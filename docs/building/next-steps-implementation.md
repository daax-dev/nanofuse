# Implementation Next Steps - flowspec-agents Firecracker Integration

**Status:** Architecture approved and merged (PR #43)
**Date:** December 22, 2025
**Decision Documents:** `dec22-architecture-decisions.jsonl`, `dec22-decisions.md`

---

## Current State

✅ **28 Architecture Decisions Approved**
- Dual runtime (container/Firecracker)
- Tiered rootfs variants (base/container)
- Pre-baked image extraction
- nanofuse-gateway in-VM proxy
- Host-managed VM lifecycle
- Per-user persistent workspace
- SLSA 1.2 supply chain security
- And 20+ more...

✅ **nanofuse Platform Status**
- Core infrastructure ~60% complete
- Pivoting from Trigger.dev to AI sandbox use case
- Firecracker VM lifecycle management exists
- CLI and API daemon foundations in place

⚠️ **Open Items**
- Q9: Resource monitoring approach (need input)
- Q14: Complete threat model document (in progress)

---

## Implementation Approach

Per nanofuse **mandatory spec-driven development** workflow:
1. Create specification (`.specify/features/{branch}/spec.md`)
2. Generate technical plan (`.specify/features/{branch}/plan.md`)
3. Create implementation tasks (`.specify/features/{branch}/tasks.md`)
4. Execute implementation with code review

**No coding without specs** - This is non-negotiable per `CLAUDE.md`.

---

## Option A: Full Integration Feature (Recommended)

### Feature Scope
**"flowspec-agents Integration with nanofuse Firecracker Platform"**

Delivers complete end-to-end capability for AI agents to execute code in secure Firecracker microVMs.

### Components Included

#### 1. Rootfs Build Pipeline (ARCH-026)
- **Extract flowspec-agents** Docker image → rootfs
- **Extract nanofuse-gateway** Docker image → rootfs
- **Build tiered variants:**
  - `flowspec-base.rootfs` (~50MB) - Basic agents
  - `flowspec-container.rootfs` (~150MB) - Platform/SRE agents with containerd
- **Sign images** with cosign (ARCH-023)
- **Generate SBOMs** for compliance (ARCH-023)
- **GitHub Actions pipeline** triggered by upstream releases

#### 2. VM Lifecycle Integration (ARCH-005, ARCH-006, ARCH-016)
- **Agent selection** via `AGENT_TYPE` environment variable
- **Task communication** via mounted drives:
  - Input: `/mnt/task/task.json`, `/mnt/task/context/*`
  - Output: `/mnt/results/output.json`, `/mnt/results/artifacts/*`
- **Host-managed lifecycle:**
  - nanofused controls VM spawn/terminate
  - No autonomous timeouts in microVM
  - Centralized policy management

#### 3. Component Deployment (ARCH-019, ARCH-013)
- **flowspec-agents** runs as native process (no Docker runtime)
- **nanofuse-gateway** runs as in-VM transparent proxy
- **Integration:**
  - Both extracted from Docker images at build
  - Flattened into single rootfs
  - Launched via tini init (ARCH-009)

#### 4. Persistent Storage (ARCH-015)
- **Per-user workspace:** `/workspace/<user-id>` mounted from host
- **Git-based state management** (primary persistence mechanism)
- **Disk quotas** to prevent abuse
- **Periodic cleanup** of old workspaces

#### 5. Secrets Management (ARCH-012)
- **Short-lived GitHub tokens** (refreshable)
- **Cached subscription logins** for services
- **Delivery mechanisms:**
  - Environment variables for tokens
  - Mounted secrets for exceptions

#### 6. Phase 1 MVP Scope
- **File-based results** → evolving to log collector API (ARCH-011)
- **Local filesystem distribution** → S3 later (ARCH-027)
- **Cold start VMs** with session token caching (ARCH-018)
- **Basic crash/retry** → CRIU research Phase 2 (ARCH-017)

### What's NOT Included (Phase 2)
- ❌ Log collector API (file-based for MVP)
- ❌ S3 object storage distribution (local FS for MVP)
- ❌ CRIU checkpointing (research phase)
- ❌ Advanced resource monitoring (Q9 open)
- ❌ Complete threat model (Q14 in progress)

### Command Sequence
```bash
# Step 1: Assess complexity and select approach
/flow:assess

# Step 2: Create specification
/flow:specify "flowspec-agents integration with nanofuse for AI agent execution in Firecracker microVMs"

# Step 3: Research and validation
/flow:research

# Step 4: Generate technical plan (Architect + Platform Engineer agents)
/flow:plan

# Step 5: Begin implementation (Engineer agents with code review)
/flow:implement

# Step 6: Validation (QA, security, docs, release management)
/flow:validate

# Step 7: Operations (CI/CD, DevSecOps, resilience)
/flow:operate
```

### Success Criteria
- [ ] Can spawn flowspec-agents in Firecracker VM
- [ ] Agent type selectable via API parameter
- [ ] Task input/output via mounted drives
- [ ] Results captured and retrievable
- [ ] Per-user workspace persists across sessions
- [ ] GitHub tokens work in VM environment
- [ ] nanofuse-gateway proxies network traffic
- [ ] Sub-200ms boot time achieved
- [ ] Rootfs images signed and verified
- [ ] GitHub Actions pipeline builds images automatically

### Estimated Scope
- **Spec phase:** 1-2 sessions
- **Plan phase:** 1-2 sessions
- **Implementation:** 4-6 sessions (multi-component)
- **Testing/validation:** 2-3 sessions

---

## Option B: Incremental - Rootfs Build Pipeline First

### Feature Scope
**"Automated Rootfs Build Pipeline for flowspec-agents Firecracker Images"**

Foundational component that enables everything else. Smaller, focused scope.

### Components Included

#### 1. Docker Image Extraction
- Pull flowspec-agents Docker image from Docker Hub
- Pull nanofuse-gateway Docker image
- Extract layers using `docker export`
- Merge with Alpine base filesystem

#### 2. Rootfs Variant Building
- **Base variant:**
  - Alpine Linux + Python + flowspec-cli
  - No container runtime
  - ~50MB target size
  - For: tech-writer, researcher, product-manager, business-validator
- **Container variant:**
  - Base + containerd + nerdctl
  - Container building support
  - ~150MB target size
  - For: backend-engineer, frontend-engineer, platform-engineer, sre-agent

#### 3. Init System Setup
- Add tini for lightweight init
- Configure entrypoint script for agent routing
- Handle AGENT_TYPE environment variable

#### 4. Supply Chain Security
- Sign rootfs images with cosign/sigstore
- Generate SBOMs using syft or trivy
- Store attestations with images

#### 5. GitHub Actions Pipeline
- Trigger on flowspec-agents release
- Trigger on nanofuse-gateway release
- Manual workflow_dispatch trigger
- Build both variants
- Sign and publish to GHCR or local artifact storage

#### 6. Local Filesystem Distribution
- Store images in `/var/lib/nanofuse/rootfs/`
- Content-addressable naming (hash-based)
- Version tagging (semantic versioning)

### What Comes Next (Future Features)
- VM lifecycle integration
- Task communication plumbing
- nanofuse-gateway deployment
- Secrets management
- Workspace mounting

### Command Sequence
```bash
# Step 1: Assess complexity and select approach
/flow:assess

# Step 2: Create specification
/flow:specify "Automated rootfs build pipeline for extracting and packaging flowspec-agents into Firecracker-ready images with tiered variants and SLSA 1.2 compliance"

# Step 3: Research (if needed)
/flow:research

# Step 4: Generate technical plan
/flow:plan

# Step 5: Implement
/flow:implement

# Step 6: Validate (QA, security, docs)
/flow:validate

# Step 7: Operations setup (CI/CD pipeline)
/flow:operate
```

### Success Criteria
- [ ] GitHub Actions workflow exists
- [ ] Can extract flowspec-agents Docker image
- [ ] Can extract nanofuse-gateway Docker image
- [ ] Builds base variant (50MB)
- [ ] Builds container variant (150MB)
- [ ] Images are signed with cosign
- [ ] SBOMs are generated
- [ ] Images stored in correct location
- [ ] Can boot extracted rootfs in Firecracker
- [ ] Agent type selection works

### Estimated Scope
- **Spec phase:** 1 session
- **Plan phase:** 1 session
- **Implementation:** 2-3 sessions
- **Testing/validation:** 1-2 sessions

---

## Comparison: Option A vs Option B

| Aspect | Option A (Full Integration) | Option B (Pipeline First) |
|--------|----------------------------|---------------------------|
| **Scope** | Complete end-to-end feature | Foundational component only |
| **User Value** | Can run AI agents in VMs | Can build VM images |
| **Complexity** | High (6 major components) | Medium (1 major component) |
| **Time to Value** | Longer (4-6 sessions) | Shorter (2-3 sessions) |
| **Dependencies** | All at once | Sequential features |
| **Risk** | Higher (more moving parts) | Lower (isolated scope) |
| **Spec Effort** | Larger specification | Smaller specification |
| **Testing** | Complex integration tests | Focused build tests |
| **Iteration** | Big bang delivery | Incremental delivery |

---

## Recommendation

**Start with Option B (Rootfs Build Pipeline)**

### Rationale

1. **Foundation First:** Can't run anything without images; build capability is prerequisite
2. **Lower Risk:** Isolated scope with clear success criteria
3. **Faster Feedback:** 2-3 sessions to working pipeline vs 4-6 for full integration
4. **Better Testing:** Can validate image building independently
5. **Iterative Approach:** Matches agile/lean principles
6. **Clear Dependencies:** Makes Option A implementation easier once pipeline exists

### Iteration Plan

**Sprint 1:** Rootfs build pipeline (Option B)
- Deliverable: Working GitHub Actions pipeline
- Output: Signed, bootable rootfs images in local storage
- Validation: Can boot extracted image in Firecracker

**Sprint 2:** VM lifecycle integration
- Deliverable: Agent selection and task communication
- Output: Can spawn agent, pass task, get results
- Validation: End-to-end agent execution

**Sprint 3:** Component integration
- Deliverable: nanofuse-gateway, workspace, secrets
- Output: Full production-ready system
- Validation: All Phase 1 MVP features working

**Sprint 4:** Polish and Phase 2 planning
- Deliverable: Documentation, monitoring, operations
- Output: Production deployment guide
- Next: Plan Phase 2 features (log collector, S3, CRIU)

---

## Open Questions to Resolve

Before starting implementation, consider addressing:

### Q9: Resource Monitoring ⚠️
**Question:** Host-based or in-VM telemetry for resource monitoring?

**Options:**
- **Host-based:** Firecracker API metrics (simpler, less overhead)
- **In-VM agent:** Lightweight telemetry daemon (more detailed)
- **Hybrid:** Host metrics + selective in-VM when needed

**Impact:** Affects rootfs contents and nanofused implementation

**Recommendation needed:** Which approach for MVP?

---

### Q14: Threat Model 🚧
**Question:** Complete attack matrix for microVMs and devcontainers?

**Scenarios to document:**
1. Malicious user code execution
2. Compromised AI agent
3. Supply chain attacks (poisoned dependencies)
4. Data exfiltration attempts
5. Resource abuse (DoS)
6. Privilege escalation
7. VM escape attempts
8. Network-based attacks
9. Side-channel attacks
10. Secrets exposure

**Impact:** Security controls in rootfs, network policies, monitoring

**Action item:** Create comprehensive threat model document

---

## Implementation Checklist

### Pre-Implementation (Required)
- [ ] **Choose approach:** Option A (full) or Option B (pipeline first)
- [ ] **Answer Q9:** Resource monitoring approach
- [ ] **Complete Q14:** Threat model document (or defer to Phase 2)
- [ ] **Verify dependencies:**
  - [ ] flowspec CLI available (check with `flowspec --version` or `/flow:assess`)
  - [ ] Backlog initialized (`ls backlog/`)
  - [ ] GitHub Actions access configured
  - [ ] GHCR or local storage ready

### Assessment Phase
- [ ] Run `/flow:assess` to analyze task complexity and select approach
- [ ] Review mode selection and recommendations

### Specification Phase
- [ ] Run `/flow:specify` with chosen feature description
- [ ] Review generated `spec.md` for completeness
- [ ] Get human approval on specification

### Research Phase
- [ ] Run `/flow:research` for market and technical validation (if needed)
- [ ] Review research findings and business validation

### Planning Phase
- [ ] Run `/flow:plan` to generate technical design
- [ ] Review generated artifacts:
  - [ ] `plan.md` - Technical implementation plan
  - [ ] `research.md` - Research findings
  - [ ] `data-model.md` - Data models (if applicable)
  - [ ] `contracts/` - API contracts (if applicable)
- [ ] Get human approval on plan

### Implementation
- [ ] Run `/flow:implement` to execute implementation with engineer agents
- [ ] Track progress: `backlog task list --status "In Progress"`
- [ ] Update task status as work progresses
- [ ] Run code reviews (automatic via pr-review-toolkit)

### Validation
- [ ] Run `/flow:validate` for QA, security, docs, release management
- [ ] Address any findings
- [ ] Complete all acceptance criteria
- [ ] Create PR and merge

### Post-Implementation
- [ ] Update architecture decisions if needed
- [ ] Document lessons learned
- [ ] Plan next iteration
- [ ] Update project status

---

## Architecture Decision References

All decisions documented in:
- **JSONL Format:** `docs/building/dec22-architecture-decisions.jsonl`
- **Markdown Format:** `docs/building/dec22-decisions.md`

**Relevant Decision IDs for Implementation:**

**Rootfs Build Pipeline:**
- ARCH-004: Pre-baked Rootfs Extraction
- ARCH-003: Tiered Rootfs Variants
- ARCH-026: Automated Rootfs Build Pipeline
- ARCH-023: SLSA 1.2 Supply Chain Security
- ARCH-010: flowspec-agents Alpine Base Unchanged
- ARCH-019: Container Images as Build Artifacts

**VM Lifecycle:**
- ARCH-005: Agent Selection via Environment Variable
- ARCH-006: Task Communication via Mounted Drive
- ARCH-016: Host-Managed VM Lifecycle
- ARCH-008: Ephemeral VMs with External State

**Component Integration:**
- ARCH-013: Network Proxy via nanofuse-gateway
- ARCH-015: Persistent Per-User Workspace
- ARCH-012: Secrets Management Strategy
- ARCH-020: Non-Root Agent Execution

**Operational:**
- ARCH-011: Task Result Communication Evolution
- ARCH-027: Local Filesystem Image Distribution
- ARCH-028: Tag-Based Versioning with Config Consumption
- ARCH-018: Cold Start with Session Token Caching

---

## Commands Summary

### Option A: Full Integration
```bash
/flow:assess
/flow:specify "flowspec-agents integration with nanofuse for AI agent execution in Firecracker microVMs"
/flow:research
/flow:plan
/flow:implement
/flow:validate
/flow:operate
```

### Option B: Pipeline First (Recommended)
```bash
/flow:assess
/flow:specify "Automated rootfs build pipeline for extracting and packaging flowspec-agents into Firecracker-ready images with tiered variants and SLSA 1.2 compliance"
/flow:research  # Optional, if market/tech validation needed
/flow:plan
/flow:implement
/flow:validate
/flow:operate
```

---

## Document History
- **2025-12-22:** Initial next steps document created after PR #43 merge
- **Status:** Ready for approach selection and specification phase

---

## Next Action Required

**Decision Point:** Choose Option A (full integration) or Option B (pipeline first)

Once decided, proceed with `/flow:assess` then `/flow:specify` to begin flowspec workflow.
