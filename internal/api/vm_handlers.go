package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/daax-dev/nanofuse/internal/network"
	"github.com/daax-dev/nanofuse/internal/types"
	"github.com/daax-dev/nanofuse/internal/vmm"
	"github.com/google/uuid"
)

var (
	errNetworkSetupDisabled           = errors.New("network setup disabled; use network mode none or enable network.setup")
	errRuntimeEgressPolicyUnsupported = errors.New("runtime-managed networking does not support nanofuse egress_policy")
)

const (
	defaultExecTimeoutSeconds = 30
	maxExecTimeoutSeconds     = 600
)

// handleListVMs lists all VMs
func (s *Server) handleListVMs(w http.ResponseWriter, r *http.Request) {
	stateFilter := r.URL.Query().Get("state")

	vms, err := s.db.ListVMs(stateFilter)
	if err != nil {
		s.logger.Printf("ERROR: Failed to list VMs: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, "Failed to list VMs", nil)
		return
	}

	items := make([]types.VMListItem, 0, len(vms))
	for _, vm := range vms {
		item := types.VMListItem{
			ID:           vm.ID,
			Name:         vm.Name,
			State:        vm.State,
			Image:        vm.Image,
			ImageDigest:  vm.ImageDigest,
			Architecture: vm.Architecture,
			Config:       vm.Config,
			Runtime:      vm.Runtime,
			CreatedAt:    vm.CreatedAt,
		}

		// Calculate uptime if running
		if vm.State == types.StateRunning && vm.Runtime != nil {
			uptime := int(time.Since(vm.UpdatedAt).Seconds())
			item.UptimeSeconds = &uptime
		}

		items = append(items, item)
	}

	response := types.ListVMsResponse{
		VMs:   items,
		Total: len(items),
	}

	writeJSON(w, http.StatusOK, response)
}

// validateAndResolveImage validates image field and returns image from DB or runtime image provider.
func (s *Server) validateAndResolveImage(imageRef string) (*types.Image, string, error) {
	var image *types.Image
	var err error

	// If imageRef is a digest (sha256:...), look up by digest
	if strings.HasPrefix(imageRef, "sha256:") {
		image, err = s.db.GetImage(imageRef)
		if err != nil {
			return nil, "", fmt.Errorf("database error: %w", err)
		}
	} else {
		// Otherwise, it's a tag reference - look up by tag
		image, err = s.db.GetImageByTag(imageRef)
		if err != nil {
			return nil, "", fmt.Errorf("database error: %w", err)
		}
	}

	if image == nil {
		if provider, ok := s.runtimeManager.(vmm.ImageProvider); ok {
			image, err = provider.ResolveImage(imageRef)
			if err != nil {
				return nil, "", fmt.Errorf("image not found: %s: %w", imageRef, err)
			}
			if image != nil {
				if err := s.db.UpsertImage(image); err != nil {
					return nil, "", fmt.Errorf("database error: %w", err)
				}
			}
		}
		if image == nil {
			return nil, "", fmt.Errorf("image not found: %s", imageRef)
		}
	}

	return image, image.Digest, nil
}

// buildVMConfig creates base config and applies user overrides
func buildVMConfig(image *types.Image, req *types.CreateVMRequest) types.VMConfig {
	config := types.VMConfig{
		VCPUs:      2,
		MemoryMiB:  512,
		KernelArgs: "console=ttyS0 root=/dev/vda rw init=/lib/systemd/systemd",
		Network: types.NetworkConfig{
			Mode: "nat",
		},
	}
	if image.RootfsPath != "" {
		config.Disks = []types.DiskConfig{
			{
				DriveID:      "rootfs",
				PathOnHost:   image.RootfsPath,
				IsReadOnly:   false,
				IsRootDevice: true,
			},
		}
	}

	// Apply user config overrides
	if req.Config != nil {
		if req.Config.VCPUs != nil {
			config.VCPUs = *req.Config.VCPUs
		}
		if req.Config.MemoryMiB != nil {
			config.MemoryMiB = *req.Config.MemoryMiB
		}
		if req.Config.KernelArgs != nil {
			config.KernelArgs = *req.Config.KernelArgs
		}
		if req.Config.SSHPublicKey != nil {
			config.SSHPublicKey = *req.Config.SSHPublicKey
		}
		if req.Config.Network != nil {
			if req.Config.Network.Mode != nil {
				config.Network.Mode = *req.Config.Network.Mode
			}
			if req.Config.Network.BridgeName != nil {
				config.Network.BridgeName = req.Config.Network.BridgeName
			}
			if req.Config.Network.MACAddress != nil {
				config.Network.MACAddress = *req.Config.Network.MACAddress
			}
			if req.Config.Network.PortForwards != nil {
				config.Network.PortForwards = *req.Config.Network.PortForwards
			}
			if req.Config.Network.EgressPolicy != nil {
				config.Network.EgressPolicy = req.Config.Network.EgressPolicy
			}
		}
	}

	return config
}

func vmStorageDir(dataDir, vmID string) string {
	return filepath.Join(dataDir, "vms", vmID)
}

func vmRootfsPath(dataDir, vmID string) string {
	return filepath.Join(vmStorageDir(dataDir, vmID), "rootfs.ext4")
}

func copyFileAtomic(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source rootfs: %w", err)
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0700); err != nil {
		return fmt.Errorf("failed to create VM storage directory: %w", err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(dst), ".rootfs-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temporary rootfs: %w", err)
	}
	tmpPath := tmp.Name()
	cleanupTmp := true
	defer func() {
		if cleanupTmp {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := io.Copy(tmp, in); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("failed to copy rootfs: %w", err)
	}
	if err := tmp.Chmod(mode); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("failed to chmod temporary rootfs: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("failed to sync temporary rootfs: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("failed to close temporary rootfs: %w", err)
	}

	if err := os.Rename(tmpPath, dst); err != nil {
		return fmt.Errorf("failed to install VM rootfs: %w", err)
	}
	cleanupTmp = false
	return nil
}

// materializeWritableRootDisks gives each VM its own writable root disk.
// Registered image rootfs files remain immutable sources; VM state persists in
// storage.data_dir/vms/<vm-id>/rootfs.ext4 until the VM is deleted.
func materializeWritableRootDisks(dataDir, vmID string, config *types.VMConfig) error {
	for i := range config.Disks {
		disk := &config.Disks[i]
		if !disk.IsRootDevice || disk.IsReadOnly {
			continue
		}

		dest := vmRootfsPath(dataDir, vmID)
		srcAbs, err := filepath.Abs(disk.PathOnHost)
		if err != nil {
			return fmt.Errorf("failed to resolve source rootfs path: %w", err)
		}
		destAbs, err := filepath.Abs(dest)
		if err != nil {
			return fmt.Errorf("failed to resolve VM rootfs path: %w", err)
		}

		if srcAbs == destAbs {
			continue
		}

		if _, err := os.Stat(dest); err == nil {
			disk.PathOnHost = dest
			continue
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("failed to inspect VM rootfs: %w", err)
		}

		if err := copyFileAtomic(disk.PathOnHost, dest, 0600); err != nil {
			return err
		}
		disk.PathOnHost = dest
	}

	return nil
}

func cleanupVMStorage(dataDir, vmID string) error {
	if err := os.RemoveAll(vmStorageDir(dataDir, vmID)); err != nil {
		return fmt.Errorf("failed to remove VM storage: %w", err)
	}
	return nil
}

func (s *Server) cleanupCreatedVMResources(vmID string, config types.VMConfig) {
	if cleanupErr := cleanupVMStorage(s.config.Storage.DataDir, vmID); cleanupErr != nil {
		s.logger.Printf("WARN: Failed to cleanup VM storage after create failure: %v", cleanupErr)
	}
	if config.Network.EgressPolicy != nil && config.Network.EgressPolicy.Enabled {
		if cleanupErr := network.CleanupEgressPolicy(vmID, config.Network.TapDevice, config.Network.IPAddress); cleanupErr != nil {
			s.logger.Printf("WARN: Failed to cleanup egress policy after create failure: %v", cleanupErr)
		}
	}
	if config.Network.TapDevice != "" {
		if cleanupErr := network.DeleteTAPDevice(config.Network.TapDevice); cleanupErr != nil {
			s.logger.Printf("WARN: Failed to cleanup TAP device after create failure: %v", cleanupErr)
		}
	}
	if config.Network.IPAddress != "" {
		s.ipam.ReleaseIP(vmID)
	}
}

func (s *Server) cleanupDeletedVMResources(vm *types.VM) {
	if len(vm.Config.Network.PortForwards) > 0 && vm.Config.Network.IPAddress != "" {
		s.logger.Printf("INFO: Cleaning up %d port forward(s) for deleted VM %s", len(vm.Config.Network.PortForwards), vm.Name)
		if err := network.CleanupPortForwards(vm.Config.Network.IPAddress, vm.Config.Network.PortForwards); err != nil {
			s.logger.Printf("WARN: Failed to cleanup port forwards: %v", err)
		}
	}

	if vm.Config.Network.EgressPolicy != nil && vm.Config.Network.EgressPolicy.Enabled {
		if err := network.CleanupEgressPolicy(vm.ID, vm.Config.Network.TapDevice, vm.Config.Network.IPAddress); err != nil {
			s.logger.Printf("WARN: Failed to cleanup egress policy: %v", err)
		}
	}

	if vm.Config.Network.TapDevice != "" {
		if err := network.DeleteTAPDevice(vm.Config.Network.TapDevice); err != nil {
			s.logger.Printf("WARN: Failed to delete TAP device %s: %v", vm.Config.Network.TapDevice, err)
		} else {
			s.logger.Printf("INFO: Deleted TAP device: %s", vm.Config.Network.TapDevice)
		}
	}

	if vm.Config.Network.IPAddress != "" {
		s.ipam.ReleaseIP(vm.ID)
		s.logger.Printf("INFO: Released IP address: %s", vm.Config.Network.IPAddress)
	}

	if err := cleanupVMStorage(s.config.Storage.DataDir, vm.ID); err != nil {
		s.logger.Printf("WARN: Failed to cleanup VM storage for deleted VM %s: %v", vm.ID, err)
	}
}

// validateVMResourceLimits checks if VM config exceeds limits
func (s *Server) validateVMResourceLimits(config types.VMConfig) error {
	if config.VCPUs > s.config.Limits.MaxVCPUsPerVM {
		return fmt.Errorf("vCPUs exceed limit: %d > %d", config.VCPUs, s.config.Limits.MaxVCPUsPerVM)
	}

	if config.MemoryMiB > s.config.Limits.MaxMemoryPerVMMiB {
		return fmt.Errorf("memory exceeds limit: %d > %d", config.MemoryMiB, s.config.Limits.MaxMemoryPerVMMiB)
	}

	// Check global VM count limit
	allVMs, err := s.db.ListVMs("")
	if err != nil {
		return fmt.Errorf("failed to check VM count: %w", err)
	}

	if len(allVMs) >= s.config.Limits.MaxVMs {
		return fmt.Errorf("maximum VM count exceeded: %d >= %d", len(allVMs), s.config.Limits.MaxVMs)
	}

	return nil
}

// setupVMNetworking configures networking for a VM
func (s *Server) setupVMNetworking(vmID string, config *types.VMConfig) error {
	if config.Network.Mode == "none" {
		return nil
	}
	if selectedRuntimeDriver(s.config) != "firecracker" {
		if config.Network.EgressPolicy != nil && config.Network.EgressPolicy.Enabled {
			return errRuntimeEgressPolicyUnsupported
		}
		return nil
	}
	if !s.config.Network.Setup {
		return errNetworkSetupDisabled
	}

	// Allocate IP address
	ip, err := s.ipam.AllocateIP(vmID)
	if err != nil {
		return fmt.Errorf("failed to allocate IP: %w", err)
	}

	// Create TAP device
	tapName := fmt.Sprintf("tap-%s", vmID[:8])
	if err := network.CreateTAPDevice(tapName); err != nil {
		s.ipam.ReleaseIP(vmID) // Cleanup
		return fmt.Errorf("failed to create TAP device: %w", err)
	}

	// Attach TAP to bridge
	if err := network.AttachTAPToBridge(tapName, network.BridgeName); err != nil {
		if err := network.DeleteTAPDevice(tapName); err != nil {
			s.logger.Printf("WARN: Failed to cleanup TAP device: %v", err)
		}
		s.ipam.ReleaseIP(vmID)
		return fmt.Errorf("failed to attach TAP to bridge: %w", err)
	}

	// Generate MAC address if not provided
	if config.Network.MACAddress == "" {
		mac, err := network.GenerateMAC()
		if err != nil {
			if err := network.DeleteTAPDevice(tapName); err != nil {
				s.logger.Printf("WARN: Failed to cleanup TAP device: %v", err)
			}
			s.ipam.ReleaseIP(vmID)
			return fmt.Errorf("failed to generate MAC: %w", err)
		}
		config.Network.MACAddress = mac
	}

	// Update network config
	config.Network.TapDevice = tapName
	config.Network.IPAddress = ip
	config.Network.Gateway = network.BridgeGateway
	config.Network.Netmask = "255.255.255.0"

	// Update kernel args to include IP configuration
	// Force classic interface naming so the kernel 'ip=' setting targets eth0 as expected
	// Otherwise Ubuntu's predictable names (en*) may prevent the static IP from applying
	config.KernelArgs = fmt.Sprintf(
		"console=ttyS0 root=/dev/vda rw init=/lib/systemd/systemd ip=%s::%s:255.255.255.0::eth0:off net.ifnames=0 biosdevname=0",
		ip, network.BridgeGateway,
	)

	s.logger.Printf("INFO: Configured network for VM %s: IP=%s TAP=%s MAC=%s",
		vmID[:8], ip, tapName, config.Network.MACAddress)

	if err := network.SetupEgressPolicy(vmID, tapName, ip, network.BridgeGateway, config.Network.EgressPolicy); err != nil {
		if err := network.DeleteTAPDevice(tapName); err != nil {
			s.logger.Printf("WARN: Failed to cleanup TAP device after egress setup failure: %v", err)
		}
		s.ipam.ReleaseIP(vmID)
		return fmt.Errorf("failed to setup egress policy: %w", err)
	}

	return nil
}

func writeNetworkSetupError(w http.ResponseWriter, err error, networkMode string) bool {
	if !errors.Is(err, errNetworkSetupDisabled) {
		if !errors.Is(err, errRuntimeEgressPolicyUnsupported) {
			return false
		}
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidConfig, errRuntimeEgressPolicyUnsupported.Error(), map[string]interface{}{
			"network_mode": networkMode,
		})
		return true
	}

	types.WriteError(w, http.StatusBadRequest, types.ErrInvalidConfig, errNetworkSetupDisabled.Error(), map[string]interface{}{
		"network_mode":         networkMode,
		"network_setup":        false,
		"allowed_network_mode": "none",
	})
	return true
}

// registerSPIREWorkload handles SPIRE workload registration for a new VM.
// Returns the spiffeID if registered, empty string otherwise, and any error encountered.
// Registration is best-effort and won't fail VM creation.
func (s *Server) registerSPIREWorkload(ctx context.Context, vmID string, req *types.CreateVMRequest) (string, error) {
	shouldRegister := req.OwnerUserID != "" && req.GroupID != ""
	if req.AutoRegisterSPIFFE != nil {
		shouldRegister = *req.AutoRegisterSPIFFE && shouldRegister
	}

	if !shouldRegister || s.spireService == nil || !s.spireService.IsEnabled() {
		return "", nil
	}

	spiffeID, err := s.spireService.CreateVMWorkloadEntry(ctx, vmID, req.GroupID, req.OwnerUserID)
	if err != nil {
		s.logger.Printf("WARN: Failed to auto-register SPIRE workload entry: %v", err)
		return "", fmt.Errorf("SPIRE registration failed: %w", err)
	}

	s.logger.Printf("INFO: Auto-registered SPIRE workload entry: %s", spiffeID)
	return spiffeID, nil
}

// handleCreateVM creates a new VM
func (s *Server) handleCreateVM(w http.ResponseWriter, r *http.Request) {
	var req types.CreateVMRequest
	if err := readJSON(r, &req); err != nil {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest, "Invalid request body", nil)
		return
	}

	// Validate required fields
	if req.Image == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest, "Image is required", nil)
		return
	}

	// Check if VM with name already exists (idempotent by name)
	if req.Name != "" {
		existing, err := s.db.GetVM(req.Name)
		if err != nil {
			s.logger.Printf("ERROR: Failed to check existing VM: %v", err)
			types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, "Failed to check existing VM", nil)
			return
		}
		if existing != nil {
			writeJSON(w, http.StatusOK, existing)
			return
		}
	}

	// Validate and resolve image
	image, imageDigest, err := s.validateAndResolveImage(req.Image)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			types.WriteError(w, http.StatusNotFound, types.ErrImageNotFound,
				fmt.Sprintf("Image '%s' not found locally. Pull it first.", req.Image),
				map[string]interface{}{"image": req.Image})
		} else {
			s.logger.Printf("ERROR: Failed to get image: %v", err)
			types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, "Failed to get image", nil)
		}
		return
	}

	// Generate VM ID and name
	vmID := uuid.New().String()
	vmName := req.Name
	if vmName == "" {
		vmName = vmID
	}

	// Build VM config with defaults and user overrides
	config := buildVMConfig(image, &req)

	// Validate resource limits
	if err := s.validateVMResourceLimits(config); err != nil {
		s.logger.Printf("ERROR: Resource validation failed: %v", err)
		types.WriteError(w, http.StatusUnprocessableEntity, types.ErrResourceLimitExceeded, err.Error(), nil)
		return
	}

	// Create VM-specific writable root disk before any privileged network setup.
	if err := materializeWritableRootDisks(s.config.Storage.DataDir, vmID, &config); err != nil {
		s.logger.Printf("ERROR: Rootfs materialization failed: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, err.Error(), nil)
		return
	}

	// Setup networking
	if err := s.setupVMNetworking(vmID, &config); err != nil {
		if cleanupErr := cleanupVMStorage(s.config.Storage.DataDir, vmID); cleanupErr != nil {
			s.logger.Printf("WARN: Failed to cleanup VM storage after network setup failure: %v", cleanupErr)
		}
		if writeNetworkSetupError(w, err, config.Network.Mode) {
			return
		}
		s.logger.Printf("ERROR: Network setup failed: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, err.Error(), nil)
		return
	}

	// Create VM
	now := time.Now()
	// Attempt SPIRE registration (best-effort, won't fail VM creation)
	spiffeID, spireErr := s.registerSPIREWorkload(r.Context(), vmID, &req)
	if spireErr != nil {
		// Log error but continue - SPIRE registration is optional
		s.logger.Printf("WARN: %v", spireErr)
	}

	vm := &types.VM{
		ID:           vmID,
		Name:         vmName,
		State:        types.StateCreated,
		Image:        req.Image,
		ImageDigest:  imageDigest,
		Architecture: image.Architecture,
		Config:       config,
		CreatedAt:    now,
		UpdatedAt:    now,
		OwnerUserID:  req.OwnerUserID,
		GroupID:      req.GroupID,
		SpiffeID:     spiffeID,
	}

	if err := s.db.CreateVM(vm); err != nil {
		s.logger.Printf("ERROR: Failed to create VM: %v", err)
		s.cleanupCreatedVMResources(vmID, config)
		// Cleanup SPIRE entry on failure (best-effort)
		if vm.SpiffeID != "" && s.spireService != nil {
			if delErr := s.spireService.DeleteVMWorkloadEntry(r.Context(), vm.SpiffeID); delErr != nil {
				s.logger.Printf("WARN: Failed to cleanup SPIRE workload entry for VM %s (%s): %v", vm.Name, vm.ID, delErr)
			}
		}
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, "Failed to create VM", nil)
		return
	}

	s.logger.Printf("INFO: Created VM: %s (%s)", vm.Name, vm.ID)
	writeJSON(w, http.StatusCreated, vm)
}

// handleGetVM gets a specific VM
func (s *Server) handleGetVM(w http.ResponseWriter, r *http.Request, vmID string) {
	vm, err := s.db.GetVM(vmID)
	if err != nil {
		s.logger.Printf("ERROR: Failed to get VM: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, "Failed to get VM", nil)
		return
	}

	if vm == nil {
		types.WriteError(w, http.StatusNotFound, types.ErrVMNotFound,
			fmt.Sprintf("Virtual machine with ID '%s' does not exist", vmID),
			map[string]interface{}{"vm_id": vmID})
		return
	}

	writeJSON(w, http.StatusOK, vm)
}

// handleDeleteVM deletes a VM
func (s *Server) handleDeleteVM(w http.ResponseWriter, r *http.Request, vmID string) {
	// Get VM
	vm, err := s.db.GetVM(vmID)
	if err != nil {
		s.logger.Printf("ERROR: Failed to get VM: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, "Failed to get VM", nil)
		return
	}

	if vm == nil {
		types.WriteError(w, http.StatusNotFound, types.ErrVMNotFound,
			fmt.Sprintf("Virtual machine with ID '%s' does not exist", vmID),
			map[string]interface{}{"vm_id": vmID})
		return
	}

	// Try to acquire lock
	if err := s.db.AcquireLock(vm.ID, "delete"); err != nil {
		types.WriteError(w, http.StatusConflict, types.ErrVMLocked, "VM is locked by another operation", nil)
		return
	}
	defer func() {
		if err := s.db.ReleaseLock(vm.ID); err != nil {
			s.logger.Printf("WARN: Failed to release lock: %v", err)
		}
	}()

	// Stop VM if running
	if (vm.State == types.StateRunning || vm.State == types.StatePaused) && vmHasRuntimeHandle(vm) {
		if err := s.runtimeManager.Kill(vm); err != nil {
			s.logger.Printf("ERROR: Failed to kill VM before delete: %v", err)
			types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError,
				"Failed to stop VM runtime before delete", map[string]interface{}{"vm_id": vm.ID})
			return
		}
	}
	if err := s.runtimeManager.Delete(vm); err != nil {
		s.logger.Printf("ERROR: Failed to delete VM runtime resources: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError,
			"Failed to delete VM runtime resources", map[string]interface{}{"vm_id": vm.ID})
		return
	}

	// Cleanup SPIRE workload entry if registered
	if vm.SpiffeID != "" && s.spireService != nil && s.spireService.IsEnabled() {
		if err := s.spireService.DeleteVMWorkloadEntry(r.Context(), vm.SpiffeID); err != nil {
			s.logger.Printf("WARN: Failed to delete SPIRE workload entry for VM %s: %v", vm.ID, err)
		} else {
			s.logger.Printf("INFO: Deleted SPIRE workload entry: %s", vm.SpiffeID)
		}
	}

	s.cleanupDeletedVMResources(vm)

	// Delete from database
	if err := s.db.DeleteVM(vm.ID); err != nil {
		s.logger.Printf("ERROR: Failed to delete VM: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, "Failed to delete VM", nil)
		return
	}

	s.logger.Printf("INFO: Deleted VM: %s (%s)", vm.Name, vm.ID)
	w.WriteHeader(http.StatusNoContent)
}

// startRuntimeAndSetupNetwork starts the configured runtime and sets up Linux host port forwards when needed.
func (s *Server) startRuntimeAndSetupNetwork(vm *types.VM, image *types.Image) error {
	if err := s.runtimeManager.Start(vm, image); err != nil {
		return fmt.Errorf("failed to start runtime: %w", err)
	}

	// Firecracker uses host iptables for port forwards. Runtime-managed backends
	// translate port forwards before launch.
	if selectedRuntimeDriver(s.config) == "firecracker" && len(vm.Config.Network.PortForwards) > 0 {
		s.logger.Printf("INFO: Setting up %d port forward(s) for VM %s", len(vm.Config.Network.PortForwards), vm.Name)
		if err := network.SetupPortForwards(vm.Config.Network.IPAddress, vm.Config.Network.PortForwards); err != nil {
			_ = s.runtimeManager.Kill(vm) // Cleanup
			return fmt.Errorf("failed to setup port forwards: %w", err)
		}
		for _, pf := range vm.Config.Network.PortForwards {
			s.logger.Printf("INFO: Port forward: host:%d -> %s:%d (%s)",
				pf.HostPort, vm.Config.Network.IPAddress, pf.VMPort, pf.Protocol)
		}
	}

	return nil
}

// performVMStart executes the VM start operation
func (s *Server) performVMStart(vm *types.VM) error {
	// Update state to starting
	vm.State = types.StateStarting
	if err := s.db.UpdateVM(vm); err != nil {
		s.logger.Printf("ERROR: Failed to update VM: %v", err)
		return fmt.Errorf("failed to update VM state")
	}

	// Get image
	image, err := s.db.GetImage(vm.ImageDigest)
	if err != nil || image == nil {
		s.logger.Printf("ERROR: Failed to get image: %v", err)
		vm.State = types.StateFailed
		_ = s.db.UpdateVM(vm)
		return fmt.Errorf("failed to get image")
	}

	// Start runtime and setup network
	if err := s.startRuntimeAndSetupNetwork(vm, image); err != nil {
		s.logger.Printf("ERROR: %v", err)
		vm.State = types.StateFailed
		_ = s.db.UpdateVM(vm)
		return err
	}

	// Update state to running
	vm.State = types.StateRunning
	if err := s.db.UpdateVM(vm); err != nil {
		s.logger.Printf("ERROR: Failed to update VM: %v", err)
	}

	return nil
}

// stopVMAndCleanup stops Firecracker and cleans up port forwards
func (s *Server) stopVMAndCleanup(vm *types.VM, timeout int) error {
	// Stop VM
	if err := s.runtimeManager.Stop(vm, timeout); err != nil {
		return fmt.Errorf("failed to stop VM: %w", err)
	}

	// Cleanup Linux host port forwards if configured.
	if selectedRuntimeDriver(s.config) == "firecracker" && len(vm.Config.Network.PortForwards) > 0 {
		s.logger.Printf("INFO: Cleaning up %d port forward(s) for VM %s", len(vm.Config.Network.PortForwards), vm.Name)
		if err := network.CleanupPortForwards(vm.Config.Network.IPAddress, vm.Config.Network.PortForwards); err != nil {
			s.logger.Printf("WARN: Failed to cleanup port forwards: %v", err)
		}
	}

	return nil
}

// performVMStop executes the VM stop operation
func (s *Server) performVMStop(vm *types.VM, timeout int) error {
	// Update state
	vm.State = types.StateStopping
	if err := s.db.UpdateVM(vm); err != nil {
		s.logger.Printf("ERROR: Failed to update VM: %v", err)
	}

	// Stop VM and cleanup
	if err := s.stopVMAndCleanup(vm, timeout); err != nil {
		s.logger.Printf("ERROR: %v", err)
		vm.State = types.StateFailed
		_ = s.db.UpdateVM(vm)
		return err
	}

	// Update state
	vm.State = types.StateStopped
	vm.Runtime = nil
	if err := s.db.UpdateVM(vm); err != nil {
		s.logger.Printf("ERROR: Failed to update VM: %v", err)
	}

	return nil
}

// ============================================================================
// Go 1.22+ Path Parameter Wrappers
// These handlers extract the VM ID from the URL path using r.PathValue()
// ============================================================================

// handleGetVMByPath handles GET /vms/{id} using path parameters
func (s *Server) handleGetVMByPath(w http.ResponseWriter, r *http.Request) {
	vmID := r.PathValue("id")
	if vmID == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest, "VM ID is required", nil)
		return
	}
	s.handleGetVM(w, r, vmID)
}

// handleDeleteVMByPath handles DELETE /vms/{id} using path parameters
func (s *Server) handleDeleteVMByPath(w http.ResponseWriter, r *http.Request) {
	vmID := r.PathValue("id")
	if vmID == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest, "VM ID is required", nil)
		return
	}
	s.handleDeleteVM(w, r, vmID)
}

// handleVMStartByPath handles POST /vms/{id}/start using path parameters
func (s *Server) handleVMStartByPath(w http.ResponseWriter, r *http.Request) {
	vmID := r.PathValue("id")
	if vmID == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest, "VM ID is required", nil)
		return
	}

	// Get VM
	vm, err := s.db.GetVM(vmID)
	if err != nil {
		s.logger.Printf("ERROR: Failed to get VM: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, "Failed to get VM", nil)
		return
	}

	if vm == nil {
		types.WriteError(w, http.StatusNotFound, types.ErrVMNotFound,
			fmt.Sprintf("Virtual machine with ID '%s' does not exist", vmID),
			map[string]interface{}{"vm_id": vmID})
		return
	}

	// Check if already running (idempotent)
	if vm.State == types.StateRunning {
		writeJSON(w, http.StatusOK, vm)
		return
	}

	// Validate state transition
	if vm.State != types.StateCreated && vm.State != types.StateStopped {
		types.WriteError(w, http.StatusConflict, types.ErrInvalidStateTransition,
			fmt.Sprintf("Cannot start VM in state '%s'", vm.State),
			map[string]interface{}{"current_state": vm.State})
		return
	}

	// Acquire lock
	if err := s.db.AcquireLock(vm.ID, "start"); err != nil {
		types.WriteError(w, http.StatusConflict, types.ErrVMLocked, "VM is locked by another operation", nil)
		return
	}
	defer func() {
		if err := s.db.ReleaseLock(vm.ID); err != nil {
			s.logger.Printf("WARN: Failed to release lock: %v", err)
		}
	}()

	// Perform start operation
	if err := s.performVMStart(vm); err != nil {
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, err.Error(), nil)
		return
	}

	s.logger.Printf("INFO: Started VM: %s (%s)", vm.Name, vm.ID)
	writeJSON(w, http.StatusOK, vm)
}

// handleVMStopByPath handles POST /vms/{id}/stop using path parameters
func (s *Server) handleVMStopByPath(w http.ResponseWriter, r *http.Request) {
	vmID := r.PathValue("id")
	if vmID == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest, "VM ID is required", nil)
		return
	}

	// Parse request body
	var req types.StopVMRequest
	if r.ContentLength > 0 {
		if err := readJSON(r, &req); err != nil {
			types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest, "Invalid request body", nil)
			return
		}
	}

	timeout := 30
	if req.TimeoutSeconds != nil {
		timeout = *req.TimeoutSeconds
	}

	// Get VM
	vm, err := s.db.GetVM(vmID)
	if err != nil {
		s.logger.Printf("ERROR: Failed to get VM: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, "Failed to get VM", nil)
		return
	}

	if vm == nil {
		types.WriteError(w, http.StatusNotFound, types.ErrVMNotFound,
			fmt.Sprintf("Virtual machine with ID '%s' does not exist", vmID),
			map[string]interface{}{"vm_id": vmID})
		return
	}

	// Check if already stopped (idempotent)
	if vm.State == types.StateStopped {
		writeJSON(w, http.StatusOK, vm)
		return
	}

	// Validate state transition
	if vm.State != types.StateRunning && vm.State != types.StatePaused {
		types.WriteError(w, http.StatusConflict, types.ErrInvalidStateTransition,
			fmt.Sprintf("Cannot stop VM in state '%s'", vm.State),
			map[string]interface{}{"current_state": vm.State})
		return
	}

	// Acquire lock
	if err := s.db.AcquireLock(vm.ID, "stop"); err != nil {
		types.WriteError(w, http.StatusConflict, types.ErrVMLocked, "VM is locked by another operation", nil)
		return
	}
	defer func() {
		if err := s.db.ReleaseLock(vm.ID); err != nil {
			s.logger.Printf("WARN: Failed to release lock: %v", err)
		}
	}()

	// Perform stop operation
	if err := s.performVMStop(vm, timeout); err != nil {
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, err.Error(), nil)
		return
	}

	s.logger.Printf("INFO: Stopped VM: %s (%s)", vm.Name, vm.ID)
	writeJSON(w, http.StatusOK, vm)
}

// handleVMKillByPath handles POST /vms/{id}/kill using path parameters
func (s *Server) handleVMKillByPath(w http.ResponseWriter, r *http.Request) {
	vmID := r.PathValue("id")
	if vmID == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest, "VM ID is required", nil)
		return
	}

	// Get VM
	vm, err := s.db.GetVM(vmID)
	if err != nil {
		s.logger.Printf("ERROR: Failed to get VM: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, "Failed to get VM", nil)
		return
	}

	if vm == nil {
		types.WriteError(w, http.StatusNotFound, types.ErrVMNotFound,
			fmt.Sprintf("Virtual machine with ID '%s' does not exist", vmID),
			map[string]interface{}{"vm_id": vmID})
		return
	}

	// Kill VM
	if vmHasRuntimeHandle(vm) {
		if err := s.runtimeManager.Kill(vm); err != nil {
			s.logger.Printf("WARN: Failed to kill VM: %v", err)
		}
	}

	// Update state
	vm.State = types.StateStopped
	vm.Runtime = nil
	if err := s.db.UpdateVM(vm); err != nil {
		s.logger.Printf("ERROR: Failed to update VM: %v", err)
	}

	s.logger.Printf("INFO: Killed VM: %s (%s)", vm.Name, vm.ID)
	writeJSON(w, http.StatusOK, vm)
}

func vmHasRuntimeHandle(vm *types.VM) bool {
	return vm != nil && vm.Runtime != nil && (vm.Runtime.PID != 0 || vm.Runtime.ExternalID != "")
}

func normalizeExecCommand(command []string) []string {
	if len(command) == 0 || strings.TrimSpace(command[0]) == "" {
		return nil
	}
	return append([]string(nil), command...)
}

// handleVMPauseByPath handles POST /vms/{id}/pause using path parameters
func (s *Server) handleVMPauseByPath(w http.ResponseWriter, r *http.Request) {
	vmID := r.PathValue("id")
	if vmID == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest, "VM ID is required", nil)
		return
	}

	// Get VM
	vm, err := s.db.GetVM(vmID)
	if err != nil {
		s.logger.Printf("ERROR: Failed to get VM: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, "Failed to get VM", nil)
		return
	}

	if vm == nil {
		types.WriteError(w, http.StatusNotFound, types.ErrVMNotFound,
			fmt.Sprintf("Virtual machine with ID '%s' does not exist", vmID),
			map[string]interface{}{"vm_id": vmID})
		return
	}

	// Check if VM is running
	if vm.State != types.StateRunning {
		types.WriteError(w, http.StatusConflict, types.ErrInvalidStateTransition,
			fmt.Sprintf("Cannot pause VM in state '%s'", vm.State),
			map[string]interface{}{"current_state": vm.State})
		return
	}

	if err := s.db.AcquireLock(vm.ID, "pause"); err != nil {
		types.WriteError(w, http.StatusConflict, types.ErrVMLocked, "VM is locked by another operation", nil)
		return
	}
	defer func() {
		if err := s.db.ReleaseLock(vm.ID); err != nil {
			s.logger.Printf("WARN: Failed to release lock: %v", err)
		}
	}()

	vm.State = types.StatePausing
	if err := s.db.UpdateVM(vm); err != nil {
		s.logger.Printf("ERROR: Failed to update VM: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, "Failed to update VM state", nil)
		return
	}

	if err := s.runtimeManager.Pause(vm); err != nil {
		s.logger.Printf("ERROR: Failed to pause VM: %v", err)
		vm.State = types.StateRunning
		_ = s.db.UpdateVM(vm)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError,
			fmt.Sprintf("Failed to pause VM: %v", err), nil)
		return
	}

	vm.State = types.StatePaused
	if err := s.db.UpdateVM(vm); err != nil {
		s.logger.Printf("ERROR: Failed to update VM: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, "Failed to update VM state", nil)
		return
	}

	s.logger.Printf("INFO: Paused VM: %s (%s)", vm.Name, vm.ID)
	writeJSON(w, http.StatusOK, vm)
}

// handleVMResumeByPath handles POST /vms/{id}/resume using path parameters
func (s *Server) handleVMResumeByPath(w http.ResponseWriter, r *http.Request) {
	vmID := r.PathValue("id")
	if vmID == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest, "VM ID is required", nil)
		return
	}

	// Get VM
	vm, err := s.db.GetVM(vmID)
	if err != nil {
		s.logger.Printf("ERROR: Failed to get VM: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, "Failed to get VM", nil)
		return
	}

	if vm == nil {
		types.WriteError(w, http.StatusNotFound, types.ErrVMNotFound,
			fmt.Sprintf("Virtual machine with ID '%s' does not exist", vmID),
			map[string]interface{}{"vm_id": vmID})
		return
	}

	// Check if VM is paused
	if vm.State != types.StatePaused {
		types.WriteError(w, http.StatusConflict, types.ErrInvalidStateTransition,
			fmt.Sprintf("Cannot resume VM in state '%s'", vm.State),
			map[string]interface{}{"current_state": vm.State})
		return
	}

	if err := s.db.AcquireLock(vm.ID, "resume"); err != nil {
		types.WriteError(w, http.StatusConflict, types.ErrVMLocked, "VM is locked by another operation", nil)
		return
	}
	defer func() {
		if err := s.db.ReleaseLock(vm.ID); err != nil {
			s.logger.Printf("WARN: Failed to release lock: %v", err)
		}
	}()

	vm.State = types.StateResuming
	if err := s.db.UpdateVM(vm); err != nil {
		s.logger.Printf("ERROR: Failed to update VM: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, "Failed to update VM state", nil)
		return
	}

	if err := s.runtimeManager.Resume(vm); err != nil {
		s.logger.Printf("ERROR: Failed to resume VM: %v", err)
		vm.State = types.StatePaused
		_ = s.db.UpdateVM(vm)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError,
			fmt.Sprintf("Failed to resume VM: %v", err), nil)
		return
	}

	vm.State = types.StateRunning
	if err := s.db.UpdateVM(vm); err != nil {
		s.logger.Printf("ERROR: Failed to update VM: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, "Failed to update VM state", nil)
		return
	}

	s.logger.Printf("INFO: Resumed VM: %s (%s)", vm.Name, vm.ID)
	writeJSON(w, http.StatusOK, vm)
}

// handleVMLogsByPath handles GET /vms/{id}/logs using path parameters
func (s *Server) handleVMLogsByPath(w http.ResponseWriter, r *http.Request) {
	vmID := r.PathValue("id")
	if vmID == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest, "VM ID is required", nil)
		return
	}

	// Parse tail parameter
	tailLines := 0
	if tailStr := r.URL.Query().Get("tail"); tailStr != "" {
		parsed, err := strconv.Atoi(tailStr)
		if err != nil {
			types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest,
				fmt.Sprintf("Invalid tail parameter: %s", tailStr), nil)
			return
		}
		if parsed < 0 {
			types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest,
				"tail parameter must be non-negative", nil)
			return
		}
		tailLines = parsed
	}

	// Get VM
	vm, err := s.db.GetVM(vmID)
	if err != nil {
		s.logger.Printf("ERROR: Failed to get VM: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, "Failed to get VM", nil)
		return
	}

	if vm == nil {
		types.WriteError(w, http.StatusNotFound, types.ErrVMNotFound,
			fmt.Sprintf("Virtual machine with ID '%s' does not exist", vmID),
			map[string]interface{}{"vm_id": vmID})
		return
	}

	// Get console logs
	logs, err := s.runtimeManager.GetConsoleLogs(vm, tailLines)
	if err != nil {
		s.logger.Printf("ERROR: Failed to get console logs: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, "Failed to get console logs", nil)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	// Prevent the browser from MIME-sniffing console output into active content (XSS).
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	// XSS is mitigated by Content-Type text/plain + X-Content-Type-Options: nosniff
	// set above; gosec's taint tracker does not model response headers.
	if _, err := w.Write(logs); err != nil { //nolint:gosec // text/plain + nosniff prevents content sniffing
		s.logger.Printf("WARN: Failed to write logs response: %v", err)
	}
}

// handleVMExecByPath handles POST /vms/{id}/exec using path parameters.
func (s *Server) handleVMExecByPath(w http.ResponseWriter, r *http.Request) {
	vmID := r.PathValue("id")
	if vmID == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest, "VM ID is required", nil)
		return
	}

	var req types.VMExecRequest
	if err := readJSON(r, &req); err != nil {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest, "Invalid request body", nil)
		return
	}

	command := normalizeExecCommand(req.Command)
	if len(command) == 0 {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest, "command is required", nil)
		return
	}

	timeoutSeconds := defaultExecTimeoutSeconds
	if req.TimeoutSeconds != nil {
		timeoutSeconds = *req.TimeoutSeconds
	}
	if timeoutSeconds <= 0 || timeoutSeconds > maxExecTimeoutSeconds {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest,
			fmt.Sprintf("timeout_seconds must be between 1 and %d", maxExecTimeoutSeconds), nil)
		return
	}

	vm, err := s.db.GetVM(vmID)
	if err != nil {
		s.logger.Printf("ERROR: Failed to get VM: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, "Failed to get VM", nil)
		return
	}
	if vm == nil {
		types.WriteError(w, http.StatusNotFound, types.ErrVMNotFound,
			fmt.Sprintf("Virtual machine with ID '%s' does not exist", vmID),
			map[string]interface{}{"vm_id": vmID})
		return
	}
	if vm.State != types.StateRunning {
		types.WriteError(w, http.StatusConflict, types.ErrInvalidStateTransition,
			fmt.Sprintf("Cannot exec in VM in state '%s'", vm.State),
			map[string]interface{}{"current_state": vm.State})
		return
	}
	if !vmHasRuntimeHandle(vm) {
		types.WriteError(w, http.StatusConflict, types.ErrInvalidStateTransition,
			"Cannot exec because VM runtime handle is unavailable", map[string]interface{}{"vm_id": vm.ID})
		return
	}

	execer, ok := s.runtimeManager.(vmm.CommandExecutor)
	if !ok {
		types.WriteError(w, http.StatusNotImplemented, types.ErrUnsupportedOperation,
			"Runtime does not support VM exec", map[string]interface{}{"vm_id": vm.ID})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	result, err := execer.Exec(ctx, vm, command)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			types.WriteError(w, http.StatusGatewayTimeout, types.ErrServiceUnavailable,
				"VM exec timed out", map[string]interface{}{"vm_id": vm.ID, "timeout_seconds": timeoutSeconds})
			return
		}
		s.logger.Printf("ERROR: Failed to exec in VM: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError,
			"Failed to exec in VM", map[string]interface{}{"vm_id": vm.ID})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// handleVMProcessExit handles VM process termination
// This is called by the firecracker manager when a VM process exits
// It updates VM state and cleans up resources
func (s *Server) handleVMProcessExit(vmID string, exitCode *int, err error) {
	s.logger.Printf("INFO: Handling process exit for VM %s", vmID)

	// Get the VM from database
	vm, dbErr := s.db.GetVM(vmID)
	if dbErr != nil {
		s.logger.Printf("ERROR: Failed to get VM %s during process exit handling: %v", vmID, dbErr)
		return
	}
	if vm == nil {
		s.logger.Printf("WARN: VM %s not found during process exit handling (may have been deleted)", vmID)
		return
	}

	// Determine new state based on exit
	var newState types.VMState
	if err != nil {
		// Process exited with error (could be signal, crash, etc.)
		if exitCode != nil && *exitCode == 0 {
			// Clean exit
			newState = types.StateStopped
		} else {
			// Failed exit
			newState = types.StateFailed
		}
	} else {
		// Normal exit
		newState = types.StateStopped
	}

	// Only update if VM is in a running state (avoid overwriting deliberate state changes)
	if vm.State == types.StateRunning || vm.State == types.StateStarting {
		s.logger.Printf("INFO: VM %s process exited, updating state from %s to %s", vmID, vm.State, newState)

		vm.State = newState
		vm.UpdatedAt = time.Now()

		if err := s.db.UpdateVM(vm); err != nil {
			s.logger.Printf("ERROR: Failed to update VM %s state after process exit: %v", vmID, err)
		}

		// Clean up network resources if VM had networking
		if vm.Config.Network.Mode != "none" {
			s.cleanupVMNetwork(vm)
		}
	} else {
		s.logger.Printf("DEBUG: VM %s already in state %s, not updating after process exit", vmID, vm.State)
	}
}

// cleanupVMNetwork cleans up network resources for a VM
func (s *Server) cleanupVMNetwork(vm *types.VM) {
	// Clean up TAP device
	if err := s.runtimeManager.CleanupNetwork(vm); err != nil {
		s.logger.Printf("WARN: Failed to cleanup network for VM %s: %v", vm.ID, err)
	}

	// Release IP address (uses VM ID, not IP address)
	if vm.Config.Network.IPAddress != "" {
		s.ipam.ReleaseIP(vm.ID)
		s.logger.Printf("INFO: Released IP %s for VM %s", vm.Config.Network.IPAddress, vm.ID)
	}

	// Clean up port forwards
	if len(vm.Config.Network.PortForwards) > 0 {
		if err := network.CleanupPortForwards(vm.Config.Network.IPAddress, vm.Config.Network.PortForwards); err != nil {
			s.logger.Printf("WARN: Failed to cleanup port forwards for VM %s: %v", vm.ID, err)
		}
	}

	if vm.Config.Network.EgressPolicy != nil && vm.Config.Network.EgressPolicy.Enabled {
		if err := network.CleanupEgressPolicy(vm.ID, vm.Config.Network.TapDevice, vm.Config.Network.IPAddress); err != nil {
			s.logger.Printf("WARN: Failed to cleanup egress policy for VM %s: %v", vm.ID, err)
		}
	}
}
