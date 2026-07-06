package gondolin

import (
	"strings"
	"testing"
)

// findDiv returns the first divergence matching feature, or nil.
func findDiv(divs []Divergence, feature string) *Divergence {
	for i := range divs {
		if divs[i].Feature == feature {
			return &divs[i]
		}
	}
	return nil
}

func TestConvert_CleanFields(t *testing.T) {
	sb := &Sandbox{
		Image:     "ghcr.io/acme/agent:latest",
		Resources: &Resources{VCPUs: iptr(4), MemoryMiB: iptr(2048)},
	}
	req, divs, err := Convert(sb, Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Image != "ghcr.io/acme/agent:latest" {
		t.Fatalf("image = %q", req.Image)
	}
	if req.Config.VCPUs != 4 || req.Config.MemoryMiB != 2048 {
		t.Fatalf("resources = %d/%d", req.Config.VCPUs, req.Config.MemoryMiB)
	}
	if req.Config.Network.Mode != "nat" {
		t.Fatalf("network mode = %q", req.Config.Network.Mode)
	}
	if req.Config.Network.EgressPolicy != nil {
		t.Fatalf("expected no egress policy, got %+v", req.Config.Network.EgressPolicy)
	}
	if len(divs) != 0 {
		t.Fatalf("expected no divergences, got %+v", divs)
	}
}

func TestConvert_MissingImageErrors(t *testing.T) {
	_, _, err := Convert(&Sandbox{}, Options{})
	if err == nil || !strings.Contains(err.Error(), "image is required") {
		t.Fatalf("expected image-required error, got %v", err)
	}
}

func TestConvert_DefaultedResourcesDisclosed(t *testing.T) {
	req, divs, err := Convert(&Sandbox{Image: "img"}, Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Config.VCPUs != DefaultVCPUs || req.Config.MemoryMiB != DefaultMemoryMiB {
		t.Fatalf("defaults not applied: %d/%d", req.Config.VCPUs, req.Config.MemoryMiB)
	}
	d := findDiv(divs, "resources")
	if d == nil {
		t.Fatalf("expected disclosed resources divergence, got %+v", divs)
	}
	if d.Severity != SeverityInfo {
		t.Fatalf("resources severity = %q, want info", d.Severity)
	}
	if !strings.Contains(d.Detail, "no CPU/memory model") {
		t.Fatalf("resources detail = %q", d.Detail)
	}
}

func TestConvert_PartialResourcesDisclosed(t *testing.T) {
	// vcpus set, memory unset -> memory defaulted + disclosed.
	req, divs, err := Convert(&Sandbox{Image: "img", Resources: &Resources{VCPUs: iptr(8)}}, Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Config.VCPUs != 8 || req.Config.MemoryMiB != DefaultMemoryMiB {
		t.Fatalf("resources = %d/%d", req.Config.VCPUs, req.Config.MemoryMiB)
	}
	if d := findDiv(divs, "resources.memory_mib"); d == nil || d.Severity != SeverityInfo {
		t.Fatalf("expected memory_mib info divergence, got %+v", divs)
	}
	if d := findDiv(divs, "resources.vcpus"); d != nil {
		t.Fatalf("did not expect vcpus divergence, got %+v", d)
	}
}

func TestConvert_AllowHostDropAndWarn(t *testing.T) {
	sb := &Sandbox{
		Image:     "img",
		Resources: &Resources{VCPUs: iptr(2), MemoryMiB: iptr(512)},
		AllowHost: []string{"api.github.com", "*.example.com"},
	}
	req, divs, err := Convert(sb, Options{})
	if err != nil {
		t.Fatalf("unexpected error (allow-host must not fail closed): %v", err)
	}
	ep := req.Config.Network.EgressPolicy
	if ep == nil {
		t.Fatalf("expected egress policy")
	}
	if !ep.Enabled || ep.DefaultAction != "deny" {
		t.Fatalf("expected default-deny egress, got %+v", ep)
	}
	if len(ep.AllowRules) != 0 {
		t.Fatalf("expected no allow rules (safe degrade), got %+v", ep.AllowRules)
	}
	d := findDiv(divs, "--allow-host")
	if d == nil || d.Severity != SeverityWarn {
		t.Fatalf("expected --allow-host warn divergence, got %+v", divs)
	}
	if !strings.Contains(d.Detail, "default-deny") {
		t.Fatalf("allow-host detail = %q", d.Detail)
	}
}

func TestConvert_ResolveEgressWithResolver(t *testing.T) {
	resolver := func(host string) ([]string, error) {
		switch host {
		case "api.github.com":
			return []string{"140.82.113.5"}, nil
		default:
			return nil, nil
		}
	}
	sb := &Sandbox{
		Image:     "img",
		Resources: &Resources{VCPUs: iptr(2), MemoryMiB: iptr(512)},
		AllowHost: []string{"api.github.com", "*.example.com", "https://x/y"},
	}
	req, divs, err := Convert(sb, Options{ResolveEgress: true, Resolver: resolver})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ep := req.Config.Network.EgressPolicy
	if ep == nil || len(ep.AllowRules) != 1 {
		t.Fatalf("expected 1 resolved allow rule, got %+v", ep)
	}
	if ep.AllowRules[0].CIDR != "140.82.113.5/32" || ep.AllowRules[0].Port != 443 {
		t.Fatalf("resolved rule = %+v", ep.AllowRules[0])
	}
	// The wildcard and URL entries must still be reported as dropped.
	d := findDiv(divs, "--allow-host")
	if d == nil || !strings.Contains(d.Detail, "*.example.com") {
		t.Fatalf("expected dropped wildcard reported, got %+v", divs)
	}
}

func TestConvert_ResolveEgressCanonicalizesAndSkipsInvalid(t *testing.T) {
	resolver := func(host string) ([]string, error) {
		switch host {
		case "v6.example.com":
			return []string{"2606:4700:4700::1111"}, nil
		case "mapped.example.com":
			return []string{"::ffff:1.2.3.4"}, nil // IPv4-in-IPv6 -> plain /32
		case "bad.example.com":
			return []string{"not-an-ip"}, nil // skipped, host dropped
		default:
			return nil, nil
		}
	}
	sb := &Sandbox{
		Image:     "img",
		Resources: &Resources{VCPUs: iptr(2), MemoryMiB: iptr(512)},
		AllowHost: []string{"v6.example.com", "mapped.example.com", "bad.example.com"},
	}
	req, divs, err := Convert(sb, Options{ResolveEgress: true, Resolver: resolver})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cidrs := map[string]bool{}
	for _, r := range req.Config.Network.EgressPolicy.AllowRules {
		cidrs[r.CIDR] = true
	}
	if !cidrs["2606:4700:4700::1111/128"] {
		t.Errorf("expected IPv6 /128 rule; got %v", cidrs)
	}
	if !cidrs["1.2.3.4/32"] {
		t.Errorf("expected IPv4-in-IPv6 canonicalized to 1.2.3.4/32; got %v", cidrs)
	}
	if cidrs["::ffff:1.2.3.4/32"] {
		t.Errorf("must not emit the IPv4-in-IPv6 literal form")
	}
	// The non-IP resolver result means bad.example.com produced no rule -> dropped.
	if d := findDiv(divs, "--allow-host"); d == nil || !strings.Contains(d.Detail, "bad.example.com") {
		t.Errorf("expected bad.example.com reported as dropped; got %+v", divs)
	}
}

func TestConvert_ResolveEgressWithoutResolver(t *testing.T) {
	sb := &Sandbox{
		Image:     "img",
		Resources: &Resources{VCPUs: iptr(2), MemoryMiB: iptr(512)},
		AllowHost: []string{"api.github.com"},
	}
	_, divs, err := Convert(sb, Options{ResolveEgress: true, Resolver: nil})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	d := findDiv(divs, "--allow-host")
	if d == nil || !strings.Contains(d.Detail, "no resolver is configured") {
		t.Fatalf("expected missing-resolver warning, not a use--resolve-egress hint; got %+v", divs)
	}
	if strings.Contains(d.Detail, "Use --resolve-egress") {
		t.Errorf("must not tell the user to use --resolve-egress when they already did")
	}
}

// unrepresentableCases drives the fail-closed / allow-lossy behaviour for every
// gondolin feature that has no faithful nanofuse equivalent.
func TestConvert_UnrepresentableFeatures(t *testing.T) {
	base := func() *Sandbox {
		return &Sandbox{Image: "img", Resources: &Resources{VCPUs: iptr(2), MemoryMiB: iptr(512)}}
	}
	cases := []struct {
		name    string
		mutate  func(*Sandbox)
		feature string
	}{
		{"vmm", func(s *Sandbox) { s.VMM = "krun" }, "--vmm"},
		{"cwd", func(s *Sandbox) { s.Cwd = "/workspace" }, "--cwd"},
		{"env", func(s *Sandbox) { s.Env = map[string]string{"K": "V"} }, "--env"},
		{"host_secret", func(s *Sandbox) { s.HostSecret = []string{"TOK"} }, "--host-secret"},
		{"mount_hostfs", func(s *Sandbox) { s.MountHostFS = []string{"/a:/b"} }, "--mount-hostfs"},
		{"mount_memfs", func(s *Sandbox) { s.MountMemFS = []string{"/tmp/x"} }, "--mount-memfs"},
		{"ssh_allow_host", func(s *Sandbox) { s.SSHAllowHost = []string{"h:22"} }, "--ssh-allow-host"},
		{"tcp_map", func(s *Sandbox) { s.TCPMap = []string{"a=b:1"} }, "--tcp-map"},
		{"rootfs_size", func(s *Sandbox) { s.RootfsSize = "8G" }, "--rootfs-size"},
		{"dns", func(s *Sandbox) { s.DNS = "open" }, "--dns"},
	}

	for _, tc := range cases {
		t.Run(tc.name+"/fail-closed", func(t *testing.T) {
			sb := base()
			tc.mutate(sb)
			_, divs, err := Convert(sb, Options{})
			if err == nil {
				t.Fatalf("expected fail-closed error for %s", tc.feature)
			}
			if !strings.Contains(err.Error(), "fail-closed") || !strings.Contains(err.Error(), tc.feature) {
				t.Fatalf("error = %q, want fail-closed mentioning %s", err.Error(), tc.feature)
			}
			d := findDiv(divs, tc.feature)
			if d == nil || d.Severity != SeverityBlocking {
				t.Fatalf("expected blocking divergence for %s, got %+v", tc.feature, divs)
			}
		})

		t.Run(tc.name+"/allow-lossy", func(t *testing.T) {
			sb := base()
			tc.mutate(sb)
			_, divs, err := Convert(sb, Options{AllowLossy: true})
			if err != nil {
				t.Fatalf("allow-lossy must proceed, got error: %v", err)
			}
			d := findDiv(divs, tc.feature)
			if d == nil || d.Severity != SeverityWarn {
				t.Fatalf("expected downgraded warn for %s, got %+v", tc.feature, divs)
			}
			if !strings.Contains(d.Detail, "LOSSY (--allow-lossy)") {
				t.Fatalf("expected loud drop marker, got %q", d.Detail)
			}
		})
	}
}

func TestConvert_NoBlockingSilentlyDropped(t *testing.T) {
	// A sandbox exercising every unrepresentable feature must report ALL of
	// them; none may be silently dropped.
	sb := &Sandbox{
		Image:        "img",
		VMM:          "qemu",
		Cwd:          "/w",
		Env:          map[string]string{"A": "1"},
		AllowHost:    []string{"h.example.com"},
		HostSecret:   []string{"S"},
		MountHostFS:  []string{"/a:/b"},
		MountMemFS:   []string{"/m"},
		SSHAllowHost: []string{"h:22"},
		TCPMap:       []string{"a=b:1"},
		DNS:          "trusted",
		RootfsSize:   "16G",
	}
	_, divs, err := Convert(sb, Options{})
	if err == nil {
		t.Fatalf("expected fail-closed")
	}
	want := []string{"--vmm", "--cwd", "--env", "--host-secret", "--mount-hostfs",
		"--mount-memfs", "--ssh-allow-host", "--tcp-map", "--dns", "--rootfs-size"}
	for _, f := range want {
		d := findDiv(divs, f)
		if d == nil {
			t.Fatalf("feature %s silently dropped (not in report): %+v", f, divs)
		}
		if d.Severity != SeverityBlocking {
			t.Fatalf("feature %s severity = %q, want blocking", f, d.Severity)
		}
	}
	// allow-host is a safe degrade (warn), still reported.
	if d := findDiv(divs, "--allow-host"); d == nil || d.Severity != SeverityWarn {
		t.Fatalf("allow-host must be a reported warn, got %+v", d)
	}
}

func TestConvert_Deterministic(t *testing.T) {
	sb := &Sandbox{Image: "img", Env: map[string]string{"B": "2", "A": "1"},
		HostSecret: []string{"X"}, MountHostFS: []string{"/a:/b"}}
	_, d1, _ := Convert(sb, Options{AllowLossy: true})
	_, d2, _ := Convert(sb, Options{AllowLossy: true})
	if len(d1) != len(d2) {
		t.Fatalf("length mismatch %d vs %d", len(d1), len(d2))
	}
	for i := range d1 {
		if d1[i] != d2[i] {
			t.Fatalf("divergence %d differs: %+v vs %+v", i, d1[i], d2[i])
		}
	}
}

func TestRenderSpecYAML_Golden(t *testing.T) {
	sb := &Sandbox{
		Image:     "ghcr.io/acme/agent:latest",
		Resources: &Resources{VCPUs: iptr(4), MemoryMiB: iptr(2048)},
		AllowHost: []string{"api.github.com"},
		DNS:       "synthetic",
	}
	// allow-lossy so the dns blocking divergence does not fail the conversion;
	// we are asserting the rendered spec shape.
	req, _, err := Convert(sb, Options{AllowLossy: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out, err := RenderSpecYAML(req)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	want := `image: ghcr.io/acme/agent:latest
config:
    vcpus: 4
    memory_mib: 2048
    network:
        mode: nat
        egress_policy:
            enabled: true
            default_action: deny
            allow_dns: true
`
	if string(out) != want {
		t.Fatalf("rendered spec mismatch:\n--- got ---\n%s\n--- want ---\n%s", out, want)
	}
}

func TestRenderSpecYAML_ResolvedRules(t *testing.T) {
	req, _, err := Convert(&Sandbox{
		Image:     "img",
		AllowHost: []string{"api.github.com"},
	}, Options{ResolveEgress: true, Resolver: func(string) ([]string, error) {
		return []string{"1.2.3.4"}, nil
	}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out, err := RenderSpecYAML(req)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if !strings.Contains(string(out), "cidr: 1.2.3.4/32") {
		t.Fatalf("expected resolved cidr in output:\n%s", out)
	}
}

func iptr(i int) *int { return &i }

func TestConvert_ExplicitNonPositiveResourcesRejected(t *testing.T) {
	for _, tc := range []struct {
		name string
		res  *Resources
	}{
		{"zero vcpus", &Resources{VCPUs: iptr(0), MemoryMiB: iptr(512)}},
		{"negative vcpus", &Resources{VCPUs: iptr(-1), MemoryMiB: iptr(512)}},
		{"zero memory", &Resources{VCPUs: iptr(2), MemoryMiB: iptr(0)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := Convert(&Sandbox{Image: "img", Resources: tc.res}, Options{})
			if err == nil {
				t.Fatal("expected an error for an explicit non-positive resource value, got nil")
			}
			if !strings.Contains(err.Error(), "must be > 0") {
				t.Errorf("error = %v, want must-be-positive rejection", err)
			}
		})
	}
}

func TestConvert_ResolveEgressDedupesDuplicateIPs(t *testing.T) {
	resolver := func(host string) ([]string, error) {
		// Same IP twice + a second distinct IP.
		return []string{"1.2.3.4", "1.2.3.4", "5.6.7.8"}, nil
	}
	sb := &Sandbox{Image: "img", Resources: &Resources{VCPUs: iptr(2), MemoryMiB: iptr(512)}, AllowHost: []string{"h"}}
	req, _, err := Convert(sb, Options{ResolveEgress: true, Resolver: resolver})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n := len(req.Config.Network.EgressPolicy.AllowRules); n != 2 {
		t.Fatalf("expected 2 deduped rules, got %d: %+v", n, req.Config.Network.EgressPolicy.AllowRules)
	}
}

func TestConvert_HostSecretValuesNotEchoed(t *testing.T) {
	sb := &Sandbox{Image: "img", Resources: &Resources{VCPUs: iptr(2), MemoryMiB: iptr(512)},
		HostSecret: []string{"SUPERSECRETVALUE123"}}
	_, divs, _ := Convert(sb, Options{AllowLossy: true})
	d := findDiv(divs, "--host-secret")
	if d == nil {
		t.Fatal("expected a --host-secret divergence")
	}
	if strings.Contains(d.Detail, "SUPERSECRETVALUE123") {
		t.Errorf("host-secret value must not be echoed in the report: %q", d.Detail)
	}
}

func TestConvert_UnknownDNSModeRejected(t *testing.T) {
	sb := &Sandbox{Image: "img", Resources: &Resources{VCPUs: iptr(2), MemoryMiB: iptr(512)}, DNS: "syntheti"}
	if _, _, err := Convert(sb, Options{AllowLossy: true}); err == nil {
		t.Fatal("expected error for unknown dns mode, got nil")
	} else if !strings.Contains(err.Error(), "unknown gondolin dns mode") {
		t.Errorf("error = %v, want unknown-dns-mode rejection", err)
	}
	// A known mode is accepted.
	sb.DNS = "synthetic"
	if _, _, err := Convert(sb, Options{AllowLossy: true}); err != nil {
		t.Errorf("known dns mode should be accepted, got: %v", err)
	}
}

func TestConvert_SetsValidDefaultKernelArgs(t *testing.T) {
	sb := &Sandbox{Image: "img", Resources: &Resources{VCPUs: iptr(2), MemoryMiB: iptr(512)}}
	req, _, err := Convert(sb, Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Config.KernelArgs == "" {
		t.Error("converted request must carry non-empty KernelArgs so it is submittable")
	}
}
