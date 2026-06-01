package api

import (
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

func TestMaterializeWritableRootDisksCopiesRootfsPerVM(t *testing.T) {
	dataDir := t.TempDir()
	imageDir := t.TempDir()
	sourceRootfs := filepath.Join(imageDir, "source-rootfs.ext4")
	if err := os.WriteFile(sourceRootfs, []byte("source-rootfs"), 0600); err != nil {
		t.Fatalf("write source rootfs: %v", err)
	}

	cfg := types.VMConfig{
		Disks: []types.DiskConfig{
			{
				DriveID:      "rootfs",
				PathOnHost:   sourceRootfs,
				IsReadOnly:   false,
				IsRootDevice: true,
			},
		},
	}

	if err := materializeWritableRootDisks(dataDir, "vm-123", &cfg); err != nil {
		t.Fatalf("materialize rootfs: %v", err)
	}

	wantPath := filepath.Join(dataDir, "vms", "vm-123", "rootfs.ext4")
	if cfg.Disks[0].PathOnHost != wantPath {
		t.Fatalf("root disk path = %q, want %q", cfg.Disks[0].PathOnHost, wantPath)
	}

	got, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("read VM rootfs: %v", err)
	}
	if string(got) != "source-rootfs" {
		t.Fatalf("VM rootfs contents = %q, want source copy", got)
	}

	source, err := os.ReadFile(sourceRootfs)
	if err != nil {
		t.Fatalf("read source rootfs: %v", err)
	}
	if string(source) != "source-rootfs" {
		t.Fatalf("source rootfs mutated: %q", source)
	}
}

func TestMaterializeWritableRootDisksPreservesExistingVMDisk(t *testing.T) {
	dataDir := t.TempDir()
	imageDir := t.TempDir()
	sourceRootfs := filepath.Join(imageDir, "source-rootfs.ext4")
	if err := os.WriteFile(sourceRootfs, []byte("source-rootfs"), 0600); err != nil {
		t.Fatalf("write source rootfs: %v", err)
	}

	existingRootfs := vmRootfsPath(dataDir, "vm-123")
	if err := os.MkdirAll(filepath.Dir(existingRootfs), 0700); err != nil {
		t.Fatalf("create VM storage: %v", err)
	}
	if err := os.WriteFile(existingRootfs, []byte("persisted-state"), 0600); err != nil {
		t.Fatalf("write existing rootfs: %v", err)
	}

	cfg := types.VMConfig{
		Disks: []types.DiskConfig{
			{
				DriveID:      "rootfs",
				PathOnHost:   sourceRootfs,
				IsReadOnly:   false,
				IsRootDevice: true,
			},
		},
	}

	if err := materializeWritableRootDisks(dataDir, "vm-123", &cfg); err != nil {
		t.Fatalf("materialize rootfs: %v", err)
	}

	got, err := os.ReadFile(existingRootfs)
	if err != nil {
		t.Fatalf("read existing rootfs: %v", err)
	}
	if string(got) != "persisted-state" {
		t.Fatalf("existing VM rootfs overwritten: %q", got)
	}
}

func TestMaterializeWritableRootDisksSkipsReadOnlyRootfs(t *testing.T) {
	dataDir := t.TempDir()
	sourceRootfs := filepath.Join(t.TempDir(), "source-rootfs.ext4")
	if err := os.WriteFile(sourceRootfs, []byte("source-rootfs"), 0600); err != nil {
		t.Fatalf("write source rootfs: %v", err)
	}

	cfg := types.VMConfig{
		Disks: []types.DiskConfig{
			{
				DriveID:      "rootfs",
				PathOnHost:   sourceRootfs,
				IsReadOnly:   true,
				IsRootDevice: true,
			},
		},
	}

	if err := materializeWritableRootDisks(dataDir, "vm-123", &cfg); err != nil {
		t.Fatalf("materialize rootfs: %v", err)
	}
	if cfg.Disks[0].PathOnHost != sourceRootfs {
		t.Fatalf("read-only rootfs path changed to %q", cfg.Disks[0].PathOnHost)
	}
	if _, err := os.Stat(vmRootfsPath(dataDir, "vm-123")); !os.IsNotExist(err) {
		t.Fatalf("read-only rootfs copy exists or stat failed: %v", err)
	}
}

func TestMaterializeWritableRootDisksContinuesAfterExistingDestinationPath(t *testing.T) {
	dataDir := t.TempDir()
	sourceRootfs := filepath.Join(t.TempDir(), "source-rootfs.ext4")
	if err := os.WriteFile(sourceRootfs, []byte("source-rootfs"), 0600); err != nil {
		t.Fatalf("write source rootfs: %v", err)
	}

	existingDestination := vmRootfsPath(dataDir, "vm-123")
	cfg := types.VMConfig{
		Disks: []types.DiskConfig{
			{
				DriveID:      "rootfs-existing",
				PathOnHost:   existingDestination,
				IsReadOnly:   false,
				IsRootDevice: true,
			},
			{
				DriveID:      "rootfs-source",
				PathOnHost:   sourceRootfs,
				IsReadOnly:   false,
				IsRootDevice: true,
			},
		},
	}

	if err := materializeWritableRootDisks(dataDir, "vm-123", &cfg); err != nil {
		t.Fatalf("materialize rootfs: %v", err)
	}

	got, err := os.ReadFile(existingDestination)
	if err != nil {
		t.Fatalf("read VM rootfs: %v", err)
	}
	if string(got) != "source-rootfs" {
		t.Fatalf("VM rootfs contents = %q, want source copy", got)
	}
	if cfg.Disks[1].PathOnHost != existingDestination {
		t.Fatalf("second root disk path = %q, want %q", cfg.Disks[1].PathOnHost, existingDestination)
	}
}

func TestCleanupVMStorageRemovesVMDirectory(t *testing.T) {
	dataDir := t.TempDir()
	vmDir := vmStorageDir(dataDir, "vm-123")
	if err := os.MkdirAll(vmDir, 0700); err != nil {
		t.Fatalf("create VM dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(vmDir, "rootfs.ext4"), []byte("state"), 0600); err != nil {
		t.Fatalf("write VM state: %v", err)
	}

	if err := cleanupVMStorage(dataDir, "vm-123"); err != nil {
		t.Fatalf("cleanup VM storage: %v", err)
	}
	if _, err := os.Stat(vmDir); !os.IsNotExist(err) {
		t.Fatalf("VM storage still exists or stat failed: %v", err)
	}
}

func TestSetupVMNetworkingAllowsNoneWhenNetworkSetupDisabled(t *testing.T) {
	server := &Server{
		config: &config.Config{
			Runtime: config.RuntimeConfig{Driver: "firecracker"},
			Network: config.NetworkConfig{Setup: false},
		},
	}
	vmConfig := types.VMConfig{
		Network: types.NetworkConfig{Mode: "none"},
	}

	if err := server.setupVMNetworking("vm-123", &vmConfig); err != nil {
		t.Fatalf("setupVMNetworking() error = %v", err)
	}
}

func TestSetupVMNetworkingRejectsManagedModeWhenNetworkSetupDisabled(t *testing.T) {
	server := &Server{
		config: &config.Config{
			Runtime: config.RuntimeConfig{Driver: "firecracker"},
			Network: config.NetworkConfig{Setup: false},
		},
	}
	vmConfig := types.VMConfig{
		Network: types.NetworkConfig{Mode: "nat"},
	}

	err := server.setupVMNetworking("vm-123", &vmConfig)
	if err == nil {
		t.Fatal("expected setupVMNetworking() to reject managed networking")
	}
	if !errors.Is(err, errNetworkSetupDisabled) {
		t.Fatalf("setupVMNetworking() error = %q", err)
	}
}

func TestSetupVMNetworkingRejectsNoneForRuntimeManagedNetworking(t *testing.T) {
	server := &Server{
		config: &config.Config{
			Runtime: config.RuntimeConfig{Driver: "apple_container"},
			Network: config.NetworkConfig{Setup: false},
		},
	}
	vmConfig := types.VMConfig{
		Network: types.NetworkConfig{Mode: "none"},
	}

	err := server.setupVMNetworking("vm-123", &vmConfig)
	if err == nil {
		t.Fatal("expected setupVMNetworking() to reject network none for runtime-managed networking")
	}
	if !errors.Is(err, errRuntimeNetworkModeUnsupported) {
		t.Fatalf("setupVMNetworking() error = %q", err)
	}
}

func TestWriteNetworkSetupErrorUsesInvalidConfig(t *testing.T) {
	rec := httptest.NewRecorder()

	if !writeNetworkSetupError(rec, errNetworkSetupDisabled, "nat") {
		t.Fatal("expected network setup error to be handled")
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var apiErr types.APIError
	if err := json.NewDecoder(rec.Body).Decode(&apiErr); err != nil {
		t.Fatalf("decode API error: %v", err)
	}
	if apiErr.Error.Code != types.ErrInvalidConfig {
		t.Fatalf("error code = %s, want %s", apiErr.Error.Code, types.ErrInvalidConfig)
	}
	if apiErr.Error.Details["network_mode"] != "nat" {
		t.Fatalf("network_mode detail = %v, want nat", apiErr.Error.Details["network_mode"])
	}
	if apiErr.Error.Details["network_setup"] != false {
		t.Fatalf("network_setup detail = %v, want false", apiErr.Error.Details["network_setup"])
	}
}

func TestWriteNetworkSetupErrorUsesInvalidConfigForRuntimeManagedNetworkNone(t *testing.T) {
	rec := httptest.NewRecorder()

	if !writeNetworkSetupError(rec, errRuntimeNetworkModeUnsupported, "none") {
		t.Fatal("expected runtime network mode error to be handled")
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var apiErr types.APIError
	if err := json.NewDecoder(rec.Body).Decode(&apiErr); err != nil {
		t.Fatalf("decode API error: %v", err)
	}
	if apiErr.Error.Code != types.ErrInvalidConfig {
		t.Fatalf("error code = %s, want %s", apiErr.Error.Code, types.ErrInvalidConfig)
	}
	if apiErr.Error.Details["network_mode"] != "none" {
		t.Fatalf("network_mode detail = %v, want none", apiErr.Error.Details["network_mode"])
	}
	if apiErr.Error.Details["allowed_network_mode"] != "nat" {
		t.Fatalf("allowed_network_mode detail = %v, want nat", apiErr.Error.Details["allowed_network_mode"])
	}
}

func TestWriteNetworkSetupErrorIgnoresOtherErrors(t *testing.T) {
	rec := httptest.NewRecorder()

	if writeNetworkSetupError(rec, errors.New("other network failure"), "nat") {
		t.Fatal("expected unrelated error to remain unhandled")
	}
}

func TestValidateAndResolveImageDoesNotTreatRuntimeProviderErrorAsImageMissing(t *testing.T) {
	db, err := storage.New(filepath.Join(t.TempDir(), "nanofuse.db"))
	if err != nil {
		t.Fatalf("storage.New: %v", err)
	}
	defer db.Close()

	providerErr := errors.New("apple container binary not found")
	server := &Server{
		db: db,
		runtimeManager: &runtimeImageProviderStub{
			err: providerErr,
		},
	}

	_, _, err = server.validateAndResolveImage("alpine:3.20")
	if err == nil {
		t.Fatal("expected runtime provider error")
	}
	var missing *imageNotFoundError
	if errors.As(err, &missing) {
		t.Fatalf("provider error was classified as image missing: %v", err)
	}
	if !errors.Is(err, providerErr) {
		t.Fatalf("provider error was not preserved: %v", err)
	}
	if !strings.Contains(err.Error(), "runtime image resolution failed") {
		t.Fatalf("error = %q, want runtime image resolution failure", err.Error())
	}
}

func TestValidateAndResolveImageMissingImageUsesTypedStableError(t *testing.T) {
	db, err := storage.New(filepath.Join(t.TempDir(), "nanofuse.db"))
	if err != nil {
		t.Fatalf("storage.New: %v", err)
	}
	defer db.Close()

	server := &Server{db: db}

	_, _, err = server.validateAndResolveImage("missing:not-found-in-error-text")
	if err == nil {
		t.Fatal("expected image missing error")
	}
	var missing *imageNotFoundError
	if !errors.As(err, &missing) {
		t.Fatalf("error = %T %v, want imageNotFoundError", err, err)
	}
	if err.Error() != "image not found" {
		t.Fatalf("error message = %q, want stable image not found", err.Error())
	}
	if missing.imageRef != "missing:not-found-in-error-text" {
		t.Fatalf("image ref = %q, want missing:not-found-in-error-text", missing.imageRef)
	}
}

func TestHandleCreateVMRuntimeProviderNotFoundTextReturnsInternalError(t *testing.T) {
	db, err := storage.New(filepath.Join(t.TempDir(), "nanofuse.db"))
	if err != nil {
		t.Fatalf("storage.New: %v", err)
	}
	defer db.Close()

	logger, err := logging.New(logging.Config{Level: "error"})
	if err != nil {
		t.Fatalf("logging.New: %v", err)
	}
	server := &Server{
		db:     db,
		logger: logger,
		runtimeManager: &runtimeImageProviderStub{
			err: errors.New("apple container binary not found"),
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/vms", strings.NewReader(`{"image":"alpine:3.20"}`))
	rec := httptest.NewRecorder()
	server.handleCreateVM(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
	var apiErr types.APIError
	if err := json.NewDecoder(rec.Body).Decode(&apiErr); err != nil {
		t.Fatalf("decode API error: %v", err)
	}
	if apiErr.Error.Code != types.ErrInternalError {
		t.Fatalf("error code = %s, want %s", apiErr.Error.Code, types.ErrInternalError)
	}
}

func TestHandleCreateVMMissingImageReturnsImageNotFound(t *testing.T) {
	db, err := storage.New(filepath.Join(t.TempDir(), "nanofuse.db"))
	if err != nil {
		t.Fatalf("storage.New: %v", err)
	}
	defer db.Close()

	logger, err := logging.New(logging.Config{Level: "error"})
	if err != nil {
		t.Fatalf("logging.New: %v", err)
	}
	server := &Server{
		db:     db,
		logger: logger,
	}

	req := httptest.NewRequest(http.MethodPost, "/vms", strings.NewReader(`{"image":"missing:latest"}`))
	rec := httptest.NewRecorder()
	server.handleCreateVM(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
	var apiErr types.APIError
	if err := json.NewDecoder(rec.Body).Decode(&apiErr); err != nil {
		t.Fatalf("decode API error: %v", err)
	}
	if apiErr.Error.Code != types.ErrImageNotFound {
		t.Fatalf("error code = %s, want %s", apiErr.Error.Code, types.ErrImageNotFound)
	}
}

func TestVMHasRuntimeHandleAcceptsPIDOrExternalID(t *testing.T) {
	tests := []struct {
		name string
		vm   *types.VM
		want bool
	}{
		{name: "nil VM", vm: nil, want: false},
		{name: "nil runtime", vm: &types.VM{}, want: false},
		{
			name: "firecracker PID",
			vm: &types.VM{Runtime: &types.VMRuntime{
				PID: 1234,
			}},
			want: true,
		},
		{
			name: "apple container external ID",
			vm: &types.VM{Runtime: &types.VMRuntime{
				ExternalID: "nf-abc123",
			}},
			want: true,
		},
		{
			name: "empty runtime",
			vm:   &types.VM{Runtime: &types.VMRuntime{}},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := vmHasRuntimeHandle(tt.vm)
			if got != tt.want {
				t.Fatalf("vmHasRuntimeHandle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandleDeleteVMKeepsMetadataWhenRuntimeDeleteFails(t *testing.T) {
	db, err := storage.New(filepath.Join(t.TempDir(), "nanofuse.db"))
	if err != nil {
		t.Fatalf("storage.New: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("db.Close: %v", err)
		}
	}()

	logger, err := logging.New(logging.Config{Level: "error"})
	if err != nil {
		t.Fatalf("logging.New: %v", err)
	}

	runtimeManager := &runtimeImageProviderStub{deleteErr: errors.New("runtime delete failed")}
	server := &Server{
		db:             db,
		logger:         logger,
		runtimeManager: runtimeManager,
	}
	vm := &types.VM{
		ID:           "vm-delete-runtime-fail",
		Name:         "delete-runtime-fail",
		State:        types.StateStopped,
		Image:        "docker.io/library/alpine:3.20",
		ImageDigest:  "sha256:test",
		Architecture: "arm64",
		Config: types.VMConfig{
			VCPUs:     1,
			MemoryMiB: 256,
			Network:   types.NetworkConfig{Mode: "nat"},
		},
		Runtime:   &types.VMRuntime{Driver: "apple_container", ExternalID: "nf-test"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := db.CreateVM(vm); err != nil {
		t.Fatalf("CreateVM: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/vms/"+vm.ID, nil)
	rec := httptest.NewRecorder()
	server.handleDeleteVM(rec, req, vm.ID)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
	if runtimeManager.deleteCalls != 1 {
		t.Fatalf("runtime delete calls = %d, want 1", runtimeManager.deleteCalls)
	}
	got, err := db.GetVM(vm.ID)
	if err != nil {
		t.Fatalf("GetVM: %v", err)
	}
	if got == nil {
		t.Fatal("VM metadata was deleted after runtime cleanup failure")
	}
}

func TestHandleDeleteVMWithoutRuntimeHandleSkipsKillAndDeletes(t *testing.T) {
	db, err := storage.New(filepath.Join(t.TempDir(), "nanofuse.db"))
	if err != nil {
		t.Fatalf("storage.New: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("db.Close: %v", err)
		}
	}()

	logger, err := logging.New(logging.Config{Level: "error"})
	if err != nil {
		t.Fatalf("logging.New: %v", err)
	}

	runtimeManager := &runtimeImageProviderStub{}
	server := &Server{
		config:         &config.Config{Storage: config.StorageConfig{DataDir: t.TempDir()}},
		db:             db,
		logger:         logger,
		runtimeManager: runtimeManager,
	}
	vm := &types.VM{
		ID:           "vm-delete-no-runtime-handle",
		Name:         "delete-no-runtime-handle",
		State:        types.StateRunning,
		Image:        "docker.io/library/alpine:3.20",
		ImageDigest:  "sha256:test",
		Architecture: "arm64",
		Config: types.VMConfig{
			VCPUs:     1,
			MemoryMiB: 256,
			Network:   types.NetworkConfig{Mode: "nat"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := db.CreateVM(vm); err != nil {
		t.Fatalf("CreateVM: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/vms/"+vm.ID, nil)
	rec := httptest.NewRecorder()
	server.handleDeleteVM(rec, req, vm.ID)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
	if runtimeManager.killCalls != 0 {
		t.Fatalf("runtime kill calls = %d, want 0", runtimeManager.killCalls)
	}
	if runtimeManager.deleteCalls != 1 {
		t.Fatalf("runtime delete calls = %d, want 1", runtimeManager.deleteCalls)
	}
	got, err := db.GetVM(vm.ID)
	if err != nil {
		t.Fatalf("GetVM: %v", err)
	}
	if got != nil {
		t.Fatal("VM metadata was not deleted")
	}
}

func TestHandleVMExecRunsThroughRuntime(t *testing.T) {
	db, err := storage.New(filepath.Join(t.TempDir(), "nanofuse.db"))
	if err != nil {
		t.Fatalf("storage.New: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("db.Close: %v", err)
		}
	}()

	logger, err := logging.New(logging.Config{Level: "error"})
	if err != nil {
		t.Fatalf("logging.New: %v", err)
	}

	runtimeManager := &runtimeImageProviderStub{
		execResult: &types.VMExecResult{
			Command:   []string{"uname", "-a"},
			ExitCode:  0,
			Stdout:    "Linux test\n",
			RuntimeID: "nf-test",
		},
	}
	server := &Server{
		db:             db,
		logger:         logger,
		runtimeManager: runtimeManager,
	}
	vm := &types.VM{
		ID:           "vm-exec",
		Name:         "exec",
		State:        types.StateRunning,
		Image:        "docker.io/library/alpine:3.20",
		ImageDigest:  "sha256:test",
		Architecture: "arm64",
		Config: types.VMConfig{
			VCPUs:     1,
			MemoryMiB: 256,
			Network:   types.NetworkConfig{Mode: "nat"},
		},
		Runtime:   &types.VMRuntime{Driver: "apple_container", ExternalID: "nf-test"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := db.CreateVM(vm); err != nil {
		t.Fatalf("CreateVM: %v", err)
	}

	body := `{"command":["uname","-a"],"timeout_seconds":5}`
	req := httptest.NewRequest(http.MethodPost, "/vms/"+vm.ID+"/exec", strings.NewReader(body))
	req.SetPathValue("id", vm.ID)
	rec := httptest.NewRecorder()
	server.handleVMExecByPath(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if runtimeManager.execCalls != 1 {
		t.Fatalf("runtime exec calls = %d, want 1", runtimeManager.execCalls)
	}
	if strings.Join(runtimeManager.execCommand, " ") != "uname -a" {
		t.Fatalf("runtime exec command = %#v", runtimeManager.execCommand)
	}
	var result types.VMExecResult
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode result: %v", err)
	}
	if result.Stdout != "Linux test\n" {
		t.Fatalf("stdout = %q, want Linux test", result.Stdout)
	}
}
