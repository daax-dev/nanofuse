---
id: task-016
title: Refactor HTTP Routing Logic
status: Done
assignee: []
created_date: '2025-11-25'
labels:
  - Tech Debt
  - Medium
dependencies: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Objective: Replace manual string parsing in HTTP router with robust pattern matching.

The current routing logic in `internal/api/server.go` uses manual string parsing (e.g., `strings.HasSuffix`) which is brittle and non-idiomatic.

## Acceptance Criteria

### AC1: Modern ServeMux Patterns Used
**Given** the refactor is complete
**When** reviewing the code
**Then** Go 1.22+ ServeMux patterns or a standard router is used

**Verification:**
```bash
# Check for modern pattern syntax (Go 1.22+)
grep -E 'HandleFunc\("(GET|POST|PUT|DELETE|PATCH) ' internal/api/server.go || \
grep -E 'mux\.Handle.*Method' internal/api/server.go || \
grep -E 'chi\.|gorilla\.|gin\.' internal/api/server.go
# Expected: exit code 0 (one of these patterns found)

# Check that old string parsing is removed
! grep -E 'strings\.HasSuffix.*("/|Handler)' internal/api/server.go
# Expected: exit code 0 (old pattern not found)
```

### AC2: Routes Are Clearly Defined
**Given** the refactor is complete
**When** reviewing the routing setup
**Then** all routes are defined in one clear section

**Verification:**
```bash
# Routes should be grouped together (not scattered)
# Count route definitions in a 50-line window
ROUTE_BLOCK=$(grep -n "HandleFunc\|Handle(" internal/api/server.go | head -20)
FIRST_LINE=$(echo "$ROUTE_BLOCK" | head -1 | cut -d: -f1)
LAST_LINE=$(echo "$ROUTE_BLOCK" | tail -1 | cut -d: -f1)
SPAN=$((LAST_LINE - FIRST_LINE))

echo "Routes span $SPAN lines"
[ $SPAN -lt 100 ]  # Routes should be within 100 lines
# Expected: exit code 0
```

### AC3: All Existing API Endpoints Still Work
**Given** the refactor is complete
**When** testing each API endpoint
**Then** all return expected responses

**Verification:**
```bash
# Start daemon and run API tests
sudo systemctl restart nanofused
sleep 3

# Test key endpoints
PASS=0

# Health check
curl -sf http://localhost:8080/health && ((PASS++)) || echo "FAIL: /health"

# List VMs
curl -sf http://localhost:8080/vms && ((PASS++)) || echo "FAIL: GET /vms"

# List images
curl -sf http://localhost:8080/images && ((PASS++)) || echo "FAIL: GET /images"

# Create VM (should work or return validation error, not 404)
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/vms -d '{}')
[ "$STATUS" != "404" ] && ((PASS++)) || echo "FAIL: POST /vms returned 404"

echo "Result: $PASS/4 endpoints working"
[ $PASS -eq 4 ]
# Expected: exit code 0
```

### AC4: VM CRUD Operations Work
**Given** the refactor is complete
**When** performing VM lifecycle operations via API
**Then** all CRUD operations succeed

**Verification:**
```bash
# Test full CRUD cycle via API
API="http://localhost:8080"

# Create
RESP=$(curl -sf -X POST "$API/vms" -H "Content-Type: application/json" \
  -d '{"name":"api-test","image":"base"}')
echo "Create: $RESP"

# Read
curl -sf "$API/vms/api-test" | grep -q "api-test"
# Expected: exit code 0

# Start
curl -sf -X POST "$API/vms/api-test/start"
# Expected: exit code 0

# Stop
sleep 5
curl -sf -X POST "$API/vms/api-test/stop"
# Expected: exit code 0

# Delete
curl -sf -X DELETE "$API/vms/api-test"
# Expected: exit code 0

echo "All CRUD operations passed"
```

### AC5: Path Parameter Extraction Works
**Given** routes with path parameters (e.g., /vms/{id})
**When** calling endpoints with various IDs
**Then** the correct VM is operated on

**Verification:**
```bash
# Create VMs with distinct names
sudo nanofuse vm create route-test-alpha --image base
sudo nanofuse vm create route-test-beta --image base

# Verify correct VM is returned
curl -sf http://localhost:8080/vms/route-test-alpha | grep -q "route-test-alpha"
# Expected: exit code 0

curl -sf http://localhost:8080/vms/route-test-beta | grep -q "route-test-beta"
# Expected: exit code 0

# Cleanup
sudo nanofuse vm delete route-test-alpha
sudo nanofuse vm delete route-test-beta
```

### AC6: No Regression in API Contract
**Given** the refactor is complete
**When** comparing API behavior
**Then** response formats are unchanged

**Verification:**
```bash
# Compare response structure before/after
# This assumes api/openapi.yaml exists

# Validate against OpenAPI spec if available
if [ -f api/openapi.yaml ]; then
  # Basic check: responses match expected structure
  curl -sf http://localhost:8080/vms | jq -e '. | type == "array"'
  # Expected: exit code 0
fi

# Manual verification: key fields present
curl -sf http://localhost:8080/health | jq -e '.status'
# Expected: exit code 0
```

## Definition of Done
- [ ] All 6 acceptance criteria pass
- [ ] No string parsing in routing logic
- [ ] All existing API tests pass
- [ ] Code review approved

Priority: Medium
Implementation Location: `internal/api/server.go`
<!-- SECTION:DESCRIPTION:END -->
