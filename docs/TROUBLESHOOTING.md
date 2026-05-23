# NanoFuse Troubleshooting Guide

This guide covers common issues, diagnostic commands, and resolutions for NanoFuse.

## Quick Reference: Common Symptoms

| Symptom | Likely Cause | Jump To |
|---------|--------------|---------|
| `nanofuse` command not found | Binary not installed | [Daemon Issues](#daemon-issues) |
| "connection refused" | Daemon not running | [Daemon Issues](#daemon-issues) |
| VM hangs during boot | Kernel version mismatch | [VM Lifecycle Issues](#vm-lifecycle-issues) |
| Services don't start in VM | Systemd configuration | [Service Issues](#service-issues) |
| No network connectivity | Bridge/TAP issues | [Network Issues](#network-issues) |
| Image pull fails | Authentication | [Image Issues](#image-issues) |
| "readonly database" | Permission issues | [Daemon Issues](#daemon-issues) |

---

## Daemon Issues

### Daemon Not Running

**Symptoms:**
- `connection refused` errors
- `nanofuse` commands hang or fail

**Diagnostic Commands:**
```bash
# Check if daemon is running
sudo systemctl status nanofused

# Check daemon logs
sudo journalctl -u nanofused -n 50 --no-pager

# Check if socket/port is listening
sudo ss -tlnp | grep -E '8080|nanofuse'
ls -la /run/nanofused.sock
```

**Resolution:**
1. Start the daemon:
   ```bash
   sudo systemctl start nanofused
   ```

2. If it fails to start, check logs:
   ```bash
   sudo journalctl -u nanofused -f
   ```

3. Common startup failures:
   - **Firecracker binary missing**: Install at `/usr/local/bin/firecracker`
   - **Network bridge creation failed**: Requires root/CAP_NET_ADMIN
   - **Database permission error**: Check `/var/lib/nanofuse/` ownership

### Database Permission Errors

**Symptoms:**
- "readonly database" errors
- Database locked errors

**Diagnostic Commands:**
```bash
# Check database ownership
ls -la /var/lib/nanofuse/nanofuse.db

# Check directory permissions
ls -la /var/lib/nanofuse/
```

**Resolution:**
```bash
# Fix ownership (daemon runs as root)
sudo chown -R root:root /var/lib/nanofuse/
sudo chmod 755 /var/lib/nanofuse/
sudo chmod 644 /var/lib/nanofuse/nanofuse.db
```

---

## VM Lifecycle Issues

### VM Fails to Boot / Hangs

**Symptoms:**
- VM stuck in "starting" state
- No console output after kernel load
- Kernel panic messages

**Diagnostic Commands:**
```bash
# Get VM ID
VM_ID=$(sudo nanofuse vm inspect <vm-name> --json | jq -r '.id')

# Check console log
sudo cat /var/lib/nanofuse/vms/${VM_ID}/console.log

# Check last 50 lines of console
sudo tail -50 /var/lib/nanofuse/vms/${VM_ID}/console.log

# Look for kernel panics
sudo grep -i "panic\|error\|failed" /var/lib/nanofuse/vms/${VM_ID}/console.log
```

**Common Causes:**

1. **Kernel Version Mismatch**
   - Old kernel (4.14.x) missing required drivers
   - Fix: Use kernel 5.10.204 or newer
   - Verify: `strings vmlinux | grep -i "Linux version"`

2. **Invalid rootfs**
   - Image built with `docker save` instead of `docker export`
   - Rootfs contains OCI manifests instead of filesystem
   - Fix: Rebuild image using correct export method

   ```bash
   # Check if rootfs is valid (should show directories, not JSON)
   sudo debugfs -R "ls -l /" /path/to/rootfs.ext4
   ```

3. **Missing init system**
   - Kernel can't find `/lib/systemd/systemd` or `/sbin/init`
   - Fix: Ensure image has systemd installed and kernel cmdline has `init=/lib/systemd/systemd`

### VM Won't Start After Creation

**Diagnostic Commands:**
```bash
# Check VM state
sudo nanofuse vm status <vm-name> --json | jq '.state'

# Check daemon logs for errors
sudo journalctl -u nanofused -n 100 | grep -i "<vm-name>\|error"
```

**Resolution:**
1. Delete and recreate:
   ```bash
   sudo nanofuse vm delete <vm-name> -f
   sudo nanofuse vm create <image> <vm-name>
   sudo nanofuse vm start <vm-name>
   ```

2. Check for resource issues (memory, storage)

---

## Network Issues

### VM Has No Network / No IP Address

**Diagnostic Commands:**
```bash
# Check bridge exists
ip addr show nanofuse0

# Check route exists
ip route show 172.16.0.0/24

# Check TAP devices
ip link show | grep tap

# Check iptables NAT rules
sudo iptables -t nat -L -n | grep 172.16
```

**Resolution:**

1. **Bridge not created:**
   ```bash
   # Restart daemon (creates bridge on startup)
   sudo systemctl restart nanofused

   # Or manually create
   sudo ip link add nanofuse0 type bridge
   sudo ip addr add 172.16.0.1/24 dev nanofuse0
   sudo ip link set nanofuse0 up
   ```

2. **NAT not configured:**
   ```bash
   sudo iptables -t nat -A POSTROUTING -s 172.16.0.0/24 -o eth0 -j MASQUERADE
   sudo sysctl -w net.ipv4.ip_forward=1
   ```

### VM Has IP But Services Unreachable

**Diagnostic Commands:**
```bash
# Get VM IP
VM_IP=$(sudo nanofuse vm status <vm-name> --json | jq -r '.runtime.network_info.guest_ip')

# Test connectivity
ping -c 3 $VM_IP

# Test HTTP
curl -v --max-time 5 http://${VM_IP}/
curl -v --max-time 5 http://${VM_IP}/health

# Check from inside VM (if SSH available)
ssh root@${VM_IP} "systemctl status todo-backend"
ssh root@${VM_IP} "netstat -tlnp"
```

**Resolution:**
1. Services may still be starting - wait 10-20 seconds after boot
2. Check service logs inside VM (see [Service Issues](#service-issues))
3. Verify firewall isn't blocking traffic

---

## Image Issues

### Image Pull Authentication Failed

**Symptoms:**
- "unauthorized" or "authentication required" errors
- Image pull hangs

**Diagnostic Commands:**
```bash
# Test Docker login
echo $GITHUB_TOKEN | docker login ghcr.io -u x-access-token --password-stdin

# Check token is set
echo ${GITHUB_TOKEN:0:10}...
```

**Resolution:**
1. Set GITHUB_TOKEN:
   ```bash
   export GITHUB_TOKEN=$(cat ~/.aws/vault/GITHUB)  # Or your token location
   ```

2. Verify token has `read:packages` scope for GHCR

3. For private repositories, ensure token has `repo` scope

### Image Registration Fails

**Diagnostic Commands:**
```bash
# List registered images
sudo nanofuse image list

# Check image files exist
ls -la /path/to/kernel /path/to/rootfs.ext4
```

**Resolution:**
1. Ensure kernel and rootfs paths are correct
2. Verify rootfs is ext4 format:
   ```bash
   file rootfs.ext4
   # Should show: "Linux rev 1.0 ext4 filesystem"
   ```

### Image Build Issues

**Common Problem: rootfs contains OCI metadata instead of filesystem**

**Diagnostic:**
```bash
# Mount and check rootfs contents
sudo mkdir -p /mnt/rootfs
sudo mount -o loop rootfs.ext4 /mnt/rootfs
ls /mnt/rootfs
# Should see: bin, etc, lib, usr, var, etc.
# NOT: blobs/, index.json, oci-layout
sudo umount /mnt/rootfs
```

**Fix:**
Use `docker export` instead of `docker save`:
```bash
# Correct way to extract filesystem
CONTAINER_ID=$(docker create myimage:latest)
docker export $CONTAINER_ID | tar -C rootfs_dir -xf -
docker rm $CONTAINER_ID

# Then create ext4 image
# ... (see build scripts)
```

---

## Service Issues

### Services Don't Start Inside VM

**Diagnostic Commands:**
```bash
# SSH into VM (if available)
VM_IP=$(sudo nanofuse vm status <vm-name> --json | jq -r '.runtime.network_info.guest_ip')
ssh root@${VM_IP}

# Inside VM:
systemctl status
systemctl list-units --failed
journalctl -b --no-pager | tail -100
```

**Common Causes:**

1. **Systemd not running as init**
   - Check kernel cmdline has `init=/lib/systemd/systemd`
   - Check Dockerfile ends with `CMD ["/lib/systemd/systemd"]`

2. **Service file errors**
   - Check service file syntax: `systemctl cat todo-backend`
   - Check service enabled: `systemctl is-enabled todo-backend`

3. **Dependency issues**
   - Check service dependencies: `systemctl list-dependencies todo-backend`
   - Look for failed dependencies in journal

### Health Endpoint Not Responding

**Diagnostic Commands:**
```bash
# From host
curl -v http://${VM_IP}/health

# From inside VM
curl -v http://localhost/health
curl -v http://localhost:80/health

# Check what's listening
netstat -tlnp | grep -E ':80|:8080'
```

**Resolution:**
1. Verify backend binary is running:
   ```bash
   ps aux | grep todo-server
   ```

2. Check backend logs:
   ```bash
   journalctl -u todo-backend -n 50
   ```

3. Verify port configuration in systemd service file

---

## Log Locations

| Log | Location | Purpose |
|-----|----------|---------|
| Daemon logs | `journalctl -u nanofused` | Daemon operations, API requests |
| VM console | `/var/lib/nanofuse/vms/<vm-id>/console.log` | Kernel boot, systemd, service output |
| VM config | `/var/lib/nanofuse/vms/<vm-id>/vm.json` | VM configuration |
| Database | `/var/lib/nanofuse/nanofuse.db` | VM state, images, IPs |
| Snapshots | `/var/lib/nanofuse/snapshots/<vm-id>/` | VM snapshots |

**Useful Log Commands:**
```bash
# Follow daemon logs
sudo journalctl -u nanofused -f

# Get VM ID for a named VM
VM_ID=$(sudo nanofuse vm inspect myvm --json | jq -r '.id')

# View VM console log
sudo cat /var/lib/nanofuse/vms/${VM_ID}/console.log

# Search for errors in console log
sudo grep -iE "error|fail|panic" /var/lib/nanofuse/vms/${VM_ID}/console.log

# View systemd boot timeline (inside VM via SSH)
systemd-analyze blame
```

---

## Diagnostic Scripts

NanoFuse includes several diagnostic scripts:

```bash
# Health check for a running VM
./scripts/health-check.sh <vm-name>
./scripts/health-check.sh <vm-name> --json

# Full rebuild and test cycle
sudo ./scripts/dev-rebuild.sh

# Network end-to-end test
sudo ./scripts/test-network-e2e.sh

# End-to-end workflow test
sudo ./test/e2e/full-workflow-test.sh --verbose
```

---

## Known Issues & Workarounds

### SSH Host Key Warnings

**Issue:** All VMs have identical SSH host keys (generated at image build time).

**Workaround:**
```bash
# Remove old host key before connecting
ssh-keygen -R $VM_IP

# Or skip host key checking (less secure)
ssh -o StrictHostKeyChecking=no root@$VM_IP
```

**Proper Fix:** Regenerate host keys on first boot (included in newer images).

### IP Address Reuse

**Issue:** Stopped VMs retain IP allocations, new VMs may get different IPs.

**Workaround:**
```bash
# Delete VM to release IP
sudo nanofuse vm delete <vm-name> -f

# Check current allocations
sudo nanofuse vm list
```

### Docker Network Conflicts

**Issue:** Docker's iptables rules can conflict with NanoFuse networking.

**Workaround:**
```bash
# Restart nanofused after Docker operations
sudo systemctl restart nanofused

# Or manually re-add NAT rules
sudo iptables -t nat -A POSTROUTING -s 172.16.0.0/24 ! -d 172.16.0.0/24 -j MASQUERADE
```

---

## Getting Help

1. Check this troubleshooting guide
2. Review daemon logs: `sudo journalctl -u nanofused -n 100`
3. Review VM console logs: `/var/lib/nanofuse/vms/<vm-id>/console.log`
4. Run diagnostic scripts: `./scripts/health-check.sh <vm-name>`
5. File an issue: https://github.com/daax-dev/nanofuse/issues
