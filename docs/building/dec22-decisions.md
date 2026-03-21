# Resolved Architecture Decisions (Dec 22, 2025)

This document contains architecture questions that have been definitively answered and decided.

---

## Q1: Task Result Communication ✅ DECIDED

**Decision:** File-based with future evolution to nanofuse log collector API

**Answer:**
- **Phase 1 (MVP):** File-based - Write to `/mnt/results/output.json` and artifacts
- **Phase 2:** Evolve to nanofuse log collector interface/API for aggregation and forwarding

**Rationale:**
- File-based is simplest for MVP and consistent with task input mechanism
- Provides path to centralized logging without blocking initial implementation
- nanofuse can monitor mounted drive and forward to collector

**Status:** ✅ Approved - Implement file-based, plan for API evolution

---

## Q2: Secrets Management ✅ DECIDED

**Decision:** Short-lived GitHub tokens + subscription-based cached logins

**Answer:**
- **Primary:** Short-lived GitHub tokens (refreshable)
- **Secondary:** Subscription-based cached logins for services
- **Delivery:** Environment variables where possible, mounted secrets for exceptions

**Secrets types:**
1. GitHub tokens (short-lived, refreshable)
2. Cached authentication sessions (subscription services)
3. Optional: LLM API keys (supported but not preferred)

**Rationale:**
- Minimal sensitive data exposure
- Short-lived credentials reduce blast radius
- Environment variables simplest for short-lived tokens
- Mounted secrets for exceptions (large creds, binary tokens)

**Status:** ✅ Approved - Short-lived tokens primary, cached sessions secondary

---

## Q3: Network Isolation Model ✅ DECIDED

**Decision:** Transparent proxy via nanofuse-gateway (separate container, same microVM)

**Answer:**
- **Architecture:** nanofuse-gateway runs as separate process in same microVM
- **Initial deployment:** Separate container, flattened into same Firecracker microVM
- **Function:** All traffic proxied through gateway for logging and security
- **Network:** TAP interface with gateway as default route

**Capabilities:**
- Request/response logging
- Domain allow-listing
- Rate limiting
- Security monitoring

**Rationale:**
- Provides visibility and control without complex host-side networking
- Gateway process isolation within same VM
- Can evolve to separate VM if needed
- Simplifies VM networking (single gateway endpoint)

**Status:** ✅ Approved - Implement nanofuse-gateway in-VM proxy

---

## Q4: Package Installation Strategy ✅ DECIDED

**Decision:** Pre-installed toolchains, exception-based on-demand packages

**Answer:**
- **Pre-installed (rootfs):** Python, Node.js, Rust, Go toolchains + common packages
- **On-demand:** Rare packages installed via pip/npm/cargo during task
- **By exception only:** Use cases requiring unusual dependencies

**Trade-off:**
- Larger rootfs (~200-300MB) but faster task startup
- No network dependency for 90%+ of tasks
- Predictable, reproducible environment

**Rationale:**
- Sub-200ms boot time incompatible with on-demand installs
- Most AI coding tasks use common dependencies
- Exceptions can still install packages (network available)

**Status:** ✅ Approved - Fat rootfs with common packages

---

## Q5: Persistent Workspace ✅ DECIDED

**Decision:** Persistent per-user workspace with git/worktree workflow

**Answer:**
- **Primary persistence:** Git commits pushed to remote branches
- **Workspace:** Per-user persistent workspace mounted at `/workspace/<user-id>`
- **Alternative:** git worktree for multi-branch workflows

**Downsides of persistent workspace:**
- State accumulation over time (disk usage)
- Potential for stale data
- Cleanup complexity
- Security: residual data between sessions

**Mitigations:**
- Periodic cleanup of old workspaces
- Workspace quotas per user
- Git-based workflow encourages pushing work
- Ephemeral task state in `/mnt/task/` (cleaned per-task)

**Rationale:**
- Users expect persistent workspace (IDE-like experience)
- Git provides natural state checkpoint mechanism
- Enables multi-session workflows (resume where left off)

**Status:** ✅ Approved - Per-user persistent workspace + git workflow

---

## Q6: VM Timeout Policy ✅ DECIDED

**Decision:** Command-driven lifecycle, no autonomous timeouts in microVM

**Answer:**
- **Trigger:** VM lifecycle controlled by nanofuse API commands only
- **No autonomous timeouts:** microVM does not self-terminate
- **Host responsibility:** nanofuse daemon enforces timeouts and monitors health

**Rationale:**
- Keep intelligence in host, not in microVM
- Simpler microVM implementation (less code to maintain)
- Centralized policy management
- Easier to adjust timeouts per-agent or per-user

**Parameters (host-side):**
- Max task runtime: Configurable per agent type
- Idle timeout: Configurable, default disabled
- Heartbeat: Host monitors VM health

**Status:** ✅ Approved - Host-managed lifecycle, dumb VMs

---

## Q7: Crash Detection & Recovery ✅ DECIDED

**Decision:** Crash and retry with CRIU investigation

**Answer:**
- **Detection:** Host monitors VM exit codes and heartbeats
- **Recovery:** Automatic retry with fresh VM
- **Advanced:** Investigate CRIU for VM snapshots/checkpoints

**CRIU Considerations:**
- **Use case:** Snapshot VM state for crash recovery
- **Challenge:** Firecracker CRIU support is experimental
- **Risk:** Snapshot corruption could lose state
- **Alternative:** Application-level checkpointing (git commits)

**Questions for CRIU:**
1. How to snapshot filesystem without risk?
2. Can we snapshot just memory state, not FS?
3. Does Firecracker support CRIU officially?

**Status:** ✅ Approved - Crash/retry immediate, CRIU research for Phase 2

---

## Q8: VM Reuse (Warm Pools) ✅ DECIDED

**Decision:** Cold start only, with session reuse exception

**Answer:**
- **Default:** Fresh microVM per task (~200ms boot acceptable)
- **Exception:** Session reuse to avoid repeated logins
  - *Example:* 5 Claude sessions shouldn't require 5 separate logins
- **Mechanism:** Session tokens cached in persistent workspace

**Rationale:**
- 200ms boot is fast enough for most use cases
- Security benefits of clean VMs outweigh marginal speed gains
- Session token reuse provides UX benefit without VM reuse complexity

**Status:** ✅ Approved - Cold start default, session token caching for UX

---

## Q9: Resource Monitoring ✅ DECIDED

**Decision:** TBD - Host-based or in-VM telemetry, open to suggestions

**Answer:**
- **Current stance:** No strong opinion, open to best practices
- **Options being considered:**
  1. Host-based monitoring via Firecracker API
  2. In-VM telemetry agent (lightweight)
  3. Hybrid approach

**Questions:**
- What metrics are most critical?
- What's the overhead of in-VM telemetry?
- Should metrics be real-time or batch?

**Status:** ⚠️ Open for recommendations - Need input on best practices

---

## Q10: Docker-in-Docker Requirements ✅ DECIDED

**Decision:** Containers as packages, flattened into rootfs

**Answer:**
- **flowspec-agents:** Extracted from Docker image, flattened into rootfs
- **nanofuse-gateway:** Extracted from Docker image, flattened into rootfs
- **No runtime containerd:** Both run as native processes, not containers
- **Exception:** Platform/SRE agents MAY need containerd for building user containers

**Architecture:**
```
Build time:
  flowspec-agents:latest (Docker) → extract → rootfs
  nanofuse-gateway:latest (Docker) → extract → rootfs

Runtime:
  Firecracker VM:
    ├─ flowspec-agent (native process)
    └─ nanofuse-gateway (native process)
```

**Status:** ✅ Approved - Containers as build artifacts, not runtime

---

## Q11: Elevated Privileges ✅ DECIDED

**Decision:** No root access for agents

**Answer:**
- **All agents run as non-root user**
- **Rationale:**
  - Minimal credentials in VMs
  - Mostly ephemeral work datasets
  - Code validated before pushing to remote branches

**Security posture:**
- Agents cannot modify system files
- Cannot install system packages (only user-level)
- Cannot bind privileged ports (<1024)
- Rootless container builds if needed (buildkit, podman)

**Status:** ✅ Approved - Non-root agents only

---

## Q12: Sub-process Spawning ✅ DECIDED

**Decision:** Unrestricted subprocess spawning within resource limits

**Answer:**
- **Allowed:** Agents can spawn arbitrary subprocesses
- **Monitoring:** Keep an eye on process tree growth
- **Limits:** Firecracker cgroup limits prevent abuse
- **Management:** tini reaps zombie processes

**Status:** ✅ Approved - Allow subprocesses, monitor resource usage

---

## Q13: Log Collection ✅ DECIDED

**Decision:** nanofuse log collector API

**Answer:**
- **Primary:** Structured logging to nanofuse log collector API
- **Fallback:** File-based logs in `/mnt/results/logs/` for persistence
- **Format:** JSON structured logs for parsing

**Architecture:**
```
Agent → nanofuse log collector API (HTTP/gRPC)
     ↓
nanofuse aggregates/forwards logs to:
  - Central logging (Loki, CloudWatch, etc.)
  - File storage for audit
```

**Status:** ✅ Approved - Implement log collector API

---

## Q14: Threat Model ✅ DECIDED

**Decision:** Document complete attack matrix for microVMs and devcontainers

**Answer:**
- **Action item:** Create comprehensive threat model document
- **Scope:**
  1. microVM attack vectors
  2. devcontainer attack vectors
  3. Supply chain attacks
  4. Data exfiltration scenarios
  5. Resource abuse scenarios

**Status:** 🚧 In progress - Threat model document to be created

---

## Q15: Supply Chain Security ✅ DECIDED

**Decision:** SLSA 1.2 + image/hash signing + SBOMs

**Answer:**
- **SLSA Level:** 1.2 (build platform attestation)
- **Signing:** Image or hash signing with cosign/sigstore
- **Provenance:** SBOM (Software Bill of Materials) for all rootfs images
- **Verification:** Signature verification before VM launch

**Implementation:**
```
GitHub Actions builds rootfs
  ↓
Generate SBOM (syft, trivy)
  ↓
Sign image with cosign
  ↓
Publish with attestations
  ↓
nanofuse verifies signature before launch
```

**Status:** ✅ Approved - SLSA 1.2 with signing and SBOMs

---

## Q16: Data Access Control ✅ DECIDED

**Decision:** Case-by-case basis via MCP or mounted shares

**Answer:**
- **Default:** Only task-specific data mounted
- **MCP:** Agents can request additional context via Model Context Protocol
- **Mounted shares:** Explicit shares for specific use cases
- **Audit:** All data access logged

**Principle:** Least privilege, explicit grants, full audit trail

**Status:** ✅ Approved - Task-specific default, MCP/shares for exceptions

---

## Q17: Output Sanitization ✅ DECIDED

**Decision:** Minimal sensitive data exposure (short-lived secrets only)

**Answer:**
- **Philosophy:** Don't put sensitive data in VMs in the first place
- **Secrets:** Only short-lived tokens (auto-expire)
- **Sanitization:** Minimal pattern-based redaction for known patterns
- **Trust model:** Agents trusted within ephemeral scope

**Rationale:**
- Short-lived credentials (GitHub tokens, session caches) expire quickly
- Ephemeral VMs destroyed after task
- Network isolation prevents exfiltration
- Output review catches edge cases

**Status:** ✅ Approved - Prevention over detection, minimal sanitization

---

## Q18: Rootfs Build Pipeline ✅ DECIDED

**Decision:** GitHub Actions triggered by flowspec-agents and nanofuse-gateway updates

**Answer:**
- **Triggers:**
  1. flowspec-agents new release (tag/release event)
  2. nanofuse-gateway new release (tag/release event)
  3. Manual trigger (workflow_dispatch)
  4. Security patches (scheduled or manual)

**Pipeline:**
```
Trigger event
  ↓
Pull source images (Docker)
  ↓
Extract and build rootfs variants
  ↓
Sign with cosign + generate SBOMs
  ↓
Publish to registry (GHCR or S3)
  ↓
Update nanofuse config (version references)
```

**Status:** ✅ Approved - Automated GitHub Actions pipeline

---

## Q19: Image Distribution ✅ DECIDED

**Decision:** Local filesystem for now, scale to object storage later

**Answer:**
- **Phase 1 (current):** Local filesystem (`/var/lib/nanofuse/rootfs/`)
  - Pre-downloaded images
  - Simple file paths
  - No network dependency at runtime
- **Phase 2 (scale):** Object storage (S3, GCS) with local caching
  - Content-addressable (hash-based names)
  - Versioning
  - Global distribution

**Rationale:**
- Local filesystem simplest for MVP
- ~150MB images load fast from local disk
- Can migrate to S3 without API changes (transparent cache)

**Status:** ✅ Approved - Local FS now, S3 later

---

## Q20: Update Strategy ✅ DECIDED

**Decision:** Tag-based versioning with config-driven consumption

**Answer:**
- **Versioning:** Semantic versioning tags (v1.2.3)
- **Config:** nanofuse config specifies which version to use
- **Updates:** New versions built, consumed based on config update
- **Rollback:** Change config to previous version tag

**Deployment process:**
```
1. New rootfs version built (v1.2.3)
2. Published with tag
3. Testing on subset of VMs
4. Update nanofuse config to use v1.2.3
5. New VMs use new version
6. Rollback: Update config back to v1.2.2
```

**Status:** ✅ Approved - Immutable versions, config-driven consumption

---

## Summary

**Total Decided:** 20/20
**Status:**
- ✅ Fully decided: 18
- 🚧 In progress: 1 (Q14 - threat model doc)
- ⚠️ Open for input: 1 (Q9 - resource monitoring)

**Next Steps:**
1. Create threat model document (Q14)
2. Decide on resource monitoring approach (Q9)
3. Create design branch and PR with these decisions

---

**Document History:**
- 2025-12-22: All questions answered and decided
