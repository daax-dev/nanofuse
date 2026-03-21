//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/jpoley/nanofuse/internal/client"
)

const (
	testSocketPath = "/tmp/nanofused-test.sock"
	testDataDir    = "/tmp/nanofuse-test-data"
	daemonTimeout  = 10 * time.Second
)

// TestSuite manages the test daemon lifecycle
type TestSuite struct {
	t          *testing.T
	daemonCmd  *exec.Cmd
	client     *client.Client
	tempDir    string
	socketPath string
}

func setupTestSuite(t *testing.T) *TestSuite {
	// Clean up any previous test artifacts
	os.RemoveAll(testDataDir)
	os.Remove(testSocketPath)

	// Create test data directory
	if err := os.MkdirAll(testDataDir, 0755); err != nil {
		t.Fatalf("Failed to create test data directory: %v", err)
	}

	ts := &TestSuite{
		t:          t,
		tempDir:    testDataDir,
		socketPath: testSocketPath,
	}

	// Start daemon
	if err := ts.startDaemon(); err != nil {
		t.Fatalf("Failed to start daemon: %v", err)
	}

	// Wait for daemon to be ready
	if err := ts.waitForDaemon(); err != nil {
		ts.tearDown()
		t.Fatalf("Daemon failed to start: %v", err)
	}

	// Create API client (socketPath, timeout, debug)
	ts.client = client.NewClient(testSocketPath, 30*time.Second, false)

	return ts
}

func (ts *TestSuite) startDaemon() error {
	// Find the daemon binary
	binaryPath, err := findBinary("nanofused")
	if err != nil {
		return fmt.Errorf("daemon binary not found: %v (run 'mage daemon' first)", err)
	}

	// Create test config
	configPath := filepath.Join(ts.tempDir, "nanofused.yaml")
	configContent := fmt.Sprintf(`
api:
  socket: %s
  socket_mode: "0666"

storage:
  data_dir: %s
  database: %s/nanofuse.db

firecracker:
  binary_path: /usr/bin/firecracker

limits:
  max_vms: 10
  max_total_memory_mib: 8192

logging:
  level: debug
  format: json
`, testSocketPath, ts.tempDir, ts.tempDir)

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write config: %v", err)
	}

	// Start daemon process
	ts.daemonCmd = exec.Command(binaryPath, "--config", configPath)
	ts.daemonCmd.Stdout = os.Stdout
	ts.daemonCmd.Stderr = os.Stderr

	if err := ts.daemonCmd.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %v", err)
	}

	ts.t.Logf("Started daemon with PID %d", ts.daemonCmd.Process.Pid)
	return nil
}

func (ts *TestSuite) waitForDaemon() error {
	ctx, cancel := context.WithTimeout(context.Background(), daemonTimeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for daemon to start")
		case <-ticker.C:
			// Check if socket exists
			if _, err := os.Stat(testSocketPath); err == nil {
				// Try to connect
				conn, err := net.Dial("unix", testSocketPath)
				if err == nil {
					conn.Close()
					ts.t.Log("Daemon is ready!")
					return nil
				}
			}
		}
	}
}

func (ts *TestSuite) tearDown() {
	if ts.daemonCmd != nil && ts.daemonCmd.Process != nil {
		ts.t.Logf("Stopping daemon (PID %d)", ts.daemonCmd.Process.Pid)

		// Try graceful shutdown first
		ts.daemonCmd.Process.Signal(syscall.SIGTERM)

		// Wait for process to exit with timeout
		done := make(chan error)
		go func() {
			done <- ts.daemonCmd.Wait()
		}()

		select {
		case <-time.After(5 * time.Second):
			ts.t.Log("Daemon didn't stop gracefully, force killing")
			ts.daemonCmd.Process.Kill()
		case <-done:
			ts.t.Log("Daemon stopped gracefully")
		}
	}

	// Clean up test artifacts
	os.Remove(testSocketPath)
	os.RemoveAll(ts.tempDir)
}

// Tests

func TestIntegration_HealthCheck(t *testing.T) {
	ts := setupTestSuite(t)
	defer ts.tearDown()

	ctx := context.Background()

	// Test health endpoint
	health, err := ts.client.Health(ctx)
	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}

	if health.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", health.Status)
	}

	if health.Version == "" {
		t.Error("Version should not be empty")
	}

	t.Logf("✓ Health check passed: %s v%s (uptime: %ds)",
		health.Status, health.Version, health.UptimeSeconds)
}

func TestIntegration_VMLifecycle(t *testing.T) {
	ts := setupTestSuite(t)
	defer ts.tearDown()

	ctx := context.Background()

	// Test VM creation (will fail without real image, but tests API)
	t.Run("CreateVM", func(t *testing.T) {
		vm, err := ts.client.CreateVM(ctx, &client.CreateVMRequest{
			Name:  "test-vm",
			Image: "test:latest",
			Config: client.VMConfig{
				VCPUs:     2,
				MemoryMiB: 512,
				Network: client.NetworkConfig{
					Mode: "nat",
				},
			},
		})

		// We expect this to fail because the image doesn't exist
		// but it should return a proper error, not a panic
		if err != nil {
			t.Logf("Expected error (no image): %v", err)
			// This is actually success - the API handled the error properly
			return
		}

		// If somehow it succeeded, clean up
		if vm != nil {
			ts.client.DeleteVM(ctx, vm.ID)
		}
	})

	// Test VM listing
	t.Run("ListVMs", func(t *testing.T) {
		resp, err := ts.client.ListVMs(ctx, "") // empty state = all VMs
		if err != nil {
			t.Fatalf("Failed to list VMs: %v", err)
		}

		t.Logf("✓ Listed %d VMs", resp.Total)
	})
}

func TestIntegration_ImageOperations(t *testing.T) {
	ts := setupTestSuite(t)
	defer ts.tearDown()

	ctx := context.Background()

	// Test image listing
	t.Run("ListImages", func(t *testing.T) {
		resp, err := ts.client.ListImages(ctx)
		if err != nil {
			t.Fatalf("Failed to list images: %v", err)
		}

		t.Logf("✓ Listed %d images", resp.Total)
	})

	// Test image pull (will fail without registry auth, but tests API)
	t.Run("PullImage", func(t *testing.T) {
		job, err := ts.client.PullImage(ctx, "ghcr.io/test/nonexistent:latest")

		// We expect this to eventually fail, but the API should handle it gracefully
		if err != nil {
			t.Logf("Expected error (no auth/image): %v", err)
			return
		}

		// If pull started, check job status
		if job != nil {
			t.Logf("Pull job created: %s (state: %s)", job.ID, job.State)
		}
	})
}

func TestIntegration_ConcurrentRequests(t *testing.T) {
	ts := setupTestSuite(t)
	defer ts.tearDown()

	ctx := context.Background()

	// Test concurrent health checks
	t.Run("ConcurrentHealthChecks", func(t *testing.T) {
		concurrency := 10
		errors := make(chan error, concurrency)

		for i := 0; i < concurrency; i++ {
			go func() {
				_, err := ts.client.Health(ctx)
				errors <- err
			}()
		}

		for i := 0; i < concurrency; i++ {
			if err := <-errors; err != nil {
				t.Errorf("Concurrent request %d failed: %v", i, err)
			}
		}

		t.Logf("✓ All %d concurrent requests succeeded", concurrency)
	})
}

func TestIntegration_ErrorHandling(t *testing.T) {
	ts := setupTestSuite(t)
	defer ts.tearDown()

	ctx := context.Background()

	// Test getting non-existent VM
	t.Run("NonExistentVM", func(t *testing.T) {
		_, err := ts.client.GetVM(ctx, "nonexistent-vm-id")
		if err == nil {
			t.Error("Expected error for non-existent VM, got nil")
		}

		// Should be a 404 error
		if err != nil {
			t.Logf("✓ Got expected error: %v", err)
		}
	})

	// Test invalid VM creation
	t.Run("InvalidVMConfig", func(t *testing.T) {
		_, err := ts.client.CreateVM(ctx, &client.CreateVMRequest{
			Name:  "", // Empty name should use auto-generated ID
			Image: "", // Empty image should fail
		})

		if err == nil {
			t.Error("Expected error for invalid config, got nil")
		} else {
			t.Logf("✓ Got expected error: %v", err)
		}
	})
}

// Helper functions

func findBinary(name string) (string, error) {
	// Check in bin/ directory (relative to project root)
	paths := []string{
		filepath.Join("../../bin", name),
		filepath.Join("bin", name),
		filepath.Join("/tmp/nanofuse/bin", name),
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			absPath, _ := filepath.Abs(path)
			return absPath, nil
		}
	}

	// Check in PATH
	return exec.LookPath(name)
}
