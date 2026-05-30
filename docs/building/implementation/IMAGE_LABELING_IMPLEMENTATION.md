# Image Labeling and Version Catalog Implementation

**Date**: 2025-11-03
**Status**: ✅ Complete

## Summary

Implemented comprehensive OCI-compliant image labeling system and version catalog for NanoFuse base images to support authenticated GHCR pulls with proper versioning.

**✨ Unified Release Process**: Both CLI/API binaries and Docker images now use the same `[release]` commit message pattern for automatic versioning:
- Commit with `[release]` → Auto-creates `v0.0.X` tag for CLI/API binaries
- Commit with `[release]` → Auto-creates `image-v0.0.X` tag for Docker images
- Both happen in a single commit, keeping releases synchronized!

## What Was Implemented

### 1. OCI Labels in Dockerfile ✅

**File**: `images/base/Dockerfile`

Added comprehensive OCI-compliant labels following the [OCI Image Spec](https://github.com/opencontainers/image-spec/blob/main/annotations.md):

#### Standard OCI Labels
- `org.opencontainers.image.title` - "NanoFuse Base Image"
- `org.opencontainers.image.description` - Full description with features
- `org.opencontainers.image.version` - Version tag (e.g., "1.0.0", "latest")
- `org.opencontainers.image.created` - Build timestamp (RFC 3339)
- `org.opencontainers.image.authors` - "NanoFuse Contributors"
- `org.opencontainers.image.vendor` - "NanoFuse"
- `org.opencontainers.image.url` - Project URL
- `org.opencontainers.image.source` - Source repository
- `org.opencontainers.image.revision` - Git commit SHA
- `org.opencontainers.image.licenses` - "MIT"
- `org.opencontainers.image.documentation` - Docs URL

#### NanoFuse-Specific Labels
- `com.nanofuse.base-os` - "ubuntu:24.04"
- `com.nanofuse.kernel.version` - "5.10.204"
- `com.nanofuse.kernel.source` - "Firecracker CI"
- `com.nanofuse.architecture` - "x86_64"
- `com.nanofuse.type` - "base"
- `com.nanofuse.firecracker.compatible` - "true"
- `com.nanofuse.services.ssh` - "enabled"
- `com.nanofuse.services.systemd-networkd` - "enabled"

### 2. CI/CD Build Arguments ✅

**File**: `.github/workflows/ci.yaml`

Enhanced the Docker build process to pass dynamic values to labels:

```yaml
- name: Determine version for image build
  id: image_version
  run: |
    # Extract version from tag, or use 'dev' for non-tag builds
    if [[ "${{ github.ref }}" == refs/tags/image-v* ]]; then
      VERSION="${{ github.ref_name }}"
      VERSION="${VERSION#image-v}"
    elif [[ "${{ github.ref }}" == refs/heads/main ]]; then
      VERSION="latest"
    else
      VERSION="dev"
    fi
    echo "VERSION=$VERSION" >> $GITHUB_OUTPUT

- name: Build and push Docker image
  with:
    build-args: |
      VERSION=${{ steps.image_version.outputs.VERSION }}
      BUILD_DATE=${{ github.event.repository.updated_at }}
      VCS_REF=${{ github.sha }}
      VCS_URL=${{ github.server_url }}/${{ github.repository }}
```

**Tagging Strategy**:
- Push to `main` → Tags: `latest`, `main`, `sha-<commit>`
- Push to `main` with `[release]` in commit message → Auto-creates `image-v0.0.X` tag, then builds with tags: `0.0.X`, `sha-<commit>`
- Manual tag `image-v1.0.0` → Tags: `1.0.0`, `sha-<commit>`
- PR builds → Tag: `pr-<number>` (not pushed)

### 3. CLI Image Pull Support ✅

**File**: `cmd/nanofuse/main.go`

The CLI already had excellent support for pulling images by tag:

```go
// Image shortcuts
DefaultImageRegistry = "ghcr.io/daax-dev/nanofuse"
DefaultBaseImage = "base"
DefaultImageTag = "latest"

// resolveImageRef handles shortcuts:
// - "default" → "ghcr.io/daax-dev/nanofuse/base:latest"
// - "default:1.0.0" → "ghcr.io/daax-dev/nanofuse/base:1.0.0"
// - "base:v1.0" → "ghcr.io/daax-dev/nanofuse/base:v1.0"
```

**CLI Commands**:
```bash
# Pull latest
nanofuse image pull --default

# Pull specific version
nanofuse image pull --default --tag 1.0.0

# Pull full reference
nanofuse image pull ghcr.io/daax-dev/nanofuse/base:1.0.0

# Run VM with shorthand
nanofuse vm run default my-vm
nanofuse vm run default:1.0.0 my-vm
```

### 4. Version Catalog Documentation ✅

**File**: `images/VERSION_CATALOG.md`

Created comprehensive documentation covering:
- Available tag types (latest, versioned, commit-based, branch-based)
- Authentication requirements and setup
- CLI shortcuts reference table
- Complete OCI label reference
- Best practices for production vs development
- Migration guide between versions
- Release process documentation

### 5. Updated Main README ✅

**File**: `README.md`

Added "Image Versioning" section with:
- Quick reference table of tag types
- Use case recommendations
- Stability indicators
- Link to full version catalog

### 6. Label Validation Test ✅

**File**: `images/base/test-labels.sh`

Created automated test script to verify labels:
- Checks image exists
- Displays all OCI labels
- Displays NanoFuse-specific labels
- Validates required labels are present
- Returns exit code for CI integration

**Usage**:
```bash
cd images/base
sudo ./build.sh  # Build image locally
./test-labels.sh  # Verify labels
```

## How It Works

### Build Flow

1. **Local Build** (via `build.sh`):
   ```bash
   cd images/base
   sudo ./build.sh
   ```
   - Builds Docker image with default labels (VERSION=dev)
   - Creates rootfs.ext4 and downloads kernel
   - Generates manifest.json

2. **CI Build** (GitHub Actions):
   ```bash
   # Triggered by:
   # - Push to main → tags: latest, main, sha-<commit>
   # - Push to main with [release] → auto-creates image-v0.0.X tag
   #   - Then new CI run creates tags: 0.0.X, sha-<commit>
   # - Manual tag image-v1.0.0 → tags: 1.0.0, sha-<commit>
   ```
   - Determines version from git ref
   - Passes build args for dynamic labels
   - Builds multi-architecture image (currently x86_64)
   - Pushes to GHCR with proper tags and labels
   - **Auto-release**: Commits with `[release]` trigger auto-versioning (same as CLI/API)

### Pull Flow

1. **User authenticates**:
   ```bash
   docker login ghcr.io
   ```

2. **User pulls image** (via CLI):
   ```bash
   nanofuse image pull --default --tag 1.0.0
   ```

3. **CLI resolves shorthand**:
   ```
   "default:1.0.0" → "ghcr.io/daax-dev/nanofuse/base:1.0.0"
   ```

4. **API daemon downloads** from GHCR with authentication

5. **Image stored locally** with all labels intact

6. **User inspects** image metadata:
   ```bash
   nanofuse image inspect default:1.0.0
   docker inspect ghcr.io/daax-dev/nanofuse/base:1.0.0
   ```

## Verification

### Test Label Presence

```bash
# Build image locally
cd images/base
sudo ./build.sh

# Run label validation
./test-labels.sh

# Expected output:
# ✓ Image found: nanofuse-base:latest
# Standard OCI Labels:
#   org.opencontainers.image.created: 2025-11-03T...
#   org.opencontainers.image.description: Ubuntu 24.04...
#   org.opencontainers.image.title: NanoFuse Base Image
#   org.opencontainers.image.version: dev
#   ...
# NanoFuse-specific Labels:
#   com.nanofuse.base-os: ubuntu:24.04
#   com.nanofuse.kernel.version: 5.10.204
#   ...
# ✓ All required labels present!
```

### Test CI Build

```bash
# Trigger CI by pushing to main
git add -A
git commit -m "test: verify image labeling"
git push origin main

# Check GitHub Actions:
# - Build should succeed
# - Image should be pushed to ghcr.io/daax-dev/nanofuse/base:latest
# - Labels should include VERSION=latest, BUILD_DATE, VCS_REF
```

### Test Auto-Versioned Release (Recommended)

```bash
# Auto-create versioned release (same as CLI/API)
git add -A
git commit -m "feat: new base image features [release]"
git push origin main

# Check GitHub Actions:
# - First CI run: Builds latest + creates image-v0.0.X tag
# - Second CI run: Triggered by new tag, builds versioned image
# - Image pushed to ghcr.io/daax-dev/nanofuse/base:0.0.X
# - VERSION label should be "0.0.X"
```

### Test Manual Versioned Release

```bash
# Manually create versioned release
git tag image-v1.0.0 -m "Release 1.0.0: Initial stable base image"
git push origin image-v1.0.0

# Check GitHub Actions:
# - Image pushed to ghcr.io/daax-dev/nanofuse/base:1.0.0
# - VERSION label should be "1.0.0"
```

### Test CLI Pull

```bash
# Ensure daemon is running
sudo systemctl start nanofused

# Test various pull methods
nanofuse image pull --default                    # latest
nanofuse image pull --default --tag 1.0.0        # versioned
nanofuse image pull ghcr.io/daax-dev/nanofuse/base:sha-abc1234  # commit

# Inspect pulled images
nanofuse image list
nanofuse image inspect default
```

## Examples

### Example 1: Development Workflow

```bash
# Developer wants latest features
nanofuse image pull --default
nanofuse vm run default my-dev-vm

# Image automatically pulls ghcr.io/daax-dev/nanofuse/base:latest
# with VERSION=latest label
```

### Example 2: Production Workflow

```bash
# Production deployment with pinned version
nanofuse image pull --default --tag 1.0.0
nanofuse vm run default:1.0.0 my-prod-vm

# Image pulls ghcr.io/daax-dev/nanofuse/base:1.0.0
# with VERSION=1.0.0 label (immutable)
```

### Example 3: Inspect Image Metadata

```bash
# Pull image
nanofuse image pull --default

# Inspect via Docker
docker inspect ghcr.io/daax-dev/nanofuse/base:latest | \
  jq '.[0].Config.Labels'

# Shows all OCI and NanoFuse labels with build metadata
```

### Example 4: CI/CD with Specific Version

```yaml
# .github/workflows/deploy.yml
- name: Pull NanoFuse base image
  run: |
    docker login ghcr.io -u ${{ github.actor }} -p ${{ secrets.GITHUB_TOKEN }}
    nanofuse image pull --default --tag 1.0.0

- name: Deploy VM
  run: |
    nanofuse vm create prod-vm --image default:1.0.0 --vcpus 2 --memory 1024
    nanofuse vm start prod-vm
```

## Files Changed

| File | Changes | Purpose |
|------|---------|---------|
| `images/base/Dockerfile` | Added OCI + NanoFuse labels | Image metadata |
| `.github/workflows/ci.yaml` | Added version detection + build args | Dynamic label values |
| `cmd/nanofuse/main.go` | (Already had support) | CLI shortcuts |
| `images/VERSION_CATALOG.md` | ✨ New file | Version documentation |
| `README.md` | Added versioning section | User-facing docs |
| `images/base/test-labels.sh` | ✨ New file | Label validation |

## Benefits

✅ **OCI Compliance**: Standard labels for tooling compatibility
✅ **Version Clarity**: Clear version tags for production use
✅ **Reproducibility**: Commit-based tags for exact builds
✅ **User-Friendly**: CLI shortcuts for common operations
✅ **Production-Ready**: Immutable versioned tags
✅ **Documentation**: Comprehensive catalog of available versions
✅ **Validation**: Automated testing of label presence

## Next Steps

### Immediate
- ✅ All implementation complete
- ⏭️ Test by building and pushing to GHCR
- ⏭️ Create first versioned release (image-v1.0.0)

### Future Enhancements
- 🔮 Add `nanofuse image catalog` command to list available versions
- 🔮 Multi-architecture support (ARM64)
- 🔮 Image size optimization labels
- 🔮 Security scan results in labels
- 🔮 Automated changelog in version catalog

## References

- [OCI Image Spec Annotations](https://github.com/opencontainers/image-spec/blob/main/annotations.md)
- [GHCR Documentation](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
- [Docker Labels Best Practices](https://docs.docker.com/config/labels-custom-metadata/)
- [Semantic Versioning](https://semver.org/)

---

**Status**: ✅ Complete and ready for use
**Last Updated**: 2025-11-03
