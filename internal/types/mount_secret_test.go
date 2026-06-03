package types

import "testing"

func TestNormalizeAndValidateMounts(t *testing.T) {
	tests := []struct {
		name    string
		in      []Mount
		wantErr bool
		check   func(t *testing.T, out []Mount)
	}{
		{name: "empty", in: nil, check: func(t *testing.T, out []Mount) {
			if out != nil {
				t.Fatalf("want nil, got %v", out)
			}
		}},
		{name: "bind defaults type", in: []Mount{{Source: "/srv/data", Target: "/data"}}, check: func(t *testing.T, out []Mount) {
			if len(out) != 1 || out[0].Type != MountTypeBind {
				t.Fatalf("want bind default, got %+v", out)
			}
		}},
		{name: "tmpfs ok", in: []Mount{{Target: "/tmp/work", Type: "tmpfs"}}},
		{name: "tmpfs with source rejected", in: []Mount{{Source: "/x", Target: "/t", Type: "tmpfs"}}, wantErr: true},
		{name: "bind missing source", in: []Mount{{Target: "/data", Type: "bind"}}, wantErr: true},
		{name: "missing target", in: []Mount{{Source: "/srv"}}, wantErr: true},
		{name: "relative target", in: []Mount{{Source: "/srv", Target: "data"}}, wantErr: true},
		{name: "bad type", in: []Mount{{Source: "/srv", Target: "/d", Type: "nfs"}}, wantErr: true},
		{name: "duplicate target", in: []Mount{{Source: "/a", Target: "/d"}, {Source: "/b", Target: "/d"}}, wantErr: true},
		{name: "duplicate target trailing slash", in: []Mount{{Source: "/a", Target: "/d"}, {Source: "/b", Target: "/d/"}}, wantErr: true},
		{name: "target normalized", in: []Mount{{Source: "/a", Target: "/data//sub/"}}, check: func(t *testing.T, out []Mount) {
			if out[0].Target != "/data/sub" {
				t.Fatalf("want cleaned target, got %q", out[0].Target)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := NormalizeAndValidateMounts(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, out)
			}
		})
	}
}

func TestNormalizeAndValidateSecrets(t *testing.T) {
	tests := []struct {
		name    string
		in      []SecretRef
		wantErr bool
		check   func(t *testing.T, out []SecretRef)
	}{
		{name: "empty", in: nil, check: func(t *testing.T, out []SecretRef) {
			if out != nil {
				t.Fatalf("want nil, got %v", out)
			}
		}},
		{name: "env defaults target to name", in: []SecretRef{{Name: "API_TOKEN", Source: "vault://kv/token"}}, check: func(t *testing.T, out []SecretRef) {
			if out[0].Type != SecretTypeEnv || out[0].Target != "API_TOKEN" {
				t.Fatalf("want env target=name, got %+v", out[0])
			}
		}},
		{name: "file ok", in: []SecretRef{{Name: "tls", Type: "file", Target: "/etc/tls/key.pem", Source: "spire://"}}},
		{name: "file target normalized", in: []SecretRef{{Name: "tls", Type: "file", Target: "/etc/tls//key.pem/", Source: "spire://"}}, check: func(t *testing.T, out []SecretRef) {
			if out[0].Target != "/etc/tls/key.pem" {
				t.Fatalf("want cleaned target, got %q", out[0].Target)
			}
		}},
		{name: "file missing target", in: []SecretRef{{Name: "tls", Type: "file", Source: "spire://"}}, wantErr: true},
		{name: "file relative target", in: []SecretRef{{Name: "tls", Type: "file", Target: "rel", Source: "spire://"}}, wantErr: true},
		{name: "missing name", in: []SecretRef{{Source: "vault://x"}}, wantErr: true},
		{name: "missing source", in: []SecretRef{{Name: "x"}}, wantErr: true},
		{name: "bad type", in: []SecretRef{{Name: "x", Type: "kms", Source: "vault://x"}}, wantErr: true},
		{name: "duplicate name", in: []SecretRef{{Name: "a", Source: "x"}, {Name: "a", Source: "y"}}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := NormalizeAndValidateSecrets(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, out)
			}
		})
	}
}
