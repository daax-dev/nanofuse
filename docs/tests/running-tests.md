# Running Tests

## Quick Reference

| Test Type | Command | Requirements |
|-----------|---------|--------------|
| Unit tests | `mage Test` | Go 1.21+ |
| Unit (quick) | `mage TestQuick` | Go 1.21+ |
| Unit (verbose) | `mage TestVerbose` | Go 1.21+ |
| Coverage | `mage TestCoverage` | Go 1.21+ |
| gdt CLI tests | `mage TestGdtCLI` | Go 1.21+, gdt-dev/gdt |
| gdt API tests | `mage TestGdtAPI` | Go 1.21+, daemon running |
| Build tests | `mage TestBuild` | Go 1.21+ |
| E2E tests | `sudo mage TestE2E` | KVM, Firecracker, sudo |
| All tests | `mage TestAll` | All of the above |
| CI simulation | `mage CI` | Go 1.21+ |

## Prerequisites

### Basic Requirements

```bash
# Go 1.21+
go version  # Should show 1.21 or higher

# Mage
mage --version  # Should be installed
# If not: go install github.com/magefile/mage@latest

# gdt-dev/gdt (for declarative tests)
go get github.com/gdt-dev/gdt
```

### For E2E Tests

```bash
# KVM access
ls -la /dev/kvm  # Should exist and be accessible

# Firecracker
which firecracker  # Should be installed

# Root access
sudo -n true  # Should not prompt for password (or use sudo)
```

## Running Unit Tests

### Standard Unit Tests

```bash
# Run all unit tests with race detector
mage Test

# Quick tests (no race detector, faster)
mage TestQuick

# Verbose output
mage TestVerbose

# With coverage report
mage TestCoverage
# Opens coverage.html in browser
```

### Running Specific Tests

```bash
# Run tests in a specific package
go test -v ./internal/network/...

# Run a specific test function
go test -v ./internal/network/... -run TestAllocateIP

# Run tests matching a pattern
go test -v ./... -run ".*MAC.*"
```

### Running With Race Detector

```bash
# Recommended for development
go test -race ./...
```

## Running gdt Declarative Tests

### CLI Tests

```bash
# Run all CLI tests
mage TestGdtCLI

# Or manually
go test -v ./test/gdt/cli/...

# With debug output
GDT_DEBUG=1 go test -v ./test/gdt/cli/...
```

### API Tests

```bash
# Start daemon first
sudo systemctl start nanofused

# Run API tests
mage TestGdtAPI

# Or manually
go test -v ./test/gdt/api/...
```

## Running Build Tests

```bash
# Validate kernel config and rootfs structure
mage TestBuild

# Or manually
go test -v ./test/build/...
```

## Running E2E Tests

### Full E2E Suite

```bash
# Requires sudo and KVM
sudo mage TestE2E

# Or use the standalone script
sudo scripts/e2e-test.sh
```

### E2E Options

```bash
# Skip cleanup for debugging
E2E_SKIP_CLEANUP=1 sudo mage TestE2E

# Increase boot timeout
E2E_BOOT_TIMEOUT=60 sudo scripts/e2e-test.sh

# Verbose output
E2E_VERBOSE=1 sudo scripts/e2e-test.sh
```

## CI Simulation

### Run All CI Checks Locally

```bash
# Runs: clean, build, lint, test
mage CI
```

### Using act (GitHub Actions locally)

```bash
# Install act
# macOS: brew install act
# Linux: curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash

# List available jobs
mage ActCIList

# Run all CI jobs
mage ActCI

# Run specific job
mage ActCIJob build-go

# Dry run (show what would run)
mage ActCIDryRun
```

## Watch Mode

```bash
# Requires entr
# Ubuntu: apt install entr
# macOS: brew install entr

# Watch and re-run tests on file changes
mage TestWatch
```

## Debugging Test Failures

### Verbose Output

```bash
# Maximum verbosity
go test -v -count=1 ./...

# With test logging
go test -v ./... 2>&1 | tee test-output.log
```

### Running Single Test

```bash
# Run one specific test
go test -v ./internal/network/... -run TestAllocateIP -count=1
```

### Examining Coverage

```bash
# Generate coverage profile
go test -coverprofile=coverage.out ./...

# View in terminal
go tool cover -func=coverage.out

# View in browser
go tool cover -html=coverage.out
```

## Test Environment Variables

| Variable | Purpose | Example |
|----------|---------|---------|
| `GDT_DEBUG` | Enable gdt debug output | `GDT_DEBUG=1` |
| `E2E_SKIP_CLEANUP` | Keep VMs after E2E test | `E2E_SKIP_CLEANUP=1` |
| `E2E_BOOT_TIMEOUT` | VM boot timeout (seconds) | `E2E_BOOT_TIMEOUT=60` |
| `E2E_VERBOSE` | Verbose E2E output | `E2E_VERBOSE=1` |

## Common Issues

### "permission denied" on /dev/kvm

```bash
# Add user to kvm group
sudo usermod -aG kvm $USER
# Log out and back in
```

### Tests hang waiting for daemon

```bash
# Check daemon status
sudo systemctl status nanofused

# View daemon logs
sudo journalctl -u nanofused -f
```

### gdt tests fail with "not found"

```bash
# Ensure gdt is installed
go get github.com/gdt-dev/gdt
go get github.com/gdt-dev/gdt/plugin/exec
go get github.com/gdt-dev/gdt/plugin/http
```
