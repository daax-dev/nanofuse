package firecracker

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/daax-dev/nanofuse/internal/types"
)

type snapshotAPICall struct {
	method string
	path   string
	body   []byte
}

func startUnixSnapshotAPIServer(t *testing.T, handler http.Handler) (string, <-chan snapshotAPICall) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("/tmp", "nf-fc-*")
	if err != nil {
		t.Fatalf("create short temp dir for unix socket: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(tmpDir)
	})

	socketPath := filepath.Join(tmpDir, "firecracker.sock")
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("listen on unix socket: %v", err)
	}

	calls := make(chan snapshotAPICall, 1)
	recordingHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioReadAll(r)
		if err != nil {
			t.Errorf("read request body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		calls <- snapshotAPICall{
			method: r.Method,
			path:   r.URL.Path,
			body:   body,
		}
		handler.ServeHTTP(w, r)
	})

	server := httptest.NewUnstartedServer(recordingHandler)
	if err := server.Listener.Close(); err != nil {
		t.Fatalf("close default listener: %v", err)
	}
	server.Listener = listener
	server.Start()
	t.Cleanup(server.Close)

	return socketPath, calls
}

func ioReadAll(r *http.Request) ([]byte, error) {
	defer r.Body.Close()
	return io.ReadAll(r.Body)
}

func snapshotTestVM(socketPath string) *types.VM {
	return &types.VM{
		ID:    "test-vm",
		State: types.StateRunning,
		Runtime: &types.VMRuntime{
			PID:        12345,
			SocketPath: socketPath,
		},
	}
}

func TestCreateSnapshotSendsFirecrackerRequest(t *testing.T) {
	socketPath, calls := startUnixSnapshotAPIServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	tmpDir := t.TempDir()
	snapshotPath := filepath.Join(tmpDir, "snapshots", "vm.snap")
	memPath := filepath.Join(tmpDir, "memory", "mem.snap")

	manager := NewManager("/usr/bin/firecracker", tmpDir)
	if err := manager.CreateSnapshot(snapshotTestVM(socketPath), snapshotPath, memPath); err != nil {
		t.Fatalf("CreateSnapshot: %v", err)
	}

	call := <-calls
	if call.method != http.MethodPut {
		t.Fatalf("method = %s, want %s", call.method, http.MethodPut)
	}
	if call.path != "/snapshot/create" {
		t.Fatalf("path = %s, want /snapshot/create", call.path)
	}

	var got snapshotCreateRequest
	if err := json.Unmarshal(call.body, &got); err != nil {
		t.Fatalf("decode request: %v", err)
	}
	if got.SnapshotType != "Full" {
		t.Fatalf("snapshot_type = %q, want Full", got.SnapshotType)
	}
	if got.SnapshotPath != snapshotPath {
		t.Fatalf("snapshot_path = %q, want %q", got.SnapshotPath, snapshotPath)
	}
	if got.MemFilePath != memPath {
		t.Fatalf("mem_file_path = %q, want %q", got.MemFilePath, memPath)
	}

	if _, err := os.Stat(filepath.Dir(snapshotPath)); err != nil {
		t.Fatalf("snapshot directory not created: %v", err)
	}
	if _, err := os.Stat(filepath.Dir(memPath)); err != nil {
		t.Fatalf("memory snapshot directory not created: %v", err)
	}
}

func TestCreateSnapshotNoRuntime(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager("/usr/bin/firecracker", tmpDir)
	vm := &types.VM{ID: "test-vm", State: types.StateStopped}

	err := manager.CreateSnapshot(vm, filepath.Join(tmpDir, "vm.snap"), filepath.Join(tmpDir, "mem.snap"))
	if err == nil {
		t.Fatal("expected error for VM without runtime")
	}
}

func TestCreateSnapshotAPIError(t *testing.T) {
	socketPath, _ := startUnixSnapshotAPIServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "snapshot failed", http.StatusInternalServerError)
	}))

	tmpDir := t.TempDir()
	manager := NewManager("/usr/bin/firecracker", tmpDir)

	err := manager.CreateSnapshot(snapshotTestVM(socketPath), filepath.Join(tmpDir, "vm.snap"), filepath.Join(tmpDir, "mem.snap"))
	if err == nil {
		t.Fatal("expected error from Firecracker API")
	}
}

func TestPauseSendsFirecrackerVMStateRequest(t *testing.T) {
	socketPath, calls := startUnixSnapshotAPIServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	manager := NewManager("/usr/bin/firecracker", t.TempDir())
	if err := manager.Pause(snapshotTestVM(socketPath)); err != nil {
		t.Fatalf("Pause: %v", err)
	}

	call := <-calls
	if call.method != http.MethodPatch {
		t.Fatalf("method = %s, want %s", call.method, http.MethodPatch)
	}
	if call.path != "/vm" {
		t.Fatalf("path = %s, want /vm", call.path)
	}

	var got vmStateRequest
	if err := json.Unmarshal(call.body, &got); err != nil {
		t.Fatalf("decode request: %v", err)
	}
	if got.State != "Paused" {
		t.Fatalf("state = %q, want Paused", got.State)
	}
}

func TestResumeSendsFirecrackerVMStateRequest(t *testing.T) {
	socketPath, calls := startUnixSnapshotAPIServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	manager := NewManager("/usr/bin/firecracker", t.TempDir())
	if err := manager.Resume(snapshotTestVM(socketPath)); err != nil {
		t.Fatalf("Resume: %v", err)
	}

	call := <-calls
	if call.method != http.MethodPatch {
		t.Fatalf("method = %s, want %s", call.method, http.MethodPatch)
	}
	if call.path != "/vm" {
		t.Fatalf("path = %s, want /vm", call.path)
	}

	var got vmStateRequest
	if err := json.Unmarshal(call.body, &got); err != nil {
		t.Fatalf("decode request: %v", err)
	}
	if got.State != "Resumed" {
		t.Fatalf("state = %q, want Resumed", got.State)
	}
}

func TestSendSnapshotLoadSendsFirecrackerRequest(t *testing.T) {
	socketPath, calls := startUnixSnapshotAPIServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	snapshotPath := "/data/snapshots/vm-1/snap/vm.snap"
	memPath := "/data/snapshots/vm-1/snap/mem.snap"

	if err := sendSnapshotLoad(socketPath, snapshotPath, memPath); err != nil {
		t.Fatalf("sendSnapshotLoad: %v", err)
	}

	call := <-calls
	if call.method != http.MethodPut {
		t.Fatalf("method = %s, want %s", call.method, http.MethodPut)
	}
	if call.path != "/snapshot/load" {
		t.Fatalf("path = %s, want /snapshot/load", call.path)
	}

	var got snapshotLoadRequest
	if err := json.Unmarshal(call.body, &got); err != nil {
		t.Fatalf("decode request: %v", err)
	}
	if got.SnapshotPath != snapshotPath {
		t.Fatalf("snapshot_path = %q, want %q", got.SnapshotPath, snapshotPath)
	}
	if got.MemBackend.BackendType != "File" {
		t.Fatalf("mem_backend.backend_type = %q, want File", got.MemBackend.BackendType)
	}
	if got.MemBackend.BackendPath != memPath {
		t.Fatalf("mem_backend.backend_path = %q, want %q", got.MemBackend.BackendPath, memPath)
	}
	if !got.ResumeVM {
		t.Fatalf("resume_vm = false, want true")
	}

	// The deprecated mem_file_path field must not be present alongside mem_backend.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(call.body, &raw); err != nil {
		t.Fatalf("decode raw request: %v", err)
	}
	if _, ok := raw["mem_file_path"]; ok {
		t.Fatalf("request must not include deprecated mem_file_path when mem_backend is set")
	}
}

func TestSendSnapshotLoadAPIError(t *testing.T) {
	socketPath, _ := startUnixSnapshotAPIServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "load failed", http.StatusBadRequest)
	}))

	if err := sendSnapshotLoad(socketPath, "/x/vm.snap", "/x/mem.snap"); err == nil {
		t.Fatal("expected error from Firecracker API")
	}
}

func TestLoadSnapshotValidatesInputs(t *testing.T) {
	manager := NewManager("/usr/bin/firecracker", t.TempDir())
	vm := &types.VM{ID: "test-vm", State: types.StateStopped}

	if err := manager.LoadSnapshot(nil, "/x/vm.snap", "/x/mem.snap"); err == nil {
		t.Fatal("expected error for nil VM")
	}
	if err := manager.LoadSnapshot(vm, "", "/x/mem.snap"); err == nil {
		t.Fatal("expected error for empty snapshot path")
	}
	if err := manager.LoadSnapshot(vm, "/x/vm.snap", ""); err == nil {
		t.Fatal("expected error for empty memory path")
	}
}

func TestLoadSnapshotMissingBackingFiles(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager("/usr/bin/firecracker", tmpDir)
	vm := &types.VM{ID: "test-vm", State: types.StateStopped}

	// Neither file exists; the load must fail before any process is spawned.
	err := manager.LoadSnapshot(vm, filepath.Join(tmpDir, "vm.snap"), filepath.Join(tmpDir, "mem.snap"))
	if err == nil {
		t.Fatal("expected error for missing snapshot files")
	}
	if !strings.Contains(err.Error(), "not accessible") {
		t.Fatalf("error = %v, want mention of inaccessible file", err)
	}
	if vm.Runtime != nil {
		t.Fatalf("VM runtime must remain unset when load fails early")
	}
}

func TestWaitForSocketReady(t *testing.T) {
	tmpDir, err := os.MkdirTemp("/tmp", "nf-fc-wait-*")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	socketPath := filepath.Join(tmpDir, "ready.sock")

	// Missing socket: must time out quickly.
	if err := waitForSocketReady(socketPath, 150*time.Millisecond); err == nil {
		t.Fatal("expected timeout for missing socket")
	}

	// Listening socket: must return promptly with no error.
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	if err := waitForSocketReady(socketPath, 2*time.Second); err != nil {
		t.Fatalf("waitForSocketReady on live socket: %v", err)
	}
}
