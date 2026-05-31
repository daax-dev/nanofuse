package applecontainer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/daax-dev/nanofuse/internal/config"
	"github.com/daax-dev/nanofuse/internal/types"
	"github.com/daax-dev/nanofuse/internal/vmm"
)

const (
	DriverName = "apple_container"
	labelVMID  = "nanofuse.vm_id"
)

type commandRunner func(args ...string) ([]byte, error)

// Manager controls macOS lightweight Linux VMs through Apple's container CLI.
type Manager struct {
	binaryPath     string
	dataDir        string
	autoStart      bool
	defaultCommand string
	runCommand     commandRunner
	onProcessExit  vmm.ProcessExitHandler
	watchMu        sync.Mutex
	watching       map[string]struct{}
}

// NewManager creates an Apple container runtime manager.
func NewManager(cfg config.AppleContainerRuntimeConfig, dataDir string) *Manager {
	binaryPath := cfg.BinaryPath
	if binaryPath == "" {
		binaryPath = "/usr/local/bin/container"
	}
	defaultCommand := cfg.DefaultCommand
	if defaultCommand == "" {
		defaultCommand = "sleep infinity"
	}
	m := &Manager{
		binaryPath:     binaryPath,
		dataDir:        dataDir,
		autoStart:      cfg.AutoStart,
		defaultCommand: defaultCommand,
		watching:       make(map[string]struct{}),
	}
	m.runCommand = m.execContainer
	return m
}

func (m *Manager) execContainer(args ...string) ([]byte, error) {
	// #nosec G204 -- binaryPath is trusted daemon configuration; arguments are constructed by Nanofuse.
	cmd := exec.Command(m.binaryPath, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, fmt.Errorf("container %s failed: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return out, nil
}

// SetProcessExitHandler sets the callback for runtime exits.
func (m *Manager) SetProcessExitHandler(handler vmm.ProcessExitHandler) {
	m.onProcessExit = handler
}

func (m *Manager) ensureAvailable() error {
	if _, err := exec.LookPath(m.binaryPath); err != nil {
		if _, statErr := os.Stat(m.binaryPath); statErr != nil {
			return fmt.Errorf("apple container binary not found at %q: %w", m.binaryPath, err)
		}
	}

	if _, err := m.runCommand("system", "status"); err == nil {
		return nil
	} else if !m.autoStart {
		return fmt.Errorf("apple container services are not running: %w", err)
	}

	if _, err := m.runCommand("system", "start", "--enable-kernel-install"); err != nil {
		return fmt.Errorf("failed to start apple container services: %w", err)
	}
	if _, err := m.runCommand("system", "status"); err != nil {
		return fmt.Errorf("apple container services did not become healthy: %w", err)
	}
	return nil
}

type imageListEntry struct {
	Reference  string          `json:"reference"`
	Descriptor imageDescriptor `json:"descriptor"`
}

type imageInspectEntry struct {
	Name     string          `json:"name"`
	Index    imageDescriptor `json:"index"`
	Variants []imageVariant  `json:"variants"`
}

type imageDescriptor struct {
	Digest    string `json:"digest"`
	MediaType string `json:"mediaType"`
	Size      int64  `json:"size"`
}

type imageVariant struct {
	Size     int64         `json:"size"`
	Platform imagePlatform `json:"platform"`
	Config   imageConfig   `json:"config"`
}

type imagePlatform struct {
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	Variant      string `json:"variant"`
}

type imageConfig struct {
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	Variant      string `json:"variant"`
}

// ListImages lists images known to Apple's local container runtime.
func (m *Manager) ListImages() ([]*types.Image, error) {
	if err := m.ensureAvailable(); err != nil {
		return nil, err
	}

	out, err := m.runCommand("images", "list", "--format", "json")
	if err != nil {
		return nil, err
	}
	var entries []imageListEntry
	if err := json.Unmarshal(out, &entries); err != nil {
		return nil, fmt.Errorf("failed to parse apple container image list: %w", err)
	}

	images := make([]*types.Image, 0, len(entries))
	for _, entry := range entries {
		if entry.Reference == "" && entry.Descriptor.Digest == "" {
			continue
		}
		images = append(images, &types.Image{
			Digest:       entry.Descriptor.Digest,
			Tags:         []string{entry.Reference},
			Architecture: runtime.GOARCH,
			SizeBytes:    entry.Descriptor.Size,
			RootfsPath:   "",
			KernelPath:   defaultKernelPath(),
			PulledAt:     time.Now().UTC(),
			Labels:       runtimeLabels(),
		})
	}
	return images, nil
}

// ResolveImage ensures an OCI image is present in Apple's local container store.
func (m *Manager) ResolveImage(imageRef string) (*types.Image, error) {
	imageRef = strings.TrimSpace(imageRef)
	if imageRef == "" {
		return nil, fmt.Errorf("image reference is required")
	}
	if err := m.ensureAvailable(); err != nil {
		return nil, err
	}

	image, err := m.inspectImage(imageRef)
	if err == nil {
		return image, nil
	}

	if _, pullErr := m.runCommand("images", "pull", "--disable-progress-updates", imageRef); pullErr != nil {
		return nil, fmt.Errorf("failed to pull apple container image %q: %w", imageRef, pullErr)
	}

	image, err = m.inspectImage(imageRef)
	if err != nil {
		return nil, err
	}
	return image, nil
}

func (m *Manager) inspectImage(imageRef string) (*types.Image, error) {
	out, err := m.runCommand("images", "inspect", imageRef)
	if err != nil {
		return nil, err
	}
	var entries []imageInspectEntry
	if err := json.Unmarshal(out, &entries); err != nil {
		return nil, fmt.Errorf("failed to parse apple container image inspect: %w", err)
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("apple container image %q not found", imageRef)
	}

	entry := entries[0]
	arch := runtime.GOARCH
	size := entry.Index.Size
	for _, variant := range entry.Variants {
		if variant.Platform.Architecture == runtime.GOARCH || variant.Config.Architecture == runtime.GOARCH {
			arch = runtime.GOARCH
			if variant.Size > 0 {
				size = variant.Size
			}
			break
		}
	}

	name := entry.Name
	if name == "" {
		name = imageRef
	}

	return &types.Image{
		Digest:       entry.Index.Digest,
		Tags:         []string{name},
		Architecture: arch,
		SizeBytes:    size,
		RootfsPath:   "",
		KernelPath:   defaultKernelPath(),
		PulledAt:     time.Now().UTC(),
		Labels:       runtimeLabels(),
	}, nil
}

func runtimeLabels() map[string]string {
	return map[string]string{
		"nanofuse.runtime":        DriverName,
		"nanofuse.virtualization": "apple-virtualization-framework",
	}
}

func defaultKernelPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, "Library", "Application Support", "com.apple.container", "kernels", "default.kernel-"+runtime.GOARCH)
}

type containerInspectEntry struct {
	Status        string                 `json:"status"`
	Networks      []containerNetwork     `json:"networks"`
	Configuration containerConfiguration `json:"configuration"`
}

type containerConfiguration struct {
	ID string `json:"id"`
}

type containerNetwork struct {
	Address string `json:"address"`
	Gateway string `json:"gateway"`
	Network string `json:"network"`
}

func (m *Manager) inspectContainer(name string) (*containerInspectEntry, bool, error) {
	out, err := m.runCommand("inspect", name)
	if err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "not found") || strings.Contains(msg, "no such") || strings.Contains(msg, "does not exist") {
			return nil, false, nil
		}
		return nil, false, err
	}
	var entries []containerInspectEntry
	if err := json.Unmarshal(out, &entries); err != nil {
		return nil, false, fmt.Errorf("failed to parse apple container inspect: %w", err)
	}
	if len(entries) == 0 {
		return nil, false, nil
	}
	return &entries[0], true, nil
}

// Start starts a macOS lightweight Linux VM for a Nanofuse VM.
func (m *Manager) Start(vm *types.VM, image *types.Image) error {
	if err := m.ensureAvailable(); err != nil {
		return err
	}

	vmDir := filepath.Join(m.dataDir, "vms", vm.ID)
	if err := os.MkdirAll(vmDir, 0755); err != nil {
		return fmt.Errorf("failed to create VM directory: %w", err)
	}
	consolePath := filepath.Join(vmDir, "console.log")
	containerName := containerNameForVM(vm.ID)

	inspect, exists, err := m.inspectContainer(containerName)
	if err != nil {
		return err
	}
	if exists && inspect.Status != "running" {
		if out, err := m.runCommand("start", containerName); err != nil {
			_ = appendConsole(consolePath, out)
			return err
		}
	} else if !exists {
		args := m.runArgs(vm, image, containerName)
		out, err := m.runCommand(args...)
		_ = appendConsole(consolePath, out)
		if err != nil {
			return err
		}
	}

	inspect, exists, err = m.inspectContainer(containerName)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("apple container %q did not exist after start", containerName)
	}
	if inspect.Status != "running" {
		return fmt.Errorf("apple container %q is %s after start", containerName, inspect.Status)
	}

	setRuntime(vm, containerName, consolePath, inspect)
	m.watchContainer(vm.ID, containerName)
	return nil
}

func (m *Manager) runArgs(vm *types.VM, image *types.Image, containerName string) []string {
	defaultCommand := strings.Fields(m.defaultCommand)
	args := make([]string, 0, 13+(2*len(vm.Config.Network.PortForwards))+len(defaultCommand))
	args = append(args,
		"run",
		"-d",
		"--name", containerName,
		"--cpus", strconv.Itoa(maxInt(vm.Config.VCPUs, 1)),
		"--memory", fmt.Sprintf("%dM", maxInt(vm.Config.MemoryMiB, 128)),
		"--label", fmt.Sprintf("%s=%s", labelVMID, vm.ID),
	)
	for _, pf := range vm.Config.Network.PortForwards {
		protocol := pf.Protocol
		if protocol == "" {
			protocol = "tcp"
		}
		args = append(args, "--publish", fmt.Sprintf("127.0.0.1:%d:%d/%s", pf.HostPort, pf.VMPort, protocol))
	}
	args = append(args, imageReference(image, vm.Image))
	args = append(args, defaultCommand...)
	return args
}

func imageReference(image *types.Image, fallback string) string {
	if image != nil {
		for _, tag := range image.Tags {
			if strings.TrimSpace(tag) != "" {
				return tag
			}
		}
		if image.Digest != "" && strings.Contains(fallback, "@") {
			return fallback
		}
	}
	return fallback
}

func setRuntime(vm *types.VM, containerName, consolePath string, inspect *containerInspectEntry) {
	vm.Runtime = &types.VMRuntime{
		PID:         0,
		Driver:      DriverName,
		ExternalID:  containerName,
		ConsolePath: consolePath,
	}
	if len(inspect.Networks) > 0 {
		net := inspect.Networks[0]
		vm.Runtime.NetworkInfo = &types.NetworkRuntimeInfo{
			HostIP:  net.Gateway,
			GuestIP: strings.TrimSuffix(strings.Split(net.Address, "/")[0], "/"),
			Gateway: net.Gateway,
		}
	}
}

func containerNameForVM(vmID string) string {
	cleaned := strings.NewReplacer("-", "", "_", "", ".", "").Replace(vmID)
	if len(cleaned) > 24 {
		cleaned = cleaned[:24]
	}
	return "nf-" + cleaned
}

func (m *Manager) watchContainer(vmID, containerName string) {
	m.watchMu.Lock()
	if _, ok := m.watching[vmID]; ok {
		m.watchMu.Unlock()
		return
	}
	m.watching[vmID] = struct{}{}
	m.watchMu.Unlock()

	go func() {
		defer func() {
			m.watchMu.Lock()
			delete(m.watching, vmID)
			m.watchMu.Unlock()
		}()

		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			inspect, exists, err := m.inspectContainer(containerName)
			if err != nil {
				if m.onProcessExit != nil {
					m.onProcessExit(vmID, nil, err)
				}
				return
			}
			if !exists || inspect.Status != "running" {
				if m.onProcessExit != nil {
					m.onProcessExit(vmID, nil, nil)
				}
				return
			}
		}
	}()
}

// Stop stops a VM gracefully.
func (m *Manager) Stop(vm *types.VM, timeoutSeconds int) error {
	name := runtimeName(vm)
	if name == "" {
		return fmt.Errorf("apple container runtime id not available")
	}
	if timeoutSeconds <= 0 {
		timeoutSeconds = 10
	}
	if _, err := m.runCommand("stop", "--time", strconv.Itoa(timeoutSeconds), name); err != nil {
		return err
	}
	return m.waitForContainerStopped(name, time.Duration(timeoutSeconds)*time.Second)
}

// Kill kills a VM forcefully.
func (m *Manager) Kill(vm *types.VM) error {
	name := runtimeName(vm)
	if name == "" {
		return fmt.Errorf("apple container runtime id not available")
	}
	if _, err := m.runCommand("kill", name); err != nil {
		return err
	}
	return m.waitForContainerStopped(name, 5*time.Second)
}

func (m *Manager) waitForContainerStopped(name string, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = time.Second
	}
	deadline := time.Now().Add(timeout)
	for {
		inspect, exists, err := m.inspectContainer(name)
		if err != nil {
			return err
		}
		if !exists || inspect.Status != "running" {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("apple container %q still running after %s", name, timeout)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// Delete deletes the stopped container backing a VM.
func (m *Manager) Delete(vm *types.VM) error {
	name := containerNameForVM(vm.ID)
	if vm.Runtime != nil && vm.Runtime.ExternalID != "" {
		name = vm.Runtime.ExternalID
	}
	if inspect, exists, err := m.inspectContainer(name); err != nil {
		return err
	} else if !exists {
		return nil
	} else if inspect.Status == "running" {
		if err := m.Kill(&types.VM{ID: vm.ID, Runtime: &types.VMRuntime{ExternalID: name}}); err != nil {
			return fmt.Errorf("failed to kill running apple container %q before delete: %w", name, err)
		}
	}
	if _, err := m.runCommand("delete", name); err != nil {
		return err
	}
	return nil
}

// Pause is not exposed by Apple container 0.4.1.
func (m *Manager) Pause(vm *types.VM) error {
	return fmt.Errorf("pause is not supported by the apple container runtime")
}

// Resume is not exposed by Apple container 0.4.1.
func (m *Manager) Resume(vm *types.VM) error {
	return fmt.Errorf("resume is not supported by the apple container runtime")
}

// CreateSnapshot is not exposed by Apple container 0.4.1.
func (m *Manager) CreateSnapshot(vm *types.VM, snapshotPath, memPath string) error {
	return fmt.Errorf("snapshots are not supported by the apple container runtime")
}

// GetConsoleLogs returns container stdio logs.
func (m *Manager) GetConsoleLogs(vm *types.VM, tailLines int) ([]byte, error) {
	name := runtimeName(vm)
	if name == "" {
		if vm.Runtime != nil && vm.Runtime.ConsolePath != "" {
			return os.ReadFile(vm.Runtime.ConsolePath)
		}
		return nil, fmt.Errorf("apple container runtime id not available")
	}
	args := []string{"logs"}
	if tailLines > 0 {
		args = append(args, "-n", strconv.Itoa(tailLines))
	}
	args = append(args, name)
	return m.runCommand(args...)
}

// CleanupNetwork is handled by Apple's container runtime.
func (m *Manager) CleanupNetwork(vm *types.VM) error {
	return nil
}

func runtimeName(vm *types.VM) string {
	if vm == nil {
		return ""
	}
	if vm.Runtime != nil && vm.Runtime.ExternalID != "" {
		return vm.Runtime.ExternalID
	}
	if vm.ID != "" {
		return containerNameForVM(vm.ID)
	}
	return ""
}

func appendConsole(path string, data []byte) error {
	if len(bytes.TrimSpace(data)) == 0 {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(append(data, '\n'))
	return err
}

func maxInt(value, min int) int {
	if value < min {
		return min
	}
	return value
}
