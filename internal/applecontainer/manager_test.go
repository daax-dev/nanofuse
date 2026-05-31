package applecontainer

import (
	"context"
	"errors"
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
	args, err := manager.runArgs(testVM(), testImage(), "nf-test")
	if err != nil {
		t.Fatalf("runArgs() error = %v", err)
	}
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

func TestRunArgsRejectsNetworkNone(t *testing.T) {
	manager := NewManager(config.AppleContainerRuntimeConfig{DefaultCommand: "sleep infinity"}, t.TempDir())
	vm := testVM()
	vm.Config.Network.Mode = "none"

	_, err := manager.runArgs(vm, testImage(), "nf-test")
	if err == nil {
		t.Fatal("expected network none to be rejected")
	}
	if !strings.Contains(err.Error(), `network mode "none"`) {
		t.Fatalf("runArgs() error = %v, want network mode none", err)
	}
}

func TestExecRunsCommandInRuntimeContainer(t *testing.T) {
	manager := NewManager(config.AppleContainerRuntimeConfig{DefaultCommand: "sleep infinity"}, t.TempDir())
	var gotArgs []string
	manager.execCommand = func(_ context.Context, args ...string) ([]byte, []byte, int, error) {
		gotArgs = append([]string(nil), args...)
		return []byte("Linux\n"), []byte(""), 0, nil
	}

	result, err := manager.Exec(context.Background(), &types.VM{
		ID:      "vm-test",
		Runtime: &types.VMRuntime{ExternalID: "nf-test"},
	}, []string{"uname", "-a"})
	if err != nil {
		t.Fatalf("Exec() error = %v", err)
	}
	if strings.Join(gotArgs, " ") != "exec nf-test uname -a" {
		t.Fatalf("exec args = %#v", gotArgs)
	}
	if result.Stdout != "Linux\n" {
		t.Fatalf("stdout = %q, want Linux", result.Stdout)
	}
	if result.RuntimeID != "nf-test" {
		t.Fatalf("runtime id = %q, want nf-test", result.RuntimeID)
	}
}

func TestExecReturnsNonZeroExitCodeWithoutTransportError(t *testing.T) {
	manager := NewManager(config.AppleContainerRuntimeConfig{DefaultCommand: "sleep infinity"}, t.TempDir())
	manager.execCommand = func(_ context.Context, args ...string) ([]byte, []byte, int, error) {
		return []byte(""), []byte("missing\n"), 127, nil
	}

	result, err := manager.Exec(context.Background(), &types.VM{
		ID:      "vm-test",
		Runtime: &types.VMRuntime{ExternalID: "nf-test"},
	}, []string{"missing-command"})
	if err != nil {
		t.Fatalf("Exec() error = %v", err)
	}
	if result.ExitCode != 127 {
		t.Fatalf("exit code = %d, want 127", result.ExitCode)
	}
	if result.Stderr != "missing\n" {
		t.Fatalf("stderr = %q, want missing", result.Stderr)
	}
}

func TestDeleteKillsRunningContainerBeforeDelete(t *testing.T) {
	binaryPath, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable() error = %v", err)
	}
	manager := NewManager(config.AppleContainerRuntimeConfig{BinaryPath: binaryPath}, t.TempDir())
	inspectStatuses := []string{"running", "stopped"}
	commands := []string{}
	manager.runCommand = func(args ...string) ([]byte, error) {
		command := strings.Join(args, " ")
		commands = append(commands, command)
		switch command {
		case "inspect nf-test":
			if len(inspectStatuses) == 0 {
				t.Fatalf("unexpected extra inspect")
			}
			status := inspectStatuses[0]
			inspectStatuses = inspectStatuses[1:]
			return []byte(`[{"status":"` + status + `"}]`), nil
		case "kill nf-test":
			return nil, nil
		case "delete nf-test":
			return nil, nil
		default:
			t.Fatalf("unexpected command: %v", args)
			return nil, nil
		}
	}

	err = manager.Delete(&types.VM{
		ID:      "vm-test",
		Runtime: &types.VMRuntime{ExternalID: "nf-test"},
	})
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	want := []string{"inspect nf-test", "kill nf-test", "inspect nf-test", "delete nf-test"}
	if strings.Join(commands, "|") != strings.Join(want, "|") {
		t.Fatalf("commands = %#v, want %#v", commands, want)
	}
}

func TestDeleteReturnsKillErrorBeforeDelete(t *testing.T) {
	binaryPath, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable() error = %v", err)
	}
	manager := NewManager(config.AppleContainerRuntimeConfig{BinaryPath: binaryPath}, t.TempDir())
	commands := []string{}
	killErr := errors.New("kill failed")
	manager.runCommand = func(args ...string) ([]byte, error) {
		command := strings.Join(args, " ")
		commands = append(commands, command)
		switch command {
		case "inspect nf-test":
			return []byte(`[{"status":"running"}]`), nil
		case "kill nf-test":
			return nil, killErr
		case "delete nf-test":
			t.Fatal("delete should not run after kill failure")
			return nil, nil
		default:
			t.Fatalf("unexpected command: %v", args)
			return nil, nil
		}
	}

	err = manager.Delete(&types.VM{
		ID:      "vm-test",
		Runtime: &types.VMRuntime{ExternalID: "nf-test"},
	})
	if err == nil {
		t.Fatal("expected Delete to return kill error")
	}
	if !strings.Contains(err.Error(), "kill failed") {
		t.Fatalf("Delete error = %v, want kill failure", err)
	}

	want := []string{"inspect nf-test", "kill nf-test"}
	if strings.Join(commands, "|") != strings.Join(want, "|") {
		t.Fatalf("commands = %#v, want %#v", commands, want)
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
