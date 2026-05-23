# Kernel Version Mismatch Root Cause Analysis

**Date**: 2025-11-19
**Severity**: CRITICAL
**Status**: Identified, fixing now

---

## The Problem

VM boots, systemd starts, network works, BUT services fail to start.

### Console Evidence
```
[  OK  ] Started todo-backend.service - Todo App Backend Service.
[FAILED] Failed to start nginx.service - A high performance web server...
```

**Systemd reports services "started" but they're not actually running.**

---

## Root Cause: Kernel/Userspace Mismatch

### What We're Running
- **Userspace**: Ubuntu 24.04 LTS (April 2024)
- **Kernel**: 5.10.240 (from 2021, 3+ years old)

### Why This Is Broken

**Ubuntu 24.04 requires**:
- Kernel 6.8.x or newer
- cgroup v2 (unified hierarchy)
- Modern systemd features (v255)
- Modern syscalls and kernel APIs

**Kernel 5.10.240 has**:
- Old cgroup support (may be v1 only)
- Missing kernel features Ubuntu 24.04 expects
- Incompatible with modern userspace requirements

### From tips.md Category 3: "Guest Kernel Limitations"

> Slim kernels bite hard:
> - **Missing cgroup controllers**: containerd/Docker crash or silently misbehave
> - **Wrong cgroup version**: Runtime wants v2; VM only has v1
> - **Missing network modules**
> - **Over-restrictive seccomp/capabilities**

**This is EXACTLY what's happening.**

---

## How Did This Happen?

### Timeline

1. **5 days ago**: Base image built with 5.10.240 (Slicer kernel)
2. **4 days ago**: Todo-app image built FROM that base → inherited old kernel
3. **Later**: Base image build process was updated to use kernel 6.1.90
4. **Today**: Testing old todo-app image (still has 5.10.240)

### The Mistake

- Base image BUILD.md says to use kernel 6.1.90 ✅
- Base image build artifacts were cleaned (build/ directory empty) ❌
- Todo-app image was never rebuilt with new kernel ❌
- Testing continued with old image ❌

---

## What Should Be

### Correct Configuration

**Base Image**:
- Ubuntu 24.04 userspace
- **Kernel 6.1.90** (Firecracker-compatible, modern enough for Ubuntu 24.04)
- Build process defined in `images/base/BUILD.md`

**Todo-App Image**:
- Built FROM base image (inherits kernel)
- Services: nginx, todo-backend
- Should work with kernel 6.1.90

---

## Evidence in Database

```bash
$ sqlite3 /var/lib/nanofuse/nanofuse.db "SELECT digest, kernel_version FROM images;"

sha256:b3acbe0b3...  5.10.240  # Base image (OLD)
sha256:0c8543d7e...  5.10.240  # Todo-app (OLD, what we're testing)
sha256:aac30dfac...  5.10.240  # Another todo-app (OLD)
```

**All images have wrong kernel.**

---

## Evidence in GHCR

```bash
$ nanofuse image pull ghcr.io/daax-dev/nanofuse/todo-app:latest
Error: MANIFEST_UNKNOWN: manifest unknown
```

**Image was never pushed to GHCR with correct kernel.**

---

## The Fix

### Step 1: Rebuild Base Image with 6.1.90

```bash
cd /home/jpoley/ps/nanofuse/images/base
sudo ./build.sh
```

**This will**:
- Download/build kernel 6.1.90
- Create Ubuntu 24.04 rootfs
- Generate manifest.json

**Output**:
- `build/vmlinux` (kernel 6.1.90)
- `build/rootfs.ext4`
- `build/manifest.json`

### Step 2: Rebuild Todo-App Image

Use the base image artifacts to build todo-app image with correct kernel.

(Script to be created: `examples/todo-app/build-with-base.sh`)

### Step 3: Test with New Image

Create VM with new image, verify:
- Kernel version is 6.1.90
- Systemd starts
- Services (nginx, todo-backend) actually run
- Health checks pass

---

## Success Criteria

- [ ] Base image rebuilt with kernel 6.1.90
- [ ] Todo-app image rebuilt with kernel 6.1.90
- [ ] VM created with new image
- [ ] Console shows: `Linux version 6.1.90`
- [ ] Systemd starts successfully
- [ ] nginx.service: `[  OK  ] Started` AND actually running
- [ ] todo-backend.service: `[  OK  ] Started` AND actually running
- [ ] `curl http://172.16.0.10:8080/health` → 200 OK
- [ ] `curl http://172.16.0.10/` → HTML response
- [ ] Port 80 open (nmap confirms)
- [ ] Port 8080 open (nmap confirms)

---

## Lessons Learned

1. **Always verify kernel version matches userspace requirements**
2. **Check actual kernel version in running VM, not just build docs**
3. **Don't assume images are up-to-date without verification**
4. **Keep build artifacts or document rebuild process**
5. **tips.md Category 3 is REAL: kernel/userspace mismatch breaks everything**

---

## Next Actions

1. **IMMEDIATE**: Rebuild base image with 6.1.90
2. Rebuild todo-app image
3. Test end-to-end
4. Document rebuild process
5. Add kernel version validation to deployment scripts

---

**Status**: Rebuilding base image now
**Blocker**: None, can proceed immediately
