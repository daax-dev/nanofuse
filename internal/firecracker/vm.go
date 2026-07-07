package firecracker

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/daax-dev/nanofuse/internal/types"
	"github.com/daax-dev/nanofuse/internal/vmm"
)

// ProcessExitHandler is called when a VM process exits
// vmID is the VM identifier, exitCode is the process exit code (nil if signal killed)
type ProcessExitHandler = vmm.ProcessExitHandler

// SPIREProxyConfig contains SPIRE-related vsock proxy configuration
type SPIREProxyConfig struct {
	Enabled     bool   // Enable vsock for SPIRE agent access
	VsockCID    uint32 // Guest CID for vsock device (must be >= 3)
	VsockPort   uint32 // Port the VM will connect to (default: 8307)
	AgentSocket string // Path to host SPIRE agent socket
}

// Manager manages Firecracker VMs
type Manager struct {
	binaryPath        string
	dataDir           string
	onProcessExit     ProcessExitHandler
	spireConfig       *SPIREProxyConfig
	vsockProxies      map[string]*VsockProxy // vmID -> proxy
	vsockMu           sync.Mutex             // protects vsockProxies map
	execSSHKey        string                 // daemon private key for `vm exec` over SSH ("" disables exec)
	execSSHUser       string                 // guest SSH user for exec (default "root")
	execVerifyHostKey bool                   // verify guest host key (accept-new TOFU) instead of disabling checks
}

// NewManager creates a new Firecracker manager
func NewManager(binaryPath, dataDir string) *Manager {
	return &Manager{
		binaryPath:   binaryPath,
		dataDir:      dataDir,
		vsockProxies: make(map[string]*VsockProxy),
	}
}

// SetExecSSH configures the daemon-side SSH key and user used to implement
// `vm exec` inside Firecracker guests. An empty keyPath leaves exec disabled,
// in which case Exec reports the operation as unsupported. When verifyHostKey
// is true, exec uses accept-new host-key verification with a known_hosts file
// under the data dir; otherwise host-key checks are disabled (see config).
func (m *Manager) SetExecSSH(keyPath, user string, verifyHostKey bool) {
	m.execSSHKey = keyPath
	m.execSSHUser = user
	m.execVerifyHostKey = verifyHostKey
}

// SetSPIREConfig sets the SPIRE proxy configuration.
// VsockCID must be in the range [3, 2^32-1]:
//   - 0 is reserved for the hypervisor
//   - 1 is reserved for local loopback
//   - 2 is reserved for the host
//   - 3+ are available for guest VMs (no upper bound validation as uint32 naturally limits to 2^32-1)
func (m *Manager) SetSPIREConfig(cfg *SPIREProxyConfig) error {
	if cfg != nil && cfg.Enabled && cfg.VsockCID < 3 {
		return fmt.Errorf("VsockCID must be >= 3 (got %d): 0-2 are reserved", cfg.VsockCID)
	}
	m.spireConfig = cfg
	return nil
}

// SetProcessExitHandler sets the callback for when VM processes exit
// This should be called before starting any VMs
func (m *Manager) SetProcessExitHandler(handler ProcessExitHandler) {
	m.onProcessExit = handler
}

// FirecrackerConfig represents Firecracker configuration
type FirecrackerConfig struct {
	BootSource        BootSource         `json:"boot-source"`
	Drives            []Drive            `json:"drives"`
	MachineConfig     MachineConfig      `json:"machine-config"`
	NetworkInterfaces []NetworkInterface `json:"network-interfaces,omitempty"`
	Vsock             *VsockConfig       `json:"vsock,omitempty"`
}

// VsockConfig represents Firecracker vsock device configuration
// Used for VM-to-host communication, including SPIRE agent access
type VsockConfig struct {
	GuestCID uint32 `json:"guest_cid"` // Guest CID (must be >= 3)
	UDSPath  string `json:"uds_path"`  // Host-side Unix domain socket path
}

type BootSource struct {
	KernelImagePath string `json:"kernel_image_path"`
	BootArgs        string `json:"boot_args,omitempty"`
}

type Drive struct {
	DriveID      string `json:"drive_id"`
	PathOnHost   string `json:"path_on_host"`
	IsReadOnly   bool   `json:"is_read_only"`
	IsRootDevice bool   `json:"is_root_device"`
}

type MachineConfig struct {
	VcpuCount  int  `json:"vcpu_count"`
	MemSizeMib int  `json:"mem_size_mib"`
	SMT        bool `json:"smt"` // Firecracker v1.13+ uses "smt" instead of "ht_enabled"
}

type NetworkInterface struct {
	IfaceID     string `json:"iface_id"`
	GuestMAC    string `json:"guest_mac"`
	HostDevName string `json:"host_dev_name"`
}

type snapshotCreateRequest struct {
	SnapshotType string `json:"snapshot_type"`
	SnapshotPath string `json:"snapshot_path"`
	MemFilePath  string `json:"mem_file_path"`
}

// memoryBackend mirrors Firecracker's MemoryBackend schema. backend_type is one
// of "File" | "Uffd"; "File" lets the kernel handle guest-memory page faults
// from backend_path.
type memoryBackend struct {
	BackendType string `json:"backend_type"`
	BackendPath string `json:"backend_path"`
}

// snapshotLoadRequest mirrors Firecracker's SnapshotLoadParams (v1.15).
// mem_backend is preferred over the deprecated mem_file_path. resume_vm is left
// false so the VM loads paused: LoadSnapshot starts the SPIRE vsock proxy before
// resuming the vCPUs (via a separate PATCH /vm) so an early guest->host vsock
// connection cannot race a not-yet-listening proxy.
type snapshotLoadRequest struct {
	SnapshotPath string        `json:"snapshot_path"`
	MemBackend   memoryBackend `json:"mem_backend"`
	ResumeVM     bool          `json:"resume_vm"`
}

type vmStateRequest struct {
	State string `json:"state"`
}

// addDrivesToConfig adds disk drives to Firecracker config
func addDrivesToConfig(config *FirecrackerConfig, disks []types.DiskConfig) {
	for _, disk := range disks {
		config.Drives = append(config.Drives, Drive{
			DriveID:      disk.DriveID,
			PathOnHost:   disk.PathOnHost,
			IsReadOnly:   disk.IsReadOnly,
			IsRootDevice: disk.IsRootDevice,
		})
	}
}

// addNetworkInterfaceToConfig adds network interface to Firecracker config if needed
func addNetworkInterfaceToConfig(config *FirecrackerConfig, networkConfig types.NetworkConfig) {
	if networkConfig.Mode != "none" {
		config.NetworkInterfaces = append(config.NetworkInterfaces, NetworkInterface{
			IfaceID:     "eth0",
			GuestMAC:    networkConfig.MACAddress,
			HostDevName: networkConfig.TapDevice,
		})
	}
}

// buildKernelArgs assembles kernel command line arguments
func buildKernelArgs(vm *types.VM) string {
	args := vm.Config.KernelArgs
	// Append SSH public key if provided (base64 encoded)
	if vm.Config.SSHPublicKey != "" {
		args += " sshkey=" + vm.Config.SSHPublicKey
	}
	return args
}

// buildFirecrackerConfig creates Firecracker configuration
func (m *Manager) buildFirecrackerConfig(vm *types.VM, image *types.Image, vmDir string) FirecrackerConfig {
	fcConfig := FirecrackerConfig{
		BootSource: BootSource{
			KernelImagePath: image.KernelPath,
			BootArgs:        buildKernelArgs(vm),
		},
		MachineConfig: MachineConfig{
			VcpuCount:  vm.Config.VCPUs,
			MemSizeMib: vm.Config.MemoryMiB,
			SMT:        false,
		},
	}

	addDrivesToConfig(&fcConfig, vm.Config.Disks)
	addNetworkInterfaceToConfig(&fcConfig, vm.Config.Network)

	// Add vsock device for SPIRE agent access if enabled
	if m.spireConfig != nil && m.spireConfig.Enabled && m.spireConfig.VsockCID >= 3 {
		vsockPath := filepath.Join(vmDir, "vsock.sock")
		fcConfig.Vsock = &VsockConfig{
			GuestCID: m.spireConfig.VsockCID,
			UDSPath:  vsockPath,
		}
	}

	return fcConfig
}

// startFirecrackerProcess starts the Firecracker process
func (m *Manager) startFirecrackerProcess(socketPath, configPath, consolePath string) (*exec.Cmd, error) {
	consoleFile, err := os.OpenFile(consolePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to create console log: %w", err)
	}
	defer consoleFile.Close()

	// A snapshot load starts Firecracker with no boot config; the snapshot
	// supplies the machine state, so --config-file is omitted when configPath
	// is empty.
	args := []string{"--api-sock", socketPath}
	if configPath != "" {
		args = append(args, "--config-file", configPath)
	}

	// #nosec G204 - binaryPath is from trusted config file, not user input
	cmd := exec.Command(m.binaryPath, args...)

	cmd.Stdout = consoleFile
	cmd.Stderr = consoleFile
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start Firecracker: %w", err)
	}

	return cmd, nil
}

// writeFirecrackerConfig writes Firecracker config to file
func writeFirecrackerConfig(configPath string, config FirecrackerConfig) error {
	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, configData, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// setupVMRuntime sets up VM runtime info
func setupVMRuntime(vm *types.VM, cmd *exec.Cmd, socketPath, consolePath string) {
	vm.Runtime = &types.VMRuntime{
		PID:         cmd.Process.Pid,
		SocketPath:  socketPath,
		ConsolePath: consolePath,
	}

	// For NAT mode, set network info
	if vm.Config.Network.Mode == "nat" {
		vm.Runtime.NetworkInfo = &types.NetworkRuntimeInfo{
			TapDevice: vm.Config.Network.TapDevice,
			HostIP:    "172.16.0.1",
			GuestIP:   vm.Config.Network.IPAddress,
			Gateway:   vm.Config.Network.Gateway,
		}
	}
}

// Start starts a Firecracker VM
func (m *Manager) Start(vm *types.VM, image *types.Image) error {
	vmDir := filepath.Join(m.dataDir, "vms", vm.ID)
	if err := os.MkdirAll(vmDir, 0755); err != nil {
		return fmt.Errorf("failed to create VM directory: %w", err)
	}

	socketPath := filepath.Join(vmDir, "firecracker.sock")
	consolePath := filepath.Join(vmDir, "console.log")
	configPath := filepath.Join(vmDir, "config.json")

	// Build and write config (includes vsock if SPIRE enabled)
	fcConfig := m.buildFirecrackerConfig(vm, image, vmDir)
	if err := writeFirecrackerConfig(configPath, fcConfig); err != nil {
		return err
	}

	// Start Firecracker process
	cmd, err := m.startFirecrackerProcess(socketPath, configPath, consolePath)
	if err != nil {
		return err
	}

	// Setup runtime info
	setupVMRuntime(vm, cmd, socketPath, consolePath)

	// Start vsock proxy for SPIRE agent access if configured
	if fcConfig.Vsock != nil {
		m.startVsockProxyIfConfigured(vm.ID, fcConfig.Vsock.UDSPath)
	}

	// Start goroutine to wait on process and reap zombie
	// This prevents zombie processes by calling Wait() when the process exits
	vmID := vm.ID // Capture for goroutine
	go m.waitForProcessExit(vmID, cmd)

	return nil
}

// waitForProcessExit waits for the VM process to exit and calls the exit handler
// This is the key fix for the zombie process bug - calling Wait() reaps the child
func (m *Manager) waitForProcessExit(vmID string, cmd *exec.Cmd) {
	// Wait for the process to exit - this reaps the zombie
	err := cmd.Wait()

	// Extract exit code if available
	var exitCode *int
	if cmd.ProcessState != nil {
		code := cmd.ProcessState.ExitCode()
		exitCode = &code
	}

	// Log the exit
	if err != nil {
		log.Printf("INFO: VM %s process exited with error: %v", vmID, err)
	} else if exitCode != nil {
		log.Printf("INFO: VM %s process exited with code %d", vmID, *exitCode)
	} else {
		log.Printf("INFO: VM %s process exited", vmID)
	}

	// Call the exit handler if set
	if m.onProcessExit != nil {
		m.onProcessExit(vmID, exitCode, err)
	}
}

// Stop stops a VM gracefully
func (m *Manager) Stop(vm *types.VM, timeoutSeconds int) error {
	if vm.Runtime == nil || vm.Runtime.PID == 0 {
		return fmt.Errorf("VM runtime info not available")
	}

	// Stop vsock proxy if running
	m.vsockMu.Lock()
	proxy, ok := m.vsockProxies[vm.ID]
	if ok {
		delete(m.vsockProxies, vm.ID)
	}
	m.vsockMu.Unlock()
	if ok {
		proxy.Stop()
		log.Printf("INFO: Stopped vsock proxy for VM %s", vm.ID)
	}

	process, err := os.FindProcess(vm.Runtime.PID)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	// Use default timeout if not specified
	if timeoutSeconds <= 0 {
		timeoutSeconds = 10
	}

	// Send SIGTERM for graceful shutdown
	if err := process.Signal(syscall.SIGTERM); err != nil {
		// Process might have already terminated
		if err == os.ErrProcessDone {
			log.Printf("INFO: VM %s (PID %d) already stopped", vm.ID, vm.Runtime.PID)
			return nil
		}
		return fmt.Errorf("failed to send SIGTERM: %w", err)
	}

	log.Printf("INFO: Sent SIGTERM to VM %s (PID %d), waiting up to %d seconds for graceful shutdown",
		vm.ID, vm.Runtime.PID, timeoutSeconds)

	// Wait for process to exit with timeout
	timeout := time.Duration(timeoutSeconds) * time.Second
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	timeoutTimer := time.NewTimer(timeout)
	defer timeoutTimer.Stop()

	for {
		select {
		case <-ticker.C:
			// Check if process is still running
			if !m.IsRunning(vm.Runtime.PID) {
				log.Printf("INFO: VM %s (PID %d) stopped gracefully", vm.ID, vm.Runtime.PID)
				return nil
			}
		case <-timeoutTimer.C:
			// Timeout expired, force kill
			log.Printf("WARN: VM %s (PID %d) did not stop gracefully after %d seconds, sending SIGKILL",
				vm.ID, vm.Runtime.PID, timeoutSeconds)

			if err := process.Signal(syscall.SIGKILL); err != nil {
				if err == os.ErrProcessDone {
					log.Printf("INFO: VM %s (PID %d) stopped during timeout", vm.ID, vm.Runtime.PID)
					return nil
				}
				return fmt.Errorf("failed to send SIGKILL: %w", err)
			}

			// Wait a short time for SIGKILL to take effect
			forceTicker := time.NewTicker(100 * time.Millisecond)
			defer forceTicker.Stop()

			forceTimeout := time.NewTimer(5 * time.Second)
			defer forceTimeout.Stop()

			for {
				select {
				case <-forceTicker.C:
					if !m.IsRunning(vm.Runtime.PID) {
						log.Printf("INFO: VM %s (PID %d) stopped after SIGKILL", vm.ID, vm.Runtime.PID)
						return nil
					}
				case <-forceTimeout.C:
					return fmt.Errorf("VM %s (PID %d) did not stop even after SIGKILL", vm.ID, vm.Runtime.PID)
				}
			}
		}
	}
}

// Kill kills a VM forcefully
func (m *Manager) Kill(vm *types.VM) error {
	if vm.Runtime == nil || vm.Runtime.PID == 0 {
		return fmt.Errorf("VM runtime info not available")
	}

	process, err := os.FindProcess(vm.Runtime.PID)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	// Send SIGKILL
	if err := process.Signal(syscall.SIGKILL); err != nil {
		return fmt.Errorf("failed to send SIGKILL: %w", err)
	}

	return nil
}

// Delete performs runtime-specific cleanup before VM metadata deletion.
func (m *Manager) Delete(vm *types.VM) error {
	return nil
}

// IsRunning checks if a VM process is running
func (m *Manager) IsRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Send signal 0 to check if process exists
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// Pause pauses a VM
func (m *Manager) Pause(vm *types.VM) error {
	socketPath, err := vmAPISocket(vm)
	if err != nil {
		return err
	}
	if err := firecrackerPATCH(socketPath, "/vm", vmStateRequest{State: "Paused"}); err != nil {
		return fmt.Errorf("failed to pause Firecracker VM %s: %w", vm.ID, err)
	}
	return nil
}

// Resume resumes a VM
func (m *Manager) Resume(vm *types.VM) error {
	socketPath, err := vmAPISocket(vm)
	if err != nil {
		return err
	}
	if err := firecrackerPATCH(socketPath, "/vm", vmStateRequest{State: "Resumed"}); err != nil {
		return fmt.Errorf("failed to resume Firecracker VM %s: %w", vm.ID, err)
	}
	return nil
}

func vmAPISocket(vm *types.VM) (string, error) {
	if vm == nil {
		return "", fmt.Errorf("VM is required")
	}
	if vm.Runtime == nil || vm.Runtime.SocketPath == "" {
		return "", fmt.Errorf("VM runtime socket path not available")
	}
	return vm.Runtime.SocketPath, nil
}

func firecrackerPUT(socketPath, endpoint string, payload any) error {
	return firecrackerJSONRequest(socketPath, http.MethodPut, endpoint, payload)
}

func firecrackerPATCH(socketPath, endpoint string, payload any) error {
	return firecrackerJSONRequest(socketPath, http.MethodPatch, endpoint, payload)
}

func firecrackerJSONRequest(socketPath, method, endpoint string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Firecracker request: %w", err)
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			dialer := net.Dialer{}
			return dialer.DialContext(ctx, "unix", socketPath)
		},
	}
	defer transport.CloseIdleConnections()

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	req, err := http.NewRequest(method, "http://unix"+endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to build Firecracker request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("firecracker API request %s failed: %w", endpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		if len(respBody) > 0 {
			return fmt.Errorf("firecracker API request %s failed with status %d: %s", endpoint, resp.StatusCode, string(respBody))
		}
		return fmt.Errorf("firecracker API request %s failed with status %d", endpoint, resp.StatusCode)
	}

	return nil
}

// CreateSnapshot creates a snapshot
func (m *Manager) CreateSnapshot(vm *types.VM, snapshotPath, memPath string) error {
	socketPath, err := vmAPISocket(vm)
	if err != nil {
		return err
	}
	if snapshotPath == "" {
		return fmt.Errorf("snapshot path is required")
	}
	if memPath == "" {
		return fmt.Errorf("snapshot memory path is required")
	}
	if err := os.MkdirAll(filepath.Dir(snapshotPath), 0755); err != nil {
		return fmt.Errorf("failed to create snapshot directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(memPath), 0755); err != nil {
		return fmt.Errorf("failed to create memory snapshot directory: %w", err)
	}

	req := snapshotCreateRequest{
		SnapshotType: "Full",
		SnapshotPath: snapshotPath,
		MemFilePath:  memPath,
	}
	if err := firecrackerPUT(socketPath, "/snapshot/create", req); err != nil {
		return fmt.Errorf("failed to create Firecracker snapshot for VM %s: %w", vm.ID, err)
	}
	return nil
}

// startVsockProxyIfConfigured starts the host-side SPIRE vsock proxy for a VM
// and tracks it in m.vsockProxies for later Stop cleanup. It is best-effort: a
// failure is logged and not fatal (matching Start's behaviour). It is a no-op
// when SPIRE/vsock is not configured, so the same gate that decides whether the
// Firecracker vsock device exists also decides whether the proxy runs. Shared by
// Start and LoadSnapshot so a snapshot-resumed VM gets the same wiring.
// stopVsockProxy stops and removes the tracked vsock proxy for vmID, if any, and
// reports whether one was stopped. The map is mutated under vsockMu but the
// (potentially blocking) Stop runs outside the lock, matching Manager.Stop.
func (m *Manager) stopVsockProxy(vmID string) bool {
	m.vsockMu.Lock()
	proxy, ok := m.vsockProxies[vmID]
	if ok {
		delete(m.vsockProxies, vmID)
	}
	m.vsockMu.Unlock()
	if ok {
		proxy.Stop()
	}
	return ok
}

func (m *Manager) startVsockProxyIfConfigured(vmID, vsockPath string) {
	if m.spireConfig == nil || !m.spireConfig.Enabled || m.spireConfig.VsockCID < 3 {
		return
	}
	// Stop any stale proxy for this vmID before replacing its map entry. If the
	// Firecracker process died without Stop being called, the old proxy is still
	// running but orphaned from the map; overwriting the entry without stopping it
	// would leak the goroutine/socket.
	if m.stopVsockProxy(vmID) {
		log.Printf("INFO: Stopped stale vsock proxy for VM %s before restart", vmID)
	}

	proxy, err := NewVsockProxy(vsockPath, m.spireConfig.AgentSocket, m.spireConfig.VsockPort)
	if err != nil {
		log.Printf("WARN: Failed to create vsock proxy for VM %s: %v", vmID, err)
		return
	}
	if err := proxy.Start(); err != nil {
		log.Printf("WARN: Failed to start vsock proxy for VM %s: %v", vmID, err)
		return
	}
	m.vsockMu.Lock()
	m.vsockProxies[vmID] = proxy
	m.vsockMu.Unlock()
	log.Printf("INFO: Started vsock proxy for VM %s (port %d -> %s)",
		vmID, m.spireConfig.VsockPort, m.spireConfig.AgentSocket)
}

// LoadSnapshot boots a fresh Firecracker process from a previously created
// snapshot and resumes it. Unlike Resume (which unpauses an already-running
// process), a snapshot load requires a brand-new Firecracker process with no
// boot source: the snapshot supplies the full machine state. Any host tap
// devices referenced by the snapshot must already exist and be reachable by the
// new process (re-established by the caller before invoking this method).
func (m *Manager) LoadSnapshot(vm *types.VM, snapshotPath, memPath string) error {
	if vm == nil {
		return fmt.Errorf("VM is required")
	}
	if snapshotPath == "" {
		return fmt.Errorf("snapshot path is required")
	}
	if memPath == "" {
		return fmt.Errorf("snapshot memory path is required")
	}
	if _, err := os.Stat(snapshotPath); err != nil {
		return fmt.Errorf("snapshot state file not accessible at %s: %w", snapshotPath, err)
	}
	if _, err := os.Stat(memPath); err != nil {
		return fmt.Errorf("snapshot memory file not accessible at %s: %w", memPath, err)
	}

	vmDir := filepath.Join(m.dataDir, "vms", vm.ID)
	// 0750 keeps the API socket (which controls the microVM) reachable only by
	// the daemon's user/group, not world-accessible. MkdirAll does not tighten an
	// existing directory (an earlier Start created it 0755), so Chmod enforces the
	// mode for resumed VMs too.
	if err := os.MkdirAll(vmDir, 0750); err != nil {
		return fmt.Errorf("failed to create VM directory: %w", err)
	}
	if err := os.Chmod(vmDir, 0750); err != nil {
		return fmt.Errorf("failed to set VM directory permissions: %w", err)
	}
	socketPath := filepath.Join(vmDir, "firecracker.sock")
	consolePath := filepath.Join(vmDir, "console.log")

	// Remove any stale API socket so the fresh Firecracker process can bind.
	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove stale API socket %s: %w", socketPath, err)
	}

	// Start Firecracker with no boot config; the snapshot supplies machine state.
	cmd, err := m.startFirecrackerProcess(socketPath, "", consolePath)
	if err != nil {
		return err
	}

	// The API socket is not ready the instant the process starts; wait for it
	// before issuing the load request.
	if err := waitForSocketReady(socketPath, 10*time.Second); err != nil {
		killAndReap(cmd)
		return fmt.Errorf("firecracker API socket not ready for VM %s: %w", vm.ID, err)
	}

	// Load the snapshot paused (resume_vm:false). The vCPUs are resumed further
	// down, only after the vsock proxy is listening.
	if err := sendSnapshotLoad(socketPath, snapshotPath, memPath); err != nil {
		killAndReap(cmd)
		return fmt.Errorf("failed to load snapshot for VM %s: %w", vm.ID, err)
	}

	// Restart the host-side SPIRE vsock proxy the same way Start does, so a
	// resumed VM keeps access to host services over vsock and the proxy is
	// tracked in m.vsockProxies for Stop cleanup (otherwise it would leak).
	// Started while the VM is still paused so it is listening before any guest
	// vsock traffic can arrive.
	m.startVsockProxyIfConfigured(vm.ID, filepath.Join(vmDir, "vsock.sock"))

	// Now that the proxy is up, resume the loaded VM's vCPUs.
	if err := firecrackerPATCH(socketPath, "/vm", vmStateRequest{State: "Resumed"}); err != nil {
		m.stopVsockProxy(vm.ID)
		killAndReap(cmd)
		return fmt.Errorf("failed to resume loaded snapshot for VM %s: %w", vm.ID, err)
	}

	// Record runtime info and reap the process on exit only after a fully
	// successful resume. Doing it earlier would let a failed resume PATCH (which
	// kills the process above) leave vm.Runtime pointing at a dead PID/socket that
	// the caller could persist.
	setupVMRuntime(vm, cmd, socketPath, consolePath)
	go m.waitForProcessExit(vm.ID, cmd)

	return nil
}

// sendSnapshotLoad issues the Firecracker PUT /snapshot/load request against an
// already-listening API socket. Split out from LoadSnapshot so the request
// schema can be unit-tested without spawning a Firecracker process.
func sendSnapshotLoad(socketPath, snapshotPath, memPath string) error {
	req := snapshotLoadRequest{
		SnapshotPath: snapshotPath,
		MemBackend: memoryBackend{
			BackendType: "File",
			BackendPath: memPath,
		},
		ResumeVM: false, // load paused; LoadSnapshot resumes after the proxy is up
	}
	if err := firecrackerPUT(socketPath, "/snapshot/load", req); err != nil {
		return fmt.Errorf("firecracker /snapshot/load failed: %w", err)
	}
	return nil
}

// waitForSocketReady blocks until the Firecracker API unix socket accepts
// connections or the timeout elapses.
func waitForSocketReady(socketPath string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for {
		conn, err := net.DialTimeout("unix", socketPath, 200*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		lastErr = err
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out after %s waiting for API socket %s: %w", timeout, socketPath, lastErr)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// killAndReap terminates a Firecracker process and reaps it synchronously to
// avoid leaving a zombie when a load fails before the reaper goroutine starts.
func killAndReap(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	_ = cmd.Process.Kill()
	_ = cmd.Wait()
}

// GetConsoleLogs reads console logs
func (m *Manager) GetConsoleLogs(vm *types.VM, tailLines int) ([]byte, error) {
	if vm.Runtime == nil {
		return nil, fmt.Errorf("VM runtime info not available")
	}

	data, err := os.ReadFile(vm.Runtime.ConsolePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read console logs: %w", err)
	}

	// If tailLines is 0 or negative, return all logs
	if tailLines <= 0 {
		return data, nil
	}

	// Split the data into lines
	lines := bytes.Split(data, []byte("\n"))

	// If we have fewer lines than requested, return all
	if len(lines) <= tailLines {
		return data, nil
	}

	// Get the last N lines
	// Note: If the file ends with a newline, the last element will be empty
	// We should include it to preserve the original format
	tailedLines := lines[len(lines)-tailLines:]

	return bytes.Join(tailedLines, []byte("\n")), nil
}

// SetupNetwork sets up network for a VM
func (m *Manager) SetupNetwork(vm *types.VM, ipam interface{}) error {
	if vm.Config.Network.Mode == "none" {
		return nil
	}

	// Import network package functionality
	// This will be called from the API layer which has access to IPAM
	if vm.Config.Network.TapDevice == "" {
		vm.Config.Network.TapDevice = fmt.Sprintf("tap%s", vm.ID[:8])
	}

	if vm.Config.Network.MACAddress == "" {
		vm.Config.Network.MACAddress = generateMAC()
	}

	return nil
}

// CleanupNetwork cleans up network for a VM
func (m *Manager) CleanupNetwork(vm *types.VM) error {
	// Skip if network mode is "none" or no TAP device configured
	if vm.Config.Network.Mode == "none" || vm.Config.Network.TapDevice == "" {
		return nil
	}

	// Note: This method is intended to be called from the Firecracker manager layer.
	// The full cleanup (including port forwards, IP release, etc.) is handled at the
	// API layer. This method focuses on TAP device removal.

	tapDevice := vm.Config.Network.TapDevice

	// Check if TAP device exists before attempting deletion
	checkCmd := exec.Command("ip", "link", "show", tapDevice)
	if err := checkCmd.Run(); err != nil {
		// Device doesn't exist - this is fine, might have been cleaned up already
		log.Printf("INFO: TAP device %s already removed or doesn't exist", tapDevice)
		return nil
	}

	// Delete the TAP device
	// The device will be automatically detached from any bridge it's connected to
	deleteCmd := exec.Command("ip", "link", "delete", tapDevice)
	if err := deleteCmd.Run(); err != nil {
		// Don't treat this as a fatal error - log warning but return nil
		// This could happen if:
		// - Device was manually deleted between check and delete
		// - Permission issues (shouldn't happen with proper setup)
		// - Device is in an unexpected state
		log.Printf("WARN: Failed to delete TAP device %s: %v", tapDevice, err)
		return nil
	}

	log.Printf("INFO: Cleaned up TAP device: %s", tapDevice)
	return nil
}

func generateMAC() string {
	// Generate a random MAC address with Firecracker OUI
	return fmt.Sprintf("AA:FC:00:%02x:%02x:%02x",
		randomByte(), randomByte(), randomByte())
}

func randomByte() byte {
	b := make([]byte, 1)
	if _, err := rand.Read(b); err != nil {
		// Fallback to zero on error (extremely unlikely with crypto/rand)
		return 0x00
	}
	return b[0]
}
