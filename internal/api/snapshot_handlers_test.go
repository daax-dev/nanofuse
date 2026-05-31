package api

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

	"github.com/daax-dev/nanofuse/internal/firecracker"
	"github.com/daax-dev/nanofuse/internal/logging"
	"github.com/daax-dev/nanofuse/internal/storage"
	"github.com/daax-dev/nanofuse/internal/types"
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
		db:        db,
		fcManager: firecracker.NewManager("/usr/bin/firecracker", t.TempDir()),
		logger:    logger,
		startTime: time.Now(),
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
