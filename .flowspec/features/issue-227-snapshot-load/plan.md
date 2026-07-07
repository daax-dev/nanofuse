# Plan: Snapshot-Load / Resume Path (issue #227)

## Primary-source contract (Firecracker v1.15, repo target per scripts/download-fixtures.sh)
`PUT /snapshot/load` (SnapshotLoadParams, verified against
`firecracker/v1.15.0/src/firecracker/swagger/firecracker.yaml`):
- `snapshot_path` (required) — microVM state file.
- Exactly one memory source. Use `mem_backend` (preferred; `mem_file_path` is
  deprecated): `{ "backend_type": "File", "backend_path": <mem file> }`.
  `backend_type` enum is `File | Uffd`; `File` uses kernel page-fault handling.
- `resume_vm: false` — the snapshot loads paused; LoadSnapshot resumes the
  vCPUs with an explicit `PATCH /vm {state:Resumed}` *after* the SPIRE vsock
  proxy is listening, so early guest->host vsock traffic cannot race a
  not-yet-listening proxy.
- Load must happen on a FRESH Firecracker process with no boot source (no
  `--config-file`, no instance-start). Host tap devices referenced by the
  snapshot must already exist and be reachable by the new process.

## Layers to change

### 1. Runtime (internal/firecracker/vm.go)
- Add request structs: `memoryBackend{backend_type,backend_path}` and
  `snapshotLoadRequest{snapshot_path, mem_backend, resume_vm}`.
- Refactor `startFirecrackerProcess` to omit `--config-file` when `configPath`
  is empty (single new call site; existing behavior unchanged for real paths).
- Add `sendSnapshotLoad(socketPath, snapshotPath, memPath) error` — builds the
  request and calls `firecrackerPUT(.../snapshot/load)`. Unit-testable against a
  unix socket, mirroring the existing CreateSnapshot test.
- Add `waitForSocketReady(socketPath, timeout) error` — dials the unix socket
  until it accepts connections or times out.
- Add `Manager.LoadSnapshot(vm, snapshotPath, memPath) error`:
  validate inputs and file existence -> ensure vmDir -> remove stale API socket
  -> start fresh Firecracker (no config) -> wait for socket -> `sendSnapshotLoad`
  -> `setupVMRuntime` + reaper goroutine. On any post-spawn failure, kill and
  reap the process (no zombie) and return an actionable error.
  The host-side SPIRE vsock proxy IS re-wired on resume (started while the VM is
  paused, before the resume PATCH) so a resumed VM keeps host-service access;
  it is tracked for Stop cleanup. Cross-node/in-guest SPIRE agent work remains
  out of scope (that is #17 / AC2 territory).

### 2. Runtime interface + other implementors
- Add `LoadSnapshot(vm, snapshotPath, memPath) error` to `vmm.Manager`.
- `applecontainer.Manager.LoadSnapshot` returns `vmm.ErrUnsupportedOperation`
  (matches existing Pause/Resume/CreateSnapshot idiom).
- Extend the api test stub (`runtimeImageProviderStub`) with a recording
  `LoadSnapshot`.

### 3. API (internal/api/snapshot_handlers.go / vm_handlers.go)
- `handleVMResumeByPath`: parse optional `ResumeVMRequest` body. If a non-empty
  `SnapshotID` is present, delegate to a new `resumeVMFromSnapshot`; otherwise
  keep the existing paused-resume path unchanged.
- `resumeVMFromSnapshot(w, vm, snapshotID)`:
  1. Reject if VM state is running/resuming (409, "stop the VM first").
  2. Load snapshot record; 404 if missing.
  3. 400 if `snapshot.VMID != vm.ID`.
  4. `os.Stat` both backing files; 404 with the offending path if missing.
  5. Acquire VM lock (409 if locked).
  6. `ensureVMTapForResume` — recreate the host tap (idempotent) only when the VM
     uses networking and a tap name is recorded; keeps unit tests hermetic.
  7. Set state Resuming, persist, call `runtimeManager.LoadSnapshot`. On failure
     restore prior state; map `ErrUnsupportedOperation` to 501.
  8. On success set state Running, persist, return the VM.

### 4. CLI / client
- Already wired: `--from-snapshot` -> `ResumeVM(ctx,id,snapshotID)` ->
  `POST /vms/{id}/resume {snapshot_id}`. No change needed.

## Tests (TDD, written first)
- firecracker: request schema (`/snapshot/load`, `mem_backend.backend_type=File`,
  `backend_path`, `resume_vm=false`); input/file validation; `waitForSocketReady`
  success + timeout.
- api: happy path (state Running, LoadSnapshot args), running-VM conflict,
  snapshot not found, wrong owner, missing files, unsupported runtime -> 501.

## Verification boundary
Unit tests cover request schema, validation, state machine, and error mapping.
The full process-spawn + `/snapshot/load` round-trip against real Firecracker
requires KVM and is exercised via e2e (documented; run under a root/KVM sandbox).
