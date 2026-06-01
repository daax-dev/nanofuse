# Design Artifact: Runtime Capability and API Error Boundary

**Date**: 2026-05-31
**Status**: Accepted
**Related ADR**: `docs/adr/adr-003-runtime-capability-and-api-error-boundary.md`

## Context

Nanofuse now supports more than one local runtime backend:

- Linux/KVM hosts use Firecracker.
- macOS hosts use Apple `container`, which launches Linux guests through Virtualization.framework.
- Windows hosts are API/tray clients unless they target a reachable daemon.

The daemon API is the stable product boundary. Runtime backends differ in capabilities: Apple `container` can start, stop, kill, delete, list images, pull images, expose host port forwards, and execute commands in a running runtime container. It does not expose Firecracker-style pause/resume snapshots in the currently targeted command surface.

## Decision

Runtime-specific missing features are represented as an explicit unsupported-operation condition, then translated at the API boundary to:

- HTTP status: `501 Not Implemented`
- API error code: `UNSUPPORTED_OPERATION`

Invalid operator configuration remains a create-time validation error:

- HTTP status: `400 Bad Request`
- API error code: `INVALID_CONFIG`

For Apple-container VMs, `network.mode=none` is rejected during VM creation because the backend currently uses runtime-managed NAT networking and cannot honor a no-network request.

## API Contract

| Condition | HTTP status | Error code | Notes |
| --- | ---: | --- | --- |
| Runtime does not implement VM exec | 501 | `UNSUPPORTED_OPERATION` | Returned when the selected runtime lacks command execution support. |
| Runtime exec exceeds API timeout | 504 | `SERVICE_UNAVAILABLE` | Timeout is a transport/runtime failure, not a guest command exit code. |
| Guest command exits non-zero | 200 | n/a | The exit code is returned in `exit_code`; this is a completed guest command. |
| Runtime does not implement VM pause/resume | 501 | `UNSUPPORTED_OPERATION` | Apple-container returns this for pause and resume. |
| Runtime does not implement snapshots | 501 | `UNSUPPORTED_OPERATION` | Apple-container returns this for snapshots. |
| Runtime-managed networking receives `network.mode=none` | 400 | `INVALID_CONFIG` | The VM is not created because the requested network posture cannot be honored. |

## Validation

Regression tests cover:

- Context deadline errors are reported before `exec.ExitError` so timed-out execs reach the API as `504`.
- Apple-container `network.mode=none` fails during VM networking setup validation.
- Unsupported pause, resume, and snapshot operations return `501 UNSUPPORTED_OPERATION`.

## Open Follow-Up

Snapshot semantics must be revisited if a non-Firecracker runtime later exposes compatible pause/snapshot/resume primitives.
