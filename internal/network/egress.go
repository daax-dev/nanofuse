package network

import (
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/daax-dev/nanofuse/internal/types"
)

type commandRunner interface {
	Run(name string, args ...string) error
}

type execRunner struct{}

func (execRunner) Run(name string, args ...string) error {
	return exec.Command(name, args...).Run() //nolint:gosec // daemon-controlled iptables/sysctl invocation
}

// SetupEgressPolicy installs per-VM outbound firewall policy.
func SetupEgressPolicy(vmID, tapDevice, vmIP, gateway string, policy *types.EgressPolicy) error {
	return setupEgressPolicy(execRunner{}, vmID, tapDevice, vmIP, gateway, policy)
}

// CleanupEgressPolicy removes per-VM outbound firewall policy.
func CleanupEgressPolicy(vmID, tapDevice, vmIP string) error {
	return cleanupEgressPolicy(execRunner{}, vmID, tapDevice, vmIP)
}

func setupEgressPolicy(runner commandRunner, vmID, tapDevice, vmIP, gateway string, policy *types.EgressPolicy) error {
	if policy == nil || !policy.Enabled {
		return nil
	}
	if err := validateEgressPolicy(policy); err != nil {
		return err
	}
	if vmID == "" || vmIP == "" {
		return fmt.Errorf("vmID and vmIP are required for egress policy")
	}

	if err := enableBridgeNetfilter(runner); err != nil {
		return fmt.Errorf("failed to enable bridge netfilter for egress policy: %w", err)
	}

	chain := egressChainName(vmID)
	_ = runner.Run("iptables", "-N", chain)
	if err := runner.Run("iptables", "-F", chain); err != nil {
		return fmt.Errorf("failed to flush egress chain %s: %w", chain, err)
	}

	if err := installEgressJumpRules(runner, chain, tapDevice, vmIP); err != nil {
		_ = cleanupEgressPolicy(runner, vmID, tapDevice, vmIP)
		return err
	}

	if policy.AllowDNS {
		if err := appendAllowRule(runner, chain, "udp", gateway, 53); err != nil {
			_ = cleanupEgressPolicy(runner, vmID, tapDevice, vmIP)
			return err
		}
		if err := appendAllowRule(runner, chain, "tcp", gateway, 53); err != nil {
			_ = cleanupEgressPolicy(runner, vmID, tapDevice, vmIP)
			return err
		}
	}

	if policy.Proxy != nil {
		proto := normalizedProtocol(policy.Proxy.Protocol)
		if err := appendAllowRule(runner, chain, proto, policy.Proxy.IP, policy.Proxy.Port); err != nil {
			_ = cleanupEgressPolicy(runner, vmID, tapDevice, vmIP)
			return err
		}
	}

	if !policy.ProxyOnly {
		for _, rule := range policy.AllowRules {
			if err := appendAllowRule(runner, chain, rule.Protocol, rule.CIDR, rule.Port); err != nil {
				_ = cleanupEgressPolicy(runner, vmID, tapDevice, vmIP)
				return err
			}
		}
	}

	if egressDefaultAction(policy) == "allow" {
		if err := runner.Run("iptables", "-A", chain, "-j", "ACCEPT"); err != nil {
			_ = cleanupEgressPolicy(runner, vmID, tapDevice, vmIP)
			return fmt.Errorf("failed to append egress accept rule: %w", err)
		}
		return nil
	}

	if err := runner.Run("iptables", "-A", chain, "-j", "DROP"); err != nil {
		_ = cleanupEgressPolicy(runner, vmID, tapDevice, vmIP)
		return fmt.Errorf("failed to append egress drop rule: %w", err)
	}
	return nil
}

func cleanupEgressPolicy(runner commandRunner, vmID, tapDevice, vmIP string) error {
	if vmID == "" {
		return nil
	}

	chain := egressChainName(vmID)
	var lastErr error
	for _, args := range egressJumpRuleArgs(chain, tapDevice, vmIP) {
		deleteArgs := append([]string{"-D", "FORWARD"}, args...)
		if err := runner.Run("iptables", deleteArgs...); err != nil {
			lastErr = err
		}
	}
	if err := runner.Run("iptables", "-F", chain); err != nil {
		lastErr = err
	}
	if err := runner.Run("iptables", "-X", chain); err != nil {
		lastErr = err
	}
	return lastErr
}

func enableBridgeNetfilter(runner commandRunner) error {
	// modprobe can fail when br_netfilter is built in; sysctl is authoritative.
	_ = runner.Run("modprobe", "br_netfilter")
	return runner.Run("sysctl", "-w", "net.bridge.bridge-nf-call-iptables=1")
}

func installEgressJumpRules(runner commandRunner, chain, tapDevice, vmIP string) error {
	for _, args := range egressJumpRuleArgs(chain, tapDevice, vmIP) {
		deleteArgs := append([]string{"-D", "FORWARD"}, args...)
		_ = runner.Run("iptables", deleteArgs...)

		insertArgs := append([]string{"-I", "FORWARD", "1"}, args...)
		if err := runner.Run("iptables", insertArgs...); err != nil {
			return fmt.Errorf("failed to install egress jump rule: %w", err)
		}
	}
	return nil
}

func egressJumpRuleArgs(chain, tapDevice, vmIP string) [][]string {
	rules := [][]string{
		{"-i", BridgeName, "-s", vmIP, "-j", chain},
	}
	if tapDevice != "" {
		rules = append(rules, []string{"-m", "physdev", "--physdev-in", tapDevice, "-s", vmIP, "-j", chain})
	}
	return rules
}

func appendAllowRule(runner commandRunner, chain, protocol, destination string, port int) error {
	if err := runner.Run("iptables", "-A", chain,
		"-p", normalizedProtocol(protocol),
		"-d", destination,
		"--dport", fmt.Sprintf("%d", port),
		"-j", "ACCEPT"); err != nil {
		return fmt.Errorf("failed to append egress allow rule for %s/%s:%d: %w", protocol, destination, port, err)
	}
	return nil
}

func validateEgressPolicy(policy *types.EgressPolicy) error {
	action := egressDefaultAction(policy)
	if action != "deny" && action != "allow" {
		return fmt.Errorf("egress default_action must be deny or allow")
	}
	if policy.ProxyOnly && policy.Proxy == nil {
		return fmt.Errorf("egress proxy_only requires proxy")
	}
	if policy.Proxy != nil {
		if err := validateEgressDestination(policy.Proxy.IP); err != nil {
			return fmt.Errorf("invalid egress proxy IP: %w", err)
		}
		if err := validateEgressPort(policy.Proxy.Port); err != nil {
			return fmt.Errorf("invalid egress proxy port: %w", err)
		}
		if err := validateEgressProtocol(normalizedProtocol(policy.Proxy.Protocol)); err != nil {
			return fmt.Errorf("invalid egress proxy protocol: %w", err)
		}
	}
	for i, rule := range policy.AllowRules {
		if err := validateEgressDestination(rule.CIDR); err != nil {
			return fmt.Errorf("invalid egress allow rule %d CIDR: %w", i, err)
		}
		if err := validateEgressPort(rule.Port); err != nil {
			return fmt.Errorf("invalid egress allow rule %d port: %w", i, err)
		}
		if err := validateEgressProtocol(rule.Protocol); err != nil {
			return fmt.Errorf("invalid egress allow rule %d protocol: %w", i, err)
		}
	}
	return nil
}

func validateEgressDestination(destination string) error {
	if destination == "" {
		return fmt.Errorf("destination is required")
	}
	if ip := net.ParseIP(destination); ip != nil {
		return nil
	}
	if _, _, err := net.ParseCIDR(destination); err != nil {
		return err
	}
	return nil
}

func validateEgressPort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("port %d must be between 1 and 65535", port)
	}
	return nil
}

func validateEgressProtocol(protocol string) error {
	switch protocol {
	case "tcp", "udp":
		return nil
	default:
		return fmt.Errorf("protocol %q must be tcp or udp", protocol)
	}
}

func normalizedProtocol(protocol string) string {
	if protocol == "" {
		return "tcp"
	}
	return strings.ToLower(protocol)
}

func egressDefaultAction(policy *types.EgressPolicy) string {
	if policy == nil || policy.DefaultAction == "" {
		return "deny"
	}
	return strings.ToLower(policy.DefaultAction)
}

func egressChainName(vmID string) string {
	clean := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r <= 'Z':
			return r
		case r >= '0' && r <= '9':
			return r
		default:
			return -1
		}
	}, vmID)
	if len(clean) > 12 {
		clean = clean[:12]
	}
	if clean == "" {
		clean = "unknown"
	}
	return "NF-EG-" + clean
}
