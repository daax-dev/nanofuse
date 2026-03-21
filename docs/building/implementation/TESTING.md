# NanoFuse Testing Guide

This document explains how to test NanoFuse both locally and in CI.

## Philosophy: Test Locally First

**Key Principle**: The same commands run locally and in GitHub Actions CI.

All CI checks use Mage targets, so you can run them locally before pushing:

```bash
# Run full CI suite locally (exactly what GitHub Actions runs)
mage ci

# This runs:
# 1. Clean
# 2. Build (CLI + Daemon)
# 3. Lint (go fmt, go vet, golangci-lint)
# 4. Test (unit tests with race detector)
# 5. Security check (if installed)
```

## Prerequisites

```bash
# Install Mage (build tool)
go install github.com/magefile/mage@latest

# Add to PATH (if needed)
export PATH=$PATH:~/go/bin

# Verify installation
mage -l
```

## Quick Start

```bash
# Build everything
mage all

# Run tests
mage test

# Run full CI locally
mage ci
```

## Mage Targets (18 total)

### Build Targets

```bash
mage all          # Build both CLI and daemon (default)
mage cli          # Build CLI only (bin/nanofuse)
mage daemon       # Build daemon only (bin/nanofused)
mage clean        # Remove build artifacts
mage install      # Install binaries to /usr/local/bin (requires sudo)
```

### Test Targets

```bash
# Unit Tests
mage test              # Run unit tests with race detector
mage testQuick         # Fast tests (no race detector)
mage testVerbose       # Verbose test output
mage testCoverage      # Tests + HTML coverage report
mage testWatch         # Auto-run tests on file changes (requires entr)

# Integration Tests
mage testIntegration   # Run integration tests (starts daemon)
mage testAll           # Run both unit and integration tests
```

### Quality Targets

```bash
mage lint              # Run go fmt, go vet, golangci-lint
mage securityCheck     # Run gosec security scanner
mage validate          # Quick sanity check
mage check             # Check all dependencies installed
```

### CI Target

```bash
mage ci                # Run FULL CI suite (same as GitHub Actions)
```

This is the **most important target** - run before pushing!

### Docker Target

```bash
mage imageBuild        # Build Docker image (images/base)
```

## Test Categories

### 1. Unit Tests (✅ Currently: 8 tests passing)

**Location**: `*_test.go` files throughout codebase

**Run**:
```bash
mage test
# Or directly:
go test -v -race ./...
```

**Coverage**: Currently 8.3% (needs improvement)

**Current Tests**:
- `cmd/nanofuse`: Package builds
- `internal/api`: Health endpoint, server start
- `internal/client`: HTTP client operations

### 2. Integration Tests (✅ Currently: 5/5 suites passing)

**Location**: `test/integration/`

**Run**:
```bash
mage testIntegration
# Or directly:
go test -v -tags=integration ./test/integration/...
```

**Status**: ✅ **All tests passing** (0.527s)

**Test Suites**:
- TestIntegration_HealthCheck
- TestIntegration_VMLifecycle (CreateVM, ListVMs)
- TestIntegration_ImageOperations (ListImages, PullImage)
- TestIntegration_ConcurrentRequests (10 concurrent health checks)
- TestIntegration_ErrorHandling (NonExistentVM, InvalidVMConfig)

**What They Test**:
- Full daemon lifecycle (start/stop gracefully)
- API endpoints with real HTTP over Unix socket
- SQLite database initialization
- Concurrent requests
- Error handling with proper status codes
- Client library end-to-end

### 3. Build Tests

**Run**:
```bash
mage validate
```

**What It Tests**:
- Go version compatibility
- go.mod validity
- Build succeeds
- Quick test pass

## Pre-Push Checklist

Before pushing to GitHub, **always** run:

```bash
mage ci
```

This ensures:
- ✅ Code builds successfully
- ✅ All tests pass
- ✅ Code is formatted correctly
- ✅ No linter errors
- ✅ No obvious security issues

If `mage ci` passes locally, GitHub Actions will pass too.

## CI/CD Pipeline

### GitHub Actions Workflow

**Trigger**: On push to `main` or pull request

**Jobs**:

1. **build-go** - Runs `mage ci`
   - Build binaries
   - Run tests
   - Upload artifacts

2. **lint** - Runs `mage lint`
   - Format check
   - Linter checks

3. **validate** - Runs `mage validate`
   - Quick sanity checks

4. **build-image** - Build Docker image
   - Build base image
   - Push to GHCR (on main only)

5. **security-scan** - Security scanning
   - Trivy container scan
   - Go vulnerability check

6. **release** - Create GitHub release (on tags only)
   - Build release binaries
   - Create release with artifacts

### View CI Results

```bash
# View recent workflow runs
gh run list

# Watch current run
gh run watch

# View logs
gh run view --log
```

## Testing Best Practices

### 1. Test Before Committing

```bash
# Quick check
mage validate

# Full check
mage ci
```

### 2. Watch Mode During Development

```bash
# Install entr (optional)
brew install entr  # macOS
apt-get install entr  # Ubuntu

# Auto-run tests on file changes
mage testWatch
```

### 3. Check Coverage

```bash
mage testCoverage

# Opens coverage.html in browser
# Shows line-by-line coverage
```

**Coverage Goals**:
- Overall: >70%
- Critical packages: >80%
- Current: 8.3% (needs work!)

### 4. Integration Testing

```bash
# Start daemon in one terminal
./bin/nanofused --config /path/to/test-config.yaml

# Run integration tests in another terminal
mage testIntegration
```

## Debugging Failed Tests

### If `mage ci` fails:

1. **Build fails**:
   ```bash
   mage clean
   mage all
   ```

2. **Tests fail**:
   ```bash
   mage testVerbose
   # Look for specific failing test
   go test -v -run TestName ./path/to/package
   ```

3. **Lint fails**:
   ```bash
   mage lint
   # Fix reported issues
   go fmt ./...
   ```

4. **Race detector fails**:
   ```bash
   # This means there's a concurrency bug
   mage test
   # Fix the race condition
   ```

## Writing New Tests

### Unit Test Example

```go
package mypackage

import "testing"

func TestMyFunction(t *testing.T) {
    result := MyFunction(input)

    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

### Integration Test Example

```go
//go:build integration
// +build integration

package integration

func TestEndToEnd(t *testing.T) {
    // Setup daemon
    // Run operations
    // Verify results
}
```

## Performance Testing

### Benchmark Tests

```bash
# Run benchmarks
go test -bench=. ./...

# With memory profiling
go test -bench=. -benchmem ./...
```

### Load Testing

```bash
# Use integration tests with higher concurrency
# Or use external tools like vegeta, hey, k6
```

## Continuous Testing

### Pre-Commit Hook

```bash
# Add to .git/hooks/pre-commit
#!/bin/bash
mage validate || exit 1
```

### Pre-Push Hook

```bash
# Add to .git/hooks/pre-push
#!/bin/bash
mage ci || exit 1
```

## Common Issues

### "mage: command not found"

```bash
go install github.com/magefile/mage@latest
export PATH=$PATH:~/go/bin
```

### "golangci-lint not found"

```bash
# Optional - CI skips if not installed
brew install golangci-lint  # macOS
# Or download from: https://golangci-lint.run/
```

### "Integration tests fail"

Integration tests should now work! If they fail:

```bash
# Make sure daemon binary has CGO enabled:
mage daemon

# Run tests with verbose output:
mage testIntegration

# Check for SQLite errors - daemon needs CGO_ENABLED=1
```

### "Coverage is low"

```bash
# Add more tests!
# See CONTRIBUTING.md for test guidelines
```

## Summary

| Command | Purpose | When to Use |
|---------|---------|-------------|
| `mage ci` | Full CI suite | Before every push |
| `mage test` | Unit tests | During development |
| `mage validate` | Quick check | After small changes |
| `mage testCoverage` | Check coverage | When adding tests |
| `mage lint` | Format/lint | Before commit |
| `mage all` | Build everything | After pulling code |

**Remember**: If `mage ci` passes locally, it will pass in GitHub Actions!

---

For more information:
- See `magefile.go` for all target definitions
- See `.github/workflows/ci.yaml` for CI configuration
- See `CONTRIBUTING.md` for development workflow
