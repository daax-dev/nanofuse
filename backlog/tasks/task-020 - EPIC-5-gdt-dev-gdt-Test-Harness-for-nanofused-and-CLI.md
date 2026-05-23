---
id: task-020
title: 'EPIC 5: gdt-dev/gdt Test Harness for nanofused and CLI'
status: Done
assignee: []
created_date: '2025-11-27'
updated_date: '2025-12-30 18:38'
labels:
  - Epic
  - P0
  - Testing
  - Critical
dependencies: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Outcome: Comprehensive declarative test harness for nanofused daemon and CLI using gdt-dev/gdt framework.

## Background and Rationale

### Why gdt-dev/gdt?

**Source**: [gdt-dev/gdt GitHub Repository](https://github.com/gdt-dev/gdt)

gdt (Go Declarative Testing) is a YAML-based testing framework that provides:

1. **Declarative Test Definitions**: Tests written in YAML, not Go code
2. **Plugin Architecture**: exec, http, kube plugins for different test types
3. **Fixture System**: Setup/teardown with reusable test fixtures
4. **Assertion DSL**: Rich assertion syntax for stdout, stderr, exit codes, JSON
5. **CI Integration**: Native support for Go testing, works with `go test`

### Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| YAML over Go tests | Easier to read, modify, and maintain by non-Go developers |
| exec plugin for CLI | Direct process execution with assertion on stdout/stderr |
| http plugin for API | REST API testing with JSON path assertions |
| Focus on boundaries | Critical edge cases, not coverage percentage |
| Mage integration | Single command to run all tests |

### Critical Boundaries to Test

**CLI Boundaries:**
- Command parsing edge cases (invalid flags, missing args)
- Error message formatting and exit codes
- Connection failure handling (daemon not running)
- Authentication/authorization errors
- Timeout handling

**Daemon Boundaries:**
- API request validation (malformed JSON, missing fields)
- Resource limits (pool exhaustion, disk space)
- Concurrent request handling
- State transitions (invalid state changes)
- Crash recovery and data persistence

## Acceptance Criteria
<!-- AC:BEGIN -->
### AC1: gdt-dev/gdt Integration Configured

**Given** the test harness work is complete
**When** examining the project structure
**Then** gdt-dev/gdt is properly integrated

**Verification:**
```bash
# Check go.mod includes gdt-dev/gdt
grep -q "github.com/gdt-dev/gdt" go.mod
# Expected: exit code 0

# Check test directory structure exists
test -d test/gdt && test -d test/gdt/fixtures
# Expected: exit code 0

# Check YAML test files exist
ls test/gdt/*.yaml 2>/dev/null | head -1
# Expected: at least one .yaml file
```

### AC2: CLI Test Suite Exists and Passes

**Given** the CLI test suite is implemented
**When** running the CLI tests
**Then** all critical boundary tests pass

**Verification:**
```bash
# Run CLI tests
go test -v ./test/gdt/cli/... 2>&1 | tee /tmp/cli-tests.log
grep -q "PASS" /tmp/cli-tests.log
# Expected: exit code 0

# Check for specific boundary tests
grep -qiE "invalid|error|missing|timeout" test/gdt/cli/*.yaml
# Expected: exit code 0 (tests exist for error cases)
```

### AC3: Daemon API Test Suite Exists and Passes

**Given** the daemon API test suite is implemented
**When** running the API tests
**Then** all critical boundary tests pass

**Verification:**
```bash
# Run API tests (may need daemon running)
go test -v ./test/gdt/api/... 2>&1 | tee /tmp/api-tests.log

# Check test file exists with HTTP assertions
grep -qiE "http:|status:" test/gdt/api/*.yaml
# Expected: exit code 0
```

### AC4: Mage Target Exists

**Given** the test harness is integrated
**When** listing mage targets
**Then** a gdt test target exists

**Verification:**
```bash
# Check for gdt-related mage target
mage -l | grep -qiE "gdt|harness|declarative"
# Expected: exit code 0

# Run the target (dry-run style check)
mage -h 2>&1 | grep -qiE "test"
# Expected: exit code 0
```

### AC5: Tests Run in GitHub Actions

**Given** the test harness is implemented
**When** examining CI configuration
**Then** gdt tests are included in the CI pipeline

**Verification:**
```bash
# Check GitHub Actions workflow includes gdt tests
grep -rqiE "gdt|TestGdt|test/gdt" .github/workflows/*.yaml
# Expected: exit code 0 (may need to be added)
```

### AC6: Test Fixtures for Common Scenarios

**Given** the test harness uses fixtures
**When** examining fixture definitions
**Then** reusable fixtures exist for common test scenarios

**Verification:**
```bash
# Check for fixture files
ls test/gdt/fixtures/*.go 2>/dev/null || ls test/gdt/fixtures/*.yaml 2>/dev/null
# Expected: at least one fixture file

# Check fixtures cover key scenarios
grep -rqiE "setup|teardown|fixture" test/gdt/
# Expected: exit code 0
```

### AC7: Documentation Exists in docs/tests/

**Given** the test harness is implemented
**When** checking documentation
**Then** comprehensive docs exist in docs/tests/

**Verification:**
```bash
# Check docs/tests directory exists
test -d docs/tests
# Expected: exit code 0

# Check CLI test docs exist
test -f docs/tests/cli-test-harness.md
# Expected: exit code 0

# Check API test docs exist
test -f docs/tests/api-test-harness.md
# Expected: exit code 0
```

## Technical Implementation

### Directory Structure

```
test/
├── gdt/
│   ├── cli/                     # CLI test suites
│   │   ├── cli_test.go          # Go test wrapper
│   │   ├── vm_commands.yaml     # VM command tests
│   │   ├── image_commands.yaml  # Image command tests
│   │   └── error_handling.yaml  # Error boundary tests
│   ├── api/                     # API test suites
│   │   ├── api_test.go          # Go test wrapper
│   │   ├── vm_api.yaml          # VM API tests
│   │   ├── health_api.yaml      # Health endpoint tests
│   │   └── error_responses.yaml # Error boundary tests
│   └── fixtures/
│       ├── daemon.go            # Daemon lifecycle fixture
│       ├── testvm.go            # Test VM fixture
│       └── cleanup.go           # Cleanup fixture
docs/
└── tests/
    ├── cli-test-harness.md      # CLI testing documentation
    ├── api-test-harness.md      # API testing documentation
    └── running-tests.md         # How to run tests
```

### Example YAML Test (gdt format)

```yaml
# test/gdt/cli/error_handling.yaml
name: CLI Error Handling Boundaries
description: Tests CLI behavior at error boundaries

tests:
  - name: missing-vm-returns-error
    exec:
      command: nanofuse vm start nonexistent-vm-12345
      assert:
        exit_code: 1
        stderr:
          contains:
            - "not found"
            - "nonexistent-vm-12345"

  - name: daemon-not-running
    skip_if:
      - daemon_running
    exec:
      command: nanofuse vm list
      assert:
        exit_code: 1
        stderr:
          contains_one_of:
            - "connection refused"
            - "daemon not running"
```

### References

- **gdt-dev/gdt**: https://github.com/gdt-dev/gdt
- **gdt exec plugin**: https://github.com/gdt-dev/gdt/tree/main/plugin/exec
- **gdt http plugin**: https://github.com/gdt-dev/gdt/tree/main/plugin/http
- **Go testing best practices**: https://go.dev/doc/tutorial/add-a-test

## Definition of Done
- [ ] #1 All 7 acceptance criteria pass
- [ ] #2 gdt-dev/gdt dependency added to go.mod
- [ ] #3 CLI test suite with 5+ boundary tests
- [ ] #4 API test suite with 5+ boundary tests
- [ ] #5 Test fixtures for daemon and VM lifecycle
- [ ] #6 `mage TestGdt` target works
- [ ] #7 GitHub Actions runs gdt tests
- [ ] #8 Documentation in docs/tests/

Priority: P0 (MUST HAVE)
Output Files:
- `test/gdt/cli/*.yaml`
- `test/gdt/api/*.yaml`
- `test/gdt/fixtures/*.go`
- `docs/tests/cli-test-harness.md`
- `docs/tests/api-test-harness.md`
<!-- SECTION:DESCRIPTION:END -->

<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## 2025-12-30: Implementation Complete

PR #77 submitted: https://daax-dev/nanofuse/pull/77

### Deliverables
- CLI test suite: 11 boundary tests (error_handling.yaml + help_usage.yaml)
- API test suite: 10 boundary tests (error_responses.yaml + health.yaml)
- Mage targets: TestGdtCLI, TestGdtAPI
- All 7 acceptance criteria verified

### Key Implementation Notes
- gdt YAML format requires `shell: bash` and `exec:` as single line
- API tests use curl -sf for error detection (exit code 22 on HTTP errors)
- API tests skip in CI (daemon not running) but CLI tests always run

## 2025-12-30: Task Completed

PR #80 merged to main. gdt test harness fully implemented with CLI and API test suites.
<!-- SECTION:NOTES:END -->
