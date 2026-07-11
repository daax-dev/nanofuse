package config

import (
	"strings"
	"testing"
)

func baseValidConfig() *Config {
	c := DefaultConfig()
	// Force a host-agnostic runtime so Validate does not depend on the test OS.
	c.Runtime.Driver = "auto"
	return c
}

func TestSnapshotStoreDisabledByDefault(t *testing.T) {
	if DefaultConfig().SnapshotStore.Backend != SnapshotBackendDisabled {
		t.Fatalf("snapshot store should be disabled by default")
	}
}

func TestValidateSnapshotStore(t *testing.T) {
	tests := []struct {
		name    string
		store   SnapshotStoreConfig
		wantErr string
	}{
		{name: "disabled ok", store: SnapshotStoreConfig{}},
		{name: "filesystem ok", store: SnapshotStoreConfig{Backend: "filesystem", Path: "/var/lib/nanofuse/tier"}},
		{name: "filesystem requires path", store: SnapshotStoreConfig{Backend: "filesystem"}, wantErr: "snapshot_store.path is required"},
		{name: "unknown backend", store: SnapshotStoreConfig{Backend: "s3", Path: "bucket"}, wantErr: "not supported"},
		{name: "bad compression", store: SnapshotStoreConfig{Backend: "filesystem", Path: "/x", Compression: "gzip"}, wantErr: "compression"},
		{name: "zstd compression ok", store: SnapshotStoreConfig{Backend: "filesystem", Path: "/x", Compression: "zstd"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := baseValidConfig()
			c.SnapshotStore = tt.store
			err := c.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Validate() = %v, want error containing %q", err, tt.wantErr)
			}
		})
	}
}
