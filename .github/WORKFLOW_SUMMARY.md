# GitHub Actions Workflow Summary

This document provides a quick reference for all GitHub Actions workflows in NanoFuse.

## Workflows Overview

### 1. CI/CD Pipeline (`ci.yaml`)

**File**: `.github/workflows/ci.yaml`

**Purpose**: Main build, test, and release pipeline

**Triggers**:
- Push to `main` branch
- Pull requests to `main`
- Tags matching `v*` pattern

**Jobs**:

| Job | Purpose | Artifacts | Runs On |
|-----|---------|-----------|---------|
| `build-go` | Build Go binaries and run tests | `nanofuse-cli`, `nanofused-daemon` | All events |
| `build-image` | Build Docker image, push to GHCR | Docker image in GHCR | All events (push on main/tags) |
| `security-scan` | Vulnerability scanning | SARIF results in Security tab | All events |
| `lint` | Code quality checks | None | All events |
| `release` | Create GitHub release | Release with binaries | Tags only |

**Environment Variables**:
- `GO_VERSION`: `1.22`
- `REGISTRY`: `ghcr.io`
- `IMAGE_NAME`: `${{ github.repository }}`

**Secrets Used**:
- `GITHUB_TOKEN` (automatically provided)

**Publishing**:
- **Pull Requests**: Build only, no publishing
- **Main Branch**: Publish Docker image with `latest` and `sha-*` tags
- **Tags**: Publish Docker image + create GitHub release with binaries

---

### 2. PR Comment (`pr-comment.yaml`)

**File**: `.github/workflows/pr-comment.yaml`

**Purpose**: Post helpful information on pull requests

**Triggers**:
- Pull request opened
- Pull request synchronized (new commits)

**Actions**:
- Posts comment with build information
- Lists CI jobs that will run
- Explains publishing behavior
- Shows commit SHA

**Permissions**:
- `pull-requests: write`

---

### 3. Dependabot (`dependabot.yml`)

**File**: `.github/dependabot.yml`

**Purpose**: Automated dependency updates

**Update Schedule**:
- **Go modules**: Monday 09:00 UTC
- **Docker images**: Tuesday 09:00 UTC
- **GitHub Actions**: Wednesday 09:00 UTC

**PR Limits**:
- Go: 10 open PRs max
- Docker: 5 open PRs max
- Actions: 5 open PRs max

**Labels Applied**:
- `dependencies`
- `go` / `docker` / `github-actions`

---

## Workflow Execution Flow

### Pull Request Flow

```
PR Created/Updated
       ↓
   build-go ─────┐
                  ├─→ All jobs run in parallel
   build-image ──┤
                  │
   security-scan ┤
                  │
   lint ─────────┘
       ↓
   Results posted to PR
   (No publishing)
```

### Main Branch Push Flow

```
Push to main
       ↓
   build-go ─────┐
                  ├─→ Jobs run in parallel
   build-image ──┤   (image pushed to GHCR)
                  │
   security-scan ┤
                  │
   lint ─────────┘
       ↓
   Docker image published:
   - ghcr.io/jpoley/nanofuse/base:latest
   - ghcr.io/jpoley/nanofuse/base:sha-abc123
```

### Tag Release Flow

```
Tag v1.0.0 pushed
       ↓
   build-go ─────┐
                  ├─→ Jobs run in parallel
   build-image ──┤   (image pushed to GHCR)
                  │
   security-scan ┤
                  │
   lint ─────────┤
                  ↓
              release job
                  ↓
   GitHub Release Created:
   - Binaries: nanofuse, nanofused
   - Release notes generated
   - Docker image tagged: v1.0.0
```

---

## Status Badges

Add to README.md:

```markdown
[![CI/CD Pipeline](https://github.com/jpoley/nanofuse/actions/workflows/ci.yaml/badge.svg)](https://github.com/jpoley/nanofuse/actions/workflows/ci.yaml)
```

---

## Quick Reference Commands

### Trigger Workflows Locally

```bash
# Run CI checks locally
make ci

# Check build before pushing
./scripts/check-build.sh

# Verify CI configuration
./scripts/verify-ci.sh
```

### Manual Workflow Triggers

GitHub Actions workflows can be manually triggered from:
1. Go to "Actions" tab in GitHub
2. Select workflow on left sidebar
3. Click "Run workflow" button

### View Workflow Runs

- **All workflows**: `https://github.com/jpoley/nanofuse/actions`
- **Specific workflow**: `https://github.com/jpoley/nanofuse/actions/workflows/ci.yaml`
- **Specific run**: Click on run from actions page

### Download Artifacts

```bash
# Via GitHub CLI
gh run download <run-id>

# Or from web UI:
# Actions → Select run → Artifacts section
```

---

## Troubleshooting

### Workflow Fails to Trigger

**Check**:
- Branch protection rules
- Workflow permissions
- YAML syntax errors

**Debug**:
```bash
# Validate YAML locally
yamllint .github/workflows/ci.yaml

# Check workflow syntax
gh workflow view ci.yaml
```

### Job Fails

**Steps**:
1. Click on failed job in Actions tab
2. Expand failed step
3. Read error logs
4. Reproduce locally: `make ci`

### Permission Denied Errors

**Check**:
- Workflow permissions in Settings → Actions → General
- Job-level permissions in workflow file
- Token scopes for GITHUB_TOKEN

### Docker Push Fails

**Common Issues**:
- Not authenticated to GHCR
- Missing `packages: write` permission
- Workflow triggered on PR (expected behavior)

---

## Performance Optimization

### Current Timings

- **build-go**: ~2 minutes
- **build-image**: ~3-5 minutes (depends on cache)
- **security-scan**: ~1 minute
- **lint**: ~1 minute
- **Total**: ~5-8 minutes

### Caching Strategy

**Go Modules**:
- Cached via `actions/setup-go@v5`
- Key: `go-${{ runner.os }}-${{ hashFiles('**/go.sum') }}`

**Docker Layers**:
- Cached via GitHub Actions cache
- Type: `gha`, Mode: `max`

**Artifacts**:
- Retention: 7 days
- Size: ~10-20 MB (both binaries)

---

## Security Considerations

### Secrets

**Never commit**:
- API tokens
- Passwords
- Private keys

**Use GitHub Secrets** for sensitive data:
- Settings → Secrets and variables → Actions
- Reference in workflows: `${{ secrets.SECRET_NAME }}`

### Permissions

**Principle of Least Privilege**:
- Grant only required permissions
- Use job-level permissions when possible
- Prefer `GITHUB_TOKEN` over PATs

**Current Permissions**:
```yaml
permissions:
  contents: read        # Read repository
  packages: write       # Push to GHCR
  pull-requests: write  # Comment on PRs
  security-events: write # Upload SARIF
```

### Dependency Scanning

**Enabled**:
- Dependabot alerts
- Trivy vulnerability scanning
- govulncheck for Go vulnerabilities

**SARIF Upload**:
- Results visible in Security tab
- Integration with GitHub Code Scanning

---

## Future Enhancements

### Planned

- [ ] Multi-architecture builds (ARM64)
- [ ] Integration tests in CI
- [ ] Performance benchmarking
- [ ] Canary deployments
- [ ] SLSA provenance generation

### Under Consideration

- [ ] Matrix builds for multiple Go versions
- [ ] Scheduled nightly builds
- [ ] Automated changelog generation
- [ ] Container signing with Cosign
- [ ] Deployment to test environment

---

## Support

For workflow issues:
- 📚 Check [CI/CD documentation](../docs/CI_CD.md)
- 🐛 Open an issue with `ci` label
- 💬 Ask in GitHub Discussions

---

**Last Updated**: 2025-10-30
**Workflow Version**: 1.0.0
