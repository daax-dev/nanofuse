# Decision Log: Layer Composition and Boot Test

**Date:** 2025-12-28
**Author:** Claude (automated)
**Status:** SUCCESS

## Summary

Successfully implemented layer-based image composition for NanoFuse microVMs:
1. Created base-os layer with Ubuntu 24.04 + systemd
2. Created image composition script
3. Composed test image from multiple layers
4. Successfully booted in Firecracker

## Decisions Made

### 1. Base OS Layer Design

**Decision:** Use Ubuntu 24.04 with systemd as the base layer.

**Rationale:**
- Ubuntu 24.04 is LTS with support until 2029
- systemd provides modern init, service management, and networking
- Consistent with falcondev devcontainer base

**Included packages:**
- systemd, systemd-sysv, dbus (init system)
- iproute2, iputils-ping, netplan.io (networking)
- openssh-server (debugging access)
- ca-certificates (HTTPS support)

**Masked services:** systemd-resolved, networkd-wait-online, plymouth, etc.

### 2. Image Composition Approach

**Decision:** Use mke2fs with -d option instead of loop mount.

**Rationale:**
- Avoids requiring sudo/root privileges
- Works in non-interactive terminals
- Creates ext4 image directly from directory tree
- Simpler error handling

**Alternative considered:** Loop mount with sudo - rejected due to permission issues in CI/CD.

### 3. Kernel Selection

**Decision:** Use 5.10.245 kernel for production, 6.1.155 for development.

**Rationale:**
- 5.10.245 has virtio-blk support built-in
- 6.1.155 is missing virtio-blk (modular, not loaded)
- Both are LTS kernels with long support

**Future work:** Build custom 6.1 kernel with virtio-blk built-in.

### 4. Layer Stacking Order

**Decision:** Layers are applied in manifest order (base → runtime → application).

**Rationale:**
- Simple rsync merge strategy
- Later layers override earlier files
- Dependencies enforced by manifest order

## Results

### Built Layers

| Layer | Size | SHA256 |
|-------|------|--------|
| base-os | 185MB | e7547db93f2e6023594b887ae0dbbaa06858dfcdccb5c78e764aaea1acdbc029 |
| python-runtime | 159MB | 3844d42524668ae631a344d3b6eab3e78eda0e4df820890019b398b6107343a7 |
| node-runtime | 466MB | bc42f39e7ced1531e17390422d0e986f73a06eab0d45295b33ac531a6ed4bf00 |
| go-runtime | 87MB | 6d817e5f0d5f305f55c370671d6eb1515d66a7e8fbd61a91c7a873a1f905eab1 |
| recording-agent | 81MB | b65eaffc954ec6c4e635d918a12f298d02fad9b75e6cfacf26f768c11e36dc67 |
| agent-tools | 132MB | 2af888fde9475643a954fc4b77ca2d889b3d12d436c071ae5278c08382cbf8ed |

### Composed Images

| Image | Size | Layers |
|-------|------|--------|
| test-boot.ext4 | 631MB | base-os, python-runtime, node-runtime |
| falcondev-agents.ext4 | 706MB | base-os, python-runtime, node-runtime, recording-agent, agent-tools |

### Boot Test Results - test-boot.ext4

- **Kernel:** 5.10.245-no-acpi
- **Boot time:** ~1 second
- **Services started:** systemd, journald, networkd, serial-getty
- **Serial console:** Working (with minor buffer overflow warnings)
- **Root filesystem:** Mounted successfully

### Boot Test Results - falcondev-agents.ext4

- **Kernel:** 5.10.245-no-acpi
- **Boot time:** ~1.5 seconds
- **Services started:** systemd, journald, networkd, serial-getty, dbus, logind, record-agent
- **Record agent:** ✅ Started successfully
- **SSH service:** ❌ Failed (missing host keys - needs first-boot generation)
- **Serial console:** Working (login prompt visible)
- **Multi-user target:** Reached successfully

## Files Created

```
layers/
├── base-os/
│   ├── Dockerfile
│   ├── layer.yaml
│   ├── rootfs/
│   └── base-os.tar.gz
├── python-runtime/
│   ├── Dockerfile
│   ├── layer.yaml
│   ├── hooks/post-install.sh
│   ├── rootfs/
│   └── python-runtime.tar.gz
├── node-runtime/
│   ├── Dockerfile
│   ├── layer.yaml
│   ├── hooks/post-install.sh
│   ├── rootfs/
│   └── node-runtime.tar.gz
├── go-runtime/
│   ├── Dockerfile
│   ├── layer.yaml
│   ├── rootfs/
│   └── go-runtime.tar.gz
├── recording-agent/
│   ├── Dockerfile
│   ├── layer.yaml
│   ├── hooks/post-install.sh
│   ├── rootfs/
│   └── recording-agent.tar.gz
└── agent-tools/
    ├── Dockerfile
    ├── layer.yaml
    ├── hooks/post-install.sh
    ├── rootfs/
    └── agent-tools.tar.gz

scripts/
├── compose-image.sh
├── build-layer.sh
└── test-boot.sh

images/
├── test-boot/
│   └── image.manifest.yaml
└── falcondev-agents/
    └── image.manifest.yaml

build/
├── test-boot.ext4
└── falcondev-agents.ext4
```

## Next Steps

1. ✅ ~~Create recording-agent layer with vsock support~~
2. ✅ ~~Build agent-tools layer with Claude Code CLI~~
3. ✅ ~~Compose full falcondev-agents image~~
4. ✅ ~~Fix SSH service failure~~ (switched to socket activation)
5. ✅ ~~Fix file ownership issues~~ (use fakeroot in composition)
6. Test with actual agent workloads (Claude Code in VM)
7. Build custom 6.1 kernel with virtio-blk built-in

## Issues Resolved (2025-12-29)

### 1. File Ownership Issue (FIXED)

**Problem:** Files in composed image owned by UID 1000 instead of root (0).

**Root cause:** `mke2fs -d` copies files with real UID/GID from source directory.

**Fix:** Use `fakeroot` to track fake root ownership during rsync and mke2fs:
```bash
fakeroot -s "$fakeroot_state" -i "$fakeroot_state" -- rsync -a ...
fakeroot -i "$fakeroot_state" -- mke2fs -t ext4 -d ...
```

### 2. SSH Service Failure (FIXED)

**Problem:** `ssh.service` failed to start on boot.

**Root cause:** Both `ssh.service` AND `ssh.socket` were enabled, causing port conflict.
Ubuntu 24.04 uses socket activation by default.

**Fix:** Enable only `ssh.socket` in base-os Dockerfile:
```dockerfile
RUN systemctl enable ssh.socket  # NOT ssh.service
```

## Remaining Minor Issues

1. **Serial buffer overflow warnings** - Cosmetic, doesn't affect functionality
2. **Missing kernel modules** - `autofs4`, `unix` not found (non-critical)

## Related Tasks

- task-43 (IMG-001): falcondev-agents manifest ✅ COMPLETE
- task-45 (IMG-003): runtime layers ✅ COMPLETE
- task-31 (T008): recording-agent layer ✅ COMPLETE
