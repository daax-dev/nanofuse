package firecracker

import (
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/jpoley/nanofuse/internal/types"
)

func TestCreateSnapshot(t *testing.T) {
	t.Skip("Skipping: CreateSnapshot not yet implemented")

	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "nanofuse-snapshot-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a mock Firecracker API server
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/snapshot/create" && r.Method == http.MethodPut {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	// Create Unix socket server
	socketPath := filepath.Join(tmpDir, "test.sock")
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("Failed to create Unix socket: %v", err)
	}
	defer listener.Close()

	// Start test server
	server := httptest.NewUnstartedServer(handler)
	server.Listener.Close()
	server.Listener = listener
	server.Start()
	defer server.Close()

	// Create test VM with runtime info
	vm := &types.VM{
		ID:    "test-vm",
		State: types.StateRunning,
		Runtime: &types.VMRuntime{
			PID:        12345,
			SocketPath: socketPath,
		},
	}

	// Create manager
	manager := NewManager("/usr/bin/firecracker", tmpDir)

	// Test snapshot creation
	snapshotPath := filepath.Join(tmpDir, "vm.snap")
	memPath := filepath.Join(tmpDir, "mem.snap")

	err = manager.CreateSnapshot(vm, snapshotPath, memPath)
	if err != nil {
		t.Errorf("CreateSnapshot failed: %v", err)
	}

	// Verify directories were created
	if _, err := os.Stat(filepath.Dir(snapshotPath)); os.IsNotExist(err) {
		t.Error("Snapshot directory was not created")
	}
	if _, err := os.Stat(filepath.Dir(memPath)); os.IsNotExist(err) {
		t.Error("Memory file directory was not created")
	}
}

func TestCreateSnapshotNoRuntime(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "nanofuse-snapshot-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	vm := &types.VM{
		ID:      "test-vm",
		State:   types.StateStopped,
		Runtime: nil,
	}

	manager := NewManager("/usr/bin/firecracker", tmpDir)

	snapshotPath := filepath.Join(tmpDir, "vm.snap")
	memPath := filepath.Join(tmpDir, "mem.snap")

	err = manager.CreateSnapshot(vm, snapshotPath, memPath)
	if err == nil {
		t.Error("Expected error for VM without runtime, got nil")
	}
}

func TestCreateSnapshotAPIError(t *testing.T) {
	t.Skip("Skipping: CreateSnapshot not yet implemented")

	tmpDir, err := os.MkdirTemp("", "nanofuse-snapshot-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a mock Firecracker API server that returns an error
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
	})

	socketPath := filepath.Join(tmpDir, "test.sock")
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("Failed to create Unix socket: %v", err)
	}
	defer listener.Close()

	server := httptest.NewUnstartedServer(handler)
	server.Listener.Close()
	server.Listener = listener
	server.Start()
	defer server.Close()

	vm := &types.VM{
		ID:    "test-vm",
		State: types.StateRunning,
		Runtime: &types.VMRuntime{
			PID:        12345,
			SocketPath: socketPath,
		},
	}

	manager := NewManager("/usr/bin/firecracker", tmpDir)

	snapshotPath := filepath.Join(tmpDir, "vm.snap")
	memPath := filepath.Join(tmpDir, "mem.snap")

	err = manager.CreateSnapshot(vm, snapshotPath, memPath)
	if err == nil {
		t.Error("Expected error from Firecracker API, got nil")
	}
}
