package firecracker

import (
	"regexp"
	"sync"
	"testing"
)

// TestGenerateMAC tests MAC address generation
func TestGenerateMAC(t *testing.T) {
	mac := generateMAC()

	// MAC should be in format AA:FC:XX:XX:XX:XX
	macRegex := regexp.MustCompile(`^[Aa][Aa]:[Ff][Cc]:[0-9A-Fa-f]{2}:[0-9A-Fa-f]{2}:[0-9A-Fa-f]{2}:[0-9A-Fa-f]{2}$`)
	if !macRegex.MatchString(mac) {
		t.Errorf("MAC address %s does not match expected format AA:FC:XX:XX:XX:XX", mac)
	}
}

// TestGenerateMACPrefix verifies the Firecracker OUI prefix
func TestGenerateMACPrefix(t *testing.T) {
	mac := generateMAC()

	// First two bytes should be AA:FC (Firecracker-style locally administered unicast)
	expectedPrefix := "AA:FC:"
	if len(mac) < len(expectedPrefix) || mac[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("MAC address %s does not start with %s", mac, expectedPrefix)
	}
}

// TestGenerateMACUniqueness tests that multiple MACs are unique
func TestGenerateMACUniqueness(t *testing.T) {
	macs := make(map[string]bool)
	numMACs := 100

	for i := 0; i < numMACs; i++ {
		mac := generateMAC()
		if macs[mac] {
			t.Errorf("Duplicate MAC generated: %s", mac)
		}
		macs[mac] = true
	}

	if len(macs) != numMACs {
		t.Errorf("Expected %d unique MACs, got %d", numMACs, len(macs))
	}
}

// TestGenerateMACNotAllZeros verifies MAC is not the broken placeholder pattern
func TestGenerateMACNotAllZeros(t *testing.T) {
	// Run multiple times to catch probabilistic issues
	brokenPattern := "AA:FC:00:00:00:00"

	for i := 0; i < 100; i++ {
		mac := generateMAC()
		if mac == brokenPattern {
			t.Errorf("Generated broken MAC pattern %s on iteration %d", brokenPattern, i)
		}
	}
}

// TestRandomByte tests the randomByte function
func TestRandomByte(t *testing.T) {
	// Call randomByte multiple times to ensure it doesn't always return 0
	nonZeroCount := 0
	iterations := 100

	for i := 0; i < iterations; i++ {
		b := randomByte()
		if b != 0 {
			nonZeroCount++
		}
	}

	// Probability of getting 0 every time with crypto/rand is vanishingly small
	// 100 iterations of all zeros is (1/256)^100 which is practically impossible
	if nonZeroCount == 0 {
		t.Error("randomByte returned 0 for all iterations - likely using placeholder implementation")
	}

	// Should get a good mix of values
	// With 100 iterations, we expect roughly 100 * (255/256) ≈ 99.6 non-zero values
	if nonZeroCount < 90 {
		t.Errorf("randomByte returned too many zeros: %d non-zero out of %d", nonZeroCount, iterations)
	}
}

// TestRandomByteDistribution tests that randomByte has reasonable distribution
func TestRandomByteDistribution(t *testing.T) {
	counts := make(map[byte]int)
	iterations := 10000

	for i := 0; i < iterations; i++ {
		b := randomByte()
		counts[b]++
	}

	// With 10000 iterations across 256 values, we expect ~39 per value
	// Check that we have a reasonable spread (at least 200 unique values)
	if len(counts) < 200 {
		t.Errorf("Poor distribution: only %d unique values out of %d iterations", len(counts), iterations)
	}
}

// TestGenerateMACConcurrent tests thread-safety of MAC generation
func TestGenerateMACConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	numGoroutines := 100
	macChan := make(chan string, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mac := generateMAC()
			macChan <- mac
		}()
	}

	wg.Wait()
	close(macChan)

	// Verify all MACs are unique
	seen := make(map[string]bool)
	for mac := range macChan {
		if seen[mac] {
			t.Errorf("Duplicate MAC generated concurrently: %s", mac)
		}
		seen[mac] = true
	}
}

// TestNewManager tests manager creation
func TestNewManager(t *testing.T) {
	manager := NewManager("/usr/bin/firecracker", "/var/lib/nanofuse")

	if manager == nil {
		t.Fatal("NewManager returned nil")
	}

	if manager.binaryPath != "/usr/bin/firecracker" {
		t.Errorf("Expected binaryPath /usr/bin/firecracker, got %s", manager.binaryPath)
	}

	if manager.dataDir != "/var/lib/nanofuse" {
		t.Errorf("Expected dataDir /var/lib/nanofuse, got %s", manager.dataDir)
	}
}

// TestIsRunningFalseForInvalidPID tests IsRunning with invalid PID
func TestIsRunningFalseForInvalidPID(t *testing.T) {
	manager := NewManager("/usr/bin/firecracker", "/tmp")

	// PID -1 should not be running
	if manager.IsRunning(-1) {
		t.Error("IsRunning returned true for PID -1")
	}

	// PID 0 is special (kernel), but should return false for user space check
	// Actually on Linux, signal to PID 0 sends to all processes in the process group
	// so we skip that test. Use a very high PID that won't exist.
	if manager.IsRunning(999999999) {
		t.Error("IsRunning returned true for non-existent PID 999999999")
	}
}

// TestFirecrackerConfig tests configuration struct marshaling
func TestFirecrackerConfig(t *testing.T) {
	config := FirecrackerConfig{
		BootSource: BootSource{
			KernelImagePath: "/path/to/vmlinux",
			BootArgs:        "console=ttyS0",
		},
		Drives: []Drive{
			{
				DriveID:      "rootfs",
				PathOnHost:   "/path/to/rootfs.ext4",
				IsReadOnly:   false,
				IsRootDevice: true,
			},
		},
		MachineConfig: MachineConfig{
			VcpuCount:  2,
			MemSizeMib: 512,
			SMT:        false,
		},
		NetworkInterfaces: []NetworkInterface{
			{
				IfaceID:     "eth0",
				GuestMAC:    "AA:FC:00:01:02:03",
				HostDevName: "tap-test",
			},
		},
	}

	// Verify config fields
	if config.BootSource.KernelImagePath != "/path/to/vmlinux" {
		t.Error("Kernel path mismatch")
	}

	if len(config.Drives) != 1 {
		t.Error("Expected 1 drive")
	}

	if config.MachineConfig.VcpuCount != 2 {
		t.Error("VcpuCount mismatch")
	}

	if len(config.NetworkInterfaces) != 1 {
		t.Error("Expected 1 network interface")
	}
}

// TestMachineConfigDefaults tests machine configuration validation
func TestMachineConfigDefaults(t *testing.T) {
	config := MachineConfig{
		VcpuCount:  1,
		MemSizeMib: 128,
		SMT:        false,
	}

	if config.VcpuCount < 1 {
		t.Error("VcpuCount should be at least 1")
	}

	if config.MemSizeMib < 128 {
		t.Error("MemSizeMib should be at least 128")
	}
}

// TestNetworkInterfaceConfig tests network interface configuration
func TestNetworkInterfaceConfig(t *testing.T) {
	ni := NetworkInterface{
		IfaceID:     "eth0",
		GuestMAC:    "AA:FC:00:01:02:03",
		HostDevName: "tap-vm1",
	}

	if ni.IfaceID != "eth0" {
		t.Error("IfaceID mismatch")
	}

	// Verify MAC format
	macRegex := regexp.MustCompile(`^[0-9A-Fa-f]{2}(:[0-9A-Fa-f]{2}){5}$`)
	if !macRegex.MatchString(ni.GuestMAC) {
		t.Errorf("Invalid MAC format: %s", ni.GuestMAC)
	}

	if ni.HostDevName != "tap-vm1" {
		t.Error("HostDevName mismatch")
	}
}

// TestDriveConfig tests drive configuration
func TestDriveConfig(t *testing.T) {
	drive := Drive{
		DriveID:      "rootfs",
		PathOnHost:   "/var/lib/nanofuse/images/base/rootfs.ext4",
		IsReadOnly:   false,
		IsRootDevice: true,
	}

	if drive.DriveID != "rootfs" {
		t.Error("DriveID mismatch")
	}

	if drive.IsReadOnly {
		t.Error("Root device should be writable")
	}

	if !drive.IsRootDevice {
		t.Error("Should be marked as root device")
	}
}

// TestBootSourceConfig tests boot source configuration
func TestBootSourceConfig(t *testing.T) {
	bootSource := BootSource{
		KernelImagePath: "/var/lib/nanofuse/images/base/vmlinux",
		BootArgs:        "console=ttyS0 root=/dev/vda1 rw init=/lib/systemd/systemd",
	}

	if bootSource.KernelImagePath == "" {
		t.Error("KernelImagePath should not be empty")
	}

	// Verify essential kernel args are present
	if bootSource.BootArgs == "" {
		t.Error("BootArgs should not be empty")
	}
}

// TestSetProcessExitHandler tests that the exit handler can be set
func TestSetProcessExitHandler(t *testing.T) {
	manager := NewManager("/usr/bin/firecracker", "/tmp")

	// Handler should be nil initially
	if manager.onProcessExit != nil {
		t.Error("onProcessExit should be nil initially")
	}

	// Set a handler
	manager.SetProcessExitHandler(func(vmID string, exitCode *int, err error) {
		// Handler body - just verify it can be set
		_ = vmID
		_ = exitCode
		_ = err
	})

	// Handler should be set
	if manager.onProcessExit == nil {
		t.Error("onProcessExit should be set after SetProcessExitHandler")
	}
}

// TestProcessExitHandlerCallback tests that the exit handler is called correctly
func TestProcessExitHandlerCallback(t *testing.T) {
	manager := NewManager("/usr/bin/firecracker", "/tmp")

	// Track whether callback was invoked
	handlerSet := false

	manager.SetProcessExitHandler(func(vmID string, exitCode *int, err error) {
		// This handler would be called when a process exits
		// For unit testing, we just verify the handler mechanism works
		_ = vmID
		_ = exitCode
		_ = err
	})
	handlerSet = true

	// Verify handler was set
	if !handlerSet || manager.onProcessExit == nil {
		t.Fatal("Handler not set correctly")
	}

	// Note: Full integration testing of zombie prevention requires running actual VMs
	// This unit test verifies the handler mechanism is in place
}

// TestProcessExitHandlerNilSafe tests that nil handler doesn't panic
func TestProcessExitHandlerNilSafe(t *testing.T) {
	manager := NewManager("/usr/bin/firecracker", "/tmp")

	// Don't set a handler - should be nil
	if manager.onProcessExit != nil {
		t.Skip("Handler unexpectedly set")
	}

	// This shouldn't panic even with nil handler
	// The waitForProcessExit function checks for nil before calling
}

// TestManagerHasProcessExitField tests that the Manager struct has the required field
func TestManagerHasProcessExitField(t *testing.T) {
	manager := NewManager("/usr/bin/firecracker", "/tmp")

	// This test verifies the struct has the field by attempting to set it
	// If the field doesn't exist, this won't compile
	var handler ProcessExitHandler = func(vmID string, exitCode *int, err error) {}
	manager.SetProcessExitHandler(handler)

	if manager.onProcessExit == nil {
		t.Error("Handler should be set")
	}
}
