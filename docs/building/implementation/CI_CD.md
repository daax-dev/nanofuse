# CI/CD Pipeline Documentation

This document describes the NanoFuse CI/CD pipeline implementation, architecture, and best practices.

## Overview

NanoFuse uses GitHub Actions for continuous integration and deployment. The pipeline is designed for:

- **Fast feedback**: Parallel job execution, caching
- **Security**: Vulnerability scanning, dependency checks
- **Quality**: Linting, testing, coverage
- **Automation**: Automated releases, image publishing

## Pipeline Architecture

### Workflows

#### 1. Main CI/CD Pipeline (`.github/workflows/ci.yaml`)

**Triggers:**
- Push to `main` branch
- Pull requests to `main`
- Tags matching `v*` pattern

**Jobs:**

##### Job 1: `build-go` - Build and Test Go Binaries

**Purpose**: Build CLI and daemon binaries, run tests

**Steps:**
1. Checkout code
2. Setup Go 1.22 with caching
3. Download and verify dependencies
4. Run tests with race detection and coverage
5. Upload coverage to Codecov
6. Build CLI binary (static, CGO_ENABLED=0)
7. Build daemon binary (static, CGO_ENABLED=0)
8. Upload binaries as artifacts (7-day retention)

**Outputs:**
- `nanofuse-cli` artifact
- `nanofused-daemon` artifact
- Coverage report

**Optimizations:**
- Go module cache enabled
- Parallel dependency download

##### Job 2: `build-image` - Build Docker Image

**Purpose**: Build base microVM image and publish to GHCR

**Steps:**
1. Checkout code
2. Setup QEMU (multi-arch support)
3. Setup Docker Buildx
4. Login to GHCR (skip on PRs)
5. Extract metadata (tags, labels)
6. Build and push image (skip push on PRs)

**Outputs:**
- Docker image in GHCR (on main/tags only)
- Build cache for faster rebuilds

**Tags Generated:**
- `main` - Latest main branch build
- `pr-123` - Pull request builds (not pushed)
- `sha-abc123` - Commit-specific builds
- `v1.0.0` - Semantic version tags
- `latest` - Latest stable release

**Optimizations:**
- GitHub Actions cache for Docker layers
- Multi-stage builds (future)

##### Job 3: `security-scan` - Security Scanning

**Purpose**: Identify vulnerabilities in code and dependencies

**Steps:**
1. Checkout code
2. Run Trivy filesystem scan
3. Upload results to GitHub Security tab
4. Run Go vulnerability check (govulncheck)

**Security Tools:**
- **Trivy**: Container and filesystem vulnerability scanner
- **govulncheck**: Go-specific vulnerability database

**Severity Levels:**
- CRITICAL: Always fail build (production)
- HIGH: Warning (development)
- MEDIUM/LOW: Informational

**Exit Codes:**
- Currently set to `0` (don't fail on vulnerabilities)
- Will change to `1` for production releases

##### Job 4: `lint` - Code Quality

**Purpose**: Enforce code quality standards

**Steps:**
1. Checkout code
2. Setup Go with caching
3. Run golangci-lint

**Linters Enabled:**
- `errcheck` - Unchecked error detection
- `gosimple` - Code simplification
- `govet` - Official Go static analyzer
- `staticcheck` - Advanced static analysis
- `gosec` - Security issues
- `revive` - Fast, configurable linter
- And more (see `.golangci.yml`)

##### Job 5: `release` - GitHub Release Creation

**Purpose**: Create GitHub releases with binaries and notes

**Triggers**: Only on tags matching `v*`

**Dependencies**: Requires all other jobs to pass

**Steps:**
1. Checkout code with full history
2. Download CLI and daemon artifacts
3. Rename binaries for release
4. Generate release notes
5. Create GitHub release
6. Upload binaries to release

**Release Types:**
- **Prerelease**: Tags containing `alpha`, `beta`, or `rc`
- **Stable**: All other `v*` tags

**Outputs:**
- GitHub release with binaries
- Release notes (auto-generated + manual)
- Downloadable artifacts

#### 2. PR Comment Workflow (`.github/workflows/pr-comment.yaml`)

**Purpose**: Provide helpful information on pull requests

**Triggers**: PR opened or updated

**Actions:**
- Posts comment with build information
- Lists jobs that will run
- Explains what gets published (nothing on PRs)
- Provides commit SHA for traceability

#### 3. Dependabot (`.github/dependabot.yml`)

**Purpose**: Automated dependency updates

**Update Schedules:**
- **Go modules**: Weekly (Monday 09:00)
- **Docker images**: Weekly (Tuesday 09:00)
- **GitHub Actions**: Weekly (Wednesday 09:00)

**PR Limits:**
- Go: 10 open PRs max
- Docker: 5 open PRs max
- Actions: 5 open PRs max

**Labels Applied:**
- `dependencies`
- `go` / `docker` / `github-actions`

## Build Optimization

### Caching Strategy

**Go Module Cache:**
- Cached via `actions/setup-go@v5` with `cache: true`
- Key: `go-${{ runner.os }}-${{ hashFiles('**/go.sum') }}`
- Speeds up dependency downloads

**Docker Layer Cache:**
- Type: GitHub Actions cache (`type=gha`)
- Mode: `max` (cache all layers)
- Significantly reduces build times

**Artifact Retention:**
- Binaries: 7 days (GitHub Actions artifacts)
- Docker images: Unlimited (GHCR)

### Parallel Execution

Jobs run in parallel where possible:

```
build-go ─────┐
              ├──> release (on tags)
build-image ──┤
              │
security-scan ┤
              │
lint ─────────┘
```

Only `release` job has dependencies (waits for all others).

## Publishing Strategy

### Docker Images

**Registry**: GitHub Container Registry (GHCR)

**Image Naming:**
```
ghcr.io/daax-dev/nanofuse/base:<tag>
```

**Publishing Rules:**
- **Pull Requests**: Build only, no push
- **Main Branch**: Push with `latest` and `sha-*` tags
- **Version Tags**: Push with `v*` and `latest` tags

**Permissions:**
- Requires `packages: write` permission
- Uses `GITHUB_TOKEN` automatically

### Binary Releases

**Location**: GitHub Releases

**Release Triggers**: Git tags matching `v*` pattern

**Artifacts:**
- `nanofuse` - CLI binary (Linux x86_64)
- `nanofused` - Daemon binary (Linux x86_64)

**Naming Convention:**
- Semantic versioning: `v1.0.0`, `v1.0.1`, etc.
- Prereleases: `v1.0.0-alpha.1`, `v1.0.0-beta.1`, `v1.0.0-rc.1`

## Security Best Practices

### Implemented

- ✅ Pinned action versions (e.g., `@v4`, not `@latest`)
- ✅ Minimal permissions (principle of least privilege)
- ✅ Secrets not exposed in logs
- ✅ Vulnerability scanning (Trivy + govulncheck)
- ✅ SARIF results uploaded to GitHub Security
- ✅ Dependabot for dependency updates

### Future Enhancements

- 🔄 SBOM (Software Bill of Materials) generation
- 🔄 Container image signing with Cosign
- 🔄 Provenance attestations (SLSA framework)
- 🔄 Static analysis with CodeQL

## Secrets Management

### Required Secrets

**None currently!** Pipeline uses `GITHUB_TOKEN` automatically.

### Optional Secrets (Future)

- `CODECOV_TOKEN` - For private repos (Codecov integration)
- `SLACK_WEBHOOK` - For build notifications
- `GHCR_PAT` - For custom GHCR authentication (if needed)

## Monitoring and Observability

### Build Metrics

- **Build Duration**: Track via GitHub Actions insights
- **Test Coverage**: Codecov dashboard
- **Vulnerability Trends**: GitHub Security tab
- **Dependency Updates**: Dependabot insights

### Badges

README includes badges for:
- ✅ CI/CD status
- ✅ Go Report Card
- ✅ License

Add these for production:
- Code coverage (Codecov)
- Security score (OpenSSF)
- Latest release version

## Troubleshooting

### Common Issues

#### Build Fails on Go Tests

**Symptom**: `go test` fails with race conditions

**Solution:**
```bash
# Run locally with race detector
make test

# Fix race conditions in code
```

#### Docker Build Fails

**Symptom**: `docker build` fails with layer errors

**Solution:**
```bash
# Build locally to debug
make build-image

# Check Dockerfile syntax
docker build -f images/base/Dockerfile images/base
```

#### Linting Failures

**Symptom**: `golangci-lint` reports errors

**Solution:**
```bash
# Run linters locally
make lint

# Auto-fix issues
make lint-fix
```

#### GHCR Push Fails

**Symptom**: `denied: permission_denied`

**Solution:**
- Ensure `packages: write` permission in workflow
- Check GHCR visibility settings (public vs private)
- Verify `GITHUB_TOKEN` has correct scopes

### Debug Locally

Run CI checks locally before pushing:

```bash
# Run all CI checks
make ci

# Or individual steps
make deps
make vet
make lint
make test
```

## Performance Targets

### Current

- **Build Go binaries**: < 2 minutes
- **Run tests**: < 1 minute
- **Build Docker image**: < 5 minutes (with cache)
- **Total pipeline**: < 10 minutes

### Optimization Ideas

- Use matrix builds for multi-arch (parallel)
- Cache test results (rerun only on changes)
- Split integration tests into separate workflow

## Release Process

### Manual Release Steps

1. **Prepare Release**
   ```bash
   # Update version in code if needed
   git checkout main
   git pull
   ```

2. **Create Tag**
   ```bash
   # Create annotated tag
   git tag -a v1.0.0 -m "Release v1.0.0: Initial stable release"

   # Verify tag
   git tag -v v1.0.0
   ```

3. **Push Tag**
   ```bash
   # Push to trigger release
   git push origin v1.0.0
   ```

4. **Monitor Pipeline**
   - Watch GitHub Actions for build status
   - Verify all jobs pass
   - Check release created successfully

5. **Verify Artifacts**
   ```bash
   # Test downloading released binary
   curl -LO https://github.com/daax-dev/nanofuse/releases/download/v1.0.0/nanofuse
   chmod +x nanofuse
   ./nanofuse --version

   # Test pulling Docker image
   docker pull ghcr.io/daax-dev/nanofuse/base:v1.0.0
   ```

### Release Checklist

- [ ] All tests passing on main branch
- [ ] Documentation updated (README, CHANGELOG)
- [ ] Version numbers updated in code
- [ ] Release notes drafted
- [ ] Tag created and pushed
- [ ] Pipeline completed successfully
- [ ] Artifacts verified (binaries + images)
- [ ] Release announced (GitHub, discussions, etc.)

## Future Enhancements

### Phase 2
- Multi-architecture builds (ARM64)
- Integration test suite in CI
- Performance benchmarks

### Phase 3
- SLSA provenance
- Container signing
- Automated changelog generation

### Phase 4
- Matrix builds for multiple Go versions
- Canary deployments
- Automated rollback on failures

## References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Docker Buildx](https://github.com/docker/buildx)
- [golangci-lint](https://golangci-lint.run/)
- [Trivy](https://aquasecurity.github.io/trivy/)
- [Semantic Versioning](https://semver.org/)

---

**Last Updated**: 2025-10-30
**Pipeline Version**: 1.0.0
