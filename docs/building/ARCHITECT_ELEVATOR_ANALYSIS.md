# NanoFuse Architectural Analysis
## Applying the Architect Elevator Framework

**Date**: 2025-11-23
**Architect**: Software Architect Enhanced Agent
**Framework**: Gregor Hohpe's Architect Elevator + Platform Strategy + EIP
**Status**: CRITICAL - Strategic realignment required

---

## Executive Summary

NanoFuse has achieved significant technical milestones (networking works, VMs boot, binaries compile) but is operating **without architectural governance**. The project exhibits classic symptoms of "engine room-only" development: extensive code exists but lacks strategic framing, specification-driven development, or task management discipline.

**Critical Gap**: The project has been building in the engine room without riding the elevator to validate strategic alignment or establish governance frameworks.

**Recommendation**: Immediate implementation of governance structures (spec-driven development, backlog management) before any new feature work. Current state is **technically functional but organizationally unsustainable**.

---

# Part I: Strategic Framing (Penthouse View)

## 1.1 Business Value Alignment

### Strategic Objectives (Stated)
1. **Learning Goal**: Rebuild Firecracker-based microVM system from first principles
2. **Production Foundation**: Enable Trigger.dev dual-environment workloads
3. **Platform Capability**: "Pull and run like Slicer" - simple microVM management

### Value Stream Analysis

**Current Value Delivery**:
- ✅ **Technical Learning**: Achieved - networking, VM lifecycle, image building understood
- ⚠️ **Production Readiness**: Partial - services not starting, no E2E validation
- ❌ **Platform Usability**: Blocked - cannot demonstrate full workflow

**Value at Risk**:
```
Investment: ~2-3 weeks of development effort
Technical Debt: High - 50+ documentation files, unclear which reflect reality
Organizational Debt: CRITICAL - no specs, no task tracking, no governance
Risk of Rework: HIGH - without specifications, changes are ad-hoc
```

### Selling Options Analysis

**Current Position**: High volatility (learning project, evolving requirements) demands architectural investment.

**Options on the Table**:

| Option | Strike Price | Volatility | Keep or Sell? |
|--------|-------------|-----------|---------------|
| **Build custom kernel** | 3-5 days effort | Medium | SELL - Using proven kernel (5.10.240) deferred correctly |
| **Multi-architecture support** | 2-3 days effort | Low | KEEP - Future ARM64 need is certain |
| **Snapshot/resume capability** | 5-7 days effort | Medium | KEEP - Core to "fast cold start" value prop |
| **Trigger.dev integration** | 10-14 days effort | High | KEEP - Business justification unclear, defer |
| **Spec-driven development** | 1-2 days setup | Zero | EXECUTE NOW - Reduces all future volatility |
| **Web UI for API** | 5-7 days effort | Medium | SELL - CLI sufficient for MVP |

**Recommendation**: Stop accumulating options. Execute the low-volatility, high-impact decision (spec-driven development) immediately.

### Architectural Technical Debt Assessment

**Debt Categories**:

1. **Governance Debt (CRITICAL)**:
   - No .specify/ directory - spec-driven development not implemented
   - Backlog exists but empty (0 tasks found)
   - No Architecture Decision Records (ADRs)
   - Constitution exists but not enforced (specs requirement ignored)

2. **Documentation Debt (HIGH)**:
   - 50+ markdown files in docs/building/
   - Multiple "status reports" with conflicting information
   - Unclear which docs represent current reality vs. planning artifacts
   - No single source of truth

3. **Testing Debt (MEDIUM)**:
   - Unit tests exist and pass (8/8)
   - Integration test broken (CreateSnapshot test failing)
   - No E2E validation (services not starting)
   - Test coverage unknown (no coverage reporting in CI)

4. **Implementation Debt (MEDIUM)**:
   - 15 TODO/FIXME markers in codebase
   - Stub implementations (snapshot creation)
   - Services failing in VMs (systemd init parameter issue suspected)

**Debt Service Cost**: Without governance, every new feature adds to technical debt. Current trajectory is unsustainable.

---

## 1.2 Phase Completion Reality Check

### Claimed vs. Actual Status

**Documentation Claims**:
- "Phase 1 Complete" (EXECUTION_PLAN.md)
- "Networking Implementation Complete" (READY-TO-TEST.md)
- "Phase 1C/1D: Services Fix Plan" (PHASE1CD_COMPREHENSIVE_PLAN.md)

**Reality** (Evidence-Based):

| Phase | Claimed Status | Evidence | Actual Status |
|-------|---------------|----------|---------------|
| **Phase 0: Architecture** | ✅ Complete | Docs exist (40+ files) | ⚠️ Partial - No specs, no ADRs |
| **Phase 1A: CLI** | ✅ Complete | Binary compiles (8.5MB) | ✅ Builds, untested E2E |
| **Phase 1B: API** | ✅ Complete | Binary compiles (9.0MB) | ✅ Builds, partial implementation |
| **Phase 1C: Base Image** | ✅ Complete | Image builds successfully | ⚠️ Services not starting |
| **Phase 1D: Networking** | ✅ Complete | VMs pingable (0.4ms) | ✅ Proven working |
| **Phase 1E: Services** | ⏳ In Progress | Nginx/backend failing | ❌ Blocked on init parameter |
| **Phase 2: Snapshot** | 📋 Planned | CreateSnapshot test fails | ❌ Not started |
| **Phase 3: Trigger.dev** | 📋 Planned | No code exists | ❌ Not started |

**Honest Assessment**: Phase 1 is **70% complete** (not 100%). The gap between claimed and actual status indicates process failure.

### Root Cause: No Testable Boundaries

**Pattern Observed**: Phases were "completed" by agents without validation gates.

**Missing Quality Gates**:
- ✅ Code compiles → **Present** (Go build succeeds)
- ❌ Tests pass → **Missing** (integration test fails, ignored)
- ❌ E2E workflow validated → **Missing** (services don't start)
- ❌ Specification approved → **Missing** (no specs exist)
- ❌ Backlog task closed → **Missing** (no tasks tracked)

**Architect Elevator Insight**: The project operated entirely in the engine room. No one rode the elevator to the penthouse to ask "Does this actually deliver value?" or "Can we prove it works?"

---

# Part II: Decision Model Application (Strategy View)

## 2.1 Platform Strategy Assessment

### Is NanoFuse a Coherent Platform?

**Platform Strategy Framework (7 C's)**:

| Criterion | Score | Evidence | Gap |
|-----------|-------|----------|-----|
| **Clarity** | 6/10 | Goal stated clearly | No specs define boundaries |
| **Consistency** | 7/10 | Go codebase uniform | Documentation inconsistent |
| **Compliance** | 3/10 | Constitution ignored | No spec-driven development |
| **Composability** | 5/10 | CLI + API + Images | Integration unproven |
| **Coverage** | 4/10 | VM lifecycle partial | Snapshot/resume missing |
| **Consumption** | 2/10 | CLI exists | E2E workflow broken |
| **Credibility** | 3/10 | Networking proven | Services failing destroys trust |

**Overall Platform Maturity**: **4.3/10** - Early-stage platform with good bones, poor discipline.

**Platform Strategy Verdict**: NanoFuse has potential but is not yet a **platform as a product (PaaP)**. It's a collection of components without proven integration.

### Fruit Salad vs. Fruit Basket

**Current State**: **Fruit Salad** (unstandardized)
- Multiple documentation formats (planning/, implementation/, reports/, diagnostics/)
- No standard workflow (some phases have specs, most don't)
- Ad-hoc decision making (no ADRs)
- Inconsistent testing (some components tested, some not)

**Required State**: **Fruit Basket** (highly opinionated)
- **ONE way to start a feature**: /jpspec.specify
- **ONE way to track work**: backlog tasks
- **ONE way to document decisions**: ADRs in backlog/decisions/
- **ONE way to validate completion**: E2E test passes

**Gap**: The project needs to transition from flexibility to standardization. This is a one-way door decision.

### Build Abstractions Not Illusions

**Current Abstractions**:

✅ **Good Abstractions**:
- `internal/network/` - Clean separation of TAP/bridge/NAT logic
- `internal/firecracker/` - Isolates Firecracker API
- `internal/client/` - Type-safe API client

⚠️ **Leaky Abstractions**:
- VM creation claims success but services don't start (hides systemd init failure)
- Image pull succeeds but boot fails (kernel/rootfs mismatch hidden until boot)
- Snapshot API exists but returns "not implemented" (illusion of capability)

❌ **Missing Abstractions**:
- No unified error handling (errors vary in format/detail)
- No observability layer (cannot introspect VM state beyond console logs)
- No resource cleanup orchestration (TAP devices, snapshots, logs - manual cleanup)

**Recommendation**: Expose failure modes explicitly. A VM that boots but has no services is **worse** than a VM that fails fast with clear error.

---

## 2.2 One-Way vs. Two-Way Door Decisions

### One-Way Doors (Irreversible - Decide Now)

**Already Committed** (Correct Decisions):
1. ✅ **Go as implementation language** - Static binaries, excellent for CLI/API
2. ✅ **Ubuntu 24.04 base** - Modern systemd, package availability
3. ✅ **GHCR as registry** - Private, authenticated, GitHub-integrated
4. ✅ **Slicer kernel reuse** - Pragmatic (5.10.240 proven stable)

**Pending Critical Decisions** (Must Decide Now):

1. **Spec-Driven Development Adoption**
   - **Decision**: Implement jp-spec-kit workflow for ALL future work
   - **Why One-Way**: Once adopted, reverting would invalidate all specs/plans
   - **Cost of Delay**: Every day without specs increases rework risk
   - **Recommendation**: ✅ **COMMIT NOW**

2. **Backlog-First Task Management**
   - **Decision**: ALL work tracked in backlog/ before implementation
   - **Why One-Way**: Task history becomes organizational memory
   - **Cost of Delay**: Untracked work cannot be audited or prioritized
   - **Recommendation**: ✅ **COMMIT NOW**

3. **Constitution as Supreme Authority**
   - **Decision**: Enforce .claude/constitution.md rules strictly
   - **Why One-Way**: Governance requires consistency; partial enforcement fails
   - **Cost of Delay**: Repository entropy increases daily
   - **Recommendation**: ✅ **COMMIT NOW**

### Two-Way Doors (Reversible - Can Defer)

**Correctly Deferred**:
1. ✅ **Web UI** - CLI sufficient for MVP, can add later
2. ✅ **Custom kernel compilation** - Using proven kernel, can experiment later
3. ✅ **Multi-architecture (ARM64)** - x86_64 first, expand when validated

**Should Defer Further**:
1. **Trigger.dev Integration** (Phase 3)
   - **Rationale**: Phase 1 (base platform) not proven yet
   - **Risk of Premature Optimization**: Building for use case before platform stable
   - **Recommendation**: ⏸️ **DEFER until Phase 1 validated**

2. **S3 Snapshot Storage**
   - **Rationale**: Local snapshot/restore must work first
   - **Cost**: Additional integration complexity
   - **Recommendation**: ⏸️ **DEFER until Phase 2 complete**

3. **Advanced Networking (WireGuard, etc.)**
   - **Rationale**: Basic NAT/bridge working, additional modes can wait
   - **Risk**: Scope creep before core validated
   - **Recommendation**: ⏸️ **DEFER until production use case demands it**

---

# Part III: Architectural Blueprint (Engine Room View)

## 3.1 Current Architecture Map

### Component Inventory (What Actually Exists)

```
NanoFuse System Architecture (As-Built)
========================================

┌─────────────────────────────────────────────────────────────────┐
│ Host System (Linux, KVM required)                               │
│                                                                  │
│  ┌────────────────────────────────────────────────────────┐    │
│  │ User Space                                              │    │
│  │                                                          │    │
│  │  ┌──────────────┐         ┌──────────────────────────┐ │    │
│  │  │ nanofuse CLI │────────▶│ nanofused API Daemon     │ │    │
│  │  │  (8.5MB)     │  HTTP   │   (9.0MB)                │ │    │
│  │  │              │  or Unix│                           │ │    │
│  │  │  Commands:   │  Socket │  Endpoints:              │ │    │
│  │  │  - image pull│         │  - POST /images/pull     │ │    │
│  │  │  - vm run    │         │  - POST /vms             │ │    │
│  │  │  - vm stop   │         │  - PUT /vms/:id/stop     │ │    │
│  │  │  - vm status │         │  - GET /vms/:id          │ │    │
│  │  │  - vm logs   │         │  - POST /snapshots       │ │    │
│  │  └──────────────┘         │  - PUT /snapshots/resume │ │    │
│  │                            │                           │ │    │
│  │                            │  State:                  │ │    │
│  │                            │  - SQLite DB             │ │    │
│  │                            │  - /var/lib/nanofuse/    │ │    │
│  │                            └──────────┬───────────────┘ │    │
│  │                                       │                  │    │
│  │  ┌────────────────────────────────────┴────────────────┐│    │
│  │  │ Network Infrastructure (internal/network/)          ││    │
│  │  │  - nanofuse0 Bridge (172.16.0.1/24)                 ││    │
│  │  │  - IPAM (172.16.0.10-254)                           ││    │
│  │  │  - NAT via iptables                                 ││    │
│  │  │  - TAP devices per VM                               ││    │
│  │  └─────────────────────────────────────────────────────┘│    │
│  │                                       │                  │    │
│  └───────────────────────────────────────┼──────────────────┘    │
│                                          │                       │
│  ┌───────────────────────────────────────▼──────────────────┐   │
│  │ Firecracker Processes                                     │   │
│  │                                                            │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐      │   │
│  │  │ MicroVM 1   │  │ MicroVM 2   │  │ MicroVM n   │      │   │
│  │  │ 172.16.0.10 │  │ 172.16.0.11 │  │ 172.16.0.n  │      │   │
│  │  │             │  │             │  │             │      │   │
│  │  │ Ubuntu 24.04│  │ Ubuntu 24.04│  │ Custom Image│      │   │
│  │  │ + systemd   │  │ + nginx     │  │ + app       │      │   │
│  │  │             │  │ + todo-app  │  │             │      │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘      │   │
│  │                                                            │   │
│  │  Status: Boots ✅  Network ✅  Services ❌               │   │
│  └────────────────────────────────────────────────────────────┘   │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘

External Dependencies:
- GHCR (ghcr.io/jpoley/nanofuse/base:*) - Image registry
- GitHub Actions - CI/CD pipeline
- Docker - Image building (build-time only)
```

### Integration Patterns (EIP Taxonomy)

**Message Endpoints**:
- **CLI → API**: Request-Reply pattern (HTTP REST or Unix socket)
- **API → Firecracker**: Command Message (one-way process invocation)
- **API → Database**: Synchronous query (SQLite local storage)

**Messaging Channels**:
- **Control Plane**: HTTP/1.1 REST API (no message broker - direct invocation)
- **Data Plane**: File system (rootfs, console logs, snapshots)
- **Observability**: Console logs to files (no structured logging broker)

**Message Routing**:
- **Content-Based Router**: API endpoints route by HTTP path/method
- **Message Filter**: VM ID validation, image existence checks
- **Recipient List**: Not applicable (single-host architecture)

**Message Transformation**:
- **Data Format Transformer**: JSON ↔ Go structs (encoding/json)
- **Content Enricher**: API adds network config to VM creation request
- **Normalizer**: CLI abstracts "default" to full GHCR image reference

**Process Orchestration**:
- **Process Manager**: API daemon manages Firecracker process lifecycle
- **Routing Slip**: VM creation workflow (allocate IP → create TAP → start Firecracker)
- **Compensation**: Partial - TAP cleanup on VM deletion (no full transaction)

**Critical Gaps in Integration Patterns**:

❌ **Missing: Message Translator for Errors**
- Error formats vary (Go errors, HTTP status codes, Firecracker exit codes)
- No unified error schema for CLI consumption
- **Recommendation**: Define error contract in API_CONTRACT.md

❌ **Missing: Idempotent Receiver**
- VM creation not idempotent (duplicate IDs cause failure)
- No "create or retrieve" semantic
- **Recommendation**: Add idempotency key support

❌ **Missing: Dead Letter Channel**
- Failed VM starts leave orphaned resources (TAP devices, IP allocations)
- No cleanup queue or retry mechanism
- **Recommendation**: Implement resource cleanup reconciler

❌ **Missing: Event-Driven Consumer**
- API uses polling (check VM status via Firecracker API)
- No event subscription (VM lifecycle events)
- **Recommendation**: Acceptable for MVP, consider for scale

---

## 3.2 Golden Path Definition

### Current State: No Golden Path Exists

**Evidence**: Documentation shows 4+ different workflows:
1. EXECUTION_PLAN.md - Phased approach
2. READY-TO-TEST.md - Network testing
3. PHASE1CD_COMPREHENSIVE_PLAN.md - Service debugging
4. COMPREHENSIVE_TESTING_PLAN.md - Layer-by-layer validation

**Problem**: Developers have no single "blessed path" to:
- Add a new feature
- Fix a bug
- Test changes
- Deploy to production

### Proposed Golden Path

**For New Features**:

```bash
# 1. Create specification (technology-agnostic)
/jpspec.specify "Feature: <description>"
# Outputs: .specify/features/<branch>/spec.md

# 2. Review and approve spec (HUMAN GATE)
# - Read spec.md
# - Validate success criteria are measurable
# - Approve or request clarification

# 3. Generate implementation plan
/jpspec.plan
# Outputs: plan.md, tasks.md, contracts/

# 4. Create backlog tasks
backlog task create "Implement feature: <name>" \
  --label "Feature" \
  --milestone "Phase X"

# 5. Implement with TDD
/jpspec.implement
# Outputs: Code + tests

# 6. Validate (QA, security, docs)
/jpspec.validate

# 7. Merge to main (CI validates)
git push origin <branch>
# CI runs: build, test, lint, security scan
```

**For Bug Fixes**:

```bash
# 1. Create bug task in backlog
backlog task create "Bug: <description>" --label "Bug"

# 2. Write failing test (RED)
# - Unit test or integration test
# - Demonstrates bug

# 3. Fix implementation (GREEN)
# - Minimal fix to pass test

# 4. Refactor if needed
# - Improve code quality while tests pass

# 5. Close backlog task
backlog task archive <task-id>

# 6. Merge via PR
git push origin <branch>
```

**For Testing Changes**:

```bash
# Unit tests
mage test

# Integration tests (requires KVM)
mage testIntegration

# E2E test (full workflow)
./scripts/test-e2e.sh  # To be created

# Validate specific component
./scripts/validate-<component>.sh
```

**For Releases**:

```bash
# Go binary release
./release.sh  # Auto-increments version

# Docker image release
./image-release.sh  # Auto-increments image version

# Verify in CI
gh run list --workflow=release.yaml
```

### Golden Path Enforcement

**Quality Gates** (Required before merge):
1. ✅ Specification exists and approved
2. ✅ All tests pass (unit + integration)
3. ✅ Lint checks pass (golangci-lint)
4. ✅ Security scans pass (no critical vulnerabilities)
5. ✅ Documentation updated
6. ✅ Backlog task closed

**Tool Support**:
- **Mage**: Enforces build consistency (mage ci runs all checks)
- **GitHub Actions**: Enforces CI gates (blocks merge on failure)
- **Backlog.md**: Enforces task tracking (all work visible)
- **jp-spec-kit**: Enforces specification discipline (no code without spec)

---

## 3.3 What MUST Be Completed to Prove Value

### Minimum Viable Platform (MVP)

**Definition**: The smallest complete workflow that demonstrates NanoFuse value proposition.

**Value Proposition**: "Pull and run microVMs like Slicer with fast snapshot/resume for Trigger.dev workloads"

**MVP Scope** (Must Work End-to-End):

```
1. Image Management
   ✅ Pull base image from GHCR (ghcr.io/jpoley/nanofuse/base:latest)
   ✅ Authenticate with GitHub token
   ✅ Store locally for reuse

2. VM Lifecycle
   ✅ Create VM from image
   ✅ Start VM with networking
   ❌ Services run inside VM (nginx, custom app)  ← BLOCKING
   ✅ Stop VM gracefully
   ✅ Delete VM and cleanup resources

3. Networking
   ✅ VM gets IP address (172.16.0.x)
   ✅ Host can ping VM
   ❌ Services accessible on ports (80, 8080)  ← BLOCKING
   ⏳ VM can reach internet (untested)

4. Snapshot/Resume (Phase 2 - Deferred)
   ❌ Create snapshot of running VM
   ❌ Resume from snapshot
   ❌ Sub-2-second cold start

5. Platform Usability
   ✅ CLI commands work
   ✅ API responds
   ❌ Full workflow documented and tested  ← BLOCKING
   ❌ Error messages actionable  ← PARTIAL
```

**MVP Blockers** (Must Fix to Ship):

1. **Services Not Starting in VMs** (CRITICAL)
   - **Symptom**: nginx, todo-backend fail to start
   - **Hypothesis**: Missing `init=/lib/systemd/systemd` kernel parameter
   - **Evidence**: Console logs show "[FAILED] Failed to start nginx.service"
   - **Fix Required**: Add init parameter to kernel args in vm_handlers.go
   - **Validation**: curl http://172.16.0.10:80 returns HTML

2. **E2E Workflow Not Validated** (HIGH)
   - **Symptom**: No end-to-end test script
   - **Gap**: Cannot prove full workflow works
   - **Fix Required**: Create scripts/test-e2e.sh
   - **Validation**: Script runs green start-to-finish

3. **Error Handling Inconsistent** (MEDIUM)
   - **Symptom**: Some errors clear, some cryptic
   - **Gap**: No unified error contract
   - **Fix Required**: Define error schema in API_CONTRACT.md
   - **Validation**: All API errors include: code, message, context

**MVP Timeline Estimate**:
- Fix services (1): 2-4 hours (update kernel args, rebuild, test)
- E2E test script (2): 4-6 hours (write script, validate all paths)
- Error handling (3): 6-8 hours (define schema, update all handlers)

**Total MVP Completion**: 1-2 days focused effort

### Production-Ready Additions (Post-MVP)

**After MVP Proven** (Priority Order):

1. **Snapshot/Resume** (Phase 2)
   - Business value: Fast cold starts for Trigger.dev
   - Effort: 5-7 days
   - Dependency: MVP working first

2. **Multi-VM Isolation**
   - Business value: Run multiple workloads safely
   - Effort: 3-4 days (network policies, resource limits)
   - Dependency: Single VM stable

3. **Observability**
   - Business value: Production debugging, metrics
   - Effort: 4-5 days (structured logging, Prometheus metrics)
   - Dependency: MVP proven in production

4. **Trigger.dev Images** (Phase 3)
   - Business value: Actual use case deployment
   - Effort: 10-14 days (build images, test integration)
   - Dependency: Phases 1-2 complete and stable

---

# Part IV: Organizational Actions (Elevator Action)

## 4.1 Governance Framework

### Current Governance State: ABSENT

**Evidence**:
- ❌ No .specify/ directory (spec-driven development not implemented)
- ❌ Backlog empty (0 tasks, despite weeks of work)
- ❌ Constitution exists but not followed (specs requirement ignored)
- ❌ No ADRs (decisions undocumented)
- ❌ No quality gates (phases marked "complete" without validation)

**Organizational Debt**: The project has accumulated significant organizational debt. Code exists, but the process to create it was ad-hoc.

### Recommended Governance Structure

**Governance Hierarchy**:

```
1. Constitution (.claude/constitution.md)
   ↓ SUPREME AUTHORITY

2. Specifications (.specify/features/{branch}/spec.md)
   ↓ WHAT to build (technology-agnostic)

3. Implementation Plans (.specify/features/{branch}/plan.md)
   ↓ HOW to build (technology-specific)

4. Backlog Tasks (backlog/tasks/)
   ↓ WHEN and WHO (execution tracking)

5. Code (internal/, cmd/)
   ↓ Actual implementation

6. Tests (unit, integration, E2E)
   ↓ Validation that code matches spec
```

**Governance Roles**:

| Role | Responsibility | Authority | Frequency |
|------|---------------|-----------|-----------|
| **Human Authority** | Approve specs, unblock decisions | Final say on all one-way doors | As needed |
| **Architect Agent** | Create specs, validate alignment | Can reject non-compliant work | Per feature |
| **Engineer Agent** | Implement from specs, write tests | Can propose spec changes | Continuous |
| **CI/CD Pipeline** | Enforce quality gates | Can block merge | Per commit |

---

## 4.2 Spec-Driven Development Requirements

### Implementation Plan

**Phase 1: Bootstrap Governance** (1 day)

1. **Initialize .specify/ structure**
   ```bash
   mkdir -p .specify/{features,templates,scripts/bash,memory}

   # Move constitution to correct location
   cp .claude/constitution.md .specify/memory/constitution.md

   # Create template
   # (Use jp-spec-kit template or create minimal version)
   ```

2. **Document existing architecture**
   - Create .specify/features/phase1-foundation/spec.md
   - Document WHAT was built (VM lifecycle, networking)
   - Define success criteria (retroactively)
   - Status: IMPLEMENTED (mark historical)

3. **Create ADRs for past decisions**
   - ADR-001: Choice of Go as implementation language
   - ADR-002: Ubuntu 24.04 as base image
   - ADR-003: GHCR as image registry
   - ADR-004: Slicer kernel reuse
   - Location: backlog/decisions/

**Phase 2: Governance for Current Work** (ongoing)

1. **Fix MVP blockers using spec-driven approach**
   ```bash
   # For "Services not starting" bug
   /jpspec.specify "Fix: Services must start in VMs"
   # Creates spec defining problem, success criteria

   # Generate fix plan
   /jpspec.plan

   # Track in backlog
   backlog task create "Fix systemd init parameter" --label "Bug" --milestone "MVP"

   # Implement with TDD
   # - Write test: VM boots, curl port 80 returns HTML
   # - Fix: Add init parameter
   # - Validate: Test passes

   # Close task
   backlog task archive <task-id>
   ```

2. **E2E test as specification**
   ```bash
   # Create spec for E2E workflow
   /jpspec.specify "E2E Test: Full VM Lifecycle"

   # Spec defines:
   # - WHAT: Pull image → Create VM → Services work → Snapshot → Resume → Delete
   # - SUCCESS: Script runs without errors, all validations pass
   # - FAILURE MODES: Each step has clear error message

   # Implement as test script
   # scripts/test-e2e.sh
   ```

**Phase 3: Enforce for Future Work** (permanent)

1. **Quality gate in CI**
   ```yaml
   # .github/workflows/ci.yaml

   - name: Validate specification exists
     run: |
       # For feature branches, require spec
       if [[ "$BRANCH" == feature/* ]]; then
         FEATURE=$(echo $BRANCH | cut -d/ -f2)
         if [ ! -f ".specify/features/$FEATURE/spec.md" ]; then
           echo "ERROR: No specification found for $FEATURE"
           exit 1
         fi
       fi
   ```

2. **Pre-commit hook**
   ```bash
   # .git/hooks/pre-commit

   # Check if new Go files added without corresponding spec
   # (Except for bug fixes in existing features)
   ```

---

## 4.3 Task Management Structure

### Backlog Organization

**Current State**: Backlog initialized but empty (0 tasks despite weeks of work)

**Required State**: ALL work tracked in backlog

**Backlog Categories**:

```
backlog/
├── tasks/
│   ├── mvp-blockers/          # Critical path to MVP
│   ├── phase1-completion/     # Finish Phase 1 items
│   ├── phase2-snapshot/       # Future: Snapshot/resume
│   ├── phase3-trigger/        # Future: Trigger.dev integration
│   ├── tech-debt/             # Refactoring, cleanup
│   └── documentation/         # Doc updates, guides
├── completed/
│   └── phase1-foundation/     # Historical work (retroactive)
├── decisions/
│   ├── ADR-001-go-language.md
│   ├── ADR-002-ubuntu-base.md
│   └── ...
└── docs/
    └── workflow-guide.md      # How to use backlog
```

**Task Template**:

```markdown
# Task: <Title>

**ID**: TASK-001
**Status**: To Do | In Progress | Done
**Priority**: P0 (Critical) | P1 (High) | P2 (Medium) | P3 (Low)
**Milestone**: MVP | Phase 2 | Phase 3
**Labels**: Bug | Feature | Tech Debt | Docs
**Effort**: <hours estimate>

## Description

<What needs to be done>

## Specification Reference

- Spec: .specify/features/<branch>/spec.md
- Plan: .specify/features/<branch>/plan.md

## Acceptance Criteria

- [ ] Criterion 1
- [ ] Criterion 2
- [ ] Tests pass
- [ ] Docs updated

## Dependencies

- Blocked by: TASK-XXX
- Blocks: TASK-YYY

## Notes

<Implementation notes, links, context>
```

### Initial Backlog Population

**Immediate Tasks** (MVP Blockers):

```bash
# TASK-001: Fix systemd init parameter
backlog task create \
  "Fix: Add init=/lib/systemd/systemd to kernel args" \
  --label "Bug,P0,MVP" \
  --milestone "MVP" \
  --description "Services failing in VMs. Add init parameter to vm_handlers.go"

# TASK-002: Create E2E test script
backlog task create \
  "Create scripts/test-e2e.sh for full workflow validation" \
  --label "Testing,P0,MVP" \
  --milestone "MVP" \
  --description "E2E test: pull image → create VM → validate services → cleanup"

# TASK-003: Unified error handling
backlog task create \
  "Define and implement unified error schema in API" \
  --label "Enhancement,P1,MVP" \
  --milestone "MVP" \
  --description "All API errors must include: code, message, context"
```

**Phase 1 Cleanup Tasks**:

```bash
# TASK-004: Document actual status
backlog task create \
  "Update all docs to reflect actual vs claimed status" \
  --label "Documentation,P1" \
  --milestone "Phase 1" \
  --description "Mark which phases are truly complete, which are partial"

# TASK-005: Fix CreateSnapshot test
backlog task create \
  "Implement CreateSnapshot or mark Phase 2 feature" \
  --label "Bug,P2" \
  --milestone "Phase 2" \
  --description "Test failing with 'not yet implemented' - implement or defer to Phase 2"
```

**Governance Tasks**:

```bash
# TASK-006: Initialize .specify/ structure
backlog task create \
  "Bootstrap spec-driven development structure" \
  --label "Governance,P0" \
  --milestone "Governance" \
  --description "Create .specify/ directories, templates, move constitution"

# TASK-007: Create ADRs for past decisions
backlog task create \
  "Document architectural decisions as ADRs" \
  --label "Documentation,P1,Governance" \
  --milestone "Governance" \
  --description "ADRs for: Go, Ubuntu 24.04, GHCR, Slicer kernel"
```

---

## 4.4 Review and Approval Process

### Current State: No Review Process

**Evidence**: Code merged without:
- Specification approval
- Architecture review
- Security review
- E2E validation

### Recommended Review Process

**Feature Development Flow**:

```
1. Specification Phase
   Developer: Create spec via /jpspec.specify
   ↓
   Architect: Review spec for completeness
   ↓
   Human: Approve or request changes
   ↓ GATE: Spec must be approved before implementation

2. Implementation Planning
   Developer: Generate plan via /jpspec.plan
   ↓
   Architect: Review for architectural alignment
   ↓ GATE: Plan must align with platform strategy

3. Task Creation
   Developer: Create backlog tasks
   ↓
   Human: Prioritize and assign milestone
   ↓ GATE: Task must be in backlog before code

4. Implementation
   Developer: Write tests (RED) → Implement (GREEN) → Refactor
   ↓
   CI: Run all checks (build, test, lint, security)
   ↓ GATE: All CI checks must pass

5. Code Review
   Human or Senior Engineer: Review PR
   ↓
   Checks:
   - Tests exist and pass
   - Spec requirements met
   - Code quality acceptable
   - No security issues
   ↓ GATE: Approved PR required to merge

6. Integration Validation
   CI: Run E2E tests on merged code
   ↓
   Monitor: Check for regressions
   ↓ GATE: E2E tests pass on main

7. Task Closure
   Developer: Close backlog task
   ↓
   Architect: Validate against spec
   ↓ COMPLETE: Task archived
```

**Review Roles**:

| Stage | Reviewer | Approval Criteria | Authority |
|-------|----------|------------------|-----------|
| **Spec** | Architect Agent | Complete, measurable, technology-agnostic | Can request revisions |
| **Plan** | Architect Agent | Aligns with platform, uses Golden Path | Can reject non-compliance |
| **Code** | Human or Senior Dev | Tests pass, code quality good | Can request changes |
| **Security** | CI (Trivy, gosec) | No critical vulnerabilities | Can block merge |
| **Integration** | CI (E2E tests) | Full workflow works | Can block deployment |

**Approval Thresholds**:

- **One-Way Door Decisions**: Require human approval (architecture, dependencies, data models)
- **Two-Way Door Decisions**: Engineer can approve (variable names, internal refactoring)
- **Bug Fixes**: Expedited review (tests required, minimal approval)
- **Documentation**: Self-merge allowed (after CI lint passes)

---

## 4.5 Documentation Standards

### Current State: Documentation Chaos

**Problems**:
- 50+ markdown files in docs/building/
- Conflicting information (multiple status reports)
- Unclear which docs are current vs. historical
- No single source of truth
- Planning docs mixed with implementation docs

### Recommended Documentation Structure

**Documentation Hierarchy**:

```
docs/
├── README.md                    # User-facing overview
├── GETTING_STARTED.md          # Quick start guide
├── API_REFERENCE.md            # API contract (generated from spec)
├── CLI_REFERENCE.md            # CLI commands (generated from code)
├── ARCHITECTURE.md             # High-level architecture (stable)
└── TROUBLESHOOTING.md          # Common issues and fixes

docs/building/                   # Developer documentation
├── README.md                    # Overview for contributors
├── DEVELOPMENT_GUIDE.md        # Setup, build, test
├── GOLDEN_PATH.md              # Standard workflow (NEW)
├── implementation/
│   ├── API_CONTRACT.md         # Current API spec
│   ├── CLI_SPEC.md             # Current CLI spec
│   └── TESTING.md              # Test strategy
├── planning/                    # Historical planning docs
│   ├── EXECUTION_PLAN.md       # Original plan (archived)
│   └── BIG-IDEAS.md            # Future ideas (backlog)
└── reports/                     # Status reports (dated)
    ├── 2025-11-23-architect-elevator.md  # This document
    └── ACTUAL_STATUS_REPORT.md # Historical (archived)

.specify/                        # Specifications (source of truth)
├── features/
│   ├── phase1-foundation/
│   │   ├── spec.md             # WHAT was built
│   │   ├── plan.md             # HOW it was built
│   │   └── tasks.md            # Work breakdown
│   ├── mvp-services-fix/       # Current work
│   │   ├── spec.md
│   │   ├── plan.md
│   │   └── tasks.md
│   └── phase2-snapshot/        # Future work
│       └── spec.md
├── templates/
│   └── spec-template.md        # Standard template
└── memory/
    └── constitution.md         # Supreme authority

backlog/
├── tasks/                       # Current work items
├── completed/                   # Historical tasks
├── decisions/                   # ADRs (architectural decisions)
│   ├── ADR-001-go-language.md
│   └── ...
└── docs/
    └── workflow-guide.md       # How to use backlog
```

**Documentation Principles**:

1. **Single Source of Truth**
   - Specs in .specify/ are authoritative
   - Docs are derived from specs
   - Code is validated against specs

2. **Versioned Documentation**
   - User docs (docs/) match released version
   - Builder docs (docs/building/) match current main branch
   - Specs (.specify/) version-controlled with code

3. **Generated Where Possible**
   - API reference from OpenAPI spec
   - CLI reference from cobra commands
   - Test reports from CI artifacts

4. **Clear Ownership**
   - User docs: Technical writer (or assigned engineer)
   - API contract: Architect
   - Implementation guides: Senior engineer
   - ADRs: Decision maker (architect or human authority)

5. **Lifecycle Management**
   - **Current**: In use, actively maintained
   - **Historical**: Archived, read-only, dated
   - **Deprecated**: Marked for removal, redirect to replacement

**Documentation Review**:
- All docs reviewed in PR (like code)
- Outdated docs moved to planning/ or deleted
- README.md updated with each release

---

# Part V: Tactical Implementation Roadmap

## 5.1 Priority Order for Completion

### Phase 0: Governance Bootstrap (IMMEDIATE - 1-2 days)

**Goal**: Establish process before resuming feature work

**Tasks**:
1. ✅ Create .specify/ structure
2. ✅ Initialize backlog with current work
3. ✅ Create ADRs for past decisions
4. ✅ Define Golden Path
5. ✅ Update README with governance requirements

**Deliverables**:
- .specify/memory/constitution.md (moved from .claude/)
- backlog/ populated with TASK-001 through TASK-007
- backlog/decisions/ADR-001 through ADR-004
- docs/building/GOLDEN_PATH.md
- Updated README.md with governance section

**Validation**:
- All current work tracked in backlog
- Specs exist for work in progress
- Team understands workflow

**Owner**: Architect + Human Authority

---

### Phase 1: MVP Completion (HIGH PRIORITY - 2-3 days)

**Goal**: Prove core value proposition end-to-end

**Blockers to Resolve**:

**TASK-001: Fix Services Starting** (P0, 2-4 hours)
- **Spec**: .specify/features/mvp-services-fix/spec.md
- **Implementation**:
  1. Update internal/api/vm_handlers.go line 99 and 217
  2. Add `init=/lib/systemd/systemd` to kernel args
  3. Rebuild daemon: `mage daemon`
  4. Test: Create VM, curl http://172.16.0.10:80
- **Success Criteria**: nginx returns HTML, todo-backend returns JSON

**TASK-002: E2E Test Script** (P0, 4-6 hours)
- **Spec**: .specify/features/e2e-validation/spec.md
- **Implementation**:
  1. Create scripts/test-e2e.sh
  2. Test full workflow: pull → create → validate services → delete
  3. Include all validation steps
  4. Add to CI (mage testE2E)
- **Success Criteria**: Script runs green, validates all MVP requirements

**TASK-003: Unified Error Handling** (P1, 6-8 hours)
- **Spec**: .specify/features/error-handling/spec.md
- **Implementation**:
  1. Define error schema in docs/building/implementation/API_CONTRACT.md
  2. Update all handlers to use schema
  3. Add error code constants
  4. Update CLI to parse and display errors
- **Success Criteria**: All API errors have code, message, context

**Validation Gates**:
- ✅ Services run in VMs (nginx, custom apps)
- ✅ E2E test passes on CI
- ✅ Error messages actionable
- ✅ Full workflow documented

**Owner**: Platform Engineer + QA

---

### Phase 2: Snapshot/Resume (DEFER until MVP proven - 5-7 days)

**Goal**: Enable fast cold starts

**Prerequisites**:
- ✅ Phase 1 MVP validated in production
- ✅ Performance baseline established (current boot time)
- ✅ Snapshot storage strategy decided (local vs S3)

**Tasks**:
1. Implement Firecracker snapshot API
2. Add snapshot endpoints to API
3. Add CLI snapshot commands
4. Test snapshot/resume workflow
5. Benchmark cold start time

**Success Criteria**:
- Create snapshot of running VM
- Resume from snapshot in <2 seconds
- Snapshots persist across daemon restarts

**Owner**: Backend Engineer

---

### Phase 3: Trigger.dev Integration (DEFER until Phase 2 proven - 10-14 days)

**Goal**: Deploy actual use case

**Prerequisites**:
- ✅ Phase 1-2 complete and stable
- ✅ Business requirements validated (Trigger.dev team consulted)
- ✅ Multi-VM isolation tested

**Tasks**:
1. Build Trigger.dev web image
2. Build Trigger.dev worker image
3. Configure inter-VM networking
4. Test full Trigger.dev deployment
5. Document deployment guide

**Success Criteria**:
- Trigger.dev web and worker run in separate VMs
- Web can communicate with worker
- Job execution works end-to-end

**Owner**: Backend Engineer + Trigger.dev stakeholder

---

## 5.2 Integration Points and Patterns

### Critical Integration Points

**1. CLI ↔ API**
- **Pattern**: Request-Reply (HTTP or Unix socket)
- **Contract**: REST API (docs/building/implementation/API_CONTRACT.md)
- **Error Handling**: Unified error schema (TASK-003)
- **Security**: Unix socket file permissions OR (future) API key for TCP
- **Testing**: Unit tests for client, integration tests for full flow

**2. API ↔ Firecracker**
- **Pattern**: Command Message (process invocation)
- **Contract**: Firecracker API spec (external dependency)
- **Error Handling**: Parse Firecracker exit codes, logs
- **Resource Management**: Process cleanup, TAP device cleanup
- **Testing**: Mock Firecracker for unit tests, real Firecracker for integration

**3. API ↔ Database**
- **Pattern**: Synchronous query (SQLite)
- **Contract**: internal/storage/db.go schema
- **Error Handling**: SQL errors wrapped with context
- **Concurrency**: SQLite handles locking
- **Testing**: In-memory DB for unit tests, file DB for integration

**4. Image Registry ↔ API**
- **Pattern**: Pull-based (OCI registry)
- **Contract**: OCI image spec, Docker registry API
- **Error Handling**: Network errors, auth failures, image not found
- **Security**: GHCR token authentication
- **Testing**: Mock registry for unit tests, real GHCR for integration

**5. VM ↔ Host Network**
- **Pattern**: Bridge + TAP + NAT (L2/L3 networking)
- **Contract**: Linux networking (iproute2, iptables)
- **Error Handling**: TAP creation failures, IP conflicts
- **Resource Management**: IPAM, TAP cleanup on VM delete
- **Testing**: Network namespace for unit tests, real TAP for integration

### Integration Testing Strategy

**Unit Tests** (Fast, isolated):
- Mock all external dependencies
- Test logic in isolation
- No network, no filesystem, no Firecracker

**Integration Tests** (Slower, real dependencies):
- Use test Firecracker VM
- Use test network namespace
- Use in-memory or temp DB
- Clean up resources after test

**E2E Tests** (Slowest, full system):
- Pull real image from GHCR (or use local)
- Create real VM with Firecracker
- Validate services work
- Test full lifecycle
- Run on CI with KVM runner

**Test Pyramid**:
```
     ┌──────────────┐
     │  E2E Tests   │  3-5 tests (critical paths)
     │  (Slowest)   │
     ├──────────────┤
     │ Integration  │  20-30 tests (component interactions)
     │   Tests      │
     ├──────────────┤
     │  Unit Tests  │  100+ tests (all logic)
     │  (Fastest)   │
     └──────────────┘
```

---

## 5.3 Quality Gates Per Phase

### Phase 0: Governance (Entry Gates)

**Before starting any new feature**:
- ✅ Backlog task exists for the work
- ✅ Specification created and approved
- ✅ Implementation plan reviewed
- ✅ Dependencies identified and available
- ✅ Test strategy defined

**Exit Gates**:
- ✅ Governance structure documented
- ✅ Team trained on Golden Path
- ✅ Backlog populated with current work
- ✅ ADRs created for past decisions

---

### Phase 1: MVP (Quality Gates)

**Entry Gates**:
- ✅ Phase 0 governance complete
- ✅ MVP scope defined and agreed
- ✅ Blockers identified and prioritized

**Development Gates** (Per Task):
- ✅ Test written (RED)
- ✅ Test fails as expected
- ✅ Implementation passes test (GREEN)
- ✅ Code refactored
- ✅ All tests still pass
- ✅ CI checks pass (lint, security)

**Integration Gates**:
- ✅ Component tests pass
- ✅ Integration tests pass
- ✅ No regressions in existing tests

**Exit Gates** (Phase 1 Complete):
- ✅ All MVP blockers resolved
- ✅ E2E test passes
- ✅ Services run in VMs (nginx, custom app)
- ✅ Error handling consistent
- ✅ Documentation updated
- ✅ Demo to stakeholder successful
- ✅ No critical bugs
- ✅ Backlog tasks closed

---

### Phase 2: Snapshot/Resume (Quality Gates)

**Entry Gates**:
- ✅ Phase 1 validated in production
- ✅ Performance baseline established
- ✅ Snapshot storage strategy decided
- ✅ Specification approved

**Development Gates**:
- ✅ Snapshot creation works
- ✅ Snapshot restore works
- ✅ Performance target met (<2s cold start)
- ✅ Snapshots persist across restarts
- ✅ Resource cleanup works

**Exit Gates**:
- ✅ Cold start <2 seconds (measured)
- ✅ E2E test includes snapshot/resume
- ✅ No memory leaks (tested with 100 snapshot cycles)
- ✅ Documentation complete
- ✅ Demo to stakeholder

---

### Phase 3: Trigger.dev (Quality Gates)

**Entry Gates**:
- ✅ Phase 1-2 complete
- ✅ Business requirements validated
- ✅ Trigger.dev images buildable
- ✅ Multi-VM testing passed

**Exit Gates**:
- ✅ Trigger.dev web and worker run
- ✅ Inter-VM communication works
- ✅ Job execution successful
- ✅ Performance acceptable
- ✅ Deployment guide complete
- ✅ Stakeholder acceptance

---

## 5.4 Success Metrics

### Phase 0: Governance

| Metric | Target | Measurement |
|--------|--------|-------------|
| **Spec coverage** | 100% of new features | Count specs vs features |
| **Task coverage** | 100% of work | Count backlog tasks vs commits |
| **ADR coverage** | 100% of arch decisions | Count ADRs vs decisions |
| **Governance compliance** | 100% | Audit PRs for spec/task |

---

### Phase 1: MVP

| Metric | Target | Measurement |
|--------|--------|-------------|
| **Service availability** | 100% | curl nginx, todo-backend |
| **E2E test pass rate** | 100% | CI test results |
| **VM boot time** | <30 seconds | Timestamp logs |
| **Host→VM latency** | <1ms | ping statistics |
| **Error clarity** | 100% actionable | User testing |
| **Test coverage** | >70% | go test -cover |

---

### Phase 2: Snapshot/Resume

| Metric | Target | Measurement |
|--------|--------|-------------|
| **Cold start time** | <2 seconds | Timestamp snapshot→running |
| **Snapshot size** | <500MB | du -sh snapshot files |
| **Resume success rate** | >99% | 100 cycles, count failures |
| **Memory overhead** | <10% | Before/after snapshot memory |

---

### Phase 3: Trigger.dev

| Metric | Target | Measurement |
|--------|--------|-------------|
| **Job success rate** | >99% | Trigger.dev metrics |
| **Inter-VM latency** | <5ms | Network measurements |
| **Deployment time** | <5 minutes | Timed deployment script |
| **Resource usage** | <2GB RAM per VM | Monitor memory |

---

# Part VI: Risk Assessment and Mitigation

## 6.1 Architectural Risks

### RISK-001: Technical Debt Accumulation (CRITICAL)

**Risk**: Continued development without governance will make codebase unmaintainable

**Impact**: HIGH - Future changes become expensive, onboarding new developers impossible

**Likelihood**: HIGH - Already happening (50+ docs, no specs, empty backlog)

**Mitigation**:
1. **Immediate**: Freeze new features until Phase 0 governance complete
2. **Short-term**: Retroactively document existing work (ADRs, specs)
3. **Long-term**: Enforce governance in CI (block PRs without spec)

**Owner**: Architect + Human Authority

**Status**: ACTIVE - Mitigation in progress (this document)

---

### RISK-002: Service Startup Failure (HIGH)

**Risk**: MVP blocked on services not starting in VMs

**Impact**: HIGH - Cannot demonstrate value, project appears non-functional

**Likelihood**: MEDIUM - Hypothesis exists (init parameter), fix straightforward

**Mitigation**:
1. **Immediate**: Validate hypothesis (read console logs)
2. **Short-term**: Implement fix (add init parameter)
3. **Validation**: E2E test proves services work

**Owner**: Platform Engineer

**Status**: ACTIVE - Fix planned (TASK-001)

---

### RISK-003: Scope Creep (MEDIUM)

**Risk**: Adding features before MVP proven (Web UI, S3, Trigger.dev)

**Impact**: MEDIUM - Delays MVP, increases complexity

**Likelihood**: MEDIUM - BIG-IDEAS.md shows many future features

**Mitigation**:
1. **Immediate**: Defer all Phase 2-3 work until MVP proven
2. **Process**: Require spec + approval before new features
3. **Governance**: Backlog milestone gates (cannot start Phase 2 until Phase 1 done)

**Owner**: Human Authority

**Status**: ACTIVE - Governance enforces prioritization

---

### RISK-004: Integration Fragility (MEDIUM)

**Risk**: Components work in isolation but fail when integrated

**Impact**: MEDIUM - E2E failures, poor user experience

**Likelihood**: MEDIUM - Integration tests exist but incomplete

**Mitigation**:
1. **Immediate**: Create comprehensive E2E test (TASK-002)
2. **CI**: Run E2E test on every commit to main
3. **Monitoring**: Detect regressions early

**Owner**: QA + SRE

**Status**: PLANNED - E2E test creation in backlog

---

### RISK-005: Knowledge Silos (LOW)

**Risk**: Only one person understands each component

**Impact**: MEDIUM - Bus factor = 1, difficult to get help

**Likelihood**: LOW - Well-documented, but risk if team grows

**Mitigation**:
1. **Documentation**: Comprehensive specs and ADRs
2. **Pairing**: Pair programming for complex changes
3. **Reviews**: All PRs reviewed by second person

**Owner**: Team Lead

**Status**: MONITORING - Acceptable for current team size

---

## 6.2 Organizational Risks

### RISK-006: Governance Resistance (MEDIUM)

**Risk**: Team finds spec-driven development too heavyweight

**Impact**: MEDIUM - Governance ignored, returns to ad-hoc development

**Likelihood**: MEDIUM - Change management challenge

**Mitigation**:
1. **Education**: Show value (reduced rework, clearer requirements)
2. **Streamline**: Make specs lightweight (max 3 clarifications)
3. **Tooling**: Automate spec creation (/jpspec.specify)
4. **Success**: Demonstrate on MVP (faster with specs than without)

**Owner**: Architect + Human Authority

**Status**: ACTIVE - Governance rollout must demonstrate value

---

### RISK-007: Process Overhead (LOW)

**Risk**: Governance slows development velocity

**Impact**: LOW - Acceptable if quality improves

**Likelihood**: MEDIUM - Initially slower, faster long-term

**Mitigation**:
1. **Measurement**: Track velocity before/after governance
2. **Optimization**: Reduce friction (templates, automation)
3. **Balance**: Lightweight specs, not waterfall docs

**Owner**: Team Lead

**Status**: MONITORING - Measure and adjust

---

# Part VII: Recommendations and Next Actions

## 7.1 Immediate Actions (Next 24 Hours)

### For Human Authority

1. **Review and Approve This Analysis**
   - Read this document end-to-end
   - Approve governance approach or request changes
   - Authorize Phase 0 governance bootstrap

2. **Approve Phase 0 Plan**
   - Confirm 1-2 day investment in governance acceptable
   - Approve backlog task priorities
   - Commit to enforcing governance going forward

3. **Decide on MVP Scope**
   - Confirm MVP definition (services working, E2E test passing)
   - Defer Phase 2-3 work until MVP proven
   - Set deadline for MVP completion (suggest 1 week)

---

### For Development Team

1. **Pause New Features**
   - No new capabilities until governance in place
   - No new "big ideas" until MVP proven
   - Focus on Phase 0 governance bootstrap

2. **Initialize Governance**
   - Run governance bootstrap script (to be created)
   - Create .specify/ structure
   - Populate backlog with TASK-001 through TASK-007
   - Create ADRs for past decisions

3. **Fix MVP Blockers**
   - TASK-001: Fix systemd init parameter (2-4 hours)
   - TASK-002: Create E2E test script (4-6 hours)
   - TASK-003: Unified error handling (6-8 hours)

---

## 7.2 One-Week Plan

### Week 1: Governance + MVP

**Day 1-2: Phase 0 Governance** (1-2 days)
- Initialize .specify/ structure
- Create backlog tasks for all current work
- Write ADRs for past decisions
- Document Golden Path
- Team training on new workflow

**Day 3-5: MVP Completion** (2-3 days)
- Fix services not starting (TASK-001)
- Create E2E test script (TASK-002)
- Implement unified error handling (TASK-003)
- Validate full workflow works
- Update documentation

**Day 6-7: MVP Validation** (1-2 days)
- Run E2E test repeatedly (stability check)
- Demo to stakeholders
- Gather feedback
- Decide on Phase 2 go/no-go

**Exit Criteria**:
- ✅ Governance in place and enforced
- ✅ MVP proven end-to-end
- ✅ E2E test green on CI
- ✅ Services run in VMs
- ✅ Documentation accurate

---

## 7.3 One-Month Roadmap

### Weeks 1-2: Governance + MVP (As Above)

### Weeks 3-4: Phase 2 - Snapshot/Resume

**Prerequisites**:
- ✅ MVP validated
- ✅ Performance baseline established
- ✅ Snapshot storage strategy decided

**Tasks**:
1. Spec: Define snapshot/resume requirements
2. Plan: Design snapshot storage, API changes
3. Implement: Firecracker snapshot integration
4. Test: Validate <2s cold start
5. Validate: E2E test includes snapshot/resume

**Exit Criteria**:
- ✅ Snapshot creation works
- ✅ Resume from snapshot works
- ✅ Performance target met (<2s)
- ✅ No memory leaks
- ✅ Documentation complete

---

### Future (Month 2+): Phase 3 - Trigger.dev

**Prerequisites**:
- ✅ Phase 1-2 complete and stable
- ✅ Business requirements validated
- ✅ Multi-VM isolation tested

**Tasks**:
1. Build Trigger.dev images
2. Configure inter-VM networking
3. Test full deployment
4. Document deployment guide
5. Stakeholder acceptance

**Exit Criteria**:
- ✅ Trigger.dev web + worker running
- ✅ Job execution works
- ✅ Performance acceptable
- ✅ Deployment guide complete

---

## 7.4 Definition of Done (Per Phase)

### Phase 0: Governance

**Specification**:
- [ ] .specify/memory/constitution.md exists and enforced
- [ ] .specify/ structure created
- [ ] Backlog populated with all current/future work
- [ ] ADRs created for past decisions

**Process**:
- [ ] Golden Path documented
- [ ] Team trained on workflow
- [ ] CI enforces spec requirement

**Validation**:
- [ ] Next feature follows spec-driven approach
- [ ] All work tracked in backlog
- [ ] ADRs created for new decisions

---

### Phase 1: MVP

**Specification**:
- [ ] Services run in VMs (nginx, custom apps)
- [ ] E2E test passes
- [ ] Error handling consistent
- [ ] Full workflow documented

**Code**:
- [ ] Systemd init parameter added
- [ ] E2E test script created
- [ ] Unified error schema implemented
- [ ] All tests pass (unit + integration + E2E)

**Validation**:
- [ ] Demo to stakeholder successful
- [ ] No critical bugs
- [ ] Documentation accurate
- [ ] Can deploy new image and run services

---

### Phase 2: Snapshot/Resume

**Specification**:
- [ ] Cold start <2 seconds (measured)
- [ ] Snapshots persist across restarts
- [ ] No memory leaks (100 cycles)
- [ ] Resource cleanup works

**Code**:
- [ ] Snapshot creation implemented
- [ ] Resume from snapshot implemented
- [ ] Storage strategy implemented
- [ ] All tests pass

**Validation**:
- [ ] Performance target met
- [ ] E2E test includes snapshot/resume
- [ ] Demo to stakeholder
- [ ] Production-ready (stability, security)

---

### Phase 3: Trigger.dev

**Specification**:
- [ ] Trigger.dev web and worker run
- [ ] Inter-VM communication works
- [ ] Job execution successful
- [ ] Deployment guide complete

**Code**:
- [ ] Web image built
- [ ] Worker image built
- [ ] Networking configured
- [ ] All tests pass

**Validation**:
- [ ] End-to-end job execution
- [ ] Performance acceptable
- [ ] Stakeholder acceptance
- [ ] Production deployment successful

---

# Part VIII: Conclusion

## 8.1 Strategic Summary

NanoFuse has achieved significant **technical milestones** but lacks **organizational discipline**. The project demonstrates solid engineering (networking works, VMs boot, binaries compile) but operates without the governance structures required for sustainable development.

**The Elevator is Stuck in the Engine Room**.

The project has been building code without riding the elevator to:
- The **penthouse** to validate strategic alignment
- The **mid-floors** to establish architecture and governance
- Back to the **engine room** with clear specifications and quality gates

## 8.2 Critical Path Forward

**Immediate** (Next 24 hours):
1. Human authority reviews and approves this analysis
2. Freeze new features until governance in place
3. Initialize Phase 0 governance bootstrap

**Short-term** (Next week):
1. Complete Phase 0 governance (1-2 days)
2. Fix MVP blockers (2-3 days)
3. Validate full workflow end-to-end
4. Demo to stakeholders

**Medium-term** (Next month):
1. Snapshot/resume implementation (Phase 2)
2. Performance validation (<2s cold start)
3. Production readiness assessment

**Long-term** (Month 2+):
1. Trigger.dev integration (Phase 3)
2. Platform expansion based on validated use cases
3. Team scaling with governance in place

## 8.3 Key Decisions Required

**One-Way Doors** (Decide Now):
1. ✅ **Adopt spec-driven development**: Use jp-spec-kit for all features
2. ✅ **Enforce backlog-first task management**: No code without backlog task
3. ✅ **Constitution as supreme authority**: Enforce governance strictly

**Two-Way Doors** (Can Defer):
1. ⏸️ **Trigger.dev integration**: Wait until MVP proven
2. ⏸️ **S3 snapshot storage**: Wait until Phase 2 complete
3. ⏸️ **Web UI**: CLI sufficient for now

## 8.4 Success Criteria

**Phase 0 Success** (Governance):
- All work tracked in backlog
- Specs exist for features
- ADRs document decisions
- Team follows Golden Path

**Phase 1 Success** (MVP):
- Services run in VMs
- E2E test passes
- Error handling consistent
- Demo successful

**Phase 2 Success** (Snapshot):
- Cold start <2 seconds
- No memory leaks
- Production-ready

**Phase 3 Success** (Trigger.dev):
- Job execution works
- Stakeholder acceptance
- Deployment guide complete

## 8.5 Final Recommendations

**To Human Authority**:

1. **Approve governance immediately** - Every day without governance increases debt
2. **Enforce quality gates strictly** - Partial enforcement fails
3. **Defer Phase 2-3 work** - Prove MVP before expanding scope
4. **Invest in Phase 0** - 1-2 days now saves weeks of rework later

**To Development Team**:

1. **Embrace spec-driven development** - It reduces rework, not increases it
2. **Track all work in backlog** - Untracked work cannot be prioritized
3. **Write ADRs for decisions** - Future you will thank you
4. **Follow the Golden Path** - Consistency enables automation

**To Architect**:

1. **Ride the elevator regularly** - Bridge penthouse strategy and engine room execution
2. **Sell options, don't prescribe solutions** - Frame trade-offs, let humans decide
3. **Enforce testable boundaries** - Every phase must have measurable exit criteria
4. **Champion governance** - Lead by example, demonstrate value

---

## 8.6 Closing Thought

**"The quality of an architecture is inversely proportional to the number of documents it takes to explain it."** - Gregor Hohpe

NanoFuse has 50+ documentation files because it lacks a **single source of truth** (specifications) and a **clear process** (Golden Path). Once governance is in place, complexity will decrease, velocity will increase, and the platform will become sustainable.

**The elevator is ready. Time to ride it.**

---

**Document Metadata**:
- **Author**: Software Architect Enhanced Agent
- **Date**: 2025-11-23
- **Version**: 1.0
- **Next Review**: After Phase 0 governance complete
- **Status**: DRAFT - Awaiting human authority approval

**Change Log**:
- 2025-11-23: Initial analysis using Architect Elevator framework
