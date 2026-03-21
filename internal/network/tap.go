package network

import (
	"fmt"
	"os/exec"
)

// CreateTAPDevice creates a TAP network interface
func CreateTAPDevice(name string) error {
	// ip tuntap add <name> mode tap
	cmd := exec.Command("ip", "tuntap", "add", name, "mode", "tap")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create TAP device %s: %w", name, err)
	}

	// ip link set <name> up
	cmd = exec.Command("ip", "link", "set", name, "up")
	if err := cmd.Run(); err != nil {
		// Cleanup on failure
		_ = exec.Command("ip", "link", "delete", name).Run()
		return fmt.Errorf("failed to bring up TAP device %s: %w", name, err)
	}

	return nil
}

// AttachTAPToBridge attaches TAP device to bridge
func AttachTAPToBridge(tapName, bridgeName string) error {
	cmd := exec.Command("ip", "link", "set", tapName, "master", bridgeName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to attach TAP %s to bridge %s: %w", tapName, bridgeName, err)
	}
	return nil
}

// DeleteTAPDevice removes a TAP device
func DeleteTAPDevice(name string) error {
	cmd := exec.Command("ip", "link", "delete", name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete TAP device %s: %w", name, err)
	}
	return nil
}

// TAPDeviceExists checks if a TAP device exists
func TAPDeviceExists(name string) bool {
	cmd := exec.Command("ip", "link", "show", name)
	return cmd.Run() == nil
}
