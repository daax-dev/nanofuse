# API Test Harness

## Overview

The API test harness uses [gdt-dev/gdt](https://github.com/gdt-dev/gdt) with the HTTP plugin to provide declarative YAML-based testing for the nanofused REST API.

## Why gdt HTTP Plugin?

1. **Declarative HTTP Tests**: Request/response in YAML
2. **JSON Path Assertions**: Assert on JSON response fields
3. **Status Code Checks**: Built-in HTTP status validation
4. **Header Assertions**: Validate response headers

## Test Location

```
test/gdt/api/
├── api_test.go           # Go test wrapper
├── health_api.yaml       # Health endpoint tests
├── vm_api.yaml           # VM CRUD tests
├── image_api.yaml        # Image API tests
└── error_responses.yaml  # Error boundary tests
```

## Running API Tests

```bash
# Start daemon first
sudo systemctl start nanofused

# Run all API tests
mage TestGdtAPI

# Run specific test file
go test -v ./test/gdt/api/... -run TestVMAPI

# Run with verbose gdt output
GDT_DEBUG=1 go test -v ./test/gdt/api/...
```

## Writing API Tests

### Basic Test Structure

```yaml
# test/gdt/api/example.yaml
name: Example API Tests
description: Tests for nanofused REST API

defaults:
  http:
    url: http://localhost:8080

tests:
  - name: health-endpoint-returns-200
    http:
      GET: /health
      assert:
        status: 200
        json:
          paths:
            $.status: healthy
            $.version: "0.1.0"
```

### Testing VM CRUD

```yaml
tests:
  - name: create-vm
    http:
      POST: /vms
      body:
        name: test-vm
        image: base
        memory_mb: 256
        vcpus: 1
      assert:
        status: 201
        json:
          paths:
            $.name: test-vm
            $.state: created

  - name: get-vm
    http:
      GET: /vms/{id}
      assert:
        status: 200
        json:
          paths:
            $.name: test-vm

  - name: delete-vm
    http:
      DELETE: /vms/{id}
      assert:
        status: 204
```

### Testing Error Responses

```yaml
tests:
  - name: vm-not-found-returns-404
    http:
      GET: /vms/nonexistent-vm-12345
      assert:
        status: 404
        json:
          paths:
            $.error.code: VM_NOT_FOUND

  - name: invalid-json-returns-400
    http:
      POST: /vms
      body: "not valid json"
      headers:
        Content-Type: application/json
      assert:
        status: 400
        json:
          paths:
            $.error.code: INVALID_REQUEST

  - name: missing-required-field
    http:
      POST: /vms
      body:
        # name is missing
        image: base
      assert:
        status: 400
```

## Critical Boundaries Tested

| Boundary | Endpoint | Expected |
|----------|----------|----------|
| Invalid JSON | POST /vms | 400 INVALID_REQUEST |
| Missing field | POST /vms | 400 INVALID_REQUEST |
| VM not found | GET /vms/x | 404 VM_NOT_FOUND |
| Image not found | POST /vms | 404 IMAGE_NOT_FOUND |
| Invalid state transition | POST /vms/x/start | 409 INVALID_STATE |
| Method not allowed | POST /health | 405 |
| VM locked | POST /vms/x/snapshot | 409 VM_LOCKED |

## Go Test Wrapper

```go
package api_test

import (
    "testing"

    "github.com/gdt-dev/gdt"
    _ "github.com/gdt-dev/gdt/plugin/http"
)

func TestAPI(t *testing.T) {
    // Skip if daemon not running
    resp, err := http.Get("http://localhost:8080/health")
    if err != nil {
        t.Skip("daemon not running, skipping API tests")
    }
    resp.Body.Close()

    s, err := gdt.From(".")
    if err != nil {
        t.Fatalf("failed to load tests: %s", err)
    }
    ctx := gdt.NewContext()
    s.Run(ctx, t)
}
```

## Fixtures for API Tests

```go
// test/gdt/fixtures/daemon.go
package fixtures

import (
    "github.com/gdt-dev/gdt"
)

type DaemonFixture struct{}

func (f *DaemonFixture) Setup(ctx *gdt.Context) error {
    // Ensure daemon is running
    // Could start daemon if not running
    return nil
}

func (f *DaemonFixture) Teardown(ctx *gdt.Context) error {
    // Cleanup resources
    return nil
}
```

## References

- [gdt http plugin docs](https://github.com/gdt-dev/gdt/tree/main/plugin/http)
- [JSON Path syntax](https://goessner.net/articles/JsonPath/)
- [nanofuse API types](../../internal/types/)
