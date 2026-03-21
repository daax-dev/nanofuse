# NanoFuse Bug Fixes and Improvements Release

This document describes the bug fixes and improvements included in this release, focusing on platform reliability and testability.

## Summary

This release closes critical bug fixes and adds comprehensive unit testing for core components:

| Task ID | Priority | Description | Status |
|---------|----------|-------------|--------|
| Task-14 | Critical | Fix IPAM State Loss on Restart | Completed |
| Task-15 | High | Fix Broken MAC Address Generation | Completed |
| Task-16 | Medium | Refactor HTTP Routing Logic | Completed |
| Task-17 | Medium | Add Unit Tests for Core Components | Completed |

## Bug Fixes

### Task-14: IPAM State Loss on Restart (Critical)

**Problem**: When the `nanofused` daemon restarted, it initialized a fresh IP pool without knowledge of IPs assigned to currently running VMs. This caused IP conflicts.

**Solution**: The `loadExistingAllocations()` function is now called during server startup to restore IP allocations from the database.

**Verification**:
```bash
# Verify LoadAllocations is called during startup
grep -n "LoadAllocations" internal/api/server.go
# Should show: ipam.LoadAllocations(allocations)
```

**Impact**: VMs maintain their IP addresses across daemon restarts, preventing network conflicts.

### Task-15: MAC Address Generation (High)

**Problem**: The `randomByte()` function was using a placeholder implementation that returned `0x00`, resulting in identical MAC addresses (`AA:FC:00:00:00:00`) for all VMs.

**Solution**: Now uses `crypto/rand` for cryptographically secure random byte generation.

**Verification**:
```bash
# Verify crypto/rand is used
grep 'crypto/rand' internal/firecracker/vm.go
# Should show: "crypto/rand"
```

**Impact**: Each VM now gets a unique MAC address, enabling multiple VMs on the same bridge network.

### Task-16: HTTP Routing Refactor (Medium)

**Problem**: The API routing used manual string parsing (`strings.HasSuffix`) which was brittle and non-idiomatic.

**Solution**: Refactored to use Go 1.22+ ServeMux method-aware patterns:
- Routes now use `HandleFunc("METHOD /path/{param}", handler)` syntax
- Path parameters extracted via `r.PathValue("param")`
- Method validation handled by the router, not handlers

**Example**:
```go
// Before (manual routing)
if strings.HasSuffix(path, "/start") {
    server.handleVMStart(w, r)
}

// After (Go 1.22+ pattern)
mux.HandleFunc("POST /vms/{id}/start", server.handleVMStartByPath)
```

**Impact**: Cleaner, more maintainable routing code with proper 405 Method Not Allowed responses.

## New Unit Tests

### IPAM Tests (`internal/network/ipam_test.go`)

| Test | Description |
|------|-------------|
| `TestNewIPAM` | Verifies IPAM initialization with 245 available IPs |
| `TestAllocateIP` | Tests basic IP allocation |
| `TestAllocateIPIdempotent` | Verifies repeated allocation returns same IP |
| `TestReleaseIP` | Tests IP release and return to pool |
| `TestPoolExhaustion` | Tests error handling when pool is exhausted |
| `TestLoadAllocations` | Tests restoring IPAM state after restart |
| `TestConcurrentAllocations` | Tests thread-safety of IPAM |
| `TestConcurrentReleases` | Tests thread-safety of IP release |

### VM Tests (`internal/firecracker/vm_test.go`)

| Test | Description |
|------|-------------|
| `TestGenerateMAC` | Verifies MAC address format |
| `TestGenerateMACPrefix` | Verifies AA:FC prefix |
| `TestGenerateMACUniqueness` | Tests that 100 MACs are unique |
| `TestGenerateMACNotAllZeros` | Verifies no broken placeholder pattern |
| `TestRandomByte` | Tests crypto/rand usage |
| `TestRandomByteDistribution` | Tests random distribution quality |
| `TestGenerateMACConcurrent` | Tests thread-safety of MAC generation |

## Running Tests

```bash
# Run all unit tests
go test ./... -v

# Run with race detector
go test -race ./...

# Run specific package tests
go test ./internal/network/... -v
go test ./internal/firecracker/... -v

# Check coverage
go test ./internal/network/... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

## API Changes

The API contract remains unchanged. The routing refactor is an internal implementation detail:

- All existing endpoints work as before
- Response formats are unchanged
- Error responses remain consistent
- 405 Method Not Allowed is now returned by the router (previously by handlers)

## Backward Compatibility

These changes are fully backward compatible:
- No API changes
- No configuration changes
- No database schema changes
- Existing VMs continue to work

## Upgrade Path

1. Stop the daemon: `sudo systemctl stop nanofused`
2. Replace binaries with new versions
3. Start the daemon: `sudo systemctl start nanofused`

Existing VMs will have their IP allocations restored automatically on restart.
