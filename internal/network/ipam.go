package network

import (
	"fmt"
	"sync"
)

// IPAM (IP Address Management) manages IP address allocation for VMs
type IPAM struct {
	mu        sync.Mutex
	allocated map[string]string // vmID -> IP
	available []string          // Available IPs in the pool
}

// NewIPAM creates a new IPAM instance
// IP Range: 172.16.0.10 - 172.16.0.254 (245 addresses)
// 172.16.0.1 is reserved for the bridge/gateway
func NewIPAM() *IPAM {
	ipam := &IPAM{
		allocated: make(map[string]string),
		available: make([]string, 0, 245),
	}

	// Generate IP pool: 172.16.0.10 - 172.16.0.254
	for i := 10; i <= 254; i++ {
		ipam.available = append(ipam.available, fmt.Sprintf("172.16.0.%d", i))
	}

	return ipam
}

// AllocateIP allocates an IP address for a VM
// Returns the IP address (without CIDR) and error if pool is exhausted
func (ipam *IPAM) AllocateIP(vmID string) (string, error) {
	ipam.mu.Lock()
	defer ipam.mu.Unlock()

	// Check if already allocated
	if ip, exists := ipam.allocated[vmID]; exists {
		return ip, nil
	}

	// Check if pool is exhausted
	if len(ipam.available) == 0 {
		return "", fmt.Errorf("IP address pool exhausted")
	}

	// Allocate first available IP
	ip := ipam.available[0]
	ipam.available = ipam.available[1:]
	ipam.allocated[vmID] = ip

	return ip, nil
}

// ReleaseIP releases an IP address when a VM is deleted
func (ipam *IPAM) ReleaseIP(vmID string) {
	ipam.mu.Lock()
	defer ipam.mu.Unlock()

	if ip, exists := ipam.allocated[vmID]; exists {
		delete(ipam.allocated, vmID)
		ipam.available = append(ipam.available, ip)
	}
}

// GetAllocatedIP returns the allocated IP for a VM, if any
func (ipam *IPAM) GetAllocatedIP(vmID string) (string, bool) {
	ipam.mu.Lock()
	defer ipam.mu.Unlock()

	ip, exists := ipam.allocated[vmID]
	return ip, exists
}

// GetAvailableCount returns the number of available IPs
func (ipam *IPAM) GetAvailableCount() int {
	ipam.mu.Lock()
	defer ipam.mu.Unlock()

	return len(ipam.available)
}

// GetAllocatedCount returns the number of allocated IPs
func (ipam *IPAM) GetAllocatedCount() int {
	ipam.mu.Lock()
	defer ipam.mu.Unlock()

	return len(ipam.allocated)
}

// GetAllAllocations returns a copy of all allocations
func (ipam *IPAM) GetAllAllocations() map[string]string {
	ipam.mu.Lock()
	defer ipam.mu.Unlock()

	allocations := make(map[string]string, len(ipam.allocated))
	for vmID, ip := range ipam.allocated {
		allocations[vmID] = ip
	}

	return allocations
}

// LoadAllocations loads existing allocations (for daemon restart)
func (ipam *IPAM) LoadAllocations(allocations map[string]string) {
	ipam.mu.Lock()
	defer ipam.mu.Unlock()

	// Clear current state
	ipam.allocated = make(map[string]string)
	ipam.available = make([]string, 0)

	// Rebuild available pool
	allocated := make(map[string]bool)
	for _, ip := range allocations {
		allocated[ip] = true
	}

	for i := 10; i <= 254; i++ {
		ip := fmt.Sprintf("172.16.0.%d", i)
		if !allocated[ip] {
			ipam.available = append(ipam.available, ip)
		}
	}

	// Copy allocations
	for vmID, ip := range allocations {
		ipam.allocated[vmID] = ip
	}
}
