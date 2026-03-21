# Why a Linux Binary (like Firecracker) Works from Shell but Dies as a Subprocess

## 1) Environment & PATH aren’t the same
Different `PATH`, `HOME`, `SHELL`, `PWD`, `TMPDIR`, locale, etc.
`s‍udo`/`systemd` often scrub or override environment variables.

### Triage
```bash
# From the good interactive shell:
( set -o posix ; set ) | sort > /tmp/env.good

# From the failing launcher (script/python/systemd/Docker), print env right before exec:
env | sort > /tmp/env.bad

diff -u /tmp/env.good /tmp/env.bad
```

---

## 2) Working directory changed → relative paths break
Any relative path in args/configs (kernel image, rootfs, API socket) may now resolve elsewhere.

### Triage
```bash
pwd; ls -l
# Make all paths absolute in your JSON/CLI:
# e.g., "kernel_image_path": "/abs/path/vmlinux", not "vmlinux"
```

---

## 3) Different user, groups, or no privileges
Running under another user drops access to files/devices.
Missing group membership like `kvm` (for `/dev/kvm`) or `tun`.

### Triage
```bash
id -a
ls -l /dev/kvm /dev/net/tun /dev/vhost-vsock 2>/dev/null
getfacl /dev/kvm 2>/dev/null
sudo usermod -aG kvm "$USER"
```

---

## 4) File descriptors & stdio weirdness
Subprocesses may lack a TTY or have closed stdin.
If the tool expects to read until EOF and the parent never closes stdin, it can hang or bail.

### Triage
```bash
ls -l /proc/$$/fd
# If using pipelines:
stdbuf -oL -eL <command>
```

---

## 5) ulimit / resource limits differ
`RLIMIT_NOFILE`, `RLIMIT_NPROC`, `RLIMIT_MEMLOCK`, etc.
Systemd units and containers often impose stricter limits.

### Triage
```bash
ulimit -a
cat /proc/$$/limits
```

---

## 6) Security frameworks blocking syscalls or devices
**Seccomp**, **AppArmor**, **SELinux**, **NoNewPrivileges**.

### Triage
```bash
aa-status 2>/dev/null | sed -n '1,80p'
getenforce 2>/dev/null
dmesg | tail -n 200 | egrep -i 'apparmor|selinux|seccomp|denied'
```

---

## 7) Capabilities missing
Firecracker/jailer may need `CAP_SYS_ADMIN`, `CAP_NET_ADMIN`, `CAP_MKNOD`, etc.

### Triage
```bash
capsh --print | sed -n '1,80p'
```

### Docker example
```bash
docker run --rm -it --device /dev/kvm \
  --cap-add NET_ADMIN --cap-add SYS_ADMIN \
  --security-opt seccomp=unconfined --security-opt apparmor=unconfined \
  -v /dev/kvm:/dev/kvm <image> firecracker ...
```

### systemd example
```
CapabilityBoundingSet=CAP_SYS_ADMIN CAP_NET_ADMIN CAP_MKNOD CAP_SETPCAP
AmbientCapabilities=CAP_SYS_ADMIN CAP_NET_ADMIN CAP_MKNOD
NoNewPrivileges=false
DeviceAllow=/dev/kvm rwm
PrivateDevices=false
ProtectKernelTunables=false
ProtectControlGroups=false
```

---

## 8) Mount options: `noexec`, `nodev`, `nosuid`
If binary or interpreter is on a `noexec` mount, it can’t run.

### Triage
```bash
mount | egrep ' on /(tmp|home|mnt|opt) '
sudo mount -o remount,exec /path
```

---

## 9) Dynamic linker / library path differences
`LD_LIBRARY_PATH` may differ or missing libs in minimal container.

### Triage
```bash
ldd "$(command -v firecracker)" || echo "static? (ok)"
strace -f -e file,openat,execve -o /tmp/trace.txt your-launcher ...
grep -i 'ENOENT.*\.so' /tmp/trace.txt
```

---

## 10) Systemd sandboxing flags
Flags like `PrivateTmp`, `ProtectHome`, or `DynamicUser` can isolate or break permissions.

### Triage
```bash
systemctl cat your-service.service
systemd-analyze security your-service.service
```

---

## 11) Cgroups throttling / OOM kill
Memory or CPU limits can kill child processes silently.

### Triage
```bash
cat /proc/self/cgroup
dmesg | egrep -i 'killed process|oom' | tail
```

---

## 12) Signals & session semantics
Parent may send SIGHUP/SIGINT or pipeline SIGPIPEs.

### Triage
```bash
trap -p
nohup your-cmd </dev/null >/var/log/your.log 2>&1 &
```

---

## 13) Stale sockets, ports, or files
Firecracker API socket may already exist or be owned by another user.

### Triage
```bash
lsof -n | egrep 'firecracker|api|socket'
rm -f /path/to/api.socket
```

---

## 14) Line endings / encoding / corrupted artifacts
Windows CRLF or compressed kernel image cause “InvalidElfMagicNumber”.

### Triage
```bash
file microvm.json vmlinux* bzImage* rootfs*
dos2unix microvm.json
```

---

## 15) Firecracker-specific landmines
- `/dev/kvm` missing or denied by AppArmor
- `PrivateDevices=true` hides `/dev/kvm`
- `bzImage` vs ELF confusion
- relative paths under `jailer`
- missing capabilities for tap/vsock

### Triage
```bash
# Devices
ls -l /dev/kvm /dev/vhost-vsock /dev/net/tun

# Validate KVM
kvm-ok 2>/dev/null || { egrep -i 'vmx|svm' /proc/cpuinfo | head; lsmod | egrep kvm; dmesg | egrep -i kvm; }

# Validate artifacts
file vmlinux.bin bzImage kernel* rootfs*

# Ensure absolute paths
jq . microvm.json

# Trace the run
RUST_LOG=debug strace -f -yy -o /tmp/fc.strace \
  firecracker --no-api --config-file /abs/microvm.json 2>&1 | tee /tmp/fc.log
grep -E 'ENOENT|EACCES|EPERM|InvalidElf|Operation not permitted' /tmp/fc.strace /tmp/fc.log
```

---

## 16) Filesystem permissions/ownership drift
Parent creates files owned by root, subprocess runs unprivileged.

### Triage
```bash
umask
namei -l /path/to/*
```

---

## 17) Architecture / CPU feature mismatches
Binary built for wrong arch or CPU flags unavailable inside sandbox.

### Triage
```bash
file "$(command -v firecracker)"
cat /proc/cpuinfo | egrep -i 'vmx|svm|sse|avx'
```

---

## 18) Kernel lockdown / Secure Boot
Secure Boot lockdown may deny `/dev/kvm` or module access.

### Triage
```bash
dmesg | egrep -i 'Lockdown|secureboot'
mokutil --sb-state 2>/dev/null
```

---

# Quick Fix Checklist
1. Diff env + cwd  
2. Ensure absolute paths in JSON  
3. Check `/dev/kvm` and group perms  
4. Add needed capabilities or disable sandbox temporarily  
5. `strace` the failing subprocess  
6. Check AppArmor/SELinux logs  
7. Verify ulimits and cgroup quotas  
8. Confirm artifacts are valid (`file`, `dos2unix`, etc.)

---

# Example debug recipe
```bash
strace -f -yy -o /tmp/fc.strace firecracker --config-file /abs/microvm.json
grep -E 'EACCES|EPERM|ENOENT' /tmp/fc.strace | tail
```

That’ll tell you exactly *why* it dies. Once you see which syscall failed, the fix is almost always one of the 18 reasons above.
