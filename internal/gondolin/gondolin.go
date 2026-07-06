// Package gondolin converts a nanofuse-authored "gondolin mirror" sandbox
// definition into a nanofuse CreateVMRequest, together with an explicit
// divergence report.
//
// Reframe (important): the gondolin project (earendil-works/gondolin) has NO
// declarative sandbox config file. Its sandbox is defined imperatively through
// `gondolin bash|exec` CLI flags (verified against docs/cli.md) and a
// TypeScript VM.create() API. There is therefore nothing to "parse" from
// gondolin itself. This package instead parses a nanofuse-authored YAML that
// MIRRORS gondolin's flag surface (image, allow-host, host-secret,
// mount-hostfs/memfs, ssh-allow-host, tcp-map, dns, env, cwd, vmm,
// rootfs-size, plus nanofuse resource hints) and reports, for every gondolin
// feature, whether it maps faithfully to nanofuse or diverges.
//
// The honest value here is the precise divergence report, not a false claim of
// equivalence: gondolin's L7 HTTP allowlist, host secret injection, host/mem
// filesystem mounts, SSH egress, TCP mapping and DNS modes have no faithful
// nanofuse equivalent, and this package refuses to silently drop them.
package gondolin

import (
	"bytes"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// Sandbox is the nanofuse-authored mirror of gondolin's `bash`/`exec` flag
// surface. Field/flag correspondence (gondolin flag -> mirror key):
//
//	--image           -> image
//	--vmm             -> vmm            (qemu|krun; nanofuse only runs firecracker)
//	--cwd             -> cwd
//	--env KEY=VALUE   -> env
//	--allow-host      -> allow_host     (L7 HTTP host patterns)
//	--host-secret     -> host_secret
//	--mount-hostfs    -> mount_hostfs
//	--mount-memfs     -> mount_memfs
//	--ssh-allow-host  -> ssh_allow_host
//	--tcp-map         -> tcp_map
//	--dns MODE        -> dns            (synthetic|trusted|open)
//	--rootfs-size     -> rootfs_size
//	(no gondolin flag) -> resources     (nanofuse-authored; gondolin has no
//	                                     CPU/memory model)
type Sandbox struct {
	Image        string            `yaml:"image"`
	VMM          string            `yaml:"vmm"`
	Cwd          string            `yaml:"cwd"`
	Env          map[string]string `yaml:"env"`
	AllowHost    []string          `yaml:"allow_host"`
	HostSecret   []string          `yaml:"host_secret"`
	MountHostFS  []string          `yaml:"mount_hostfs"`
	MountMemFS   []string          `yaml:"mount_memfs"`
	SSHAllowHost []string          `yaml:"ssh_allow_host"`
	TCPMap       []string          `yaml:"tcp_map"`
	DNS          string            `yaml:"dns"`
	RootfsSize   string            `yaml:"rootfs_size"`
	Resources    *Resources        `yaml:"resources"`
}

// Resources carries nanofuse resource hints. Gondolin has no resource model,
// so any values here are a nanofuse-side authoring convenience and any defaults
// applied on their absence are disclosed as an assumption in the report.
type Resources struct {
	VCPUs     int `yaml:"vcpus"`
	MemoryMiB int `yaml:"memory_mib"`
}

// Parse decodes a gondolin mirror YAML document. Unknown keys are rejected
// (KnownFields) so that typos and adversarial/unexpected input fail loudly
// rather than being silently ignored. Empty input is an error.
func Parse(data []byte) (*Sandbox, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, fmt.Errorf("empty gondolin mirror document")
	}

	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)

	var sb Sandbox
	if err := dec.Decode(&sb); err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("empty gondolin mirror document")
		}
		return nil, fmt.Errorf("parse gondolin mirror: %w", err)
	}

	// Reject trailing documents: a single sandbox per file keeps the mapping
	// unambiguous and blocks smuggling a second doc past validation.
	var extra Sandbox
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("multiple YAML documents in gondolin mirror: only one sandbox per file is supported")
		}
		return nil, fmt.Errorf("parse gondolin mirror: %w", err)
	}

	return &sb, nil
}
