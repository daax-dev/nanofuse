# nanofuse Vagrant Dev Environment

Disposable VM with full sudo + nested KVM for end-to-end nanofuse development and testing. Designed for AI agents (Claude) and humans alike.

## What You Get

One `vagrant up` gives you a fully provisioned Ubuntu 24.04 VM with:

- **Go 1.24.3** + mage build system
- **Firecracker 1.7.0** + jailer
- **Docker** (for base image builds)
- **nanofuse** built from source (CLI + daemon + register-local-image)
- **Base microVM image** (kernel + rootfs) built and registered
- **nanofused** configured as systemd service
- **Network tools** (iptables, dnsmasq, iproute2) for security layer testing
- **Full sudo** — install packages, modify iptables, run Firecracker, anything

## Quick Start

```bash
# From this directory
vagrant up           # ~10-15 min first run (kernel build)
vagrant ssh          # full sudo inside
vagrant destroy -f   # clean slate
```

## Placement

This is designed to work from two locations:

### Inside nanofuse repo (recommended)
```
nanofuse/
  dev/
    vagrant/
      Vagrantfile    # auto-detects ../../go.mod
      setup.sh
      verify.sh
      README.md
      docs/
```

### Standalone (with env var)
```bash
NANOFUSE_SRC=~/ps/daax/nanofuse vagrant up
```

## Claude Workflow

From the host, Claude runs commands inside the VM:

```bash
# Build and test
vagrant ssh -c "cd /nanofuse && sudo mage all"
vagrant ssh -c "cd /nanofuse && sudo mage testAll"

# Start daemon and create VMs
vagrant ssh -c "sudo systemctl start nanofused"
vagrant ssh -c "nanofuse health"
vagrant ssh -c "nanofuse vm list"

# Test security layers
vagrant ssh -c "sudo iptables -L -n"
vagrant ssh -c "sudo /nanofuse/scripts/e2e-test.sh"

# Re-provision after source changes (re-syncs /nanofuse)
vagrant provision

# Nuclear option
vagrant destroy -f && vagrant up
```

## What's Inside the VM

| Path | Contents |
|------|----------|
| `/nanofuse` | Source code (rsynced from host) |
| `/usr/local/bin/nanofuse` | CLI binary |
| `/usr/local/bin/nanofused` | Daemon binary |
| `/usr/local/bin/firecracker` | Firecracker binary |
| `/etc/nanofuse/nanofused.yaml` | Daemon config |
| `/var/lib/nanofuse/` | Data directory (DB, images) |
| `/var/lib/nanofuse/images/` | vmlinux + rootfs.ext4 |
| `/vagrant-scripts/` | These setup/verify scripts |

## VM Specs

- Ubuntu 24.04 (bento/ubuntu-24.04)
- 4 vCPUs, 4GB RAM
- KVM host-passthrough (nested virtualization)
- vagrant-libvirt provider

## Security Layer Testing

This VM is the testbed for porting microvm-sandbox's proven security layers into nanofuse:

1. **Escape containment** — 13 attack vectors (port from `microvm-sandbox/escape/`)
2. **Filesystem isolation** — sentinel file tests
3. **L3/L4 firewall** — per-TAP iptables rules (nanofuse advantage over Kata)
4. **DNS filtering** — dnsmasq domain allow/block
5. **SPIFFE identity + Vault** — JWT-SVID issuance and secret retrieval

See `docs/` for the full reuse analysis.

## Troubleshooting

```bash
# VM won't boot
vagrant up --debug 2>&1 | tail -50

# nanofused won't start
vagrant ssh -c "sudo journalctl -u nanofused -n 50"

# Rebuild nanofuse after source changes
vagrant ssh -c "cd /nanofuse && sudo mage all && sudo cp bin/* /usr/local/bin/"

# Re-sync source without full reprovision
vagrant rsync

# KVM not available inside VM
# Ensure host has: lv.cpu_mode = "host-passthrough" and lv.nested = true
```
