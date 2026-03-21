package network

import (
	"fmt"
	"sync"
	"testing"
)

// TestNewIPAM verifies IPAM initialization
func TestNewIPAM(t *testing.T) {
	ipam := NewIPAM()

	if ipam == nil {
		t.Fatal("NewIPAM returned nil")
	}

	// Verify pool size: 172.16.0.10 - 172.16.0.254 = 245 addresses
	expectedSize := 245
	if ipam.GetAvailableCount() != expectedSize {
		t.Errorf("Expected %d available IPs, got %d", expectedSize, ipam.GetAvailableCount())
	}

	if ipam.GetAllocatedCount() != 0 {
		t.Errorf("Expected 0 allocated IPs, got %d", ipam.GetAllocatedCount())
	}
}

// TestAllocateIP tests basic IP allocation
func TestAllocateIP(t *testing.T) {
	ipam := NewIPAM()

	vmID := "vm-test-001"
	ip, err := ipam.AllocateIP(vmID)
	if err != nil {
		t.Fatalf("AllocateIP failed: %v", err)
	}

	// First IP should be 172.16.0.10
	expectedIP := "172.16.0.10"
	if ip != expectedIP {
		t.Errorf("Expected first IP to be %s, got %s", expectedIP, ip)
	}

	// Verify counts updated
	if ipam.GetAllocatedCount() != 1 {
		t.Errorf("Expected 1 allocated IP, got %d", ipam.GetAllocatedCount())
	}

	if ipam.GetAvailableCount() != 244 {
		t.Errorf("Expected 244 available IPs, got %d", ipam.GetAvailableCount())
	}

	// Verify allocation is tracked
	allocatedIP, exists := ipam.GetAllocatedIP(vmID)
	if !exists {
		t.Error("Allocation not tracked")
	}
	if allocatedIP != ip {
		t.Errorf("Tracked IP mismatch: expected %s, got %s", ip, allocatedIP)
	}
}

// TestAllocateIPIdempotent verifies that repeated allocation returns same IP
func TestAllocateIPIdempotent(t *testing.T) {
	ipam := NewIPAM()

	vmID := "vm-test-002"

	// First allocation
	ip1, err := ipam.AllocateIP(vmID)
	if err != nil {
		t.Fatalf("First AllocateIP failed: %v", err)
	}

	// Second allocation for same VM should return same IP
	ip2, err := ipam.AllocateIP(vmID)
	if err != nil {
		t.Fatalf("Second AllocateIP failed: %v", err)
	}

	if ip1 != ip2 {
		t.Errorf("Idempotent allocation failed: first=%s, second=%s", ip1, ip2)
	}

	// Should still show only 1 allocation
	if ipam.GetAllocatedCount() != 1 {
		t.Errorf("Expected 1 allocated IP after idempotent allocation, got %d", ipam.GetAllocatedCount())
	}
}

// TestReleaseIP tests IP release
func TestReleaseIP(t *testing.T) {
	ipam := NewIPAM()

	vmID := "vm-test-003"
	ip, err := ipam.AllocateIP(vmID)
	if err != nil {
		t.Fatalf("AllocateIP failed: %v", err)
	}

	// Release the IP
	ipam.ReleaseIP(vmID)

	// Verify counts updated
	if ipam.GetAllocatedCount() != 0 {
		t.Errorf("Expected 0 allocated IPs after release, got %d", ipam.GetAllocatedCount())
	}

	if ipam.GetAvailableCount() != 245 {
		t.Errorf("Expected 245 available IPs after release, got %d", ipam.GetAvailableCount())
	}

	// Verify IP is no longer tracked
	_, exists := ipam.GetAllocatedIP(vmID)
	if exists {
		t.Error("IP still tracked after release")
	}

	// Released IP should be returned to pool (at the end)
	// Allocate 244 more IPs, then the 245th should be the released one
	for i := 0; i < 244; i++ {
		_, err := ipam.AllocateIP(fmt.Sprintf("vm-fill-%d", i))
		if err != nil {
			t.Fatalf("Allocation %d failed: %v", i, err)
		}
	}

	// This allocation should return the previously released IP
	lastIP, err := ipam.AllocateIP("vm-last")
	if err != nil {
		t.Fatalf("Last allocation failed: %v", err)
	}

	if lastIP != ip {
		t.Errorf("Released IP was not returned to pool: expected %s, got %s", ip, lastIP)
	}
}

// TestReleaseIPNonExistent verifies releasing non-existent VM is safe
func TestReleaseIPNonExistent(t *testing.T) {
	ipam := NewIPAM()

	// Should not panic or cause issues
	ipam.ReleaseIP("non-existent-vm")

	if ipam.GetAllocatedCount() != 0 {
		t.Error("Allocated count changed after releasing non-existent VM")
	}
}

// TestPoolExhaustion tests allocation when pool is exhausted
func TestPoolExhaustion(t *testing.T) {
	ipam := NewIPAM()

	// Allocate all 245 IPs
	for i := 0; i < 245; i++ {
		vmID := fmt.Sprintf("vm-exhaust-%d", i)
		_, err := ipam.AllocateIP(vmID)
		if err != nil {
			t.Fatalf("Allocation %d failed unexpectedly: %v", i, err)
		}
	}

	// Verify pool is exhausted
	if ipam.GetAvailableCount() != 0 {
		t.Errorf("Expected 0 available IPs, got %d", ipam.GetAvailableCount())
	}

	// Next allocation should fail
	_, err := ipam.AllocateIP("vm-overflow")
	if err == nil {
		t.Error("Expected error when pool is exhausted, got nil")
	}
}

// TestLoadAllocations tests restoring IPAM state after restart
func TestLoadAllocations(t *testing.T) {
	ipam := NewIPAM()

	// Simulate existing allocations (as if loaded from database)
	existingAllocations := map[string]string{
		"vm-existing-1": "172.16.0.50",
		"vm-existing-2": "172.16.0.100",
		"vm-existing-3": "172.16.0.150",
	}

	ipam.LoadAllocations(existingAllocations)

	// Verify allocations are restored
	if ipam.GetAllocatedCount() != 3 {
		t.Errorf("Expected 3 allocated IPs after load, got %d", ipam.GetAllocatedCount())
	}

	// Verify specific allocations
	for vmID, expectedIP := range existingAllocations {
		ip, exists := ipam.GetAllocatedIP(vmID)
		if !exists {
			t.Errorf("VM %s not found in allocations", vmID)
			continue
		}
		if ip != expectedIP {
			t.Errorf("VM %s: expected IP %s, got %s", vmID, expectedIP, ip)
		}
	}

	// Verify those IPs are not available for new allocations
	// Available pool should be 245 - 3 = 242
	if ipam.GetAvailableCount() != 242 {
		t.Errorf("Expected 242 available IPs, got %d", ipam.GetAvailableCount())
	}

	// New allocations should not conflict with existing ones
	for i := 0; i < 10; i++ {
		ip, err := ipam.AllocateIP(fmt.Sprintf("vm-new-%d", i))
		if err != nil {
			t.Fatalf("New allocation %d failed: %v", i, err)
		}

		// Verify no conflict with existing allocations
		for existingVM, existingIP := range existingAllocations {
			if ip == existingIP {
				t.Errorf("New allocation conflicts with existing: new=%s, existing=%s (VM %s)",
					ip, existingIP, existingVM)
			}
		}
	}
}

// TestLoadAllocationsEmpty tests loading empty allocations
func TestLoadAllocationsEmpty(t *testing.T) {
	ipam := NewIPAM()

	// Pre-allocate some IPs
	ipam.AllocateIP("vm-pre-1")
	ipam.AllocateIP("vm-pre-2")

	// Load empty allocations (simulating restart with no existing VMs)
	ipam.LoadAllocations(map[string]string{})

	// Should reset to full pool
	if ipam.GetAllocatedCount() != 0 {
		t.Errorf("Expected 0 allocated IPs after loading empty, got %d", ipam.GetAllocatedCount())
	}

	if ipam.GetAvailableCount() != 245 {
		t.Errorf("Expected 245 available IPs after loading empty, got %d", ipam.GetAvailableCount())
	}
}

// TestGetAllAllocations tests retrieval of all allocations
func TestGetAllAllocations(t *testing.T) {
	ipam := NewIPAM()

	// Make some allocations
	ipam.AllocateIP("vm-1")
	ipam.AllocateIP("vm-2")
	ipam.AllocateIP("vm-3")

	allocations := ipam.GetAllAllocations()

	if len(allocations) != 3 {
		t.Errorf("Expected 3 allocations, got %d", len(allocations))
	}

	// Verify we got a copy (modifying returned map shouldn't affect IPAM)
	allocations["vm-fake"] = "172.16.0.200"
	_, exists := ipam.GetAllocatedIP("vm-fake")
	if exists {
		t.Error("Modifying returned map affected internal state")
	}
}

// TestConcurrentAllocations tests thread-safety of IPAM
func TestConcurrentAllocations(t *testing.T) {
	ipam := NewIPAM()

	var wg sync.WaitGroup
	numGoroutines := 50
	allocatedIPs := make(chan string, numGoroutines)

	// Allocate IPs concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			vmID := fmt.Sprintf("vm-concurrent-%d", id)
			ip, err := ipam.AllocateIP(vmID)
			if err != nil {
				t.Errorf("Concurrent allocation %d failed: %v", id, err)
				return
			}
			allocatedIPs <- ip
		}(i)
	}

	wg.Wait()
	close(allocatedIPs)

	// Verify all IPs are unique
	seen := make(map[string]bool)
	for ip := range allocatedIPs {
		if seen[ip] {
			t.Errorf("Duplicate IP allocated: %s", ip)
		}
		seen[ip] = true
	}

	// Verify correct count
	if ipam.GetAllocatedCount() != numGoroutines {
		t.Errorf("Expected %d allocations, got %d", numGoroutines, ipam.GetAllocatedCount())
	}
}

// TestConcurrentReleases tests thread-safety of IP release
func TestConcurrentReleases(t *testing.T) {
	ipam := NewIPAM()

	// Pre-allocate IPs
	numVMs := 50
	for i := 0; i < numVMs; i++ {
		ipam.AllocateIP(fmt.Sprintf("vm-release-%d", i))
	}

	var wg sync.WaitGroup

	// Release IPs concurrently
	for i := 0; i < numVMs; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ipam.ReleaseIP(fmt.Sprintf("vm-release-%d", id))
		}(i)
	}

	wg.Wait()

	// Verify all IPs released
	if ipam.GetAllocatedCount() != 0 {
		t.Errorf("Expected 0 allocations after concurrent release, got %d", ipam.GetAllocatedCount())
	}

	if ipam.GetAvailableCount() != 245 {
		t.Errorf("Expected 245 available after concurrent release, got %d", ipam.GetAvailableCount())
	}
}

// TestIPAddressFormat verifies IP addresses are in expected format
func TestIPAddressFormat(t *testing.T) {
	ipam := NewIPAM()

	ip, _ := ipam.AllocateIP("vm-format-test")

	// IP should be in 172.16.0.x format (x from 10-254)
	expectedPrefix := "172.16.0."
	if len(ip) < len(expectedPrefix) || ip[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("IP %s does not have expected prefix %s", ip, expectedPrefix)
	}
}
