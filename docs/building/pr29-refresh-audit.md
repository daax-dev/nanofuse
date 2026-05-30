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
- Replace the previous snapshot stubs with real Firecracker snapshot create/load API calls on the existing manager.
- Keep an audit trail for why the scope was narrowed.

## Implemented Snapshot Slice

The refreshed branch implements the Firecracker snapshot API calls needed before any future warm-pool work can be valid:

- `Manager.CreateSnapshot` sends `PUT /snapshot/create` over the VM Firecracker Unix socket.
- Snapshot creation uses `snapshot_type: "Full"`, `snapshot_path`, and `mem_file_path`.
- `Manager.LoadSnapshot` sends `PUT /snapshot/load` over the VM Firecracker Unix socket.
- Snapshot loading uses `snapshot_path`, `mem_backend.backend_type: "File"`, `mem_backend.backend_path`, `enable_diff_snapshots: false`, and defaults `resume_vm` to `true`.
- `Manager.LoadSnapshotWithResume` exposes the same load path with explicit `resume_vm` control.
- Unit tests use an Unix-socket `httptest` server and validate request path, method, and JSON body.

This targets the repo-pinned Firecracker v1.7.0 API. The Vagrant setup pins Firecracker 1.7.0 in `dev/vagrant/setup.sh`, and the v1.7.0 swagger is the primary contract for the exact request fields. Firecracker upgrades must include an API-field review before changing these request shapes.

This is not a VM pool implementation. It does not make a cold-start latency claim.

Primary reference: Firecracker v1.7.0 swagger, `https://raw.githubusercontent.com/firecracker-microvm/firecracker/v1.7.0/src/firecracker/swagger/firecracker.yaml`.

Supporting reference: Firecracker snapshot support documentation, `https://github.com/firecracker-microvm/firecracker/blob/main/docs/snapshotting/snapshot-support.md`.

Future implementation should start from a new spec and should land as smaller independently validated slices:

- Snapshot restore API integration before any VM pool performance target.
- VM lifecycle-owned identity registration/revocation before SVID rotation.
- Disk isolation strategy selection and validation before overlayfs or CoW claims.

## PR Summary Text

Problem:

PR #29 was closed and stale. Its implementation overclaimed production behavior for VM pooling, SVID rotation, and CoW layers. The branch also predated current-main module identity and PR #46 API/egress/capabilities changes.

Approach:

Merged current main non-destructively, resolved conflicts toward current main, removed the unverified PR29 feature files, preserved `github.com/daax-dev/nanofuse`, implemented the narrow Firecracker snapshot create/load API client slice, and documented the audit findings. Copilot review blockers against the old VMPool/SVID/CoW files are closed by removal of the affected stale files rather than by carrying partial fixes.

Validation:

See the final branch handoff for exact commands and results.
