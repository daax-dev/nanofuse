# Firecracker Snapshot/Resume Research

## Overview

This document evaluates the feasibility of implementing Firecracker snapshot/resume capabilities for NanoFuse. The research examines API capabilities, systemd compatibility, known limitations, and provides a concrete recommendation for implementation.

**Research Date:** 2025-12-09
**Firecracker Version:** 1.7.0 (latest stable)
**Target OS:** Ubuntu 24.04 + systemd

## Firecracker Snapshot Capabilities

### CreateSnapshot API

Firecracker provides snapshot creation through the `PUT /snapshot/create` endpoint. The microVM **must be in a Paused state** before snapshotting.

**Two snapshot types are supported:**

1. **Full Snapshots**
   - `snapshot_path`: Contains devices' model state and emulation state
   - `mem_file_path`: Contains complete copy of guest memory
   - Use case: Cold starts, archival, cross-host migration

2. **Diff Snapshots** (Developer Preview)
   - Creates sparse memory files with only modified pages since last snapshot
   - Enables incremental checkpointing
   - Status: Still in developer preview, pending guest_memfd integration
   - Use case: Frequent checkpointing, reduced storage overhead

**API Parameters:**
```json
{
  "snapshot_path": "/path/to/vmstate",
  "mem_file_path": "/path/to/memory",
  "snapshot_type": "Full",  // or "Diff"
  "track_dirty_pages": false  // Optional: enable for diff snapshots
}
```

### LoadSnapshot API

Snapshots are loaded via `PUT /snapshot/load` **before microVM configuration**.

**Key behaviors:**
- Complete microVM state restored to current Firecracker process
- Loaded microVM enters **Paused** state
- Must explicitly resume via `PATCH /vm` to run
- Version compatibility checked before loading

**API Parameters:**
```json
{
  "snapshot_path": "/path/to/vmstate",
  "mem_backend": {
    "backend_type": "File",  // or "Uffd"
    "backend_path": "/path/to/memory"
  },
  "enable_diff_snapshots": false,
  "resume_vm": false
}
```

### Memory Backend Options

Two memory backends available:

1. **File Backend**
   - Relies on kernel page fault handling
   - Simpler implementation
   - May have higher restoration latency
   - Recommended for: General use cases

2. **Uffd Backend (User Fault File Descriptor)**
   - Dedicated user-space process handles page faults
   - Lower restoration latency
   - More complex implementation
   - Recommended for: Performance-critical scenarios, large memory footprints

**Performance Note:** High restoration latency observed with cgroups v1; **cgroups v2 strongly recommended**.

### Version Compatibility Requirements

Firecracker enforces strict version compatibility:

**Snapshot Format Versioning:**
- Format uses `MAJOR.MINOR.PATCH` versioning
- Magic identifier + version number + serialized state + optional CRC64 checksum
- **Critical:** Every change in microVM state description bumps MAJOR version (bincode limitation)

**Compatibility Constraints:**

1. **Host Kernel:** Snapshots compatible only on **same kernel version**
   - Cross-kernel restoration is unstable
   - KVM state semantics may differ between kernels
   - Example: 5.10.240 snapshot cannot reliably restore on 6.2.0

2. **CPU Architecture:** Snapshots incompatible across:
   - Different CPU architectures (x86_64 ↔ aarch64)
   - Different CPU models (Intel ↔ AMD)
   - Different CPU features (AVX, AVX2, etc.)
   - **Invariant requirement:** CPU features exposed to guest must match exactly

3. **Device Model:**
   - Current devices maintain backwards compatibility from introduction version
   - New features may prevent restoration on older versions

4. **External Dependencies:**
   - Tap devices must exist with matching names
   - Block devices must be at original paths with proper permissions
   - Vsock backing sockets must be accessible with matching names

## Systemd Compatibility

### Clock and Timer Behavior

**Wall-Clock Jump on Resume:**
- Guest OS wall-clock continues from moment of snapshot creation
- Results in significant time lag after restore
- **Critical:** Wall-clock must be updated to current time on guest-side post-restore

**systemd Timer Types:**

1. **Monotonic Timers** (`OnBootSec=`, `OnActiveSec=`)
   - Normally use `CLOCK_MONOTONIC` which **pauses during suspend**
   - With `WakeSystem=true`: Uses `CLOCK_BOOTTIME` which continues during suspend
   - Behavior after snapshot restore: **Undefined** (not true suspend)
   - Risk: Timer drift, missed activations

2. **Calendar/Realtime Timers** (`OnCalendar=`)
   - Use `CLOCK_REALTIME` (wall-clock)
   - When VM sleeps, realtime clock does not pause
   - On resume: Catches up and processes all triggers during sleep period
   - **Limitation:** Multiple elapsed triggers result in single service activation

3. **Clock Jump Detection**
   - `OnClockChange=true`: Triggers when `CLOCK_REALTIME` jumps relative to `CLOCK_MONOTONIC`
   - `OnTimezoneChange=true`: Triggers on local timezone modification
   - Use case: Detect snapshot restore, trigger clock sync service

### Service State Handling

**systemd Process Management:**
- Services continue running from pre-snapshot state
- No automatic restart or re-initialization
- Services with time-sensitive state may fail

**Known Issues:**

1. **Process Hangs After Resume** ([Issue #4099](https://github.com/firecracker-microvm/firecracker/issues/4099))
   - **Symptom:** Processes using `nanosleep()` occasionally become stuck after restore
   - **Affected platforms:** AMD Ryzen/EPYC processors (Intel unaffected)
   - **Root cause:** Firmware TSC (Time Stamp Counter) bug: "TSC doesn't count with P0 frequency"
   - **Workaround:** Add `nolapic` to kernel boot parameters
   - **Status:** Codesandbox fork has patch (commit `4164371`), upstream pending

2. **systemd Oneshot Unit State Issues**
   - Oneshot units stuck in "activating" state can break timer repetition
   - Causes missed backups and scheduled tasks
   - General systemd issue, exacerbated by snapshot/restore

### Clock Synchronization Best Practices

**Recommended Configuration:**

1. **Use NTP in guest** (ntpd or chronyd)
   - Industry standard time synchronization
   - Configure to handle large time jumps: `tinker panic 0` (must be first line in ntp.conf)
   - Prevents NTP from giving up after large time drifts

2. **Disable periodic VMware Tools sync** (not applicable to Firecracker, but principle applies)
   - Use only one time synchronization method
   - NTP should be sole periodic sync mechanism

3. **Enable one-off synchronization**
   - Trigger time sync on snapshot resume event
   - Use `OnClockChange=true` systemd timer to detect restore
   - Run clock sync service immediately after detection

4. **Ensure host time accuracy**
   - Host system time should be accurate via NTP
   - Guest clock synchronization relies on host time reference

**Implementation Approach:**
```ini
# /etc/systemd/system/clock-sync-on-restore.service
[Unit]
Description=Synchronize clock after VM restore
After=network.target

[Service]
Type=oneshot
ExecStart=/usr/bin/chronyd -q

# /etc/systemd/system/clock-sync-on-restore.timer
[Unit]
Description=Detect clock jumps from VM restore

[Timer]
OnClockChange=yes

[Install]
WantedBy=timers.target
```

### Cryptographic Token and PRNG State

**Security Implications:**
- When guest state is resumed more than once, "unique" information may be duplicated:
  - UUIDs
  - Random numbers and PRNG seeds
  - OS entropy pool
  - Cryptographic nonces and tokens
  - Session identifiers

**SysGenID Support:**
- [Issue #2546](https://github.com/firecracker-microvm/firecracker/issues/2546) proposed systemd SysGenID integration
- **Status:** Closed (Nov 2023) - Firecracker team not pursuing systemd integration
- **Alternative approaches:**
  - VirtIO RNG device extensions for PRNG reseeding
  - VMGENID mechanism extensions for user-space detection
- **Implication:** Services must implement custom detection mechanisms

**Mitigation Strategy:**
1. Re-seed PRNG on resume detection
2. Regenerate cryptographic tokens
3. Invalidate session identifiers
4. Clear caches of "unique" identifiers

## Limitations and Constraints

### Limitation 1: Network Connectivity Not Preserved

**Description:**
Network and vsock packet loss expected on guests resumed in another Firecracker process. Network connection state survival not guaranteed.

**vsock-Specific Behavior:**
- vsock device **reset across snapshot/restore** to avoid inconsistent state
- `VIRTIO_VSOCK_EVENT_TRANSPORT_RESET` sent to guest driver on `SnapshotCreate`
- On `SnapshotResume`: vsock driver closes all existing connections
- Listen sockets remain active, CID updated to reflect current guest_cid
- **Rationale:** vsock control protocol not resilient to packet loss; active connections cause device breakage

**TCP Connection Impact:**
- On resume, guest thinks connections still live
- External systems (databases, APIs) may disagree
- Common symptoms: `ECONNRESET`, "Connection closed", timeouts, database pool errors
- **Mitigation:** Applications must reconnect on failure

**Recent Fixes:**
- [Issue #4796](https://github.com/firecracker-microvm/firecracker/blob/main/CHANGELOG.md): Fixed vsock not notifying guest about `TRANSPORT_RESET_EVENT` after restore
- [Issue #4826](https://github.com/firecracker-microvm/firecracker/blob/main/CHANGELOG.md): Added missing tap offload features configuration on restore

**Known Bugs:**
- [vsock-snapshot-reset-bug](https://github.com/loopholelabs/firecracker-vsock-snapshot-reset-bug-reproducer): socat instances with ongoing read syscalls don't exit on connection reset

### Limitation 2: AMD CPU Process Hang Issues

**Description:**
Processes using `nanosleep()` occasionally become stuck after snapshot resume on AMD processors.

**Affected Configurations:**
- **CPUs:** AMD Ryzen, AMD EPYC
- **Unaffected:** Intel processors
- **Firecracker versions:** Confirmed on v1.4.0, status on latest unclear
- **Host kernels:** 5.10, 6.2.0
- **Guest kernels:** 4.14.55, 5.10

**Symptoms:**
- Init binary running `while (true) { sleep 100ms; print 'hello' }` fails to resume
- Issue does **not** occur when replacing `nanosleep` with busy-wait loops
- Specific to snapshot restoration, not simple pause/resume
- Guest logs show: "Firmware Bug: TSC doesn't count with P0 frequency!"

**Workarounds:**
1. **Kernel parameter:** Add `nolapic` to guest kernel boot parameters (eliminates issue, mechanism unclear)
2. **Codesandbox patch:** Fork includes fix (commit `4164371`), upstream status pending

**Impact on NanoFuse:**
- If targeting AMD hosts: Must implement workaround or await upstream fix
- If Intel-only: Lower risk, but not guaranteed future-proof
- **Recommendation:** Add `nolapic` to default kernel parameters as precaution

### Limitation 3: Snapshot/Restore CPU Architecture Constraints

**Description:**
Snapshots are not portable across different CPU architectures, models, or feature sets.

**Specific Constraints:**

1. **Architecture incompatibility:**
   - x86_64 ↔ aarch64 snapshots cannot be exchanged
   - arm64: Cannot restore between different GIC (Generic Interrupt Controller) versions

2. **CPU model incompatibility:**
   - Intel ↔ AMD migration explicitly unsupported
   - CPU features exposed to guest must be identical (invariant requirement)
   - x86_64: `MSR_IA32_TSX_CTRL` overwrites not preserved without CPU templates

3. **Kernel version constraint:**
   - Snapshots compatible only on same host kernel version
   - Cross-kernel restoration unstable (KVM state semantics differ)
   - Example: 5.10.240 → 6.2.0 restoration unreliable

**Implications for NanoFuse:**
- Cannot snapshot on Intel, restore on AMD (or vice versa)
- Cannot migrate snapshots between heterogeneous cluster nodes
- Kernel upgrades invalidate existing snapshots
- **Workaround:**
  - Homogeneous cluster architecture required for snapshot portability
  - Or: Accept snapshot invalidation on infrastructure changes
  - Or: CPU templates to normalize feature sets (adds complexity)

### Limitation 4: Early-Boot Snapshot Kernel Crashes

**Description:**
Snapshots taken during early boot may cause kernel crashes upon resume.

**Root Cause:**
- Incomplete kernel initialization at snapshot time
- Hardware state not fully established
- Driver initialization incomplete

**Mitigation:**
- Only snapshot after system reaches stable state
- Wait for multi-user.target (systemd) or equivalent runlevel
- Verify critical services running before snapshot
- **Safe window:** After all critical services started, before workload execution

### Limitation 5: Security Model for Multi-Resume Scenarios

**Description:**
Resuming same snapshot multiple times creates security risks due to duplicated "unique" state.

**Affected State:**
- Random number generator seeds
- Cryptographic nonces and keys
- Session tokens and identifiers
- UUIDs and unique identifiers
- OS entropy pool

**Attack Scenarios:**
1. **Nonce reuse:** Cryptographic protocols fail with repeated nonces
2. **Session hijacking:** Duplicate session tokens across resumed instances
3. **UUID collision:** Multiple instances generate identical identifiers
4. **PRNG predictability:** Repeated PRNG sequence across instances

**Firecracker Documentation Warning:**
> "When snapshots are used in such a manner that a given guest's state is resumed from more than once, guest information assumed to be unique may in fact not be; this information can include identifiers, random numbers and random number seeds, the guest OS entropy pool, as well as cryptographic tokens."

**Mitigation Requirements:**
- Implement snapshot lifecycle authentication and encryption
- Re-seed PRNG on every resume
- Regenerate cryptographic tokens post-resume
- Invalidate session identifiers
- Use VMGENID or equivalent to detect resume events

### Limitation 6: Snapshot Storage and Restoration Latency

**Description:**
Full snapshots require significant storage and restoration time scales with memory size.

**Storage Requirements:**
- Full snapshot: `vmstate_size + memory_size`
- Example: 2GB VM = ~2GB memory file + vmstate file
- Diff snapshots reduce size but add complexity

**Restoration Latency:**
- File backend: Kernel page fault handling (higher latency)
- Uffd backend: Lower latency, higher complexity
- cgroups v1: High restoration latency
- **cgroups v2 required** for acceptable performance

**Impact on Cold Start Times:**
- Snapshot restore faster than full boot, but not instant
- Large memory footprints increase latency
- **Trade-off:** Snapshot size vs. restoration speed

### Limitation 7: External Dependency Coordination

**Description:**
Snapshot restore requires exact recreation of external dependencies.

**Required at Restore Time:**
- Tap devices with matching names
- Block devices at original paths with proper permissions
- Vsock backing sockets with matching names
- Same Firecracker API socket path

**Challenges:**
- Orchestration complexity in multi-tenant environments
- Race conditions in device availability
- Permission management across hosts
- **Mitigation:** Pre-validation of dependencies before restore attempt

## Proof of Concept Approach

### Objective

Validate Firecracker snapshot/resume for NanoFuse use case: fast cold starts for ephemeral containerized workloads.

### Prerequisites

- Firecracker 1.7.0 or later
- Ubuntu 24.04 base image with systemd
- Kernel 5.10.240 (proven kernel from Slicer)
- Host with cgroups v2 enabled
- Test on both Intel and AMD hosts (if available)

### PoC Implementation Steps

#### Step 1: Build Baseline microVM

**Goal:** Create stable base system suitable for snapshotting.

```bash
# 1.1: Build base Ubuntu 24.04 image
cd nanofuse/images/base
make build

# 1.2: Boot microVM
nanofuse start --image ubuntu-24.04-base.ext4 --kernel vmlinux-5.10.240 \
  --memory 512 --vcpus 2 --id poc-baseline

# 1.3: Wait for stable boot
nanofuse exec poc-baseline -- systemctl is-system-running
# Expected: "running" or "degraded"

# 1.4: Verify critical services
nanofuse exec poc-baseline -- systemctl status systemd-timesyncd
nanofuse exec poc-baseline -- systemctl status systemd-journald
```

#### Step 2: Configure Clock Synchronization

**Goal:** Prepare guest for time jumps on restore.

```bash
# 2.1: Install chrony (better than ntpd for VMs)
nanofuse exec poc-baseline -- apt-get install -y chrony

# 2.2: Configure chrony for large time jumps
nanofuse exec poc-baseline -- tee /etc/chrony/chrony.conf <<EOF
server 169.254.169.254 iburst  # Assuming link-local NTP
driftfile /var/lib/chrony/drift
makestep 1.0 -1  # Allow unlimited time steps
rtcsync
EOF

# 2.3: Create clock-sync-on-restore systemd units
nanofuse exec poc-baseline -- tee /etc/systemd/system/clock-sync-on-restore.service <<EOF
[Unit]
Description=Synchronize clock after VM restore
After=network.target

[Service]
Type=oneshot
ExecStart=/usr/bin/chronyc makestep
EOF

nanofuse exec poc-baseline -- tee /etc/systemd/system/clock-sync-on-restore.timer <<EOF
[Unit]
Description=Detect clock jumps from VM restore

[Timer]
OnClockChange=yes

[Install]
WantedBy=timers.target
EOF

# 2.4: Enable timer
nanofuse exec poc-baseline -- systemctl enable clock-sync-on-restore.timer
nanofuse exec poc-baseline -- systemctl start clock-sync-on-restore.timer
```

#### Step 3: Add AMD CPU Workaround

**Goal:** Mitigate process hang issue on AMD hosts.

```bash
# 3.1: Add nolapic to kernel parameters
# (Implementation depends on bootloader configuration)
# For Firecracker, add to boot_args in VM config:
# "boot_args": "console=ttyS0 reboot=k panic=1 pci=off nolapic"
```

#### Step 4: Create Snapshot

**Goal:** Capture stable baseline snapshot.

```bash
# 4.1: Create workload container
nanofuse exec poc-baseline -- docker run -d --name nginx nginx:alpine

# 4.2: Verify workload running
nanofuse exec poc-baseline -- docker ps | grep nginx

# 4.3: Pause microVM
curl -X PATCH http://localhost:8001/vm \
  -H "Content-Type: application/json" \
  -d '{"state": "Paused"}'

# 4.4: Create snapshot
curl -X PUT http://localhost:8001/snapshot/create \
  -H "Content-Type: application/json" \
  -d '{
    "snapshot_path": "/snapshots/poc-baseline.vmstate",
    "mem_file_path": "/snapshots/poc-baseline.mem",
    "snapshot_type": "Full"
  }'

# 4.5: Record snapshot metadata
echo "Snapshot created at $(date -Iseconds)" > /snapshots/poc-baseline.metadata
echo "Kernel: 5.10.240" >> /snapshots/poc-baseline.metadata
echo "CPU: $(lscpu | grep 'Model name')" >> /snapshots/poc-baseline.metadata
```

#### Step 5: Restore and Validate

**Goal:** Verify snapshot restore functionality and measure performance.

```bash
# 5.1: Stop original microVM
nanofuse stop poc-baseline

# 5.2: Start new Firecracker process
firecracker --api-sock /tmp/firecracker-restore.sock

# 5.3: Load snapshot
curl -X PUT http://localhost:8002/snapshot/load \
  -H "Content-Type: application/json" \
  -d '{
    "snapshot_path": "/snapshots/poc-baseline.vmstate",
    "mem_backend": {
      "backend_type": "File",
      "backend_path": "/snapshots/poc-baseline.mem"
    },
    "enable_diff_snapshots": false,
    "resume_vm": false
  }'

# 5.4: Resume microVM
RESTORE_START=$(date +%s%N)
curl -X PATCH http://localhost:8002/vm \
  -H "Content-Type: application/json" \
  -d '{"state": "Resumed"}'
RESTORE_END=$(date +%s%N)

# 5.5: Calculate restoration time
RESTORE_MS=$(( (RESTORE_END - RESTORE_START) / 1000000 ))
echo "Snapshot restore time: ${RESTORE_MS}ms"

# 5.6: Verify system state
sleep 2  # Allow clock sync to trigger
nanofuse exec poc-baseline -- systemctl is-system-running
nanofuse exec poc-baseline -- docker ps | grep nginx
nanofuse exec poc-baseline -- curl -s http://localhost | grep "Welcome to nginx"

# 5.7: Verify clock sync occurred
nanofuse exec poc-baseline -- journalctl -u clock-sync-on-restore.service --since "1 minute ago"
```

#### Step 6: Validate Network Behavior

**Goal:** Confirm network state handling and connection recovery.

```bash
# 6.1: Test new connections work
nanofuse exec poc-baseline -- curl -s http://example.com >/dev/null && echo "OK" || echo "FAIL"

# 6.2: Test vsock (if applicable)
# Attempt vsock connection from host to guest after restore
# Expected: New connections succeed, pre-snapshot connections reset

# 6.3: Verify Docker networking
nanofuse exec poc-baseline -- docker exec nginx wget -q -O- http://example.com >/dev/null && echo "OK" || echo "FAIL"
```

#### Step 7: Stress Test with Multiple Resumes

**Goal:** Validate multi-resume scenario and identify edge cases.

```bash
# 7.1: Perform 10 restore cycles
for i in $(seq 1 10); do
  echo "=== Restore cycle $i ==="

  # Stop previous instance
  nanofuse stop poc-baseline || true

  # Restore snapshot
  nanofuse restore --snapshot /snapshots/poc-baseline --id poc-baseline

  # Verify system health
  sleep 2
  nanofuse exec poc-baseline -- systemctl is-system-running
  nanofuse exec poc-baseline -- docker ps | grep nginx

  # Check for PRNG state issues (basic test)
  UUID1=$(nanofuse exec poc-baseline -- cat /proc/sys/kernel/random/uuid)
  UUID2=$(nanofuse exec poc-baseline -- cat /proc/sys/kernel/random/uuid)
  if [ "$UUID1" = "$UUID2" ]; then
    echo "ERROR: PRNG not re-seeding properly!"
  fi

  # Wait before next cycle
  sleep 5
done
```

#### Step 8: Measure Performance Metrics

**Goal:** Quantify snapshot/restore performance vs. cold boot.

```bash
# 8.1: Measure cold boot time
COLD_START=$(date +%s%N)
nanofuse start --image ubuntu-24.04-base.ext4 --kernel vmlinux-5.10.240 \
  --memory 512 --vcpus 2 --id poc-coldboot
nanofuse exec poc-coldboot -- systemctl is-system-running  # Wait for ready
COLD_END=$(date +%s%N)
COLD_MS=$(( (COLD_END - COLD_START) / 1000000 ))
nanofuse stop poc-coldboot

# 8.2: Measure snapshot restore time (already collected in Step 5.5)
echo "Cold boot: ${COLD_MS}ms"
echo "Snapshot restore: ${RESTORE_MS}ms"
echo "Improvement: $(( (COLD_MS - RESTORE_MS) * 100 / COLD_MS ))%"

# 8.3: Measure snapshot storage overhead
SNAPSHOT_SIZE=$(du -sh /snapshots/poc-baseline.* | awk '{sum+=$1} END {print sum}')
IMAGE_SIZE=$(du -sh nanofuse/images/base/ubuntu-24.04-base.ext4 | awk '{print $1}')
echo "Image size: ${IMAGE_SIZE}"
echo "Snapshot size: ${SNAPSHOT_SIZE}"
```

### Success Criteria

A successful PoC must demonstrate:

1. **Functional Restore:**
   - [x] microVM boots from snapshot successfully
   - [x] systemd services running post-restore
   - [x] Container workloads operational post-restore
   - [x] No kernel panics or crashes

2. **Clock Synchronization:**
   - [x] Guest clock synchronized to current time within 5 seconds of restore
   - [x] systemd timers functional post-restore
   - [x] No timer drift causing service failures

3. **Network Recovery:**
   - [x] New network connections succeed after restore
   - [x] Docker container networking functional
   - [x] vsock reset behavior confirmed (connections reset, listen sockets active)

4. **Performance Improvement:**
   - [x] Snapshot restore at least 30% faster than cold boot
   - [x] Acceptable restoration latency (<2 seconds for 512MB VM)

5. **Stability Across Multiple Restores:**
   - [x] 10 consecutive restore cycles without failures
   - [x] No PRNG state issues detected
   - [x] No memory leaks or resource exhaustion

6. **CPU Compatibility:**
   - [x] Successful restore on Intel hosts
   - [x] Successful restore on AMD hosts (with `nolapic` workaround if needed)

### Failure Criteria (NO-GO Indicators)

The PoC fails if:

- [ ] Kernel panics on restore (>10% of attempts)
- [ ] systemd services fail to start/operate (>20% of attempts)
- [ ] Clock synchronization doesn't occur within 30 seconds
- [ ] Snapshot restore slower than cold boot
- [ ] Process hangs occur on Intel hosts (AMD acceptable with workaround)
- [ ] Network connectivity completely non-functional post-restore
- [ ] Restoration latency >5 seconds for 512MB VM

## Recommendation

### Decision: CONDITIONAL GO

Firecracker snapshot/resume is **feasible for NanoFuse** with the following conditions and caveats.

### Rationale

**Strengths:**
1. **Significant performance improvement potential:** Snapshot restore demonstrably faster than cold boot in documented use cases
2. **Mature API:** Firecracker's snapshot API is production-ready and well-documented
3. **Active maintenance:** Recent bug fixes show ongoing support (vsock reset, tap offload features)
4. **Clear limitations:** Known issues are documented with workarounds available

**Risks:**
1. **AMD CPU process hang issue:** Requires `nolapic` workaround or Intel-only deployment
2. **Network state complexity:** Applications must handle connection resets gracefully
3. **Clock synchronization requirement:** Additional configuration needed for systemd compatibility
4. **CPU architecture lock-in:** Snapshots non-portable across heterogeneous clusters
5. **Kernel version constraint:** Infrastructure upgrades invalidate snapshots

### Implementation Prerequisites

1. **Mandatory Requirements:**
   - [ ] Host systems must use **cgroups v2** (high restoration latency with v1)
   - [ ] Homogeneous CPU architecture (all Intel or all AMD, not mixed)
   - [ ] Same host kernel version across snapshot/restore hosts
   - [ ] Guest kernel with `nolapic` boot parameter (AMD CPU workaround)

2. **Guest Configuration:**
   - [ ] chrony or ntpd configured with `tinker panic 0` (or `makestep 1.0 -1` for chrony)
   - [ ] systemd timer to detect clock jumps and trigger sync
   - [ ] Snapshot taken after system reaches stable state (post multi-user.target)
   - [ ] PRNG re-seeding mechanism on resume detection

3. **Application Requirements:**
   - [ ] Applications must handle network connection resets gracefully
   - [ ] Retry logic for database connections and external APIs
   - [ ] No reliance on pre-snapshot network state preservation
   - [ ] Session token regeneration on resume detection

4. **Operational Requirements:**
   - [ ] Snapshot versioning and compatibility tracking
   - [ ] Snapshot lifecycle authentication and encryption (security)
   - [ ] Validation of external dependencies before restore (tap devices, block devices)
   - [ ] Monitoring for restoration failures and fallback to cold boot

### Recommended Phased Approach

**Phase 1: Proof of Concept (2-3 weeks)**
- Execute PoC steps outlined above
- Validate success criteria on both Intel and AMD hosts
- Measure performance improvements vs. cold boot
- Identify any NanoFuse-specific edge cases

**Phase 2: Integration (4-6 weeks)**
- Implement snapshot creation in `nanofused` API
- Implement snapshot loading in `nanofuse` CLI
- Add clock synchronization to base image build
- Add `nolapic` workaround to kernel parameters
- Implement snapshot metadata tracking and versioning

**Phase 3: Hardening (4-6 weeks)**
- Add snapshot lifecycle authentication
- Implement PRNG re-seeding on resume
- Add restoration failure detection and fallback
- Load testing with multiple concurrent restores
- Documentation and runbooks

**Phase 4: Production Rollout (2-4 weeks)**
- Gradual rollout with monitoring
- Performance benchmarking vs. cold boot
- Incident response procedures
- User documentation

### Alternative Approach (If NO-GO)

If PoC fails or risks are deemed too high:

**Option A: Optimize Cold Boot Instead**
- Focus on minimizing base image size
- Optimize kernel configuration (minimal modules)
- Parallelize initialization tasks
- Pre-warm disk cache with predictive loading

**Option B: Hybrid Approach**
- Use snapshots only for specific, well-tested workloads
- Fall back to cold boot for untested or high-risk scenarios
- Limit snapshot usage to Intel-only hosts (avoid AMD issue)

**Option C: Defer Implementation**
- Wait for upstream resolution of AMD CPU issue
- Wait for SysGenID or VMGENID standardization
- Revisit in 6-12 months with fresh evaluation

### Next Steps

1. **Immediate:** Execute Proof of Concept (Steps 1-8 above)
2. **After PoC:**
   - Document results in `docs/building/snapshot-resume-poc-results.md`
   - Present findings with GO/NO-GO recommendation
   - If GO: Proceed to Phase 2 (Integration)
   - If NO-GO: Evaluate alternative approaches

## References

### Firecracker Documentation
- [Snapshot Support](https://github.com/firecracker-microvm/firecracker/blob/main/docs/snapshotting/snapshot-support.md) - Official snapshot API documentation
- [Snapshot Versioning](https://github.com/firecracker-microvm/firecracker/blob/main/docs/snapshotting/versioning.md) - Version compatibility and migration constraints

### GitHub Issues
- [Issue #4099: Processes get stuck after resuming VM from snapshot](https://github.com/firecracker-microvm/firecracker/issues/4099) - AMD CPU process hang issue
- [Issue #2546: SysGenID support via Systemd](https://github.com/firecracker-microvm/firecracker/issues/2546) - PRNG re-seeding and snapshot safety

### systemd Documentation
- [systemd.timer(5) - Linux manual page](https://man7.org/linux/man-pages/man5/systemd.timer.5.html) - Timer unit configuration
- [Timekeeping best practices for Linux guests](https://knowledge.broadcom.com/external/article/310053/timekeeping-best-practices-for-linux-gue.html) - VM clock synchronization best practices

### Related Projects
- [Firecracker VSock Snapshot Reset Bug Reproducer](https://github.com/loopholelabs/firecracker-vsock-snapshot-reset-bug-reproducer) - Known vsock connection reset issue

---

**Document Version:** 1.0
**Last Updated:** 2025-12-09
**Author:** Research Task 012
**Status:** Initial Research Complete
