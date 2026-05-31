package config

import (
	"os"
	"path/filepath"
	"testing"
)

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
