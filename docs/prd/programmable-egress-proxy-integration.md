# nanofuse Integration: Programmable Egress Proxy (forced mode, v1)

**Status**: Design — not yet implemented
**Canonical PRD**: `daax-dev/dx` → `arch/prd/programmable-egress-proxy.md` (v1.3.0)
**Tracking issue**: [daax-dev/dx#55](https://github.com/daax-dev/dx/issues/55)
**Builds on**: the L3/L4 egress policy (`daax-dev/dx` → `arch/prd/nanofuse-network-egress-policy.md`)
**Related nanofuse issues**: #19 (credential isolation), #17 (SPIFFE SVID), #31 (signed capability grants)

> This doc is the **nanofuse-specific** slice of the cross-cutting PRD. It does not restate the
> full design — read the canonical PRD first. nanofuse is the **v1 target** because it is the only
> place the secret-isolation guarantee actually holds (forced enforcement).

---

## Why nanofuse is the forced-mode target

A nanofuse microVM guest has **no network path off the VM except through the host**. The L3/L4
layer already drops `0.0.0.0/0` on the TAP interface and permits only an allowlist. By pointing
that allowlist at the egress proxy and *nothing else*, the guest is physically unable to bypass
the proxy — there is no other route. This is what makes "the agent uses a credential it never
possesses" a real guarantee rather than a convention.

This is the property the daax-devtools container **cannot** provide (shared kernel, bypassable
`HTTPS_PROXY`), which is why devtools is Phase 2 and clearly labelled cooperative.

---

## ⚠ Blocking prerequisite — L3 PRD "proxy-enabled job" contract

The existing L3/L4 PRD assumes immutable job-start policy with **pinned upstream IPs**: the guest
resolves allowed hostnames and reaches those upstream IPs directly. **The proxy model changes
this.** For a proxy-enabled job, the guest egress allowlist must be **only**:

```
{ proxy IP:port (3128/3129), controlled DNS resolver, boot/provisioning services }
```

Upstream API IPs are **never** reachable directly from the guest — only the host proxy dials them.
This must be ratified as an explicit "proxy-enabled job" mode in the L3 PRD before this integration
can be implemented (PRD milestone M5 hard dependency). Until then, forced enforcement is undefined.

---

## What lands in nanofuse

The proxy data plane and policy engine are a **shared Go library** (per the canonical PRD). The
nanofuse repo owns the **adapter** — the wiring that makes enforcement *forced*:

| Area | nanofuse-specific responsibility | Touches |
|------|----------------------------------|---------|
| Forced routing | nftables on the TAP: drop-all-except-{proxy, DNS, boot}; **fail-closed** (deny installed before guest's first packet, before proxy readiness) | `internal/network/` (extends `nat.go`, `tap.go`, portforward) |
| Proxy placement | Run the egress proxy as a `nanofused` component (or sibling daemon) bound to the VM bridge gateway IP | `cmd/nanofused`, `nanofused.service` |
| Client identity | Per-sandbox **mTLS**: provision a client cert/key into the guest at boot; TAP source binds identity (candidate SPIFFE SVID, #17) | guest provisioning path |
| CA trust | Install the **public** cert of the per-run CA into the guest trust store at image build/boot; CA **private key never enters the guest** | guest image / boot provisioning |
| IPv6 off | Disable IPv6 in the guest stack; ensure DNS returns no AAAA; proxy refuses IPv6 upstream (FR-19, all three layers) | guest config + `internal/network/` |
| Job spec | Extend the job/sandbox spec so a tenant declares egress allow rules + secret references (names only) | job spec schema |
| Audit | Emit `egress_brokered`/`egress_denied` JSONL into the nanofuse audit stream with `enforcement: "forced"` | audit log |
| Teardown | Delete per-run CA material, revoke client cert, remove nftables chain on VM teardown | VM lifecycle (`Drop` on TAP handle) |

---

## Job spec sketch (illustrative — final schema TBD with canonical PRD)

```yaml
sandbox:
  network:
    egress_proxy:
      enabled: true          # switches the job into proxy-enabled mode (L3 contract above)
      allow:
        - host: api.anthropic.com
          port: 443
          methods: [POST]
          path: /v1/messages
          secret: anthropic   # name only; value lives on the host
          ttl: 1h
        - host: "*.github.com"
          port: 443
          methods: [GET, POST]
          secret: ghpush
```

When `egress_proxy.enabled: true`, the L3 layer installs the proxy-only allowlist; when false,
existing L3/L4 behaviour is unchanged.

---

## Acceptance criteria (nanofuse-scoped subset of the PRD)

- [ ] Guest calls an allowlisted upstream with **no key in the guest**; the request leaving the
      guest (captured at the intercept boundary) has no `Authorization`; the proxy injects it.
- [ ] `grep` of guest fs/env/memory fixtures + audit log for the secret value → no match.
- [ ] Direct IP connect / raw socket / alt port / IPv6 from the guest → blocked at nftables.
- [ ] Proxy stopped → guest egress fails closed (no NAT fallback); health/audit signal emitted.
- [ ] Per-run CA private key absent from guest image/mounts; only public cert present.
- [ ] nftables chain + CA material removed cleanly on teardown (no leaks).

---

## Out of scope here (see canonical PRD / Phase 2)

devtools cooperative mode, OAuth refresh, HTTP/2 & WebSocket, tunnel mode, HTTP API/SDK,
multi-tenant RBAC store, response redaction, deny rules, signed capability grants.
