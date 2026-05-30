# PR29 Refresh Audit

Date: 2026-05-30

Branch: `fix/issues-11-12-13-v2`

Base target: `origin/main` at `fb56363` after PR #46

## Outcome

PR #29 was closed and not merged. Its original branch claimed three production features:

- VM pool with snapshot-based cold start targeting P50 under 100 ms.
- Per-VM SPIFFE SVIDs with one-hour TTL and rotation.
- Copy-on-write filesystem layers with fast snapshots.

Those claims are not supported by the branch implementation. The refresh removes the unverified implementation files and preserves current-main behavior from PR #46, including API capability reporting, egress policy fields and cleanup, writable per-VM rootfs materialization, current Flowspec guidance, and the canonical `github.com/daax-dev/nanofuse` module identity.

No Firecracker hardware validation was run for this refresh. No latency claim is made.

## Removed Implementation

The following stale files are removed from the refreshed branch:

- `internal/firecracker/pool.go`
- `internal/firecracker/pool_test.go`
- `internal/layer/cow.go`
- `internal/layer/cow_test.go`
- `internal/spire/svid.go`
- `internal/spire/svid_test.go`

Removal is intentional. The files were not current-main-compatible production code.

## Blocking Findings Addressed

The stale `VMPool` implementation was removed instead of patched because:

- `warmOneContext` claimed snapshot restore while calling `Manager.Start`.
- No Firecracker `LoadSnapshot` flow was wired into the manager.
- The 100 ms cold-start target had no hardware validation evidence.
- `MaxSize` was bypassed by on-demand warm paths.
- `Release` could underflow `inFlight` and did not prove ownership of the VM being released.
- `Release` returned used VMs to the ready pool without reset, snapshot reload, disk cleanup, network cleanup, or identity reinitialization.

The stale SVID store was removed instead of patched because:

- It introduced unused auth/SVID configuration outside the current-main daemon auth model.
- It was not wired into VM lifecycle create/delete/start/stop paths.
- `renewSVID` used a background context instead of a caller-controlled lifecycle context.
- Rotation used a hard-coded buffer that could drift from config.
- Renewal could recreate SPIRE state after revoke because candidate rotation was not coordinated with revocation.
- Tests validated only disabled or manually inserted in-memory state, not guest SVID issuance or rotation.

The stale CoW layer was removed instead of patched because:

- It was not integrated into VM disk creation or image/layer storage.
- It required privileged overlayfs mounts that unit tests skipped.
- Tests did not validate actual CoW isolation, overlay semantics, or snapshot timing.

The old CI and lint changes were not retained because:

- Current main already owns the golangci-lint v2 configuration.
- Broad gosec suppressions from the old branch had inaccurate trust-boundary rationales.
- `config.Load` path handling remains current-main behavior; no config-read gosec suppression is added.
- The old `govulncheck continue-on-error` stdlib-vulnerability workaround is not carried forward unless a current-main gate proves it is still needed.

## Current Scope

The refreshed branch is intentionally narrow:

- Merge current main into the old PR29 branch.
- Resolve conflicts toward current main for guidance, tooling, dependencies, API, SPIRE service, and config.
- Remove unverified VM pool, SVID rotation, and CoW layer code.
- Replace the previous snapshot stubs with real Firecracker snapshot create and VM pause/resume API calls on the existing manager.
- Keep an audit trail for why the scope was narrowed.

## Implemented Snapshot Slice

The refreshed branch implements the Firecracker snapshot API calls needed before any future warm-pool work can be valid:

- `Manager.CreateSnapshot` sends `PUT /snapshot/create` over the VM Firecracker Unix socket.
- Snapshot creation uses `snapshot_type: "Full"`, `snapshot_path`, and `mem_file_path`.
- `POST /vms/{id}/snapshots` now rejects non-paused VMs before calling Firecracker.
- `Manager.Pause` and `Manager.Resume` send `PATCH /vm` with `state: "Paused"` or `state: "Resumed"` so the public API can move running VMs into the valid snapshot state.
- Unit tests use an Unix-socket `httptest` server and validate request path, method, and JSON body.

This targets the repo-pinned Firecracker v1.7.0 API. The Vagrant setup pins Firecracker 1.7.0 in `dev/vagrant/setup.sh`, and the v1.7.0 swagger is the primary contract for the exact request fields. Firecracker upgrades must include an API-field review before changing these request shapes.

This is not a VM pool implementation. It does not make a cold-start latency claim.

Primary reference: Firecracker v1.7.0 swagger, `https://raw.githubusercontent.com/firecracker-microvm/firecracker/v1.7.0/src/firecracker/swagger/firecracker.yaml`.

Supporting reference: Firecracker snapshot support documentation, `https://github.com/firecracker-microvm/firecracker/blob/main/docs/snapshotting/snapshot-support.md`.

## Replacement PR Copilot Audit

Closed PR #29 still retained Copilot inline comments, but the comments inspected before this replacement pass were attached to stale commit `a74d20904c51175b6a6fdde52c34bc7193ea836d` and reported as outdated review threads. No current-head Copilot comments were found against branch head `a2c1c5910d7505ffbd4c5283c4fddaf24ead070f`.

The stale comments target old VM pool, SVID rotation, and CoW layer files that are no longer in the effective branch diff. Those implementations remain removed rather than patched in place.

The current-head gap found during this pass was the public snapshot API accepting running VMs even though Firecracker snapshot creation requires a paused microVM. The replacement branch now rejects non-paused snapshot requests and wires the existing pause/resume routes to Firecracker `PATCH /vm` state transitions.

Fresh Copilot review on replacement PR #47 correctly identified that a low-level `PUT /snapshot/load` helper using an already-started VM runtime socket cannot restore snapshots in production because Firecracker requires snapshot loading before microVM configuration. That helper has been removed. Snapshot restore requires a future restore-specific launch path that starts Firecracker with an API socket, loads the snapshot before normal configuration, and then reconciles daemon lifecycle state.

Future implementation should start from a new spec and should land as smaller independently validated slices:

- Snapshot restore-specific launch path before any VM pool performance target.
- VM lifecycle-owned identity registration/revocation before SVID rotation.
- Disk isolation strategy selection and validation before overlayfs or CoW claims.

## PR Summary Text

Problem:

PR #29 was closed and stale. Its implementation overclaimed production behavior for VM pooling, SVID rotation, and CoW layers. The branch also predated current-main module identity and PR #46 API/egress/capabilities changes.

Approach:

Merged current main non-destructively, resolved conflicts toward current main, removed the unverified PR29 feature files, preserved `github.com/daax-dev/nanofuse`, implemented the narrow Firecracker snapshot create and VM pause/resume API client slice, and documented the audit findings. Copilot review blockers against the old VMPool/SVID/CoW files are closed by removal of the affected stale files rather than by carrying partial fixes.

Replacement PR hardening also makes the public snapshot endpoint require a paused VM, matching the Firecracker snapshot lifecycle requirement instead of accepting running VMs and failing later inside the VMM. The same hardening replaces the pause/resume handler stubs with Firecracker `PATCH /vm` calls.

Validation:

See the final branch handoff for exact commands and results.
