# Next Steps: flowspec-agents Firecracker Integration

**Date:** December 22, 2025
**Architecture:** Approved and merged (PR #43)
**Decisions:** 28 architecture decisions documented

---

## TL;DR - Quick Start

**Recommended approach:** Start with **Option B** (Rootfs Build Pipeline)

```bash
/flow:assess
/flow:specify "Automated rootfs build pipeline for extracting flowspec-agents into Firecracker images with SLSA 1.2 compliance"
/flow:plan
/flow:implement
/flow:validate
/flow:operate
```

---

## Current State

### ✅ Architecture Complete
- **28 decisions approved** - Dual runtime, tiered rootfs, pre-baked extraction, nanofuse-gateway proxy, SLSA 1.2 security
- **Documents:** `dec22-architecture-decisions.jsonl`, `dec22-decisions.md`
- **nanofuse:** ~60% complete, pivoting to AI sandbox use case

### ⚠️ Open Items
- **Q9:** Resource monitoring approach (host-based vs in-VM)
- **Q14:** Threat model document (in progress)

---

## Two Approaches

### Option A: Full Integration (End-to-End)

**Scope:** Complete flowspec-agents integration with nanofuse
**Timeline:** 4-6 sessions
**Risk:** Higher (6 major components)

**Delivers:**
- Rootfs build pipeline
- VM lifecycle integration
- Component deployment (flowspec-agents + nanofuse-gateway)
- Persistent workspace
- Secrets management
- Phase 1 MVP (file-based results, local FS, cold start)

**Commands:**
```bash
/flow:assess
/flow:specify "flowspec-agents integration with nanofuse for AI agent execution in Firecracker microVMs"
/flow:research
/flow:plan
/flow:implement
/flow:validate
/flow:operate
```

---

### Option B: Rootfs Build Pipeline First (Recommended)

**Scope:** Foundational image building capability
**Timeline:** 2-3 sessions
**Risk:** Lower (isolated component)

**Delivers:**
- Docker image extraction (flowspec-agents + nanofuse-gateway)
- Tiered rootfs variants (base ~50MB, container ~150MB)
- Init system setup (tini + AGENT_TYPE routing)
- SLSA 1.2 security (cosign signing, SBOMs)
- GitHub Actions pipeline
- Local filesystem distribution

**Commands:**
```bash
/flow:assess
/flow:specify "Automated rootfs build pipeline for extracting flowspec-agents into Firecracker images with SLSA 1.2 compliance"
/flow:plan
/flow:implement
/flow:validate
/flow:operate
```

**Why recommended:**
- ✅ Foundation first (can't run without images)
- ✅ Lower risk, faster feedback (2-3 sessions)
- ✅ Independent testing
- ✅ Enables iterative delivery

---

## Iteration Plan (Option B → Full System)

### Sprint 1: Rootfs Build Pipeline
**Output:** Signed, bootable rootfs images
**Validation:** Can boot in Firecracker
**Commands:** `/flow:assess` → `/flow:specify` → `/flow:plan` → `/flow:implement` → `/flow:validate` → `/flow:operate`

### Sprint 2: VM Lifecycle Integration
**Output:** Agent selection, task communication
**Validation:** End-to-end agent execution

### Sprint 3: Component Integration
**Output:** nanofuse-gateway, workspace, secrets
**Validation:** Full Phase 1 MVP working

### Sprint 4: Polish & Phase 2 Planning
**Output:** Documentation, monitoring, operations
**Next:** Log collector API, S3 distribution, CRIU

---

## Flowspec 7-Phase Workflow

All features follow this mandatory workflow:

```
/flow:assess      # Analyze complexity, select approach (auto/strategic/explorative)
       ↓
/flow:specify     # PM Planner creates technology-agnostic specification
       ↓
/flow:research    # Market and technical research, business validation (optional)
       ↓
/flow:plan        # Architect + Platform Engineer design artifacts
       ↓
/flow:implement   # Frontend/Backend engineers with code review
       ↓
/flow:validate    # QA, security, documentation, release management
       ↓
/flow:operate     # SRE for CI/CD, DevSecOps, resilience
```

**No coding without specs** - Mandatory per `CLAUDE.md`

---

## Success Criteria

### Option A (Full Integration)
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

### Option B (Pipeline First)
- [ ] GitHub Actions workflow exists
- [ ] Can extract flowspec-agents Docker image
- [ ] Can extract nanofuse-gateway Docker image
- [ ] Builds base variant (~50MB)
- [ ] Builds container variant (~150MB)
- [ ] Images are signed with cosign
- [ ] SBOMs are generated
- [ ] Images stored in `/var/lib/nanofuse/rootfs/`
- [ ] Can boot extracted rootfs in Firecracker
- [ ] Agent type selection works (AGENT_TYPE environment variable)

---

## Architecture Decision Quick Reference

### Rootfs Build Pipeline
- **ARCH-004:** Pre-baked Rootfs Extraction
- **ARCH-003:** Tiered Rootfs Variants (base/container)
- **ARCH-026:** Automated Build Pipeline (GitHub Actions)
- **ARCH-023:** SLSA 1.2 Security (signing, SBOMs)
- **ARCH-019:** Containers as Build Artifacts

### VM Lifecycle
- **ARCH-005:** Agent Selection (AGENT_TYPE env var)
- **ARCH-006:** Task Communication (mounted drives)
- **ARCH-016:** Host-Managed Lifecycle
- **ARCH-008:** Ephemeral VMs with External State

### Components
- **ARCH-013:** nanofuse-gateway In-VM Proxy
- **ARCH-015:** Per-User Persistent Workspace
- **ARCH-012:** Secrets Management (short-lived tokens)
- **ARCH-020:** Non-Root Execution

### Operations
- **ARCH-011:** Results Evolution (file → API)
- **ARCH-027:** Local FS Distribution → S3
- **ARCH-028:** Tag-Based Versioning
- **ARCH-018:** Cold Start with Session Caching

**Full details:** `docs/building/dec22-architecture-decisions.jsonl`

---

## Pre-Flight Checklist

Before running `/flow:assess`:

- [ ] Choose approach (A or B)
- [ ] flowspec CLI available (`/flow:assess` to verify)
- [ ] Backlog initialized (`ls backlog/`)
- [ ] GitHub Actions access configured
- [ ] GHCR or local storage ready
- [ ] (Optional) Answer Q9: Resource monitoring approach
- [ ] (Optional) Complete Q14: Threat model document

---

## What Gets Created

### Specification Phase (`/flow:specify`)
- `.specify/features/{branch}/spec.md` - Technology-agnostic specification
- Automated branch creation
- Quality validation

### Planning Phase (`/flow:plan`)
- `.specify/features/{branch}/plan.md` - Technical implementation plan
- `.specify/features/{branch}/research.md` - Research findings
- `.specify/features/{branch}/data-model.md` - Data models (if applicable)
- `.specify/features/{branch}/contracts/` - API contracts (if applicable)

### Implementation Phase (`/flow:implement`)
- Code changes with automatic code review
- Tests (TDD enforced)
- Progress tracking in backlog

### Validation Phase (`/flow:validate`)
- QA analysis
- Security review
- Documentation updates
- Release readiness

### Operations Phase (`/flow:operate`)
- CI/CD pipeline setup
- DevSecOps integration
- Monitoring and observability
- Resilience testing

---

## Quick Command Reference

### Option A (Full Integration)
```bash
/flow:assess
/flow:specify "flowspec-agents integration with nanofuse for AI agent execution in Firecracker microVMs"
/flow:research  # Market/tech validation
/flow:plan      # Technical design
/flow:implement # Engineering
/flow:validate  # QA + security
/flow:operate   # CI/CD + ops
```

### Option B (Pipeline First) - **RECOMMENDED**
```bash
/flow:assess
/flow:specify "Automated rootfs build pipeline for extracting flowspec-agents into Firecracker images with SLSA 1.2 compliance"
/flow:plan      # Skip /flow:research for infrastructure work
/flow:implement
/flow:validate
/flow:operate
```

---

## Open Questions

### Q9: Resource Monitoring ⚠️
**Options:** Host-based (Firecracker API) vs In-VM agent vs Hybrid
**Impact:** Affects rootfs contents and nanofused implementation
**Status:** Need decision before implementation

### Q14: Threat Model 🚧
**Need:** Complete attack matrix for microVMs and devcontainers
**Scenarios:** Malicious code, compromised agents, supply chain, data exfiltration, resource abuse, privilege escalation, VM escape, network attacks, side-channels, secrets exposure
**Status:** In progress
**Impact:** Security controls in rootfs, network policies, monitoring

---

## Phase 1 vs Phase 2 Scope

### Phase 1 MVP (Current)
- ✅ File-based results
- ✅ Local filesystem distribution
- ✅ Cold start VMs
- ✅ Basic crash/retry
- ✅ Short-lived secrets
- ✅ Pre-installed packages

### Phase 2 (Future)
- ⏭️ Log collector API
- ⏭️ S3 object storage distribution
- ⏭️ CRIU checkpointing
- ⏭️ Advanced resource monitoring
- ⏭️ Warm VM pools

---

## Getting Started

**Decision needed:** Option A or Option B?

**Recommended:** Start with **Option B** for faster, lower-risk delivery.

**First command:**
```bash
/flow:assess
```

This will analyze the task complexity and recommend the best approach (auto/strategic/explorative mode).

**Next command:**
```bash
/flow:specify "Automated rootfs build pipeline for extracting flowspec-agents into Firecracker images with SLSA 1.2 compliance"
```

Then follow the 7-phase workflow to completion. 🚀

---

## Document References

- **Architecture Decisions (JSONL):** `docs/building/dec22-architecture-decisions.jsonl`
- **Architecture Decisions (Markdown):** `docs/building/dec22-decisions.md`
- **Original Questions:** `docs/building/dec22-questions.md`
- **Detailed Implementation Plan:** `docs/building/next-steps-implementation.md`

---

**Status:** Ready for `/flow:assess` → Choose your approach and begin! 🎯
