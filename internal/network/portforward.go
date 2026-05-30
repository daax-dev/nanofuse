package network

import (
	"fmt"
	"os/exec"

	"github.com/daax-dev/nanofuse/internal/types"
)

// ValidatePortForward validates a port forward configuration
func ValidatePortForward(pf types.PortForward) error {
	// Validate host port
	if pf.HostPort < 1 || pf.HostPort > 65535 {
		return fmt.Errorf("invalid host port %d: must be between 1 and 65535", pf.HostPort)
	}

	// Validate VM port
	if pf.VMPort < 1 || pf.VMPort > 65535 {
		return fmt.Errorf("invalid VM port %d: must be between 1 and 65535", pf.VMPort)
	}

	// Validate protocol
	if pf.Protocol != "tcp" && pf.Protocol != "udp" {
		return fmt.Errorf("invalid protocol %q: must be 'tcp' or 'udp'", pf.Protocol)
	}

	return nil
}

// SetupPortForwards creates iptables DNAT rules for all port forwards
func SetupPortForwards(vmIP string, portForwards []types.PortForward) error {
	for _, pf := range portForwards {
		if err := ValidatePortForward(pf); err != nil {
			return fmt.Errorf("invalid port forward: %w", err)
		}

		if err := addPortForwardRule(vmIP, pf); err != nil {
			// Clean up any rules we already added
			_ = CleanupPortForwards(vmIP, portForwards)
			return fmt.Errorf("failed to add port forward %d:%d: %w", pf.HostPort, pf.VMPort, err)
		}
	}

	return nil
}

// CleanupPortForwards removes all iptables DNAT rules for the given port forwards
func CleanupPortForwards(vmIP string, portForwards []types.PortForward) error {
	var lastErr error

	for _, pf := range portForwards {
		if err := removePortForwardRule(vmIP, pf); err != nil {
			lastErr = err
			// Continue removing other rules even if one fails
		}
	}

	return lastErr
}

// addPortForwardRule adds iptables DNAT rules for a single port forward
func addPortForwardRule(vmIP string, pf types.PortForward) error {
	proto := pf.Protocol
	destination := fmt.Sprintf("%s:%d", vmIP, pf.VMPort)

	// Check if PREROUTING rule already exists
	checkCmd := exec.Command("iptables", "-t", "nat", "-C", "PREROUTING",
		"-p", proto, "--dport", fmt.Sprintf("%d", pf.HostPort),
		"-j", "DNAT", "--to-destination", destination)

	if checkCmd.Run() == nil {
		// Rule already exists, skip
		return nil
	}

	// Add DNAT rule for external connections (PREROUTING chain)
	// Use -I to insert at top, before Docker rules
	preroutingCmd := exec.Command("iptables", "-t", "nat", "-I", "PREROUTING", "1",
		"-p", proto, "--dport", fmt.Sprintf("%d", pf.HostPort),
		"-j", "DNAT", "--to-destination", destination)

	if err := preroutingCmd.Run(); err != nil {
		return fmt.Errorf("failed to add PREROUTING rule: %w", err)
	}

	// Add DNAT rule for localhost connections (OUTPUT chain)
	// This allows connections from the host itself to work
	// CRITICAL: Must match specifically on 127.0.0.1 destination (like Docker does)
	// Matching only on port would affect ALL traffic on that port
	checkOutputCmd := exec.Command("iptables", "-t", "nat", "-C", "OUTPUT",
		"-d", "127.0.0.1",
		"-p", proto, "--dport", fmt.Sprintf("%d", pf.HostPort),
		"-j", "DNAT", "--to-destination", destination)

	if checkOutputCmd.Run() != nil {
		// Rule doesn't exist, add it
		// CRITICAL: Use -I (INSERT) not -A (APPEND) to put rule BEFORE Docker/other rules
		outputCmd := exec.Command("iptables", "-t", "nat", "-I", "OUTPUT", "1",
			"-d", "127.0.0.1",
			"-p", proto, "--dport", fmt.Sprintf("%d", pf.HostPort),
			"-j", "DNAT", "--to-destination", destination)

		if err := outputCmd.Run(); err != nil {
			// Try to remove PREROUTING rule we just added
			_ = exec.Command("iptables", "-t", "nat", "-D", "PREROUTING",
				"-p", proto, "--dport", fmt.Sprintf("%d", pf.HostPort),
				"-j", "DNAT", "--to-destination", destination).Run()
			return fmt.Errorf("failed to add OUTPUT rule for localhost: %w", err)
		}
	}

	// Ensure FORWARD chain allows the traffic
	checkForwardCmd := exec.Command("iptables", "-C", "FORWARD",
		"-p", proto, "-d", vmIP, "--dport", fmt.Sprintf("%d", pf.VMPort),
		"-j", "ACCEPT")

	if checkForwardCmd.Run() != nil {
		// Rule doesn't exist, add it
		forwardCmd := exec.Command("iptables", "-A", "FORWARD",
			"-p", proto, "-d", vmIP, "--dport", fmt.Sprintf("%d", pf.VMPort),
			"-j", "ACCEPT")

		if err := forwardCmd.Run(); err != nil {
			// Try to clean up the rules we just added
			_ = exec.Command("iptables", "-t", "nat", "-D", "PREROUTING",
				"-p", proto, "--dport", fmt.Sprintf("%d", pf.HostPort),
				"-j", "DNAT", "--to-destination", destination).Run()
			_ = exec.Command("iptables", "-t", "nat", "-D", "OUTPUT",
				"-d", "127.0.0.1",
				"-p", proto, "--dport", fmt.Sprintf("%d", pf.HostPort),
				"-j", "DNAT", "--to-destination", destination).Run()
			return fmt.Errorf("failed to add FORWARD rule: %w", err)
		}
	}

	// Add MASQUERADE for general port-forwarded traffic (external connections)
	checkMasqueradeCmd := exec.Command("iptables", "-t", "nat", "-C", "POSTROUTING",
		"-p", proto, "-d", vmIP, "--dport", fmt.Sprintf("%d", pf.VMPort),
		"-j", "MASQUERADE")

	if checkMasqueradeCmd.Run() != nil {
		// Rule doesn't exist, add it
		masqueradeCmd := exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING",
			"-p", proto, "-d", vmIP, "--dport", fmt.Sprintf("%d", pf.VMPort),
			"-j", "MASQUERADE")

		if err := masqueradeCmd.Run(); err != nil {
			// Try to clean up the rules we just added
			_ = exec.Command("iptables", "-t", "nat", "-D", "PREROUTING",
				"-p", proto, "--dport", fmt.Sprintf("%d", pf.HostPort),
				"-j", "DNAT", "--to-destination", destination).Run()
			_ = exec.Command("iptables", "-t", "nat", "-D", "OUTPUT",
				"-d", "127.0.0.1",
				"-p", proto, "--dport", fmt.Sprintf("%d", pf.HostPort),
				"-j", "DNAT", "--to-destination", destination).Run()
			_ = exec.Command("iptables", "-D", "FORWARD",
				"-p", proto, "-d", vmIP, "--dport", fmt.Sprintf("%d", pf.VMPort),
				"-j", "ACCEPT").Run()
			return fmt.Errorf("failed to add MASQUERADE rule: %w", err)
		}
	}

	// For localhost connections, we need explicit SNAT to the bridge IP
	// This is required because locally-originated packets with DNAT in OUTPUT
	// may not properly route responses without source NAT
	// Match on outgoing interface to bridge and destination VM
	bridgeIP := "172.16.0.1"
	bridgeName := "nanofuse0"
	checkSnatCmd := exec.Command("iptables", "-t", "nat", "-C", "POSTROUTING",
		"-o", bridgeName, "-p", proto, "-d", vmIP, "--dport", fmt.Sprintf("%d", pf.VMPort),
		"-j", "SNAT", "--to-source", bridgeIP)

	if checkSnatCmd.Run() != nil {
		// Rule doesn't exist, add it
		snatCmd := exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING",
			"-o", bridgeName, "-p", proto, "-d", vmIP, "--dport", fmt.Sprintf("%d", pf.VMPort),
			"-j", "SNAT", "--to-source", bridgeIP)

		if err := snatCmd.Run(); err != nil {
			// Try to clean up the rules we just added (non-fatal, log and continue)
			_ = exec.Command("iptables", "-t", "nat", "-D", "PREROUTING",
				"-p", proto, "--dport", fmt.Sprintf("%d", pf.HostPort),
				"-j", "DNAT", "--to-destination", destination).Run()
			_ = exec.Command("iptables", "-t", "nat", "-D", "OUTPUT",
				"-d", "127.0.0.1",
				"-p", proto, "--dport", fmt.Sprintf("%d", pf.HostPort),
				"-j", "DNAT", "--to-destination", destination).Run()
			_ = exec.Command("iptables", "-D", "FORWARD",
				"-p", proto, "-d", vmIP, "--dport", fmt.Sprintf("%d", pf.VMPort),
				"-j", "ACCEPT").Run()
			_ = exec.Command("iptables", "-t", "nat", "-D", "POSTROUTING",
				"-p", proto, "-d", vmIP, "--dport", fmt.Sprintf("%d", pf.VMPort),
				"-j", "MASQUERADE").Run()
			return fmt.Errorf("failed to add SNAT rule for localhost: %w", err)
		}
	}

	return nil
}

// removePortForwardRule removes iptables DNAT rules for a single port forward
func removePortForwardRule(vmIP string, pf types.PortForward) error {
	proto := pf.Protocol
	destination := fmt.Sprintf("%s:%d", vmIP, pf.VMPort)

	// Remove PREROUTING rule (ignore errors - rule might not exist)
	_ = exec.Command("iptables", "-t", "nat", "-D", "PREROUTING",
		"-p", proto, "--dport", fmt.Sprintf("%d", pf.HostPort),
		"-j", "DNAT", "--to-destination", destination).Run()

	// Remove OUTPUT rule (ignore errors - rule might not exist)
	_ = exec.Command("iptables", "-t", "nat", "-D", "OUTPUT",
		"-d", "127.0.0.1",
		"-p", proto, "--dport", fmt.Sprintf("%d", pf.HostPort),
		"-j", "DNAT", "--to-destination", destination).Run()

	// Remove MASQUERADE rule (ignore errors - rule might not exist)
	_ = exec.Command("iptables", "-t", "nat", "-D", "POSTROUTING",
		"-p", proto, "-d", vmIP, "--dport", fmt.Sprintf("%d", pf.VMPort),
		"-j", "MASQUERADE").Run()

	// Remove SNAT rule for localhost (ignore errors - rule might not exist)
	bridgeIP := "172.16.0.1"
	bridgeName := "nanofuse0"
	_ = exec.Command("iptables", "-t", "nat", "-D", "POSTROUTING",
		"-o", bridgeName, "-p", proto, "-d", vmIP, "--dport", fmt.Sprintf("%d", pf.VMPort),
		"-j", "SNAT", "--to-source", bridgeIP).Run()

	// Note: We don't remove FORWARD rules as they might be used by other port forwards
	// or by the general NAT setup

	return nil
}
