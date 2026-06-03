package main

import (
	"fmt"
	"strings"

	"github.com/daax-dev/nanofuse/internal/client"
	"github.com/spf13/cobra"
)

// parseMountSpecs parses repeatable --mount flags into client.Mount values.
//
// Each spec is a comma-separated list of key=value pairs:
//
//	--mount src=/host/data,dst=/data,type=bind,ro
//	--mount type=tmpfs,dst=/scratch
//
// A two-field shorthand "src:dst" is also accepted for bind mounts, with an
// optional ":ro" suffix:
//
//	--mount /host/data:/data
//	--mount /host/data:/data:ro
func parseMountSpecs(specs []string) ([]client.Mount, error) {
	if len(specs) == 0 {
		return nil, nil
	}
	mounts := make([]client.Mount, 0, len(specs))
	for _, spec := range specs {
		spec = strings.TrimSpace(spec)
		if spec == "" {
			continue
		}
		var m client.Mount
		if !strings.Contains(spec, "=") {
			// shorthand src:dst[:ro|:rw]. Strip the optional mode suffix first,
			// then split on the LAST colon so a Windows drive-letter source
			// (e.g. C:\data:/data) parses correctly.
			body := spec
			switch {
			case strings.HasSuffix(body, ":ro"):
				m.ReadOnly = true
				body = strings.TrimSuffix(body, ":ro")
			case strings.HasSuffix(body, ":rw"):
				body = strings.TrimSuffix(body, ":rw")
			}
			idx := strings.LastIndex(body, ":")
			if idx <= 0 || idx == len(body)-1 {
				return nil, fmt.Errorf("invalid --mount %q: expected src:dst[:ro] or key=value list", spec)
			}
			m.Source = body[:idx]
			m.Target = body[idx+1:]
			m.Type = "bind"
			mounts = append(mounts, m)
			continue
		}
		for _, field := range strings.Split(spec, ",") {
			field = strings.TrimSpace(field)
			if field == "" {
				continue
			}
			key, value, hasValue := strings.Cut(field, "=")
			key = strings.ToLower(strings.TrimSpace(key))
			value = strings.TrimSpace(value)
			switch key {
			case "src", "source":
				m.Source = value
			case "dst", "target", "destination":
				m.Target = value
			case "type":
				m.Type = strings.ToLower(value)
			case "ro", "readonly", "read_only":
				// Bare "ro" means read-only; an explicit value is parsed
				// case-insensitively as a boolean.
				if !hasValue {
					m.ReadOnly = true
					break
				}
				switch strings.ToLower(value) {
				case "", "true", "1", "yes", "on":
					m.ReadOnly = true
				case "false", "0", "no", "off":
					m.ReadOnly = false
				default:
					return nil, fmt.Errorf("invalid --mount %q: ro must be true or false", spec)
				}
			default:
				return nil, fmt.Errorf("invalid --mount %q: unknown key %q", spec, key)
			}
		}
		mounts = append(mounts, m)
	}
	return mounts, nil
}

// parseSecretSpecs parses repeatable --secret flags into client.SecretRef values.
//
// Each spec is a comma-separated list of key=value pairs (references only, never
// a value):
//
//	--secret name=API_TOKEN,source=vault://kv/token
//	--secret name=tls_key,type=file,target=/etc/tls/key.pem,source=spire://
func parseSecretSpecs(specs []string) ([]client.SecretRef, error) {
	if len(specs) == 0 {
		return nil, nil
	}
	secrets := make([]client.SecretRef, 0, len(specs))
	for _, spec := range specs {
		spec = strings.TrimSpace(spec)
		if spec == "" {
			continue
		}
		var s client.SecretRef
		for _, field := range strings.Split(spec, ",") {
			field = strings.TrimSpace(field)
			if field == "" {
				continue
			}
			key, value, _ := strings.Cut(field, "=")
			key = strings.ToLower(strings.TrimSpace(key))
			value = strings.TrimSpace(value)
			switch key {
			case "name":
				s.Name = value
			case "source", "src", "ref":
				s.Source = value
			case "type":
				s.Type = strings.ToLower(value)
			case "target", "dst", "as":
				s.Target = value
			case "value", "secret":
				return nil, fmt.Errorf("invalid --secret %q: secret values are not accepted; pass a reference via source=", spec)
			default:
				return nil, fmt.Errorf("invalid --secret %q: unknown key %q", spec, key)
			}
		}
		if s.Name == "" {
			return nil, fmt.Errorf("invalid --secret %q: name is required", spec)
		}
		secrets = append(secrets, s)
	}
	return secrets, nil
}

type vmMountsOutput struct {
	ID     string         `json:"id"`
	Name   string         `json:"name"`
	State  string         `json:"state"`
	Mounts []client.Mount `json:"mounts"`
}

type vmSecretsOutput struct {
	ID      string             `json:"id"`
	Name    string             `json:"name"`
	State   string             `json:"state"`
	Secrets []client.SecretRef `json:"secrets"`
}

func collectVMsForQuery(cmd *cobra.Command, args []string, op string) ([]client.VM, error) {
	if len(args) == 1 {
		vm, err := apiClient.GetVM(cmd.Context(), args[0])
		if err != nil {
			return nil, handleAPIErrorWithResource(err, op, args[0])
		}
		return []client.VM{*vm}, nil
	}
	resp, err := apiClient.ListVMs(cmd.Context(), "")
	if err != nil {
		return nil, handleAPIError(err, op)
	}
	return resp.VMs, nil
}

var vmMountsCmd = &cobra.Command{
	Use:   "mounts [vm-id]",
	Short: "Show configured VM filesystem mounts",
	Long: `Show operator-declared filesystem mounts for one or all VMs.

Mounts are an operator-visible inventory surface: source, target, type, and
read-only intent. Runtime enforcement depends on the daemon backend.

Examples:
  nanofuse vm mounts
  nanofuse vm mounts my-vm`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vms, err := collectVMsForQuery(cmd, args, "get VM mounts")
		if err != nil {
			return err
		}
		out := make([]vmMountsOutput, 0, len(vms))
		for _, vm := range vms {
			out = append(out, vmMountsOutput{ID: vm.ID, Name: vm.Name, State: vm.State, Mounts: vm.Config.Mounts})
		}
		if jsonOutput {
			return formatter.PrintJSON(map[string]interface{}{"vms": out, "total": len(out)})
		}
		for _, vm := range out {
			if len(vm.Mounts) == 0 {
				fmt.Printf("%s [%s]: no mounts configured\n", displayVMLabel(vm.ID, vm.Name), vm.State)
				continue
			}
			for _, m := range vm.Mounts {
				mode := "rw"
				if m.ReadOnly {
					mode = "ro"
				}
				typ := m.Type
				if typ == "" {
					typ = "bind"
				}
				source := m.Source
				if source == "" {
					source = "-"
				}
				fmt.Printf("%s [%s] %s %s -> %s (%s)\n",
					displayVMLabel(vm.ID, vm.Name), vm.State, typ, source, m.Target, mode)
			}
		}
		return nil
	},
}

var vmSecretsCmd = &cobra.Command{
	Use:   "secrets [vm-id]",
	Short: "Show VM secret references",
	Long: `Show secret references attached to one or all VMs.

Secret references never include values: only the logical name, source
reference, delivery type, and in-guest target are shown.

Examples:
  nanofuse vm secrets
  nanofuse vm secrets my-vm`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vms, err := collectVMsForQuery(cmd, args, "get VM secrets")
		if err != nil {
			return err
		}
		out := make([]vmSecretsOutput, 0, len(vms))
		for _, vm := range vms {
			out = append(out, vmSecretsOutput{ID: vm.ID, Name: vm.Name, State: vm.State, Secrets: vm.Config.Secrets})
		}
		if jsonOutput {
			return formatter.PrintJSON(map[string]interface{}{"vms": out, "total": len(out)})
		}
		for _, vm := range out {
			if len(vm.Secrets) == 0 {
				fmt.Printf("%s [%s]: no secret references configured\n", displayVMLabel(vm.ID, vm.Name), vm.State)
				continue
			}
			for _, s := range vm.Secrets {
				typ := s.Type
				if typ == "" {
					typ = "env"
				}
				source := s.Source
				if source == "" {
					source = "-"
				}
				target := s.Target
				if target == "" {
					target = s.Name
				}
				fmt.Printf("%s [%s] %s (%s) source=%s target=%s\n",
					displayVMLabel(vm.ID, vm.Name), vm.State, s.Name, typ, source, target)
			}
		}
		return nil
	},
}
