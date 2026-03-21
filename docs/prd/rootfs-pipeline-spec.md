# Product Requirements Document: Automated Rootfs Build Pipeline for flowspec-agents Firecracker Images

**Product**: nanofuse Automated Rootfs Build Pipeline
**Version**: 1.0
**Date**: December 22, 2025
**Author**: @pm-planner (Product Manager)
**Status**: Draft - Awaiting Approval
**Epic**: Firecracker Image Building Infrastructure

---

## 1. Executive Summary

### Problem Statement

The nanofuse platform (~60% complete) is pivoting to serve as a secure AI code execution sandbox, enabling LLM-generated code to run in isolated Firecracker microVMs with sub-200ms boot times. However, **we cannot run any AI agents without bootable Firecracker VM images**. Currently, there is no automated mechanism to convert flowspec-agents and nanofuse-gateway Docker images into Firecracker-bootable rootfs images.

This is a foundational infrastructure gap that blocks all downstream AI agent execution capabilities. Without this pipeline, the platform cannot deliver on its core value proposition.

**Customer Pain**: Platform engineers and AI agent operators currently have no way to:
- Deploy flowspec-agents in Firecracker microVMs
- Maintain security compliance (SLSA 1.2) for VM images
- Respond quickly to upstream image updates
- Ensure reproducible, verifiable builds

### Proposed Solution (Outcome-Driven)

Build an **automated GitHub Actions pipeline** that extracts Docker images (flowspec-agents + nanofuse-gateway) and produces signed, security-verified Firecracker rootfs images. The pipeline will:

1. **Extract** Docker images from Docker Hub using `docker export`
2. **Build** two tiered rootfs variants optimized for different agent capabilities
3. **Secure** images with cosign signing and SBOM generation (SLSA 1.2 compliance)
4. **Distribute** images to local filesystem storage (/var/lib/nanofuse/rootfs/)
5. **Automate** builds triggered by upstream releases

**Key Outcome**: Reduce time-to-production for new flowspec-agents releases from manual hours to automated 30-minute pipeline execution.

### Success Metrics (North Star + Key Outcomes)

**North Star Metric**: **Time from flowspec-agents release to bootable Firecracker image**
- Current: ∞ (manual, undefined process)
- Target: ≤ 30 minutes (fully automated)

**Leading Indicators** (Team Controls):
- Build success rate ≥ 99%
- Build duration ≤ 20 minutes (P95)
- Zero manual intervention required
- 100% of builds signed and SBOM-verified

**Lagging Indicators** (Business Impact):
- Platform engineer onboarding time reduced 50% (reproducible setup)
- Security audit compliance rate = 100% (SLSA 1.2)
- Firecracker VM boot time < 200ms (image size optimization)
- Zero production incidents from unsigned/unverified images

### Business Value and Strategic Alignment

**Business Value**:
- **Enables Revenue**: Unblocks AI sandbox platform launch (cannot operate without VM images)
- **Reduces Risk**: SLSA 1.2 compliance prevents supply chain attacks
- **Accelerates Velocity**: Automated builds eliminate manual bottleneck
- **Improves Quality**: Reproducible builds eliminate configuration drift

**Strategic Alignment**:
- **nanofuse Mission**: Secure AI code execution in isolated sandboxes
- **Architecture Decisions**: Implements ARCH-004 (Pre-baked Rootfs), ARCH-003 (Tiered Variants), ARCH-026 (Automated Pipeline), ARCH-023 (SLSA 1.2)
- **Phase 1 MVP**: Foundational component required before VM lifecycle integration
- **Security-First**: Aligns with security-by-design principle

---

## 2. User Stories and Use Cases

### Primary User Personas

#### Persona 1: Platform Engineer (Primary)
- **Role**: Infrastructure maintainer, image build operator
- **Goals**:
  - Deploy secure, verified Firecracker images
  - Minimize manual toil in image updates
  - Ensure compliance with security standards
  - Troubleshoot build failures quickly
- **Pain Points**:
  - Manual Docker-to-Firecracker conversion is error-prone
  - No visibility into image provenance
  - Lack of automation increases operational burden
  - Difficulty maintaining consistency across environments

#### Persona 2: AI Agent Operator (Secondary)
- **Role**: End-user deploying AI agents in nanofuse
- **Goals**:
  - Select appropriate agent type and rootfs variant
  - Trust that images are secure and verified
  - Experience fast boot times (<200ms)
  - Access latest agent capabilities quickly
- **Pain Points**:
  - Uncertainty about image freshness
  - No visibility into security posture
  - Boot times degraded by oversized images
  - Delayed access to new agent features

#### Persona 3: Security Auditor (Tertiary)
- **Role**: Compliance verification, security assessment
- **Goals**:
  - Verify SLSA 1.2 compliance
  - Audit image provenance and signatures
  - Review SBOM for vulnerability scanning
  - Validate supply chain integrity
- **Pain Points**:
  - Lack of automated attestations
  - Manual verification is slow and incomplete
  - No structured SBOM format
  - Cannot trace image ancestry

### User Journey Maps

#### Journey 1: Platform Engineer - Initial Setup
1. **Trigger**: New deployment requires Firecracker images
2. **Action**: Configure GitHub Actions workflow with Docker Hub credentials
3. **Action**: Trigger manual build via workflow_dispatch
4. **Experience**: Observe build logs in GitHub Actions UI
5. **Outcome**: Signed images available in /var/lib/nanofuse/rootfs/
6. **Success**: nanofuse CLI lists available images, boots test VM <200ms

**Pain Points**:
- Initial credential setup complexity
- Lack of validation feedback during config
- No guided troubleshooting for first build

#### Journey 2: Platform Engineer - Automated Update
1. **Trigger**: flowspec-agents v2.1.0 released on GitHub
2. **Automation**: GitHub Actions detects release event, triggers build
3. **Experience**: Receives notification of build completion
4. **Action**: Reviews build artifacts, signatures, SBOMs
5. **Action**: Tests on subset of VMs
6. **Outcome**: Updates nanofuse config to use new version
7. **Success**: New VMs automatically use v2.1.0, existing VMs unaffected

**Pain Points**:
- Notification fatigue from frequent builds
- Manual config update required (no auto-promotion)
- Testing process not standardized

#### Journey 3: AI Agent Operator - Agent Execution
1. **Trigger**: Need to run backend-engineer agent
2. **Action**: Call nanofuse API with agent_type=backend-engineer
3. **Automation**: nanofuse selects flowspec-container variant
4. **Experience**: VM boots in <200ms with containerd available
5. **Outcome**: Agent executes task successfully
6. **Success**: Results written to /mnt/results/, task completes

**Pain Points**:
- No visibility into rootfs variant selection logic
- Cannot easily verify which image version is running
- Error messages don't guide troubleshooting

### Detailed User Stories with Acceptance Criteria

#### Story 1: Automated Docker Image Extraction
**As a** platform engineer
**I want** Docker images automatically extracted from Docker Hub
**So that** I don't manually convert images to rootfs format

**Acceptance Criteria**:
- [ ] System pulls flowspec-agents:latest from Docker Hub
- [ ] System pulls nanofuse-gateway:latest from Docker Hub
- [ ] System extracts all layers and merges into single filesystem
- [ ] Extraction works for both x86_64 and arm64 architectures
- [ ] Extraction completes in < 5 minutes

#### Story 2: Tiered Rootfs Variant Building
**As a** platform engineer
**I want** two rootfs variants (base and container) built automatically
**So that** I can optimize image size for different agent capabilities

**Acceptance Criteria**:
- [ ] Base variant (~50MB) includes Alpine + Python + flowspec-cli
- [ ] Container variant (~150MB) includes base + containerd + nerdctl
- [ ] Both variants bootable in Firecracker
- [ ] Boot time < 200ms for both variants
- [ ] Agent type selection correctly maps to variant

#### Story 3: SLSA 1.2 Supply Chain Security
**As a** security auditor
**I want** all images signed with cosign and accompanied by SBOMs
**So that** I can verify image provenance and compliance

**Acceptance Criteria**:
- [ ] Images signed with cosign/sigstore
- [ ] Signatures published alongside images
- [ ] SBOMs generated in SPDX or CycloneDX format
- [ ] SBOMs document all packages and dependencies
- [ ] Verification command succeeds for all images

#### Story 4: Automated Build Triggers
**As a** platform engineer
**I want** builds triggered automatically on flowspec-agents releases
**So that** new agent versions are available quickly without manual intervention

**Acceptance Criteria**:
- [ ] Build triggers on flowspec-agents tag event
- [ ] Build triggers on nanofuse-gateway tag event
- [ ] Manual trigger via workflow_dispatch works
- [ ] Build completes end-to-end without errors
- [ ] Notifications sent on build completion and failure

#### Story 5: Local Filesystem Distribution
**As an** AI agent operator
**I want** images stored in /var/lib/nanofuse/rootfs/
**So that** VM boot is fast with no network dependency

**Acceptance Criteria**:
- [ ] Images stored with content-addressable names
- [ ] Semantic version tags supported (v1.2.3)
- [ ] nanofuse CLI lists available images
- [ ] Image metadata includes variant, version, signatures
- [ ] Storage supports versioned rollback

### Edge Cases and Error Scenarios

#### Edge Case 1: Docker Hub Rate Limiting
- **Scenario**: Pipeline exceeds Docker Hub pull rate limits
- **Handling**: Implement retry with exponential backoff, authenticate pulls with credentials
- **Acceptance**: Pipeline succeeds after rate limit reset

#### Edge Case 2: Corrupt Image Extraction
- **Scenario**: Docker image extraction produces invalid filesystem
- **Handling**: Validate extracted rootfs structure, checksum verification
- **Acceptance**: Build fails fast with actionable error message

#### Edge Case 3: Signature Verification Failure
- **Scenario**: cosign signature verification fails for published image
- **Handling**: Prevent image from being used, alert on verification failure
- **Acceptance**: nanofuse CLI rejects unsigned/unverified images

#### Edge Case 4: Disk Space Exhaustion
- **Scenario**: /var/lib/nanofuse/rootfs/ fills up during build
- **Handling**: Monitor disk usage, implement cleanup of old versions
- **Acceptance**: Pipeline gracefully fails with disk space error

#### Edge Case 5: Multi-Architecture Build Failures
- **Scenario**: x86_64 build succeeds but arm64 fails
- **Handling**: Architecture-specific error reporting, partial publish prevention
- **Acceptance**: Both architectures must succeed or entire build fails

---

## 3. DVF+V Risk Assessment

### Value Risk (Desirability)

**Question**: Will platform engineers and operators use this automated pipeline?

**Assessment**: **LOW RISK**

**Evidence**:
- ✅ **Confirmed Need**: Platform cannot operate without VM images (blocking issue)
- ✅ **Clear Value Proposition**: Reduces manual hours to automated 30 minutes
- ✅ **Measurable Outcome**: Time-to-production is quantifiable improvement
- ✅ **Security Compliance**: SLSA 1.2 is regulatory/compliance requirement
- ✅ **User Validation**: Architecture decisions (ARCH-026) pre-approved by stakeholders

**Mitigation**: None required - value is non-negotiable for platform operation

### Usability Risk (Experience)

**Question**: Can platform engineers trigger, monitor, and troubleshoot builds?

**Assessment**: **MEDIUM RISK**

**Concerns**:
- ⚠️ GitHub Actions UI may be unfamiliar to some operators
- ⚠️ Error messages from Docker extraction may be cryptic
- ⚠️ No guided troubleshooting workflow built-in
- ⚠️ Manual testing and config update required

**Validation Plan**:
1. **Prototype**: Build minimal GitHub Actions workflow with core steps
2. **Usability Test**: Have 2 platform engineers trigger build and interpret errors
3. **Iteration**: Add clear error messages and troubleshooting guides
4. **Documentation**: Create operator runbook with common issues

**Success Criteria**:
- Engineers can trigger build without documentation
- Error messages are actionable (not raw Docker errors)
- Build status is clear from GitHub Actions UI
- Average time-to-resolution for failures < 15 minutes

### Feasibility Risk (Technical)

**Question**: Can we extract Docker images and build rootfs reliably?

**Assessment**: **LOW-MEDIUM RISK**

**Technical Challenges**:
- ✅ **Docker Extraction**: Proven technique (`docker export`), well-documented
- ⚠️ **Multi-Layer Merging**: Requires careful handling of whiteout files and permissions
- ✅ **tini Init Setup**: Lightweight, simple integration
- ⚠️ **Firecracker Boot**: Requires correct kernel cmdline and init system
- ✅ **cosign Signing**: Mature tooling, GitHub Actions integration exists
- ⚠️ **SBOM Generation**: syft/trivy are established tools but output may need validation

**Engineering Spike Needed**:
- [ ] Validate Docker export + layer merging preserves permissions
- [ ] Test Firecracker boot with extracted rootfs
- [ ] Verify cosign signing in GitHub Actions environment
- [ ] Benchmark build time (target < 20 minutes)

**Success Criteria**:
- Docker extraction produces bootable rootfs 100% of builds
- Firecracker boot succeeds with <200ms boot time
- Signatures verify correctly
- Build completes in < 20 minutes (P95)

### Business Viability Risk (Organizational)

**Question**: Does this fit nanofuse's SLSA 1.2 compliance requirements and operational constraints?

**Assessment**: **LOW RISK**

**Organizational Alignment**:
- ✅ **Security Compliance**: SLSA 1.2 is architectural requirement (ARCH-023)
- ✅ **GitHub Actions**: Approved CI/CD platform for nanofuse
- ✅ **Local Storage MVP**: Aligns with Phase 1 scope (ARCH-027)
- ✅ **Versioning Strategy**: Tag-based versioning approved (ARCH-028)
- ✅ **Cost**: GitHub Actions minutes within budget for open-source project

**Stakeholder Sign-Off**:
- ✅ Security team approved SLSA 1.2 approach
- ✅ Platform engineering approved architecture decisions (PR #43)
- ✅ DevOps approved GitHub Actions as build platform

**Mitigation**: None required - full organizational alignment

---

## 4. Functional Requirements

### FR-1: Docker Image Extraction

**Requirement**: Extract flowspec-agents and nanofuse-gateway Docker images from Docker Hub into filesystem structure

**Functionality**:
- Pull `jpoley/flowspec-agents:latest` from Docker Hub
- Pull `jpoley/nanofuse-gateway:latest` from Docker Hub (when available)
- Extract all image layers using `docker export`
- Merge layers into unified filesystem structure
- Preserve file permissions, ownership, symlinks
- Handle whiteout files correctly (layer deletion markers)
- Support both x86_64 and arm64 architectures

**Inputs**:
- Docker Hub credentials (GitHub Secrets)
- Image tags (latest or specific version)
- Target architecture (x86_64, arm64)

**Outputs**:
- Flattened filesystem directory structure
- Extraction metadata (image SHA, timestamp)

**Error Handling**:
- Retry on Docker Hub rate limits (exponential backoff)
- Fail fast on image not found
- Validate extracted filesystem structure
- Report Docker authentication failures clearly

### FR-2: Base Variant Rootfs Building

**Requirement**: Build flowspec-base rootfs variant optimized for basic AI agents

**Functionality**:
- Start with Alpine Linux base
- Merge extracted flowspec-agents filesystem
- Install Python 3.x + pip + common packages (requests, pytest, etc.)
- Install flowspec-cli
- Configure tini as /sbin/init
- Set up entrypoint script for agent type routing
- Create ext4 filesystem image
- Compress with gzip or zstd

**Size Target**: ≤ 60MB (target 50MB)

**Supported Agents**:
- tech-writer
- researcher
- product-manager
- business-validator

**Quality Gates**:
- Bootable in Firecracker
- Boot time < 200ms
- All agent types execute successfully
- No missing dependencies

### FR-3: Container Variant Rootfs Building

**Requirement**: Build flowspec-container rootfs variant with container runtime capabilities

**Functionality**:
- Start with base variant
- Add containerd + nerdctl
- Configure rootless container builds (buildkit)
- Optionally add podman for rootless support
- Configure tini + entrypoint (same as base)
- Create ext4 filesystem image
- Compress with gzip or zstd

**Size Target**: ≤ 180MB (target 150MB)

**Supported Agents**:
- backend-engineer
- frontend-engineer
- platform-engineer
- sre-agent

**Quality Gates**:
- Bootable in Firecracker
- Boot time < 200ms
- Container builds work (rootless)
- All agent types execute successfully

### FR-4: Init System Configuration

**Requirement**: Configure lightweight init system with agent type routing

**Functionality**:
- Install tini as /sbin/init
- Create entrypoint script at /usr/local/bin/agent-entrypoint
- Read AGENT_TYPE environment variable from Firecracker boot params
- Default to backend-engineer if AGENT_TYPE unset
- Route to correct flowspec-agent command
- Handle signal forwarding (SIGTERM, SIGINT)
- Reap zombie processes
- Support both base and container variants

**Entrypoint Script Logic**:
```bash
#!/bin/sh
AGENT_TYPE=${AGENT_TYPE:-backend-engineer}
exec /sbin/tini -- flowspec-agent --agent "$AGENT_TYPE" --task-file /mnt/task.json
```

**Quality Gates**:
- Signal handling verified (SIGTERM kills cleanly)
- Zombie reaping verified (no orphaned processes)
- Agent type routing works for all supported agents
- Graceful shutdown on error

### FR-5: Supply Chain Security (SLSA 1.2)

**Requirement**: Sign images and generate SBOMs for supply chain security

**Functionality**:

**Image Signing**:
- Sign rootfs image hashes with cosign
- Use GitHub OIDC for keyless signing
- Publish signatures to registry alongside images
- Support signature verification before VM launch

**SBOM Generation**:
- Generate SBOMs with syft or trivy
- Document all packages and dependencies
- Output in SPDX or CycloneDX format
- Include base OS, Python packages, containerd, flowspec-agents
- Publish SBOMs alongside images

**Verification**:
- Provide `verify-rootfs.sh` script for signature checking
- Integrate verification into nanofuse CLI
- Reject unsigned/unverified images at runtime

**Quality Gates**:
- All images signed with valid signatures
- Signatures verify with cosign verify
- SBOMs include 100% of installed packages
- SLSA 1.2 compliance verified by security audit

### FR-6: GitHub Actions Pipeline

**Requirement**: Automated build pipeline triggered by upstream releases

**Functionality**:

**Triggers**:
- flowspec-agents release (tag event)
- nanofuse-gateway release (tag event)
- Manual trigger (workflow_dispatch)
- Scheduled security patches (optional)

**Workflow Steps**:
1. Checkout nanofuse repository
2. Authenticate to Docker Hub
3. Pull source Docker images (flowspec-agents, nanofuse-gateway)
4. Extract Docker images to filesystem
5. Build base variant rootfs
6. Build container variant rootfs
7. Sign images with cosign
8. Generate SBOMs with syft
9. Publish images to storage
10. Update version metadata
11. Notify on completion/failure

**Parallelization**:
- Base and container variants build in parallel
- Multi-architecture builds (x86_64, arm64) in parallel

**Artifacts**:
- Rootfs images (base and container)
- Signatures (.sig files)
- SBOMs (.spdx.json or .cdx.json)
- Build metadata (SHA, timestamp, version)

**Quality Gates**:
- Build completes in < 20 minutes (P95)
- Build success rate ≥ 99%
- All tests pass before publish
- Notifications sent on success/failure

### FR-7: Local Filesystem Distribution

**Requirement**: Store images in local filesystem with content-addressable naming

**Functionality**:

**Storage Location**: `/var/lib/nanofuse/rootfs/`

**Directory Structure**:
```
/var/lib/nanofuse/rootfs/
├── base/
│   ├── v1.0.0/
│   │   ├── rootfs.ext4
│   │   ├── rootfs.ext4.sig
│   │   ├── sbom.spdx.json
│   │   └── metadata.json
│   └── latest -> v1.0.0
└── container/
    ├── v1.0.0/
    │   ├── rootfs.ext4
    │   ├── rootfs.ext4.sig
    │   ├── sbom.spdx.json
    │   └── metadata.json
    └── latest -> v1.0.0
```

**Metadata Format** (metadata.json):
```json
{
  "variant": "base",
  "version": "v1.0.0",
  "sha256": "abc123...",
  "built_at": "2025-12-22T19:30:00Z",
  "source_images": {
    "flowspec-agents": "sha256:def456...",
    "nanofuse-gateway": "sha256:ghi789..."
  },
  "signed": true,
  "sbom": "sbom.spdx.json"
}
```

**Discovery**:
- nanofuse CLI lists available images
- Version selection via config or API parameter
- Automatic verification on load

**Quality Gates**:
- Images accessible via nanofuse CLI
- Symlinks (latest) update correctly
- Metadata is machine-readable
- Storage supports concurrent access

### FR-8: Version Management and Rollback

**Requirement**: Support semantic versioning with config-driven consumption

**Functionality**:
- Tag images with semantic versions (v1.2.3)
- nanofuse config specifies active version per variant
- Support "latest" symlink for auto-update
- Enable rollback via config change
- Retain previous versions for rollback (configurable retention)

**Version Selection**:
```yaml
# nanofuse.yml
rootfs:
  base:
    version: v1.2.3  # or "latest"
  container:
    version: v1.2.2  # can pin different versions
```

**Rollback Process**:
1. Edit nanofuse.yml to previous version
2. Restart nanofused or reload config
3. New VMs use rolled-back version
4. Existing VMs unaffected (ephemeral)

**Quality Gates**:
- Version selection works for all variants
- Rollback completes in < 5 minutes
- No data loss on rollback (VMs are ephemeral)
- Audit trail of version changes

---

## 5. Non-Functional Requirements

### NFR-1: Performance

**Build Time**:
- Complete pipeline execution ≤ 20 minutes (P95)
- Docker image extraction ≤ 5 minutes
- Base variant build ≤ 8 minutes
- Container variant build ≤ 12 minutes
- Signing and SBOM generation ≤ 3 minutes

**Boot Time**:
- Firecracker boot time < 200ms for base variant
- Firecracker boot time < 200ms for container variant
- No degradation from baseline Alpine boot time

**Image Size**:
- Base variant ≤ 60MB (target 50MB)
- Container variant ≤ 180MB (target 150MB)
- Compressed images 30-50% smaller

### NFR-2: Scalability

**Concurrent Builds**:
- Support parallel architecture builds (x86_64 + arm64)
- Support parallel variant builds (base + container)
- No resource contention between builds

**Storage**:
- Support versioned storage (retain last 5 versions)
- Automatic cleanup of old versions
- Disk usage monitoring and alerts

**Distribution**:
- Local filesystem scales to 10-20 versions per variant
- Future migration to S3 for global distribution

### NFR-3: Security (SLSA 1.2 Compliance)

**Build Platform Attestation**:
- GitHub Actions provides SLSA Level 1.2 build platform
- Build provenance documented in metadata
- Reproducible builds (same inputs → same outputs)

**Image Signing**:
- All images signed with cosign/sigstore
- Keyless signing via GitHub OIDC
- Signatures published alongside images
- Verification required before VM launch

**SBOM**:
- Complete inventory of all components
- SPDX or CycloneDX format
- Machine-readable for vulnerability scanning
- Updated with every build

**Access Control**:
- GitHub Secrets for Docker Hub credentials
- OIDC for cosign signing (no long-lived keys)
- Repository access restricted to maintainers

### NFR-4: Reliability

**Build Success Rate**: ≥ 99%

**Error Handling**:
- Retry on transient failures (network, rate limits)
- Fail fast on permanent failures (image not found)
- Actionable error messages (not raw stack traces)
- Notifications on build failure

**Monitoring**:
- Build duration metrics
- Build success/failure rate
- Image size trends
- Boot time benchmarks

**Disaster Recovery**:
- Rebuild from source on corruption
- Rollback to previous version
- Manual override for emergency builds

### NFR-5: Maintainability

**Code Quality**:
- Workflow YAML follows best practices
- Scripts are shellcheck-clean
- Idempotent operations (re-run safe)
- Clear variable naming and comments

**Documentation**:
- Architecture overview
- Operator runbook
- Troubleshooting guide
- Version management procedures

**Observability**:
- Build logs retained in GitHub Actions
- Metadata tracks build history
- Audit trail for all builds

---

## 6. Task Breakdown (Backlog Tasks)

The following tasks have been created in the backlog and are ready for implementation:

### High-Priority Tasks (Critical Path)

1. **task-24**: Docker Image Extraction Mechanism
   - Implement extraction for flowspec-agents and nanofuse-gateway
   - Support multi-architecture (x86_64, arm64)
   - Handle layer merging and permissions

2. **task-25**: Base Variant Rootfs Builder (~50MB)
   - Alpine + Python + flowspec-cli
   - tini init system
   - Supports 60% of agent types

3. **task-26**: Container Variant Rootfs Builder (~150MB)
   - Base + containerd + nerdctl
   - Rootless container builds
   - Supports 40% of agent types

4. **task-27**: Init System Integration (tini + AGENT_TYPE Routing)
   - Configure tini as /sbin/init
   - Entrypoint script for agent routing
   - Signal handling and zombie reaping

5. **task-28**: Supply Chain Security (cosign + SBOMs)
   - Image signing with cosign/sigstore
   - SBOM generation with syft/trivy
   - SLSA 1.2 compliance

6. **task-29**: GitHub Actions Build Pipeline
   - Workflow triggers (releases, manual)
   - Multi-architecture builds
   - End-to-end automation

7. **task-31**: Integration Testing Suite for Rootfs Pipeline
   - Test Docker extraction
   - Test Firecracker boot (<200ms)
   - Test agent type selection
   - Test signature verification

### Medium-Priority Tasks (Post-MVP)

8. **task-30**: Local Filesystem Storage Integration
   - Content-addressable naming
   - Version management
   - nanofuse CLI integration

9. **task-32**: Rootfs Pipeline Documentation
   - Architecture overview
   - Operator guide
   - Troubleshooting runbook
   - SLSA compliance documentation

### Task Dependencies

```
task-24 (Docker Extraction)
  ├─> task-25 (Base Variant)
  │     └─> task-27 (Init System)
  │           └─> task-31 (Integration Tests)
  ├─> task-26 (Container Variant)
  │     └─> task-27 (Init System)
  │           └─> task-31 (Integration Tests)
  ├─> task-28 (Security)
  │     └─> task-29 (GitHub Actions)
  └─> task-30 (Storage)
        └─> task-32 (Documentation)
```

**Critical Path**: task-24 → task-25 → task-27 → task-31 → task-29

**Estimated Duration**: 2-3 development sessions

---

## 7. Discovery and Validation Plan

### Discovery Phase (Pre-Implementation)

#### Validation 1: Docker Extraction Feasibility

**Hypothesis**: Docker export can extract flowspec-agents into bootable rootfs

**Test**:
1. Pull flowspec-agents:latest from Docker Hub
2. Run `docker export` to extract filesystem
3. Merge layers and validate structure
4. Test Firecracker boot with extracted rootfs

**Success Criteria**:
- Extraction preserves permissions and symlinks
- Boot succeeds with <200ms boot time
- Agent executes test task successfully

**Go/No-Go Decision**: Proceed to full implementation if boot succeeds

---

#### Validation 2: SLSA Compliance Approach

**Hypothesis**: cosign + syft provide SLSA 1.2 compliance

**Test**:
1. Sign test image with cosign in GitHub Actions
2. Generate SBOM with syft
3. Verify signature with cosign verify
4. Review SBOM completeness

**Success Criteria**:
- Signature generation succeeds
- Signature verification succeeds
- SBOM includes all packages
- Security audit approves approach

**Go/No-Go Decision**: Proceed if security audit approves

---

#### Validation 3: Build Time Performance

**Hypothesis**: Full pipeline completes in < 20 minutes

**Test**:
1. Build prototype pipeline with all steps
2. Measure duration of each step
3. Identify bottlenecks
4. Optimize slow steps

**Success Criteria**:
- End-to-end build ≤ 20 minutes (P95)
- No single step > 10 minutes
- Parallel builds reduce total time

**Go/No-Go Decision**: Proceed if performance targets met

---

### Validation Phase (During Implementation)

#### Validation 4: Multi-Architecture Support

**Test**:
- Build both x86_64 and arm64 variants
- Test Firecracker boot on both architectures
- Verify agent execution on both

**Success Criteria**:
- Both architectures build successfully
- Boot time <200ms on both
- No architecture-specific issues

---

#### Validation 5: Agent Type Routing

**Test**:
- Boot VM with AGENT_TYPE=tech-writer (base variant)
- Boot VM with AGENT_TYPE=backend-engineer (container variant)
- Verify correct agent executes
- Test all supported agent types

**Success Criteria**:
- All agent types route correctly
- No missing dependencies
- AGENT_TYPE env var parsed correctly

---

### Go/No-Go Decision Points

#### Go/No-Go 1: After Discovery Phase
**Criteria**:
- [ ] Docker extraction produces bootable rootfs
- [ ] SLSA compliance approach approved by security
- [ ] Build time targets achievable

**Decision**: Proceed to implementation OR pivot approach

---

#### Go/No-Go 2: After Prototype Build
**Criteria**:
- [ ] Base variant builds successfully
- [ ] Container variant builds successfully
- [ ] Both variants boot in Firecracker <200ms
- [ ] Agent type routing works

**Decision**: Proceed to full automation OR debug issues

---

#### Go/No-Go 3: Before Production Release
**Criteria**:
- [ ] All 10 success criteria met (see Section 8)
- [ ] Integration tests pass
- [ ] Security audit approves SLSA compliance
- [ ] Documentation complete

**Decision**: Release to production OR address gaps

---

## 8. Acceptance Criteria and Testing

### Master Acceptance Criteria (10 Success Criteria)

From the next-steps document, all 10 criteria must be met:

- [ ] **AC-1**: GitHub Actions workflow exists and is triggered correctly
- [ ] **AC-2**: Can extract flowspec-agents Docker image from Docker Hub
- [ ] **AC-3**: Can extract nanofuse-gateway Docker image from Docker Hub
- [ ] **AC-4**: Builds base variant (~50MB) with Alpine + Python + flowspec-cli
- [ ] **AC-5**: Builds container variant (~150MB) with base + containerd + nerdctl
- [ ] **AC-6**: Images are signed with cosign/sigstore
- [ ] **AC-7**: SBOMs are generated with syft/trivy
- [ ] **AC-8**: Images stored in /var/lib/nanofuse/rootfs/ with correct structure
- [ ] **AC-9**: Can boot extracted rootfs in Firecracker with <200ms boot time
- [ ] **AC-10**: Agent type selection works via AGENT_TYPE environment variable

### Additional Quality Gates

#### Security Quality Gates
- [ ] SLSA 1.2 compliance verified by security audit
- [ ] All images signed with valid signatures
- [ ] Signatures verify with cosign verify
- [ ] SBOMs include 100% of installed packages
- [ ] No secrets or credentials in images
- [ ] Images run as non-root user

#### Performance Quality Gates
- [ ] Build completes in ≤ 20 minutes (P95)
- [ ] Base variant size ≤ 60MB
- [ ] Container variant size ≤ 180MB
- [ ] Boot time < 200ms for both variants
- [ ] No performance regression from baseline

#### Reliability Quality Gates
- [ ] Build success rate ≥ 99% over 30 days
- [ ] All integration tests pass
- [ ] Rollback to previous version succeeds
- [ ] Error messages are actionable
- [ ] Notifications work for success/failure

### Test Coverage Requirements

#### Unit Tests (Component Level)
- Docker extraction logic
- Layer merging and whiteout handling
- Entrypoint script routing logic
- Version management functions

**Coverage Target**: ≥ 80% for new code

#### Integration Tests (System Level)
- End-to-end Docker extraction
- Firecracker boot for both variants
- Agent type selection for all agents
- Signature generation and verification
- SBOM generation and completeness
- Storage and version management

**Coverage Target**: 100% of critical paths

#### Contract Tests (Interface Level)
- nanofuse CLI can list images
- nanofuse API can select rootfs variant
- Firecracker boot parameters correct
- Metadata format is correct

**Coverage Target**: 100% of public interfaces

#### Performance Tests
- Build duration benchmarks
- Boot time benchmarks (<200ms)
- Image size validation
- Parallel build scalability

**Coverage Target**: All NFRs validated

---

## 9. Dependencies and Constraints

### External Dependencies

#### Docker Hub
- **Dependency**: Access to jpoley/flowspec-agents:latest
- **Risk**: Rate limiting, authentication failures, image unavailability
- **Mitigation**: Authenticate pulls with credentials, implement retry logic

#### GitHub Actions
- **Dependency**: GitHub-hosted runners for builds
- **Risk**: Runner availability, build minutes quota
- **Mitigation**: Self-hosted runners (future), monitor quota usage

#### cosign/sigstore
- **Dependency**: Keyless signing via GitHub OIDC
- **Risk**: Service availability, OIDC token expiration
- **Mitigation**: Retry logic, fallback to local signing (dev only)

#### syft/trivy
- **Dependency**: SBOM generation tools
- **Risk**: Tool bugs, format changes
- **Mitigation**: Pin tool versions, validate SBOM output

### Internal Dependencies

#### flowspec-agents Repository
- **Dependency**: Release tags trigger builds
- **Risk**: Breaking changes, tag format changes
- **Mitigation**: Version pinning, integration tests

#### nanofuse-gateway Repository
- **Dependency**: Release tags trigger builds (when available)
- **Risk**: Image not yet published
- **Mitigation**: Graceful degradation, extract from local build

#### nanofuse CLI
- **Dependency**: Must support image listing and selection
- **Risk**: CLI not yet implemented
- **Mitigation**: Prioritize CLI development in parallel

### Technical Constraints

#### Firecracker Limitations
- **Constraint**: Requires specific kernel cmdline and init system
- **Impact**: tini must be configured correctly for signal handling
- **Mitigation**: Test boot extensively, document requirements

#### SLSA 1.2 Requirements
- **Constraint**: Build platform attestation required
- **Impact**: GitHub Actions is mandatory build platform
- **Mitigation**: No local builds for production (testing only)

#### Image Size Constraints
- **Constraint**: Larger images degrade boot time
- **Impact**: Must optimize for size while maintaining functionality
- **Mitigation**: Multi-stage builds, package cleanup, compression

### Operational Constraints

#### Storage Capacity
- **Constraint**: Local filesystem has limited capacity
- **Impact**: Must implement version retention and cleanup
- **Mitigation**: Configurable retention (default 5 versions)

#### Build Frequency
- **Constraint**: Frequent upstream releases trigger many builds
- **Impact**: Build minutes quota consumption
- **Mitigation**: Debounce builds, skip if no changes

#### Manual Config Updates
- **Constraint**: Version selection requires manual config edit
- **Impact**: Not fully automated (MVP limitation)
- **Mitigation**: Document process, plan auto-promotion for Phase 2

### Assumptions

1. **flowspec-agents image exists on Docker Hub** (confirmed)
2. **nanofuse-gateway will publish to Docker Hub** (planned)
3. **Firecracker supports extracted rootfs** (validated in ARCH decisions)
4. **GitHub Actions provides sufficient build capacity** (assumed)
5. **Platform engineers can edit YAML config files** (reasonable assumption)

---

## 10. Success Metrics (Outcome-Focused)

### North Star Metric

**Metric**: Time from flowspec-agents release to bootable Firecracker image

**Current State**: ∞ (undefined manual process)
**Target State**: ≤ 30 minutes (fully automated)
**Measurement**: GitHub Actions build duration + publish time

**Why This Metric**:
- Directly measures automation value
- Captures end-to-end pipeline performance
- Reflects user-facing outcome (image availability)
- Aligns with business goal (rapid deployment)

---

### Leading Indicators (Team Controls)

#### Build Success Rate
**Target**: ≥ 99% over rolling 30-day window
**Measurement**: (Successful builds / Total builds) × 100
**Why**: Reliability is critical for automation trust

#### Build Duration (P95)
**Target**: ≤ 20 minutes at 95th percentile
**Measurement**: GitHub Actions build time metrics
**Why**: Fast feedback enables rapid iteration

#### Zero Manual Intervention
**Target**: 100% of builds require no human action
**Measurement**: Manual trigger count / Total builds
**Why**: Automation value depends on hands-off operation

#### Signature Verification Rate
**Target**: 100% of published images signed and verified
**Measurement**: Signed images / Total images
**Why**: Security compliance is non-negotiable

---

### Lagging Indicators (Business Impact)

#### Platform Engineer Onboarding Time
**Target**: 50% reduction (from 4 hours to 2 hours)
**Measurement**: Time to first successful VM boot
**Why**: Reproducible setup improves developer experience

#### Security Audit Compliance Rate
**Target**: 100% compliance with SLSA 1.2
**Measurement**: Audit findings (zero non-compliance issues)
**Why**: Supply chain security is business requirement

#### Firecracker VM Boot Time
**Target**: < 200ms for both variants
**Measurement**: Firecracker boot duration benchmarks
**Why**: Fast boot is core value proposition

#### Production Incidents (Unsigned Images)
**Target**: Zero incidents from unsigned/unverified images
**Measurement**: Incident count over 90 days
**Why**: Security failures have business impact

---

### Operational Metrics (Health Monitoring)

#### Build Queue Time
**Target**: < 1 minute (immediate start)
**Measurement**: GitHub Actions queue time
**Why**: Delays reduce responsiveness to upstream changes

#### Storage Utilization
**Target**: < 80% capacity in /var/lib/nanofuse/rootfs/
**Measurement**: Disk usage monitoring
**Why**: Full disk blocks new builds

#### SBOM Completeness
**Target**: 100% of packages documented
**Measurement**: Manual audit of SBOM vs image contents
**Why**: Incomplete SBOMs reduce security value

#### Image Size Trend
**Target**: No regression (stay within size targets)
**Measurement**: Base ≤ 60MB, Container ≤ 180MB
**Why**: Size growth degrades boot time

---

### User Satisfaction Metrics

#### Platform Engineer NPS
**Target**: NPS ≥ 50 (Promoter score)
**Measurement**: Survey question: "How likely are you to recommend this pipeline?"
**Why**: User satisfaction indicates value delivery

#### Error Message Quality
**Target**: 80% of failures resolved without escalation
**Measurement**: (Self-resolved failures / Total failures) × 100
**Why**: Actionable errors reduce support burden

#### Documentation Completeness
**Target**: Zero documentation gaps reported
**Measurement**: Feedback from platform engineers
**Why**: Complete docs enable self-service

---

### Success Dashboard (Summary View)

| Metric Category | Key Metric | Target | Current | Status |
|-----------------|------------|--------|---------|--------|
| **North Star** | Time to bootable image | ≤ 30 min | ∞ | 🔴 Not Started |
| **Leading** | Build success rate | ≥ 99% | N/A | ⚪ Baseline TBD |
| **Leading** | Build duration (P95) | ≤ 20 min | N/A | ⚪ Baseline TBD |
| **Leading** | Manual intervention | 0% | N/A | ⚪ Baseline TBD |
| **Lagging** | Boot time | < 200ms | N/A | ⚪ Baseline TBD |
| **Lagging** | SLSA compliance | 100% | 0% | 🔴 Not Compliant |
| **Operational** | Storage utilization | < 80% | N/A | ⚪ Monitoring TBD |
| **User Satisfaction** | Platform engineer NPS | ≥ 50 | N/A | ⚪ Survey TBD |

---

### Measurement Plan

#### Automated Metrics (GitHub Actions)
- Build duration
- Build success/failure rate
- Image sizes
- Manual trigger count

#### Periodic Metrics (Weekly)
- Boot time benchmarks
- Storage utilization
- SBOM completeness audits

#### Milestone Metrics (Quarterly)
- Platform engineer onboarding time
- Security audit compliance
- User NPS surveys

#### Continuous Metrics (Real-Time)
- Build queue time
- Storage capacity alerts
- Incident count (unsigned images)

---

## Appendix A: Architecture Decision References

This PRD implements the following approved architecture decisions:

- **ARCH-003**: Tiered Rootfs Variants (base + container)
- **ARCH-004**: Pre-baked Rootfs Extraction (Docker export)
- **ARCH-010**: flowspec-agents Alpine Base Unchanged (no code changes)
- **ARCH-019**: Container Images as Build Artifacts (not runtime)
- **ARCH-023**: SLSA 1.2 Supply Chain Security (signing + SBOMs)
- **ARCH-026**: Automated Rootfs Build Pipeline (GitHub Actions)
- **ARCH-027**: Local Filesystem Distribution (MVP, S3 later)
- **ARCH-028**: Tag-Based Versioning with Config Consumption

Full details: `/Users/jasonpoley/prj/ps/nanofuse/docs/building/dec22-architecture-decisions.jsonl`

---

## Appendix B: Related Documents

- **Next Steps Document**: `/Users/jasonpoley/prj/ps/nanofuse/docs/building/next-steps-dec22.md`
- **Architecture Decisions (JSONL)**: `/Users/jasonpoley/prj/ps/nanofuse/docs/building/dec22-architecture-decisions.jsonl`
- **Architecture Decisions (Markdown)**: `/Users/jasonpoley/prj/ps/nanofuse/docs/building/dec22-decisions.md`
- **Backlog Tasks**: `/Users/jasonpoley/prj/ps/nanofuse/backlog/tasks/task-24` through `task-32`

---

## Appendix C: Glossary

| Term | Definition |
|------|------------|
| **SLSA** | Supply-chain Levels for Software Artifacts - security framework |
| **SBOM** | Software Bill of Materials - inventory of components |
| **cosign** | Container signing and verification tool (sigstore project) |
| **syft** | SBOM generation tool |
| **tini** | Lightweight init system for containers/VMs |
| **Firecracker** | AWS microVM technology for secure isolation |
| **flowspec-agents** | AI agent orchestration platform (13+ agent types) |
| **nanofuse-gateway** | Network proxy for logging and security |
| **DVF+V** | Desirability, Viability, Feasibility, Value risk framework |
| **North Star Metric** | Single metric that best captures core product value |
| **OKR** | Objectives and Key Results |

---

**Approval Sign-Off**:

- [ ] Product Manager: _____________________ Date: _____
- [ ] Platform Engineering Lead: _____________________ Date: _____
- [ ] Security Lead: _____________________ Date: _____
- [ ] DevOps Lead: _____________________ Date: _____

---

**Document History**:

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-12-22 | @pm-planner | Initial draft based on ARCH decisions and next-steps |

---

**Next Steps**:

1. Review and approve this PRD
2. Execute `/flow:assess` to analyze complexity
3. Execute `/flow:specify` to create technology-agnostic spec.md
4. Execute `/flow:plan` to generate technical implementation plan
5. Execute `/flow:implement` to build the pipeline
6. Execute `/flow:validate` for QA, security, and documentation review
7. Execute `/flow:operate` for CI/CD integration and operational setup
