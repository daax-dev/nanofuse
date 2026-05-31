package applecontainer

import (
	"os"
	"strings"
	"testing"

	"github.com/daax-dev/nanofuse/internal/config"
	"github.com/daax-dev/nanofuse/internal/types"
)

func TestResolveImageParsesAppleContainerInspect(t *testing.T) {
	binaryPath, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable() error = %v", err)
	}
	manager := NewManager(config.AppleContainerRuntimeConfig{
		BinaryPath:     binaryPath,
		DefaultCommand: "sleep infinity",
	}, t.TempDir())
	manager.runCommand = func(args ...string) ([]byte, error) {
		switch strings.Join(args, " ") {
		case "system status":
			return []byte("apiserver is running"), nil
		case "images inspect alpine:3.20":
			return []byte(`[{
				"name":"docker.io/library/alpine:3.20",
				"index":{"digest":"sha256:index","mediaType":"application/vnd.oci.image.index.v1+json","size":9226},
				"variants":[{
					"size":4093973,
					"platform":{"os":"linux","architecture":"arm64","variant":"v8"},
					"config":{"os":"linux","architecture":"arm64","variant":"v8"}
				}]
			}]`), nil
		default:
			t.Fatalf("unexpected command: %v", args)
			return nil, nil
		}
	}

	image, err := manager.ResolveImage("alpine:3.20")
	if err != nil {
		t.Fatalf("ResolveImage() error = %v", err)
	}
	if image.Digest != "sha256:index" {
		t.Fatalf("Digest = %q", image.Digest)
	}
	if len(image.Tags) != 1 || image.Tags[0] != "docker.io/library/alpine:3.20" {
		t.Fatalf("Tags = %#v", image.Tags)
	}
	if image.Labels["nanofuse.runtime"] != DriverName {
		t.Fatalf("runtime label = %q", image.Labels["nanofuse.runtime"])
	}
}

func TestRunArgsMapsVMConfigToAppleContainerCLI(t *testing.T) {
	manager := NewManager(config.AppleContainerRuntimeConfig{DefaultCommand: "sleep infinity"}, t.TempDir())
	args := manager.runArgs(testVM(), testImage(), "nf-test")
	got := strings.Join(args, " ")

	for _, want := range []string{
		"run -d",
		"--name nf-test",
		"--cpus 2",
		"--memory 512M",
		"--label nanofuse.vm_id=vm-1234567890abcdef",
		"--publish 127.0.0.1:18081:8080/tcp",
		"docker.io/library/alpine:3.20",
		"sleep infinity",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("args %q missing %q", got, want)
		}
	}
}

func testVM() *types.VM {
	return &types.VM{
		ID:    "vm-1234567890abcdef",
		Image: "alpine:3.20",
		Config: types.VMConfig{
			VCPUs:     2,
			MemoryMiB: 512,
			Network: types.NetworkConfig{
				PortForwards: []types.PortForward{{
					HostPort: 18081,
					VMPort:   8080,
					Protocol: "tcp",
				}},
			},
		},
	}
}

func testImage() *types.Image {
	return &types.Image{Tags: []string{"docker.io/library/alpine:3.20"}}
}
