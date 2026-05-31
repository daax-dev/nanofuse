package config

import (
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
