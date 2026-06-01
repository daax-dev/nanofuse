# ADR-004: VM Inventory and Exposure Metadata

## Status

Proposed

## Context

Nanofuse clients need a Docker-ps-like view of running microVMs, plus enough metadata to manage security posture without shelling into the guest or runtime host. The immediate tray and CLI surfaces already show VM state and configured port forwards. Upcoming management features need to track these per VM:

- Mounts and writable/persistent storage.
- Host ingress exposure and guest port targets.
- Egress policy and proxy mode.
- Secret or identity brokers used by the VM.

The tray app, CLI, and future SDKs must get this information through `nanofused`; desktop clients must not inspect runtime-private state directly.

## Decision

Nanofuse will model VM operational inventory as daemon-owned API metadata:

- VM list/detail responses remain the source of truth for state, image, runtime handle, and configured host-to-VM port forwards.
- Future mount, ingress, egress, and secret-use fields will be added to VM config/status objects rather than embedded only in tray-local state.
- The tray and CLI will render these fields from API responses and will not call Firecracker, Apple `container`, Docker, SSH, or host networking tools directly for authoritative state.
- `host_port: 0` on VM create means "allocate a concrete host port on the daemon host" and the assigned port is returned in VM metadata.

## Consequences

### Positive

- Mac, Windows, and remote clients see the same inventory.
- Runtime-specific implementations can differ while preserving one API contract.
- Security posture can be reviewed from one daemon-owned object graph.

### Negative

- VM schemas need explicit evolution as mounts, ingress/egress policy, and secret brokers mature.
- The daemon must normalize runtime-specific details into stable API fields.

### Neutral

- Host reachability checks such as `nanofuse vm ports` can still probe localhost TCP from the client, but they are diagnostics, not the source of truth.

## Alternatives Considered

### Tray-Local Runtime Inspection

- **Pros:** Faster to prototype for a local desktop.
- **Cons:** Breaks Windows/remote clients, duplicates runtime knowledge, and bypasses daemon ownership.
- **Why rejected:** It violates the API/client boundary.

### Free-Form Labels Only

- **Pros:** Avoids schema work.
- **Cons:** Makes ports, mounts, egress, and secret-use non-queryable and hard to validate.
- **Why rejected:** The management UI needs structured state and state-aware controls.

## References

- `api/openapi.yaml`
- `internal/types/vm.go`
- `docs/TRAY_APP.md`
- `docs/OPERATING_LOCAL_MICROVMS.md`
