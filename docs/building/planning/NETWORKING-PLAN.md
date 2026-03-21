# NanoFuse VM Networking - Implementation Plan

**Goal**: Enable full network connectivity for Firecracker VMs with internet access

---

## Current State

❌ **Not Working**:
- TAP device creation is stubbed
- No IP address assignment
- No routing/NAT setup
- VMs boot but have no network connectivity
- Cannot test SSH, curl, wget, etc.

---

## Architecture Decision: NAT Mode

We'll implement **NAT (Network Address Translation)** mode first:

```
Internet
   ↓
Host eth0 (Public IP)
   ↓
[iptables NAT]
   ↓
Host Bridge (172.16.0.1/24)
   ↓
TAP Devices (tap0, tap1, ...)
   ↓
VM eth0 (172.16.0.10, 172.16.0.11, ...)
```

**Why NAT?**:
- ✅ Simple to implement
- ✅ Works without special host network permissions
- ✅ VMs get outbound internet access
- ✅ VMs can talk to each other
- ✅ Host acts as gateway/firewall
- ⚠️ Inbound connections require port forwarding (acceptable for now)

**Alternative (Bridge Mode)** - Future enhancement:
- Requires VMs on same L2 network as host
- More complex permission requirements
- Better for production clusters

---

## Implementation Steps

### Phase 1: TAP Device Management

**What**: Create and configure TAP network interfaces

**How**:
```go
// Create TAP device
ip tuntap add tap0 mode tap

// Bring it up
ip link set tap0 up

// Add to bridge (if using bridge)
ip link set tap0 master br0
```

**Implementation**:
- Use Go `exec.Command` to call `ip` commands
- Alternative: Use `netlink` library for pure Go implementation
- Store TAP device name in VM runtime state
- Cleanup TAP device when VM is deleted

### Phase 2: Bridge Setup

**What**: Create a bridge for all TAP devices

**How**:
```bash
# Create bridge (one-time setup)
ip link add nanofuse0 type bridge
ip addr add 172.16.0.1/24 dev nanofuse0
ip link set nanofuse0 up

# Each TAP device joins this bridge
ip link set tap-vm1 master nanofuse0
```

**Implementation**:
- Create bridge on daemon startup (if doesn't exist)
- All TAP devices attach to this bridge
- Bridge acts as gateway (172.16.0.1)

### Phase 3: IP Address Allocation

**What**: Assign unique IP addresses to each VM

**How**:
- DHCP server (complex) OR
- **Static assignment via metadata** (simple, chosen approach)

**IP Range**: 172.16.0.0/24
- 172.16.0.1 - Bridge/Gateway
- 172.16.0.10-254 - VMs (245 addresses available)

**Implementation**:
- Track allocated IPs in database
- Assign next available IP on VM create
- Free IP on VM delete
- Pass IP to VM via kernel cmdline or metadata

### Phase 4: NAT and IP Forwarding

**What**: Enable VMs to access internet via host

**How**:
```bash
# Enable IP forwarding (one-time)
sysctl -w net.ipv4.ip_forward=1

# Add iptables NAT rule (one-time)
iptables -t nat -A POSTROUTING -s 172.16.0.0/24 -o eth0 -j MASQUERADE

# Allow forwarding for the bridge
iptables -A FORWARD -i nanofuse0 -o eth0 -j ACCEPT
iptables -A FORWARD -i eth0 -o nanofuse0 -m state --state RELATED,ESTABLISHED -j ACCEPT
```

**Implementation**:
- Run these commands on daemon startup
- Detect primary network interface (eth0, wlan0, etc.)
- Store iptables rules to cleanup on shutdown

### Phase 5: Guest VM Configuration

**What**: Configure network inside the VM

**Problem**: VM needs to know:
- IP address (172.16.0.10)
- Gateway (172.16.0.1)
- DNS servers (8.8.8.8, 8.8.4.4)

**Solutions**:

**Option A: Kernel Command Line** (Simplest)
```
ip=172.16.0.10::172.16.0.1:255.255.255.0::eth0:off
```
- Limited functionality
- No DNS configuration

**Option B: Metadata Service** (Better)
- Implement minimal IMDS (Instance Metadata Service) on 169.254.169.254
- VM curls metadata to get network config
- More flexible

**Option C: Cloud-Init** (Best, but complex)
- Add cloud-init to base image
- Provide cloud-init config via ISO or metadata
- Handles full network configuration

**Chosen: Option A + Manual DNS for MVP**, then Option B for production

### Phase 6: DNS Configuration

**What**: VMs need DNS to resolve hostnames

**How**:
```bash
# In VM, configure DNS
echo "nameserver 8.8.8.8" > /etc/resolv.conf
echo "nameserver 8.8.4.4" >> /etc/resolv.conf
```

**Implementation**:
- Add to base image Dockerfile
- Or via metadata service
- Or via cloud-init

---

## Code Changes Required

### 1. New File: `internal/network/tap.go`

```go
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
        return fmt.Errorf("failed to create TAP device: %w", err)
    }

    // ip link set <name> up
    cmd = exec.Command("ip", "link", "set", name, "up")
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("failed to bring up TAP device: %w", err)
    }

    return nil
}

// AttachTAPToBridge attaches TAP device to bridge
func AttachTAPToBridge(tapName, bridgeName string) error {
    cmd := exec.Command("ip", "link", "set", tapName, "master", bridgeName)
    return cmd.Run()
}

// DeleteTAPDevice removes a TAP device
func DeleteTAPDevice(name string) error {
    cmd := exec.Command("ip", "link", "delete", name)
    return cmd.Run()
}
```

### 2. New File: `internal/network/bridge.go`

```go
package network

import (
    "fmt"
    "os/exec"
)

const (
    BridgeName = "nanofuse0"
    BridgeIP   = "172.16.0.1/24"
)

// SetupBridge creates and configures the nanofuse bridge
func SetupBridge() error {
    // Check if bridge already exists
    cmd := exec.Command("ip", "link", "show", BridgeName)
    if cmd.Run() == nil {
        // Bridge exists, skip creation
        return nil
    }

    // Create bridge
    cmd = exec.Command("ip", "link", "add", BridgeName, "type", "bridge")
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("failed to create bridge: %w", err)
    }

    // Assign IP
    cmd = exec.Command("ip", "addr", "add", BridgeIP, "dev", BridgeName)
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("failed to assign IP to bridge: %w", err)
    }

    // Bring up
    cmd = exec.Command("ip", "link", "set", BridgeName, "up")
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("failed to bring up bridge: %w", err)
    }

    return nil
}
```

### 3. New File: `internal/network/nat.go`

```go
package network

import (
    "fmt"
    "os/exec"
)

// SetupNAT configures IP forwarding and iptables NAT rules
func SetupNAT(primaryInterface string) error {
    // Enable IP forwarding
    cmd := exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1")
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("failed to enable IP forwarding: %w", err)
    }

    // Check if NAT rule exists
    cmd = exec.Command("iptables", "-t", "nat", "-C", "POSTROUTING",
        "-s", "172.16.0.0/24", "-o", primaryInterface, "-j", "MASQUERADE")

    if cmd.Run() != nil {
        // Rule doesn't exist, add it
        cmd = exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING",
            "-s", "172.16.0.0/24", "-o", primaryInterface, "-j", "MASQUERADE")
        if err := cmd.Run(); err != nil {
            return fmt.Errorf("failed to add NAT rule: %w", err)
        }
    }

    // Allow forwarding
    cmd = exec.Command("iptables", "-C", "FORWARD",
        "-i", BridgeName, "-o", primaryInterface, "-j", "ACCEPT")

    if cmd.Run() != nil {
        cmd = exec.Command("iptables", "-A", "FORWARD",
            "-i", BridgeName, "-o", primaryInterface, "-j", "ACCEPT")
        if err := cmd.Run(); err != nil {
            return fmt.Errorf("failed to add forward rule: %w", err)
        }
    }

    return nil
}

// GetPrimaryInterface returns the primary network interface
func GetPrimaryInterface() (string, error) {
    // Try common interface names
    interfaces := []string{"eth0", "ens3", "enp0s3", "wlan0"}

    for _, iface := range interfaces {
        cmd := exec.Command("ip", "link", "show", iface)
        if cmd.Run() == nil {
            return iface, nil
        }
    }

    // Fallback: parse `ip route` output
    cmd := exec.Command("ip", "route", "show", "default")
    output, err := cmd.Output()
    if err != nil {
        return "", fmt.Errorf("failed to detect primary interface")
    }

    // Parse "default via X.X.X.X dev <interface>"
    // ... parsing logic ...

    return "eth0", nil // Fallback
}
```

### 4. New File: `internal/network/ipam.go`

```go
package network

import (
    "fmt"
    "net"
    "sync"
)

// IPAM manages IP address allocation
type IPAM struct {
    mu        sync.Mutex
    allocated map[string]string // vmID -> IP
    pool      []string           // Available IPs
}

// NewIPAM creates a new IPAM instance
func NewIPAM() *IPAM {
    ipam := &IPAM{
        allocated: make(map[string]string),
        pool:      make([]string, 0),
    }

    // Generate IP pool: 172.16.0.10 - 172.16.0.254
    for i := 10; i <= 254; i++ {
        ipam.pool = append(ipam.pool, fmt.Sprintf("172.16.0.%d", i))
    }

    return ipam
}

// AllocateIP allocates an IP for a VM
func (ipam *IPAM) AllocateIP(vmID string) (string, error) {
    ipam.mu.Lock()
    defer ipam.mu.Unlock()

    if len(ipam.pool) == 0 {
        return "", fmt.Errorf("no IPs available")
    }

    ip := ipam.pool[0]
    ipam.pool = ipam.pool[1:]
    ipam.allocated[vmID] = ip

    return ip, nil
}

// ReleaseIP releases an IP when VM is deleted
func (ipam *IPAM) ReleaseIP(vmID string) {
    ipam.mu.Lock()
    defer ipam.mu.Unlock()

    if ip, ok := ipam.allocated[vmID]; ok {
        delete(ipam.allocated, vmID)
        ipam.pool = append(ipam.pool, ip)
    }
}
```

### 5. Update: `internal/firecracker/vm.go`

```go
// In SetupNetwork function
func (m *Manager) SetupNetwork(vm *types.VM) error {
    if vm.Config.Network.Mode == "none" {
        return nil
    }

    // Create TAP device
    tapName := fmt.Sprintf("tap-%s", vm.ID[:8])
    if err := network.CreateTAPDevice(tapName); err != nil {
        return fmt.Errorf("failed to create TAP device: %w", err)
    }

    // Attach to bridge
    if err := network.AttachTAPToBridge(tapName, network.BridgeName); err != nil {
        network.DeleteTAPDevice(tapName) // Cleanup
        return fmt.Errorf("failed to attach TAP to bridge: %w", err)
    }

    // Assign to VM config
    vm.Config.Network.TapDevice = tapName

    // Generate MAC address
    if vm.Config.Network.MACAddress == "" {
        vm.Config.Network.MACAddress = generateMAC()
    }

    return nil
}

// In CleanupNetwork function
func (m *Manager) CleanupNetwork(vm *types.VM) error {
    if vm.Config.Network.TapDevice != "" {
        return network.DeleteTAPDevice(vm.Config.Network.TapDevice)
    }
    return nil
}
```

### 6. Update: `cmd/nanofused/main.go`

```go
// On daemon startup, setup network
func main() {
    // ... existing code ...

    // Setup networking
    log.Info("Setting up network infrastructure...")

    if err := network.SetupBridge(); err != nil {
        log.Fatalf("Failed to setup bridge: %v", err)
    }

    primaryIface, err := network.GetPrimaryInterface()
    if err != nil {
        log.Warnf("Could not detect primary interface: %v", err)
        primaryIface = "eth0" // Fallback
    }

    if err := network.SetupNAT(primaryIface); err != nil {
        log.Fatalf("Failed to setup NAT: %v", err)
    }

    log.Info("Network infrastructure ready")

    // ... start API server ...
}
```

### 7. Update Base Image: `images/base/Dockerfile`

```dockerfile
# Add DNS configuration
RUN echo "nameserver 8.8.8.8" > /etc/resolv.conf && \
    echo "nameserver 8.8.4.4" >> /etc/resolv.conf

# Create network startup script
COPY units/configure-network.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/configure-network.sh

# Add systemd unit to configure network on boot
COPY units/configure-network.service /etc/systemd/system/
RUN systemctl enable configure-network.service
```

**New file**: `images/base/units/configure-network.sh`
```bash
#!/bin/bash
# Configure network from kernel cmdline
# Expects: ip=<IP>::<gateway>:<netmask>::eth0:off

IP_CONFIG=$(cat /proc/cmdline | grep -o 'ip=[^ ]*')
if [ -n "$IP_CONFIG" ]; then
    # Parse and configure using ip command
    # This is a simplified version
    echo "Configuring network from cmdline: $IP_CONFIG"
    # ... parsing and configuration ...
fi
```

---

## Testing Plan

### Test 1: Bridge and TAP Creation
```bash
# Start daemon
sudo ./bin/nanofused

# Check bridge exists
ip link show nanofuse0
# Should show: nanofuse0: <BROADCAST,MULTICAST,UP,LOWER_UP>

# Check bridge IP
ip addr show nanofuse0
# Should show: inet 172.16.0.1/24

# Create VM
./bin/nanofuse vm create test-vm --image nanofuse-base:latest

# Check TAP device created
ip link show tap-test-vm
# Should exist and be UP

# Check TAP attached to bridge
bridge link show
# Should show tap-test-vm attached to nanofuse0
```

### Test 2: NAT Rules
```bash
# Check IP forwarding enabled
sysctl net.ipv4.ip_forward
# Should show: net.ipv4.ip_forward = 1

# Check iptables NAT rule
sudo iptables -t nat -L POSTROUTING -v
# Should show MASQUERADE rule for 172.16.0.0/24

# Check forward rules
sudo iptables -L FORWARD -v
# Should show ACCEPT rules for nanofuse0
```

### Test 3: VM Network Connectivity
```bash
# Start VM
./bin/nanofuse vm start test-vm

# Wait for boot
sleep 10

# From host, ping VM
ping -c 3 172.16.0.10
# Should succeed

# SSH into VM (if SSH keys configured)
ssh root@172.16.0.10

# From inside VM:
# Test 3a: Ping gateway
ping -c 3 172.16.0.1
# Should succeed

# Test 3b: Ping external IP
ping -c 3 8.8.8.8
# Should succeed (proves routing works)

# Test 3c: DNS resolution
ping -c 3 google.com
# Should succeed (proves DNS works)

# Test 3d: HTTP access
curl -I https://google.com
# Should return 200 OK
```

### Test 4: Multiple VMs
```bash
# Create second VM
./bin/nanofuse vm create test-vm2 --image nanofuse-base:latest
./bin/nanofuse vm start test-vm2

# VM1 pings VM2
# From VM1: ping 172.16.0.11

# Both VMs can access internet
# From both: curl https://google.com
```

---

## Permission Requirements

**Daemon must run as root** OR user must have:
- CAP_NET_ADMIN capability
- Access to /dev/net/tun
- Sudo access for ip and iptables commands

**Production Setup**:
```bash
# Add user to required groups
sudo usermod -a -G kvm,netdev $USER

# Grant specific capabilities
sudo setcap cap_net_admin=ep /usr/local/bin/nanofused

# Or run as systemd service with appropriate capabilities
```

---

## Rollout Plan

1. ✅ Create detailed plan (this document)
2. ⏭️ Implement network package (tap.go, bridge.go, nat.go, ipam.go)
3. ⏭️ Update firecracker/vm.go with real network setup
4. ⏭️ Update daemon startup to initialize network
5. ⏭️ Update base image with DNS and network config
6. ⏭️ Test on clean system
7. ⏭️ Document in START-VM-GUIDE.md
8. ⏭️ Create NETWORK-TROUBLESHOOTING.md

---

## Future Enhancements

- **Bridge Mode**: VMs on host LAN
- **Port Forwarding**: Expose VM ports to host
- **IPv6 Support**: Dual-stack networking
- **DHCP Server**: Dynamic IP assignment
- **Metadata Service**: Cloud-init style config
- **Network Isolation**: Per-VM network namespaces
- **Bandwidth Limits**: QoS per VM

---

**Status**: Ready to implement
**Estimated Time**: 4-6 hours for full implementation and testing
**Dependencies**: Root access or NET_ADMIN capability
