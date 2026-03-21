# All Things That Can Go Wrong When a Container Runs Inside a Firecracker VM  
*(Even if host → VM ping works)*

If host → Firecracker VM ping works, congrats: you’ve proven precisely **one thing** — that ICMP between those two IPs isn’t totally dead. Everything else can still be a dumpster fire. Here’s the complete failure safari, organized bottom-to-top.

---

## 1. Firecracker / VM Network Plumbing

Ping only proves link-local reachability. Common failures:

- **Bridge/TAP wired wrong for outbound**
  - TAP ↔ bridge OK, but:
    - No NAT/masquerade for VM → Internet.
    - VM attached to an isolated bridge with no routing.
- **Asymmetric routing**
  - Host→VM works.
  - VM→Internet leaves via the wrong interface → return packets die.
- **Reverse path filtering (`rp_filter`)**
  - Host or VM drops outbound replies that don’t match expected interface.

---

## 2. Guest OS (Inside the VM) Networking Issues

Containers inherit problems here.

- **Missing / wrong default route**
  - Local ping works; anything off-subnet fails.
- **Broken DNS**
  - Ping by IP works; hostnames fail.
- **Guest firewall (iptables/nftables/ufw)**
  - ICMP allowed; TCP/UDP blocked.
- **MTU mismatch**
  - Small pings OK, large packets fail or hang.
- **IP forwarding disabled**
  - Containers can’t reach outside via the VM because `ip_forward=0`.

---

## 3. Guest Kernel Limitations (Common in Firecracker Images)

Slim kernels bite hard:

- **Missing cgroup controllers**
  - containerd/Docker crash or silently misbehave.
- **Missing network modules**
  - Missing `overlay`, `br_netfilter`, `vxlan` = CNI fails.
- **Wrong cgroup version**
  - Runtime wants v2; VM only has v1.
- **Over-restrictive seccomp/capabilities**
  - Blocks syscalls needed by the container runtime.

---

## 4. Container Runtime / CNI Layer

Assuming guest networking “works,” now CNI may still break.

- **Docker/CNI bridge not created**
  - `docker0` / `cni0` missing or down.
- **Container not attached to any network**
  - `--network=none` or plugin failure.
- **Broken CNI config**
  - Wrong subnet, invalid plugin path, missing `.conflist`.
- **veth setup failure**
  - veth not moved into container namespace.
- **Missing NAT rules**
  - No `MASQUERADE` → container packets reach nowhere.
- **No default route inside container**
  - Looks up but doesn’t leave the namespace.

---

## 5. Firewalls Everywhere (Host ↔ Guest ↔ Container)

Three firewalls = three chances to ruin your day.

- **Host firewall only allows ICMP**
  - Ping OK; TCP/UDP dead.
- **Guest firewall blocks forwarding**
  - `FORWARD DROP` by default.
- **nftables vs iptables conflict**
  - Rules applied to the wrong backend.
- **Conntrack exhaustion**
  - Ping works (stateless-ish); connections fail under load.

---

## 6. Ping-Specific Container Issues

If *containers* can’t ping:

- **Missing `CAP_NET_RAW`**
  - Non-privileged containers can’t create ICMP sockets.
- **Ping binary absent**
  - Distroless, Alpine minimal, scratch-based images.
- **Wrong `ping_group_range`**
  - Kernel blocks unprivileged ping.

---

## 7. Addressing / Subnet Problems

Ping might still work locally while everything else collapses.

- **Overlapping CIDRs**
  - container subnet == VM subnet == host subnet → routing chaos.
- **Duplicate IPs**
  - VM shares an address accidentally.
- **Wrong gateway for containers**
  - CNI assigns a gateway not reachable from the bridge.

---

## 8. DNS / Proxy / Name-Resolution Disaster Class

Ping by IP works; everything else soaks in misery.

- **Bad `/etc/resolv.conf` inside VM or container**
- **VPN split-DNS rules not visible inside VM**
- **Missing proxy configuration**
  - Host uses proxies; VM does not.

---

## 9. Subtle Network Bugs & Performance

These can be misleading:

- **Bad offloading settings (TSO/GSO/GRO)**
  - Some combos corrupt packets going through TAP/veth.
- **Broken PMTU / no ICMP fragmentation needed**
  - Only large packets fail; pings succeed.
- **Clock skew in VM**
  - TLS and auth failures look like “network dead.”

---

## 10. Firecracker / Jailer Isolation Oddities

More esoteric, but real:

- **Firecracker in jailer netns**
  - TAP reachable from host but invisible to your NAT bridge.
- **Multiple microVMs sharing interfaces**
  - ARP confusion; packets vanish into the ether.
- **vsock confusion**
  - App expects IP networking but only vsock is available.

---

## 11. Container Behavior vs VM Lifecycle

Powering off/on or ephemeral behavior breaks networking indirectly.

- **Runtime state not persisted**
  - containerd metadata gone → containers restart w/o network.
- **OverlayFS + ephemeral root**
  - Network configs silently wiped.
- **Race conditions at boot**
  - CNI starts before network is up; containers boot with no routes.

