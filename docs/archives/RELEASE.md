# Release Process

## TL;DR

Want to create a release? Just include the word **`release`** in your commit message:

```bash
git commit -m "feat: add new feature [release]"
git push origin main
```

CI will automatically:
1. ✅ Build and test
2. ✅ If all passes, create tag (e.g., `v0.0.2`)
3. ✅ Build release binaries
4. ✅ Publish to GitHub Releases

## How It Works

### Normal Commits (No Release)
```
Commit: "feat: add feature"
    ↓
Push to main
    ↓
CI: Build → Test → Lint → Security Scan
    ↓
Done (no tag, no release)
```

### Release Commits
```
Commit: "feat: add feature [release]"
    ↓
Push to main
    ↓
CI: Build → Test → Lint → Security Scan
    ↓
All passed ✅
    ↓
Auto-release job:
  - Create tag v0.0.X
  - Push tag to origin
  - Build binaries with tag version
  - Generate release notes
  - Publish GitHub Release
    ↓
Done! Release is live
```

## Release Examples

Any of these will trigger a release:

```bash
# Explicit release commit
git commit -m "release: version 0.1.0"

# Feature with release tag
git commit -m "feat: add snapshots [release]"

# Fix with release
git commit -m "fix: memory leak (release)"

# Anything with the word "release"
git commit -m "Ready for release!"
```

## Version Numbering

- **Automatic**: Patch version bumps automatically (0.0.1 → 0.0.2)
- **Manual**: To specify version, use release.sh (see below)

Current scheme: `v{major}.{minor}.{patch}`
- Patch: Auto-incremented for each release
- Minor/Major: Currently manual (edit ci.yaml if needed)

## Safety Checks

The auto-release job ONLY runs if:
- ✅ Commit pushed to `main` branch
- ✅ Commit message contains "release"
- ✅ Build job succeeded
- ✅ Lint job succeeded
- ✅ Security scan succeeded

If any check fails, no tag is created, no release happens.

## Manual Release (Alternative)

If you prefer manual control, use `release.sh`:

```bash
# After pushing to main and CI passes
./release.sh
```

This script:
- Verifies you're on main, synced with origin
- Checks CI passed on current commit
- Creates and pushes tag
- Triggers release workflow

## Versioning in Binaries

Built binaries get version from git tags:
- **With tags**: `nanofuse version` → `0.0.2`
- **No tags**: `nanofuse version` → `0.0.0-dev`

Version is set at build time via ldflags.

## Workflows

### CI Workflow (`.github/workflows/ci.yaml`)
**Triggers:** Push to main, Pull requests
**Jobs:**
1. Build and test Go binaries
2. Lint and format check
3. Security scanning
4. Auto-release (if commit has "release"):
   - Creates and pushes tag
   - Builds binaries with tag version
   - Generates release notes
   - Publishes GitHub Release

## Tips

### Batching Multiple Commits

Don't want every commit released? Just don't say "release":

```bash
git commit -m "feat: add feature A"
git push

git commit -m "feat: add feature B"
git push

git commit -m "fix: typo"
git push

# Now release all three together
git commit -m "release: v0.1.0 with features A, B and fixes" --allow-empty
git push
```

### Skip Release Despite "Release" in Description

If you're working on release-related code but don't want to trigger a release:

```bash
# This will NOT trigger (not in commit message)
git commit -m "refactor: update release script" -m "Details about the release process..."
```

Only the first line of the commit message (subject) is checked.

## Troubleshooting

### "Release didn't happen"

Check:
1. Did commit message contain "release"?
2. Did CI pass all checks?
3. Check Actions tab on GitHub for logs

### "Version is wrong"

The version comes from the most recent `v*` tag. Check:
```bash
git tag -l --sort=-v:refname 'v[0-9]*' | head -1
```

### "Want to delete a tag"

```bash
# Delete locally
git tag -d v0.0.2

# Delete on GitHub
git push origin :refs/tags/v0.0.2

# Also delete the GitHub Release manually via UI
```

## Architecture

```
Developer
    ↓
Commits with "release"
    ↓
GitHub Actions CI
    ├→ build-go
    ├→ lint
    ├→ security-scan
    └→ auto-release (needs: all above)
         ├→ Create tag v0.0.X
         ├→ Push tag to origin
         ├→ Build binaries with tag version
         ├→ Generate release notes
         └→ Publish GitHub Release
              ↓
         Release is live!
```

## Future Enhancements

Possible improvements:
- Support release types: `[release:major]`, `[release:minor]`
- Parse version from commit: `release: v0.2.0`
- Auto-generate changelog
- Pre-release tags: `[release:beta]`
