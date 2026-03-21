# SSH Host Keys Issue - CRITICAL

## Problem

**All VMs built from the same image have IDENTICAL SSH host keys.**

This is a **critical security issue**.

### What's Happening

1. When the Docker image is built, openssh-server generates SSH host keys:
   - `/etc/ssh/ssh_host_rsa_key`
   - `/etc/ssh/ssh_host_ecdsa_key`
   - `/etc/ssh/ssh_host_ed25519_key`

2. These keys are baked into the image

3. Every VM created from that image has the **exact same keys**

4. Multiple VMs with identical SSH host keys = **cryptographic disaster**

### Why This Is Bad

1. **SSH clients get MitM warnings**
   ```
   @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
   @    WARNING: REMOTE HOST IDENTIFICATION HAS CHANGED!
   @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
   ```

2. **SSH known_hosts gets corrupted**
   ```bash
   # You add VM1 to known_hosts:
   172.16.0.10 ssh-ed25519 AAAAC3NzaC1...

   # Later, VM2 (different IP, SAME HOST KEY!) tries to connect:
   ssh root@172.16.0.11
   # SSH warns: "Host key has changed!"
   ```

3. **Security compromise**
   - Host keys should be unique per system
   - Identical keys allow impersonation
   - Defeats SSH host verification

4. **User friction**
   - Users have to `ssh-keygen -R` to fix known_hosts
   - Or accept host key warnings (bad practice)
   - Or disable host key verification (worse practice)

---

## Current Implementation

### What Happens Now

1. Image is built with openssh-server
2. dpkg generates host keys during package installation
3. Keys are baked into the image at `/etc/ssh/ssh_host_*_key`
4. Every VM from that image has identical keys

### What Firstboot Does (Currently)

The `units/firstboot.service` only **logs** the keys:

```bash
echo "[$(date -Iseconds)] SSH host keys:" | tee -a /var/log/nanofuse/firstboot.log
ls -la /etc/ssh/ssh_host_*_key.pub | tee -a /var/log/nanofuse/firstboot.log
```

It does NOT regenerate them.

---

## Solution: Regenerate Keys on First Boot

### Option 1: Remove Keys from Image, Generate on First Boot (RECOMMENDED)

**Dockerfile**:
```dockerfile
# Remove pre-generated SSH host keys so they're regenerated on first boot
RUN rm -f /etc/ssh/ssh_host_*_key /etc/ssh/ssh_host_*_key.pub
```

**Firstboot service** (in units/firstboot.service):
```bash
# Regenerate SSH host keys on first boot
if [ ! -f /etc/ssh/ssh_host_rsa_key ]; then
    echo "Regenerating SSH host keys..."
    dpkg-reconfigure openssh-server 2>&1 | tee -a /var/log/nanofuse/firstboot.log
    systemctl restart ssh
    echo "SSH host keys regenerated"
fi
```

**Pros**:
- ✅ Each VM gets unique keys
- ✅ Simple and reliable
- ✅ Standard approach used by cloud images
- ✅ No security compromise

**Cons**:
- Boot takes ~1-2 seconds longer (SSH key generation)
- First boot log will show regeneration process

**Verdict**: Best option - use this.

### Option 2: Mount /etc/ssh as Ephemeral

Use systemd-tmpfiles to create transient SSH keys:

```bash
# /etc/tmpfiles.d/ssh.conf
d /etc/ssh/keys 0755 root root -
L+ /etc/ssh/ssh_host_rsa_key - - - - /run/ssh/ssh_host_rsa_key
L+ /etc/ssh/ssh_host_rsa_key.pub - - - - /run/ssh/ssh_host_rsa_key.pub
```

**Pros**:
- Keys never stored to disk
- Unique per VM automatically

**Cons**:
- Fragile, breaks SSH client expectations
- Hard to debug
- Complex to implement
- Not standard practice

**Verdict**: Too complex, don't use.

### Option 3: Generate in Dockerfile, Don't Cache

```dockerfile
# Build SSH keys fresh (not cached)
RUN ssh-keygen -A
```

With `docker build --no-cache`, forces regeneration each build.

**Pros**:
- Simple command

**Cons**:
- ❌ Still creates identical keys for all VMs (doesn't solve problem)
- Requires `--no-cache` flag always
- Still bakes keys into image

**Verdict**: Doesn't solve the problem.

---

## Recommended Fix

### Step 1: Update Dockerfile

Remove pre-generated host keys:

```dockerfile
# Remove pre-generated SSH host keys - they must be regenerated per VM on first boot
# This ensures each VM has unique SSH host keys
RUN rm -f /etc/ssh/ssh_host_*_key /etc/ssh/ssh_host_*_key.pub
```

### Step 2: Update Firstboot Service

Regenerate keys on first boot:

```bash
# In units/firstboot.service ExecStart:

# Regenerate SSH host keys if missing
if [ ! -f /etc/ssh/ssh_host_rsa_key ]; then
    echo "[$(date -Iseconds)] Regenerating SSH host keys..." | tee -a /var/log/nanofuse/firstboot.log
    dpkg-reconfigure openssh-server 2>&1 | tee -a /var/log/nanofuse/firstboot.log
    systemctl restart ssh
    echo "[$(date -Iseconds)] SSH host keys regenerated" | tee -a /var/log/nanofuse/firstboot.log
fi
```

### Step 3: Test

```bash
# Build image
sudo make build

# Create two VMs
./bin/nanofuse vm create default vm1
./bin/nanofuse vm create default vm2

# Start both
./bin/nanofuse vm start vm1
./bin/nanofuse vm start vm2

# Check keys are different
./bin/nanofuse vm exec vm1 cat /etc/ssh/ssh_host_ed25519_key.pub
./bin/nanofuse vm exec vm2 cat /etc/ssh/ssh_host_ed25519_key.pub
# Should be DIFFERENT

# SSH to both (should not get MitM warnings)
ssh root@<vm1-ip>
ssh root@<vm2-ip>
```

---

## Implementation Status

### Current State
- ❌ Host keys baked into image (WRONG)
- ❌ Firstboot doesn't regenerate them (INCOMPLETE)
- ❌ All VMs share identical keys (SECURITY ISSUE)

### What Needs To Happen
1. Update Dockerfile to remove host keys
2. Update firstboot.service to regenerate keys
3. Test that each VM gets unique keys
4. Verify SSH works without MitM warnings

---

## Timeline Impact

- **Boot time**: +1-2 seconds (for first boot only)
- **Subsequent boots**: No impact (keys already exist)
- **User experience**: Better (no SSH warnings)
- **Security**: Dramatically improved (unique keys per VM)

---

## See Also

- Cloud Images Standard: https://cloudinit.readthedocs.io/
- AWS EC2: Generates unique keys per instance
- OpenStack: Regenerates keys per boot
- Best Practice: Always regenerate ephemeral system keys

---

## Questions for You

1. **Should we fix this now or after Phase 1?**
   - Recommend: Fix now (critical security issue)
   - Impact: 1-2 lines in Dockerfile, 5-10 lines in firstboot.service

2. **Is firstboot the right place for this?**
   - Alternative: Use a separate ssh-keygen.service
   - Current: firstboot seems reasonable since it runs once

3. **Should we document this for users?**
   - Yes - users need to know keys are regenerated on first boot
   - Add note to README: "Each VM gets unique SSH host keys"

---

## Action Items

- [ ] Remove host keys from Dockerfile
- [ ] Add SSH key regeneration to firstboot.service
- [ ] Test with multiple VMs
- [ ] Verify SSH works without warnings
- [ ] Update documentation
- [ ] Update Firecracker image guide
