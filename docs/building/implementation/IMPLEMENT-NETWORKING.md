# Implementing VM Networking - Step-by-Step Guide

**Status**: Network package created ✅, Integration needed ⏭️

---

## What We Have So Far

✅ **Network Package Created** (`internal/network/`):
- `tap.go` - TAP device creation/deletion
- `bridge.go` - Bridge setup and management
- `nat.go` - NAT/iptables configuration
- `ipam.go` - IP address allocation

✅ **Detailed Plan**: See `NETWORKING-PLAN.md`

⏭️ **Integration Needed**: Wire everything together

---

## Quick Test: Verify Network Package Works

Before integrating, let's test the network functions directly:

```bash
cd /home/jpoley/src/_mine/nanofuse

# Create a simple test program
cat > test-network.go << 'EOF'
package main

import (
	"fmt"
	"log"
	"github.com/daax-dev/nanofuse/internal/network"
)

func main() {
	fmt.Println("=== Testing NanoFuse Network Setup ===\n")

	// Test 1: Setup bridge
	fmt.Println("[1/5] Setting up bridge...")
	if err := network.SetupBridge(); err != nil {
		log.Fatalf("Failed to setup bridge: %v", err)
	}
	fmt.Println("✓ Bridge created: nanofuse0")

	// Test 2: Detect primary interface
	fmt.Println("\n[2/5] Detecting primary network interface...")
	primaryIface, err := network.GetPrimaryInterface()
	if err != nil {
		log.Fatalf("Failed to detect interface: %v", err)
	}
	fmt.Printf("✓ Primary interface: %s\n", primaryIface)

	// Test 3: Setup NAT
	fmt.Println("\n[3/5] Setting up NAT...")
	if err := network.SetupNAT(primaryIface); err != nil {
		log.Fatalf("Failed to setup NAT: %v", err)
	}
	fmt.Println("✓ NAT rules configured")

	// Test 4: Create TAP device
	fmt.Println("\n[4/5] Creating test TAP device...")
	if err := network.CreateTAPDevice("tap-test123"); err != nil {
		log.Fatalf("Failed to create TAP: %v", err)
	}
	fmt.Println("✓ TAP device created: tap-test123")

	// Test 5: Attach to bridge
	fmt.Println("\n[5/5] Attaching TAP to bridge...")
	if err := network.AttachTAPToBridge("tap-test123", network.BridgeName); err != nil {
		log.Fatalf("Failed to attach TAP: %v", err)
	}
	fmt.Println("✓ TAP attached to bridge")

	fmt.Println("\n=== All Tests Passed! ===")
	fmt.Println("\nVerify manually:")
	fmt.Println("  ip link show nanofuse0")
	fmt.Println("  ip link show tap-test123")
	fmt.Println("  bridge link show")
	fmt.Println("  iptables -t nat -L POSTROUTING -v")
	fmt.Println("\nCleanup:")
	fmt.Println("  sudo ip link delete tap-test123")
	fmt.Println("  sudo ip link delete nanofuse0")
}
EOF

# Build and run (requires sudo)
go build -o test-network test-network.go
sudo ./test-network
```

**Expected output:**
```
=== Testing NanoFuse Network Setup ===

[1/5] Setting up bridge...
✓ Bridge created: nanofuse0

[2/5] Detecting primary network interface...
✓ Primary interface: eth0

[3/5] Setting up NAT...
✓ NAT rules configured

[4/5] Creating test TAP device...
✓ TAP device created: tap-test123

[5/5] Attaching TAP to bridge...
✓ TAP attached to bridge

=== All Tests Passed! ===
```

**If this works**, the network package is good. Now we need to integrate it.

---

## Integration Step 1: Update Server to Include IPAM

**File**: `internal/api/server.go`

Add IPAM to the Server struct:

```go
// Find the Server struct and add:
type Server struct {
	// ... existing fields ...
	ipam    *network.IPAM  // Add this line
}

// In NewServer function, add:
func NewServer(config *config.Config, db *storage.DB, fc *firecracker.Manager) *Server {
	s := &Server{
		config: config,
		db:     db,
		fc:     fc,
		logger: log.New(os.Stdout, "[API] ", log.LstdFlags),
		ipam:   network.NewIPAM(),  // Add this line
	}
	// ... rest of function
}
```

---

## Integration Step 2: Update VM Creation to Setup Network

**File**: `internal/api/vm_handlers.go`

In `handleCreateVM`, after creating the VM but before saving:

```go
// After vm := &types.VM{ ... }
// Add network setup:

if vm.Config.Network.Mode != "none" {
	// Allocate IP address
	ip, err := s.ipam.AllocateIP(vm.ID)
	if err != nil {
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, "Failed to allocate IP", nil)
		return
	}

	// Create TAP device
	tapName := fmt.Sprintf("tap-%s", vm.ID[:8])
	if err := network.CreateTAPDevice(tapName); err != nil {
		s.ipam.ReleaseIP(vm.ID) // Cleanup
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, "Failed to create TAP device", nil)
		return
	}

	// Attach to bridge
	if err := network.AttachTAPToBridge(tapName, network.BridgeName); err != nil {
		network.DeleteTAPDevice(tapName) // Cleanup
		s.ipam.ReleaseIP(vm.ID)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, "Failed to attach TAP to bridge", nil)
		return
	}

	// Update VM config
	vm.Config.Network.TapDevice = tapName
	vm.Config.Network.MACAddress = generateMAC() // You'll need this helper
	vm.Config.Network.IPAddress = ip
	vm.Config.Network.Gateway = network.BridgeGateway
	vm.Config.Network.Netmask = "255.255.255.0"

	// Update kernel args to include IP configuration
	vm.Config.KernelArgs = fmt.Sprintf(
		"console=ttyS0 root=/dev/vda1 rw panic=1 reboot=k ip=%s::%s:255.255.255.0::eth0:off",
		ip, network.BridgeGateway,
	)

	s.logger.Printf("INFO: Configured network for VM %s: IP=%s TAP=%s", vm.ID, ip, tapName)
}
```

---

## Integration Step 3: Update VM Deletion to Cleanup Network

**File**: `internal/api/vm_handlers.go`

In `handleDeleteVM`, before deleting from DB:

```go
// Cleanup network resources
if vm.Config.Network.TapDevice != "" {
	if err := network.DeleteTAPDevice(vm.Config.Network.TapDevice); err != nil {
		s.logger.Printf("WARN: Failed to delete TAP device: %v", err)
	}
}
if vm.Config.Network.IPAddress != "" {
	s.ipam.ReleaseIP(vm.ID)
}
```

---

## Integration Step 4: Update Daemon Startup

**File**: `cmd/nanofused/main.go`

Add network initialization before starting the API server:

```go
import (
	// ... existing imports ...
	"github.com/daax-dev/nanofuse/internal/network"
)

func main() {
	// ... existing code ...

	// Setup network infrastructure (requires root)
	log.Println("Setting up network infrastructure...")

	if err := network.SetupBridge(); err != nil {
		log.Fatalf("Failed to setup bridge: %v (are you running as root?)", err)
	}
	log.Println("✓ Bridge configured: nanofuse0")

	primaryIface, err := network.GetPrimaryInterface()
	if err != nil {
		log.Printf("WARNING: Could not detect primary interface: %v", err)
		primaryIface = "eth0" // Fallback
	}
	log.Printf("✓ Detected primary interface: %s", primaryIface)

	if err := network.SetupNAT(primaryIface); err != nil {
		log.Fatalf("Failed to setup NAT: %v", err)
	}
	log.Println("✓ NAT configured")

	log.Println("Network infrastructure ready")

	// ... continue with API server setup ...
}
```

---

## Integration Step 5: Update Base Image with DNS

**File**: `images/base/Dockerfile`

Add permanent DNS configuration:

```dockerfile
# After the systemd setup, add:

# Configure DNS (8.8.8.8, 8.8.4.4)
RUN echo "nameserver 8.8.8.8" > /etc/resolv.conf && \
    echo "nameserver 8.8.4.4" >> /etc/resolv.conf && \
    echo "nameserver 1.1.1.1" >> /etc/resolv.conf

# Make resolv.conf immutable to prevent systemd-resolved from overwriting
RUN chattr +i /etc/resolv.conf || true
```

---

## Rebuild and Test

### 1. Rebuild Base Image

```bash
cd images/base
sudo ./build.sh
```

### 2. Rebuild Binaries

```bash
cd /home/jpoley/src/_mine/nanofuse
mage all
```

### 3. Copy Updated Image

```bash
sudo cp images/base/build/* /tmp/nanofuse/images/nanofuse-base/latest/
sudo chown -R jpoley:jpoley /tmp/nanofuse/images/
sudo chmod 664 /tmp/nanofuse/images/nanofuse-base/latest/rootfs.ext4
```

### 4. Start Daemon (requires sudo for networking)

```bash
sudo ./bin/nanofused --config ~/.config/nanofuse/config.yaml
```

**Expected output:**
```
Setting up network infrastructure...
✓ Bridge configured: nanofuse0
✓ Detected primary interface: eth0
✓ NAT configured
Network infrastructure ready
INFO[0000] Starting NanoFuse API daemon
INFO[0000] API server listening on 127.0.0.1:8080
```

### 5. Create and Start VM

```bash
# In another terminal
export NANOFUSE_API_URL="http://127.0.0.1:8080"

./bin/nanofuse vm create test-net --image nanofuse-base:latest --vcpus 2 --memory 512
./bin/nanofuse vm start test-net
```

### 6. Test Network Connectivity

```bash
# Get VM IP
VM_IP=$(./bin/nanofuse vm inspect test-net --json | jq -r '.config.network.ip_address')
echo "VM IP: $VM_IP"

# Wait for VM to boot
sleep 10

# Test 1: Ping VM from host
ping -c 3 $VM_IP

# Test 2: View VM console (should show network configured)
./bin/nanofuse vm logs test-net | grep -i "network\|eth0"

# Test 3: SSH into VM (if you added SSH keys)
ssh root@$VM_IP

# From inside VM:
# - ping 172.16.0.1  (gateway)
# - ping 8.8.8.8      (internet)
# - ping google.com   (DNS)
# - curl https://google.com
```

---

## If You Get Stuck

### Issue: "Permission denied" creating bridge

**Solution**: Run daemon with sudo
```bash
sudo ./bin/nanofused --config ~/.config/nanofuse/config.yaml
```

### Issue: "Failed to detect primary interface"

**Solution**: Check your network interface name
```bash
ip route show default
# Look for "dev <interface>"

# Manually specify in nat.go or pass as config
```

### Issue: VM boots but no network

**Check**:
1. Bridge exists: `ip link show nanofuse0`
2. TAP device exists: `ip link show tap-<vmid>`
3. TAP attached: `bridge link show`
4. IP forwarding: `sysctl net.ipv4.ip_forward`
5. iptables rules: `sudo iptables -t nat -L -v`

**Debug VM**:
```bash
# View console
./bin/nanofuse vm logs test-net

# Look for:
# - "eth0" interface coming up
# - IP address assignment
# - Errors related to network
```

---

## What to Do Next

1. ✅ Test network package standalone (use test-network.go above)
2. ⏭️ Make the integration changes listed above
3. ⏭️ Rebuild everything
4. ⏭️ Test end-to-end
5. ⏭️ Document results

**Want me to**:
- A) Make all the integration changes for you automatically?
- B) Create a migration script that makes the changes?
- C) Guide you through making each change manually?

Let me know and I'll continue!
