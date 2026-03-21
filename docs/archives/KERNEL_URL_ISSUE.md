# Kernel URL Issue - Clarification

## The Situation

Your build failed because the versioned Firecracker kernel URL (v5.10.204) returned **404 Not Found**.

```
curl: (22) The requested URL returned error: 404
```

## Analysis

The original `build.sh` URL is:
```bash
https://s3.amazonaws.com/spec.ccfc.min/img/quickstart_guide/x86_64/kernels/vmlinux.bin
```

This generic URL **works and downloads a kernel**.

The attempted fix (version-pinned to 5.10.204):
```bash
https://s3.amazonaws.com/spec.ccfc.min/firecracker-ci/v1.7/x86_64/vmlinux-5.10.204
```

This URL **does not exist** - returns 404.

## Current Status

✅ **build.sh has been reverted** to use the working generic S3 URL

Now it will download whatever kernel Firecracker's S3 provides (currently 4.14.174).

## The Real Issue

The kernel that's provided by Firecracker's S3 quickstart (4.14.174) **is not ideal** but it's what's officially available.

Your console output showed:
```
Linux version 4.14.174
```

This kernel:
- ✅ Does boot (not completely broken)
- ❌ Is old (2021)
- ❌ Might have performance issues
- ❌ Might have missing features

## Why The VM "Hung"

Looking at your console output more carefully, the kernel **is booting** - you see kernel messages, systemd startup, etc. The "hang" is probably:

1. **systemd is hanging** during service startup
2. **Network initialization timing out** (waiting for DHCP)
3. **Device initialization slow** (old kernel with modern hardware)
4. **SSH key issue** - kernel booted fine, but can't SSH

## Next Steps

### Option 1: Use Current Kernel (Quickest)

Just rebuild with the current generic URL and investigate what's actually hanging:

```bash
cd images/base
sudo make clean
sudo make build

# Test
./bin/nanofuse vm create default test-vm
./bin/nanofuse vm start test-vm

# Watch full boot
tail -f /tmp/nanofuse/vms/*/console.log

# Wait 30 seconds - if you see systemd messages, it's booting
# If it gets stuck, note WHERE it's stuck
```

### Option 2: Find Alternative Kernel URL

If 4.14.174 doesn't work, we could try:
- Pre-built 5.10 kernels from other sources
- Building a custom kernel
- Using Docker Hub's Firecracker images

### Option 3: Accept 4.14 and Optimize

The 4.14 kernel might work fine - we just need to:
1. Let it boot completely (30+ seconds possible)
2. Check if systemd reaches multi-user
3. Test SSH access
4. If it works, we're good

## Recommendation

**Rebuild with current settings and debug what's actually hanging.**

The kernel itself is booting (we see boot messages in console.log). The hang is likely:
- Not the kernel binary
- Not Firecracker process
- But something in the boot sequence (systemd, network, drivers)

**Let's see the full console output** once the rebuild completes, and identify exactly where it stops.

---

## Technical Notes

### Why version-pinned URL didn't work

Firecracker's S3 bucket structure doesn't have a `firecracker-ci/v1.7/` path with individual kernel files. The quickstart URL is the public interface they maintain.

### What the quickstart URL serves

The generic URL:
```
https://s3.amazonaws.com/spec.ccfc.min/img/quickstart_guide/x86_64/kernels/vmlinux.bin
```

Points to their latest quickstart kernel (currently 4.14.174). This is the **officially recommended kernel** for Firecracker quickstart.

### Dockerfile Comments vs Reality

The Dockerfile has comments saying "5.10.204" but that was aspirational. The actual kernel available is 4.14.174, which is what we should work with.

---

## Summary

| Issue | Status | Action |
|-------|--------|--------|
| build.sh URL error | FIXED | Reverted to working generic URL |
| Kernel version | KNOWN | 4.14.174 (from S3 quickstart) |
| "Hang" cause | UNKNOWN | Need full console output to diagnose |
| Next step | DO THIS | Rebuild and investigate boot sequence |
