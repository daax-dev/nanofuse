# Open Architecture Questions (Dec 22, 2025)

This document tracks unmade decisions and open questions following the flowspec-agents + Firecracker architecture analysis.

**Status Legend:**
- 🔴 **Critical** - Blocks implementation
- 🟡 **Important** - Needed soon
- 🟢 **Nice to have** - Can defer

---

## Architecture & Design

### 🔴 Q1: Task Result Communication
**Question:** How do agents communicate task results back to nanofuse?

**Options:**
- **A. File-based** - Write to `/mnt/results/output.json` (mounted drive)
- **B. stdout/stderr** - Capture VM console output
- **C. Callback API** - Agent POSTs to nanofuse endpoint
- **D. Metadata service** - nanofuse polls metadata endpoint

**Implications:**
- File-based is simplest and already decided for task input
- stdout requires parsing and has size limits
- Callback requires network and error handling
- Metadata service adds complexity

**Recommendation:** File-based (consistent with input mechanism)

---

### 🔴 Q2: Secrets Management
**Question:** How are API keys and secrets provided to agents?

**Options:**
- **A. Environment variables** - Set in Firecracker boot params
- **B. Mounted secret file** - Read-only drive with secrets
- **C. Metadata service** - Like AWS EC2 instance metadata
- **D. Host-provided socket** - Unix socket for secret retrieval

**Security considerations:**
- Environment visible in `/proc/<pid>/environ`
- Mounted file needs encryption at rest
- Metadata service enables credential rotation
- Socket provides best isolation but adds complexity

**Recommendation:** Start with mounted secret file (B), evolve to metadata service (C) if rotation needed

---

### 🟡 Q3: Network Isolation Model
**Question:** What network access should agents have?

**Options:**
- **A. Full internet** - No restrictions (simplest)
- **B. Allow-list** - Only approved domains (package registries, git hosts, AI APIs)
- **C. Transparent proxy** - All traffic through proxy for logging
- **D. Full isolation** - No network (extreme)

**Required access:**
- Package managers (PyPI, npm, cargo, etc.)
- Git repositories (GitHub, GitLab, etc.)
- AI APIs (OpenAI, Anthropic, etc.)
- Documentation sites (docs.python.org, etc.)

**Recommendation:** Allow-list (B) with transparent proxy (C) for logging

---

### 🟡 Q4: Package Installation Strategy
**Question:** How are runtime dependencies handled?

**Options:**
- **A. Pre-installed** - All common packages in rootfs
- **B. On-demand** - Install via pip/npm during task
- **C. Cached** - Shared package cache on host
- **D. Hybrid** - Common pre-installed, rare on-demand

**Trade-offs:**
- Pre-installed: Faster startup, larger image, may waste space
- On-demand: Smaller image, slower startup, requires network
- Cached: Fast + small, but complex management

**Recommendation:** Hybrid (D) - Python/Node/Rust toolchains pre-installed, packages on-demand with cache

---

### 🟡 Q5: Persistent Workspace
**Question:** Do agents need persistent storage across tasks?

**Use cases:**
- Git repository cache
- Package manager cache
- User workspace/history
- Build artifacts

**Options:**
- **A. Ephemeral** - Fresh VM every task (current assumption)
- **B. Persistent per-user** - Mounted user workspace
- **C. Shared cache** - Read-only cache for packages/repos

**Recommendation:** B. Persistent per-user workspace with git/worktree workflow

---

## VM Lifecycle Management

### 🔴 Q6: VM Timeout Policy
**Question:** When should VMs be terminated?

**Scenarios:**
- Task completes successfully
- Task exceeds max runtime
- Task hangs (no progress)
- VM becomes unresponsive

**Parameters needed:**
- Max task runtime (e.g., 1 hour)
- Idle timeout (e.g., 5 minutes no activity)
- Heartbeat interval (e.g., 30 seconds)

**Recommendation:** Define per-agent timeout configurations (some tasks take longer than others)

---

### 🟡 Q7: Crash Detection & Recovery
**Question:** How are crashed/hung VMs detected and handled?

**Detection methods:**
- VM exits unexpectedly
- No heartbeat for N seconds
- No file writes for N minutes
- OOM kills

**Recovery actions:**
- Terminate VM
- Collect crash logs/dumps
- Notify user of failure
- Optionally retry

**Recommendation:** Heartbeat-based detection + automatic cleanup + manual retry decision

---

### 🟢 Q8: VM Reuse (Warm Pools)
**Question:** Can VMs be reused across tasks to reduce cold start latency?

**Benefits:**
- Faster task startup (~150ms vs ~500ms)
- Better resource utilization

**Challenges:**
- State contamination between tasks
- Security concerns (residual data)
- Complexity of pool management

**Recommendation:** Defer to Phase 2; start with cold start for simplicity and security

---

### 🟡 Q9: Resource Monitoring
**Question:** How is VM resource usage monitored and enforced?

**Metrics needed:**
- CPU usage
- Memory usage
- Disk I/O
- Network bandwidth
- API call count/cost

**Enforcement:**
- Firecracker cgroups for CPU/memory limits
- Disk quota on mounted drives
- Network rate limiting
- AI API cost tracking via proxy

**Recommendation:** Use Firecracker resource limits + telemetry export to nanofuse

---

## Agent Capabilities

### 🔴 Q10: Docker-in-Docker Requirements
**Question:** Which agents need container building capabilities?

**Candidates:**
- **platform-engineer** - Definitely (infrastructure focus)
- **sre-agent** - Definitely (K8s, containers)
- **backend-engineer** - Maybe (might build Docker images)
- **frontend-engineer** - Unlikely (typically NPM builds)

**Recommendation:**
- Include containerd in `flowspec-container` rootfs
- Map platform-engineer, sre-agent, backend-engineer to container-enabled rootfs
- Document which agents have container capabilities

---

### 🟢 Q11: Elevated Privileges
**Question:** Do any agents need root/sudo access?

**Rationale:**
- Security best practice: minimize privileges
- Container builds can run rootless (buildkit, podman)
- System changes shouldn't be needed

**Recommendation:** NO agents get root; all run as non-root user

---

### 🟡 Q12: Sub-process Spawning
**Question:** How are agent-spawned processes managed?

**Use cases:**
- Running tests
- Building code
- Git operations
- Package installations

**Process management:**
- tini reaps zombie processes
- Resource limits apply to all processes
- Process tree termination on timeout

**Recommendation:** Allow unrestricted subprocess spawning within resource limits

---

### 🟡 Q13: Log Collection
**Question:** How are agent logs collected and stored?

**Log sources:**
- Agent stdout/stderr
- Application logs
- Build logs
- Test output

**Collection methods:**
- **A. Console output** - Firecracker serial console
- **B. Log files** - Written to `/mnt/results/logs/`
- **C. Syslog** - Forward to host syslog
- **D. Structured logging** - JSON to stdout

**Recommendation:** Structured JSON to stdout + log files in `/mnt/results/` for persistence

---

## Security

### 🔴 Q14: Threat Model
**Question:** What threats are we defending against?

**Threat scenarios:**
1. **Untrusted user code** - User provides malicious code to execute
2. **Compromised agent** - AI agent acts maliciously
3. **Supply chain attack** - Poisoned dependencies in rootfs
4. **Data exfiltration** - Agent leaks sensitive data
5. **Resource abuse** - Agent consumes excessive resources

**Mitigations:**
- VM isolation (Firecracker)
- Network restrictions
- Resource limits
- Secrets management
- Audit logging

**Recommendation:** Document formal threat model and mitigations

---

### 🟡 Q15: Supply Chain Security
**Question:** How are rootfs images verified and secured?

**SLSA compliance considerations:**
- Image signing (who built this?)
- Attestations (how was it built?)
- Provenance (what's in it?)
- Reproducible builds

**Implementation:**
- Sign rootfs images with sigstore/cosign
- Generate SLSA attestations
- Store in content-addressable registry
- Verify signatures before VM launch

**Recommendation:** Implement SLSA Level 2 for rootfs builds (aligns with nanofuse principles)

---

### 🟡 Q16: Data Access Control
**Question:** What data can agents access?

**Data categories:**
1. **Task input** - User-provided code, instructions
2. **Context** - Project files, git history
3. **Secrets** - API keys, credentials
4. **Results** - Previous task outputs
5. **External data** - Internet access

**Access model:**
- Principle of least privilege
- Only task-specific data mounted
- No access to other users' data
- Audit log of data access

**Recommendation:** Mount only task-specific data; implement audit logging

---

### 🟡 Q17: Output Sanitization
**Question:** How is sensitive data in agent outputs handled?

**Concerns:**
- API keys in logs
- PII in generated code
- Credentials in error messages
- Internal paths/hostnames

**Approaches:**
- **A. Pattern-based redaction** - Regex filters for known patterns
- **B. LLM-based detection** - Use AI to detect sensitive data
- **C. User review** - Flag suspicious outputs for review
- **D. No sanitization** - Trust agent + network isolation

**Recommendation:** Pattern-based redaction (A) for known secrets, user review for unknown

---

## Operations

### 🟡 Q18: Rootfs Build Pipeline
**Question:** How are rootfs images built and versioned?

**Build triggers:**
- flowspec-agents new release
- Security patch to base Alpine
- Manual rebuild request

**CI/CD pipeline:**
1. Monitor flowspec-agents releases
2. Extract Docker image
3. Build rootfs variants (base, container)
4. Sign images
5. Generate attestations
6. Push to registry
7. Update nanofuse image references

**Recommendation:** Automate with GitHub Actions, triggered by flowspec-agents releases

---

### 🟡 Q19: Image Distribution
**Question:** Where are rootfs images stored and how are they distributed?

**Options:**
- **A. OCI registry** - Standard container registry (e.g., Docker Hub)
- **B. Object storage** - S3, GCS, etc.
- **C. Local filesystem** - Pre-downloaded to nanofuse hosts
- **D. Content-addressable storage** - IPFS, etc.

**Requirements:**
- Fast download (~150MB in <1s)
- Versioning
- Access control
- Caching on hosts

**Recommendation:** Object storage (B) with local caching; OCI registry for metadata

---

### 🟢 Q20: Update Strategy
**Question:** How are rootfs images updated in production?

**Strategies:**
- **A. Immediate** - New tasks use new version immediately
- **B. Gradual** - Percentage-based rollout
- **C. Blue-Green** - Deploy alongside, switch traffic
- **D. Immutable** - Version pinning, manual updates

**Rollback considerations:**
- Can revert to previous version?
- How to handle in-flight tasks?

**Recommendation:** Immutable versioning (D) with gradual rollout (B) for major updates

---

## Summary Statistics

- **Total Questions:** 20
- **Critical (🔴):** 5 - Must answer before implementation
- **Important (🟡):** 13 - Should answer soon
- **Nice to have (🟢):** 2 - Can defer

## Next Steps

1. **Immediate:** Answer critical questions (Q1, Q2, Q10, Q14)
2. **Short-term:** Address important questions in design phase
3. **Long-term:** Revisit nice-to-have questions after MVP

---

**Document History:**
- 2025-12-22: Initial draft based on sequential thinking analysis
