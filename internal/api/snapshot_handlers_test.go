package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/daax-dev/nanofuse/internal/config"
	"github.com/daax-dev/nanofuse/internal/firecracker"
	"github.com/daax-dev/nanofuse/internal/logging"
	"github.com/daax-dev/nanofuse/internal/storage"
	"github.com/daax-dev/nanofuse/internal/types"
	"github.com/daax-dev/nanofuse/internal/vmm"
)

type vmAPICall struct {
	method string
	path   string
	body   []byte
}

func startUnixVMAPIServer(t *testing.T) (string, <-chan vmAPICall) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("/tmp", "nf-api-fc-*")
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

	calls := make(chan vmAPICall, 1)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		calls <- vmAPICall{
			method: r.Method,
			path:   r.URL.Path,
			body:   body,
		}
		w.WriteHeader(http.StatusNoContent)
	})

	server := httptest.NewUnstartedServer(handler)
	if err := server.Listener.Close(); err != nil {
		t.Fatalf("close default listener: %v", err)
	}
	server.Listener = listener
	server.Start()
	t.Cleanup(server.Close)

	return socketPath, calls
}

func newSnapshotHandlerTestServer(t *testing.T, db *storage.DB) *Server {
	t.Helper()

	logger, err := logging.New(logging.Config{Level: "error"})
	if err != nil {
		t.Fatalf("logging.New: %v", err)
	}

	return &Server{
		db:             db,
		runtimeManager: firecracker.NewManager("/usr/bin/firecracker", t.TempDir()),
		logger:         logger,
		startTime:      time.Now(),
	}
}

func createSnapshotHandlerTestVM(t *testing.T, db *storage.DB, state types.VMState, socketPath string) *types.VM {
	t.Helper()

	now := time.Now()
	vm := &types.VM{
		ID:           "vm-" + string(state),
		Name:         "test-" + string(state),
		State:        state,
		Image:        "docker.io/library/alpine:latest",
		ImageDigest:  "sha256:test",
		Architecture: "x86_64",
		Config: types.VMConfig{
			VCPUs:     1,
			MemoryMiB: 128,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	if socketPath != "" {
		vm.Runtime = &types.VMRuntime{
			PID:        12345,
			SocketPath: socketPath,
		}
	}
	if err := db.CreateVM(vm); err != nil {
		t.Fatalf("CreateVM: %v", err)
	}
	return vm
}

func TestHandleCreateSnapshotRequiresPausedVM(t *testing.T) {
	states := []types.VMState{
		types.StateCreated,
		types.StateRunning,
		types.StateStopped,
	}

	for _, state := range states {
		t.Run(string(state), func(t *testing.T) {
			db, err := storage.New(filepath.Join(t.TempDir(), "nanofuse.db"))
			if err != nil {
				t.Fatalf("storage.New: %v", err)
			}
			defer db.Close()

			vm := createSnapshotHandlerTestVM(t, db, state, "")

			server := &Server{db: db}
			req := httptest.NewRequest(http.MethodPost, "/vms/"+vm.ID+"/snapshots", nil)
			w := httptest.NewRecorder()

			server.handleCreateSnapshot(w, req, vm.ID)

			if w.Code != http.StatusConflict {
				t.Fatalf("status = %d, want %d", w.Code, http.StatusConflict)
			}

			var response types.APIError
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("decode error response: %v", err)
			}
			if response.Error.Code != types.ErrInvalidStateTransition {
				t.Fatalf("error code = %s, want %s", response.Error.Code, types.ErrInvalidStateTransition)
			}
			if !strings.Contains(response.Error.Message, "pause the VM first") {
				t.Fatalf("error message = %q, want pause guidance", response.Error.Message)
			}
			if response.Error.Details["current_state"] != string(state) {
				t.Fatalf("current_state = %v, want %s", response.Error.Details["current_state"], state)
			}
		})
	}
}

func TestHandleCreateSnapshotMapsUnsupportedRuntimeTo501(t *testing.T) {
	db, err := storage.New(filepath.Join(t.TempDir(), "nanofuse.db"))
	if err != nil {
		t.Fatalf("storage.New: %v", err)
	}
	defer db.Close()

	vm := createSnapshotHandlerTestVM(t, db, types.StatePaused, "")
	logger, err := logging.New(logging.Config{Level: "error"})
	if err != nil {
		t.Fatalf("logging.New: %v", err)
	}
	server := &Server{
		db: db,
		config: &config.Config{
			Storage: config.StorageConfig{DataDir: t.TempDir()},
		},
		logger: logger,
		runtimeManager: &runtimeImageProviderStub{
			snapshotErr: fmt.Errorf("%w: snapshots are not supported", vmm.ErrUnsupportedOperation),
		},
	}
	req := httptest.NewRequest(http.MethodPost, "/vms/"+vm.ID+"/snapshots", nil)
	w := httptest.NewRecorder()

	server.handleCreateSnapshot(w, req, vm.ID)

	if w.Code != http.StatusNotImplemented {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusNotImplemented, w.Body.String())
	}
	var response types.APIError
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if response.Error.Code != types.ErrUnsupportedOperation {
		t.Fatalf("error code = %s, want %s", response.Error.Code, types.ErrUnsupportedOperation)
	}
	if response.Error.Details["operation"] != "VM snapshots" {
		t.Fatalf("operation detail = %v, want VM snapshots", response.Error.Details["operation"])
	}
}

func TestHandleVMPauseCallsFirecrackerAndUpdatesState(t *testing.T) {
	socketPath, calls := startUnixVMAPIServer(t)
	db, err := storage.New(filepath.Join(t.TempDir(), "nanofuse.db"))
	if err != nil {
		t.Fatalf("storage.New: %v", err)
	}
	defer db.Close()

	vm := createSnapshotHandlerTestVM(t, db, types.StateRunning, socketPath)
	server := newSnapshotHandlerTestServer(t, db)
	req := httptest.NewRequest(http.MethodPost, "/vms/"+vm.ID+"/pause", nil)
	req.SetPathValue("id", vm.ID)
	w := httptest.NewRecorder()

	server.handleVMPauseByPath(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	call := <-calls
	if call.method != http.MethodPatch {
		t.Fatalf("method = %s, want %s", call.method, http.MethodPatch)
	}
	if call.path != "/vm" {
		t.Fatalf("path = %s, want /vm", call.path)
	}
	var requestBody struct {
		State string `json:"state"`
	}
	if err := json.Unmarshal(call.body, &requestBody); err != nil {
		t.Fatalf("decode Firecracker request: %v", err)
	}
	if requestBody.State != "Paused" {
		t.Fatalf("state request = %q, want Paused", requestBody.State)
	}

	updated, err := db.GetVM(vm.ID)
	if err != nil {
		t.Fatalf("GetVM: %v", err)
	}
	if updated.State != types.StatePaused {
		t.Fatalf("VM state = %s, want %s", updated.State, types.StatePaused)
	}
}

func TestHandleVMPauseMapsUnsupportedRuntimeTo501(t *testing.T) {
	db, err := storage.New(filepath.Join(t.TempDir(), "nanofuse.db"))
	if err != nil {
		t.Fatalf("storage.New: %v", err)
	}
	defer db.Close()

	vm := createSnapshotHandlerTestVM(t, db, types.StateRunning, "")
	logger, err := logging.New(logging.Config{Level: "error"})
	if err != nil {
		t.Fatalf("logging.New: %v", err)
	}
	server := &Server{
		db:     db,
		logger: logger,
		runtimeManager: &runtimeImageProviderStub{
			pauseErr: fmt.Errorf("%w: pause is not supported", vmm.ErrUnsupportedOperation),
		},
	}
	req := httptest.NewRequest(http.MethodPost, "/vms/"+vm.ID+"/pause", nil)
	req.SetPathValue("id", vm.ID)
	w := httptest.NewRecorder()

	server.handleVMPauseByPath(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusNotImplemented, w.Body.String())
	}
	var response types.APIError
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if response.Error.Code != types.ErrUnsupportedOperation {
		t.Fatalf("error code = %s, want %s", response.Error.Code, types.ErrUnsupportedOperation)
	}
	if response.Error.Details["operation"] != "VM pause" {
		t.Fatalf("operation detail = %v, want VM pause", response.Error.Details["operation"])
	}

	updated, err := db.GetVM(vm.ID)
	if err != nil {
		t.Fatalf("GetVM: %v", err)
	}
	if updated.State != types.StateRunning {
		t.Fatalf("VM state = %s, want %s", updated.State, types.StateRunning)
	}
}

func TestHandleVMResumeCallsFirecrackerAndUpdatesState(t *testing.T) {
	socketPath, calls := startUnixVMAPIServer(t)
	db, err := storage.New(filepath.Join(t.TempDir(), "nanofuse.db"))
	if err != nil {
		t.Fatalf("storage.New: %v", err)
	}
	defer db.Close()

	vm := createSnapshotHandlerTestVM(t, db, types.StatePaused, socketPath)
	server := newSnapshotHandlerTestServer(t, db)
	req := httptest.NewRequest(http.MethodPost, "/vms/"+vm.ID+"/resume", nil)
	req.SetPathValue("id", vm.ID)
	w := httptest.NewRecorder()

	server.handleVMResumeByPath(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	call := <-calls
	if call.method != http.MethodPatch {
		t.Fatalf("method = %s, want %s", call.method, http.MethodPatch)
	}
	if call.path != "/vm" {
		t.Fatalf("path = %s, want /vm", call.path)
	}
	var requestBody struct {
		State string `json:"state"`
	}
	if err := json.Unmarshal(call.body, &requestBody); err != nil {
		t.Fatalf("decode Firecracker request: %v", err)
	}
	if requestBody.State != "Resumed" {
		t.Fatalf("state request = %q, want Resumed", requestBody.State)
	}

	updated, err := db.GetVM(vm.ID)
	if err != nil {
		t.Fatalf("GetVM: %v", err)
	}
	if updated.State != types.StateRunning {
		t.Fatalf("VM state = %s, want %s", updated.State, types.StateRunning)
	}
}

// snapshotDirFor returns the managed snapshot directory for a VM/snapshot pair,
// mirroring the layout the create handler uses ({DataDir}/snapshots/{vm}/{id}).
func snapshotDirFor(s *Server, vmID, id string) string {
	return filepath.Join(s.config.Storage.DataDir, "snapshots", vmID, id)
}

// writeSnapshotRecord creates a snapshot DB record with real backing files under
// the server's managed snapshots root.
func writeSnapshotRecord(t *testing.T, s *Server, id, vmID string) *types.Snapshot {
	t.Helper()
	dir := snapshotDirFor(s, vmID, id)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatalf("mkdir snapshot dir: %v", err)
	}
	snapPath := filepath.Join(dir, "vm.snap")
	memPath := filepath.Join(dir, "mem.snap")
	if err := os.WriteFile(snapPath, []byte("state"), 0o600); err != nil {
		t.Fatalf("write snap file: %v", err)
	}
	if err := os.WriteFile(memPath, []byte("memory"), 0o600); err != nil {
		t.Fatalf("write mem file: %v", err)
	}
	snap := &types.Snapshot{
		ID:               id,
		VMID:             vmID,
		Name:             id,
		SnapshotFilePath: snapPath,
		MemoryFilePath:   memPath,
		CreatedAt:        time.Now(),
	}
	if err := s.db.CreateSnapshot(snap); err != nil {
		t.Fatalf("CreateSnapshot: %v", err)
	}
	return snap
}

func newSnapshotResumeServer(t *testing.T, db *storage.DB, stub *runtimeImageProviderStub) *Server {
	t.Helper()
	logger, err := logging.New(logging.Config{Level: "error"})
	if err != nil {
		t.Fatalf("logging.New: %v", err)
	}
	return &Server{
		db:             db,
		logger:         logger,
		runtimeManager: stub,
		startTime:      time.Now(),
		config:         &config.Config{Storage: config.StorageConfig{DataDir: t.TempDir()}},
	}
}

func resumeFromSnapshotRequest(t *testing.T, vmID, snapshotID string) *http.Request {
	t.Helper()
	body, err := json.Marshal(types.ResumeVMRequest{SnapshotID: &snapshotID})
	if err != nil {
		t.Fatalf("marshal resume request: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/vms/"+vmID+"/resume", strings.NewReader(string(body)))
	req.SetPathValue("id", vmID)
	return req
}

func TestHandleVMResumeFromSnapshotLoadsAndUpdatesState(t *testing.T) {
	db, err := storage.New(filepath.Join(t.TempDir(), "nanofuse.db"))
	if err != nil {
		t.Fatalf("storage.New: %v", err)
	}
	defer db.Close()

	vm := createSnapshotHandlerTestVM(t, db, types.StateStopped, "")
	stub := &runtimeImageProviderStub{}
	server := newSnapshotResumeServer(t, db, stub)
	snap := writeSnapshotRecord(t, server, "snapshot-load-1", vm.ID)

	w := httptest.NewRecorder()
	server.handleVMResumeByPath(w, resumeFromSnapshotRequest(t, vm.ID, snap.ID))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
	if stub.loadSnapshotCalls != 1 {
		t.Fatalf("LoadSnapshot calls = %d, want 1", stub.loadSnapshotCalls)
	}
	if stub.loadSnapshotSnapPath != snap.SnapshotFilePath {
		t.Fatalf("snapshot path = %q, want %q", stub.loadSnapshotSnapPath, snap.SnapshotFilePath)
	}
	if stub.loadSnapshotMemPath != snap.MemoryFilePath {
		t.Fatalf("mem path = %q, want %q", stub.loadSnapshotMemPath, snap.MemoryFilePath)
	}

	updated, err := db.GetVM(vm.ID)
	if err != nil {
		t.Fatalf("GetVM: %v", err)
	}
	if updated.State != types.StateRunning {
		t.Fatalf("VM state = %s, want %s", updated.State, types.StateRunning)
	}
}

func TestHandleVMResumeFromSnapshotRejectsRunningVM(t *testing.T) {
	db, err := storage.New(filepath.Join(t.TempDir(), "nanofuse.db"))
	if err != nil {
		t.Fatalf("storage.New: %v", err)
	}
	defer db.Close()

	vm := createSnapshotHandlerTestVM(t, db, types.StateRunning, "")
	stub := &runtimeImageProviderStub{}
	server := newSnapshotResumeServer(t, db, stub)
	snap := writeSnapshotRecord(t, server, "snapshot-load-2", vm.ID)

	w := httptest.NewRecorder()
	server.handleVMResumeByPath(w, resumeFromSnapshotRequest(t, vm.ID, snap.ID))

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusConflict, w.Body.String())
	}
	if stub.loadSnapshotCalls != 0 {
		t.Fatalf("LoadSnapshot must not be called for a running VM")
	}
	var response types.APIError
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if response.Error.Code != types.ErrInvalidStateTransition {
		t.Fatalf("error code = %s, want %s", response.Error.Code, types.ErrInvalidStateTransition)
	}
}

func TestHandleVMResumeFromSnapshotRejectsLiveRuntimeStates(t *testing.T) {
	// Snapshot resume must be limited to states with no live runtime. Paused and
	// Stopping both still hold (or may still hold) a live Firecracker process, so
	// loading a snapshot into a fresh process would orphan it.
	for _, state := range []types.VMState{types.StatePaused, types.StateStopping} {
		t.Run(string(state), func(t *testing.T) {
			db, err := storage.New(filepath.Join(t.TempDir(), "nanofuse.db"))
			if err != nil {
				t.Fatalf("storage.New: %v", err)
			}
			defer db.Close()

			vm := createSnapshotHandlerTestVM(t, db, state, "")
			stub := &runtimeImageProviderStub{}
			server := newSnapshotResumeServer(t, db, stub)
			snap := writeSnapshotRecord(t, server, "snapshot-load-"+string(state), vm.ID)

			w := httptest.NewRecorder()
			server.handleVMResumeByPath(w, resumeFromSnapshotRequest(t, vm.ID, snap.ID))

			if w.Code != http.StatusConflict {
				t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusConflict, w.Body.String())
			}
			if stub.loadSnapshotCalls != 0 {
				t.Fatalf("LoadSnapshot must not be called from state %s", state)
			}
		})
	}
}

func TestHandleVMResumeFromSnapshotNotFound(t *testing.T) {
	db, err := storage.New(filepath.Join(t.TempDir(), "nanofuse.db"))
	if err != nil {
		t.Fatalf("storage.New: %v", err)
	}
	defer db.Close()

	vm := createSnapshotHandlerTestVM(t, db, types.StateStopped, "")
	stub := &runtimeImageProviderStub{}
	server := newSnapshotResumeServer(t, db, stub)

	w := httptest.NewRecorder()
	server.handleVMResumeByPath(w, resumeFromSnapshotRequest(t, vm.ID, "does-not-exist"))

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusNotFound, w.Body.String())
	}
	if stub.loadSnapshotCalls != 0 {
		t.Fatalf("LoadSnapshot must not be called when snapshot is missing")
	}
	updated, _ := db.GetVM(vm.ID)
	if updated.State != types.StateStopped {
		t.Fatalf("VM state = %s, want unchanged %s", updated.State, types.StateStopped)
	}
}

func TestHandleVMResumeFromSnapshotWrongOwner(t *testing.T) {
	db, err := storage.New(filepath.Join(t.TempDir(), "nanofuse.db"))
	if err != nil {
		t.Fatalf("storage.New: %v", err)
	}
	defer db.Close()

	vm := createSnapshotHandlerTestVM(t, db, types.StateStopped, "")
	// The snapshot belongs to a different, existing VM.
	otherVM := &types.VM{
		ID:           "vm-other-owner",
		Name:         "other-owner",
		State:        types.StateStopped,
		Image:        "docker.io/library/alpine:latest",
		ImageDigest:  "sha256:test",
		Architecture: "x86_64",
		Config:       types.VMConfig{VCPUs: 1, MemoryMiB: 128},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := db.CreateVM(otherVM); err != nil {
		t.Fatalf("CreateVM(other): %v", err)
	}
	stub := &runtimeImageProviderStub{}
	server := newSnapshotResumeServer(t, db, stub)
	snap := writeSnapshotRecord(t, server, "snapshot-load-3", otherVM.ID)

	w := httptest.NewRecorder()
	server.handleVMResumeByPath(w, resumeFromSnapshotRequest(t, vm.ID, snap.ID))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusBadRequest, w.Body.String())
	}
	if stub.loadSnapshotCalls != 0 {
		t.Fatalf("LoadSnapshot must not be called for a mismatched owner")
	}
	var response types.APIError
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if response.Error.Code != types.ErrInvalidRequest {
		t.Fatalf("error code = %s, want %s", response.Error.Code, types.ErrInvalidRequest)
	}
}

func TestHandleVMResumeFromSnapshotMissingFiles(t *testing.T) {
	db, err := storage.New(filepath.Join(t.TempDir(), "nanofuse.db"))
	if err != nil {
		t.Fatalf("storage.New: %v", err)
	}
	defer db.Close()

	vm := createSnapshotHandlerTestVM(t, db, types.StateStopped, "")
	stub := &runtimeImageProviderStub{}
	server := newSnapshotResumeServer(t, db, stub)
	// Record points at non-existent files WITHIN the managed snapshots root so
	// the missing-file check (not the path-traversal guard) is exercised.
	dir := snapshotDirFor(server, vm.ID, "snapshot-load-4")
	snap := &types.Snapshot{
		ID:               "snapshot-load-4",
		VMID:             vm.ID,
		SnapshotFilePath: filepath.Join(dir, "missing-vm.snap"),
		MemoryFilePath:   filepath.Join(dir, "missing-mem.snap"),
		CreatedAt:        time.Now(),
	}
	if err := db.CreateSnapshot(snap); err != nil {
		t.Fatalf("CreateSnapshot: %v", err)
	}

	w := httptest.NewRecorder()
	server.handleVMResumeByPath(w, resumeFromSnapshotRequest(t, vm.ID, snap.ID))

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusNotFound, w.Body.String())
	}
	if stub.loadSnapshotCalls != 0 {
		t.Fatalf("LoadSnapshot must not be called when backing files are missing")
	}
}

func TestHandleVMResumeFromSnapshotRejectsPathOutsideRoot(t *testing.T) {
	db, err := storage.New(filepath.Join(t.TempDir(), "nanofuse.db"))
	if err != nil {
		t.Fatalf("storage.New: %v", err)
	}
	defer db.Close()

	vm := createSnapshotHandlerTestVM(t, db, types.StateStopped, "")
	stub := &runtimeImageProviderStub{}
	server := newSnapshotResumeServer(t, db, stub)

	// Craft real files OUTSIDE the managed snapshots root; the guard must refuse
	// them even though they exist.
	outside := t.TempDir()
	snapPath := filepath.Join(outside, "vm.snap")
	memPath := filepath.Join(outside, "mem.snap")
	if err := os.WriteFile(snapPath, []byte("x"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := os.WriteFile(memPath, []byte("x"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	snap := &types.Snapshot{
		ID:               "snapshot-outside",
		VMID:             vm.ID,
		SnapshotFilePath: snapPath,
		MemoryFilePath:   memPath,
		CreatedAt:        time.Now(),
	}
	if err := db.CreateSnapshot(snap); err != nil {
		t.Fatalf("CreateSnapshot: %v", err)
	}

	w := httptest.NewRecorder()
	server.handleVMResumeByPath(w, resumeFromSnapshotRequest(t, vm.ID, snap.ID))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusInternalServerError, w.Body.String())
	}
	if stub.loadSnapshotCalls != 0 {
		t.Fatalf("LoadSnapshot must not be called for an out-of-root snapshot path")
	}
}

func TestHandleVMResumeFromSnapshotUnsupportedRuntime(t *testing.T) {
	db, err := storage.New(filepath.Join(t.TempDir(), "nanofuse.db"))
	if err != nil {
		t.Fatalf("storage.New: %v", err)
	}
	defer db.Close()

	vm := createSnapshotHandlerTestVM(t, db, types.StateStopped, "")
	stub := &runtimeImageProviderStub{
		loadSnapshotErr: fmt.Errorf("%w: snapshot resume is not supported", vmm.ErrUnsupportedOperation),
	}
	server := newSnapshotResumeServer(t, db, stub)
	snap := writeSnapshotRecord(t, server, "snapshot-load-5", vm.ID)

	w := httptest.NewRecorder()
	server.handleVMResumeByPath(w, resumeFromSnapshotRequest(t, vm.ID, snap.ID))

	if w.Code != http.StatusNotImplemented {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusNotImplemented, w.Body.String())
	}
	var response types.APIError
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if response.Error.Code != types.ErrUnsupportedOperation {
		t.Fatalf("error code = %s, want %s", response.Error.Code, types.ErrUnsupportedOperation)
	}
	if response.Error.Details["operation"] != "VM snapshot resume" {
		t.Fatalf("operation detail = %v, want VM snapshot resume", response.Error.Details["operation"])
	}
	// State must be restored to the pre-resume value on failure.
	updated, _ := db.GetVM(vm.ID)
	if updated.State != types.StateStopped {
		t.Fatalf("VM state = %s, want restored %s", updated.State, types.StateStopped)
	}
}

func TestHandleVMResumeMapsUnsupportedRuntimeTo501(t *testing.T) {
	db, err := storage.New(filepath.Join(t.TempDir(), "nanofuse.db"))
	if err != nil {
		t.Fatalf("storage.New: %v", err)
	}
	defer db.Close()

	vm := createSnapshotHandlerTestVM(t, db, types.StatePaused, "")
	logger, err := logging.New(logging.Config{Level: "error"})
	if err != nil {
		t.Fatalf("logging.New: %v", err)
	}
	server := &Server{
		db:     db,
		logger: logger,
		runtimeManager: &runtimeImageProviderStub{
			resumeErr: fmt.Errorf("%w: resume is not supported", vmm.ErrUnsupportedOperation),
		},
	}
	req := httptest.NewRequest(http.MethodPost, "/vms/"+vm.ID+"/resume", nil)
	req.SetPathValue("id", vm.ID)
	w := httptest.NewRecorder()

	server.handleVMResumeByPath(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusNotImplemented, w.Body.String())
	}
	var response types.APIError
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if response.Error.Code != types.ErrUnsupportedOperation {
		t.Fatalf("error code = %s, want %s", response.Error.Code, types.ErrUnsupportedOperation)
	}
	if response.Error.Details["operation"] != "VM resume" {
		t.Fatalf("operation detail = %v, want VM resume", response.Error.Details["operation"])
	}

	updated, err := db.GetVM(vm.ID)
	if err != nil {
		t.Fatalf("GetVM: %v", err)
	}
	if updated.State != types.StatePaused {
		t.Fatalf("VM state = %s, want %s", updated.State, types.StatePaused)
	}
}
