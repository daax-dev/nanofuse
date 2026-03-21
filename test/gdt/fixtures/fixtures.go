// Package fixtures provides shared test fixtures for gdt tests.
package fixtures

import (
	"os"
	"os/exec"
	"path/filepath"
)

// ProjectRoot returns the project root directory.
func ProjectRoot() string {
	// Try to find project root by looking for go.mod
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			return ""
		}
		dir = parent
	}
}

// BuildArtifactsExist checks if build artifacts are present.
func BuildArtifactsExist() bool {
	root := ProjectRoot()
	if root == "" {
		return false
	}

	// Check for kernel
	kernelPath := filepath.Join(root, "images", "base", "build", "vmlinux")
	if _, err := os.Stat(kernelPath); err != nil {
		return false
	}

	// Check for rootfs
	rootfsPath := filepath.Join(root, "images", "base", "build", "rootfs.ext4")
	if _, err := os.Stat(rootfsPath); err != nil {
		return false
	}

	return true
}

// KVMAvailable checks if KVM is available.
func KVMAvailable() bool {
	_, err := os.Stat("/dev/kvm")
	return err == nil
}

// FirecrackerInstalled checks if firecracker binary is available.
func FirecrackerInstalled() bool {
	_, err := exec.LookPath("firecracker")
	return err == nil
}

// DaemonRunning checks if nanofused daemon is running.
func DaemonRunning() bool {
	// Check for socket
	socketPaths := []string{
		"/var/run/nanofused.sock",
		"/run/nanofused.sock",
		"/tmp/nanofused.sock",
	}

	for _, path := range socketPaths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	// Check systemctl
	cmd := exec.Command("systemctl", "is-active", "--quiet", "nanofused")
	return cmd.Run() == nil
}

// CanRunWithSudo checks if we have passwordless sudo access.
func CanRunWithSudo() bool {
	cmd := exec.Command("sudo", "-n", "true")
	return cmd.Run() == nil
}

// IsRoot checks if running as root.
func IsRoot() bool {
	return os.Geteuid() == 0
}
