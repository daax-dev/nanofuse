package network

import (
	"fmt"
	"os/exec"
)

const (
	// BridgeName is the name of the nanofuse bridge
	BridgeName = "nanofuse0"
	// BridgeIP is the IP address/CIDR of the bridge
	BridgeIP = "172.16.0.1/24"
	// BridgeGateway is the gateway IP (same as bridge IP)
	BridgeGateway = "172.16.0.1"
	// NetworkCIDR is the full network CIDR
	NetworkCIDR = "172.16.0.0/24"
)

// SetupBridge creates and configures the nanofuse bridge
// This should be called once on daemon startup
func SetupBridge() error {
	// Check if bridge already exists
	if BridgeExists() {
		return nil // Already configured
	}

	// Create bridge
	cmd := exec.Command("ip", "link", "add", BridgeName, "type", "bridge")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create bridge %s: %w", BridgeName, err)
	}

	// Assign IP address
	cmd = exec.Command("ip", "addr", "add", BridgeIP, "dev", BridgeName)
	if err := cmd.Run(); err != nil {
		// Cleanup bridge on failure
		_ = exec.Command("ip", "link", "delete", BridgeName).Run()
		return fmt.Errorf("failed to assign IP %s to bridge: %w", BridgeIP, err)
	}

	// Bring bridge up
	cmd = exec.Command("ip", "link", "set", BridgeName, "up")
	if err := cmd.Run(); err != nil {
		_ = exec.Command("ip", "link", "delete", BridgeName).Run()
		return fmt.Errorf("failed to bring up bridge %s: %w", BridgeName, err)
	}

	// Ensure route exists for the bridge network
	// This is critical for localhost port forwarding to work:
	// When OUTPUT chain does DNAT to 172.16.0.x, kernel needs to know how to route there
	// Route is usually auto-created by "ip addr add", but we verify/add it explicitly
	if err := ensureBridgeRoute(); err != nil {
		_ = exec.Command("ip", "link", "delete", BridgeName).Run()
		return fmt.Errorf("failed to setup bridge route: %w", err)
	}

	return nil
}

// ensureBridgeRoute ensures a route exists for the bridge network
func ensureBridgeRoute() error {
	// Check if route already exists
	checkCmd := exec.Command("ip", "route", "show", NetworkCIDR, "dev", BridgeName)
	output, _ := checkCmd.CombinedOutput()

	// If route exists, we're done
	if len(output) > 0 {
		return nil
	}

	// Add route: 172.16.0.0/24 dev nanofuse0
	// This tells kernel: packets for 172.16.0.0/24 go via nanofuse0
	cmd := exec.Command("ip", "route", "add", NetworkCIDR, "dev", BridgeName)
	if err := cmd.Run(); err != nil {
		// Route might already exist (race condition), check again
		checkCmd := exec.Command("ip", "route", "show", NetworkCIDR, "dev", BridgeName)
		output, _ := checkCmd.CombinedOutput()
		if len(output) > 0 {
			return nil // Route now exists, all good
		}
		return fmt.Errorf("failed to add route for %s: %w", NetworkCIDR, err)
	}

	return nil
}

// BridgeExists checks if the nanofuse bridge exists
func BridgeExists() bool {
	cmd := exec.Command("ip", "link", "show", BridgeName)
	return cmd.Run() == nil
}

// DeleteBridge removes the nanofuse bridge
// WARNING: This will disconnect all running VMs
func DeleteBridge() error {
	if !BridgeExists() {
		return nil
	}

	cmd := exec.Command("ip", "link", "delete", BridgeName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete bridge %s: %w", BridgeName, err)
	}

	return nil
}
