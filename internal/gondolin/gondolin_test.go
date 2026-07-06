package gondolin

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string // substring; empty => no error
		check   func(t *testing.T, sb *Sandbox)
	}{
		{
			name:  "minimal valid",
			input: "image: ghcr.io/acme/agent:latest\n",
			check: func(t *testing.T, sb *Sandbox) {
				if sb.Image != "ghcr.io/acme/agent:latest" {
					t.Fatalf("image = %q", sb.Image)
				}
			},
		},
		{
			name: "full valid",
			input: `image: ghcr.io/acme/agent:latest
vmm: krun
cwd: /workspace
env:
  API_KEY: secret
  LANG: en_US
allow_host:
  - api.github.com
  - "*.example.com"
host_secret:
  - OPENAI_API_KEY
mount_hostfs:
  - "/data:/mnt/data:ro"
mount_memfs:
  - /tmp/scratch
ssh_allow_host:
  - "git.example.com:22"
tcp_map:
  - "db:5432=10.0.0.5:5432"
dns: synthetic
rootfs_size: 8G
resources:
  vcpus: 4
  memory_mib: 2048
`,
			check: func(t *testing.T, sb *Sandbox) {
				if len(sb.Env) != 2 || sb.Env["API_KEY"] != "secret" {
					t.Fatalf("env = %v", sb.Env)
				}
				if sb.Resources == nil || sb.Resources.VCPUs == nil || *sb.Resources.VCPUs != 4 ||
					sb.Resources.MemoryMiB == nil || *sb.Resources.MemoryMiB != 2048 {
					t.Fatalf("resources = %+v", sb.Resources)
				}
				if sb.DNS != "synthetic" || sb.VMM != "krun" {
					t.Fatalf("dns/vmm = %q/%q", sb.DNS, sb.VMM)
				}
			},
		},
		{
			name:    "empty document",
			input:   "",
			wantErr: "empty gondolin mirror",
		},
		{
			name:    "whitespace only",
			input:   "   \n\t\n",
			wantErr: "empty gondolin mirror",
		},
		{
			name:    "unknown key rejected",
			input:   "image: x\nprivileged: true\n",
			wantErr: "field privileged not found",
		},
		{
			name:    "wrong type rejected",
			input:   "image:\n  - not-a-string\n",
			wantErr: "parse gondolin mirror",
		},
		{
			name: "multiple documents rejected",
			input: `image: a
---
image: b
`,
			wantErr: "multiple YAML documents",
		},
		{
			name:    "malformed yaml",
			input:   "image: [unterminated\n",
			wantErr: "parse gondolin mirror",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sb, err := Parse([]byte(tc.input))
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.wantErr)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("error = %q, want substring %q", err.Error(), tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.check != nil {
				tc.check(t, sb)
			}
		})
	}
}
