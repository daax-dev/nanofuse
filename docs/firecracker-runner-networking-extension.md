# FCRunner Networking Extension

## VM-to-VM Communication & Overlay Networks

**Status:** Design
**Extends:** [firecracker-runner-design.md](firecracker-runner-design.md)
**Last Updated:** 2025-12-21

---

## 1. Overview

The base design isolates VMs from each other (default-deny). This extension covers:

1. **Same-host VM-to-VM** - Controlled communication between VMs on one host
2. **Cross-host VM-to-VM** - Communication between VMs on different hosts
3. **WireGuard integration** - Encrypted overlay networking
4. **Tailscale integration** - Managed mesh networking with identity

---

## 2. Architecture Options

### 2.1 Options Overview

#### Option 1: Same-Host Policy Only (No Overlay)

VMs on same host can communicate via bridge + nftables policy. No cross-host capability.

| Aspect | Details |
|--------|---------|
| **Complexity** | Low - just nftables rules |
| **Cross-host** | Not supported |
| **Encryption** | None (local bridge) |
| **Identity** | IP-based |
| **Use cases** | Co-located services, single-host deployments |

#### Option 2: Host-Level WireGuard

Single WireGuard interface on host. All VM traffic NAT'd through host's WG identity.

| Aspect | Details |
|--------|---------|
| **Complexity** | Low-Medium |
| **Cross-host** | Yes |
| **Encryption** | Yes (WireGuard) |
| **Identity** | Shared (all VMs appear as host) |
| **Key management** | Simple - one keypair per host |
| **Pros** | Simple setup, easy to reason about, minimal overhead |
| **Cons** | No per-VM identity, host can inspect all traffic, can't distinguish VM traffic at remote end |
| **Use cases** | Trusted VMs, simple cross-host networking |

#### Option 3: Sidecar WireGuard (Per-VM Network Namespace)

Each VM gets dedicated WireGuard interface in host-side network namespace. Host controls keys.

| Aspect | Details |
|--------|---------|
| **Complexity** | High |
| **Cross-host** | Yes |
| **Encryption** | Yes (WireGuard) |
| **Identity** | Per-VM |
| **Key management** | Host generates/stores, VM never sees private key |
| **Pros** | Per-VM crypto identity, centralized key control, revocation easy |
| **Cons** | Complex namespace plumbing, more host resources, coordination needed |
| **Use cases** | Multi-tenant, managed infrastructure, compliance requirements |

#### Option 4: Guest WireGuard

WireGuard runs inside the VM. VM controls its own keys.

| Aspect | Details |
|--------|---------|
| **Complexity** | Medium |
| **Cross-host** | Yes |
| **Encryption** | Yes (end-to-end, host can't inspect) |
| **Identity** | Per-VM (VM-controlled) |
| **Key management** | VM/tenant manages own keys |
| **Pros** | True end-to-end encryption, host-blind, tenant autonomy |
| **Cons** | Host can't enforce network policy on encrypted traffic, harder to revoke, key distribution to guest |
| **Use cases** | Tenant-managed networking, privacy-first, BYO-network |

#### Option 5: Tailscale Subnet Router

Host joins Tailnet, advertises VM subnet. VMs don't run Tailscale.

| Aspect | Details |
|--------|---------|
| **Complexity** | Low |
| **Cross-host** | Yes (via Tailnet) |
| **Encryption** | Yes (WireGuard under the hood) |
| **Identity** | Host-level (VMs share) |
| **Key management** | Tailscale handles |
| **Pros** | NAT traversal, managed PKI, works through firewalls, ACLs via Tailscale admin |
| **Cons** | External dependency (Tailscale control plane), shared identity, cost at scale |
| **Use cases** | Access VMs from corporate network, remote development, hybrid cloud |

#### Option 6: Per-VM Tailscale

Each VM joins Tailnet with own identity.

| Aspect | Details |
|--------|---------|
| **Complexity** | Medium |
| **Cross-host** | Yes (full mesh via Tailnet) |
| **Encryption** | Yes |
| **Identity** | Per-VM (Tailscale device) |
| **Key management** | Tailscale handles, auth key per VM |
| **Pros** | Per-VM ACLs in Tailscale, auto-discovery, full mesh, MagicDNS |
| **Cons** | Tailscale dependency, auth key distribution, agent in guest, Tailscale device limits |
| **Use cases** | Full mesh networking, Tailscale-native environments, per-VM access control |

#### Option 7: Headscale (Self-Hosted Tailscale)

Same as Tailscale options but with self-hosted control plane.

| Aspect | Details |
|--------|---------|
| **Complexity** | Medium-High (need to run Headscale) |
| **Cross-host** | Yes |
| **Encryption** | Yes |
| **Identity** | Configurable (subnet-router or per-VM) |
| **Key management** | Self-managed via Headscale |
| **Pros** | No external dependency, full control, no device limits |
| **Cons** | Operational burden (run Headscale), less mature than Tailscale |
| **Use cases** | Air-gapped, compliance, cost-sensitive at scale |

### 2.2 Comparison Matrix

| Approach | Complexity | Cross-Host | Per-VM Identity | Key Control | Host Inspectable | External Deps |
|----------|------------|------------|-----------------|-------------|------------------|---------------|
| Same-host only | Low | No | N/A | N/A | Yes | None |
| Host WireGuard | Low-Med | Yes | No | Host | Yes | None |
| Sidecar WireGuard | High | Yes | Yes | Host | Yes | None |
| Guest WireGuard | Medium | Yes | Yes | Guest | No | None |
| Tailscale Subnet | Low | Yes | No | Tailscale | Yes | Tailscale |
| Per-VM Tailscale | Medium | Yes | Yes | Tailscale | Via TS | Tailscale |
| Headscale | Med-High | Yes | Configurable | Self | Configurable | None (self-host) |

### 2.3 Trust Model Implications

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        TUNNEL TERMINATION LOCATION                          │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  HOST-TERMINATED                         GUEST-TERMINATED                   │
│  ┌─────────────────────────────┐        ┌─────────────────────────────┐    │
│  │ + Host controls keys        │        │ + VM has own identity       │    │
│  │ + VM can't see plaintext    │        │ + End-to-end encryption     │    │
│  │   of other VMs' traffic     │        │ + Host can't inspect        │    │
│  │ + Centralized policy        │        │ - VM controls keys          │    │
│  │ - Host can inspect traffic  │        │ - Harder to revoke          │    │
│  │ - Shared identity (NAT)     │        │ - Per-VM auth management    │    │
│  └─────────────────────────────┘        └─────────────────────────────┘    │
│                                                                             │
│  SIDECAR (HOST, PER-VM NETNS)                                              │
│  ┌─────────────────────────────┐                                           │
│  │ + Per-VM identity           │                                           │
│  │ + Host controls keys        │                                           │
│  │ + VM can't access keys      │                                           │
│  │ + Centralized policy        │                                           │
│  │ ~ Host can inspect (owns ns)│                                           │
│  │ - More complex setup        │                                           │
│  └─────────────────────────────┘                                           │
│                                                                             │
│  Trade-off: Control vs Trust                                               │
│  - Host-terminated = operator controls everything, can inspect             │
│  - Guest-terminated = VM owner controls, operator blind                    │
│  - Sidecar = middle ground, per-VM identity but operator-controlled        │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Same-Host VM-to-VM Communication

### 3.1 Network Topology

```
┌─────────────────────────────────────────────────────────────────┐
│                           HOST                                  │
│                                                                 │
│  ┌───────────────────────────────────────────────────────────┐ │
│  │                      fcbr0 (bridge)                       │ │
│  │                       172.16.0.1                          │ │
│  └───────────────────────────────────────────────────────────┘ │
│          │              │              │              │        │
│          │              │              │              │        │
│      ┌───┴───┐      ┌───┴───┐      ┌───┴───┐      ┌───┴───┐   │
│      │ tap-a │      │ tap-b │      │ tap-c │      │ tap-d │   │
│      └───┬───┘      └───┬───┘      └───┬───┘      └───┬───┘   │
│          │              │              │              │        │
│      ┌───┴───┐      ┌───┴───┐      ┌───┴───┐      ┌───┴───┐   │
│      │ VM-A  │      │ VM-B  │      │ VM-C  │      │ VM-D  │   │
│      │ .2    │◄────►│ .3    │      │ .4    │◄────►│ .5    │   │
│      └───────┘      └───────┘      └───────┘      └───────┘   │
│         group: frontend            group: backend             │
│                                                                 │
│  Policy: frontend <-> frontend: ALLOW                          │
│          backend <-> backend: ALLOW                            │
│          frontend -> backend: ALLOW (specific ports)           │
│          backend -> frontend: DENY                             │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 3.2 Policy Schema Extension

```yaml
apiVersion: fcrunner.io/v1
kind: NetworkPolicy
metadata:
  name: allow-frontend-group
spec:
  vmSelector:
    matchLabels:
      group: frontend
  
  # Peer policies (VM-to-VM)
  peers:
    - name: allow-same-group
      direction: both
      peerSelector:
        matchLabels:
          group: frontend
      ports:
        - port: 0  # All ports
          protocol: any
    
    - name: allow-to-backend
      direction: egress
      peerSelector:
        matchLabels:
          group: backend
      ports:
        - port: 5432
          protocol: tcp
        - port: 6379
          protocol: tcp
  
  # External policies (unchanged from base)
  ingress:
    - name: allow-http
      from:
        - cidr: 0.0.0.0/0
      ports:
        - port: 80
          protocol: tcp
  
  egress:
    - name: allow-dns
      to:
        - cidr: 172.16.0.1/32
      ports:
        - port: 53
          protocol: udp
```

### 3.3 nftables Compilation for Peer Policies

```bash
#!/bin/bash
# compile-peer-policy.sh

# Generate rules for VM-to-VM communication
# Called with: compile-peer-policy.sh <vm_id> <vm_ip> <peer_ips...>

VM_ID="$1"
VM_IP="$2"
shift 2
PEER_IPS=("$@")

cat << EOF
# Peer policy for VM ${VM_ID}
table inet fcrunner_peers_${VM_ID} {
    set allowed_peers_${VM_ID} {
        type ipv4_addr
        elements = { $(IFS=,; echo "${PEER_IPS[*]}") }
    }
    
    chain forward {
        type filter hook forward priority -5; policy accept;
        
        # VM -> Peer (check source is this VM, dest is allowed peer)
        ip saddr ${VM_IP} ip daddr @allowed_peers_${VM_ID} accept
        
        # Peer -> VM (check source is allowed peer, dest is this VM)
        ip saddr @allowed_peers_${VM_ID} ip daddr ${VM_IP} accept
        
        # VM -> Other VM (not in peer set) - handled by per-VM policy (drop)
    }
}
EOF
```

### 3.4 Dynamic Peer Updates

```go
// network/peers.go
package network

import (
    "context"
    "fmt"
    "sync"
)

type PeerManager struct {
    mu       sync.RWMutex
    peers    map[string][]string  // vmID -> []peerIPs
    nft      *NFTablesClient
    policies *PolicyStore
}

// UpdatePeers recalculates peer relationships when VMs or policies change
func (pm *PeerManager) UpdatePeers(ctx context.Context, vmID string) error {
    pm.mu.Lock()
    defer pm.mu.Unlock()
    
    vm, err := pm.getVM(vmID)
    if err != nil {
        return err
    }
    
    // Find all policies that match this VM
    policies := pm.policies.FindMatchingPolicies(vm.Labels)
    
    // Collect all allowed peers
    peerIPs := make(map[string]struct{})
    for _, policy := range policies {
        for _, peer := range policy.Spec.Peers {
            // Find VMs matching peer selector
            matchingVMs := pm.findVMsBySelector(peer.PeerSelector)
            for _, peerVM := range matchingVMs {
                if peerVM.ID != vmID {
                    peerIPs[peerVM.Status.IPAddress] = struct{}{}
                }
            }
        }
    }
    
    // Convert to slice
    ips := make([]string, 0, len(peerIPs))
    for ip := range peerIPs {
        ips = append(ips, ip)
    }
    
    // Update nftables
    if err := pm.nft.UpdatePeerSet(vmID, ips); err != nil {
        return fmt.Errorf("failed to update peer set: %w", err)
    }
    
    pm.peers[vmID] = ips
    return nil
}

// OnVMCreated triggers peer recalculation for affected VMs
func (pm *PeerManager) OnVMCreated(ctx context.Context, vmID string) error {
    // Update the new VM's peers
    if err := pm.UpdatePeers(ctx, vmID); err != nil {
        return err
    }
    
    // Update all VMs that might now peer with this one
    vm, _ := pm.getVM(vmID)
    affectedVMs := pm.findVMsWithMatchingPeerSelectors(vm.Labels)
    
    for _, affectedVM := range affectedVMs {
        if err := pm.UpdatePeers(ctx, affectedVM.ID); err != nil {
            // Log but don't fail
            log.Printf("failed to update peers for %s: %v", affectedVM.ID, err)
        }
    }
    
    return nil
}
```

---

## 4. Cross-Host Communication

### 4.1 Options Comparison

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         CROSS-HOST OPTIONS                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  OPTION A: VXLAN (L2 over L3)                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  Host A                              Host B                         │   │
│  │  ┌─────────┐                        ┌─────────┐                    │   │
│  │  │ fcbr0   │                        │ fcbr0   │                    │   │
│  │  └────┬────┘                        └────┬────┘                    │   │
│  │       │                                  │                          │   │
│  │  ┌────┴────┐                        ┌────┴────┐                    │   │
│  │  │ vxlan0  │◄──────── UDP 4789 ────►│ vxlan0  │                    │   │
│  │  └─────────┘                        └─────────┘                    │   │
│  │                                                                     │   │
│  │  Pros: Native Linux, simple, same-subnet illusion                  │   │
│  │  Cons: No encryption (unless IPsec), multicast or unicast flooding │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  OPTION B: WireGuard (L3 encrypted tunnel)                                 │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  Host A                              Host B                         │   │
│  │  ┌─────────┐                        ┌─────────┐                    │   │
│  │  │ fcbr0   │                        │ fcbr0   │                    │   │
│  │  └────┬────┘                        └────┬────┘                    │   │
│  │       │ routing                          │ routing                  │   │
│  │  ┌────┴────┐                        ┌────┴────┐                    │   │
│  │  │   wg0   │◄──────── UDP 51820 ───►│   wg0   │                    │   │
│  │  └─────────┘                        └─────────┘                    │   │
│  │                                                                     │   │
│  │  Pros: Encrypted, fast, simple key management                      │   │
│  │  Cons: L3 only (no broadcast), need routing between subnets        │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  OPTION C: Tailscale (managed WireGuard mesh)                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  Host A                              Host B                         │   │
│  │  ┌─────────┐                        ┌─────────┐                    │   │
│  │  │ fcbr0   │                        │ fcbr0   │                    │   │
│  │  └────┬────┘                        └────┬────┘                    │   │
│  │       │                                  │                          │   │
│  │  ┌────┴────┐    Tailscale Control   ┌────┴────┐                    │   │
│  │  │tailscale│◄───── Coordination ───►│tailscale│                    │   │
│  │  └─────────┘                        └─────────┘                    │   │
│  │       │                                  │                          │   │
│  │       └──────────── WireGuard ──────────┘                          │   │
│  │                                                                     │   │
│  │  Pros: NAT traversal, identity, ACLs, managed                      │   │
│  │  Cons: External dependency (or self-host Headscale)                │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 5. WireGuard Integration

### 5.1 Architecture: Host-Level WireGuard

VMs share host's WireGuard identity. Traffic from all VMs appears to come from host.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  HOST A (10.0.0.1)                      HOST B (10.0.0.2)                   │
│                                                                             │
│  ┌─────────────────────────┐           ┌─────────────────────────┐         │
│  │       fcbr0             │           │       fcbr0             │         │
│  │     172.16.0.1/24       │           │     172.17.0.1/24       │         │
│  └───────────┬─────────────┘           └───────────┬─────────────┘         │
│              │                                      │                       │
│      ┌───────┴───────┐                      ┌───────┴───────┐              │
│      │               │                      │               │              │
│   ┌──┴──┐         ┌──┴──┐                ┌──┴──┐         ┌──┴──┐          │
│   │VM-A │         │VM-B │                │VM-C │         │VM-D │          │
│   │.2   │         │.3   │                │.2   │         │.3   │          │
│   └─────┘         └─────┘                └─────┘         └─────┘          │
│                                                                             │
│  ┌─────────────────────────┐           ┌─────────────────────────┐         │
│  │         wg0             │           │         wg0             │         │
│  │    192.168.100.1/24     │◄─────────►│    192.168.100.2/24     │         │
│  └─────────────────────────┘           └─────────────────────────┘         │
│                                                                             │
│  Routing:                              Routing:                            │
│  172.17.0.0/24 via 192.168.100.2       172.16.0.0/24 via 192.168.100.1    │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

Traffic: VM-A (172.16.0.2) -> VM-C (172.17.0.2)
Path:    VM-A -> fcbr0 -> routing -> wg0 -> [encrypted] -> wg0 -> fcbr0 -> VM-C
```

### 5.2 Configuration: Host-Level WireGuard

```yaml
# /opt/fcrunner/etc/wireguard/wg0.conf
apiVersion: fcrunner.io/v1
kind: WireGuardConfig
metadata:
  name: cluster-overlay
spec:
  interface:
    name: wg0
    address: 192.168.100.1/24
    listenPort: 51820
    privateKeyFile: /opt/fcrunner/etc/wireguard/private.key
  
  peers:
    - name: host-b
      publicKey: "abc123..."
      endpoint: 10.0.0.2:51820
      allowedIPs:
        - 192.168.100.2/32   # Host B's WG address
        - 172.17.0.0/24       # Host B's VM subnet
      persistentKeepalive: 25
    
    - name: host-c
      publicKey: "def456..."
      endpoint: 10.0.0.3:51820
      allowedIPs:
        - 192.168.100.3/32
        - 172.18.0.0/24
      persistentKeepalive: 25
  
  # Which VM traffic to route through WireGuard
  routing:
    # Route all cross-host VM traffic
    mode: auto  # auto | manual | none
    
    # Or manual specification
    routes:
      - destination: 172.17.0.0/24
        via: 192.168.100.2
      - destination: 172.18.0.0/24
        via: 192.168.100.3
```

### 5.3 Implementation: Host WireGuard Manager

```go
// wireguard/manager.go
package wireguard

import (
    "fmt"
    "os/exec"
    "text/template"
)

type Manager struct {
    configPath string
    keyPath    string
}

type WGConfig struct {
    Interface InterfaceConfig
    Peers     []PeerConfig
}

type InterfaceConfig struct {
    PrivateKey string
    Address    string
    ListenPort int
}

type PeerConfig struct {
    PublicKey           string
    Endpoint            string
    AllowedIPs          []string
    PersistentKeepalive int
}

const wgTemplate = `[Interface]
PrivateKey = {{ .Interface.PrivateKey }}
Address = {{ .Interface.Address }}
ListenPort = {{ .Interface.ListenPort }}

{{ range .Peers }}
[Peer]
PublicKey = {{ .PublicKey }}
{{ if .Endpoint }}Endpoint = {{ .Endpoint }}{{ end }}
AllowedIPs = {{ join .AllowedIPs ", " }}
{{ if .PersistentKeepalive }}PersistentKeepalive = {{ .PersistentKeepalive }}{{ end }}

{{ end }}`

func (m *Manager) ApplyConfig(config *WGConfig) error {
    // Write config file
    if err := m.writeConfig(config); err != nil {
        return fmt.Errorf("failed to write config: %w", err)
    }
    
    // Apply with wg-quick
    cmd := exec.Command("wg-quick", "up", "wg0")
    if output, err := cmd.CombinedOutput(); err != nil {
        return fmt.Errorf("wg-quick failed: %s: %w", output, err)
    }
    
    return nil
}

func (m *Manager) AddPeer(peer PeerConfig) error {
    // Hot-add peer without restarting interface
    args := []string{
        "set", "wg0",
        "peer", peer.PublicKey,
        "allowed-ips", strings.Join(peer.AllowedIPs, ","),
    }
    
    if peer.Endpoint != "" {
        args = append(args, "endpoint", peer.Endpoint)
    }
    
    if peer.PersistentKeepalive > 0 {
        args = append(args, "persistent-keepalive", 
            fmt.Sprintf("%d", peer.PersistentKeepalive))
    }
    
    cmd := exec.Command("wg", args...)
    if output, err := cmd.CombinedOutput(); err != nil {
        return fmt.Errorf("wg set failed: %s: %w", output, err)
    }
    
    // Add route for peer's VM subnet
    for _, cidr := range peer.AllowedIPs {
        if !strings.Contains(cidr, "/32") {
            // This is a subnet, add route
            cmd := exec.Command("ip", "route", "add", cidr, "dev", "wg0")
            cmd.Run() // Ignore error if route exists
        }
    }
    
    return nil
}

func (m *Manager) RemovePeer(publicKey string) error {
    cmd := exec.Command("wg", "set", "wg0", "peer", publicKey, "remove")
    if output, err := cmd.CombinedOutput(); err != nil {
        return fmt.Errorf("wg remove peer failed: %s: %w", output, err)
    }
    return nil
}
```

### 5.4 Architecture: Sidecar WireGuard (Per-VM Identity)

Each VM gets its own WireGuard identity, but tunnel terminates in host-side network namespace.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                                  HOST                                       │
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │                         Root Network Namespace                        │ │
│  │                                                                       │ │
│  │  ┌─────────────┐                                                     │ │
│  │  │    eth0     │  (Physical/external)                                │ │
│  │  └──────┬──────┘                                                     │ │
│  │         │                                                             │ │
│  │         │  NAT / Routing                                             │ │
│  │         │                                                             │ │
│  └─────────┼─────────────────────────────────────────────────────────────┘ │
│            │                                                                │
│  ┌─────────┼─────────────────────────────────────────────────────────────┐ │
│  │         │           Per-VM Network Namespaces                         │ │
│  │         │                                                             │ │
│  │  ┌──────┴────────────────────┐    ┌─────────────────────────────┐    │ │
│  │  │  netns: fcns-vm-a         │    │  netns: fcns-vm-b           │    │ │
│  │  │                           │    │                             │    │ │
│  │  │  ┌─────────┐ ┌─────────┐ │    │  ┌─────────┐ ┌─────────┐   │    │ │
│  │  │  │  wg-a   │ │  tap-a  │ │    │  │  wg-b   │ │  tap-b  │   │    │ │
│  │  │  │(sidecar)│ │ (to VM) │ │    │  │(sidecar)│ │ (to VM) │   │    │ │
│  │  │  │10.100.  │ │172.16.  │ │    │  │10.100.  │ │172.16.  │   │    │ │
│  │  │  │  0.1    │ │  0.1    │ │    │  │  0.2    │ │  0.1    │   │    │ │
│  │  │  └────┬────┘ └────┬────┘ │    │  └────┬────┘ └────┬────┘   │    │ │
│  │  │       │           │      │    │       │           │        │    │ │
│  │  │       │    ┌──────┴────┐ │    │       │    ┌──────┴────┐   │    │ │
│  │  │       │    │           │ │    │       │    │           │   │    │ │
│  │  └───────┼────┼───────────┼─┘    └───────┼────┼───────────┼───┘    │ │
│  │          │    │  ┌────────┴─┐            │    │  ┌────────┴─┐      │ │
│  │          │    │  │   VM-A   │            │    │  │   VM-B   │      │ │
│  │          │    │  │ 172.16.  │            │    │  │ 172.16.  │      │ │
│  │          │    │  │   0.2    │            │    │  │   0.2    │      │ │
│  │          │    │  └──────────┘            │    │  └──────────┘      │ │
│  │          │    │                          │    │                    │ │
│  │          │    │ Firecracker              │    │ Firecracker        │ │
│  │          │                               │                         │ │
│  │          └───────────────────────────────┘                         │ │
│  │                    WireGuard mesh                                  │ │
│  │                  (VM-A <-> VM-B, etc)                              │ │
│  │                                                                     │ │
│  └─────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

Key insight: WireGuard runs in per-VM network namespace on HOST.
             VM sees only tap interface.
             Host controls all keys.
             Each VM has unique WG identity.
```

### 5.5 Implementation: Sidecar WireGuard

```go
// wireguard/sidecar.go
package wireguard

import (
    "context"
    "crypto/rand"
    "encoding/base64"
    "fmt"
    "os/exec"
    
    "golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type SidecarManager struct {
    coordinator *Coordinator  // Distributes peer info across hosts
    keyStore    *KeyStore     // Stores VM WG keys
}

// SetupVMWireGuard creates a WireGuard interface in the VM's network namespace
func (sm *SidecarManager) SetupVMWireGuard(ctx context.Context, vmID string, netns string) (*WGInfo, error) {
    // Generate keys for this VM
    privateKey, err := wgtypes.GeneratePrivateKey()
    if err != nil {
        return nil, fmt.Errorf("failed to generate key: %w", err)
    }
    publicKey := privateKey.PublicKey()
    
    // Allocate WG IP for this VM
    wgIP, err := sm.allocateWGIP(vmID)
    if err != nil {
        return nil, fmt.Errorf("failed to allocate WG IP: %w", err)
    }
    
    // Create WG interface in the VM's network namespace
    wgIface := fmt.Sprintf("wg-%s", vmID[:8])
    
    commands := [][]string{
        // Create WG interface
        {"ip", "link", "add", wgIface, "type", "wireguard"},
        // Move to namespace
        {"ip", "link", "set", wgIface, "netns", netns},
    }
    
    for _, args := range commands {
        if err := exec.CommandContext(ctx, args[0], args[1:]...).Run(); err != nil {
            return nil, fmt.Errorf("failed to run %v: %w", args, err)
        }
    }
    
    // Configure WG in namespace
    nsCommands := [][]string{
        {"ip", "addr", "add", wgIP + "/24", "dev", wgIface},
        {"ip", "link", "set", wgIface, "up"},
    }
    
    for _, args := range nsCommands {
        cmd := exec.CommandContext(ctx, "ip", "netns", "exec", netns)
        cmd.Args = append(cmd.Args, args...)
        if err := cmd.Run(); err != nil {
            return nil, fmt.Errorf("failed to run in netns %v: %w", args, err)
        }
    }
    
    // Configure WireGuard keys (using wg command in namespace)
    wgCmd := exec.CommandContext(ctx, "ip", "netns", "exec", netns,
        "wg", "set", wgIface,
        "private-key", "/dev/stdin",
        "listen-port", "51820",
    )
    wgCmd.Stdin = strings.NewReader(privateKey.String())
    if err := wgCmd.Run(); err != nil {
        return nil, fmt.Errorf("failed to configure wireguard: %w", err)
    }
    
    // Store keys
    info := &WGInfo{
        VMID:       vmID,
        Interface:  wgIface,
        PrivateKey: privateKey.String(),
        PublicKey:  publicKey.String(),
        IP:         wgIP,
        Netns:      netns,
    }
    
    if err := sm.keyStore.Store(vmID, info); err != nil {
        return nil, fmt.Errorf("failed to store keys: %w", err)
    }
    
    // Announce to coordinator for peer discovery
    if err := sm.coordinator.Announce(ctx, info); err != nil {
        return nil, fmt.Errorf("failed to announce: %w", err)
    }
    
    return info, nil
}

// AddPeers configures allowed peers for a VM's WireGuard interface
func (sm *SidecarManager) AddPeers(ctx context.Context, vmID string, peers []string) error {
    info, err := sm.keyStore.Get(vmID)
    if err != nil {
        return err
    }
    
    for _, peerVMID := range peers {
        peerInfo, err := sm.coordinator.GetPeerInfo(ctx, peerVMID)
        if err != nil {
            return fmt.Errorf("failed to get peer %s info: %w", peerVMID, err)
        }
        
        // Add peer to WG config
        args := []string{
            "netns", "exec", info.Netns,
            "wg", "set", info.Interface,
            "peer", peerInfo.PublicKey,
            "allowed-ips", peerInfo.IP + "/32",
        }
        
        if peerInfo.Endpoint != "" {
            args = append(args, "endpoint", peerInfo.Endpoint)
        }
        
        cmd := exec.CommandContext(ctx, "ip", args...)
        if err := cmd.Run(); err != nil {
            return fmt.Errorf("failed to add peer: %w", err)
        }
        
        // Add route in namespace
        routeCmd := exec.CommandContext(ctx, "ip", "netns", "exec", info.Netns,
            "ip", "route", "add", peerInfo.IP + "/32", "dev", info.Interface,
        )
        routeCmd.Run() // Ignore if exists
    }
    
    return nil
}

type WGInfo struct {
    VMID       string
    Interface  string
    PrivateKey string
    PublicKey  string
    IP         string
    Netns      string
    Endpoint   string  // host:port for cross-host
}
```

---

## 6. Tailscale Integration

### 6.1 Option A: Subnet Router (Simple)

Host runs Tailscale and advertises VM subnets to the Tailnet.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              TAILNET                                        │
│                                                                             │
│  ┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐       │
│  │   Your Laptop   │     │   FCRunner A    │     │   FCRunner B    │       │
│  │  100.64.0.1     │     │  100.64.0.2     │     │  100.64.0.3     │       │
│  │                 │     │                 │     │                 │       │
│  │                 │     │ Advertises:     │     │ Advertises:     │       │
│  │                 │     │ 172.16.0.0/24   │     │ 172.17.0.0/24   │       │
│  └────────┬────────┘     └────────┬────────┘     └────────┬────────┘       │
│           │                       │                       │                 │
│           └───────────────────────┼───────────────────────┘                 │
│                                   │                                         │
│                          Tailscale Control                                  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

From laptop: curl http://172.16.0.2:8080  # Reaches VM on FCRunner A
             curl http://172.17.0.3:3000  # Reaches VM on FCRunner B
```

### 6.2 Subnet Router Configuration

```yaml
# /opt/fcrunner/etc/tailscale/config.yaml
apiVersion: fcrunner.io/v1
kind: TailscaleConfig
metadata:
  name: subnet-router
spec:
  mode: subnet-router
  
  # Tailscale auth
  authKey: ${TAILSCALE_AUTH_KEY}  # Or use OAuth
  
  # Subnets to advertise
  advertiseRoutes:
    - 172.16.0.0/24  # VM subnet
  
  # Accept routes from other subnet routers
  acceptRoutes: true
  
  # Tailscale ACL tag for this node
  tags:
    - tag:fcrunner
    - tag:subnet-router
  
  # Exit node (optional - allow VMs to use Tailscale exit node)
  exitNode: false
```

### 6.3 Implementation: Tailscale Subnet Router

```bash
#!/bin/bash
# setup-tailscale-subnet-router.sh

# Install Tailscale
curl -fsSL https://tailscale.com/install.sh | sh

# Authenticate with auth key
tailscale up \
  --authkey="${TAILSCALE_AUTH_KEY}" \
  --advertise-routes=172.16.0.0/24 \
  --accept-routes \
  --hostname="fcrunner-$(hostname)"

# Enable IP forwarding (required for subnet routing)
echo 1 > /proc/sys/net/ipv4/ip_forward

# Tailscale handles the rest
```

### 6.4 Option B: Per-VM Tailscale (Full Identity)

Each VM joins the Tailnet with its own identity.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              TAILNET                                        │
│                                                                             │
│     ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                   │
│     │  VM-A        │  │  VM-B        │  │  VM-C        │                   │
│     │ 100.64.0.10  │  │ 100.64.0.11  │  │ 100.64.0.12  │                   │
│     │              │  │              │  │              │                   │
│     │ Identity:    │  │ Identity:    │  │ Identity:    │                   │
│     │ vm-a@tailnet │  │ vm-b@tailnet │  │ vm-c@tailnet │                   │
│     └──────┬───────┘  └──────┬───────┘  └──────┬───────┘                   │
│            │                 │                 │                            │
│            └─────────────────┼─────────────────┘                            │
│                              │                                              │
│                     Tailscale Control                                       │
│                              │                                              │
│                     ACL Policy enforced                                     │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

Tailscale ACL:
{
  "acls": [
    {"action": "accept", "src": ["tag:fcrunner-vm"], "dst": ["tag:fcrunner-vm:*"]},
    {"action": "accept", "src": ["tag:admin"], "dst": ["tag:fcrunner-vm:22"]}
  ]
}
```

### 6.5 Guest Tailscale Agent

```yaml
# In guest rootfs, systemd unit for Tailscale
# /etc/systemd/system/tailscale-vm.service

[Unit]
Description=Tailscale VM Agent
After=network.target
Wants=network.target

[Service]
Type=notify
ExecStart=/usr/sbin/tailscaled --state=/var/lib/tailscale/tailscaled.state
ExecStartPost=/usr/bin/tailscale up --authkey=${TAILSCALE_AUTH_KEY} --hostname=${VM_ID}

[Install]
WantedBy=multi-user.target
```

### 6.6 Auth Key Distribution

```go
// tailscale/authkeys.go
package tailscale

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type AuthKeyManager struct {
    apiKey    string
    tailnet   string
    baseURL   string
}

type AuthKeyRequest struct {
    Capabilities AuthKeyCapabilities `json:"capabilities"`
    ExpirySeconds int                `json:"expirySeconds"`
    Description   string             `json:"description"`
}

type AuthKeyCapabilities struct {
    Devices AuthKeyDevices `json:"devices"`
}

type AuthKeyDevices struct {
    Create AuthKeyCreate `json:"create"`
}

type AuthKeyCreate struct {
    Reusable      bool     `json:"reusable"`
    Ephemeral     bool     `json:"ephemeral"`
    Preauthorized bool     `json:"preauthorized"`
    Tags          []string `json:"tags"`
}

// CreateVMAuthKey creates a single-use, ephemeral auth key for a VM
func (m *AuthKeyManager) CreateVMAuthKey(ctx context.Context, vmID string) (string, error) {
    req := AuthKeyRequest{
        Capabilities: AuthKeyCapabilities{
            Devices: AuthKeyDevices{
                Create: AuthKeyCreate{
                    Reusable:      false,  // Single use
                    Ephemeral:     true,   // Auto-cleanup when offline
                    Preauthorized: true,   // No manual approval needed
                    Tags:          []string{"tag:fcrunner-vm"},
                },
            },
        },
        ExpirySeconds: 300,  // 5 minute expiry (just for initial auth)
        Description:   fmt.Sprintf("FCRunner VM: %s", vmID),
    }
    
    body, _ := json.Marshal(req)
    
    httpReq, err := http.NewRequestWithContext(ctx, "POST",
        fmt.Sprintf("%s/api/v2/tailnet/%s/keys", m.baseURL, m.tailnet),
        bytes.NewReader(body),
    )
    if err != nil {
        return "", err
    }
    
    httpReq.Header.Set("Authorization", "Bearer " + m.apiKey)
    httpReq.Header.Set("Content-Type", "application/json")
    
    resp, err := http.DefaultClient.Do(httpReq)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("API returned %d", resp.StatusCode)
    }
    
    var result struct {
        Key string `json:"key"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return "", err
    }
    
    return result.Key, nil
}
```

### 6.7 Option C: Headscale (Self-Hosted)

For those who want Tailscale-like functionality without external dependency.

```yaml
# /opt/fcrunner/etc/headscale/config.yaml
server_url: https://headscale.internal.example.com
listen_addr: 0.0.0.0:8080
metrics_listen_addr: 127.0.0.1:9090

private_key_path: /opt/fcrunner/etc/headscale/private.key

# Use SQLite for simplicity
db_type: sqlite3
db_path: /opt/fcrunner/data/headscale.db

# OIDC integration (optional)
oidc:
  issuer: https://auth.example.com
  client_id: headscale
  client_secret: ${OIDC_CLIENT_SECRET}

# IP allocation
ip_prefixes:
  - 100.64.0.0/10

# DNS
dns_config:
  nameservers:
    - 1.1.1.1
  magic_dns: true
  base_domain: fcrunner.internal

# ACL policy (Tailscale-compatible)
acl_policy_path: /opt/fcrunner/etc/headscale/acl.json
```

---

## 7. Network Policy Integration

### 7.1 Extended Policy Schema

```yaml
apiVersion: fcrunner.io/v1
kind: NetworkPolicy
metadata:
  name: multi-network-policy
spec:
  vmSelector:
    matchLabels:
      app: webserver
  
  # Local (same-host) peers
  peers:
    - name: allow-database
      peerSelector:
        matchLabels:
          app: database
      ports:
        - port: 5432
          protocol: tcp
  
  # WireGuard overlay peers
  wireguard:
    enabled: true
    peers:
      - name: remote-cache
        # Reference to VM on another host
        vmRef:
          name: redis-cache
          host: fcrunner-host-b
        ports:
          - port: 6379
            protocol: tcp
  
  # Tailscale peers (by Tailscale identity)
  tailscale:
    enabled: true
    peers:
      - name: allow-from-admin
        identity: admin@example.com
        ports:
          - port: 22
            protocol: tcp
      
      - name: allow-from-monitoring
        tags:
          - tag:monitoring
        ports:
          - port: 9090
            protocol: tcp
  
  # Standard ingress/egress unchanged
  ingress: [...]
  egress: [...]
```

### 7.2 Policy Compiler Updates

```go
// network/policy_compiler.go

type CompiledPolicy struct {
    NFTables     string           // nftables rules
    WireGuard    *WGPeerConfig    // WG peer configuration
    Tailscale    *TSACLFragment   // Tailscale ACL fragment
}

func (c *PolicyCompiler) Compile(policy *NetworkPolicy, vm *VM) (*CompiledPolicy, error) {
    result := &CompiledPolicy{}
    
    // Compile nftables (local + basic peer rules)
    nft, err := c.compileNFTables(policy, vm)
    if err != nil {
        return nil, err
    }
    result.NFTables = nft
    
    // Compile WireGuard peers
    if policy.Spec.WireGuard != nil && policy.Spec.WireGuard.Enabled {
        wg, err := c.compileWireGuard(policy.Spec.WireGuard, vm)
        if err != nil {
            return nil, err
        }
        result.WireGuard = wg
    }
    
    // Compile Tailscale ACL fragment
    if policy.Spec.Tailscale != nil && policy.Spec.Tailscale.Enabled {
        ts, err := c.compileTailscale(policy.Spec.Tailscale, vm)
        if err != nil {
            return nil, err
        }
        result.Tailscale = ts
    }
    
    return result, nil
}
```

---

## 8. Configuration Reference

### 8.1 Overlay Network Configuration

```yaml
# /opt/fcrunner/etc/overlay.yaml
apiVersion: fcrunner.io/v1
kind: OverlayConfig
metadata:
  name: cluster-overlay
spec:
  # Choose one overlay type
  type: wireguard  # wireguard | tailscale | headscale | none
  
  wireguard:
    # Host-level or sidecar
    mode: sidecar  # host | sidecar
    
    # IP allocation for WG interfaces
    subnet: 10.100.0.0/16
    
    # Listen port
    listenPort: 51820
    
    # Key storage
    keyStore:
      type: file  # file | vault | k8s-secret
      path: /opt/fcrunner/data/wireguard/keys
    
    # Peer discovery
    discovery:
      type: static  # static | dns | etcd | kubernetes
      
      # For static
      peers:
        - name: host-b
          endpoint: 10.0.0.2:51820
          publicKey: "abc123..."
          subnets:
            - 172.17.0.0/24
      
      # For DNS-based discovery
      # dns:
      #   zone: _wireguard._udp.fcrunner.internal
      
      # For etcd-based discovery
      # etcd:
      #   endpoints: ["https://etcd1:2379"]
      #   prefix: /fcrunner/wireguard/peers
  
  tailscale:
    # Auth method
    auth:
      type: authkey  # authkey | oauth
      
      # For authkey
      keyEnvVar: TAILSCALE_AUTH_KEY
      
      # For OAuth
      # oauth:
      #   clientID: ${TS_OAUTH_CLIENT_ID}
      #   clientSecret: ${TS_OAUTH_CLIENT_SECRET}
    
    # Mode
    mode: subnet-router  # subnet-router | per-vm
    
    # For subnet-router mode
    subnetRouter:
      advertiseRoutes: true
      acceptRoutes: true
    
    # For per-vm mode
    perVM:
      ephemeral: true
      tags: ["tag:fcrunner-vm"]
  
  headscale:
    serverURL: https://headscale.internal:8080
    
    # API key for creating nodes
    apiKeyEnvVar: HEADSCALE_API_KEY
    
    # Namespace for VMs
    namespace: fcrunner
```

### 8.2 API Extensions

```protobuf
// api/v1/overlay.proto

service OverlayService {
  // WireGuard operations
  rpc GetWireGuardStatus(GetWireGuardStatusRequest) returns (WireGuardStatus);
  rpc AddWireGuardPeer(AddWireGuardPeerRequest) returns (WireGuardPeer);
  rpc RemoveWireGuardPeer(RemoveWireGuardPeerRequest) returns (google.protobuf.Empty);
  rpc ListWireGuardPeers(ListWireGuardPeersRequest) returns (ListWireGuardPeersResponse);
  
  // Tailscale operations
  rpc GetTailscaleStatus(GetTailscaleStatusRequest) returns (TailscaleStatus);
  rpc GetVMTailscaleInfo(GetVMTailscaleInfoRequest) returns (VMTailscaleInfo);
}

message WireGuardStatus {
  string interface_name = 1;
  string public_key = 2;
  string listen_port = 3;
  repeated WireGuardPeer peers = 4;
}

message WireGuardPeer {
  string public_key = 1;
  string endpoint = 2;
  repeated string allowed_ips = 3;
  int64 last_handshake_seconds = 4;
  int64 rx_bytes = 5;
  int64 tx_bytes = 6;
}

message TailscaleStatus {
  string hostname = 1;
  string tailscale_ip = 2;
  repeated string advertised_routes = 3;
  bool online = 4;
}

message VMTailscaleInfo {
  string vm_id = 1;
  string tailscale_ip = 2;
  string hostname = 3;
  repeated string tags = 4;
  bool online = 5;
}
```

---

## 9. Security Considerations

### 9.1 Overlay-Specific Threats

| Threat | WireGuard (Host) | WireGuard (Sidecar) | Tailscale |
|--------|------------------|---------------------|-----------|
| **Key compromise** | All VMs exposed | Single VM exposed | Single VM exposed |
| **Lateral movement** | Via routing rules | Via WG peers | Via Tailscale ACL |
| **Traffic inspection** | Host can inspect | Host can inspect | Host can inspect |
| **Identity spoofing** | IP-based only | Crypto identity | Crypto identity |
| **Control plane attack** | N/A (no control plane) | Coordinator | Tailscale/Headscale |

### 9.2 Mitigation Requirements

```yaml
security_requirements:
  key_management:
    - Keys MUST be stored encrypted at rest
    - Keys MUST NOT be accessible to VMs (sidecar mode)
    - Key rotation SHOULD be supported
    
  network_segmentation:
    - Overlay peers MUST be explicitly configured
    - Default MUST be no overlay connectivity
    - Cross-tenant overlay MUST be prohibited
    
  authentication:
    - WireGuard: Public key authentication
    - Tailscale: Device authentication via control plane
    - All: mTLS for management API
    
  monitoring:
    - All overlay traffic SHOULD be logged (metadata only)
    - Peer connection/disconnection MUST be logged
    - Unusual traffic patterns SHOULD trigger alerts
```

---

## 10. Implementation Sequence

```yaml
phases:
  phase_2a:  # After base networking
    name: Same-Host VM-to-VM
    deliverables:
      - Peer policy schema
      - nftables peer set compilation
      - Dynamic peer updates on VM create/delete
      - Integration tests
    
    depends_on:
      - phase_2.batch_2  # Network policy API
  
  phase_2b:
    name: WireGuard Host-Level
    deliverables:
      - WireGuard configuration schema
      - Host WG interface management
      - Static peer configuration
      - Routing integration
    
    depends_on:
      - phase_2a
  
  phase_2c:
    name: WireGuard Sidecar
    deliverables:
      - Per-VM WG in network namespace
      - Key generation and storage
      - Peer discovery/coordination
      - Cross-host VM communication
    
    depends_on:
      - phase_2b
  
  phase_2d:
    name: Tailscale Integration
    deliverables:
      - Subnet router mode
      - Auth key management
      - Per-VM Tailscale (optional)
      - Tailscale ACL integration
    
    depends_on:
      - phase_2a
    
    parallel_with:
      - phase_2b
      - phase_2c
```

---

## 11. Open Questions

| Question | Options | Considerations |
|----------|---------|----------------|
| Which overlay types to support? | WireGuard only / Tailscale only / Both / Pluggable | Maintenance burden vs flexibility |
| Default overlay behavior? | None (opt-in) / Auto-mesh / Configurable | Security vs convenience |
| WireGuard mode? | Host / Sidecar / Guest / All three | Complexity vs security model fit |
| Tailscale auth method? | Auth key / OAuth / Both | Simplicity vs enterprise requirements |
| Key storage? | File only / Vault / K8s secrets / Pluggable | Deployment complexity vs security |
| Self-hosted control plane? | Headscale required / Optional / Not supported | External dependency tolerance |
| Cross-tenant overlay? | Never / Admin-configured / Tenant-controlled | Multi-tenancy model |

---

## Appendix A: Quick Start Examples

### A.1 Enable Same-Host VM-to-VM

```bash
# Create two VMs in same group
fcrunner vm create --name web1 --label group=web
fcrunner vm create --name web2 --label group=web

# Apply policy allowing group communication
cat << EOF | fcrunner policy apply -f -
apiVersion: fcrunner.io/v1
kind: NetworkPolicy
metadata:
  name: allow-web-group
spec:
  vmSelector:
    matchLabels:
      group: web
  peers:
    - name: same-group
      direction: both
      peerSelector:
        matchLabels:
          group: web
EOF

# Verify
fcrunner vm exec web1 -- ping -c1 172.16.0.3  # web2's IP
```

### A.2 Enable WireGuard Cross-Host

```bash
# On host A
fcrunner overlay configure wireguard \
  --mode host \
  --listen-port 51820 \
  --peer host-b=10.0.0.2:51820,pubkey=abc123...,subnets=172.17.0.0/24

# On host B
fcrunner overlay configure wireguard \
  --mode host \
  --listen-port 51820 \
  --peer host-a=10.0.0.1:51820,pubkey=def456...,subnets=172.16.0.0/24

# Test
fcrunner vm exec vm-on-host-a -- ping -c1 172.17.0.2  # VM on host B
```

### A.3 Enable Tailscale Subnet Router

```bash
# Set auth key
export TAILSCALE_AUTH_KEY=tskey-auth-xxx

# Enable subnet router
fcrunner overlay configure tailscale \
  --mode subnet-router \
  --advertise-routes 172.16.0.0/24

# From any Tailscale device
curl http://172.16.0.2:8080  # Reaches VM
```
