package network

import (
	"errors"
	"strings"
	"testing"

	"github.com/daax-dev/nanofuse/internal/types"
)

type recordedCommand struct {
	name string
	args []string
}

type fakeRunner struct {
	commands []recordedCommand
	failOn   string
}

func (f *fakeRunner) Run(name string, args ...string) error {
	cmd := recordedCommand{name: name, args: append([]string{}, args...)}
	f.commands = append(f.commands, cmd)
	if f.failOn != "" && strings.Contains(commandString(cmd), f.failOn) {
		return errors.New("forced failure")
	}
	return nil
}

func commandString(cmd recordedCommand) string {
	return cmd.name + " " + strings.Join(cmd.args, " ")
}

func (f *fakeRunner) contains(fragment string) bool {
	for _, cmd := range f.commands {
		if strings.Contains(commandString(cmd), fragment) {
			return true
		}
	}
	return false
}

func TestSetupEgressPolicyDefaultDeny(t *testing.T) {
	runner := &fakeRunner{}
	policy := &types.EgressPolicy{
		Enabled:       true,
		DefaultAction: "deny",
		AllowDNS:      true,
		AllowRules: []types.EgressRule{
			{CIDR: "203.0.113.10", Protocol: "tcp", Port: 443},
		},
	}

	err := setupEgressPolicy(
		runner,
		"550e8400-e29b-41d4-a716-446655440000",
		"tap-550e8400",
		"172.16.0.10",
		"172.16.0.1",
		policy,
	)
	if err != nil {
		t.Fatalf("setup egress policy: %v", err)
	}

	assertCommandContains(t, runner, "sysctl -w net.bridge.bridge-nf-call-iptables=1")
	assertCommandContains(t, runner, "iptables -N NF-EG-550e8400e29b")
	assertCommandContains(t, runner, "iptables -I FORWARD 1 -i nanofuse0 -s 172.16.0.10 -j NF-EG-550e8400e29b")
	assertCommandContains(t, runner, "iptables -I FORWARD 1 -m physdev --physdev-in tap-550e8400 -s 172.16.0.10 -j NF-EG-550e8400e29b")
	assertCommandContains(t, runner, "iptables -A NF-EG-550e8400e29b -p udp -d 172.16.0.1 --dport 53 -j ACCEPT")
	assertCommandContains(t, runner, "iptables -A NF-EG-550e8400e29b -p tcp -d 203.0.113.10 --dport 443 -j ACCEPT")
	assertCommandContains(t, runner, "iptables -A NF-EG-550e8400e29b -j DROP")
}

func TestSetupEgressPolicyProxyOnly(t *testing.T) {
	runner := &fakeRunner{}
	policy := &types.EgressPolicy{
		Enabled:   true,
		ProxyOnly: true,
		Proxy: &types.EgressProxy{
			IP:   "172.16.0.1",
			Port: 3128,
		},
		AllowRules: []types.EgressRule{
			{CIDR: "203.0.113.10", Protocol: "tcp", Port: 443},
		},
	}

	if err := setupEgressPolicy(runner, "vm-1", "tap-vm1", "172.16.0.11", "172.16.0.1", policy); err != nil {
		t.Fatalf("setup egress policy: %v", err)
	}

	assertCommandContains(t, runner, "iptables -A NF-EG-vm1 -p tcp -d 172.16.0.1 --dport 3128 -j ACCEPT")
	if runner.contains("203.0.113.10") {
		t.Fatalf("proxy-only policy installed direct upstream allow rule")
	}
	assertCommandContains(t, runner, "iptables -A NF-EG-vm1 -j DROP")
}

func TestSetupEgressPolicyRejectsProxyOnlyWithoutProxy(t *testing.T) {
	err := setupEgressPolicy(&fakeRunner{}, "vm-1", "tap-vm1", "172.16.0.11", "172.16.0.1", &types.EgressPolicy{
		Enabled:   true,
		ProxyOnly: true,
	})
	if err == nil || !strings.Contains(err.Error(), "proxy_only requires proxy") {
		t.Fatalf("error = %v, want proxy_only validation", err)
	}
}

func TestCleanupEgressPolicyRemovesJumpAndChain(t *testing.T) {
	runner := &fakeRunner{}
	if err := cleanupEgressPolicy(runner, "vm-1", "tap-vm1", "172.16.0.11"); err != nil {
		t.Fatalf("cleanup egress policy: %v", err)
	}

	assertCommandContains(t, runner, "iptables -D FORWARD -i nanofuse0 -s 172.16.0.11 -j NF-EG-vm1")
	assertCommandContains(t, runner, "iptables -D FORWARD -m physdev --physdev-in tap-vm1 -s 172.16.0.11 -j NF-EG-vm1")
	assertCommandContains(t, runner, "iptables -F NF-EG-vm1")
	assertCommandContains(t, runner, "iptables -X NF-EG-vm1")
}

func assertCommandContains(t *testing.T, runner *fakeRunner, fragment string) {
	t.Helper()
	if runner.contains(fragment) {
		return
	}
	var commands []string
	for _, cmd := range runner.commands {
		commands = append(commands, commandString(cmd))
	}
	t.Fatalf("missing command containing %q\ncommands:\n%s", fragment, strings.Join(commands, "\n"))
}
