# Bug Fixes Technical Deep Dive

This document provides technical details for developers to understand and test the bug fixes in this release.

## Table of Contents
1. [Task-14: IPAM State Persistence](#task-14-ipam-state-persistence)
2. [Task-15: MAC Address Generation](#task-15-mac-address-generation)
3. [Task-16: HTTP Routing Refactor](#task-16-http-routing-refactor)
4. [Task-17: Unit Tests](#task-17-unit-tests)
5. [Acceptance Criteria Verification](#acceptance-criteria-verification)

---

## Task-14: IPAM State Persistence

### Background

The IPAM (IP Address Management) component manages a pool of IP addresses (172.16.0.10 - 172.16.0.254) for VM allocation. The problem was that the `LoadAllocations` method existed but was never called during daemon startup.

### Code Changes

**File**: `internal/api/server.go`

The `loadExistingAllocations` function at line 33-62 queries the database for existing VMs and restores their IP allocations:

```go
func loadExistingAllocations(db *storage.DB, ipam *network.IPAM, logger *log.Logger) error {
    vms, err := db.ListVMs("")
    if err != nil {
        return fmt.Errorf("failed to list VMs: %w", err)
    }

    allocations := make(map[string]string)
    for _, vm := range vms {
        if vm.Config.Network.IPAddress != "" {
            allocations[vm.ID] = vm.Config.Network.IPAddress
        }
    }

    if len(allocations) > 0 {
        ipam.LoadAllocations(allocations)
    }
    return nil
}
```

This is called during server initialization at line 289-293:

```go
// Load existing IP allocations from database to prevent conflicts after restart
if err := loadExistingAllocations(db, ipam, logger); err != nil {
    logger.Printf("WARN: Failed to load existing IP allocations: %v", err)
}
```

### IPAM Implementation Details

**File**: `internal/network/ipam.go`

The `LoadAllocations` method (line 106-131) rebuilds the IPAM state:

1. Clears current allocations
2. Rebuilds the available pool excluding allocated IPs
3. Copies the allocation map

Key insight: The available pool is rebuilt from scratch to ensure consistency.

### Testing

```bash
# Run IPAM tests
go test ./internal/network/... -v -run "Load"

# Specific test: TestLoadAllocations
go test ./internal/network/... -v -run TestLoadAllocations
```

### Manual Verification

```bash
# 1. Create a VM
sudo nanofuse vm create test-ipam --image base
IP1=$(sudo nanofuse vm inspect test-ipam --format json | jq -r '.config.network.ip_address')
echo "Before restart: $IP1"

# 2. Restart daemon
sudo systemctl restart nanofused
sleep 3

# 3. Verify IP preserved
IP2=$(sudo nanofuse vm inspect test-ipam --format json | jq -r '.config.network.ip_address')
echo "After restart: $IP2"

# 4. Should be equal
[ "$IP1" = "$IP2" ] && echo "PASS: IP preserved" || echo "FAIL: IP changed"
```

---

## Task-15: MAC Address Generation

### Background

The original implementation had a placeholder `randomByte()` function returning `0x00`:

```go
// BROKEN - DO NOT USE
func randomByte() byte {
    return 0x00  // Placeholder
}
```

This resulted in all VMs getting MAC `AA:FC:00:00:00:00`, breaking network communication.

### Code Changes

**File**: `internal/firecracker/vm.go`

The corrected implementation at line 440-447:

```go
func randomByte() byte {
    b := make([]byte, 1)
    if _, err := rand.Read(b); err != nil {
        // Fallback to zero on error (extremely unlikely with crypto/rand)
        return 0x00
    }
    return b[0]
}
```

Key points:
- Uses `crypto/rand` (imported at line 5)
- Creates a 1-byte buffer and fills it with cryptographic randomness
- Only falls back to 0x00 on error (practically never happens with crypto/rand)

### MAC Format

Generated MACs follow the format `AA:FC:XX:XX:XX:XX`:
- `AA` - Locally administered bit set (not a vendor OUI)
- `FC` - Firecracker identifier
- `XX:XX:XX:XX` - Random bytes

### Testing

```bash
# Run MAC tests
go test ./internal/firecracker/... -v -run "MAC|Random"

# Specific tests
go test ./internal/firecracker/... -v -run TestGenerateMACUniqueness
go test ./internal/firecracker/... -v -run TestRandomByteDistribution
```

### Statistical Verification

The `TestRandomByteDistribution` test generates 10,000 random bytes and verifies:
- At least 200 unique values (out of 256 possible)
- Approximately uniform distribution

---

## Task-16: HTTP Routing Refactor

### Background

The original routing used manual string parsing:

```go
// BEFORE - Brittle manual routing
if strings.HasSuffix(path, "/start") {
    server.handleVMStart(w, r)
} else if strings.HasSuffix(path, "/stop") {
    server.handleVMStop(w, r)
}
```

Problems:
- No method validation at router level
- Handlers had to check HTTP method
- Error-prone string matching
- Not idiomatic Go

### Go 1.22+ ServeMux Patterns

Go 1.22 introduced method-aware routing patterns. The new routing in `setupHTTPRouter` (line 116-168):

```go
func setupHTTPRouter(server *Server) *http.ServeMux {
    mux := http.NewServeMux()

    // Method + path pattern
    mux.HandleFunc("GET /health", server.handleHealth)

    // Path parameters with {name} syntax
    mux.HandleFunc("GET /vms/{id}", server.handleGetVMByPath)
    mux.HandleFunc("POST /vms/{id}/start", server.handleVMStartByPath)

    return mux
}
```

### Path Parameter Extraction

New handler wrappers extract path parameters using `r.PathValue()`:

```go
func (s *Server) handleGetVMByPath(w http.ResponseWriter, r *http.Request) {
    vmID := r.PathValue("id")  // Go 1.22+ feature
    if vmID == "" {
        types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest, "VM ID is required", nil)
        return
    }
    s.handleGetVM(w, r, vmID)
}
```

### Route Table

| Pattern | Handler | Description |
|---------|---------|-------------|
| `GET /health` | handleHealth | Health check |
| `GET /vms` | handleListVMs | List VMs |
| `POST /vms` | handleCreateVM | Create VM |
| `GET /vms/{id}` | handleGetVMByPath | Get VM |
| `DELETE /vms/{id}` | handleDeleteVMByPath | Delete VM |
| `POST /vms/{id}/start` | handleVMStartByPath | Start VM |
| `POST /vms/{id}/stop` | handleVMStopByPath | Stop VM |
| `POST /vms/{id}/kill` | handleVMKillByPath | Kill VM |
| `GET /vms/{id}/logs` | handleVMLogsByPath | Get logs |
| `GET /images` | handleImages | List images |
| `POST /images/pull` | handleImagePull | Pull image |
| `GET /images/jobs/{id}` | handleImageJobByPath | Get job status |

### Testing

```bash
# Run API tests including routing
go test ./internal/api/... -v

# Test method not allowed
curl -X POST http://localhost:8080/health
# Should return 405 Method Not Allowed
```

### Benefits

1. **Cleaner code**: Method validation handled by router
2. **Type safety**: Path parameters have defined names
3. **Standards-based**: Uses Go stdlib patterns
4. **Better errors**: Automatic 405 responses for wrong methods

---

## Task-17: Unit Tests

### Test Files Created

1. `internal/network/ipam_test.go` - 12 test functions
2. `internal/firecracker/vm_test.go` - 14 test functions

### Test Coverage Summary

| Package | Coverage | Key Areas |
|---------|----------|-----------|
| internal/network | 22.9% | IPAM fully covered, TAP/bridge requires root |
| internal/firecracker | 7.1% | MAC generation covered, VM lifecycle requires Firecracker |

Note: Low overall coverage is expected because many functions require root privileges or running Firecracker. The critical logic (IPAM, MAC) is well tested.

### IPAM Test Functions

```
TestNewIPAM                  - Initialization
TestAllocateIP               - Basic allocation
TestAllocateIPIdempotent     - Repeated allocation
TestReleaseIP                - IP release
TestReleaseIPNonExistent     - Edge case
TestPoolExhaustion           - Error handling
TestLoadAllocations          - State restore
TestLoadAllocationsEmpty     - Edge case
TestGetAllAllocations        - Data retrieval
TestConcurrentAllocations    - Thread safety
TestConcurrentReleases       - Thread safety
TestIPAddressFormat          - Format validation
```

### VM Test Functions

```
TestGenerateMAC              - Format validation
TestGenerateMACPrefix        - Prefix check
TestGenerateMACUniqueness    - Uniqueness
TestGenerateMACNotAllZeros   - No placeholder
TestRandomByte               - RNG usage
TestRandomByteDistribution   - Distribution quality
TestGenerateMACConcurrent    - Thread safety
TestNewManager               - Manager creation
TestIsRunningFalseForInvalidPID - Process check
TestFirecrackerConfig        - Config struct
TestMachineConfigDefaults    - Config validation
TestNetworkInterfaceConfig   - NIC config
TestDriveConfig              - Drive config
TestBootSourceConfig         - Boot config
```

---

## Acceptance Criteria Verification

### Task-14 Verification

```bash
# AC4: LoadAllocations called in server.go
grep -n "LoadAllocations" internal/api/server.go
# Expected: Line 56: ipam.LoadAllocations(allocations)

# AC5: Test file exists with 5+ tests
test -f internal/network/ipam_test.go && \
grep -c "func Test" internal/network/ipam_test.go
# Expected: 12 (or more)
```

### Task-15 Verification

```bash
# AC1: crypto/rand used
grep 'crypto/rand' internal/firecracker/vm.go
# Expected: "crypto/rand"

# AC1: rand.Read used in randomByte
grep -A5 "func randomByte" internal/firecracker/vm.go | grep "rand.Read"
# Expected: Found
```

### Task-16 Verification

```bash
# AC1: Modern ServeMux patterns
grep -c 'HandleFunc("GET\|HandleFunc("POST\|HandleFunc("DELETE' internal/api/server.go
# Expected: 25+ routes

# AC2: No string parsing
grep "strings.HasSuffix" internal/api/server.go
# Expected: No output (removed)
```

### Task-17 Verification

```bash
# AC1 & AC2: Test files exist
ls -la internal/network/ipam_test.go internal/firecracker/vm_test.go
# Expected: Both files exist

# AC7: All tests pass
go test -race ./internal/...
# Expected: All PASS
```

---

## Decision Rationale

### Why Go 1.22+ Patterns Instead of chi/gorilla?

1. **Zero dependencies**: Uses stdlib only
2. **Future-proof**: Go's official routing patterns
3. **Simple**: No framework learning curve
4. **Maintenance**: One less dependency to update

### Why crypto/rand Instead of math/rand?

1. **Security**: Cryptographically secure
2. **Uniqueness**: Better entropy source
3. **Thread-safe**: No global state
4. **Industry standard**: Recommended for MAC generation

### Why Unit Tests for IPAM and MAC?

1. **Critical path**: These are foundational components
2. **Bug prevention**: Catches regressions
3. **Documentation**: Tests show expected behavior
4. **CI integration**: Automated validation

---

## Running All Verifications

```bash
#!/bin/bash
# Run all verification checks

echo "=== Building ==="
go build ./...

echo "=== Running Tests ==="
go test -race ./...

echo "=== Checking Task-14 ==="
grep -n "LoadAllocations" internal/api/server.go

echo "=== Checking Task-15 ==="
grep 'crypto/rand' internal/firecracker/vm.go

echo "=== Checking Task-16 ==="
grep -c 'HandleFunc("' internal/api/server.go

echo "=== Checking Task-17 ==="
test -f internal/network/ipam_test.go && echo "ipam_test.go: OK"
test -f internal/firecracker/vm_test.go && echo "vm_test.go: OK"

echo "=== Done ==="
```
