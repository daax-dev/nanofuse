package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/daax-dev/nanofuse/internal/config"
	"github.com/daax-dev/nanofuse/internal/logging"
	"github.com/daax-dev/nanofuse/internal/snapshotstore"
	"github.com/daax-dev/nanofuse/internal/storage"
	"github.com/daax-dev/nanofuse/internal/types"
)

// snapshotWritingStub overrides CreateSnapshot to produce real files on disk so
// the object-storage tiering path has content to upload.
type snapshotWritingStub struct {
	*runtimeImageProviderStub
}

func (s *snapshotWritingStub) CreateSnapshot(_ *types.VM, snapPath, memPath string) error {
	if err := os.MkdirAll(filepath.Dir(snapPath), 0o750); err != nil {
		return err
	}
	if err := os.WriteFile(snapPath, []byte("vm-state-bytes"), 0o600); err != nil {
		return err
	}
	return os.WriteFile(memPath, []byte("memory-bytes-guest-ram"), 0o600)
}

func TestHandleCreateSnapshotTiersToObjectStore(t *testing.T) {
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

	tierRoot := t.TempDir()
	blob, err := snapshotstore.NewFSBlob(tierRoot)
	if err != nil {
		t.Fatalf("NewFSBlob: %v", err)
	}
	store := snapshotstore.NewTieredStore(blob, snapshotstore.Options{})

	server := &Server{
		db:              db,
		config:          &config.Config{Storage: config.StorageConfig{DataDir: t.TempDir()}},
		logger:          logger,
		runtimeManager:  &snapshotWritingStub{&runtimeImageProviderStub{}},
		snapshotStore:   store,
		snapshotRuntime: snapshotstore.RuntimeVersions{Firecracker: "v1.7.0", Nanofuse: "test"},
	}

	req := httptest.NewRequest(http.MethodPost, "/vms/"+vm.ID+"/snapshots", nil)
	w := httptest.NewRecorder()
	server.handleCreateSnapshot(w, req, vm.ID)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body: %s", w.Code, w.Body.String())
	}

	ctx := context.Background()
	ids, err := store.List(ctx)
	if err != nil {
		t.Fatalf("store.List: %v", err)
	}
	if len(ids) != 1 {
		t.Fatalf("tiered snapshots = %v, want exactly 1", ids)
	}
	if !strings.HasPrefix(ids[0], vm.ID+"__") {
		t.Fatalf("tiered id = %q, want prefix %q", ids[0], vm.ID+"__")
	}

	manifest, err := store.Manifest(ctx, ids[0])
	if err != nil {
		t.Fatalf("store.Manifest: %v", err)
	}
	if manifest.Runtime.Firecracker != "v1.7.0" {
		t.Errorf("manifest firecracker = %q, want v1.7.0", manifest.Runtime.Firecracker)
	}
	if len(manifest.Files) != 2 {
		t.Fatalf("manifest files = %d, want 2 (vm.snap + mem.snap)", len(manifest.Files))
	}

	// Restore the tiered snapshot and confirm byte-for-byte recovery.
	dest := t.TempDir()
	if _, err := store.Get(ctx, ids[0], dest); err != nil {
		t.Fatalf("store.Get: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dest, "vm.snap"))
	if err != nil {
		t.Fatalf("read restored vm.snap: %v", err)
	}
	if string(got) != "vm-state-bytes" {
		t.Fatalf("restored vm.snap = %q, want %q", got, "vm-state-bytes")
	}
}

func TestHandleCreateSnapshotWithoutStoreDoesNotTier(t *testing.T) {
	// Regression guard: with no store configured, the create path is unchanged.
	db, err := storage.New(filepath.Join(t.TempDir(), "nanofuse.db"))
	if err != nil {
		t.Fatalf("storage.New: %v", err)
	}
	defer db.Close()

	vm := createSnapshotHandlerTestVM(t, db, types.StatePaused, "")
	logger, _ := logging.New(logging.Config{Level: "error"})
	server := &Server{
		db:             db,
		config:         &config.Config{Storage: config.StorageConfig{DataDir: t.TempDir()}},
		logger:         logger,
		runtimeManager: &snapshotWritingStub{&runtimeImageProviderStub{}},
		// snapshotStore is nil.
	}

	req := httptest.NewRequest(http.MethodPost, "/vms/"+vm.ID+"/snapshots", nil)
	w := httptest.NewRecorder()
	server.handleCreateSnapshot(w, req, vm.ID)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body: %s", w.Code, w.Body.String())
	}
}
