# Phase 0: Governance Bootstrap - Implementation Checklist

**Goal**: Establish organizational discipline before resuming feature work
**Duration**: 1-2 days
**Status**: READY TO EXECUTE

---

## Prerequisites

- [ ] Human authority has reviewed [ARCHITECT_ELEVATOR_ANALYSIS.md](./ARCHITECT_ELEVATOR_ANALYSIS.md)
- [ ] Human authority has reviewed [EXECUTIVE_SUMMARY.md](./EXECUTIVE_SUMMARY.md)
- [ ] Governance approach approved
- [ ] Phase 0 investment (1-2 days) authorized
- [ ] MVP scope confirmed

---

## Step 1: Initialize .specify/ Structure (2-4 hours)

### Create Directory Structure

```bash
# Create .specify/ directories
mkdir -p .specify/{features,templates,scripts/bash,memory}

# Verify structure
tree .specify/
```

**Expected Output**:
```
.specify/
├── features/          # Feature-specific artifacts
├── templates/         # Standardized templates
├── scripts/bash/      # Workflow automation
└── memory/            # Project principles
```

**Checklist**:
- [ ] .specify/features/ created
- [ ] .specify/templates/ created
- [ ] .specify/scripts/bash/ created
- [ ] .specify/memory/ created

---

### Move Constitution to Correct Location

```bash
# Move constitution from .claude/ to .specify/memory/
cp .claude/constitution.md .specify/memory/constitution.md

# Verify
cat .specify/memory/constitution.md | head -20
```

**Checklist**:
- [ ] Constitution moved to .specify/memory/constitution.md
- [ ] Original .claude/constitution.md kept (backwards compatibility)
- [ ] Constitution declares itself SUPREME AUTHORITY

---

### Create Spec Template

**Option 1**: Use jp-spec-kit template (if installed)
```bash
# Check if jp-spec-kit installed
specify --help

# If not installed:
uv tool install specify-cli --from git+https://github.com/jpoley/jp-spec-kit.git

# Copy template
cp ~/.specify/templates/spec-template.md .specify/templates/spec-template.md
```

**Option 2**: Create minimal template manually

Create `.specify/templates/spec-template.md`:

```markdown
# Feature Specification: {Feature Name}

**Status**: Draft | Review | Approved | Implemented
**Owner**: {Name or Agent}
**Created**: {YYYY-MM-DD}
**Updated**: {YYYY-MM-DD}

---

## Overview

### Problem Statement

{What problem does this feature solve?}

### Proposed Solution

{WHAT to build - technology-agnostic}

### Success Criteria

1. {Measurable criterion 1}
2. {Measurable criterion 2}
3. {Measurable criterion 3}

---

## Requirements

### Functional Requirements

1. {Requirement 1}
2. {Requirement 2}

### Non-Functional Requirements

1. Performance: {Target}
2. Security: {Requirements}
3. Observability: {Requirements}

---

## Out of Scope

- {What is explicitly NOT included}

---

## Assumptions

1. {Assumption 1}
2. {Assumption 2}

---

## Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|-----------|------------|
| {Risk 1} | HIGH/MEDIUM/LOW | HIGH/MEDIUM/LOW | {Strategy} |

---

## Open Questions

1. {Question 1} - [NEEDS CLARIFICATION]
2. {Question 2} - [NEEDS CLARIFICATION]

{Maximum 3 questions - make informed guesses beyond this}

---

## Acceptance Criteria

- [ ] {Testable criterion 1}
- [ ] {Testable criterion 2}
- [ ] {Testable criterion 3}

---

## Notes

{Additional context, links, research}
```

**Checklist**:
- [ ] Template created at .specify/templates/spec-template.md
- [ ] Template follows technology-agnostic principle
- [ ] Template includes measurable success criteria

---

## Step 2: Document Existing Work Retroactively (3-4 hours)

### Create Phase 1 Foundation Spec

Create `.specify/features/phase1-foundation/spec.md`:

```markdown
# Feature Specification: Phase 1 - Foundation Platform

**Status**: Implemented (Historical)
**Owner**: Development Team
**Created**: 2025-11-23 (Retroactive)
**Updated**: 2025-11-23

---

## Overview

### Problem Statement

Build a Firecracker-based microVM platform that can pull and run OCI images with networking, similar to Slicer, as a foundation for Trigger.dev dual-environment workloads.

### Proposed Solution

Create CLI tool, API daemon, and base Ubuntu 24.04 image with:
- VM lifecycle management (create, start, stop, delete)
- Networking (NAT + bridge + TAP devices per VM)
- Image management (pull from GHCR with authentication)
- Basic observability (console logs, VM status)

### Success Criteria

1. ✅ VMs boot from OCI images stored in GHCR
2. ✅ VMs have network connectivity (pingable from host)
3. ⚠️ Services run inside VMs (nginx, custom apps) - PARTIAL
4. ✅ CLI provides user-friendly interface
5. ✅ API provides programmatic interface

---

## Requirements

### Functional Requirements

1. ✅ Pull OCI images from GHCR with authentication
2. ✅ Create VMs from pulled images
3. ✅ Start VMs with Firecracker
4. ✅ Configure networking (IP allocation, TAP devices, NAT)
5. ✅ Stop VMs gracefully (with SIGKILL fallback)
6. ✅ Delete VMs and cleanup resources
7. ✅ View VM console logs
8. ✅ List VMs and show status

### Non-Functional Requirements

1. ✅ Performance: VM boot time <30 seconds
2. ✅ Performance: Host-to-VM latency <1ms
3. ⚠️ Reliability: Services start automatically - FAILING
4. ✅ Security: GHCR authentication via tokens
5. ✅ Observability: Console logs accessible

---

## Out of Scope

- ❌ Snapshot/resume (Phase 2)
- ❌ Trigger.dev integration (Phase 3)
- ❌ Multi-architecture (ARM64) - deferred
- ❌ Web UI - deferred

---

## Implementation Summary (Historical)

### What Was Built

1. **CLI Tool** (`cmd/nanofuse/`)
   - Binary: 8.5MB
   - Commands: image pull, vm run, vm stop, vm delete, vm status, vm logs
   - Status: ✅ Compiles and runs

2. **API Daemon** (`cmd/nanofused/`)
   - Binary: 9.0MB
   - Endpoints: /images/pull, /vms, /vms/:id, /snapshots
   - Status: ✅ Compiles, partial implementation

3. **Base Image** (`images/base/`)
   - OS: Ubuntu 24.04 + systemd
   - Kernel: 5.10.240 (Slicer proven kernel)
   - Status: ✅ Builds, boots in Firecracker

4. **Networking** (`internal/network/`)
   - Bridge: nanofuse0 (172.16.0.1/24)
   - IPAM: 172.16.0.10-254
   - TAP devices per VM
   - NAT via iptables
   - Status: ✅ Working (sub-millisecond latency)

5. **CI/CD** (`.github/workflows/`)
   - Build: Go binaries + Docker images
   - Test: Unit tests, integration tests
   - Security: Trivy, gosec
   - Status: ✅ Pipeline exists

---

## Known Issues

1. **Services Not Starting in VMs** (BLOCKING MVP)
   - Symptom: nginx, todo-backend fail with "[FAILED]"
   - Hypothesis: Missing `init=/lib/systemd/systemd` kernel parameter
   - Status: OPEN (TASK-001)

2. **No E2E Validation** (BLOCKING MVP)
   - Symptom: Cannot prove full workflow works
   - Gap: No end-to-end test script
   - Status: OPEN (TASK-002)

3. **Inconsistent Error Handling**
   - Symptom: Error formats vary across components
   - Gap: No unified error schema
   - Status: OPEN (TASK-003)

---

## Acceptance Criteria

- [x] Code compiles (Go build succeeds)
- [x] Unit tests pass (8/8)
- [x] VMs boot and are pingable
- [ ] Services run in VMs - FAILING
- [ ] E2E test passes - NOT CREATED
- [x] Documentation exists
- [ ] Documentation accurate - PARTIAL

---

## Notes

This spec documents work already completed in Phase 1. It is marked "Implemented (Historical)" to capture what was built without the spec-first process. Future work must create specs BEFORE implementation.

**Lesson Learned**: Building without specs led to gaps (services not starting, no E2E test). Spec-driven development prevents this.
```

**Checklist**:
- [ ] Phase 1 foundation spec created
- [ ] Status marked "Implemented (Historical)"
- [ ] Known issues documented
- [ ] Lessons learned captured

---

### Create Spec for Current Work (MVP Fixes)

Create `.specify/features/mvp-completion/spec.md`:

```markdown
# Feature Specification: MVP Completion

**Status**: Approved
**Owner**: Platform Engineer
**Created**: 2025-11-23
**Updated**: 2025-11-23

---

## Overview

### Problem Statement

Phase 1 foundation is 70% complete but has critical blockers preventing MVP demonstration:
1. Services don't start in VMs (nginx, todo-backend failing)
2. No end-to-end test to prove full workflow
3. Error handling inconsistent across components

### Proposed Solution

Fix the three MVP blockers to achieve a demonstrable, production-ready foundation platform.

### Success Criteria

1. Services run successfully in VMs (nginx accessible on port 80, todo-backend on 8080)
2. E2E test script runs green (full workflow validated)
3. All API errors follow unified schema (code, message, context)

---

## Requirements

### Functional Requirements

1. **Fix Service Startup**
   - Add `init=/lib/systemd/systemd` to VM kernel arguments
   - Verify systemd starts as PID 1
   - Confirm nginx and todo-backend services start

2. **Create E2E Test**
   - Script validates: pull image → create VM → services work → delete VM
   - Each step has clear validation
   - Failures provide actionable error messages

3. **Unify Error Handling**
   - Define error schema in API_CONTRACT.md
   - Update all API handlers to use schema
   - CLI parses and displays errors consistently

### Non-Functional Requirements

1. Performance: No performance regression
2. Reliability: E2E test must be stable (>95% pass rate)
3. Observability: Errors include context for debugging

---

## Out of Scope

- Snapshot/resume (Phase 2)
- Trigger.dev integration (Phase 3)
- Advanced error recovery (retry logic)

---

## Acceptance Criteria

- [ ] `curl http://172.16.0.10:80` returns HTML
- [ ] `curl http://172.16.0.10:8080/health` returns JSON
- [ ] E2E test script (scripts/test-e2e.sh) runs without errors
- [ ] All API errors include: code, message, context
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Documentation updated

---

## Tasks

See backlog/tasks/ for detailed breakdown:
- TASK-001: Fix systemd init parameter
- TASK-002: Create E2E test script
- TASK-003: Unified error handling
```

**Checklist**:
- [ ] MVP completion spec created
- [ ] Status marked "Approved"
- [ ] Success criteria measurable
- [ ] Tasks reference backlog

---

## Step 3: Populate Backlog (2-3 hours)

### Create Core Tasks

```bash
# Verify backlog exists
ls -la backlog/

# Create tasks using backlog CLI or manually create markdown files
```

**TASK-001**: Fix Systemd Init Parameter

Create `backlog/tasks/TASK-001-fix-systemd-init.md`:

```markdown
# TASK-001: Fix Systemd Init Parameter

**Status**: To Do
**Priority**: P0 (Critical)
**Milestone**: MVP
**Labels**: Bug, P0, MVP
**Effort**: 2-4 hours
**Owner**: Platform Engineer

---

## Description

Services (nginx, todo-backend) fail to start in VMs with error:
```
[FAILED] Failed to start nginx.service
```

**Hypothesis**: Missing `init=/lib/systemd/systemd` kernel parameter causes wrong init or systemd failure.

---

## Specification Reference

- Spec: .specify/features/mvp-completion/spec.md
- Plan: .specify/features/mvp-completion/plan.md (to be created)

---

## Acceptance Criteria

- [ ] Add `init=/lib/systemd/systemd` to kernel args in internal/api/vm_handlers.go (lines 99, 217)
- [ ] Rebuild daemon: `mage daemon`
- [ ] Create fresh VM
- [ ] Verify systemd starts as PID 1 (check console logs)
- [ ] Verify nginx starts: `curl http://172.16.0.10:80` returns HTML
- [ ] Verify todo-backend starts: `curl http://172.16.0.10:8080/health` returns JSON
- [ ] All tests pass
- [ ] Documentation updated

---

## Dependencies

- Blocked by: None (can start immediately)
- Blocks: TASK-002 (E2E test needs services working)

---

## Implementation Notes

**Files to modify**:
1. `internal/api/vm_handlers.go` line 99
2. `internal/api/vm_handlers.go` line 217

**Changes**:
```go
// Before (line 99):
KernelArgs: "console=ttyS0 root=/dev/vda1 rw",

// After:
KernelArgs: "console=ttyS0 root=/dev/vda1 rw init=/lib/systemd/systemd",

// Before (line 217):
config.KernelArgs = fmt.Sprintf(
    "console=ttyS0 root=/dev/vda1 rw ip=%s::%s:255.255.255.0::eth0:off net.ifnames=0 biosdevname=0",
    ip, network.BridgeGateway,
)

// After:
config.KernelArgs = fmt.Sprintf(
    "console=ttyS0 root=/dev/vda1 rw init=/lib/systemd/systemd ip=%s::%s:255.255.255.0::eth0:off net.ifnames=0 biosdevname=0",
    ip, network.BridgeGateway,
)
```

**Testing**:
```bash
# Rebuild
mage daemon

# Stop old daemon
sudo killall nanofused

# Start new daemon
sudo ./bin/nanofused

# Create test VM
./bin/nanofuse vm run default test-services

# Wait 30 seconds for boot

# Test nginx
curl http://172.16.0.10:80

# Test todo-backend
curl http://172.16.0.10:8080/health

# Check console logs
sudo tail -100 /var/lib/nanofuse/vms/*/console.log | grep systemd
```
```

**Checklist**:
- [ ] TASK-001 created
- [ ] Status: To Do
- [ ] Priority: P0
- [ ] Acceptance criteria clear

---

**TASK-002**: Create E2E Test Script

Create `backlog/tasks/TASK-002-create-e2e-test.md`:

```markdown
# TASK-002: Create E2E Test Script

**Status**: To Do
**Priority**: P0 (Critical)
**Milestone**: MVP
**Labels**: Testing, P0, MVP
**Effort**: 4-6 hours
**Owner**: QA Engineer

---

## Description

No end-to-end test exists to validate the full workflow:
- Pull image from GHCR
- Create VM
- Validate services work
- Delete VM and cleanup

This blocks MVP validation and CI confidence.

---

## Specification Reference

- Spec: .specify/features/mvp-completion/spec.md
- Plan: .specify/features/e2e-testing/plan.md (to be created)

---

## Acceptance Criteria

- [ ] Script created: `scripts/test-e2e.sh`
- [ ] Script executable: `chmod +x scripts/test-e2e.sh`
- [ ] Tests full workflow:
  - [ ] GHCR authentication
  - [ ] Image pull
  - [ ] VM creation
  - [ ] VM network configuration
  - [ ] Service availability (nginx port 80, todo-backend port 8080)
  - [ ] VM stop
  - [ ] VM delete
  - [ ] Resource cleanup (TAP devices, IP allocations)
- [ ] Clear output (show progress, validate each step)
- [ ] Actionable errors (if step fails, explain why and how to fix)
- [ ] Runs on CI (add to .github/workflows/ci.yaml)
- [ ] Documentation: Update README with E2E test instructions

---

## Dependencies

- Blocked by: TASK-001 (services must start first)
- Blocks: MVP validation

---

## Implementation Notes

**Script structure**:
```bash
#!/bin/bash
set -euo pipefail

# Configuration
IMAGE="ghcr.io/daax-dev/nanofuse/base:latest"
VM_NAME="e2e-test-vm"
API_URL="http://127.0.0.1:8080"

# Cleanup function
cleanup() {
    echo "Cleaning up..."
    ./bin/nanofuse vm delete $VM_NAME || true
}
trap cleanup EXIT

# Step 1: Verify daemon running
echo "[1/7] Verifying daemon is running..."
# ...

# Step 2: Pull image
echo "[2/7] Pulling image from GHCR..."
# ...

# Step 3: Create VM
echo "[3/7] Creating VM..."
# ...

# Step 4: Wait for boot
echo "[4/7] Waiting for VM to boot..."
# ...

# Step 5: Validate networking
echo "[5/7] Validating network connectivity..."
# ...

# Step 6: Validate services
echo "[6/7] Validating services..."
# Test nginx
# Test todo-backend
# ...

# Step 7: Cleanup
echo "[7/7] Cleaning up..."
# ...

echo "✅ E2E test passed!"
```
```

**Checklist**:
- [ ] TASK-002 created
- [ ] Blocked by TASK-001 documented
- [ ] Script structure outlined

---

**TASK-003**: Unified Error Handling

Create `backlog/tasks/TASK-003-unified-errors.md`:

```markdown
# TASK-003: Unified Error Handling

**Status**: To Do
**Priority**: P1 (High)
**Milestone**: MVP
**Labels**: Enhancement, P1, MVP
**Effort**: 6-8 hours
**Owner**: Backend Engineer

---

## Description

API errors vary in format:
- Some return JSON with `error` field
- Some return plain text
- Error codes inconsistent
- No context for debugging

Users cannot reliably parse errors or understand how to fix issues.

---

## Specification Reference

- Spec: .specify/features/mvp-completion/spec.md
- Plan: .specify/features/error-handling/plan.md (to be created)

---

## Acceptance Criteria

- [ ] Error schema defined in docs/building/implementation/API_CONTRACT.md
- [ ] Schema includes: code (string), message (string), context (object)
- [ ] All API handlers updated to use schema
- [ ] Error code constants defined (e.g., ERR_VM_NOT_FOUND)
- [ ] CLI parses error schema and displays user-friendly messages
- [ ] Tests updated to validate error responses
- [ ] Documentation updated with error codes and meanings

---

## Dependencies

- Blocked by: None (can start in parallel with TASK-001)
- Blocks: None (nice-to-have for MVP)

---

## Implementation Notes

**Error Schema**:
```json
{
  "error": {
    "code": "ERR_VM_NOT_FOUND",
    "message": "Virtual machine 'test-vm' not found",
    "context": {
      "vm_id": "test-vm",
      "available_vms": ["vm1", "vm2"]
    }
  }
}
```

**Error Codes**:
- ERR_VM_NOT_FOUND
- ERR_IMAGE_NOT_FOUND
- ERR_NETWORK_FAILED
- ERR_FIRECRACKER_FAILED
- ERR_AUTH_FAILED
- ERR_INVALID_REQUEST

**Files to modify**:
- docs/building/implementation/API_CONTRACT.md (add error schema)
- internal/api/errors.go (define error types and codes)
- internal/api/*_handlers.go (update all handlers)
- internal/client/client.go (parse error responses)
- cmd/nanofuse/commands/*.go (display user-friendly errors)
```

**Checklist**:
- [ ] TASK-003 created
- [ ] Error schema defined
- [ ] Can run in parallel with TASK-001

---

### Additional Governance Tasks

**TASK-006**: Initialize .specify/ Structure

```markdown
# TASK-006: Initialize .specify/ Structure

**Status**: In Progress
**Priority**: P0 (Critical)
**Milestone**: Governance
**Labels**: Governance, P0
**Effort**: 2-4 hours
**Owner**: Architect

---

## Description

Spec-driven development requires .specify/ directory structure to organize specifications, plans, and governance artifacts.

---

## Acceptance Criteria

- [ ] .specify/ directories created
- [ ] Constitution moved to .specify/memory/
- [ ] Spec template created
- [ ] Phase 1 foundation spec created (historical)
- [ ] MVP completion spec created
- [ ] Structure documented in README
```

**TASK-007**: Create ADRs for Past Decisions

```markdown
# TASK-007: Create ADRs for Past Decisions

**Status**: To Do
**Priority**: P1 (High)
**Milestone**: Governance
**Labels**: Documentation, P1, Governance
**Effort**: 3-4 hours
**Owner**: Architect

---

## Description

Document architectural decisions made during Phase 1 as Architecture Decision Records (ADRs) for future reference.

---

## Acceptance Criteria

- [ ] ADR-001: Choice of Go as implementation language
- [ ] ADR-002: Ubuntu 24.04 as base image
- [ ] ADR-003: GHCR as image registry
- [ ] ADR-004: Slicer kernel reuse (5.10.240)
- [ ] ADR-005: Unix socket vs TCP for API
- [ ] ADRs stored in backlog/decisions/
```

**Checklist**:
- [ ] All 7 core tasks created in backlog/tasks/
- [ ] Task dependencies documented
- [ ] Priorities assigned (P0, P1)
- [ ] Owners assigned

---

## Step 4: Create ADRs (3-4 hours)

### ADR Template

Create `backlog/decisions/ADR-TEMPLATE.md`:

```markdown
# ADR-{NUMBER}: {Title}

**Status**: Proposed | Accepted | Rejected | Deprecated | Superseded
**Date**: {YYYY-MM-DD}
**Deciders**: {Names or roles}
**Context**: {What prompted this decision}

---

## Decision

{What was decided}

## Rationale

{Why this decision was made}

## Consequences

### Positive
- {Benefit 1}
- {Benefit 2}

### Negative
- {Cost or limitation 1}
- {Cost or limitation 2}

### Neutral
- {Neither good nor bad, just different}

---

## Alternatives Considered

### Alternative 1: {Name}
- Pros: {Benefits}
- Cons: {Drawbacks}
- Reason for rejection: {Why not chosen}

### Alternative 2: {Name}
- Pros: {Benefits}
- Cons: {Drawbacks}
- Reason for rejection: {Why not chosen}

---

## Related Decisions

- Supersedes: ADR-{NUMBER}
- Superseded by: ADR-{NUMBER}
- Related to: ADR-{NUMBER}

---

## Notes

{Additional context, links, references}
```

**Checklist**:
- [ ] ADR template created

---

### Create Core ADRs

**ADR-001**: Choice of Go

Create `backlog/decisions/ADR-001-go-language.md`:

```markdown
# ADR-001: Choice of Go as Implementation Language

**Status**: Accepted
**Date**: 2025-10-30 (Retroactive)
**Deciders**: Development Team, Architect
**Context**: Need to choose programming language for CLI and API daemon

---

## Decision

Use Go (golang) for all CLI and API daemon implementation.

## Rationale

1. **Static Binaries**: Go produces single-file executables with no runtime dependencies
2. **Systems Programming**: Excellent for low-level operations (process management, networking)
3. **Performance**: Compiled language with good performance characteristics
4. **Firecracker Integration**: Existing Go libraries for Firecracker API
5. **Concurrency**: Built-in goroutines for async operations
6. **Tooling**: Excellent standard library, testing framework, build tools

## Consequences

### Positive
- Static binaries easy to distribute (no dependency hell)
- Fast compilation times
- Strong typing catches errors at compile time
- Large ecosystem of libraries
- Good systemd integration

### Negative
- Learning curve for developers unfamiliar with Go
- Verbose error handling (if err != nil pattern)
- Less expressive than some languages (Python, Rust)

### Neutral
- Go modules for dependency management
- Opinionated formatting (gofmt)

---

## Alternatives Considered

### Alternative 1: Python
- Pros: Rapid development, expressive, large ecosystem
- Cons: Runtime dependency, slower execution, packaging complexity
- Reason for rejection: Static binary requirement

### Alternative 2: Rust
- Pros: Memory safety, performance, systems programming
- Cons: Steep learning curve, longer compile times, smaller ecosystem
- Reason for rejection: Team expertise, faster iteration needed

### Alternative 3: Shell Scripts
- Pros: Universal, no compilation
- Cons: Poor error handling, difficult testing, limited type safety
- Reason for rejection: Complexity beyond shell script capabilities

---

## Related Decisions

- Related to: ADR-005 (Unix socket vs TCP - Go supports both well)

---

## Notes

This decision was made early in project (October 2025) and has proven successful. Static binaries simplify deployment, and Go's standard library provided everything needed for VM management.
```

**Checklist**:
- [ ] ADR-001 created

---

**ADR-002**: Ubuntu 24.04 Base

Create `backlog/decisions/ADR-002-ubuntu-base.md`:

```markdown
# ADR-002: Ubuntu 24.04 as Base Image

**Status**: Accepted
**Date**: 2025-10-30 (Retroactive)
**Deciders**: Platform Engineer, Architect
**Context**: Choose base OS for microVM images

---

## Decision

Use Ubuntu 24.04 LTS as the base operating system for all microVM images.

## Rationale

1. **LTS Support**: Long-term support (5 years) provides stability
2. **Package Availability**: Comprehensive apt repositories
3. **Systemd**: Built-in systemd for service management
4. **Familiarity**: Team expertise with Ubuntu
5. **Documentation**: Extensive community documentation
6. **Container Compatibility**: Easy to build from ubuntu:24.04 Docker image

## Consequences

### Positive
- Proven stable base
- Easy package installation (apt)
- Systemd simplifies service management
- Large community for troubleshooting
- Can extend with custom packages

### Negative
- Larger image size than Alpine (100+ MB vs 5 MB)
- More attack surface than minimal distros
- Ubuntu-specific quirks (snap, etc.)

### Neutral
- Cloud-optimized kernel available
- Regular security updates

---

## Alternatives Considered

### Alternative 1: Alpine Linux
- Pros: Minimal size (5 MB), musl libc, security-focused
- Cons: Different package manager (apk), musl incompatibilities, no systemd
- Reason for rejection: Need systemd for complex services

### Alternative 2: Debian
- Pros: Stable, similar to Ubuntu, smaller default install
- Cons: Older packages, less familiar to team
- Reason for rejection: Ubuntu provides newer packages, more documentation

### Alternative 3: Fedora
- Pros: Latest features, Red Hat ecosystem
- Cons: Shorter support cycle, larger images
- Reason for rejection: LTS support more important than cutting edge

---

## Related Decisions

- Related to: ADR-004 (Kernel choice - Ubuntu 24.04 comes with 6.x kernel)

---

## Notes

The choice of Ubuntu 24.04 was validated during implementation. Systemd support was critical for running nginx, todo-backend, and other services. Package availability made it easy to install dependencies.

**Trade-off**: Accepted larger image size (100+ MB) for convenience and stability. For production, could consider distroless or minimal Ubuntu variants.
```

**Checklist**:
- [ ] ADR-002 created

---

**ADR-003**: GHCR Registry

Create `backlog/decisions/ADR-003-ghcr-registry.md`:

```markdown
# ADR-003: GitHub Container Registry (GHCR) for Images

**Status**: Accepted
**Date**: 2025-10-30 (Retroactive)
**Deciders**: SRE, Architect
**Context**: Choose container registry for storing and distributing microVM images

---

## Decision

Use GitHub Container Registry (ghcr.io) as the primary image registry for NanoFuse.

## Rationale

1. **GitHub Integration**: Project hosted on GitHub, seamless integration
2. **Private Repositories**: Free private storage for team members
3. **Authentication**: GitHub PAT-based auth, no separate credentials
4. **CI/CD**: GitHub Actions can push directly with GITHUB_TOKEN
5. **OCI Compatibility**: Full OCI image spec support
6. **Pricing**: Generous free tier for open source and private use

## Consequences

### Positive
- No separate registry account needed
- Authentication uses existing GitHub tokens
- CI/CD integration trivial (GITHUB_TOKEN)
- Team access controlled by GitHub repository permissions
- No additional cost

### Negative
- Tied to GitHub ecosystem (vendor lock-in)
- Rate limiting on unauthenticated pulls
- Less control than self-hosted registry

### Neutral
- Must authenticate to pull private images
- Pull rates counted against GitHub quotas

---

## Alternatives Considered

### Alternative 1: Docker Hub
- Pros: Most popular, widely supported, free tier
- Cons: Aggressive rate limiting, requires separate account, less GitHub integration
- Reason for rejection: Rate limiting problematic, separate credentials

### Alternative 2: Self-Hosted (Harbor)
- Pros: Full control, no vendor lock-in, no rate limits
- Cons: Infrastructure cost, maintenance burden, complexity
- Reason for rejection: Maintenance burden too high for project size

### Alternative 3: AWS ECR
- Pros: Scalable, integrates with AWS, reliable
- Cons: Cost, AWS-specific, requires separate credentials
- Reason for rejection: Cost and complexity not justified

---

## Related Decisions

- Related to: CI/CD implementation (GitHub Actions pushing to GHCR)

---

## Notes

GHCR has worked well for the project. Authentication with GitHub PATs is simple, and CI/CD integration is seamless. The free tier is more than sufficient for current usage.

**Future Consideration**: If moving away from GitHub or needing multi-cloud, could add additional registries. OCI spec compliance makes this reversible.
```

**Checklist**:
- [ ] ADR-003 created

---

**ADR-004**: Slicer Kernel Reuse

Create `backlog/decisions/ADR-004-slicer-kernel.md`:

```markdown
# ADR-004: Reuse Slicer's Proven Kernel (5.10.240)

**Status**: Accepted
**Date**: 2025-10-30 (Retroactive)
**Deciders**: Platform Engineer, Architect
**Context**: Choose kernel version for microVMs (compile custom vs use proven)

---

## Decision

Use Slicer's proven Linux kernel 5.10.240 bundled in images, deferring custom kernel compilation to future learning phase.

## Rationale

1. **Proven Stability**: Slicer uses 5.10.240 in production, well-tested
2. **Firecracker Compatibility**: Known to work with Firecracker
3. **Learning Focus**: Project is learning Firecracker, not kernel compilation
4. **Pragmatism**: Reduces variables (if issue arises, unlikely to be kernel)
5. **Time to MVP**: Faster to reuse than compile custom
6. **Future Learning**: Can compile custom kernel later as learning exercise

## Consequences

### Positive
- Reduced risk (proven kernel)
- Faster to MVP (no kernel compilation)
- Clear separation of concerns (focus on platform, not kernel)
- Can upgrade kernel later without platform changes

### Negative
- Miss learning opportunity (kernel configuration)
- Slight dependency on Slicer (kernel availability)
- 5.10 is older (6.x has newer features)

### Neutral
- Kernel bundled in image (VM-specific, not host-shared)
- Can experiment with 6.x later

---

## Alternatives Considered

### Alternative 1: Compile Custom Kernel (6.x)
- Pros: Latest features, full control, learning opportunity
- Cons: Time-consuming, more variables, debugging complexity
- Reason for rejection: Too many unknowns for initial implementation

### Alternative 2: Use Host Kernel
- Pros: No bundling needed, single kernel version
- Cons: Firecracker doesn't support this, VMs need own kernel
- Reason for rejection: Not supported by Firecracker architecture

### Alternative 3: Ubuntu's Default Kernel (6.x)
- Pros: Newer features, maintained by Ubuntu
- Cons: Untested with Firecracker, potential boot issues
- Reason for rejection: Unknown compatibility, risk too high for MVP

---

## Related Decisions

- Related to: ADR-002 (Ubuntu 24.04 base - has 6.x kernel available for future)
- May be superseded by: Future decision to compile custom kernel (6.x)

---

## Notes

This decision proved correct during implementation. Services failing in VMs was NOT a kernel issue (it was systemd init parameter). Using proven kernel eliminated kernel as variable.

**Future**: Once platform stable, can experiment with:
- Custom 6.x kernel compilation
- Kernel optimizations for Firecracker
- Security hardening (disable unused features)

This remains a **two-way door decision** - can change kernel later without platform changes.
```

**Checklist**:
- [ ] ADR-004 created

---

### Complete ADR Checklist

**Checklist**:
- [ ] ADR template created
- [ ] ADR-001 (Go language) created
- [ ] ADR-002 (Ubuntu base) created
- [ ] ADR-003 (GHCR registry) created
- [ ] ADR-004 (Slicer kernel) created
- [ ] ADR-005 (Unix socket vs TCP) - OPTIONAL, exists in BIG-IDEAS.md
- [ ] All ADRs stored in backlog/decisions/

---

## Step 5: Document Golden Path (2-3 hours)

Create `docs/building/GOLDEN_PATH.md`:

```markdown
# NanoFuse Golden Path - Standard Development Workflow

**Purpose**: This document defines the ONE standard way to develop features, fix bugs, and contribute to NanoFuse.

**Authority**: This Golden Path is mandated by the project constitution (.specify/memory/constitution.md).

---

## For New Features

### 1. Create Specification

```bash
# Use jp-spec-kit to create spec
/jpspec.specify "Feature: <description>"

# OR manually create spec from template
cp .specify/templates/spec-template.md .specify/features/<feature-name>/spec.md
# Edit spec.md
```

**Output**: `.specify/features/<feature-name>/spec.md`

**Review Gate**: Human authority or architect must approve spec before implementation

---

### 2. Generate Implementation Plan

```bash
# Use jp-spec-kit to create plan
/jpspec.plan

# OR manually create plan
# Create .specify/features/<feature-name>/plan.md
```

**Output**: `.specify/features/<feature-name>/plan.md`

**Review Gate**: Architect reviews for architectural alignment

---

### 3. Create Backlog Tasks

```bash
# Create task in backlog
backlog task create \
  "Implement feature: <name>" \
  --label "Feature,P1" \
  --milestone "Phase X" \
  --description "See spec: .specify/features/<feature-name>/spec.md"

# OR manually create task file
# Create backlog/tasks/TASK-XXX-<name>.md
```

**Output**: Task in `backlog/tasks/`

**Review Gate**: Human prioritizes and assigns milestone

---

### 4. Implement with TDD

```bash
# Step 1: Write failing test (RED)
# Create test file, write test that demonstrates requirement
go test ./... # Should FAIL

# Step 2: Implement minimal code to pass test (GREEN)
# Write implementation
go test ./... # Should PASS

# Step 3: Refactor
# Improve code quality while keeping tests green
go test ./... # Still PASS

# Step 4: Commit
git add .
git commit -m "feat: <description>

Implements spec: .specify/features/<feature-name>/spec.md
Closes: TASK-XXX"
```

**Review Gate**: All tests must pass, code review required

---

### 5. Validate with CI

```bash
# Push to feature branch
git push origin feature/<feature-name>

# Create PR
gh pr create --title "feat: <description>" --body "See spec.md"

# CI will run:
# - go build (compilation check)
# - go test (unit tests)
# - golangci-lint (code quality)
# - gosec (security scan)
# - E2E tests (if applicable)
```

**Review Gate**: All CI checks must pass, PR approved by human

---

### 6. Close Backlog Task

```bash
# After merge, close task
backlog task archive TASK-XXX

# Update spec status
# Edit .specify/features/<feature-name>/spec.md
# Change status to "Implemented"
```

---

## For Bug Fixes

### 1. Create Bug Task

```bash
# Create task
backlog task create \
  "Bug: <description>" \
  --label "Bug,P0" \
  --milestone "Current Sprint"
```

---

### 2. Write Failing Test (RED)

```bash
# Create test that demonstrates bug
# Test should FAIL (proving bug exists)
go test ./... # FAIL expected
```

---

### 3. Fix Bug (GREEN)

```bash
# Implement fix
# Test should PASS
go test ./... # PASS
```

---

### 4. Commit and PR

```bash
git commit -m "fix: <description>

Closes: TASK-XXX"

git push origin fix/<bug-name>
gh pr create
```

---

### 5. Close Task

```bash
backlog task archive TASK-XXX
```

---

## For Documentation Updates

### 1. Update Docs

```bash
# Edit documentation files
# User docs: docs/
# Dev docs: docs/building/
```

---

### 2. Validate

```bash
# Check for broken links, formatting
# Run markdown linter if available
```

---

### 3. Commit

```bash
git commit -m "docs: <description>"
git push origin docs/<topic>
```

**Note**: Documentation PRs can be self-merged after CI lint passes

---

## For Releases

### Go Binary Release

```bash
# Run release script (auto-increments version)
./release.sh

# CI will:
# - Create git tag (v0.0.X)
# - Build binaries
# - Create GitHub release
# - Upload artifacts
```

---

### Docker Image Release

```bash
# Run image release script
./image-release.sh

# CI will:
# - Create git tag (image-v0.0.X)
# - Build Docker image
# - Push to GHCR with version tag
# - Update latest tag
```

---

## Quality Gates (Required Before Merge)

### Entry Gates (Before Starting)
- [ ] Specification exists and approved (for features)
- [ ] Backlog task created
- [ ] Dependencies identified

### Development Gates (While Working)
- [ ] Tests written (RED)
- [ ] Tests pass (GREEN)
- [ ] Code refactored
- [ ] All tests still pass

### Integration Gates (Before Merge)
- [ ] All CI checks pass
  - [ ] Build succeeds
  - [ ] Unit tests pass
  - [ ] Lint checks pass
  - [ ] Security scans pass
  - [ ] Integration tests pass (if applicable)
  - [ ] E2E tests pass (if applicable)
- [ ] Code reviewed and approved
- [ ] Documentation updated

### Exit Gates (After Merge)
- [ ] Backlog task closed
- [ ] Spec status updated (if applicable)
- [ ] No regressions detected

---

## Tools

**Build**:
```bash
mage all        # Build all binaries
mage cli        # Build CLI only
mage daemon     # Build daemon only
```

**Test**:
```bash
mage test           # Unit tests
mage testIntegration # Integration tests
mage testE2E        # End-to-end tests
mage testAll        # All tests
```

**Quality**:
```bash
mage lint           # Code quality checks
mage securityCheck  # Security scans
mage ci             # Full CI suite locally
```

**Backlog**:
```bash
backlog task create "<description>" --label "<label>" --milestone "<milestone>"
backlog task list
backlog task archive <task-id>
backlog browser  # Open web UI
```

---

## Workflow Diagram

```
┌─────────────────┐
│ Human Authority │ Approves specs, unblocks decisions
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Create Spec     │ /jpspec.specify OR manual
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Review Spec     │ Architect validates alignment
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Create Plan     │ /jpspec.plan OR manual
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Create Tasks    │ Backlog
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Implement (TDD) │ RED → GREEN → REFACTOR
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ CI Validation   │ Build, Test, Lint, Security
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Code Review     │ Human approval
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Merge to Main   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Close Task      │ Backlog archive
└─────────────────┘
```

---

## Exceptions

**Hotfixes** (Critical production bugs):
- Can skip spec creation
- Must still create backlog task
- Must write test demonstrating bug
- Expedited review process
- Post-mortem required

**Documentation-Only**:
- No spec required
- Create backlog task optional
- Self-merge after lint passes

**Experimental**:
- Create branch prefixed with `experiment/`
- No spec required
- Not merged to main
- Deleted after learning complete

---

## Getting Help

**Questions about workflow**: Ask in team chat or GitHub Discussions
**Questions about architecture**: Ping architect
**Questions about priorities**: Ask human authority
**Questions about tooling**: Check tool documentation (mage -l, backlog --help)

---

## Summary

**The Golden Path is**:
1. Spec first (WHAT and WHY)
2. Plan second (HOW)
3. Task created (track work)
4. Implement with TDD (RED → GREEN → REFACTOR)
5. CI validates (automated quality gates)
6. Human reviews (final approval)
7. Merge and close task

**No exceptions for feature work. Consistency enables automation and quality.**
```

**Checklist**:
- [ ] GOLDEN_PATH.md created
- [ ] Workflow diagram included
- [ ] Quality gates defined
- [ ] Tools documented

---

## Step 6: Update Documentation (1-2 hours)

### Update README.md

Add section on governance:

```markdown
## Development Process

NanoFuse follows a strict spec-driven development process:

1. **Specification First**: All features start with a technology-agnostic specification
2. **Backlog Required**: All work tracked in backlog/ before implementation
3. **Test-Driven Development**: Write tests before implementation (RED → GREEN → REFACTOR)
4. **Quality Gates**: CI enforces build, test, lint, and security checks
5. **Code Review**: All PRs require approval before merge

See [Golden Path](docs/building/GOLDEN_PATH.md) for the standard development workflow.
```

---

### Create Governance Summary

Create `docs/GOVERNANCE.md`:

```markdown
# NanoFuse Governance

This document describes the governance structure for the NanoFuse project.

## Authority Hierarchy

1. **Constitution** (`.specify/memory/constitution.md`) - SUPREME AUTHORITY
2. **Specifications** (`.specify/features/*/spec.md`) - Feature requirements
3. **Implementation Plans** (`.specify/features/*/plan.md`) - Technical approach
4. **Backlog Tasks** (`backlog/tasks/`) - Work tracking
5. **Code** (`internal/`, `cmd/`) - Implementation

## Development Process

All development follows the **Golden Path**: [docs/building/GOLDEN_PATH.md](docs/building/GOLDEN_PATH.md)

## Decision Making

- **One-Way Doors** (irreversible): Require human authority approval
- **Two-Way Doors** (reversible): Engineers can decide
- **Architecture Decisions**: Documented as ADRs in `backlog/decisions/`

## Quality Gates

Before merge to main:
- ✅ Specification approved (for features)
- ✅ Tests pass (unit + integration + E2E)
- ✅ Lint checks pass
- ✅ Security scans pass
- ✅ Code reviewed and approved
- ✅ Documentation updated

## Roles

- **Human Authority**: Approves specs, unblocks one-way doors
- **Architect**: Validates architectural alignment, creates specs/ADRs
- **Engineers**: Implement from specs, write tests
- **CI/CD**: Enforces quality gates

## Tools

- **jp-spec-kit**: Specification-driven development
- **Backlog.md**: Task management via MCP server
- **Mage**: Build and test automation
- **GitHub Actions**: CI/CD pipeline

## Getting Started

1. Read [Constitution](.specify/memory/constitution.md)
2. Read [Golden Path](docs/building/GOLDEN_PATH.md)
3. Set up tools (jp-spec-kit, backlog, mage)
4. Review existing specs and ADRs
5. Pick a task from backlog
6. Follow the Golden Path

## References

- Full Analysis: [docs/building/ARCHITECT_ELEVATOR_ANALYSIS.md](docs/building/ARCHITECT_ELEVATOR_ANALYSIS.md)
- Executive Summary: [docs/building/EXECUTIVE_SUMMARY.md](docs/building/EXECUTIVE_SUMMARY.md)
```

**Checklist**:
- [ ] README.md updated with governance section
- [ ] docs/GOVERNANCE.md created
- [ ] Links to key documents included

---

## Validation

### Phase 0 Complete Checklist

- [ ] .specify/ structure created
- [ ] Constitution in correct location
- [ ] Spec template created
- [ ] Phase 1 foundation spec created (historical)
- [ ] MVP completion spec created
- [ ] Backlog populated with 7+ tasks
- [ ] ADRs created for 4+ key decisions
- [ ] Golden Path documented
- [ ] README and GOVERNANCE.md updated
- [ ] Team trained on new workflow

---

### Success Criteria

- **Governance Established**: Process documented and enforced
- **Work Tracked**: All current and future work in backlog
- **Decisions Documented**: ADRs capture architectural reasoning
- **Path Clear**: Developers know standard workflow
- **Authority Defined**: Roles and responsibilities clear

---

## Next Steps (After Phase 0)

1. **Fix MVP Blockers** (Phase 1):
   - TASK-001: Fix systemd init parameter (2-4 hours)
   - TASK-002: Create E2E test script (4-6 hours)
   - TASK-003: Unified error handling (6-8 hours)

2. **Validate MVP** (1-2 days):
   - Run E2E test
   - Demo to stakeholders
   - Decide on Phase 2 go/no-go

3. **Continue with Governance**:
   - Create specs for all new features
   - Track all work in backlog
   - Document decisions as ADRs
   - Follow Golden Path

---

**Document**: Phase 0 Governance Checklist
**Status**: READY TO EXECUTE
**Duration**: 1-2 days focused effort
**Outcome**: Sustainable development process established
