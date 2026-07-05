package api

import (
	"testing"

	"github.com/daax-dev/nanofuse/internal/types"
)

func ptrStr(s string) *string { return &s }

func TestBuildVMConfigAppliesMountsAndSecrets(t *testing.T) {
	image := &types.Image{RootfsPath: "/var/lib/nanofuse/img/rootfs.ext4"}
	req := &types.CreateVMRequest{
		Image: "base",
		Config: &types.VMConfigRequest{
			Mounts: &[]types.Mount{
				{Source: "/srv/data", Target: "/data", Type: "bind", ReadOnly: true},
				{Target: "/scratch", Type: "tmpfs"},
			},
			Secrets: &[]types.SecretRef{
				{Name: "API_TOKEN", Source: "vault://kv/token"},
				{Name: "tls", Type: "file", Target: "/etc/tls/key.pem", Source: "spire://"},
			},
		},
	}

	config := buildVMConfig(image, req)
	if len(config.Mounts) != 2 {
		t.Fatalf("want 2 mounts, got %d", len(config.Mounts))
	}
	if len(config.Secrets) != 2 {
		t.Fatalf("want 2 secrets, got %d", len(config.Secrets))
	}

	mounts, err := types.NormalizeAndValidateMounts(config.Mounts)
	if err != nil {
		t.Fatalf("mounts validation: %v", err)
	}
	if mounts[0].Type != types.MountTypeBind || !mounts[0].ReadOnly {
		t.Fatalf("unexpected mount[0]: %+v", mounts[0])
	}

	secrets, err := types.NormalizeAndValidateSecrets(config.Secrets)
	if err != nil {
		t.Fatalf("secrets validation: %v", err)
	}
	if secrets[0].Type != types.SecretTypeEnv || secrets[0].Target != "API_TOKEN" {
		t.Fatalf("unexpected secret[0]: %+v", secrets[0])
	}
}

func TestBuildVMConfigNoMountsOrSecretsByDefault(t *testing.T) {
	image := &types.Image{RootfsPath: "/x/rootfs.ext4"}
	req := &types.CreateVMRequest{Image: "base", Config: &types.VMConfigRequest{KernelArgs: ptrStr("console=ttyS0")}}
	config := buildVMConfig(image, req)
	if config.Mounts != nil || config.Secrets != nil {
		t.Fatalf("expected no mounts/secrets, got mounts=%v secrets=%v", config.Mounts, config.Secrets)
	}
}
