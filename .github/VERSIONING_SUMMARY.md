# Versioning and Release Summary

## What We Built

A dual-track versioning system for NanoFuse that separates Go binary releases from Docker image releases.

## Version Tracks

### 1. Go Binaries (CLI + Daemon)
- **Tags**: `v0.0.1`, `v0.0.2`, `v0.0.3`, etc.
- **Script**: `./release.sh`
- **Artifacts**: GitHub Releases with downloadable binaries
- **Use case**: When users want to download/install the CLI tools

### 2. Docker Images
- **Tags**: `image-v0.0.1`, `image-v0.0.2`, etc.
- **Script**: `./image-release.sh`
- **Artifacts**: GHCR images at `ghcr.io/jpoley/nanofuse/base:X.Y.Z`
- **Use case**: When base image is stable and ready for versioned release

## Files Created/Modified

### New Files:
1. **`release.sh`** - Auto-bump Go binary versions (v0.0.X)
2. **`image-release.sh`** - Auto-bump Docker image versions (image-v0.0.X)
3. **`.github/RELEASE_PROCESS.md`** - Complete release workflow documentation
4. **`.github/VERSIONING_SUMMARY.md`** - This file

### Modified Files:
1. **`.github/workflows/ci.yaml`**:
   - Added `image-v*` tag support
   - Separate build logic for `v*` vs `image-v*` tags
   - Updated metadata tagging for images
   - Binary releases only trigger on `v*` tags (not `image-v*`)

2. **`magefile.go`**:
   - Added `InstallUser()` target for ~/bin installation

3. **`README.md`**:
   - Added release workflow documentation
   - Updated GHCR tags section
   - Added ~/bin installation instructions

4. **`.gitignore`**:
   - Added `/nanofuse` and `/nanofused` for root binaries

5. **`cmd/nanofuse/main.go`**:
   - Added `--default` flag for image pulls
   - Added image reference shortcuts (default, base, etc.)
   - Enhanced authentication error messages

## How It Works

### Scenario 1: Release a new Docker image

```bash
# You've pushed several features to main
git push origin main

# Image builds look good, ready to version it
./image-release.sh
```

**What happens:**
1. Script finds latest `image-v*` tag (e.g., `image-v0.0.5`)
2. Bumps to `image-v0.0.6`
3. Creates and pushes tag
4. GitHub Actions:
   - Builds Docker image from `images/base/`
   - Tags as `ghcr.io/jpoley/nanofuse/base:0.0.6`
   - Also updates `latest` tag
5. Users can now: `nanofuse image pull --default --tag 0.0.6`

### Scenario 2: Release new Go binaries

```bash
# Ready for a new CLI release
./release.sh
```

**What happens:**
1. Script finds latest `v*` tag (e.g., `v0.0.3`)
2. Bumps to `v0.0.4`
3. Updates `magefile.go` version
4. Commits, tags, pushes
5. GitHub Actions:
   - Builds nanofuse + nanofused binaries
   - Creates GitHub Release with binaries
   - Users can download from Releases page

## CLI Improvements

### Before:
```bash
nanofuse image pull ghcr.io/jpoley/nanofuse/base:latest
nanofuse vm run ghcr.io/jpoley/nanofuse/base:latest my-vm
```

### After:
```bash
# Pull default image
nanofuse image pull --default
nanofuse vm run default my-vm

# Pull specific version
nanofuse image pull --default --tag 0.0.1
nanofuse vm run default:0.0.1 my-vm

# Shortcuts work
nanofuse vm run base my-vm           # → ghcr.io/jpoley/nanofuse/base:latest
nanofuse vm run default:0.0.5 my-vm  # → ghcr.io/jpoley/nanofuse/base:0.0.5
```

## Testing Plan

### 1. Test release.sh (dry run)
```bash
# Check what it would do without pushing
git fetch --tags
./release.sh 2>&1 | tee release-test.log
# Then reset: git reset --hard HEAD~1 && git tag -d vX.Y.Z
```

### 2. Test image-release.sh (dry run)
```bash
# Check tag creation (won't affect main)
./image-release.sh 2>&1 | tee image-release-test.log
# Then delete: git tag -d image-vX.Y.Z && git push origin :image-vX.Y.Z
```

### 3. Test CI workflow
```bash
# Push a test image tag
git tag image-v0.0.1
git push origin image-v0.0.1
# Watch: https://github.com/jpoley/nanofuse/actions
# Check GHCR: https://github.com/jpoley/nanofuse/pkgs/container/nanofuse%2Fbase
```

### 4. Test CLI shortcuts
```bash
# Authenticate first (images are private)
docker login ghcr.io

# After image is pushed
nanofuse image pull --default --tag 0.0.1
nanofuse vm run default:0.0.1 test-vm
```

## Benefits

1. **Independent versioning**: Can release images without binary releases
2. **Auto-incrementing**: No manual version number management
3. **Clear separation**: Easy to see what changed (binary vs image)
4. **User-friendly**: `--default` flag and shortcuts make CLI easier
5. **Reproducible**: Every version is tagged and available

## Migration Notes

- Existing `v*` tags remain for Go binaries
- New `image-v*` namespace for Docker images
- No breaking changes to existing workflows
- Users can still use full image references if preferred

## Future Enhancements

1. Support major/minor bumps: `./release.sh minor`
2. Auto-changelog generation
3. Slack/Discord notifications on releases
4. Multi-arch images (amd64 + arm64)
5. Release candidates: `image-v0.1.0-rc1`
