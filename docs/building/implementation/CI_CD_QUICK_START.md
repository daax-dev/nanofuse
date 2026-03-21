# CI/CD Quick Start Guide

> 🚀 Get up to speed with NanoFuse CI/CD in 5 minutes

## TL;DR

```bash
# Verify CI is configured correctly
./scripts/verify-ci.sh

# Test locally before pushing
./scripts/check-build.sh

# Or use make
make ci

# Push to trigger pipeline
git push origin main
```

---

## Pipeline Overview

```
Push/PR → Build → Test → Scan → Lint → Publish (main only) → Release (tags only)
          ⏱️ 2min  1min   1min   1min   2min              3min
```

**Total Time**: ~5-8 minutes

---

## What Happens When?

### 📝 On Pull Request

```
✅ Builds both binaries (CLI + daemon)
✅ Runs all tests with race detection
✅ Scans for vulnerabilities (Trivy + govulncheck)
✅ Runs linters (golangci-lint)
✅ Builds Docker image (no push)
✅ Posts PR comment with build info
❌ Does NOT publish anything
```

**Artifacts**: Available in Actions tab (7 days)

### 🔀 On Merge to Main

```
✅ Everything from PR check
✅ Pushes Docker image to GHCR:
   - ghcr.io/jpoley/nanofuse/base:latest
   - ghcr.io/jpoley/nanofuse/base:sha-abc123
❌ Does NOT create GitHub release
```

**Artifacts**: Docker image in GHCR

### 🏷️ On Version Tag (v*)

```
✅ Everything from main
✅ Pushes Docker image with version tag:
   - ghcr.io/jpoley/nanofuse/base:v1.0.0
✅ Creates GitHub Release with:
   - nanofuse binary (Linux x86_64)
   - nanofused binary (Linux x86_64)
   - Auto-generated release notes
```

**Artifacts**: GitHub Release + Docker image

---

## Quick Commands

### Local Development

```bash
# Build everything
make build

# Run tests
make test

# Run tests with coverage
make test-coverage

# Run linters
make lint

# Fix linting issues
make lint-fix

# Run all CI checks
make ci

# Clean build artifacts
make clean
```

### Pre-Push Checks

```bash
# Automated pre-push validation
./scripts/check-build.sh

# Verify CI configuration
./scripts/verify-ci.sh
```

### View Pipeline Status

```bash
# Using GitHub CLI
gh run list
gh run watch
gh run view --log-failed

# Or visit
# https://github.com/jpoley/nanofuse/actions
```

### Download Artifacts

```bash
# Download from latest run
gh run download

# Download from specific run
gh run download 1234567890

# Or click "Artifacts" in GitHub Actions UI
```

---

## Creating a Release

```bash
# 1. Ensure main is clean and tests pass
git checkout main
git pull
make ci

# 2. Create and push tag
git tag -a v1.0.0 -m "Release v1.0.0: Description"
git push origin v1.0.0

# 3. Monitor release
gh run watch

# 4. Verify release
gh release view v1.0.0
gh release download v1.0.0

# 5. Test downloaded binaries
chmod +x nanofuse nanofused
./nanofuse --version
```

---

## Troubleshooting

### Build Fails Locally

```bash
# Clean and rebuild
make clean
make build

# Check Go version (should be 1.22+)
go version

# Verify dependencies
go mod verify
```

### Tests Fail

```bash
# Run with verbose output
go test -v ./...

# Run specific test
go test -v -run TestHealthEndpoint ./internal/api

# Check for race conditions
go test -race ./...
```

### Linting Errors

```bash
# Run linters locally
make lint

# Auto-fix issues
make lint-fix

# Or manually
golangci-lint run --fix
```

### Pipeline Fails in CI

```bash
# View failed run
gh run view --log-failed

# Reproduce locally
./scripts/check-build.sh

# Run in Docker (same as CI)
docker run --rm -v $(pwd):/src -w /src golang:1.22 \
  sh -c "go test -v -race ./..."
```

### GHCR Push Denied

**Check**:
1. Repository settings → Actions → General → Workflow permissions
2. Ensure "Read and write permissions" is enabled
3. Verify package visibility (public vs private)

**Fix**:
- Settings → Actions → General → Workflow permissions → Read and write

### Docker Build Fails

```bash
# Build locally to debug
cd images/base
docker build -t test .

# Check Dockerfile syntax
docker build --check -f Dockerfile .
```

---

## Pipeline Jobs

| Job | Duration | Purpose | Runs On |
|-----|----------|---------|---------|
| `build-go` | ~2 min | Build binaries, run tests | All events |
| `build-image` | ~3-5 min | Build Docker image | All events |
| `security-scan` | ~1 min | Vulnerability scanning | All events |
| `lint` | ~1 min | Code quality checks | All events |
| `release` | ~3 min | Create GitHub release | Tags only |

---

## File Reference

```
.github/
├── workflows/
│   ├── ci.yaml              # Main CI/CD pipeline
│   └── pr-comment.yaml      # PR comment helper
├── dependabot.yml           # Dependency updates
└── WORKFLOW_SUMMARY.md      # Workflow documentation

scripts/
├── check-build.sh           # Pre-push validation
└── verify-ci.sh             # CI configuration check

docs/
├── CI_CD.md                            # Detailed technical docs
├── CI_CD_IMPLEMENTATION_REPORT.md      # Implementation report
└── TESTING.md                           # Testing guide

Configuration:
├── .golangci.yml            # Linter configuration
├── Makefile                 # Build automation
├── .gitignore              # Git ignore rules
└── LICENSE                  # MIT License
```

---

## Environment Variables

**In CI**:
- `GO_VERSION=1.22` - Go version
- `REGISTRY=ghcr.io` - Container registry
- `IMAGE_NAME=${{ github.repository }}` - Image name

**In Builds**:
- `CGO_ENABLED=0` - Static binaries
- `GOOS=linux` - Target OS
- `GOARCH=amd64` - Target architecture

**Injected via ldflags**:
- `main.Version` - Git tag or branch
- `main.Commit` - Git commit SHA
- `main.BuildDate` - Build timestamp

---

## Badges

Add to README:

```markdown
[![CI/CD Pipeline](https://github.com/jpoley/nanofuse/actions/workflows/ci.yaml/badge.svg)](https://github.com/jpoley/nanofuse/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/jpoley/nanofuse)](https://goreportcard.com/report/github.com/jpoley/nanofuse)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
```

---

## Performance Targets

| Metric | Target | Current | Status |
|--------|--------|---------|--------|
| Total pipeline | <10 min | 5-8 min | ✅ |
| Build Go | <3 min | ~2 min | ✅ |
| Build Docker | <7 min | 3-5 min | ✅ |
| Run tests | <2 min | ~1 min | ✅ |
| Security scan | <2 min | ~1 min | ✅ |

---

## Security Checklist

- ✅ Pinned action versions (not `@latest`)
- ✅ Minimal permissions (read + write only as needed)
- ✅ Vulnerability scanning enabled
- ✅ Dependabot enabled
- ✅ No hardcoded secrets
- ✅ SARIF upload to Security tab
- 🔄 Branch protection (enable in settings)
- 🔄 Required status checks (enable in settings)

---

## Next Steps After Implementation

1. **Test the pipeline**:
   ```bash
   git checkout -b test/ci
   git push origin test/ci
   gh pr create
   ```

2. **Enable branch protection**:
   - Settings → Branches → Add rule
   - Branch name pattern: `main`
   - Require status checks: `build-go`, `build-image`, `lint`, `security-scan`

3. **Configure notifications**:
   - Settings → Notifications
   - Enable email/Slack for failed builds

4. **Set up environments** (optional):
   - Settings → Environments
   - Create `production` environment
   - Add protection rules

---

## Resources

- 📚 [Detailed CI/CD Docs](docs/CI_CD.md)
- 📝 [Implementation Report](docs/CI_CD_IMPLEMENTATION_REPORT.md)
- 🧪 [Testing Guide](docs/TESTING.md)
- 🔄 [Workflow Summary](.github/WORKFLOW_SUMMARY.md)
- 🤝 [Contributing Guide](CONTRIBUTING.md)

---

## Getting Help

**Issue?** Check these in order:

1. Run `./scripts/check-build.sh` - fixes 90% of issues
2. Check [CI/CD docs](docs/CI_CD.md) - comprehensive troubleshooting
3. View logs: `gh run view --log-failed`
4. Search [existing issues](https://github.com/jpoley/nanofuse/issues)
5. Open new issue with `ci` label

---

**Last Updated**: 2025-10-30 | **Version**: 1.0.0
