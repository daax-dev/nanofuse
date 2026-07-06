# Feature Specification: Per-microVM Cryptographic Identity (SPIFFE SVID)

**Feature Branch**: `feat/issue-17-spiffe-svid-lifecycle`
**Created**: 2026-07-06
**Status**: Partial — issuance/rotation lifecycle delivered; production attestation deferred
**Input**: GitHub issue #17 — "Implement SPIFFE SVID issuance for per-microVM cryptographic identity"

## Overview

Every microVM must possess a unique, short-lived, verifiable cryptographic
identity so that workloads can authenticate to peers and to platform services
without long-lived shared secrets. The identity is expressed as a SPIFFE
X.509-SVID: a certificate binding exactly one SPIFFE ID to a private key,
verifiable against a trust bundle. The identity must be short-lived and rotate
automatically before expiry, and the platform must fail closed — a workload
without a valid identity must not run rather than fall back to static or
plaintext credentials.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Each microVM receives a unique identity (Priority: P1)

An operator starts an untrusted workload in a microVM. The workload is issued a
unique SPIFFE X.509-SVID at a well-known location and can use it to establish
mutual TLS with platform services and peers.

**Why this priority**: Per-VM identity is the foundation for all
identity-scoped policy (network, secrets, API access). Without it there is no
subject to authorize.

**Independent Test**: Issue an SVID for a workload identity and confirm the
certificate's single URI SAN equals the requested SPIFFE ID and that the
certificate verifies against the returned trust bundle.

**Acceptance Scenarios**:

1. **Given** a workload identity, **When** an SVID is issued, **Then** the
   certificate carries exactly one URI SAN equal to the SPIFFE ID and verifies
   against the trust bundle for both client and server authentication.
2. **Given** two distinct workloads, **When** each is issued an SVID, **Then**
   the SPIFFE IDs and key material differ.

### User Story 2 - Identity is delivered as a protected, well-known document (Priority: P1)

The issued identity is written to a fixed path inside the guest as an
owner-read-only document so an in-guest consumer can load it, and no other
principal on the guest can read the private key.

**Why this priority**: The private key is the sensitive asset; delivery must not
widen its exposure.

**Independent Test**: After issuance, confirm the document exists at the
canonical path, is mode 0400, contains the leaf chain, private key, and trust
bundle, and round-trips back to a valid SVID.

**Acceptance Scenarios**:

1. **Given** a successful issuance, **When** the document is written, **Then**
   it is created atomically with mode 0400 and is never observable in a
   partially written or world-readable state.
2. **Given** an insecure or symlinked target directory, **When** a write is
   attempted, **Then** the write is refused.

### User Story 3 - Identity rotates before expiry (Priority: P1)

The identity is short-lived and is replaced with a fresh one before it expires,
without interrupting the workload.

**Why this priority**: Short lifetimes bound the blast radius of a leaked key;
automatic rotation makes short lifetimes operable.

**Independent Test**: With a deterministic clock, advance time to the rotation
point and confirm a new certificate (new serial, new validity window) is issued
and persisted before the previous one expires.

**Acceptance Scenarios**:

1. **Given** an issued SVID with a bounded TTL, **When** the refresh lead time
   before expiry is reached, **Then** a fresh SVID is issued and persisted while
   the previous one is still valid.
2. **Given** rotation is temporarily failing, **When** the current SVID is still
   valid, **Then** it is retained and rotation retries; **When** it has expired,
   **Then** the credential is removed so consumers fail closed rather than
   present an expired identity.

### User Story 4 - Fail closed when identity cannot be obtained (Priority: P1)

If an identity cannot be obtained at startup, the workload does not start.

**Why this priority**: The security guarantee is that no workload runs without a
verifiable identity. A silent fallback to no/weak identity defeats the feature.

**Acceptance Scenarios**:

1. **Given** the issuing endpoint is unreachable, **When** startup issuance is
   attempted, **Then** startup returns an error that names the unreachability
   and starts no background work.
2. **Given** a structurally invalid, expired, or not-yet-valid SVID is returned,
   **When** it is validated, **Then** it is rejected and no credential is
   written.

## Success Criteria *(measurable)*

- SVID URI SAN equals the requested SPIFFE ID; certificate verifies against the
  trust bundle for both clientAuth and serverAuth. (verifiable by test)
- Document is written atomically at the canonical path with mode 0400.
  (verifiable by test)
- A fresh SVID is issued and persisted before the previous SVID expires, driven
  by a deterministic clock. (verifiable by test)
- Startup fails closed (no background work, named error) when issuance fails.
  (verifiable by test)
- Expired credential is removed on persistent rotation failure. (verifiable by
  test)

## Prerequisites & Assumptions

- A trusted in-guest time source. SVID validity and rotation scheduling depend
  on the guest clock; an untrusted or skewed clock undermines expiry
  enforcement. This is a documented operational prerequisite, not code in this
  increment.

## Out of Scope (this increment)

- Standing up the identity-issuing infrastructure (SPIRE server/agent
  deployment).
- The production in-guest attestation path that obtains SVIDs from a real
  Workload API over the host bridge. This is a one-way-door architectural piece
  requiring live infrastructure and cannot be exercised on a developer host; it
  is deferred and tracked as follow-up.
