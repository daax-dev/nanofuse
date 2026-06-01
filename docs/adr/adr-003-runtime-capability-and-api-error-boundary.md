# ADR-003: Runtime Capability and API Error Boundary

## Status

Accepted

## Context

Nanofuse has a stable daemon API and multiple runtime backends with different host requirements and feature coverage. Linux/KVM hosts use Firecracker. macOS hosts use Apple `container`, which launches Linux guests through Virtualization.framework. Clients on Windows and macOS can manage VMs through the same API and tray UI.

The runtime abstraction introduced a new failure class: a request can be valid for the Nanofuse API but unsupported by the selected runtime backend. Before this ADR, some unsupported runtime paths were translated as `INTERNAL_ERROR`/500, which made clients treat backend feature gaps as daemon failures.

## Decision

Nanofuse will distinguish unsupported runtime capabilities from daemon failures:

- Runtime backends return `vmm.ErrUnsupportedOperation` for capabilities they do not expose.
- API handlers translate that sentinel to HTTP `501 Not Implemented` with error code `UNSUPPORTED_OPERATION`.
- Guest command non-zero exits remain successful API responses with `exit_code`.
- Runtime command timeouts are transport failures and return HTTP `504 Gateway Timeout`.
- Runtime-managed networking rejects unsupported network posture during VM creation.

For the Apple-container backend, `network.mode=none` is rejected at VM creation because Apple `container` currently provides runtime-managed NAT networking and Nanofuse cannot truthfully create a no-network VM through that backend.

## Consequences

### Positive

- Clients can distinguish unsupported backend capabilities from transient daemon errors.
- The tray and CLI can show actionable unsupported-runtime errors.
- VM creation fails before launch when the requested network posture cannot be honored.
- Timeout behavior is deterministic for API callers.

### Negative

- Cross-runtime behavior is not feature-identical; clients must handle `UNSUPPORTED_OPERATION`.
- Apple-container VMs cannot currently satisfy a no-network posture.

### Neutral

- Firecracker remains the runtime for pause/resume snapshot semantics.
- Future runtimes can opt into pause, resume, snapshot, and exec features by implementing them instead of changing the API contract.

## Alternatives Considered

### Treat Unsupported Runtime Features as Internal Errors

- **Pros:** No new runtime sentinel needed.
- **Cons:** Clients cannot distinguish backend capability gaps from daemon faults.
- **Why rejected:** It breaks the API contract needed by CLI, tray, and remote clients.

### Hide Unsupported Operations in Clients

- **Pros:** Keeps server behavior unchanged.
- **Cons:** Every client must duplicate runtime-specific capability logic and remote clients can still hit bad states.
- **Why rejected:** The daemon owns runtime selection and must enforce the contract.

### Allow Apple-Container `network.mode=none` and Fail at Start

- **Pros:** Less create-time validation.
- **Cons:** Creates VM records that are guaranteed to fail on start and misrepresent the requested security posture.
- **Why rejected:** Invalid runtime/network combinations must fail before persistent VM creation.

## References

- `.flowspec/features/codex-goal/design.md`
- `.flowspec/features/codex-goal/spec.md`
- `internal/vmm/errors.go`
- `internal/api/runtime_errors.go`
