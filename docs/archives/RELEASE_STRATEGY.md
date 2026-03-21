# Release Strategy

How to release CLI/API binaries and base images independently.

---

## TL;DR

- **CLI/API release**: Add `[release]` to commit message
- **Image release**: Automatically tagged as separate version (`image-v*`)
- **Both release together** with single commit message containing `[release]`

---

## How It Works

NanoFuse has **two independent release tracks**:

### Track 1: CLI/API Binaries (nanofuse + nanofused)

**Tag pattern**: `v0.0.1`, `v0.0.2`, `v1.0.0`, etc.

**Triggered by**: Commit message contains `[release]` on main branch

**What gets released**:
- `nanofuse` CLI binary
- `nanofused` daemon binary
- `nanofused.service` systemd file
- `setup-service.sh` script
- Release notes

**Version bumping**: Automatically increments patch version
- Latest tag: `v0.0.1`
- New commit with `[release]`: Creates `v0.0.2`

### Track 2: Base Image (Docker/ext4)

**Tag pattern**: `image-v0.0.1`, `image-v0.0.2`, `image-v1.0.0`, etc.

**Triggered by**: Commit message contains `[release]` on main branch

**What gets released**:
- Docker image built from `images/base/Dockerfile`
- Published to GHCR at `ghcr.io/jpoley/nanofuse/base:0.0.2`
- Tagged as `latest` on main
- Versioned tag for production use

**Version bumping**: Automatically increments patch version (separate from CLI)
- Latest image tag: `image-v0.0.1`
- New commit with `[release]`: Creates `image-v0.0.2`

---

## Release Scenarios

### Scenario 1: Release CLI/API Only (Bug Fix)

```bash
# Make changes to CLI or daemon code
git add cmd/nanofuse/ cmd/nanofused/
git commit -m "fix: handle edge case in VM creation [release]"
git push origin main
```

**What happens**:
1. CI runs build-go job (compiles binaries) ✅
2. Contains `[release]` → auto-release job runs
3. Creates git tag `v0.0.2` (auto-incremented)
4. Builds release binaries
5. Creates GitHub Release with binaries ✅
6. Docker image build also runs (untagged/dev only)

**Result**:
- ✅ New CLI/API release `v0.0.2`
- ✅ Downloadable from GitHub Releases
- ❌ Image NOT released (stays at `image-v0.0.1`)

### Scenario 2: Release Image Only (New Packages)

```bash
# Make changes to base image
git add images/base/Dockerfile
git commit -m "feat: add curl and network tools to base image [release]"
git push origin main
```

**What happens**:
1. CI runs build-image job ✅
2. Contains `[release]` → auto-release-image job runs
3. Creates git tag `image-v0.0.2` (auto-incremented)
4. This triggers CI again (on the new image-v* tag)
5. Docker image built and pushed to GHCR ✅
6. Pushed as `ghcr.io/jpoley/nanofuse/base:0.0.2`
7. Also tagged as `latest` ✅

**Result**:
- ✅ New image release `image-v0.0.2` in GHCR
- ❌ CLI/API NOT released (stays at `v0.0.1`)

### Scenario 3: Release Both (New Feature + Image Update)

```bash
# Make changes to both CLI and image
git add cmd/nanofuse/main.go images/base/Dockerfile
git commit -m "feat: support new image version and add image validation [release]"
git push origin main
```

**What happens**:
1. build-go job compiles CLI/API ✅
2. build-image job builds Docker image ✅
3. Both contain `[release]` → both auto-release jobs run
4. Creates CLI tag `v0.0.2`
5. Creates image tag `image-v0.0.2`
6. GitHub Release created with binaries ✅
7. Docker image pushed to GHCR ✅

**Result**:
- ✅ New CLI/API release `v0.0.2`
- ✅ New image release `image-v0.0.2`
- Both released simultaneously

---

## Commit Message Format

### Include [release] to Trigger Release

```bash
git commit -m "feat: add SSH key rotation [release]"
git commit -m "fix: handle timeout in VM creation [release]"
git commit -m "chore: update documentation [release]"
```

### Format Examples

**Good**:
```
fix: timeout handling in VM startup [release]

- Increase timeout from 5s to 10s
- Add debug logging
- Tests pass locally
```

**Also good**:
```
[release] feat: add image validation
```

**Not good** (no `[release]`):
```
feat: add SSH key rotation
# This will NOT trigger a release
```

### Case Sensitivity

`[release]` is case-sensitive. These will NOT trigger release:
- `[Release]` ❌
- `[RELEASE]` ❌
- `[Release]` ❌

---

## Version Numbering

Both CLI and image use **semantic versioning**: `MAJOR.MINOR.PATCH`

### CLI Version Examples

- `v0.0.1` - First release
- `v0.0.2` - Bug fixes
- `v0.1.0` - New features (minor)
- `v1.0.0` - Breaking changes (major)

### Image Version Examples

- `image-v0.0.1` - First base image
- `image-v0.0.2` - Security updates
- `image-v0.1.0` - Major changes
- `image-v1.0.0` - Full rewrite

### Manual Version Bumps

If you want to control versions manually, create tags directly:

```bash
# CLI release with specific version
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0

# Image release with specific version
git tag -a image-v0.1.0 -m "Release image-v0.1.0"
git push origin image-v0.1.0
```

---

## CI/CD Pipeline Behavior

### On Every Push to main

```
┌─────────────────────────────────────┐
│ Push to main (any commit)           │
└────────┬────────────────────────────┘
         │
         ├─→ build-go (compile binaries) ✅
         ├─→ lint (code quality) ✅
         ├─→ security-scan ✅
         ├─→ validate ✅
         ├─→ build-image (Docker build) ✅
         │
         └─→ If commit message contains [release]:
             ├─→ auto-release (CLI/API)
             │   ├─ Create tag v0.0.2
             │   ├─ GitHub Release
             │   └─ Publish binaries
             │
             └─→ auto-release-image (Base Image)
                 ├─ Create tag image-v0.0.2
                 ├─ Push to GHCR
                 └─ Tag as latest
```

### On New image-v* Tag

```
┌─────────────────────────────────────┐
│ New tag pushed: image-v0.0.2        │
└────────┬────────────────────────────┘
         │
         └─→ build-image job runs again
             └─ Builds and pushes versioned image
```

---

## Publishing & Availability

### CLI/API Binaries

After `[release]`:
1. Published to GitHub Releases page
2. Download with: `gh release download v0.0.2`
3. Available immediately

### Base Image

After `[release]`:
1. Published to GHCR: `ghcr.io/jpoley/nanofuse/base:0.0.2`
2. Also tagged as `latest`: `ghcr.io/jpoley/nanofuse/base:latest`
3. Pull with: `docker pull ghcr.io/jpoley/nanofuse/base:0.0.2`
4. Available in ~5-10 minutes (build time)

### Usage After Release

**CLI/API**:
```bash
gh release download v0.0.2 -R jpoley/nanofuse
chmod +x nanofuse nanofused
sudo mv nanofuse nanofused /usr/local/bin/
```

**Base Image**:
```bash
nanofuse image pull ghcr.io/jpoley/nanofuse/base:0.0.2
# or
nanofuse image pull --default  # Gets latest
```

---

## Common Tasks

### Release a Bug Fix (CLI Only)

```bash
# Make fix
nano cmd/nanofuse/vm.go
git add cmd/nanofuse/vm.go
git commit -m "fix: VM startup race condition [release]"
git push origin main

# CI automatically:
# 1. Creates v0.0.2 tag
# 2. Publishes GitHub Release
# 3. Builds image (but doesn't release)
```

### Release Image Updates

```bash
# Update base image
nano images/base/Dockerfile
git add images/base/Dockerfile
git commit -m "feat: add required system packages [release]"
git push origin main

# CI automatically:
# 1. Creates image-v0.0.2 tag
# 2. Pushes to GHCR
# 3. Updates latest tag
# 4. Builds CLI (but doesn't release)
```

### Release Both Simultaneously

```bash
# Update both
git add cmd/nanofuse/ images/base/Dockerfile
git commit -m "feat: new image format support [release]"
git push origin main

# CI automatically:
# 1. Creates v0.0.2 tag (CLI/API)
# 2. Creates image-v0.0.2 tag (image)
# 3. Publishes both
```

### Hotfix Without Release

```bash
# Make changes
git add .
git commit -m "docs: fix typo"
git push origin main

# No [release] in message:
# - CLI/API not released
# - Image not released
# - Just a normal commit
```

---

## Important Notes

### Only Works on main Branch

Release is triggered by:
```
github.ref == 'refs/heads/main' && contains(github.event.head_commit.message, '[release]')
```

- **Feature branches**: Won't trigger release even with `[release]`
- **Pull requests**: Won't trigger release
- **Only main**: Releases only on pushed commits to main

### Automatic Version Bumping

Versions are auto-incremented at patch level:
- Latest: `v0.0.1` → Next: `v0.0.2`
- Latest: `v1.2.3` → Next: `v1.2.4`

To bump minor/major, create tag manually:
```bash
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0
```

### GitHub Token Required

Releases need GITHUB_TOKEN (provided automatically by GitHub Actions).

If you manually trigger outside of GitHub, you need:
```bash
export GITHUB_TOKEN=ghp_xxxx
gh release create v0.0.2 --generate-notes
```

---

## Troubleshooting

### No release created even with [release]

1. **Check commit is on main branch**
   ```bash
   git log --oneline -1
   # Should show: on main, last commit has [release]
   ```

2. **Check commit message exactly**
   ```bash
   git log -1 --format="%B"
   # Must contain literal string: [release]
   ```

3. **Check CI status**
   - Go to GitHub Actions
   - View latest workflow run
   - Check build-go, lint, security-scan all pass
   - Check auto-release job exists

4. **Check if tag already exists**
   ```bash
   git tag -l 'v*' | sort -V | tail -1
   # If v0.0.2 already exists, next will be v0.0.3
   ```

### Image pushed but binaries not released

This happens if build-go fails but build-image passes.

- Check build-go job in GitHub Actions
- Fix the compilation error
- Make a new commit with `[release]`

### Can't find release in GitHub Releases

1. Check GitHub repository settings → Releases
2. Check the tag was created: `git tag`
3. Check Actions tab → auto-release job passed

---

## See Also

- [CI/CD Pipeline](.github/workflows/ci.yaml) - Full workflow definition
- [README.md](../README.md) - Installation instructions
- [DEVELOPMENT_GUIDE.md](DEVELOPMENT_GUIDE.md) - Development setup
