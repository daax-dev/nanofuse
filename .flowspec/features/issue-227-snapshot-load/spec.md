# Spec: Snapshot-Load / Resume Path (issue #227)

## Why
nanofuse can create VM snapshots but cannot restore or resume from them. The
resume operation ignores the snapshot entirely, so a snapshot is a write-only
artifact: it can be produced and stored but never used to bring a VM back to a
running state. This blocks the core value of snapshots (fast restore, and,
combined with snapshot tiering, cross-node portability).

## What (scope of this change)
Provide a local restore path so that a VM can be brought back to a running state
from a previously created snapshot.

### In scope (AC1)
- A resume operation that accepts a snapshot identifier restores the VM from the
  referenced snapshot and leaves it running.
- The operation validates that the snapshot exists, belongs to the target VM,
  and that its backing files are present before attempting restore.
- The operation refuses to restore over a VM that currently has a live runtime,
  with an actionable message.
- Any host-side network resources the restored VM depends on are re-established
  when absent, so the restored VM matches the state captured in the snapshot.
- Errors are explicit and actionable; no silent failures.

### Out of scope (documented follow-up)
- **AC2 — cross-node resume** (restore on host B a snapshot produced on host A).
  This depends on snapshot object-storage tiering (issue #130), which is an
  unmerged pull request (#250). AC2 is intentionally deferred and must not be
  built on #130's code. The local restore path in this change is the prerequisite
  that AC2 will reuse.
- **Snapshot/VM compatibility validation** (deferred follow-up). A Firecracker
  snapshot is authoritative for the guest's hardware (vCPUs, memory, MAC,
  architecture). Two validations are therefore desirable but require persisting
  that metadata on the snapshot record (a storage schema change), so they are
  out of scope here:
  1. **Architecture match** — reject loading a snapshot whose architecture
     differs from the target VM. Not reachable in AC1 (a local snapshot is
     created from, and resumed on, the same VM, so architectures always match);
     becomes relevant only for imported/cross-node snapshots (AC2).
  2. **Config-drift detection** — warn or reject when the VM's current recorded
     config (memory, vCPUs, tap name) was changed after the snapshot was taken,
     since the resumed guest uses the snapshot's config, not the current record.
     Not reachable in the AC1 happy path (create -> start -> pause -> snapshot ->
     stop -> resume performs no config edit in between).

## Success Criteria (measurable)
1. Invoking resume with a valid snapshot identifier for a non-running VM results
   in the VM entering the running state and the underlying runtime being
   instructed to load that snapshot.
2. Invoking resume with a snapshot identifier that does not exist yields a
   not-found result and leaves VM state unchanged.
3. Invoking resume with a snapshot identifier that belongs to a different VM
   yields an invalid-request result and leaves VM state unchanged.
4. Invoking resume with a snapshot whose backing files are missing yields a
   not-found result with a message that identifies the missing artifact, and
   leaves VM state unchanged.
5. Invoking snapshot resume against a VM that is currently running yields a
   conflict result with guidance to stop the VM first, and leaves VM state
   unchanged.
6. When the runtime does not support snapshot restore, the caller receives a
   not-implemented result identifying the unsupported operation.
7. The restore request sent to the runtime carries the snapshot state path, the
   memory backing path, and an instruction to resume execution.

## Assumptions
- The natural local lifecycle is: create -> start -> pause -> snapshot -> stop ->
  resume-from-snapshot. Restore therefore targets a VM that is not currently
  running, and a fresh runtime instance is created to load the snapshot.
- Snapshot metadata already persists the absolute paths to the state and memory
  files.
