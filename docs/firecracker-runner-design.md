# Firecracker microVM Runner Platform

## Design & Implementation Requirements Specification

**Version:** 0.1.0-draft  
**Status:** Design Phase  
**Last Updated:** 2025-12-21

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [System Overview](#2-system-overview)
3. [Architecture](#3-architecture)
4. [Host Requirements](#4-host-requirements)
5. [Guest Requirements](#5-guest-requirements)
6. [Network Architecture](#6-network-architecture)
7. [API Gateway / Reverse Proxy](#7-api-gateway--reverse-proxy)
8. [Management API](#8-management-api)
9. [Security Model](#9-security-model)
10. [Platform Support](#10-platform-support)
11. [Implementation Phases](#11-implementation-phases)
12. [Operational Requirements](#12-operational-requirements)

---

## 1. Executive Summary

### 1.1 Purpose

Build a secure, ephemeral container execution platform using Firecracker microVMs to provide strong isolation for untrusted workloads. The system runs Docker containers inside microVMs with:

- Hardware-level isolation (KVM)
- Host-enforced network policy (L3/L4)
- API gateway with logging and future AI-powered introspection
- Remote management via secure gRPC API

### 1.2 Key Design Principles

| Principle | Rationale |
|-----------|-----------|
| **Defense in depth** | Multiple isolation layers; no single point of failure |
| **Host-side enforcement** | Never trust the guest; all security controls on host |
| **Ephemeral by default** | VMs are disposable; state is external |
| **Minimal attack surface** | Reduce exposed interfaces, syscalls, devices |
| **Observable** | Comprehensive logging at every boundary |

### 1.3 Target Use Cases

- Untrusted code execution (AI agents, user-submitted code)
- CI/CD job isolation
- Multi-tenant workload separation
- Sandboxed development environments

---

## 2. System Overview

### 2.1 Component Diagram

```
┌────────────────────────────────────────────────────────────────────────────────┐
│                                  HOST SYSTEM                                   │
│                                                                                │
│  ┌──────────────────────────────────────────────────────────────────────────┐ │
│  │                         MANAGEMENT PLANE                                  │ │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐  │ │
│  │  │   gRPC API  │  │  VM Manager │  │ Config Store│  │  Metrics/Logs   │  │ │
│  │  │   Server    │──│  (Orchest.) │──│  (SQLite/   │──│  (OTLP Export)  │  │ │
│  │  │   (mTLS)    │  │             │  │   etcd)     │  │                 │  │ │
│  │  └─────────────┘  └──────┬──────┘  └─────────────┘  └─────────────────┘  │ │
│  └──────────────────────────┼───────────────────────────────────────────────┘ │
│                             │                                                  │
│  ┌──────────────────────────┼───────────────────────────────────────────────┐ │
│  │                     DATA PLANE                                            │ │
│  │                          │                                                │ │
│  │  ┌───────────────────────▼────────────────────────────────────────────┐  │ │
│  │  │                    NETWORK FABRIC                                   │  │ │
│  │  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────────┐ │  │ │
│  │  │  │  nftables   │  │ API Gateway │  │      Traffic Controller     │ │  │ │
│  │  │  │  (L3/L4     │  │  (L7 Proxy) │  │   (Rate Limit, QoS, DPI)    │ │  │ │
│  │  │  │  Policy)    │  │  + Logging  │  │                             │ │  │ │
│  │  │  └──────┬──────┘  └──────┬──────┘  └──────────────┬──────────────┘ │  │ │
│  │  │         │                │                        │                │  │ │
│  │  │         └────────────────┴────────────────────────┘                │  │ │
│  │  │                          │                                         │  │ │
│  │  │                    ┌─────┴─────┐                                   │  │ │
│  │  │                    │  bridge0  │                                   │  │ │
│  │  │                    └─────┬─────┘                                   │  │ │
│  │  └──────────────────────────┼─────────────────────────────────────────┘  │ │
│  │                             │                                            │ │
│  │  ┌──────────────────────────┼───────────────────────────────────────┐   │ │
│  │  │                    VM INSTANCES                                   │   │ │
│  │  │      ┌───────────────────┼───────────────────┐                   │   │ │
│  │  │      │                   │                   │                   │   │ │
│  │  │  ┌───┴───┐          ┌────┴────┐         ┌────┴────┐             │   │ │
│  │  │  │ tap0  │          │  tap1   │         │  tapN   │             │   │ │
│  │  │  └───┬───┘          └────┬────┘         └────┬────┘             │   │ │
│  │  │      │                   │                   │                   │   │ │
│  │  │  ┌───┴───────────┐  ┌────┴────────────┐ ┌────┴────────────┐    │   │ │
│  │  │  │   Jailer      │  │   Jailer        │ │   Jailer        │    │   │ │
│  │  │  │ ┌───────────┐ │  │ ┌─────────────┐ │ │ ┌─────────────┐ │    │   │ │
│  │  │  │ │Firecracker│ │  │ │ Firecracker │ │ │ │ Firecracker │ │    │   │ │
│  │  │  │ │  ┌─────┐  │ │  │ │   ┌─────┐   │ │ │ │   ┌─────┐   │ │    │   │ │
│  │  │  │ │  │Guest│  │ │  │ │   │Guest│   │ │ │ │   │Guest│   │ │    │   │ │
│  │  │  │ │  │ VM  │  │ │  │ │   │ VM  │   │ │ │ │   │ VM  │   │ │    │   │ │
│  │  │  │ └──┴─────┴──┘ │  │ └───┴─────┴───┘ │ │ └───┴─────┴───┘ │    │   │ │
│  │  │  └───────────────┘  └─────────────────┘ └─────────────────┘    │   │ │
│  │  │       cgroup0            cgroup1             cgroupN           │   │ │
│  │  │       netns0             netns1              netnsN            │   │ │
│  │  └────────────────────────────────────────────────────────────────┘   │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │                     OBSERVABILITY PLANE                                │ │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌──────────────┐  │ │
│  │  │  Tetragon   │  │  Prometheus │  │   Loki /    │  │   Audit      │  │ │
│  │  │  (eBPF)     │  │  Metrics    │  │   Logging   │  │   Trail      │  │ │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └──────────────┘  │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 2.2 Data Flow

```
External Request
       │
       ▼
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Ingress    │────▶│  API Gateway │────▶│   nftables   │
│   (TLS)      │     │  (Log/Route) │     │   (Policy)   │
└──────────────┘     └──────────────┘     └──────┬───────┘
                                                  │
                            ┌─────────────────────┘
                            │ (if allowed)
                            ▼
                     ┌──────────────┐
                     │   tap iface  │
                     └──────┬───────┘
                            │
                     ┌──────▼───────┐
                     │  Firecracker │
                     │    (virtio)  │
                     └──────┬───────┘
                            │
                     ┌──────▼───────┐
                     │  Guest eth0  │
                     └──────┬───────┘
                            │
                     ┌──────▼───────┐
                     │   Docker     │
                     │  Container   │
                     └──────────────┘
```

---

## 3. Architecture

### 3.1 Component Responsibilities

| Component | Responsibility | Trust Level |
|-----------|---------------|-------------|
| **gRPC API Server** | External management interface | Trusted (mTLS) |
| **VM Manager** | Lifecycle orchestration (create/start/stop/destroy) | Trusted |
| **Jailer** | Process isolation, namespace setup | Trusted |
| **Firecracker** | VMM, device emulation | Trusted (but minimal) |
| **nftables** | L3/L4 network policy enforcement | Trusted |
| **API Gateway** | L7 proxy, logging, future introspection | Trusted |
| **Guest VM** | Runs untrusted workloads | **UNTRUSTED** |
| **Tetragon** | Host-side syscall monitoring | Trusted |

### 3.2 Trust Boundaries

```
TRUST BOUNDARY 1: Host Kernel ←→ Firecracker Process
  - Enforced by: seccomp, namespaces, capabilities
  - Attack surface: syscalls allowed by seccomp filter

TRUST BOUNDARY 2: Firecracker VMM ←→ Guest VM
  - Enforced by: KVM (hardware virtualization)
  - Attack surface: virtio devices, KVM ABI

TRUST BOUNDARY 3: Host Network ←→ Guest Network
  - Enforced by: nftables, network namespaces
  - Attack surface: L3/L4 protocols, TCP/IP stack

TRUST BOUNDARY 4: External ←→ Management API
  - Enforced by: mTLS, RBAC
  - Attack surface: gRPC endpoints
```

---

## 4. Host Requirements

### 4.1 Hardware Requirements

| Resource | Minimum | Recommended | Notes |
|----------|---------|-------------|-------|
| CPU | 4 cores | 16+ cores | VT-x/AMD-V required |
| RAM | 8 GB | 64+ GB | Overcommit not recommended |
| Storage | 50 GB SSD | 500+ GB NVMe | For VM images, logs |
| Network | 1 Gbps | 10 Gbps | Per-VM bandwidth limiting |

### 4.2 Software Requirements

#### 4.2.1 Operating System

```yaml
supported_os:
  linux:
    distributions:
      - ubuntu: "22.04+"
      - debian: "12+"
      - amazon_linux: "2023+"
      - fedora: "38+"
    kernel: "5.10+"
    required_features:
      - CONFIG_KVM=y
      - CONFIG_VHOST_NET=y
      - CONFIG_TUN=y
      - CONFIG_BRIDGE=y
      - CONFIG_NETFILTER=y
      - CONFIG_BPF=y
      - CONFIG_CGROUPS=y
```

#### 4.2.2 Required Packages

```bash
# Core virtualization
firecracker          # v1.5.0+
jailer               # Bundled with Firecracker

# Networking
nftables             # v1.0.0+
iproute2             # For ip/tc commands
iptables             # Fallback compatibility

# Container/Image handling
skopeo               # OCI image pulling
umoci                # OCI image unpacking

# Observability (optional but recommended)
tetragon             # eBPF monitoring
prometheus-node-exporter

# Build dependencies
rust                 # For custom components
go                   # For API gateway
```

#### 4.2.3 Kernel Modules

```bash
# Required modules
kvm
kvm_intel  # or kvm_amd
tun
vhost_net
bridge
nf_tables
nf_conntrack
```

### 4.3 Host Configuration

#### 4.3.1 Kernel Parameters

```bash
# /etc/sysctl.d/99-fcrunner.conf

# Security hardening
kernel.kptr_restrict=2
kernel.dmesg_restrict=1
kernel.unprivileged_bpf_disabled=1
kernel.perf_event_paranoid=3
kernel.randomize_va_space=2
kernel.yama.ptrace_scope=2

# Network security
net.ipv4.conf.all.rp_filter=1
net.ipv4.conf.default.rp_filter=1
net.ipv4.conf.all.accept_source_route=0
net.ipv4.conf.all.accept_redirects=0
net.ipv4.conf.all.send_redirects=0
net.ipv4.conf.all.log_martians=1
net.ipv4.icmp_echo_ignore_broadcasts=1
net.ipv4.icmp_ignore_bogus_error_responses=1

# Enable IP forwarding (required for VM networking)
net.ipv4.ip_forward=1
net.ipv6.conf.all.forwarding=1

# Connection tracking
net.netfilter.nf_conntrack_max=1048576

# Memory management
vm.swappiness=10
vm.overcommit_memory=0
vm.max_map_count=262144

# File descriptors
fs.file-max=2097152
fs.nr_open=2097152
```

#### 4.3.2 Cgroup v2 Setup

```bash
# Verify cgroup v2
mount | grep cgroup2

# Create hierarchy for VMs
mkdir -p /sys/fs/cgroup/fcrunner
echo "+cpu +memory +io +pids" > /sys/fs/cgroup/fcrunner/cgroup.subtree_control

# Set global limits for all VMs
echo "80" > /sys/fs/cgroup/fcrunner/cpu.weight  # 80% of host CPU max
echo "$(($(nproc) * 100000)) 100000" > /sys/fs/cgroup/fcrunner/cpu.max
```

#### 4.3.3 Directory Structure

```
/opt/fcrunner/
├── bin/
│   ├── fcrunner-api          # gRPC server binary
│   ├── fcrunner-gateway      # API gateway binary
│   └── fc-net-setup          # Network setup helper
├── etc/
│   ├── fcrunner.yaml         # Main configuration
│   ├── network-policies/     # Network policy definitions
│   │   └── default.yaml
│   ├── certs/                # TLS certificates
│   │   ├── ca.crt
│   │   ├── server.crt
│   │   └── server.key
│   └── seccomp/              # Custom seccomp profiles
│       └── firecracker.json
├── lib/
│   ├── kernels/              # Guest kernels
│   │   └── vmlinux-5.10
│   └── rootfs/               # Base root filesystems
│       └── docker-rootfs.ext4
├── run/
│   ├── vms/                  # Per-VM runtime state
│   │   └── {vm-id}/
│   │       ├── firecracker.sock
│   │       ├── root/         # Jailer chroot
│   │       └── logs/
│   └── gateway/              # Gateway runtime
│       └── gateway.sock
├── data/
│   ├── images/               # Cached container images
│   ├── overlays/             # VM overlay filesystems
│   └── state.db              # SQLite state database
└── logs/
    ├── api/
    ├── gateway/
    └── vms/
```

### 4.4 Host Security Requirements

#### 4.4.1 User/Permission Model

```bash
# Create dedicated users
useradd -r -s /usr/sbin/nologin fcrunner       # Main service user
useradd -r -s /usr/sbin/nologin fcrunner-vm    # VM process user (uid range)

# UID range for VM processes (each VM gets unique UID)
# Range: 100000-165535 (65536 UIDs)
echo "fcrunner-vm:100000:65536" >> /etc/subuid
echo "fcrunner-vm:100000:65536" >> /etc/subgid

# Permissions
chown -R fcrunner:fcrunner /opt/fcrunner
chmod 750 /opt/fcrunner/etc/certs
chmod 600 /opt/fcrunner/etc/certs/*.key

# KVM access
usermod -aG kvm fcrunner
```

#### 4.4.2 Capabilities

```yaml
# fcrunner-api service capabilities
capabilities:
  required:
    - CAP_NET_ADMIN      # Network namespace, nftables
    - CAP_SYS_ADMIN      # Mount namespaces, cgroups
    - CAP_SETUID         # Switch to VM user
    - CAP_SETGID         # Switch to VM group
  dropped:
    - ALL others

# Firecracker process (via jailer)
capabilities:
  required: []           # No capabilities needed
  dropped:
    - ALL
```

#### 4.4.3 Seccomp Profile

```json
{
  "defaultAction": "SCMP_ACT_ERRNO",
  "defaultErrnoRet": 1,
  "archMap": [
    {
      "architecture": "SCMP_ARCH_X86_64",
      "subArchitectures": []
    }
  ],
  "syscalls": [
    {
      "names": [
        "accept4", "brk", "clock_gettime", "close", "dup", "epoll_create1",
        "epoll_ctl", "epoll_pwait", "epoll_wait", "eventfd2", "exit",
        "exit_group", "fallocate", "fcntl", "fstat", "fsync", "ftruncate",
        "futex", "getrandom", "getpid", "gettid", "gettimeofday",
        "io_uring_enter", "io_uring_register", "io_uring_setup",
        "ioctl", "listen", "lseek", "madvise", "membarrier", "mmap",
        "mprotect", "mremap", "munmap", "nanosleep", "newfstatat",
        "openat", "pipe2", "ppoll", "prctl", "pread64", "preadv",
        "pwrite64", "pwritev", "read", "readv", "recvfrom", "recvmsg",
        "rt_sigaction", "rt_sigprocmask", "rt_sigreturn", "sched_getaffinity",
        "sched_yield", "sendmsg", "sendto", "set_robust_list",
        "sigaltstack", "socket", "timerfd_create", "timerfd_settime",
        "tkill", "write", "writev"
      ],
      "action": "SCMP_ACT_ALLOW"
    },
    {
      "names": ["ioctl"],
      "action": "SCMP_ACT_ALLOW",
      "args": [
        {
          "index": 1,
          "value": 44545,
          "op": "SCMP_CMP_EQ"
        }
      ],
      "comment": "KVM_RUN"
    }
  ]
}
```

### 4.5 Host Monitoring Requirements

#### 4.5.1 Tetragon Policies

```yaml
# /opt/fcrunner/etc/tetragon/firecracker-monitor.yaml
apiVersion: cilium.io/v1alpha1
kind: TracingPolicy
metadata:
  name: fcrunner-firecracker-monitor
spec:
  kprobes:
    # Detect unexpected process execution
    - call: "__x64_sys_execve"
      selectors:
        - matchBinaries:
            - operator: In
              values:
                - "/usr/bin/firecracker"
        - matchActions:
            - action: Sigkill
              argError: -1
      message: "Firecracker attempted to exec - possible escape"
    
    # Monitor file access outside allowed paths
    - call: "__x64_sys_openat"
      selectors:
        - matchBinaries:
            - operator: In
              values:
                - "/usr/bin/firecracker"
        - matchArgs:
            - index: 1
              operator: NotPrefix
              values:
                - "/opt/fcrunner/run/vms/"
                - "/dev/kvm"
                - "/dev/net/tun"
                - "/dev/urandom"
        - matchActions:
            - action: Post
              rateLimit: "1s"
      message: "Firecracker accessed unexpected file"
    
    # Monitor network syscalls
    - call: "__x64_sys_connect"
      selectors:
        - matchBinaries:
            - operator: In
              values:
                - "/usr/bin/firecracker"
        - matchActions:
            - action: Sigkill
      message: "Firecracker attempted outbound connection"
    
    # Detect ptrace attempts
    - call: "__x64_sys_ptrace"
      selectors:
        - matchBinaries:
            - operator: In
              values:
                - "/usr/bin/firecracker"
        - matchActions:
            - action: Sigkill
      message: "Firecracker attempted ptrace"
```

#### 4.5.2 Metrics to Collect

```yaml
metrics:
  host_level:
    - name: fcrunner_vms_total
      type: gauge
      labels: [status]  # running, stopped, failed
    
    - name: fcrunner_vm_cpu_usage_seconds
      type: counter
      labels: [vm_id]
    
    - name: fcrunner_vm_memory_bytes
      type: gauge
      labels: [vm_id, type]  # rss, cache, swap
    
    - name: fcrunner_vm_network_bytes
      type: counter
      labels: [vm_id, direction]  # rx, tx
    
    - name: fcrunner_vm_network_packets
      type: counter
      labels: [vm_id, direction, status]  # allowed, dropped
    
    - name: fcrunner_api_requests_total
      type: counter
      labels: [method, status]
    
    - name: fcrunner_security_events_total
      type: counter
      labels: [event_type, severity]
    
    - name: fcrunner_vm_boot_duration_seconds
      type: histogram
      labels: [image]
```

---

## 5. Guest Requirements

### 5.1 Guest Kernel

#### 5.1.1 Kernel Configuration

```bash
# Minimal kernel config for Firecracker guests
CONFIG_VIRTIO=y
CONFIG_VIRTIO_BLK=y
CONFIG_VIRTIO_NET=y
CONFIG_VIRTIO_CONSOLE=y
CONFIG_VIRTIO_MMIO=y
CONFIG_EXT4_FS=y
CONFIG_OVERLAY_FS=y
CONFIG_SQUASHFS=y

# Container support
CONFIG_NAMESPACES=y
CONFIG_CGROUPS=y
CONFIG_CGROUP_DEVICE=y
CONFIG_CGROUP_FREEZER=y
CONFIG_CGROUP_SCHED=y
CONFIG_CGROUP_CPUACCT=y
CONFIG_CGROUP_PIDS=y
CONFIG_MEMCG=y
CONFIG_BLK_CGROUP=y

# Docker requirements
CONFIG_NETFILTER=y
CONFIG_BRIDGE=y
CONFIG_VETH=y
CONFIG_NETFILTER_XT_MATCH_CONNTRACK=y
CONFIG_IP_NF_NAT=y

# Security
CONFIG_SECCOMP=y
CONFIG_SECURITY=y
CONFIG_SECURITY_APPARMOR=y

# Disable unnecessary features
# CONFIG_MODULES is not set
# CONFIG_SMP is not set (single vCPU builds)
# CONFIG_ACPI is not set
# CONFIG_PCI is not set
```

#### 5.1.2 Kernel Build

```bash
#!/bin/bash
# build-guest-kernel.sh

KERNEL_VERSION="5.10.217"
KERNEL_URL="https://cdn.kernel.org/pub/linux/kernel/v5.x/linux-${KERNEL_VERSION}.tar.xz"

# Download and extract
wget "$KERNEL_URL"
tar xf "linux-${KERNEL_VERSION}.tar.xz"
cd "linux-${KERNEL_VERSION}"

# Apply Firecracker-optimized config
cp /path/to/guest-kernel.config .config
make olddefconfig

# Build
make -j$(nproc) vmlinux

# Output: vmlinux (uncompressed kernel)
cp vmlinux /opt/fcrunner/lib/kernels/vmlinux-${KERNEL_VERSION}
```

### 5.2 Root Filesystem

#### 5.2.1 Base Image Requirements

```yaml
rootfs:
  format: ext4
  size: 2GB  # Expandable
  
  base_packages:
    - systemd (or openrc for lighter init)
    - containerd (1.7+)
    - runc (1.1+)
    - docker-cli (optional, for UX)
    - iptables
    - iproute2
    - ca-certificates
    - curl (for healthchecks)
  
  optimizations:
    - Remove documentation
    - Remove man pages
    - Minimal locale (C.UTF-8 only)
    - No package manager cache
    - Read-only where possible
```

#### 5.2.2 Rootfs Build Script

```bash
#!/bin/bash
# build-guest-rootfs.sh

set -euo pipefail

ROOTFS_SIZE="2G"
ROOTFS_FILE="/opt/fcrunner/lib/rootfs/docker-rootfs.ext4"
WORK_DIR=$(mktemp -d)
MOUNT_DIR="${WORK_DIR}/rootfs"

# Create ext4 image
truncate -s "$ROOTFS_SIZE" "$ROOTFS_FILE"
mkfs.ext4 -F "$ROOTFS_FILE"

# Mount and bootstrap
mkdir -p "$MOUNT_DIR"
mount -o loop "$ROOTFS_FILE" "$MOUNT_DIR"

# Debootstrap minimal Debian
debootstrap --variant=minbase --include=systemd,dbus,containerd,runc \
  bookworm "$MOUNT_DIR" http://deb.debian.org/debian

# Configure systemd
cat > "${MOUNT_DIR}/etc/systemd/system/containerd.service" << 'EOF'
[Unit]
Description=containerd container runtime
After=network.target

[Service]
ExecStart=/usr/bin/containerd
Delegate=yes
KillMode=process
Restart=always

[Install]
WantedBy=multi-user.target
EOF

# Enable services
chroot "$MOUNT_DIR" systemctl enable containerd

# Configure networking
cat > "${MOUNT_DIR}/etc/systemd/network/10-eth0.network" << 'EOF'
[Match]
Name=eth0

[Network]
DHCP=no
Address=172.16.0.2/24
Gateway=172.16.0.1
DNS=172.16.0.1
EOF

chroot "$MOUNT_DIR" systemctl enable systemd-networkd

# Install guest agent
cp /path/to/fcrunner-guest-agent "${MOUNT_DIR}/usr/local/bin/"
cat > "${MOUNT_DIR}/etc/systemd/system/fcrunner-agent.service" << 'EOF'
[Unit]
Description=FCRunner Guest Agent
After=network.target containerd.service

[Service]
ExecStart=/usr/local/bin/fcrunner-guest-agent
Restart=always

[Install]
WantedBy=multi-user.target
EOF

chroot "$MOUNT_DIR" systemctl enable fcrunner-agent

# Cleanup
rm -rf "${MOUNT_DIR}/var/cache/apt"/*
rm -rf "${MOUNT_DIR}/var/lib/apt/lists"/*
rm -rf "${MOUNT_DIR}/usr/share/doc"/*
rm -rf "${MOUNT_DIR}/usr/share/man"/*

# Unmount
umount "$MOUNT_DIR"
rm -rf "$WORK_DIR"

# Shrink image
e2fsck -f "$ROOTFS_FILE"
resize2fs -M "$ROOTFS_FILE"
```

### 5.3 Guest Agent

#### 5.3.1 Responsibilities

```yaml
guest_agent:
  communication: virtio-vsock  # CID assigned at boot
  
  functions:
    - name: healthcheck
      description: Report VM health status
      interval: 5s
    
    - name: container_run
      description: Execute container via containerd
      input:
        image: string
        command: [string]
        env: map[string]string
        mounts: [mount_spec]
        resources: resource_spec
    
    - name: container_stop
      description: Stop running container
      input:
        container_id: string
        timeout: duration
    
    - name: container_logs
      description: Stream container logs
      input:
        container_id: string
        follow: bool
        tail: int
    
    - name: exec
      description: Execute command in container
      input:
        container_id: string
        command: [string]
    
    - name: metrics
      description: Report resource usage
      interval: 10s
```

#### 5.3.2 Agent Protocol (vsock)

```protobuf
// guest-agent.proto
syntax = "proto3";
package fcrunner.guest;

service GuestAgent {
  // Health
  rpc Healthcheck(HealthRequest) returns (HealthResponse);
  
  // Container lifecycle
  rpc RunContainer(RunContainerRequest) returns (RunContainerResponse);
  rpc StopContainer(StopContainerRequest) returns (StopContainerResponse);
  rpc ContainerStatus(ContainerStatusRequest) returns (ContainerStatusResponse);
  
  // Streaming
  rpc ContainerLogs(ContainerLogsRequest) returns (stream LogEntry);
  rpc Exec(stream ExecRequest) returns (stream ExecResponse);
  
  // Metrics
  rpc GetMetrics(MetricsRequest) returns (MetricsResponse);
}

message RunContainerRequest {
  string image = 1;
  repeated string command = 2;
  map<string, string> env = 3;
  repeated Mount mounts = 4;
  Resources resources = 5;
  NetworkConfig network = 6;
}

message Resources {
  int64 memory_bytes = 1;
  int64 cpu_shares = 2;
  int64 cpu_quota = 3;
  int64 pids_limit = 4;
}

message NetworkConfig {
  repeated PortMapping ports = 1;
}

message PortMapping {
  int32 container_port = 1;
  int32 host_port = 2;  // Mapped at host level
  string protocol = 3;  // tcp, udp
}
```

### 5.4 Guest Security Configuration

#### 5.4.1 Boot Parameters

```bash
# Kernel command line
console=ttyS0 \
reboot=k \
panic=1 \
pci=off \
nomodules \
init=/sbin/init \
root=/dev/vda \
ro \
quiet \
systemd.journald.forward_to_console=0 \
random.trust_cpu=on
```

#### 5.4.2 Guest Hardening

```bash
# /etc/sysctl.d/99-security.conf (in guest rootfs)

# Restrict kernel pointers
kernel.kptr_restrict=2
kernel.dmesg_restrict=1

# Disable core dumps
kernel.core_pattern=|/bin/false
fs.suid_dumpable=0

# Network restrictions
net.ipv4.conf.all.accept_redirects=0
net.ipv4.conf.all.secure_redirects=0
net.ipv4.conf.all.send_redirects=0
net.ipv4.conf.all.accept_source_route=0

# Disable IPv6 if not needed
net.ipv6.conf.all.disable_ipv6=1
```

---

## 6. Network Architecture

### 6.1 Network Topology

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           HOST                                          │
│                                                                         │
│  External Network                                                       │
│       │                                                                 │
│       ▼                                                                 │
│  ┌─────────┐                                                           │
│  │  eth0   │  (Host physical/virtual NIC)                              │
│  └────┬────┘                                                           │
│       │                                                                 │
│       ▼                                                                 │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                    nftables (PREROUTING)                        │   │
│  │  - DNAT for exposed VM ports                                    │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│       │                                                                 │
│       ▼                                                                 │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                    API Gateway (Optional L7)                     │   │
│  │  - HTTP(S) termination                                          │   │
│  │  - Request logging                                               │   │
│  │  - Future: AI introspection                                     │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│       │                                                                 │
│       ▼                                                                 │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                    nftables (FORWARD)                           │   │
│  │  - L3/L4 policy enforcement                                     │   │
│  │  - Per-VM rules                                                 │   │
│  │  - Rate limiting                                                │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│       │                                                                 │
│       ▼                                                                 │
│  ┌──────────────┐                                                      │
│  │   fcbr0      │  (Linux bridge for VM connectivity)                  │
│  │  172.16.0.1  │                                                      │
│  └──────┬───────┘                                                      │
│         │                                                               │
│    ┌────┴────┬────────────┬────────────┐                              │
│    │         │            │            │                              │
│    ▼         ▼            ▼            ▼                              │
│ ┌──────┐ ┌──────┐    ┌──────┐    ┌──────┐                            │
│ │tap-a │ │tap-b │    │tap-c │    │tap-N │                            │
│ │netns │ │netns │    │netns │    │netns │                            │
│ └──┬───┘ └──┬───┘    └──┬───┘    └──┬───┘                            │
│    │        │           │           │                                 │
│ ┌──┴────┐ ┌─┴─────┐  ┌──┴────┐  ┌──┴────┐                           │
│ │ VM-A  │ │ VM-B  │  │ VM-C  │  │ VM-N  │                           │
│ │.2     │ │.3     │  │.4     │  │.X     │                           │
│ └───────┘ └───────┘  └───────┘  └───────┘                           │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘

IP Allocation:
  Bridge:     172.16.0.1/24
  VM Range:   172.16.0.2-254/24
  DHCP:       Not used (static assignment)
```

### 6.2 Network Policy Model

#### 6.2.1 Policy Schema

```yaml
# network-policy.yaml
apiVersion: fcrunner.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny
spec:
  # Applies to VMs matching these labels
  vmSelector:
    matchLabels:
      tenant: untrusted
  
  # Default policy when no rules match
  defaultAction: drop
  
  # Ingress rules (traffic TO the VM)
  ingress:
    - name: allow-http
      from:
        - cidr: 0.0.0.0/0
      ports:
        - port: 80
          protocol: tcp
        - port: 443
          protocol: tcp
    
    - name: allow-ssh-internal
      from:
        - cidr: 10.0.0.0/8
      ports:
        - port: 22
          protocol: tcp
  
  # Egress rules (traffic FROM the VM)
  egress:
    - name: allow-dns
      to:
        - cidr: 172.16.0.1/32  # Host DNS resolver
      ports:
        - port: 53
          protocol: udp
        - port: 53
          protocol: tcp
    
    - name: allow-https-out
      to:
        - cidr: 0.0.0.0/0
      ports:
        - port: 443
          protocol: tcp
    
    - name: deny-metadata
      to:
        - cidr: 169.254.169.254/32
      action: drop
    
    - name: deny-internal
      to:
        - cidr: 10.0.0.0/8
        - cidr: 172.16.0.0/12
        - cidr: 192.168.0.0/16
      action: drop
```

#### 6.2.2 Policy Compilation to nftables

```bash
#!/bin/bash
# compile-network-policy.sh

VM_ID="$1"
VM_IP="$2"
TAP_IFACE="tap-${VM_ID}"

# Generate nftables rules from policy
cat << EOF
table inet fcrunner_${VM_ID} {
    chain input {
        type filter hook input priority 0; policy drop;
        
        # Allow established/related
        ct state established,related accept
        
        # Allow ICMP (ping)
        ip protocol icmp accept
        
        # Specific ingress rules
        tcp dport 80 accept comment "allow-http"
        tcp dport 443 accept comment "allow-https"
        ip saddr 10.0.0.0/8 tcp dport 22 accept comment "allow-ssh-internal"
    }
    
    chain output {
        type filter hook output priority 0; policy drop;
        
        # Allow established/related
        ct state established,related accept
        
        # Allow DNS to host
        ip daddr 172.16.0.1 udp dport 53 accept
        ip daddr 172.16.0.1 tcp dport 53 accept
        
        # Allow HTTPS out
        tcp dport 443 accept
        
        # Block metadata service
        ip daddr 169.254.169.254 drop comment "block-metadata"
        
        # Block internal networks
        ip daddr 10.0.0.0/8 drop comment "block-internal"
        ip daddr 172.16.0.0/12 drop comment "block-internal"
        ip daddr 192.168.0.0/16 drop comment "block-internal"
    }
    
    chain forward {
        type filter hook forward priority 0; policy drop;
        
        # Apply same rules for forwarded traffic
        iifname "${TAP_IFACE}" jump output
        oifname "${TAP_IFACE}" jump input
    }
}
EOF
```

### 6.3 Network Setup Per-VM

```bash
#!/bin/bash
# setup-vm-network.sh

VM_ID="$1"
VM_IP="$2"
VM_MAC="$3"

TAP_IFACE="tap-${VM_ID}"
NETNS="fcns-${VM_ID}"
BRIDGE="fcbr0"

# Create network namespace
ip netns add "$NETNS"

# Create tap device
ip tuntap add dev "$TAP_IFACE" mode tap
ip link set "$TAP_IFACE" up

# Move tap to namespace (for isolation)
# Actually, keep tap in root ns but apply nftables per-interface
# This allows bridge connectivity while enforcing policy

# Add to bridge
ip link set "$TAP_IFACE" master "$BRIDGE"

# Set MAC address filtering (optional, defense in depth)
bridge fdb add "$VM_MAC" dev "$TAP_IFACE" master static

# Apply rate limiting via tc
tc qdisc add dev "$TAP_IFACE" root tbf rate 100mbit burst 32kbit latency 400ms

# Apply nftables policy
nft -f "/opt/fcrunner/run/vms/${VM_ID}/nftables.conf"

# Enable ARP proxy on bridge for VM
echo 1 > /proc/sys/net/ipv4/conf/${TAP_IFACE}/proxy_arp

echo "Network setup complete for VM ${VM_ID}"
echo "  TAP: ${TAP_IFACE}"
echo "  IP:  ${VM_IP}"
echo "  MAC: ${VM_MAC}"
```

### 6.4 DNS Resolution

```yaml
# Host-side DNS proxy (per-VM isolation)
dns:
  resolver: coredns  # or dnsmasq
  
  config: |
    .:53 {
      # Forward to upstream
      forward . 8.8.8.8 8.8.4.4 {
        policy sequential
      }
      
      # Block internal domain resolution for untrusted VMs
      template IN A internal.company.com {
        rcode NXDOMAIN
      }
      
      # Logging
      log
      
      # Cache
      cache 300
    }
```

---

## 7. API Gateway / Reverse Proxy

### 7.1 Overview

The API Gateway sits between external traffic and VMs, providing:
- L7 visibility and logging
- Future AI-powered request introspection
- Rate limiting and throttling
- TLS termination (optional)

### 7.2 Architecture

```
                              ┌─────────────────────────────────┐
                              │         API Gateway             │
External ──────────────────▶  │                                 │
Request                       │  ┌───────────────────────────┐  │
                              │  │     Request Pipeline      │  │
                              │  │                           │  │
                              │  │  1. TLS Termination       │  │
                              │  │  2. Request Parsing       │  │
                              │  │  3. Logging (structured)  │  │
                              │  │  4. Rate Limiting         │  │
                              │  │  5. [Future] AI Analysis  │  │
                              │  │  6. Routing Decision      │  │
                              │  │                           │  │
                              │  └───────────────────────────┘  │
                              │               │                 │
                              │               ▼                 │
                              │  ┌───────────────────────────┐  │
                              │  │      Route Table          │  │
                              │  │                           │  │
                              │  │  /api/v1/* → VM-A:8080   │  │
                              │  │  /app/*    → VM-B:3000   │  │
                              │  │  /*        → VM-C:80     │  │
                              │  └───────────────────────────┘  │
                              │               │                 │
                              └───────────────┼─────────────────┘
                                              │
                    ┌─────────────────────────┼─────────────────────────┐
                    │                         │                         │
                    ▼                         ▼                         ▼
               ┌─────────┐              ┌─────────┐              ┌─────────┐
               │  VM-A   │              │  VM-B   │              │  VM-C   │
               │ :8080   │              │ :3000   │              │ :80     │
               └─────────┘              └─────────┘              └─────────┘
```

### 7.3 Gateway Configuration

```yaml
# gateway-config.yaml
apiVersion: fcrunner.io/v1
kind: GatewayConfig
metadata:
  name: default-gateway
spec:
  listeners:
    - name: http
      port: 80
      protocol: HTTP
    
    - name: https
      port: 443
      protocol: HTTPS
      tls:
        mode: TERMINATE
        certificateRefs:
          - name: wildcard-cert
            namespace: fcrunner
  
  logging:
    enabled: true
    format: json
    fields:
      - timestamp
      - client_ip
      - method
      - path
      - query_string
      - request_headers
      - response_status
      - response_time_ms
      - upstream_vm
      - request_body_sha256  # Hash, not content
      - response_body_size
    
    # Sensitive data handling
    redact:
      headers:
        - Authorization
        - Cookie
        - X-Api-Key
      query_params:
        - token
        - key
        - secret
    
    output:
      - type: file
        path: /opt/fcrunner/logs/gateway/access.log
        rotation:
          maxSize: 100MB
          maxAge: 7d
          compress: true
      
      - type: otlp
        endpoint: http://localhost:4317
        headers:
          Authorization: "Bearer ${OTLP_TOKEN}"
  
  rateLimiting:
    enabled: true
    default:
      requestsPerSecond: 100
      burstSize: 200
    
    perRoute:
      - match:
          path: /api/v1/expensive/*
        limit:
          requestsPerSecond: 10
          burstSize: 20
  
  # Future: AI introspection
  introspection:
    enabled: false  # Phase 2
    mode: log_only  # log_only | warn | block
    models:
      - name: request-classifier
        endpoint: http://localhost:8000/classify
        timeout: 100ms
```

### 7.4 Route Configuration

```yaml
# routes.yaml
apiVersion: fcrunner.io/v1
kind: GatewayRoute
metadata:
  name: vm-routes
spec:
  routes:
    - name: api-service
      match:
        - path:
            type: PathPrefix
            value: /api/v1
        - headers:
            - name: Host
              value: api.example.com
      
      backends:
        - vmRef:
            name: vm-api-server
          port: 8080
          weight: 100
      
      # Per-route overrides
      logging:
        includeRequestBody: true
        maxBodySize: 1KB
    
    - name: web-app
      match:
        - path:
            type: PathPrefix
            value: /
      
      backends:
        - vmRef:
            name: vm-web-frontend
          port: 3000
```

### 7.5 Implementation (Go)

```go
// gateway/main.go
package main

import (
    "context"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "io"
    "net/http"
    "net/http/httputil"
    "net/url"
    "sync"
    "time"

    "go.uber.org/zap"
    "golang.org/x/time/rate"
)

type Gateway struct {
    logger      *zap.Logger
    routes      map[string]*Route
    rateLimiter *RateLimiter
    mu          sync.RWMutex
}

type Route struct {
    Name     string
    Match    RouteMatch
    Backends []Backend
    Proxy    *httputil.ReverseProxy
}

type RouteMatch struct {
    PathPrefix string
    Host       string
}

type Backend struct {
    VMID   string
    VMAddr string // IP:Port
    Weight int
}

type RequestLog struct {
    Timestamp       time.Time         `json:"timestamp"`
    RequestID       string            `json:"request_id"`
    ClientIP        string            `json:"client_ip"`
    Method          string            `json:"method"`
    Path            string            `json:"path"`
    Query           string            `json:"query"`
    Headers         map[string]string `json:"headers"`
    BodySHA256      string            `json:"body_sha256,omitempty"`
    UpstreamVM      string            `json:"upstream_vm"`
    ResponseStatus  int               `json:"response_status"`
    ResponseTimeMs  int64             `json:"response_time_ms"`
    ResponseSize    int64             `json:"response_size"`
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    startTime := time.Now()
    requestID := generateRequestID()
    
    // Find matching route
    route := g.matchRoute(r)
    if route == nil {
        http.Error(w, "No route found", http.StatusNotFound)
        return
    }
    
    // Rate limiting
    if !g.rateLimiter.Allow(r.RemoteAddr) {
        http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
        return
    }
    
    // Prepare log entry
    logEntry := &RequestLog{
        Timestamp:  startTime,
        RequestID:  requestID,
        ClientIP:   r.RemoteAddr,
        Method:     r.Method,
        Path:       r.URL.Path,
        Query:      r.URL.RawQuery,
        Headers:    sanitizeHeaders(r.Header),
        UpstreamVM: route.Backends[0].VMID,
    }
    
    // Hash request body if present
    if r.Body != nil && r.ContentLength > 0 {
        bodyBytes, _ := io.ReadAll(io.LimitReader(r.Body, 1024*1024)) // 1MB limit
        hash := sha256.Sum256(bodyBytes)
        logEntry.BodySHA256 = hex.EncodeToString(hash[:])
        r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
    }
    
    // Wrap response writer to capture status/size
    wrapped := &responseCapture{ResponseWriter: w}
    
    // Proxy request
    route.Proxy.ServeHTTP(wrapped, r)
    
    // Complete log entry
    logEntry.ResponseStatus = wrapped.status
    logEntry.ResponseSize = wrapped.size
    logEntry.ResponseTimeMs = time.Since(startTime).Milliseconds()
    
    // Emit log
    g.emitLog(logEntry)
}

func (g *Gateway) matchRoute(r *http.Request) *Route {
    g.mu.RLock()
    defer g.mu.RUnlock()
    
    for _, route := range g.routes {
        if strings.HasPrefix(r.URL.Path, route.Match.PathPrefix) {
            if route.Match.Host == "" || route.Match.Host == r.Host {
                return route
            }
        }
    }
    return nil
}

func sanitizeHeaders(headers http.Header) map[string]string {
    result := make(map[string]string)
    sensitiveHeaders := map[string]bool{
        "Authorization": true,
        "Cookie":        true,
        "X-Api-Key":     true,
    }
    
    for key, values := range headers {
        if sensitiveHeaders[key] {
            result[key] = "[REDACTED]"
        } else {
            result[key] = strings.Join(values, ", ")
        }
    }
    return result
}

type responseCapture struct {
    http.ResponseWriter
    status int
    size   int64
}

func (r *responseCapture) WriteHeader(status int) {
    r.status = status
    r.ResponseWriter.WriteHeader(status)
}

func (r *responseCapture) Write(b []byte) (int, error) {
    n, err := r.ResponseWriter.Write(b)
    r.size += int64(n)
    return n, err
}

// RateLimiter per-client rate limiting
type RateLimiter struct {
    limiters sync.Map
    rate     rate.Limit
    burst    int
}

func (rl *RateLimiter) Allow(clientIP string) bool {
    limiter, _ := rl.limiters.LoadOrStore(clientIP, rate.NewLimiter(rl.rate, rl.burst))
    return limiter.(*rate.Limiter).Allow()
}
```

### 7.6 Future: AI Introspection

```yaml
# Phase 2: AI-powered request analysis
introspection:
  enabled: true
  mode: log_only  # Start with logging only
  
  analyzers:
    - name: prompt-injection-detector
      description: Detect prompt injection attempts in request bodies
      model:
        type: local  # Run model locally
        path: /opt/fcrunner/models/prompt-injection-v1
        maxLatency: 50ms
      
      actions:
        high_confidence:
          - log
          - tag: "security:prompt-injection"
        low_confidence:
          - log
    
    - name: anomaly-detector
      description: Detect unusual request patterns
      model:
        type: statistical
        features:
          - request_rate
          - path_entropy
          - payload_size
      
      actions:
        anomaly:
          - log
          - alert
    
    - name: content-classifier
      description: Classify request content type/intent
      model:
        type: api
        endpoint: http://localhost:8000/classify
      
      actions:
        classified:
          - tag_with_category
```

---

## 8. Management API

### 8.1 API Design Principles

- **gRPC-first**: Binary protocol, strong typing, streaming support
- **mTLS required**: All connections mutually authenticated
- **RBAC**: Role-based access control for operations
- **Audit logging**: All operations logged with caller identity

### 8.2 Service Definition

```protobuf
// api/v1/fcrunner.proto
syntax = "proto3";
package fcrunner.api.v1;

import "google/protobuf/timestamp.proto";
import "google/protobuf/duration.proto";
import "google/protobuf/empty.proto";

option go_package = "github.com/yourorg/fcrunner/api/v1;apiv1";

// VM Management Service
service VMService {
  // Lifecycle
  rpc CreateVM(CreateVMRequest) returns (VM);
  rpc GetVM(GetVMRequest) returns (VM);
  rpc ListVMs(ListVMsRequest) returns (ListVMsResponse);
  rpc DeleteVM(DeleteVMRequest) returns (google.protobuf.Empty);
  
  // State management
  rpc StartVM(StartVMRequest) returns (VM);
  rpc StopVM(StopVMRequest) returns (VM);
  rpc RestartVM(RestartVMRequest) returns (VM);
  
  // Container operations (via guest agent)
  rpc RunContainer(RunContainerRequest) returns (Container);
  rpc StopContainer(StopContainerRequest) returns (google.protobuf.Empty);
  rpc GetContainerLogs(GetContainerLogsRequest) returns (stream LogEntry);
  rpc ExecInContainer(stream ExecRequest) returns (stream ExecResponse);
  
  // Observability
  rpc GetVMMetrics(GetVMMetricsRequest) returns (VMMetrics);
  rpc StreamVMEvents(StreamVMEventsRequest) returns (stream VMEvent);
}

// Network Policy Service
service NetworkPolicyService {
  rpc CreatePolicy(CreatePolicyRequest) returns (NetworkPolicy);
  rpc GetPolicy(GetPolicyRequest) returns (NetworkPolicy);
  rpc ListPolicies(ListPoliciesRequest) returns (ListPoliciesResponse);
  rpc UpdatePolicy(UpdatePolicyRequest) returns (NetworkPolicy);
  rpc DeletePolicy(DeletePolicyRequest) returns (google.protobuf.Empty);
  rpc ApplyPolicy(ApplyPolicyRequest) returns (google.protobuf.Empty);
}

// Health Service
service HealthService {
  rpc Check(HealthCheckRequest) returns (HealthCheckResponse);
  rpc Watch(HealthCheckRequest) returns (stream HealthCheckResponse);
}

// === Messages ===

message VM {
  string id = 1;
  string name = 2;
  VMState state = 3;
  VMSpec spec = 4;
  VMStatus status = 5;
  map<string, string> labels = 6;
  google.protobuf.Timestamp created_at = 7;
  google.protobuf.Timestamp updated_at = 8;
}

enum VMState {
  VM_STATE_UNSPECIFIED = 0;
  VM_STATE_PENDING = 1;
  VM_STATE_STARTING = 2;
  VM_STATE_RUNNING = 3;
  VM_STATE_STOPPING = 4;
  VM_STATE_STOPPED = 5;
  VM_STATE_FAILED = 6;
}

message VMSpec {
  // Resources
  int32 vcpus = 1;
  int64 memory_mib = 2;
  int64 disk_size_mib = 3;
  
  // Images
  string kernel_image = 4;
  string rootfs_image = 5;
  
  // Network
  NetworkSpec network = 6;
  
  // Timeouts
  google.protobuf.Duration boot_timeout = 7;
  google.protobuf.Duration idle_timeout = 8;
  google.protobuf.Duration max_lifetime = 9;
  
  // Labels for policy matching
  map<string, string> labels = 10;
}

message NetworkSpec {
  // Static IP assignment (or auto if empty)
  string ip_address = 1;
  
  // Exposed ports
  repeated PortMapping ports = 2;
  
  // Network policy reference
  string network_policy = 3;
  
  // Bandwidth limits
  int64 bandwidth_mbps = 4;
}

message PortMapping {
  int32 host_port = 1;
  int32 vm_port = 2;
  string protocol = 3;  // tcp, udp
}

message VMStatus {
  string ip_address = 1;
  string mac_address = 2;
  repeated ContainerStatus containers = 3;
  ResourceUsage resource_usage = 4;
  string last_error = 5;
}

message ResourceUsage {
  double cpu_percent = 1;
  int64 memory_bytes = 2;
  int64 disk_read_bytes = 3;
  int64 disk_write_bytes = 4;
  int64 network_rx_bytes = 5;
  int64 network_tx_bytes = 6;
}

message CreateVMRequest {
  string name = 1;
  VMSpec spec = 2;
  map<string, string> labels = 3;
  
  // Optional: start immediately after creation
  bool auto_start = 4;
}

message GetVMRequest {
  string id = 1;
}

message ListVMsRequest {
  // Filtering
  map<string, string> label_selector = 1;
  repeated VMState state_filter = 2;
  
  // Pagination
  int32 page_size = 3;
  string page_token = 4;
}

message ListVMsResponse {
  repeated VM vms = 1;
  string next_page_token = 2;
  int32 total_count = 3;
}

message DeleteVMRequest {
  string id = 1;
  bool force = 2;  // Force kill if running
}

message StartVMRequest {
  string id = 1;
  google.protobuf.Duration timeout = 2;
}

message StopVMRequest {
  string id = 1;
  google.protobuf.Duration timeout = 2;
  bool force = 3;  // SIGKILL vs graceful
}

message RestartVMRequest {
  string id = 1;
  google.protobuf.Duration timeout = 2;
}

// Container operations
message Container {
  string id = 1;
  string vm_id = 2;
  string image = 3;
  ContainerState state = 4;
  google.protobuf.Timestamp created_at = 5;
}

enum ContainerState {
  CONTAINER_STATE_UNSPECIFIED = 0;
  CONTAINER_STATE_CREATING = 1;
  CONTAINER_STATE_RUNNING = 2;
  CONTAINER_STATE_STOPPED = 3;
  CONTAINER_STATE_FAILED = 4;
}

message RunContainerRequest {
  string vm_id = 1;
  string image = 2;
  repeated string command = 3;
  map<string, string> env = 4;
  repeated Mount mounts = 5;
  ContainerResources resources = 6;
}

message Mount {
  string source = 1;  // Path in VM
  string target = 2;  // Path in container
  bool read_only = 3;
}

message ContainerResources {
  int64 memory_bytes = 1;
  int64 cpu_shares = 2;
  int64 pids_limit = 3;
}

message StopContainerRequest {
  string vm_id = 1;
  string container_id = 2;
  google.protobuf.Duration timeout = 3;
}

message GetContainerLogsRequest {
  string vm_id = 1;
  string container_id = 2;
  bool follow = 3;
  int32 tail_lines = 4;
  bool timestamps = 5;
}

message LogEntry {
  google.protobuf.Timestamp timestamp = 1;
  string stream = 2;  // stdout, stderr
  bytes content = 3;
}

message ExecRequest {
  oneof message {
    ExecStart start = 1;
    bytes stdin = 2;
    ExecResize resize = 3;
  }
}

message ExecStart {
  string vm_id = 1;
  string container_id = 2;
  repeated string command = 3;
  bool tty = 4;
}

message ExecResize {
  uint32 width = 1;
  uint32 height = 2;
}

message ExecResponse {
  oneof output {
    bytes stdout = 1;
    bytes stderr = 2;
    int32 exit_code = 3;
  }
}

// Network Policy
message NetworkPolicy {
  string id = 1;
  string name = 2;
  NetworkPolicySpec spec = 3;
  google.protobuf.Timestamp created_at = 4;
}

message NetworkPolicySpec {
  map<string, string> vm_selector = 1;
  string default_action = 2;  // allow, drop
  repeated IngressRule ingress = 3;
  repeated EgressRule egress = 4;
}

message IngressRule {
  string name = 1;
  repeated string from_cidrs = 2;
  repeated PortRule ports = 3;
  string action = 4;  // allow, drop
}

message EgressRule {
  string name = 1;
  repeated string to_cidrs = 2;
  repeated PortRule ports = 3;
  string action = 4;
}

message PortRule {
  int32 port = 1;
  string protocol = 2;
}

// Events
message VMEvent {
  google.protobuf.Timestamp timestamp = 1;
  string vm_id = 2;
  string type = 3;  // created, started, stopped, failed, etc.
  string message = 4;
  map<string, string> metadata = 5;
}

message StreamVMEventsRequest {
  // Filter by VM IDs (empty = all)
  repeated string vm_ids = 1;
  // Filter by event types (empty = all)
  repeated string event_types = 2;
}

// Metrics
message GetVMMetricsRequest {
  string vm_id = 1;
}

message VMMetrics {
  string vm_id = 1;
  google.protobuf.Timestamp timestamp = 2;
  ResourceUsage usage = 3;
  repeated ContainerMetrics containers = 4;
}

message ContainerMetrics {
  string container_id = 1;
  ResourceUsage usage = 2;
}

// Health
message HealthCheckRequest {
  string service = 1;  // Empty = overall health
}

message HealthCheckResponse {
  enum ServingStatus {
    UNKNOWN = 0;
    SERVING = 1;
    NOT_SERVING = 2;
  }
  ServingStatus status = 1;
  map<string, string> details = 2;
}
```

### 8.3 Authentication & Authorization

#### 8.3.1 mTLS Configuration

```yaml
# /opt/fcrunner/etc/tls.yaml
tls:
  server:
    cert: /opt/fcrunner/etc/certs/server.crt
    key: /opt/fcrunner/etc/certs/server.key
    ca: /opt/fcrunner/etc/certs/ca.crt
    
    # Require client certificates
    clientAuth: requireAndVerify
    
    # Minimum TLS version
    minVersion: "1.3"
  
  client:
    # Client certificates must include these SANs or CNs
    allowedSubjects:
      - CN=fcrunner-admin
      - CN=fcrunner-operator
      - CN=cicd-service
```

#### 8.3.2 RBAC Configuration

```yaml
# /opt/fcrunner/etc/rbac.yaml
roles:
  - name: admin
    description: Full access to all operations
    permissions:
      - resource: "*"
        actions: ["*"]
  
  - name: operator
    description: Manage VMs and containers
    permissions:
      - resource: vm
        actions: [create, get, list, start, stop, delete]
      - resource: container
        actions: [run, stop, logs, exec]
      - resource: metrics
        actions: [get]
  
  - name: viewer
    description: Read-only access
    permissions:
      - resource: vm
        actions: [get, list]
      - resource: container
        actions: [logs]
      - resource: metrics
        actions: [get]
  
  - name: network-admin
    description: Manage network policies
    permissions:
      - resource: networkpolicy
        actions: [create, get, list, update, delete, apply]

bindings:
  - subject: CN=fcrunner-admin
    role: admin
  
  - subject: CN=fcrunner-operator
    role: operator
  
  - subject: CN=cicd-service
    role: operator
  
  - subject: CN=monitoring-service
    role: viewer
```

### 8.4 Client SDK Example

```go
// client/example.go
package main

import (
    "context"
    "crypto/tls"
    "crypto/x509"
    "fmt"
    "io/ioutil"
    "log"
    "time"

    apiv1 "github.com/yourorg/fcrunner/api/v1"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials"
)

func main() {
    // Load TLS credentials
    cert, err := tls.LoadX509KeyPair("client.crt", "client.key")
    if err != nil {
        log.Fatal(err)
    }
    
    caCert, err := ioutil.ReadFile("ca.crt")
    if err != nil {
        log.Fatal(err)
    }
    
    caCertPool := x509.NewCertPool()
    caCertPool.AppendCertsFromPEM(caCert)
    
    tlsConfig := &tls.Config{
        Certificates: []tls.Certificate{cert},
        RootCAs:      caCertPool,
        MinVersion:   tls.VersionTLS13,
    }
    
    // Connect
    conn, err := grpc.Dial(
        "fcrunner.example.com:443",
        grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()
    
    client := apiv1.NewVMServiceClient(conn)
    ctx := context.Background()
    
    // Create a VM
    vm, err := client.CreateVM(ctx, &apiv1.CreateVMRequest{
        Name: "my-sandbox",
        Spec: &apiv1.VMSpec{
            Vcpus:     2,
            MemoryMib: 1024,
            Network: &apiv1.NetworkSpec{
                Ports: []*apiv1.PortMapping{
                    {HostPort: 8080, VmPort: 80, Protocol: "tcp"},
                },
                NetworkPolicy: "default-deny",
                BandwidthMbps: 100,
            },
        },
        Labels: map[string]string{
            "tenant": "untrusted",
            "env":    "sandbox",
        },
        AutoStart: true,
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Created VM: %s (IP: %s)\n", vm.Id, vm.Status.IpAddress)
    
    // Run a container
    container, err := client.RunContainer(ctx, &apiv1.RunContainerRequest{
        VmId:    vm.Id,
        Image:   "nginx:alpine",
        Command: []string{},
        Env:     map[string]string{"NGINX_PORT": "80"},
        Resources: &apiv1.ContainerResources{
            MemoryBytes: 256 * 1024 * 1024,
            CpuShares:   512,
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Started container: %s\n", container.Id)
    
    // Stream logs
    stream, err := client.GetContainerLogs(ctx, &apiv1.GetContainerLogsRequest{
        VmId:        vm.Id,
        ContainerId: container.Id,
        Follow:      true,
        TailLines:   100,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    for {
        entry, err := stream.Recv()
        if err != nil {
            break
        }
        fmt.Printf("[%s] %s: %s\n", 
            entry.Timestamp.AsTime().Format(time.RFC3339),
            entry.Stream,
            string(entry.Content),
        )
    }
}
```

---

## 9. Security Model

### 9.1 Threat Model

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           THREAT ACTORS                                 │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  EXTERNAL ATTACKER              MALICIOUS WORKLOAD       INSIDER THREAT │
│  - Network-based attacks        - VM escape attempts     - Misuse of    │
│  - API exploitation             - Resource exhaustion      management   │
│  - DDoS                         - Data exfiltration        API          │
│                                 - Lateral movement       - Credential   │
│                                                            theft        │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                           ATTACK SURFACE                                │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐   │
│  │ Management  │  │  Network    │  │   KVM/      │  │  virtio     │   │
│  │    API      │  │  (L3/L4/L7) │  │  Hypervisor │  │  Devices    │   │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘   │
│                                                                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐   │
│  │  Syscalls   │  │  Block      │  │   vsock     │  │  Shared     │   │
│  │  (seccomp)  │  │  Devices    │  │  Interface  │  │  Memory     │   │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘   │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### 9.2 Security Controls Matrix

| Threat | Control | Implementation | Priority |
|--------|---------|----------------|----------|
| VM escape via KVM bug | Kernel updates, minimal kernel | Host hardening | P0 |
| VM escape via device emulation | Minimal virtio devices, Rust VMM | Firecracker design | P0 |
| Syscall exploitation | Seccomp-BPF, minimal allowlist | Jailer | P0 |
| Network-based attacks | nftables L3/L4, API gateway L7 | Network policy | P0 |
| Resource exhaustion | cgroups v2 hard limits | VM Manager | P0 |
| Lateral movement | Network isolation, default-deny | Network policy | P0 |
| API exploitation | mTLS, RBAC, input validation | gRPC server | P0 |
| Data exfiltration | Egress filtering, logging | Network policy + gateway | P1 |
| Side-channel attacks | Core pinning, no SMT, kernel mitigations | Host config | P1 |
| Credential theft | Short-lived certs, rotation | PKI | P1 |
| Insider threat | Audit logging, least privilege | RBAC + logging | P1 |

### 9.3 Security Invariants

```
INVARIANT 1: Host Kernel Isolation
  - Firecracker process MUST run with:
    - No capabilities
    - Strict seccomp filter
    - Unprivileged UID
    - Namespaced (PID, NET, MNT, USER)
    - Chrooted
  
  VERIFICATION:
  - Audit jailer configuration
  - Tetragon monitoring for violations
  - Periodic capability checks

INVARIANT 2: Network Boundary Enforcement
  - All VM network traffic MUST traverse nftables
  - Default policy MUST be deny
  - No direct host network access
  - No VM-to-VM communication without explicit policy
  
  VERIFICATION:
  - Network policy compilation tests
  - Traffic capture validation
  - Policy audit logging

INVARIANT 3: Resource Containment
  - VMs MUST NOT exceed allocated resources
  - No memory overcommit
  - CPU limits enforced via cgroups
  - No shared resources between VMs
  
  VERIFICATION:
  - cgroup limit tests
  - Resource exhaustion fuzzing
  - Monitoring alerts

INVARIANT 4: API Authentication
  - All API calls MUST use mTLS
  - All operations MUST pass RBAC checks
  - All operations MUST be audit logged
  
  VERIFICATION:
  - TLS configuration tests
  - RBAC policy tests
  - Audit log completeness checks
```

### 9.4 Incident Response

```yaml
# Incident response procedures
incidents:
  - type: vm_escape_attempt
    detection:
      - Tetragon alert: unexpected syscall
      - Tetragon alert: file access outside allowed paths
      - Tetragon alert: network connection from Firecracker
    
    response:
      immediate:
        - Kill affected VM process (automatic via Tetragon)
        - Block VM network (nftables)
        - Preserve VM state for forensics
      
      investigation:
        - Collect Tetragon logs
        - Collect VM console output
        - Analyze guest memory (if preserved)
        - Check for similar activity in other VMs
      
      remediation:
        - Patch if kernel vulnerability
        - Update seccomp profile if needed
        - Review and tighten policies
  
  - type: resource_exhaustion
    detection:
      - cgroup OOM events
      - CPU throttling sustained
      - Disk quota exceeded
    
    response:
      immediate:
        - VM continues within limits (by design)
        - Alert operator
      
      investigation:
        - Review VM workload
        - Check for cryptomining/abuse
      
      remediation:
        - Terminate if malicious
        - Adjust limits if legitimate
  
  - type: network_policy_violation
    detection:
      - nftables drop with logging
      - Gateway blocked request
    
    response:
      immediate:
        - Traffic blocked (by design)
        - Log event
      
      investigation:
        - Analyze traffic patterns
        - Check for data exfiltration attempts
      
      remediation:
        - Adjust policy if false positive
        - Terminate VM if malicious
```

---

## 10. Platform Support

### 10.1 Linux (Primary Platform)

```yaml
platform: linux
status: primary
kvm_source: native

requirements:
  kernel: "5.10+"
  modules:
    - kvm
    - kvm_intel | kvm_amd
    - vhost_net
    - tun
  
distributions:
  tier1:  # Full support, CI tested
    - ubuntu-22.04
    - ubuntu-24.04
    - debian-12
  
  tier2:  # Supported, community tested
    - fedora-39+
    - amazon-linux-2023
    - rhel-9
```

### 10.2 macOS (Secondary Platform)

```yaml
platform: macos
status: secondary
kvm_source: hypervisor.framework  # Not KVM, but similar API

approach:
  # Option A: Use native Hypervisor.framework
  native:
    description: Port Firecracker to use macOS Hypervisor.framework
    pros:
      - Native performance
      - No additional software
    cons:
      - Significant development effort
      - Different API, requires abstraction layer
    status: future
  
  # Option B: Linux VM as host
  nested:
    description: Run Linux in a VM, then Firecracker inside
    pros:
      - Immediate compatibility
      - Reuse all Linux tooling
    cons:
      - Performance overhead
      - More complex setup
    implementation:
      hypervisor: colima | orbstack | lima
      inner_vm: ubuntu-22.04
    status: recommended_interim

requirements:
  macos_version: "13.0+"  # Ventura+
  architecture: arm64 | x86_64
  
  option_b_setup: |
    # Using Colima (recommended for Docker-like experience)
    brew install colima docker
    
    # Start with nested virtualization enabled
    colima start --cpu 4 --memory 8 --vm-type vz --vz-rosetta
    
    # Inside Colima VM, install fcrunner
    colima ssh
    curl -fsSL https://get.fcrunner.io | sudo bash
```

### 10.3 Windows (Tertiary Platform)

```yaml
platform: windows
status: tertiary
kvm_source: wsl2  # Uses Hyper-V backed WSL2

approach:
  wsl2:
    description: Run fcrunner inside WSL2 distribution
    pros:
      - Good integration with Windows
      - Reasonable performance
    cons:
      - WSL2 kernel may lag behind
      - Some networking complexity
    implementation:
      distribution: ubuntu-22.04
      kernel: custom (with KVM enabled)
    
requirements:
  windows_version: "11+"  # Or Windows 10 21H2+
  wsl_version: "2"
  features:
    - Hyper-V
    - Virtual Machine Platform
    - Windows Subsystem for Linux

setup_script: |
  # PowerShell (Administrator)
  
  # Enable required features
  dism.exe /online /enable-feature /featurename:Microsoft-Windows-Subsystem-Linux /all /norestart
  dism.exe /online /enable-feature /featurename:VirtualMachinePlatform /all /norestart
  dism.exe /online /enable-feature /featurename:Microsoft-Hyper-V-All /all /norestart
  
  # Reboot required
  Restart-Computer
  
  # After reboot, set WSL2 as default
  wsl --set-default-version 2
  
  # Install Ubuntu
  wsl --install -d Ubuntu-22.04
  
  # Configure nested virtualization (requires custom kernel)
  # See: https://github.com/microsoft/WSL2-Linux-Kernel

wsl2_kernel_config: |
  # Additional options for WSL2 kernel to enable KVM
  # Build custom kernel with:
  CONFIG_KVM=y
  CONFIG_KVM_INTEL=y  # or CONFIG_KVM_AMD=y
  CONFIG_VHOST_NET=y

networking:
  notes: |
    WSL2 networking requires additional configuration for
    exposing VM ports to Windows host. Use:
    - netsh interface portproxy
    - Or: WSL2 mirrored networking mode (Windows 11 23H2+)
```

### 10.4 Platform Abstraction Layer

```go
// platform/platform.go
package platform

import "runtime"

type Platform interface {
    // VM Operations
    CreateVM(spec VMSpec) (VM, error)
    StartVM(id string) error
    StopVM(id string) error
    
    // Network Operations
    CreateNetwork(spec NetworkSpec) (Network, error)
    ApplyFirewall(vmID string, rules []FirewallRule) error
    
    // Resource Management
    CreateCgroup(name string, limits ResourceLimits) error
    
    // Platform Info
    Info() PlatformInfo
}

type PlatformInfo struct {
    OS           string
    Arch         string
    KVMAvailable bool
    KVMType      string // "kvm", "hypervisor.framework", "hyper-v"
}

func NewPlatform() (Platform, error) {
    switch runtime.GOOS {
    case "linux":
        return NewLinuxPlatform()
    case "darwin":
        return NewMacOSPlatform()
    case "windows":
        return NewWindowsPlatform()
    default:
        return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
    }
}

// Linux implementation uses Firecracker directly
type LinuxPlatform struct {
    firecrackerBin string
    jailerBin      string
}

// macOS implementation wraps a Linux VM
type MacOSPlatform struct {
    innerVM     string // colima, lima, etc.
    innerClient *grpc.ClientConn
}

// Windows implementation uses WSL2
type WindowsPlatform struct {
    wslDistro string
    wslClient *grpc.ClientConn
}
```

---

## 11. Implementation Phases

### Phase 1: Core VM Runtime

```yaml
phase: 1
name: Core VM Runtime
goal: Basic VM lifecycle with Docker support

deliverables:
  batch_1:
    - Project scaffolding
    - Firecracker wrapper library (Go)
    - Jailer integration
    - Basic VM create/start/stop/delete
  
  batch_2:
    - Guest kernel build automation
    - Root filesystem with containerd
    - Guest agent (vsock communication)
    - Container run/stop via guest agent
  
  batch_3:
    - gRPC API server (basic endpoints)
    - mTLS authentication
    - SQLite state persistence
    - Basic CLI tool
  
  batch_4:
    - cgroup v2 resource limits
    - Basic logging (file-based)
    - Integration tests
    - Documentation

success_criteria:
  - Can create VM in < 500ms
  - Can run Docker container in VM
  - Can stop/delete VM cleanly
  - Resource limits enforced
  - mTLS authentication working
```

### Phase 2: Network Security

```yaml
phase: 2
name: Network Security
goal: L3/L4 network policy with basic L7 gateway
depends_on: [phase_1.batch_1, phase_1.batch_2]

deliverables:
  batch_1:
    - Network policy schema design
    - nftables rule compiler
    - Per-VM network namespace setup
  
  batch_2:
    - Network policy API endpoints
    - Dynamic policy updates (hot reload)
    - Policy validation
  
  batch_3:
    - API gateway (basic reverse proxy)
    - Request/response logging
    - Rate limiting
  
  batch_4:
    - Gateway route configuration
    - TLS termination
    - Integration tests

success_criteria:
  - Network policies block/allow as configured
  - No VM-to-VM traffic without explicit policy
  - All HTTP traffic logged
  - Rate limiting functional
```

### Phase 3: Observability & Security Hardening

```yaml
phase: 3
name: Observability & Security
goal: Production-grade monitoring and security
depends_on: [phase_1, phase_2]

deliverables:
  batch_1:
    - Prometheus metrics exporter
    - Tetragon policy deployment
    - Structured logging (JSON)
  
  batch_2:
    - OTLP export (traces, logs)
    - RBAC implementation
    - Audit logging
  
  batch_3:
    - Security hardening review
    - Penetration testing
    - Threat model validation

success_criteria:
  - All operations have metrics
  - Tetragon detects escape attempts
  - RBAC enforced on all endpoints
  - Audit log for all operations
```

### Phase 4: Multi-Platform & Production Features

```yaml
phase: 4
name: Multi-Platform & Production
goal: macOS/Windows support, HA, production readiness
depends_on: [phase_1, phase_2, phase_3]

deliverables:
  batch_1:
    - Platform abstraction layer
    - macOS support (via nested VM)
  
  batch_2:
    - Windows/WSL2 support
    - High availability (multi-node)
  
  batch_3:
    - VM migration (optional)
    - Backup/restore
    - Performance optimization
  
  batch_4:
    - Production documentation
    - Deployment guides

success_criteria:
  - Works on Linux, macOS, Windows
  - Can survive single node failure
  - Production deployment guide complete
```

### Phase 5: AI Introspection

```yaml
phase: 5
name: AI Introspection
goal: AI-powered request analysis
depends_on: [phase_2.batch_3]  # Needs gateway

deliverables:
  - Prompt injection detection
  - Anomaly detection
  - Content classification
  - Adaptive rate limiting
  - Automated threat response

status: future
```

---

## 12. Operational Requirements

### 12.1 Deployment Architecture

```yaml
deployment:
  single_node:
    description: All components on one host
    use_case: Development, small deployments
    components:
      - fcrunner-api
      - fcrunner-gateway
      - fcrunner-vms (Firecracker processes)
      - prometheus (optional)
      - tetragon (optional)
  
  multi_node:
    description: Distributed deployment
    use_case: Production, high availability
    components:
      control_plane:
        - fcrunner-api (replicated)
        - etcd (state store)
        - prometheus
      
      data_plane:
        - fcrunner-worker (per host)
        - fcrunner-gateway (per host or centralized)
        - tetragon (per host)
```

### 12.2 Configuration Management

```yaml
# /opt/fcrunner/etc/fcrunner.yaml
apiVersion: fcrunner.io/v1
kind: Config
metadata:
  name: fcrunner-config

spec:
  # API Server
  api:
    listen: "0.0.0.0:8443"
    tls:
      certFile: /opt/fcrunner/etc/certs/server.crt
      keyFile: /opt/fcrunner/etc/certs/server.key
      caFile: /opt/fcrunner/etc/certs/ca.crt
  
  # VM Defaults
  vm:
    defaults:
      vcpus: 2
      memoryMib: 1024
      bootTimeout: 30s
      idleTimeout: 1h
      maxLifetime: 24h
    
    limits:
      maxVMs: 100
      maxVCPUsPerVM: 8
      maxMemoryPerVM: 16384  # MiB
    
    images:
      kernelPath: /opt/fcrunner/lib/kernels/vmlinux-5.10
      rootfsPath: /opt/fcrunner/lib/rootfs/docker-rootfs.ext4
  
  # Network
  network:
    bridge: fcbr0
    subnet: 172.16.0.0/24
    gateway: 172.16.0.1
    dns: 172.16.0.1
    defaultPolicy: /opt/fcrunner/etc/network-policies/default.yaml
  
  # Gateway
  gateway:
    enabled: true
    listen: "0.0.0.0:80"
    tlsListen: "0.0.0.0:443"
    configPath: /opt/fcrunner/etc/gateway.yaml
  
  # Storage
  storage:
    stateDB: /opt/fcrunner/data/state.db
    imagesDir: /opt/fcrunner/data/images
    overlaysDir: /opt/fcrunner/data/overlays
    logsDir: /opt/fcrunner/logs
  
  # Observability
  observability:
    metrics:
      enabled: true
      listen: "127.0.0.1:9090"
    
    logging:
      level: info
      format: json
      output: /opt/fcrunner/logs/fcrunner.log
    
    tracing:
      enabled: false
      endpoint: http://localhost:4317
  
  # Security
  security:
    seccompProfile: /opt/fcrunner/etc/seccomp/firecracker.json
    tetragon:
      enabled: true
      policyDir: /opt/fcrunner/etc/tetragon
```

### 12.3 Monitoring & Alerting

```yaml
# prometheus/alerts.yaml
groups:
  - name: fcrunner
    rules:
      - alert: VMEscapeAttempt
        expr: increase(fcrunner_security_events_total{event_type="escape_attempt"}[5m]) > 0
        for: 0m
        labels:
          severity: critical
        annotations:
          summary: "VM escape attempt detected"
          description: "Tetragon detected suspicious activity from VM {{ $labels.vm_id }}"
      
      - alert: HighVMFailureRate
        expr: rate(fcrunner_vms_total{status="failed"}[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High VM failure rate"
          description: "More than 10% of VMs are failing"
      
      - alert: ResourceExhaustion
        expr: fcrunner_vm_memory_bytes / fcrunner_vm_memory_limit_bytes > 0.9
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "VM near memory limit"
          description: "VM {{ $labels.vm_id }} is using >90% of memory limit"
      
      - alert: APIHighLatency
        expr: histogram_quantile(0.99, rate(fcrunner_api_request_duration_seconds_bucket[5m])) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "API latency high"
          description: "99th percentile API latency is above 1 second"
```

### 12.4 Backup & Recovery

```yaml
backup:
  components:
    state_db:
      description: SQLite state database
      method: sqlite3 .backup
      frequency: hourly
      retention: 7d
    
    configuration:
      description: All YAML configs
      method: file copy
      frequency: on_change
      retention: 30d
    
    certificates:
      description: TLS certificates
      method: file copy
      frequency: on_change
      retention: forever
      notes: Store securely, consider HSM for production
    
    vm_images:
      description: Kernel and rootfs images
      method: file copy
      frequency: on_change
      retention: forever
      notes: Keep multiple versions for rollback

recovery:
  procedures:
    full_recovery:
      steps:
        - Restore configuration files
        - Restore state database
        - Restore certificates
        - Restart fcrunner services
        - Verify VM list matches expected
        - Restart VMs as needed
    
    vm_recovery:
      steps:
        - Check VM state in database
        - If VM was running, attempt restart
        - If restart fails, check logs
        - Recreate VM if necessary
```

---

## Appendix A: Glossary

| Term | Definition |
|------|------------|
| **Firecracker** | Lightweight VMM (Virtual Machine Monitor) built by AWS |
| **Jailer** | Firecracker companion that sets up isolation (namespaces, seccomp, chroot) |
| **KVM** | Kernel-based Virtual Machine - Linux kernel virtualization |
| **virtio** | Standard for virtual device emulation |
| **vsock** | Virtual socket for host-guest communication |
| **nftables** | Linux kernel packet filtering framework |
| **seccomp** | Secure Computing - syscall filtering |
| **cgroups** | Control Groups - resource limiting |
| **mTLS** | Mutual TLS - both client and server authenticate |

---

## Appendix B: References

- [Firecracker Design](https://github.com/firecracker-microvm/firecracker/blob/main/docs/design.md)
- [Firecracker Security](https://github.com/firecracker-microvm/firecracker/blob/main/docs/security.md)
- [Jailer Documentation](https://github.com/firecracker-microvm/firecracker/blob/main/docs/jailer.md)
- [KVM API](https://www.kernel.org/doc/Documentation/virtual/kvm/api.txt)
- [nftables Wiki](https://wiki.nftables.org/)
- [Tetragon Documentation](https://tetragon.io/docs/)
- [gRPC Authentication](https://grpc.io/docs/guides/auth/)

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 0.1.0 | 2024-12-20 | - | Initial draft |
| 0.1.1 | 2025-12-21 | - | Terminology standardization (microVM), cross-references |
