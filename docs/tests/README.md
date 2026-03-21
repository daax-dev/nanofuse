# Nanofuse Test Documentation

This directory contains comprehensive documentation for the nanofuse test suites.

## Test Suites Overview

| Suite | Purpose | Requirements | Mage Target |
|-------|---------|--------------|-------------|
| [CLI Test Harness](cli-test-harness.md) | CLI command testing using gdt-dev/gdt | Go 1.21+ | `mage TestGdtCLI` |
| [API Test Harness](api-test-harness.md) | REST API testing using gdt-dev/gdt | Go 1.21+, daemon running | `mage TestGdtAPI` |
| [Build Testing](build-testing.md) | Kernel and rootfs build validation | Build dependencies | `mage TestBuild` |
| [E2E Testing](e2e-testing.md) | Full lifecycle testing | KVM, sudo, Firecracker | `mage TestE2E` |

## Quick Start

```bash
# Run all unit tests
mage Test

# Run gdt declarative tests (CLI + API)
mage TestGdt

# Run build validation tests
mage TestBuild

# Run full E2E tests (requires KVM)
sudo mage TestE2E
```

## Test Philosophy

### What We Test

1. **Critical Boundaries**: Error cases, edge cases, resource limits
2. **Integration Points**: Component interactions, API contracts
3. **User Workflows**: Real usage patterns from start to finish
4. **Regression Prevention**: Tests for every bug fix

### What We Don't Chase

1. **Coverage Percentage**: Coverage is a metric, not a goal
2. **Trivial Getters/Setters**: Don't test what the compiler tests
3. **Implementation Details**: Test behavior, not implementation

## Directory Structure

```
test/
├── gdt/
│   ├── cli/           # CLI tests (YAML + Go wrapper)
│   ├── api/           # API tests (YAML + Go wrapper)
│   └── e2e/           # E2E tests (YAML + Go wrapper)
├── build/             # Build validation tests
├── integration/       # Integration tests
└── fixtures/          # Shared test fixtures
docs/tests/
├── README.md          # This file
├── cli-test-harness.md
├── api-test-harness.md
├── build-testing.md
├── e2e-testing.md
└── running-tests.md
```

## References

- [gdt-dev/gdt Framework](https://github.com/gdt-dev/gdt)
- [Go Testing Documentation](https://go.dev/doc/tutorial/add-a-test)
- [Firecracker Testing Approach](https://github.com/firecracker-microvm/firecracker/tree/main/tests)
