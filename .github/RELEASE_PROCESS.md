# Release Process

NanoFuse uses separate versioning for Go binaries and Docker images to allow independent release cycles.

## Overview

- **Go Binaries**: `v0.0.1`, `v0.0.2`, etc. → GitHub Releases with downloadable binaries
- **Docker Images**: `image-v0.0.1`, `image-v0.0.2`, etc. → GHCR images at `ghcr.io/jpoley/nanofuse/base`

## Go Binary Releases

Create a new Go binary release (CLI + daemon):

```bash
./release.sh
```

This will:
1. Find the latest `v*` tag (e.g., `v0.0.1`)
2. Bump the patch version (e.g., `v0.0.2`)
3. Update `magefile.go` with the new version
4. Commit, tag, and push to GitHub
5. GitHub Actions will:
   - Build binaries for Linux amd64
   - Run tests and security scans
   - Create a GitHub Release with binaries attached

**Manual version:**
```bash
git tag v0.1.0
git push origin v0.1.0
```

## Docker Image Releases

Create a new Docker image release:

```bash
./image-release.sh
```

This will:
1. Find the latest `image-v*` tag (e.g., `image-v0.0.1`)
2. Bump the patch version (e.g., `image-v0.0.2`)
3. Create and push the tag (no file changes needed)
4. GitHub Actions will:
   - Build the Docker image from `images/base/`
   - Tag it with the version number (e.g., `0.0.2`)
   - Push to `ghcr.io/jpoley/nanofuse/base:0.0.2`
   - Also update the `latest` tag if on main branch

**Manual version:**
```bash
git tag image-v0.1.0
git push origin image-v0.1.0
```

## Image Tags on GHCR

After `image-v0.0.5` is pushed, the following tags are available:

- `ghcr.io/jpoley/nanofuse/base:0.0.5` - Specific version
- `ghcr.io/jpoley/nanofuse/base:latest` - Latest from main branch
- `ghcr.io/jpoley/nanofuse/base:sha-abc1234` - Specific commit (for reproducibility)
- `ghcr.io/jpoley/nanofuse/base:main` - Latest main branch build

Users can pull with (after authenticating):
```bash
# Authenticate first (images are private)
docker login ghcr.io

# Latest
nanofuse image pull --default

# Specific version
nanofuse image pull --default --tag 0.0.5
nanofuse image pull ghcr.io/jpoley/nanofuse/base:0.0.5

# Specific commit
nanofuse image pull ghcr.io/jpoley/nanofuse/base:sha-abc1234
```

## Release Workflow

### Typical development cycle:

1. **Develop features** → Push to main
   - CI runs tests, builds, pushes `latest` image

2. **Ready for image release** → Run `./image-release.sh`
   - Creates `image-v0.0.X` tag
   - CI builds and tags versioned image
   - Users can pull with `--tag 0.0.X`

3. **Ready for Go binary release** → Run `./release.sh`
   - Creates `v0.0.X` tag
   - CI builds binaries and creates GitHub Release
   - Users can download from Releases page

### Example:

```bash
# Day 1: Feature work
git commit -m "feat: add port forwarding"
git push origin main
# → CI builds image as ghcr.io/jpoley/nanofuse/base:latest

# Day 2: Image is stable, create versioned release
./image-release.sh
# → Creates image-v0.0.1
# → CI builds ghcr.io/jpoley/nanofuse/base:0.0.1

# Day 5: More features
git commit -m "feat: add snapshot support"
git push origin main
# → CI updates ghcr.io/jpoley/nanofuse/base:latest

# Day 6: Ready for binary release
./release.sh
# → Creates v0.0.1
# → CI creates GitHub Release with nanofuse/nanofused binaries

# Day 10: New image features tested
./image-release.sh
# → Creates image-v0.0.2
# → CI builds ghcr.io/jpoley/nanofuse/base:0.0.2
```

## Version Bumping

Both scripts support bump types via arguments (planned for future):

```bash
# Patch bump (default): 0.0.1 → 0.0.2
./release.sh

# Minor bump: 0.0.2 → 0.1.0
./release.sh minor

# Major bump: 0.1.0 → 1.0.0
./release.sh major
```

Currently only patch bumps are implemented. Edit `release.sh` or `image-release.sh` to manually set major/minor versions.

## CI/CD Behavior

| Event | Go Binaries | Docker Images |
|-------|-------------|---------------|
| Push to main | Build, test | Build, push as `latest` + `sha-*` |
| Tag `v*` | Build, release on GitHub | Skip |
| Tag `image-v*` | Skip | Build, push as version + `latest` |
| Pull Request | Build, test | Build (no push) |

## Troubleshooting

### Image not appearing on GHCR

1. Check workflow ran: https://github.com/jpoley/nanofuse/actions
2. Ensure package is public: See `.github/GHCR_SETUP.md`
3. Check tag format: Must be exactly `image-vX.Y.Z`

### Release failed

1. Check workflow logs in GitHub Actions
2. Ensure all tests pass
3. Check permissions: Workflow needs `packages: write` and `contents: write`

### Wrong version number

If you need to re-release:
```bash
# Delete local tag
git tag -d v0.0.5

# Delete remote tag
git push origin :refs/tags/v0.0.5

# Re-run release script or manually create new tag
```

## References

- GitHub Actions workflow: `.github/workflows/ci.yaml`
- GHCR setup guide: `.github/GHCR_SETUP.md`
