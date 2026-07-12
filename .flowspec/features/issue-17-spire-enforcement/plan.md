# Implementation Plan: SPIRE Fail-Closed Enforcement

**Issue:** daax-dev/nanofuse#17 (DoD AC4)
**Spec:** ./spec.md

## Phase 1 — Fail-closed host-side enforcement (IMPLEMENT)

### Config flag
- Add `Required bool` (yaml `required`) to `config.SPIREConfig`
  (`internal/config/config.go`).
- Default `false` in `DefaultConfig()` so existing deployments are unchanged.
- Semantics: enforcement is active only when `SPIRE.Enabled && SPIRE.Required`.
- **Config safety (premortem finding):** `Config.Validate()` rejects the
  contradictory `Required && !Enabled` combination as a fatal startup error.
  Without this, an operator who sets `required` but forgets `enabled` would
  silently fail *open* (registration is skipped when disabled) — the exact
  outcome AC4 forbids.

### Reviewer premortem outcomes (adversarial review)
- **Accepted / fixed:** `Required && !Enabled` silent fail-open → added
  `validateSPIRE()` startup rejection.
- **Rejected as out-of-scope:** `cleanupCreatedVMResources` logs cleanup errors
  as WARN and continues (best-effort). This is *pre-existing* behavior shared
  with the existing DB-create-failure path; escalating logging or adding a GC
  sweep is an unrelated refactor and is left untouched per change-discipline.

### Enforcement point
- `internal/api/vm_handlers.go`, `handleCreateVM`, immediately after the
  `registerSPIREWorkload` call (~line 789). This is the correct seam because at
  that point networking and writable root disks have been provisioned but the VM
  record has NOT yet been persisted (`s.db.CreateVM` is later) and no SPIRE entry
  was created (`CreateVMWorkloadEntry` failed → empty spiffeID).
- New behavior when `spireErr != nil`:
  - If `s.spireRequired()` (`s.config.SPIRE.Enabled && s.config.SPIRE.Required`):
    call `s.cleanupCreatedVMResources(vmID, config)` (releases TAP device, egress
    policy, IPAM lease, writable storage) and return
    `types.WriteError(w, http.StatusServiceUnavailable, types.ErrServiceUnavailable, <msg>, nil)`
    where `<msg>` names SPIRE unreachability and is actionable.
  - Else: keep today's best-effort path (log `WARN`, proceed).
- No SPIRE entry cleanup is needed on this path: registration failed, so nothing
  was registered. (Entry cleanup on the later `CreateVM` DB-failure path is
  unchanged.)

### Testability seam
- `Server.spireService` is currently the concrete `*spire.Service`, whose
  `CreateVMWorkloadEntry` shells out to `docker exec` — it cannot be made to
  succeed in a unit test, so the mandatory-success case (AC-2) is untestable
  against the concrete type.
- Minimal fix: introduce a package-private interface `spireRegistrar` in
  `internal/api` with exactly the methods the handlers use
  (`IsEnabled`, `ValidateIdentityParams`, `CreateVMWorkloadEntry`,
  `DeleteVMWorkloadEntry`) and change the struct field type to it. `*spire.Service` already satisfies it; `server.go`
  assignment is unchanged. Nil-interface checks (`s.spireService == nil`) keep
  working for tests that don't set the field.
- Add a `spireRegistrarStub` in the api test package to inject success/failure,
  mirroring how `runtimeImageProviderStub` injects `snapshotErr`/`loadSnapshotErr`.

### Tests (TDD — write failing first)
1. `Required + SPIRE unreachable` → `handleCreateVM` returns 503,
   error text names SPIRE, `db.ListVMs` empty, IPAM lease released, stub
   `DeleteVMWorkloadEntry` not needed (never registered). (AC-1)
2. `Required + SPIRE ok` (stub returns spiffeID) → 201, persisted VM has
   `SpiffeID`. (AC-2)
3. `not Required + SPIRE unreachable` → 201, VM persisted with empty SpiffeID
   (WARN only). (AC-3)
4. Config default: `DefaultConfig().SPIRE.Required == false`. (AC-4)

### Quality gate
`mage test`; `go test -race ./...`; `go vet ./...`; `gofmt -l`; golangci-lint
v2.12.2 (CI pin) → 0 issues.

---

## Phase 2 — In-guest identity agent ("topology B") + trusted time (SPEC/PLAN ONLY)

TTL enforcement (SVID expiry at 60 min, refresh at 15 min before) is only
trustworthy if the component enforcing it has a trustworthy clock. This section
lays out the design space for the operator to decide. **Not implemented here.**

### Problem
An in-guest SPIRE agent (or in-guest SVID Manager) validating/rotating SVIDs must
answer "has this SVID expired?" A guest that controls its own wall clock can roll
time backwards to keep an expired SVID "valid", defeating rotation as a security
control. The guest is, by threat model, untrusted.

### The three topologies

- **Topology A (current, shipped):** Host-side SPIRE server/agent registers the
  workload; the host issues/rotates the SVID and mounts `svid.json` into the
  guest. The trusted clock is the host's. Guest consumes but does not enforce.
  Simple; already in `main`. Limitation: the SVID's private key transits the host
  mount; the guest is not directly attested to the SPIRE agent.
- **Topology B (this Phase 2):** A SPIRE agent runs *inside* each guest, attests
  over the vsock proxy to the host SPIRE server, and holds the Workload API
  socket in-guest. Stronger isolation (private key minted in-guest, per-guest
  attestation), but the in-guest agent needs trusted time.
- **Topology C (hybrid):** In-guest workload consumes SVIDs but a host-side
  monitor kills the microVM when the host clock says the SVID has expired.
  Enforcement stays host-side; guest gains no clock authority.

### Trusted-time options for Topology B

| Option | Mechanism | Trust root | Cost | Risk |
|---|---|---|---|---|
| T1 KVM PTP (`ptp_kvm`) | Guest reads host clock via paravirtual PTP device | Host TSC | Low — kernel config + guest chrony | Requires guest kernel `CONFIG_PTP_1588_CLOCK_KVM`; guest can still ignore it unless enforcement reads PTP directly |
| T2 vsock time attestation | Agent fetches signed timestamp from host over vsock before each TTL check | Host + signature | Medium — new host service + guest client | Adds a per-check round trip; replay window needs a nonce |
| T3 Roughtime/authenticated NTP | Agent queries an external authenticated time service | External authority | Medium — network egress + trust config | Needs guest egress; external dependency and its own availability |
| T4 Host-enforced expiry (=Topology C) | Host monitor, not guest, enforces expiry | Host | Low | Not true in-guest enforcement; keeps enforcement on host |

### Recommendation (for operator decision)
Adopt **Topology C / option T4 as the near-term path** and defer a full
Topology-B in-guest agent. Rationale: it delivers the security outcome AC4
targets (an expired identity cannot be used to keep a workload running) using the
host as the trust root — which the platform already is for VM lifecycle — without
introducing a guest-trusted-clock problem that has no clean solution under an
untrusted-guest threat model. If per-guest attestation with an in-guest private
key later becomes a hard requirement, pair Topology B with **T1 (`ptp_kvm`)** as
the lowest-cost trusted-time source, gated behind a guest-kernel capability check,
and treat T2 as the fallback where the guest kernel lacks PTP. This is a
one-way-door architectural choice and must be ratified by the operator before any
build.

### Phase 2 open questions
- Does the target guest kernel (6.1.x LTS) ship `ptp_kvm`? (verify before T1)
- Is per-guest in-guest private-key custody a hard requirement, or is host-mount
  custody (Topology A) acceptable with host-enforced expiry (Topology C)?
- What is the acceptable enforcement latency (host monitor poll interval)?
