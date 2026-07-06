# Spec: gondolin-style sandbox conversion (issue #18, reframed)

## Context and reframe

Issue #18 asks nanofuse to "parse gondolin config format" and provide a
`convert` command. Primary-source verification (gondolin `docs/cli.md`,
`README.md`) establishes that **gondolin has no declarative sandbox config
file**: its sandbox is defined imperatively via `gondolin bash|exec` CLI flags
and a TypeScript `VM.create()` API. The literal task ("parse gondolin config
format") is therefore unbuildable.

Reframed scope (OQ-1 resolved = option a): provide a conversion that reads a
**nanofuse-authored mirror** of gondolin's flag surface and produces a nanofuse
spec plus an explicit **divergence report**. The deliverable's value is the
precise divergence report, not a false claim of equivalence.

## What / Why

Teams evaluating gondolin-style sandboxes need to understand exactly what
carries over to nanofuse and what does not, without a tool silently dropping
isolation-relevant controls.

## Requirements

- Read a mirror document describing a gondolin sandbox (image, resource hints,
  and the gondolin flag surface: allow-host, host-secret, mount-hostfs/memfs,
  ssh-allow-host, tcp-map, dns, env, cwd, vmm, rootfs-size).
- Produce a nanofuse spec for the fields that map faithfully.
- Produce a divergence report listing every feature with no faithful nanofuse
  equivalent, one distinct entry per feature. No silent drops.
- Fail closed by default when unrepresentable features are present; allow an
  explicit opt-in to drop them and proceed.
- Degrade an L7 host allowlist safely (more restrictive), never approximate it
  into a false-security L3/L4 rule by default.
- Reject malformed / unknown-key / multi-document input loudly; never panic.

## Success criteria (measurable)

- A valid mirror with only faithfully-mapped fields converts with zero blocking
  divergences.
- Each unrepresentable feature, present alone, yields exactly one blocking
  divergence and a fail-closed error.
- The opt-in lossy mode proceeds and marks every dropped feature loudly.
- An L7 allowlist yields a locked-down (default-drop) egress policy and a
  warning, and never fails closed on its own.
- Empty, unknown-key, wrong-type, and multi-document inputs return errors, not
  panics or misleading success.

## Out of scope

- Running the converted workload end-to-end (no gondolin runtime available).
- Faithful emulation of gondolin features nanofuse lacks.

## Open questions

- OQ-1 (resolved, option a): mirror-YAML + divergence report, not a parser for a
  non-existent gondolin file format.
