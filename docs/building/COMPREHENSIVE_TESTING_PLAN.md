# Comprehensive Testing Plan for NanoFuse VM Services

**Date**: 2025-11-19
**Context**: VM boots with kernel 6.1.90, but services (nginx, todo-backend) not accessible
**Based on**: `docs/building/tips.md` - systematic debugging approach

---

## Current State

✅ **What Works**:
- Kernel 6.1.90 built and deployed
- VM creates and starts successfully
- VM is pingable (172.16.0.10)
- Firecracker/TAP/bridge networking functional

❌ **What Doesn't Work**:
- Nginx on port 80: Connection refused
- Todo-backend on port 8080: Connection refused
- Services may have started briefly then failed

---

## Testing Strategy

**Principle**: Test bottom-to-top, validate each layer before moving up.

### Layer 0: Kernel & Boot (Foundational)

**Tests**:
1. Verify kernel version in console log
2. Check if systemd started as PID 1
3. Verify boot completed (multi-user.target reached)
4. Check for kernel panics or critical errors

**Commands**:
```bash
sudo grep "Linux version" /var/lib/nanofuse/vms/[VM_ID]/console.log
sudo grep "systemd\[1\]" /var/lib/nanofuse/vms/[VM_ID]/console.log | head -20
sudo grep "Reached target multi-user" /var/lib/nanofuse/vms/[VM_ID]/console.log
sudo grep -i "panic\|oops\|bug:" /var/lib/nanofuse/vms/[VM_ID]/console.log
```

**Success Criteria**:
- Kernel 6.1.90 detected ✅
- Systemd PID 1 messages present ✅
- Boot reaches multi-user.target ✅
- No kernel panics ✅

---

### Layer 1: Systemd Service Status

**Tests**:
1. Check if nginx.service attempted to start
2. Check if todo-backend.service attempted to start
3. Look for service failure messages
4. Check for dependency issues

**Commands**:
```bash
sudo grep "nginx.service" /var/lib/nanofuse/vms/[VM_ID]/console.log
sudo grep "todo-backend.service" /var/lib/nanofuse/vms/[VM_ID]/console.log
sudo grep "\[FAILED\]" /var/lib/nanofuse/vms/[VM_ID]/console.log
sudo grep -i "dependency\|After=\|Requires=" /var/lib/nanofuse/vms/[VM_ID]/console.log
```

**Success Criteria**:
- Services attempted to start
- If failed: specific error messages visible
- Dependency chain understood

---

### Layer 2: Binary & Library Availability

**Tests**:
1. Check if nginx binary exists in rootfs
2. Check if todo-server binary exists in rootfs
3. Verify execute permissions
4. Check for library dependencies

**Commands**:
```bash
# Mount rootfs and check
sudo mkdir -p /tmp/nanofuse-test-mount
sudo mount -o loop /home/jpoley/ps/nanofuse/examples/todo-app/output/rootfs.ext4 /tmp/nanofuse-test-mount

# Check binaries
ls -la /tmp/nanofuse-test-mount/usr/sbin/nginx
ls -la /tmp/nanofuse-test-mount/usr/local/bin/todo-server

# Check systemd service files
cat /tmp/nanofuse-test-mount/etc/systemd/system/nginx.service
cat /tmp/nanofuse-test-mount/etc/systemd/system/todo-backend.service

# Check nginx config
cat /tmp/nanofuse-test-mount/etc/nginx/sites-available/default

# Unmount
sudo umount /tmp/nanofuse-test-mount
```

**Success Criteria**:
- Binaries exist ✅
- Binaries are executable ✅
- Service files are correct ✅
- Config files are present ✅

---

### Layer 3: Network Configuration (Guest OS)

**Tests**:
1. Verify VM has correct IP (172.16.0.10)
2. Check if default route is configured
3. Verify DNS configuration
4. Check guest firewall rules

**Commands**:
```bash
# Check console for network config
sudo grep "ip=" /var/lib/nanofuse/vms/[VM_ID]/console.log
sudo grep -E "eth0|ens|enp" /var/lib/nanofuse/vms/[VM_ID]/console.log

# Check for networkd messages
sudo grep "systemd-networkd" /var/lib/nanofuse/vms/[VM_ID]/console.log

# Check for firewall messages
sudo grep -iE "iptables|nftables|ufw|firewall" /var/lib/nanofuse/vms/[VM_ID]/console.log
```

**Success Criteria**:
- IP 172.16.0.10 assigned ✅
- Network interface up ✅
- Gateway configured ✅
- No blocking firewall rules ✅

---

### Layer 4: Service-Specific Issues

**For Nginx**:
1. Check if port 80 is already in use
2. Verify nginx config syntax
3. Check for permission issues
4. Look for SSL/certificate errors

**For Todo-Backend**:
1. Check if port 8080 is available
2. Verify DuckDB database path `/data/todos.db`
3. Check for Go runtime errors
4. Verify environment variables

**Commands**:
```bash
# Check console for port conflicts
sudo grep "Address already in use" /var/lib/nanofuse/vms/[VM_ID]/console.log

# Check for permission denied
sudo grep "Permission denied" /var/lib/nanofuse/vms/[VM_ID]/console.log

# Check for missing files/dirs
sudo grep "No such file or directory" /var/lib/nanofuse/vms/[VM_ID]/console.log

# Check for database issues
sudo grep -i "duckdb\|database\|/data" /var/lib/nanofuse/vms/[VM_ID]/console.log
```

**Success Criteria**:
- No port conflicts ✅
- No permission errors ✅
- All required files/dirs present ✅

---

### Layer 5: Race Conditions & Timing

**Tests**:
1. Check if services started too early (before network)
2. Look for restart attempts
3. Check service dependencies in systemd

**Commands**:
```bash
# Check service timing relative to network
sudo grep -B5 -A5 "Starting nginx" /var/lib/nanofuse/vms/[VM_ID]/console.log | grep "network"
sudo grep "Restart=" /var/lib/nanofuse/vms/[VM_ID]/console.log

# Check for multiple start attempts
sudo grep -c "Starting nginx" /var/lib/nanofuse/vms/[VM_ID]/console.log
sudo grep -c "Starting todo-backend" /var/lib/nanofuse/vms/[VM_ID]/console.log
```

**Success Criteria**:
- Services wait for network.target ✅
- Restart policies appropriate ✅
- No premature starts ✅

---

### Layer 6: Kernel Module & Cgroup Support

**Tests**:
1. Verify cgroup v2 is available
2. Check for missing kernel modules
3. Verify systemd can create cgroups

**Commands**:
```bash
# Check for cgroup messages
sudo grep -i "cgroup" /var/lib/nanofuse/vms/[VM_ID]/console.log

# Check for module loading errors
sudo grep "FATAL\|modprobe\|module" /var/lib/nanofuse/vms/[VM_ID]/console.log

# Check systemd cgroup setup
sudo grep "systemd\[1\].*cgroup" /var/lib/nanofuse/vms/[VM_ID]/console.log
```

**Success Criteria**:
- Cgroups available ✅
- No missing modules ✅
- Systemd can manage resources ✅

---

## Execution Plan

### Phase 1: Data Gathering (No Changes)
1. Run all Layer 0-6 tests
2. Collect console log
3. Document all findings
4. Identify root cause

### Phase 2: Hypothesis Formation
Based on findings, form specific hypotheses:
- **H1**: Service binaries missing or broken
- **H2**: Port binding issues
- **H3**: Race condition (services start before network ready)
- **H4**: Permission/ownership issues
- **H5**: Kernel module/cgroup incompatibility
- **H6**: Configuration file errors

### Phase 3: Targeted Fixes
For each hypothesis:
1. Create minimal test case
2. Implement fix
3. Rebuild image
4. Test
5. Document result

### Phase 4: Validation
Once services work:
1. Test all endpoints
2. Verify persistence across restarts
3. Check resource usage
4. Validate logs

---

## Test Automation Script

Create `scripts/building/run-comprehensive-tests.sh`:

```bash
#!/bin/bash
# Run all layers of testing systematically

VM_ID="$1"

if [ -z "$VM_ID" ]; then
    echo "Usage: $0 <vm-id>"
    exit 1
fi

CONSOLE_LOG="/var/lib/nanofuse/vms/$VM_ID/console.log"

echo "=== Layer 0: Kernel & Boot ==="
echo "Kernel:"
sudo grep "Linux version" "$CONSOLE_LOG" | head -1

echo "Systemd PID 1:"
sudo grep "systemd\[1\]:.*Detected" "$CONSOLE_LOG" | head -1

echo "Boot Complete:"
sudo grep "Reached target multi-user" "$CONSOLE_LOG"

echo ""
echo "=== Layer 1: Service Status ==="
echo "Nginx attempts:"
sudo grep "nginx.service" "$CONSOLE_LOG" | tail -5

echo "Todo-backend attempts:"
sudo grep "todo-backend.service" "$CONSOLE_LOG" | tail -5

echo "Failures:"
sudo grep "\[FAILED\]" "$CONSOLE_LOG"

echo ""
echo "=== Layer 3: Network ==="
echo "IP Assignment:"
sudo grep "ip=" "$CONSOLE_LOG" | head -1

echo "Network Ready:"
sudo grep "network.*target" "$CONSOLE_LOG" | tail -3

echo ""
echo "=== Layer 4: Errors ==="
echo "Permission Issues:"
sudo grep -i "permission denied" "$CONSOLE_LOG" | head -5

echo "Missing Files:"
sudo grep "No such file" "$CONSOLE_LOG" | head -5

echo "Port Conflicts:"
sudo grep "Address already in use" "$CONSOLE_LOG" | head -5

echo ""
echo "=== Layer 6: Kernel/Cgroup ==="
echo "Cgroup Messages:"
sudo grep -i "cgroup" "$CONSOLE_LOG" | head -5

echo ""
echo "Full console log: $CONSOLE_LOG"
```

---

## Success Metrics

**Minimum Viable**:
- ✅ Kernel 6.1.90 boots
- ✅ Systemd starts
- ✅ Nginx serves on port 80
- ✅ Todo-backend responds on port 8080

**Production Ready**:
- ✅ Services survive VM restart
- ✅ No memory leaks
- ✅ Performance acceptable
- ✅ Logs accessible
- ✅ Health checks pass

---

## Next Steps

1. **Run comprehensive test script** on current VM
2. **Analyze findings** layer by layer
3. **Form hypothesis** based on evidence
4. **Implement targeted fix** (not guessing)
5. **Test fix**
6. **Document resolution**
7. **Push working images to GHCR**
8. **Never lose this kernel again**

---

**Principle**: Evidence → Hypothesis → Fix → Test → Document

No more rabbit holes. Systematic debugging only.
