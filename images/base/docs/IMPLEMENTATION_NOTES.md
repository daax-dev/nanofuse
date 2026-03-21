# NanoFuse Base Image Implementation Notes

## Phase 1A Completion Status

### Deliverables Status

- ✅ `Dockerfile` - Complete base image definition
- ✅ `Makefile` - Build automation with all required targets
- ✅ `README.md` - Comprehensive documentation
- ✅ `test-boot.sh` - Firecracker boot testing script
- ✅ `validate-build.sh` - Build artifact validation
- ✅ `units/firstboot.service` - First-boot systemd unit
- ✅ `.gitignore` - Ignore build artifacts

### Design Decisions

#### 1. Kernel Selection

**Decision**: Use official Firecracker CI kernel 5.10.204 instead of Slicer's kernel

**Rationale**:
- Official Firecracker kernel ensures maximum compatibility
- Publicly accessible from AWS S3 (no authentication required)
- Well-tested by Firecracker CI pipeline
- Version 5.10.204 is production-ready and stable

**Source**: `https://s3.amazonaws.com/spec.ccfc.min/firecracker-ci/v1.7/x86_64/vmlinux-5.10.204`

**Alternative Considered**:
- Slicer kernel 5.10.240: GitHub release URL may change, authentication issues with GHCR
- Custom kernel build: Deferred to future learning phase (too complex for MVP)

#### 2. Rootfs Size

**Decision**: Default to 2GB ext4 filesystem

**Rationale**:
- Large enough for Ubuntu 24.04 + systemd + SSH + networking (~500-800MB)
- Room for user applications and data
- Can be customized via `ROOTFS_SIZE` make variable
- Balance between disk usage and flexibility

#### 3. systemd Configuration

**Decision**: Mask unnecessary systemd units, enable only essential services

**Rationale**:
- Reduces boot time by avoiding unnecessary service initialization
- Prevents conflicts in Firecracker environment (e.g., systemd-resolved, systemd-logind)
- Essential services only: ssh, systemd-networkd, serial-getty@ttyS0

**Masked Units**:
- `systemd-resolved.service` - DNS resolution (not needed in microVM)
- `systemd-timesyncd.service` - Time sync (host handles this)
- `systemd-logind.service` - Login manager (not needed for headless VM)
- `getty@.service` - Virtual terminal (using serial console instead)

#### 4. Network Configuration

**Decision**: Use systemd-networkd with DHCP

**Rationale**:
- Lightweight and built-in to systemd
- DHCP works with Firecracker's TAP networking
- No additional packages required (vs NetworkManager)
- Predictable configuration via `/etc/systemd/network/20-wired.network`

**Configuration**:
```
[Match]
Name=en*

[Network]
DHCP=yes
LinkLocalAddressing=yes
```

#### 5. SSH Configuration

**Decision**: Permit root login with key only, disable password authentication

**Rationale**:
- Security: No password-based access (prevents brute force)
- Convenience: SSH key injection via cloud-init or API (future)
- Standard practice for cloud/VM images

#### 6. Build Process

**Decision**: Multi-stage process: Docker build → Export → ext4 image creation

**Rationale**:
- Docker provides familiar build environment
- Dockerfile enables reproducible builds with layer caching
- ext4 export required for Firecracker (doesn't support Docker images directly)
- Separation of concerns: build vs runtime format

**Steps**:
1. Build Docker image with systemd and kernel
2. Create container from image
3. Export container filesystem to tarball
4. Create ext4 filesystem image
5. Mount and copy filesystem to ext4
6. Extract kernel separately
7. Generate manifest.json

#### 7. Manifest Format

**Decision**: JSON metadata file with image information

**Rationale**:
- Machine-readable format for API/CLI consumption
- Contains all necessary metadata (kernel cmdline, architecture, versions)
- Follows OCI image spec patterns
- Enables validation and introspection

**Fields**:
- `version`: Image version (semver)
- `name`: Image name
- `tag`: Image tag
- `architecture`: CPU architecture (x86_64, arm64)
- `base_os`: Base OS identifier
- `kernel`: Kernel metadata (version, source, cmdline, file)
- `rootfs`: Rootfs metadata (format, file, size)
- `services`: Enabled services
- `built_at`: Build timestamp (ISO 8601)

## Testing Strategy

### Automated Tests (test-boot.sh)

1. **Boot Test**: VM boots successfully within 30s timeout
2. **Console Test**: Console output visible on ttyS0
3. **systemd Test**: systemd reaches multi-user.target
4. **SSH Test**: SSH daemon running and listening
5. **Network Test**: systemd-networkd configured
6. **Units Test**: No failed systemd units
7. **Performance Test**: Boot time < 2s (warning if > 2s, acceptable if < 5s)

### Validation Tests (validate-build.sh)

1. **Rootfs Validation**: ext4 filesystem, correct size, integrity check
2. **Kernel Validation**: Uncompressed Linux kernel, correct size
3. **Manifest Validation**: Valid JSON, required fields present
4. **Docker Image Validation**: Image exists, correct size
5. **Contents Validation**: Required directories, systemd init, configs

### Manual Testing

Required for comprehensive validation:
- SSH connectivity test (inject key, connect)
- Network throughput test
- Snapshot/resume cycle test (future, with API)
- Multi-VM networking test (future, with bridge mode)

## Build Artifacts

### Generated Files

```
images/base/build/
├── rootfs.ext4           # 2GB ext4 filesystem (root device)
├── vmlinux               # Uncompressed kernel (5.10.204)
└── manifest.json         # Image metadata

Temporary Files (cleaned up):
├── .docker_built         # Docker build marker
├── .container_id         # Container ID for extraction
├── rootfs/               # Extracted filesystem (removed after ext4 creation)
└── mnt/                  # Mount point (removed after use)
```

### Sizes

- Docker image: ~300-400MB (with layers)
- rootfs.ext4: 2048MB (sparse file, actual usage ~500-800MB)
- vmlinux: ~30-40MB (uncompressed kernel)
- Total artifacts: ~2100MB

## Performance Targets

### Build Time

- Docker build: ~2-3 minutes (first build, depends on network)
- Docker build (cached): ~30 seconds
- Filesystem extraction: ~30 seconds
- ext4 creation: ~15 seconds
- **Total build time**: ~3-4 minutes (first), ~1-2 minutes (cached)

### Boot Time

- Target: < 2 seconds to multi-user.target
- Acceptable: < 5 seconds
- Includes: kernel boot, systemd init, service startup

## Known Issues and Limitations

### Build Issues

1. **Requires root for ext4 operations**: mkfs.ext4 and mount require root/sudo
   - Mitigation: Document requirement, provide sudo example in README
   - Future: Explore fakeroot or user namespace solutions

2. **Large build artifacts**: 2GB+ on disk
   - Mitigation: Make ROOTFS_SIZE configurable
   - Future: Implement sparse file handling, compression

### Runtime Issues

1. **No automatic SSH key injection**: Users must manually inject SSH keys
   - Mitigation: Document manual injection process
   - Future: Implement cloud-init or API-based key injection

2. **DHCP dependency**: Requires DHCP server on TAP network
   - Mitigation: API will configure DHCP for NAT mode
   - Future: Support static IP configuration

3. **No persistence across reboots**: Rootfs changes lost unless snapshot
   - Mitigation: This is expected behavior (ephemeral VMs)
   - Future: Support persistent disks via additional block devices

## Future Enhancements

### Phase 2 (Post-MVP)

1. **Cloud-init support**: Automatic SSH key injection, user-data
2. **Smaller kernel**: Compile custom minimal kernel
3. **Arm64 support**: Multi-arch builds and manifests
4. **Optimized boot**: Further reduce boot time (<1s)
5. **Read-only rootfs**: Immutable base with overlay

### Phase 3 (Production)

1. **Image signing**: Cryptographic verification
2. **Vulnerability scanning**: Automated CVE checks
3. **Minimal base**: Remove unnecessary packages
4. **Performance profiling**: systemd-analyze for boot optimization

### Phase 5+ (Advanced)

1. **Custom kernel**: Build from source with minimal config
2. **GPU support**: Cloud Hypervisor variant with PCI passthrough
3. **Confidential computing**: SEV-SNP support

## Coordination Notes

### For API Agent

The API agent can use these build artifacts for testing:

**Paths** (absolute):
- Rootfs: `/home/jpoley/src/_mine/nanofuse/images/base/build/rootfs.ext4`
- Kernel: `/home/jpoley/src/_mine/nanofuse/images/base/build/vmlinux`
- Manifest: `/home/jpoley/src/_mine/nanofuse/images/base/build/manifest.json`

**Firecracker Configuration** (example):
```json
{
  "boot-source": {
    "kernel_image_path": "/home/jpoley/src/_mine/nanofuse/images/base/build/vmlinux",
    "boot_args": "console=ttyS0 root=/dev/vda1 rw panic=1 reboot=k"
  },
  "drives": [
    {
      "drive_id": "rootfs",
      "path_on_host": "/home/jpoley/src/_mine/nanofuse/images/base/build/rootfs.ext4",
      "is_root_device": true,
      "is_read_only": false
    }
  ],
  "machine-config": {
    "vcpu_count": 2,
    "mem_size_mib": 512
  }
}
```

### For CLI Agent

The CLI will eventually pull images from GHCR, but for local testing:

**Local image reference** (mock):
- Image: `file:///home/jpoley/src/_mine/nanofuse/images/base/build`
- Metadata: Read from `manifest.json`

### For CI/CD Agent

Build process to automate in GitHub Actions:

1. Install Docker
2. Run `make build` (requires Docker in Docker or mounted socket)
3. Run `make validate` (optional, for quality gate)
4. Run `make test` (requires Firecracker + KVM, may need self-hosted runner)
5. Run `make push` (requires GITHUB_TOKEN for GHCR authentication)

**Challenges**:
- GitHub Actions runners don't have KVM by default (need self-hosted for Firecracker tests)
- Docker-in-Docker for image build
- Artifact size (2GB+) for workflow caching

## References

- [Firecracker Quickstart](https://github.com/firecracker-microvm/firecracker/blob/main/docs/getting-started.md)
- [systemd in Containers](https://systemd.io/CONTAINER_INTERFACE/)
- [Ubuntu Cloud Images](https://cloud-images.ubuntu.com/)
- [Slicer Documentation](https://docs.slicervm.com)

## Change Log

### 2025-10-30 - Initial Implementation
- Created Dockerfile FROM ubuntu:24.04
- Added Makefile with build, test, validate targets
- Implemented test-boot.sh for Firecracker boot testing
- Implemented validate-build.sh for artifact validation
- Created firstboot.service for first-boot initialization
- Documented in README.md
- Selected Firecracker CI kernel 5.10.204 (instead of Slicer kernel)
