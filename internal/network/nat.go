package network

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// SetupNAT configures IP forwarding and iptables NAT rules
// This enables VMs to access the internet through the host
func SetupNAT(primaryInterface string) error {
	// Enable IP forwarding
	if err := enableIPForwarding(); err != nil {
		return fmt.Errorf("failed to enable IP forwarding: %w", err)
	}

	// Add NAT (MASQUERADE) rule
	if err := addNATRule(primaryInterface); err != nil {
		return fmt.Errorf("failed to add NAT rule: %w", err)
	}

	// Add forwarding rules
	if err := addForwardRules(primaryInterface); err != nil {
		return fmt.Errorf("failed to add forward rules: %w", err)
	}

	return nil
}

// enableIPForwarding enables IPv4 forwarding
func enableIPForwarding() error {
	cmd := exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1")
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// addNATRule adds iptables MASQUERADE rule for NAT (idempotent)
func addNATRule(primaryInterface string) error {
	// Check if rule already exists
	checkCmd := exec.Command("iptables", "-t", "nat", "-C", "POSTROUTING",
		"-s", NetworkCIDR, "-o", primaryInterface, "-j", "MASQUERADE")
	if checkCmd.Run() == nil {
		// Rule already exists, nothing to do
		return nil
	}

	// Try to add the rule
	addCmd := exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING",
		"-s", NetworkCIDR, "-o", primaryInterface, "-j", "MASQUERADE")
	var stderr bytes.Buffer
	addCmd.Stderr = &stderr

	if err := addCmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		// Check if it failed because rule already exists (race condition or -C had permission issue)
		if strings.Contains(errMsg, "already exists") || strings.Contains(errMsg, "duplicate") {
			return nil
		}
		// Re-check if rule exists now (handles race conditions)
		recheckCmd := exec.Command("iptables", "-t", "nat", "-C", "POSTROUTING",
			"-s", NetworkCIDR, "-o", primaryInterface, "-j", "MASQUERADE")
		if recheckCmd.Run() == nil {
			return nil // Rule exists, we're good
		}
		if errMsg != "" {
			return fmt.Errorf("%s: %w", errMsg, err)
		}
		return err
	}
	return nil
}

// addForwardRules adds iptables FORWARD rules (idempotent)
func addForwardRules(primaryInterface string) error {
	// Rule 1: Allow forwarding from bridge to primary interface
	if err := addForwardRuleIdempotent(BridgeName, primaryInterface, ""); err != nil {
		return fmt.Errorf("failed to add forward rule (outbound): %w", err)
	}

	// Rule 2: Allow established connections back from primary interface to bridge
	if err := addForwardRuleIdempotent(primaryInterface, BridgeName, "RELATED,ESTABLISHED"); err != nil {
		return fmt.Errorf("failed to add forward rule (inbound): %w", err)
	}

	return nil
}

// addForwardRuleIdempotent adds a single FORWARD rule, handling "already exists" gracefully
func addForwardRuleIdempotent(inIface, outIface, state string) error {
	args := []string{"-C", "FORWARD", "-i", inIface, "-o", outIface}
	if state != "" {
		args = append(args, "-m", "state", "--state", state)
	}
	args = append(args, "-j", "ACCEPT")

	// Check if rule already exists
	checkCmd := exec.Command("iptables", args...)
	if checkCmd.Run() == nil {
		return nil // Rule exists
	}

	// Try to add the rule
	addArgs := make([]string, len(args))
	copy(addArgs, args)
	addArgs[0] = "-A" // Change -C to -A

	addCmd := exec.Command("iptables", addArgs...)
	var stderr bytes.Buffer
	addCmd.Stderr = &stderr

	if err := addCmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		// Check if it failed because rule already exists
		if strings.Contains(errMsg, "already exists") || strings.Contains(errMsg, "duplicate") {
			return nil
		}
		// Re-check if rule exists now
		recheckCmd := exec.Command("iptables", args...)
		if recheckCmd.Run() == nil {
			return nil
		}
		if errMsg != "" {
			return fmt.Errorf("%s: %w", errMsg, err)
		}
		return err
	}
	return nil
}

// GetPrimaryInterface detects the primary network interface
// Returns the interface name (e.g., "eth0", "wlan0", "ens3")
func GetPrimaryInterface() (string, error) {
	// Get default route
	cmd := exec.Command("ip", "route", "show", "default")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get default route: %w", err)
	}

	// Parse "default via X.X.X.X dev <interface>"
	fields := strings.Fields(string(output))
	for i, field := range fields {
		if field == "dev" && i+1 < len(fields) {
			return fields[i+1], nil
		}
	}

	// Fallback: try common interface names
	commonInterfaces := []string{"eth0", "ens3", "enp0s3", "ens33", "wlan0", "wlp2s0"}
	for _, iface := range commonInterfaces {
		cmd := exec.Command("ip", "link", "show", iface)
		if cmd.Run() == nil {
			return iface, nil
		}
	}

	return "", fmt.Errorf("could not detect primary network interface")
}

// CleanupNAT removes NAT rules (for daemon shutdown)
// WARNING: This will break connectivity for running VMs
func CleanupNAT(primaryInterface string) error {
	// Remove NAT rule
	cmd := exec.Command("iptables", "-t", "nat", "-D", "POSTROUTING",
		"-s", NetworkCIDR, "-o", primaryInterface, "-j", "MASQUERADE")
	_ = cmd.Run() // Ignore errors

	// Remove forward rules
	cmd = exec.Command("iptables", "-D", "FORWARD",
		"-i", BridgeName, "-o", primaryInterface, "-j", "ACCEPT")
	_ = cmd.Run()

	cmd = exec.Command("iptables", "-D", "FORWARD",
		"-i", primaryInterface, "-o", BridgeName,
		"-m", "state", "--state", "RELATED,ESTABLISHED", "-j", "ACCEPT")
	_ = cmd.Run()

	return nil
}

// CheckIPForwardingEnabled checks if IP forwarding is enabled
func CheckIPForwardingEnabled() bool {
	cmd := exec.Command("sysctl", "net.ipv4.ip_forward")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return false
	}
	return strings.Contains(out.String(), "net.ipv4.ip_forward = 1")
}
