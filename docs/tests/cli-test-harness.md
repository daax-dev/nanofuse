# CLI Test Harness

## Overview

The CLI test harness uses [gdt-dev/gdt](https://github.com/gdt-dev/gdt) to provide declarative YAML-based testing for the `nanofuse` CLI.

## Why gdt-dev/gdt?

1. **Declarative**: Tests are YAML, not Go code
2. **exec Plugin**: Perfect for CLI testing with stdout/stderr assertions
3. **Retries**: Built-in retry logic for flaky commands
4. **Fixtures**: Setup/teardown for test environments

## Test Location

```
test/gdt/cli/
├── cli_test.go           # Go test wrapper
├── vm_commands.yaml      # VM command tests
├── image_commands.yaml   # Image command tests
└── error_handling.yaml   # Error boundary tests
```

## Running CLI Tests

```bash
# Run all CLI tests
mage TestGdtCLI

# Run specific test file
go test -v ./test/gdt/cli/... -run TestVMCommands

# Run with verbose gdt output
GDT_DEBUG=1 go test -v ./test/gdt/cli/...
```

## Writing CLI Tests

### Basic Test Structure

```yaml
# test/gdt/cli/example.yaml
name: Example CLI Tests
description: Tests for nanofuse CLI commands

tests:
  - name: help-shows-usage
    exec:
      command: nanofuse --help
      assert:
        exit_code: 0
        stdout:
          contains:
            - "Usage:"
            - "nanofuse"
```

### Testing Error Cases

```yaml
tests:
  - name: invalid-command-fails
    exec:
      command: nanofuse invalid-command
      assert:
        exit_code: 1
        stderr:
          contains:
            - "unknown command"

  - name: vm-not-found
    exec:
      command: nanofuse vm start nonexistent-vm-12345
      assert:
        exit_code: 1
        stderr:
          contains_all:
            - "not found"
            - "nonexistent-vm-12345"
```

### Using Fixtures

```yaml
fixtures:
  - daemon_running

tests:
  - name: list-vms-when-daemon-running
    exec:
      command: nanofuse vm list
      assert:
        exit_code: 0
```

## Critical Boundaries Tested

| Boundary | Test | Expected |
|----------|------|----------|
| Invalid command | `nanofuse foo` | Exit 1, error message |
| Missing required arg | `nanofuse vm create` | Exit 1, usage hint |
| Daemon not running | `nanofuse vm list` | Exit 1, daemon hint |
| VM not found | `nanofuse vm start x` | Exit 1, VM ID in error |
| Invalid JSON output | `nanofuse vm show x --format json` | Valid JSON or error |

## Go Test Wrapper

The Go test wrapper (`cli_test.go`) loads and runs the YAML tests:

```go
package cli_test

import (
    "testing"

    "github.com/gdt-dev/gdt"
    _ "github.com/gdt-dev/gdt/plugin/exec"
)

func TestCLI(t *testing.T) {
    s, err := gdt.From(".")
    if err != nil {
        t.Fatalf("failed to load tests: %s", err)
    }
    ctx := gdt.NewContext()
    s.Run(ctx, t)
}
```

## References

- [gdt exec plugin docs](https://github.com/gdt-dev/gdt/tree/main/plugin/exec)
- [gdt assertion syntax](https://github.com/gdt-dev/gdt#assertions)
