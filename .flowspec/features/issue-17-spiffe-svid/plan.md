# Implementation Plan: Per-microVM SPIFFE SVID

**Spec**: `./spec.md` · **Issue**: #17 · **Status**: lifecycle delivered, production attestation deferred

## Architecture Decision: SVID Delivery Topology

The operator selected **Topology B**. The alternatives were evaluated as follows.

| # | Topology | Where the SVID private key lives | Attestation | Blast radius of a host/control-plane compromise | Verdict |
|---|----------|----------------------------------|-------------|--------------------------------------------------|---------|
| A | Host mints the SVID and injects it into the guest filesystem | Key generated host-side, copied into guest | Host asserts guest identity | Host holds every guest's private key; one host compromise forges all identities | Rejected — contradicts per-VM isolation; host becomes a universal key custodian |
| **B** | **SPIRE agent runs inside each guest, node-attested; SVID key generated and kept in-guest; SVID delivered to the guest at a fixed path** | **In-guest only by design (target: key never leaves the VM)** | **Node attestation of the guest** | **By design, a compromised host cannot read a running guest's private key; identities are per-VM** | **Selected** |
| C | Guest calls host-side Workload API for each handshake (no in-guest key) | No persistent key; per-call signing on host | Host asserts identity per call | Host sees all signing operations; strong coupling and a host-side oracle | Rejected — host-side signing oracle; heavier runtime coupling |

**Topology B rationale**: keeping the private key inside the VM boundary is the
security *goal* of this topology — it would extend the hardware isolation the
platform already provides to the identity's most sensitive asset. That
key-stays-in-guest property is a design goal that depends on the **deferred**
production in-guest Workload API Source; it is not something this increment's
code guarantees. The current increment ships a development `LocalCASource` that
mints SVIDs in-process (control-plane side), so the private key is not yet
confined to the guest — the isolation guarantee lands only when the deferred
in-guest source is built. The SVID document is delivered to the guest at
`/var/run/secrets/spiffe/svid.json`. A trusted in-guest time source is a
documented prerequisite (see spec Prerequisites).

## What This Increment Delivers

Portable, in-process issuance/rotation/mount lifecycle in package
`internal/spire`, layered on the existing workload-registration substrate
(`service.go`) and reusing `config.SPIREConfig`:

- `svid.go` — the `SVID` type, SPIFFE-ID grammar validation, X509-SVID leaf
  constraint checks, chain verification (`Verify`), and the on-disk `Document`
  (JSON, PEM material) with round-trip marshal/parse.
- `issuer.go` — the `Source` seam (`FetchSVID`) and `ErrSPIREUnavailable`
  fail-safe contract; `LocalCASource`, an in-process dev/test CA that mints
  fresh, independently verifiable leaves per call.
- `manager.go` — `Manager`: acquire → verify → atomic 0400 write → rotate
  `RefreshBefore` ahead of expiry, with an injectable `Clock`, fail-closed
  startup, retain-while-valid / remove-on-expiry rotation semantics, and
  context-cancellable shutdown.
- `credwrite_unix.go` / `credwrite_other.go` — directory-fd-anchored
  (openat/renameat, O_NOFOLLOW) atomic credential write on unix; a portable
  fallback elsewhere.

The production `Source` (a go-spiffe Workload API client dialed over the
Firecracker vsock proxy from inside the guest) is defined as an interface only
and deferred — it requires a live SPIRE agent and cannot be exercised on a
developer host. `LocalCASource` exercises the full lifecycle deterministically.

## Relationship to Credential Isolation (issue #19)

`internal/credisolation` protects `/var/run/secrets/daax` (and its canonical
`/run/secrets/daax`) via `GuardMounts`. The SVID lands at
`/var/run/secrets/spiffe/svid.json` — a **sibling** of `daax`, not a descendant.
`GuardMounts` treats siblings as allowed (see its tests), so the two features do
not collide and neither breaks the other. Hardening the SPIFFE subtree under the
same mount-guard is possible future work, not required here.

## Deferred / Follow-up

- Production in-guest SPIRE agent attestation and Workload API `Source` over
  vsock.
- SPIRE server/agent deployment and registration wiring end-to-end.
- Trusted in-guest time source provisioning.
- Runtime wiring: mounting the document into the live guest filesystem and
  aborting VM start on issuance failure.
- End-to-end test against a real SPIRE deployment (not possible on a dev host).

## Validation

- `go build ./...`, `go vet`, `gofmt`, `go test -race ./internal/spire/...`,
  `mage test` — all green.
- New lifecycle files ~80% statement coverage; lifecycle driven by a
  deterministic fake clock (no wall-clock rotation wait).
