package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateAuthTCPRequiresTLSFiles(t *testing.T) {
	tests := []struct {
		name    string
		cfg     AuthConfig
		wantErr string
	}{
		{
			name: "missing cert",
			cfg: AuthConfig{
				Enabled:      true,
				TLSKeyFile:   "/etc/nanofuse/server.key",
				ClientCAFile: "/etc/nanofuse/client-ca.pem",
			},
			wantErr: "auth.tls_cert_file is required",
		},
		{
			name: "missing key",
			cfg: AuthConfig{
				Enabled:      true,
				TLSCertFile:  "/etc/nanofuse/server.pem",
				ClientCAFile: "/etc/nanofuse/client-ca.pem",
			},
			wantErr: "auth.tls_key_file is required",
		},
		{
			name: "missing client CA",
			cfg: AuthConfig{
				Enabled:     true,
				TLSCertFile: "/etc/nanofuse/server.pem",
				TLSKeyFile:  "/etc/nanofuse/server.key",
			},
			wantErr: "auth.client_ca_file is required",
		},
		{
			name: "complete",
			cfg: AuthConfig{
				Enabled:      true,
				TLSCertFile:  "/etc/nanofuse/server.pem",
				TLSKeyFile:   "/etc/nanofuse/server.key",
				ClientCAFile: "/etc/nanofuse/client-ca.pem",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.API.TCPBind = "127.0.0.1:8080"
			cfg.Auth = tt.cfg

			err := cfg.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate returned error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate returned nil, want error containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Validate error = %q, want substring %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestValidateAuthUnixOnlyDoesNotRequireTLSFiles(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Auth.Enabled = true

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate returned error for Unix-only auth config: %v", err)
	}
}

func TestNetworkSetupDefaultsEnabled(t *testing.T) {
	cfg := DefaultConfig()
	if !cfg.Network.Setup {
		t.Fatal("expected network setup to default to enabled")
	}
}

func TestLoadNetworkSetupDisabled(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "nanofused.yaml")
	if err := os.WriteFile(configPath, []byte(`
storage:
  data_dir: /tmp/nanofuse-test
  database: /tmp/nanofuse-test/nanofuse.db
firecracker:
  binary_path: /usr/local/bin/firecracker
network:
  setup: false
`), 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Network.Setup {
		t.Fatal("expected explicit network.setup=false to be honored")
	}
}

func TestRuntimeDriverForHostRejectsIncompatibleDrivers(t *testing.T) {
	tests := []struct {
		name    string
		driver  string
		goos    string
		want    string
		wantErr string
	}{
		{name: "auto darwin", driver: "auto", goos: "darwin", want: "apple_container"},
		{name: "auto linux", driver: "auto", goos: "linux", want: "firecracker"},
		{name: "empty defaults to auto", driver: "", goos: "linux", want: "firecracker"},
		{name: "explicit firecracker linux", driver: "firecracker", goos: "linux", want: "firecracker"},
		{name: "explicit apple darwin", driver: "apple_container", goos: "darwin", want: "apple_container"},
		{name: "firecracker darwin rejected", driver: "firecracker", goos: "darwin", wantErr: "firecracker requires linux host"},
		{name: "apple linux rejected", driver: "apple_container", goos: "linux", wantErr: "apple_container requires darwin host"},
		{name: "auto windows rejected", driver: "auto", goos: "windows", wantErr: "auto does not support host OS"},
		{name: "unknown rejected", driver: "unknown", goos: "linux", wantErr: "runtime.driver must be one of"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := runtimeDriverForHost(tt.driver, tt.goos)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("runtimeDriverForHost() error = %v", err)
				}
				if got != tt.want {
					t.Fatalf("runtimeDriverForHost() = %q, want %q", got, tt.want)
				}
				return
			}
			if err == nil {
				t.Fatalf("runtimeDriverForHost() error = nil, want %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("runtimeDriverForHost() error = %q, want substring %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestValidateSPIRERequiredWithoutEnabledIsRejected(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SPIRE.Required = true
	cfg.SPIRE.Enabled = false

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() error = nil, want rejection of required-without-enabled")
	}
	if !strings.Contains(err.Error(), "spire.required") || !strings.Contains(err.Error(), "spire.enabled") {
		t.Fatalf("Validate() error = %q, want it to name spire.required and spire.enabled", err.Error())
	}
}

func TestValidateSPIRERequiredWithEnabledIsAccepted(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SPIRE.Required = true
	cfg.SPIRE.Enabled = true

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil for required+enabled", err)
	}
}
