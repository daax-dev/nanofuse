# Nanofuse Frequently Asked Questions

## General Questions

### What is Nanofuse?

Nanofuse is a microVM platform for running untrusted code in secure, isolated sandboxes. It uses [Firecracker](https://github.com/firecracker-microvm/firecracker) to provide hardware-level isolation with sub-second boot times.

### How is Nanofuse different from containers?

Containers share the host kernel, which creates a larger attack surface for untrusted code. Nanofuse runs each workload in its own microVM with a dedicated kernel, providing stronger isolation similar to traditional VMs but with container-like performance.

| Feature | Containers | Nanofuse microVMs |
|---------|------------|-------------------|
| Kernel isolation | Shared kernel | Dedicated kernel |
| Boot time | Milliseconds | Sub-second |
| Memory overhead | Low | ~5-10 MiB per VM |
| Security | Process isolation | Hardware isolation |
| Use case | Trusted workloads | Untrusted workloads |

### How is Nanofuse different from E2B?

Nanofuse is heavily inspired by [E2B](https://e2b.dev) but is designed to be self-hosted. Key differences:

- **Self-hosted**: Run on your own infrastructure
- **Customizable**: Built on open source technologies
- **No vendor lock-in**: No cloud dependency
- **Control**: Full control over security policies and data

### What are the primary use cases?

1. **AI Code Execution**: Run LLM-generated code securely
2. **Isolated Workloads**: Multi-tenant compute environments
3. **Development Sandboxes**: Fast-spinning dev environments

---

## Requirements and Compatibility

### What are the system requirements?

| Requirement | Details |
|-------------|---------|
| Operating System | Linux with KVM support |
| Architecture | x86_64 (ARM64 support planned) |
| Kernel | 4.14+ (5.10+ recommended) |
| Permissions | Root or CAP_NET_ADMIN capability |

### Does Nanofuse work on macOS or Windows?

Not directly. Nanofuse requires Linux with KVM support. For macOS or Windows:

- Use a Linux VM (e.g., via Multipass, Vagrant, or WSL2)
- Deploy to a Linux cloud instance
- Use Docker Desktop with a Linux container

### Does Nanofuse work in cloud environments?

Yes, Nanofuse works on cloud instances that support nested virtualization:

| Provider | Instance Types |
|----------|---------------|
| AWS | Metal instances, `.metal` suffix |
| GCP | N1, N2 with nested virtualization enabled |
| Azure | Dv3, Ev3 with nested virtualization |

### Can I run Nanofuse inside Docker?

Not recommended. Nanofuse requires direct access to `/dev/kvm` and network capabilities that are difficult to provide securely within a container.

---

## Installation and Setup

### How do I install Nanofuse?

Download pre-built binaries or build from source:

```bash
# Download binaries
VERSION=v0.1.0
curl -LO https://github.com/daax-dev/nanofuse/releases/download/${VERSION}/nanofuse
curl -LO https://github.com/daax-dev/nanofuse/releases/download/${VERSION}/nanofused
chmod +x nanofuse nanofused
sudo mv nanofuse nanofused /usr/local/bin/
```

See the [Quick Start Guide](QUICKSTART.md) for complete instructions.

### Why does Nanofuse require root access?

Nanofuse needs root access for:

1. **KVM access**: Creating and managing virtual machines
2. **Network bridges**: Setting up TAP devices and bridges
3. **IP routing**: Configuring NAT for VM internet access

You can run with reduced privileges using capabilities, but root is recommended for development.

### How do I authenticate with GHCR?

Create a GitHub personal access token with `read:packages` scope:

1. Go to [GitHub Token Settings](https://github.com/settings/tokens/new?scopes=read:packages)
2. Generate a token
3. Login:

```bash
echo $GITHUB_TOKEN | docker login ghcr.io -u YOUR_USERNAME --password-stdin
```

---

## microVMs and Images

### What base OS do microVMs use?

The default base image is Ubuntu 24.04 with systemd, SSH, and networking pre-configured.

### How long does a microVM take to boot?

Currently approximately 2 seconds. Target is sub-200ms with snapshot/resume (Phase 2).

### How much memory does each microVM use?

The base overhead is approximately 5-10 MiB per VM, plus whatever your workload requires.

### Can I create custom microVM images?

Yes. Create a Dockerfile extending the base image:

```dockerfile
FROM ghcr.io/daax-dev/nanofuse/base:latest
RUN apt-get update && apt-get install -y python3
```

See the [Developer Guide](DEVELOPER_GUIDE.md#building-microvm-images) for details.

### How do I SSH into a microVM?

1. Get the VM's IP address:

```bash
nanofuse vm inspect my-vm | grep ip
```

2. SSH in (if keys are configured):

```bash
ssh root@<VM_IP>
```

See [SSH Access Guide](SSH_ACCESS_QUICK_START.md) for SSH key configuration.

---

## Networking

### What IP range do microVMs use?

By default, microVMs use the `172.16.0.0/24` subnet with:

- Gateway: `172.16.0.1`
- DHCP range: `172.16.0.10-254`

### Can microVMs access the internet?

Yes, via NAT through the host. The daemon automatically configures iptables rules for outbound connectivity.

### Can microVMs communicate with each other?

Yes, VMs on the same host can communicate directly over the bridge network. Cross-host communication is planned for a future release.

### How do I expose a microVM port to the host?

Port forwarding is planned for Phase 2. Currently, access VMs directly via their IP address.

---

## Security

### How secure is Nanofuse?

Nanofuse provides multiple layers of isolation:

1. **Hardware virtualization**: KVM-based isolation
2. **Minimal VMM**: Firecracker's small attack surface (~50K lines of code)
3. **Jailer**: seccomp-bpf, cgroups, and chroot
4. **Network isolation**: Per-VM TAP devices with firewall rules

See [Project Goals](GOALS.md) for the complete security model.

### Is it safe to run untrusted code?

Nanofuse is designed for running untrusted code, but security is a continuous effort. Recommendations:

- Keep the host kernel updated
- Use the latest Firecracker release
- Implement network egress filtering
- Monitor VM resource consumption
- Review the [Firewall Capabilities](future-fw.md) documentation

### How does Nanofuse compare to gVisor or Kata Containers?

| Solution | Isolation | Performance | Complexity |
|----------|-----------|-------------|------------|
| Nanofuse/Firecracker | Hardware (KVM) | High | Low |
| Kata Containers | Hardware (KVM) | Medium | Medium |
| gVisor | User-space syscall filtering | Medium | High |

Nanofuse prioritizes simplicity and performance for the AI sandbox use case.

---

## Development and Contributing

### How do I build Nanofuse from source?

```bash
git clone https://github.com/daax-dev/nanofuse.git
cd nanofuse
./scripts/ensure-mage.sh
mage all
```

See the [Developer Guide](DEVELOPER_GUIDE.md) for complete instructions.

### How do I run tests?

```bash
# Unit tests
mage test

# Integration tests (requires running daemon)
mage testIntegration

# All tests
mage testAll
```

### How do I contribute?

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `mage ci`
5. Submit a pull request

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### What is the project roadmap?

| Phase | Focus | Status |
|-------|-------|--------|
| Phase 1 | Core infrastructure | In Progress |
| Phase 2 | Snapshot/resume for fast boot | Planned |
| Phase 3 | SDKs (Python, JavaScript) | Planned |
| Phase 4 | Advanced features | Planned |

See [Project Goals](GOALS.md) for detailed roadmap.

---

## Troubleshooting

### Common Issues

| Issue | Solution |
|-------|----------|
| `connection refused` | Start the daemon: `sudo nanofused` |
| `image not found` | Pull the image: `nanofuse image pull --default` |
| `KVM not available` | Check `/dev/kvm` exists and is accessible |
| `permission denied` | Use `sudo` or check group membership |

### Where are the logs?

| Log | Location |
|-----|----------|
| Daemon | `journalctl -u nanofused` |
| VM console | `/var/lib/nanofuse/vms/<vm-id>/console.log` |

### Where can I get help?

1. Check the [Troubleshooting Guide](TROUBLESHOOTING.md)
2. Search [GitHub Issues](https://github.com/daax-dev/nanofuse/issues)
3. Start a [Discussion](https://github.com/daax-dev/nanofuse/discussions)
4. Open a new issue with reproduction steps

---

*See also: [README](../README.md) | [Quick Start](QUICKSTART.md) | [Developer Guide](DEVELOPER_GUIDE.md)*
