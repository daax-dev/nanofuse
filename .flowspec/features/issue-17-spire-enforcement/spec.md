# Feature Specification: SPIRE Fail-Closed Enforcement for microVM Startup

**Issue:** daax-dev/nanofuse#17 (DoD AC4)
**Status:** In Progress
**Type:** Security enforcement

## Why

Static credentials in sandboxed workloads are attackable at rest. The platform
issues short-lived SPIFFE identities (SVIDs) to replace them. A security control
is only sound if it cannot be silently bypassed: if the identity provider is
unreachable, the platform must refuse to run the workload rather than start it
without an identity. The current behavior does the opposite — an identity-provider
failure is logged and the workload starts anyway (fail-open), which reopens the
exact plaintext-credential window this feature exists to close.

Issue #17 DoD AC4 states: *"If SPIRE is unavailable at startup, microVM startup
fails with an error message naming SPIRE unreachability (fail-safe, not fallback
to plaintext)."*

## What

Introduce an operator-controlled, default-OFF fail-closed mode for workload
identity provisioning. When an operator opts in, a workload whose identity cannot
be provisioned must not start, and no partial/leaked resources may remain.

## Scope

### In scope (Phase 1 — implement)
- Host-side workload-identity registration during microVM creation.
- An opt-in setting that makes identity provisioning mandatory.
- Fail-closed behavior with an actionable, identity-naming error.
- Guaranteed cleanup of any resources provisioned before the failure.

### Out of scope (Phase 2 — spec/plan only, see plan.md)
- The in-guest identity agent ("topology B") and the trusted in-guest time
  source that TTL enforcement depends on. Documented as a decision-ready design
  with trade-offs and a recommendation; not built here.

## Actors
- **Operator** — configures the platform; chooses whether identity is mandatory.
- **API client** — requests microVM creation.

## Requirements

- **FR-1** The platform MUST provide a configuration setting that makes workload
  identity provisioning mandatory. It MUST default to disabled so existing
  behavior is unchanged unless explicitly enabled.
- **FR-2** When identity is enabled AND mandatory, a failure to provision a
  workload identity during microVM creation MUST cause the creation request to
  fail.
- **FR-3** The failure response MUST carry an actionable error that names the
  identity provider (SPIRE) as the cause of unavailability, and MUST map to an
  appropriate unavailability status.
- **FR-4** On such a failure, the platform MUST release every resource
  provisioned for the request so far, leaving no persisted microVM record and no
  orphaned network/storage/identity resources.
- **FR-5** When identity is NOT mandatory, a provisioning failure MUST preserve
  today's best-effort behavior: log a warning and continue microVM creation.
- **FR-6** When identity provisioning succeeds, microVM creation MUST proceed
  regardless of the mandatory setting.

## Acceptance Criteria (measurable)

- **AC-1 (core, DoD AC4):** With identity mandatory and the provider unreachable,
  a create request fails with a response whose error text names SPIRE
  unreachability, and no microVM record, network, storage, or identity entry
  remains afterward.
- **AC-2:** With identity mandatory and the provider reachable, a create request
  succeeds and the microVM record carries the issued identity.
- **AC-3:** With identity NOT mandatory and the provider unreachable, a create
  request still succeeds; the failure is logged as a warning only and no identity
  is attached.
- **AC-4:** The mandatory setting defaults to disabled; a configuration with no
  explicit value behaves exactly as before this change.
- **AC-5:** The full local quality gate passes (build, vet, lint, race tests).

## Success Criteria
- No fail-open path exists when the operator has declared identity mandatory.
- Zero behavior change for operators who do not opt in.
