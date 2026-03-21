---
id: task-018
title: Replace nginx with go-chi for HTTP serving
status: Done
assignee: []
created_date: '2025-11-26'
labels:
  - Feature
  - Simplification
  - High
dependencies: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Objective: Remove nginx complexity from todo-app, use go-chi to serve static files and API on single port.

nginx adds unnecessary complexity:
- Extra service to configure and debug
- Additional failure point
- Harder to troubleshoot without exec/SSH access
- Overkill for microVM use case

go-chi can serve static files + API routes on one port with minimal code.

## Acceptance Criteria

### AC1: Backend Serves Static Files
**Given** the todo-backend is running
**When** requesting `/` or static assets
**Then** the Go server serves files from embedded or mounted directory

**Verification:**
```bash
curl -sf http://${VM_IP}:8080/ | grep -q '<html>'
# Expected: exit code 0
```

### AC2: API Routes Still Work
**Given** the todo-backend is running
**When** requesting `/api/*` endpoints
**Then** API responds correctly

**Verification:**
```bash
curl -sf http://${VM_IP}:8080/health | jq -e '.status'
curl -sf http://${VM_IP}:8080/api/todos
# Expected: exit code 0
```

### AC3: nginx Removed from Dockerfile
**Given** the refactor is complete
**When** reviewing the Dockerfile
**Then** nginx is not installed

**Verification:**
```bash
! grep -q nginx examples/todo-app/docker/Dockerfile
# Expected: exit code 0
```

### AC4: Single Port Configuration
**Given** the backend is running
**When** checking listening ports
**Then** only one HTTP port is open (8080 or 80)

**Verification:**
```bash
# In VM, only one service listening
netstat -tlnp | grep -E ':80|:8080' | wc -l
# Expected: 1
```

## Implementation Notes

1. Add chi router with static file serving:
```go
import "github.com/go-chi/chi/v5"

r := chi.NewRouter()
r.Handle("/*", http.FileServer(http.Dir("/var/www/html")))
r.Route("/api", func(r chi.Router) {
    // existing API routes
})
r.Get("/health", healthHandler)
```

2. Update Dockerfile to remove nginx
3. Update systemd service to use port 80 (or keep 8080)

## Definition of Done
- [ ] All 4 acceptance criteria pass
- [ ] nginx removed from Dockerfile
- [ ] go-chi serving static files + API
- [ ] Health check updated if needed

Priority: High
<!-- SECTION:DESCRIPTION:END -->
