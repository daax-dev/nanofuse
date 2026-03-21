# Network Debug (SSH) Workflow

This script boots a VM, injects your SSH public key into a working copy of the base rootfs, forwards `localhost:2222` to the VM's `22/tcp`, and keeps everything running so you can debug inside the VM.

Prereqs: build the base image (`images/base/build/*.ext4` and `vmlinux` present) and have `sudo` access.

Quick start:

```bash
cd /path/to/nanofuse
sudo chmod +x scripts/test-network-e2e-ssh.sh
sudo ./scripts/test-network-e2e-ssh.sh
```

Flags:
- `--name NAME` VM name (default `debug-net-vm`)
- `--host-port PORT` host port forwarded to VM 22 (default `2222`)
- `--pubkey PATH` public key to inject (auto-detects `~/.ssh/id_ed25519.pub` or `id_rsa.pub`)
- `--no-ssh` do not auto-open SSH after boot
- `--cleanup-on-exit` stop/delete VM, stop daemon, and remove temp data when the script exits

After boot, connect to the VM:

```bash
ssh -o StrictHostKeyChecking=no -p 2222 root@localhost
```

Useful commands:

```bash
# VM logs
./bin/nanofuse --api-url http://127.0.0.1:8080 vm logs debug-net-vm | tail -n 100

# Inspect bridge/TAP
ip link show nanofuse0
bridge link show

# Inspect iptables NAT
iptables -t nat -L -n -v | sed -n '1,120p'

# Stop and delete VM when done
./bin/nanofuse --api-url http://127.0.0.1:8080 vm stop debug-net-vm
./bin/nanofuse --api-url http://127.0.0.1:8080 vm delete debug-net-vm --force
```

Daemon logs are written to `/tmp/nanofuse-daemon-debug.log`.
