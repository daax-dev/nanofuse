package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/daax-dev/nanofuse/internal/config"
	"github.com/daax-dev/nanofuse/internal/logging"
	"github.com/daax-dev/nanofuse/internal/storage"
	"github.com/daax-dev/nanofuse/internal/types"
)

// spireRegistrarStub injects SPIRE registration success/failure into the API
// handlers so the fail-closed enforcement path can be exercised without a live
// SPIRE deployment (the concrete *spire.Service shells out to `docker exec`).
type spireRegistrarStub struct {
	enabled      bool
	spiffeID     string
	createErr    error
	deleteCalls  int
	lastDeleteID string
}

func (s *spireRegistrarStub) IsEnabled() bool { return s.enabled }

func (s *spireRegistrarStub) CreateVMWorkloadEntry(_ context.Context, _, _, _ string) (string, error) {
	if s.createErr != nil {
		return "", s.createErr
	}
	return s.spiffeID, nil
}

func (s *spireRegistrarStub) DeleteVMWorkloadEntry(_ context.Context, spiffeID string) error {
	s.deleteCalls++
	s.lastDeleteID = spiffeID
	return nil
}

func newSPIRETestServer(t *testing.T, db *storage.DB, spireCfg config.SPIREConfig, spireSvc spireRegistrar) *Server {
	t.Helper()
	logger, err := logging.New(logging.Config{Level: "error"})
	if err != nil {
		t.Fatalf("logging.New: %v", err)
	}
	// Provide a real source rootfs so materializeWritableRootDisks provisions a
	// per-VM storage directory before the SPIRE gate; this lets the fail-closed
	// test prove cleanup actually reclaimed it.
	rootfsSrc := filepath.Join(t.TempDir(), "rootfs.ext4")
	if writeErr := os.WriteFile(rootfsSrc, []byte("fake-rootfs"), 0600); writeErr != nil {
		t.Fatalf("write source rootfs: %v", writeErr)
	}
	return &Server{
		config: &config.Config{
			Storage: config.StorageConfig{DataDir: t.TempDir()},
			Runtime: config.RuntimeConfig{Driver: "apple_container"},
			Network: config.NetworkConfig{Setup: false},
			Limits: config.LimitsConfig{
				MaxVMs:            25,
				MaxVCPUsPerVM:     8,
				MaxMemoryPerVMMiB: 8192,
			},
			SPIRE: spireCfg,
		},
		db:     db,
		logger: logger,
		runtimeManager: &runtimeImageProviderStub{
			image: &types.Image{
				Digest:       "sha256:test",
				Tags:         []string{"docker.io/library/alpine:3.20"},
				Architecture: "arm64",
				RootfsPath:   rootfsSrc,
				PulledAt:     time.Now(),
			},
		},
		spireService: spireSvc,
	}
}

func newSPIRETestDB(t *testing.T) *storage.DB {
	t.Helper()
	db, err := storage.New(filepath.Join(t.TempDir(), "nanofuse.db"))
	if err != nil {
		t.Fatalf("storage.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

const spireCreateBody = `{"name":"nf-spire","image":"docker.io/library/alpine:3.20","owner_user_id":"alice","group_id":"team"}`

// AC-1 (DoD AC4): Required + SPIRE unreachable -> create fails 503 naming SPIRE,
// no VM persisted and no partial resource leak.
func TestHandleCreateVMSpireRequiredUnreachableFailsClosed(t *testing.T) {
	db := newSPIRETestDB(t)
	spireSvc := &spireRegistrarStub{enabled: true, createErr: errors.New("dial spire agent: connection refused")}
	server := newSPIRETestServer(t, db,
		config.SPIREConfig{Enabled: true, Required: true}, spireSvc)

	req := httptest.NewRequest(http.MethodPost, "/vms", strings.NewReader(spireCreateBody))
	rec := httptest.NewRecorder()
	server.handleCreateVM(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusServiceUnavailable, rec.Body.String())
	}
	var apiErr types.APIError
	if err := json.NewDecoder(rec.Body).Decode(&apiErr); err != nil {
		t.Fatalf("decode API error: %v", err)
	}
	if apiErr.Error.Code != types.ErrServiceUnavailable {
		t.Fatalf("error code = %s, want %s", apiErr.Error.Code, types.ErrServiceUnavailable)
	}
	if !strings.Contains(apiErr.Error.Message, "SPIRE") {
		t.Fatalf("error message = %q, want it to name SPIRE", apiErr.Error.Message)
	}

	// No leaked VM record.
	vms, err := db.ListVMs("")
	if err != nil {
		t.Fatalf("ListVMs: %v", err)
	}
	if len(vms) != 0 {
		t.Fatalf("VMs persisted = %d, want 0 (no leaked VM on fail-closed)", len(vms))
	}
	// Nothing was registered, so nothing should be unregistered.
	if spireSvc.deleteCalls != 0 {
		t.Fatalf("DeleteVMWorkloadEntry called %d times, want 0 (nothing was registered)", spireSvc.deleteCalls)
	}
	// Prove cleanupCreatedVMResources actually reclaimed provisioned storage:
	// with Network.Setup=false, the per-VM writable rootfs under <DataDir>/vms
	// is the only resource materialized before the SPIRE gate, so no residual
	// per-VM storage directory may remain. (TAP/IPAM/egress cleanup use the same
	// helper but require root+real networking and are not exercised in unit tests.)
	vmsDir := filepath.Join(server.config.Storage.DataDir, "vms")
	entries, err := os.ReadDir(vmsDir)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("ReadDir(%s): %v", vmsDir, err)
	}
	if len(entries) != 0 {
		t.Fatalf("leaked %d per-VM storage dir(s) under %s after fail-closed; cleanup incomplete", len(entries), vmsDir)
	}
}

// AC-2: Required + SPIRE reachable -> create succeeds and VM carries the SVID.
func TestHandleCreateVMSpireRequiredReachableSucceeds(t *testing.T) {
	db := newSPIRETestDB(t)
	const wantSpiffe = "spiffe://poley.dev/g/team/u/alice/microvm/x"
	spireSvc := &spireRegistrarStub{enabled: true, spiffeID: wantSpiffe}
	server := newSPIRETestServer(t, db,
		config.SPIREConfig{Enabled: true, Required: true}, spireSvc)

	req := httptest.NewRequest(http.MethodPost, "/vms", strings.NewReader(spireCreateBody))
	rec := httptest.NewRecorder()
	server.handleCreateVM(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var vm types.VM
	if err := json.NewDecoder(rec.Body).Decode(&vm); err != nil {
		t.Fatalf("decode VM: %v", err)
	}
	if vm.SpiffeID != wantSpiffe {
		t.Fatalf("vm.SpiffeID = %q, want %q", vm.SpiffeID, wantSpiffe)
	}
}

// AC-3: not Required + SPIRE unreachable -> create still succeeds (WARN only),
// no identity attached.
func TestHandleCreateVMSpireNotRequiredUnreachableSucceeds(t *testing.T) {
	db := newSPIRETestDB(t)
	spireSvc := &spireRegistrarStub{enabled: true, createErr: errors.New("dial spire agent: connection refused")}
	server := newSPIRETestServer(t, db,
		config.SPIREConfig{Enabled: true, Required: false}, spireSvc)

	req := httptest.NewRequest(http.MethodPost, "/vms", strings.NewReader(spireCreateBody))
	rec := httptest.NewRecorder()
	server.handleCreateVM(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var vm types.VM
	if err := json.NewDecoder(rec.Body).Decode(&vm); err != nil {
		t.Fatalf("decode VM: %v", err)
	}
	if vm.SpiffeID != "" {
		t.Fatalf("vm.SpiffeID = %q, want empty (best-effort failure attaches no identity)", vm.SpiffeID)
	}
	vms, err := db.ListVMs("")
	if err != nil {
		t.Fatalf("ListVMs: %v", err)
	}
	if len(vms) != 1 {
		t.Fatalf("VMs persisted = %d, want 1 (best-effort proceeds)", len(vms))
	}
	// Positive control for the fail-closed cleanup assertion: a successful create
	// materializes a per-VM storage directory under <DataDir>/vms. This proves the
	// fail-closed test's "0 entries" check is meaningful (cleanup removed a real dir).
	vmDir := vmStorageDir(server.config.Storage.DataDir, vm.ID)
	if _, statErr := os.Stat(vmDir); statErr != nil {
		t.Fatalf("expected per-VM storage dir %s to exist after successful create: %v", vmDir, statErr)
	}
}

// spireRequired must be false unless SPIRE is both enabled and required, so a
// Required flag alone (SPIRE disabled) never triggers enforcement.
func TestSpireRequiredGate(t *testing.T) {
	cases := []struct {
		enabled  bool
		required bool
		want     bool
	}{
		{false, false, false},
		{true, false, false},
		{false, true, false},
		{true, true, true},
	}
	for _, tc := range cases {
		s := &Server{config: &config.Config{SPIRE: config.SPIREConfig{Enabled: tc.enabled, Required: tc.required}}}
		if got := s.spireRequired(); got != tc.want {
			t.Errorf("spireRequired(enabled=%v, required=%v) = %v, want %v", tc.enabled, tc.required, got, tc.want)
		}
	}
}

// AC-4: the fail-closed setting defaults to disabled.
func TestDefaultConfigSpireRequiredDefaultsOff(t *testing.T) {
	cfg := config.DefaultConfig()
	if cfg.SPIRE.Required {
		t.Fatalf("DefaultConfig().SPIRE.Required = true, want false (fail-closed must be opt-in)")
	}
}

// Guard the stub against interface drift.
var _ spireRegistrar = (*spireRegistrarStub)(nil)
