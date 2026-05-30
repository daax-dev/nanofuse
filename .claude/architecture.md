# Architecture

Architectural decisions require operator approval before implementation.
ADRs log to `.logs/decisions/architecture.jsonl` (see `.claude/history.md`).
The project constitution (`.flowspec/memory/constitution.md` / `.claude/constitution.md`) is the supreme authority; on conflict, follow the constitution.

---

## System Shape
- Control plane = `nanofused` daemon: owns microVM lifecycle, drives Firecracker processes, manages TAP networking + IPAM, images, and snapshots.
- `nanofuse` CLI (cobra) and future SDKs are clients of the daemon REST API (`api/openapi.yaml`); they do not touch Firecracker or storage directly.
- Daemon transport: unix socket (`api.socket`, mode-restricted) with optional TCP bind; host<->guest over vsock.
- Internal package layout under `internal/`: `api` (HTTP handlers, auth middleware), `firecracker` (VMM control), `network` (TAP/IPAM), `storage` (SQLite + filesystem), `registry`/`layer`/`layerbuild`/`builder` (image pipeline), `config`, `logging`, `recording`, `spire`, `types`, `validate`.

---

## Default Patterns
- Layering: client (CLI/SDK) -> daemon REST API -> internal services -> Firecracker / storage / network. Respect this boundary; do not call lower layers from the CLI.
- API style: REST over HTTP (contract in `api/openapi.yaml`).
- Idempotency: state-changing endpoints should be idempotent or accept an idempotency key. Deviations require a logged decision.
- Configuration: YAML daemon config (`config.dev.yaml` pattern) + flags/env. Secrets via the platform secret store or operator-supplied config — never committed to git.
- Time: UTC everywhere internally. Local time is a presentation concern.
- IDs: UUID (`github.com/google/uuid`).
- Errors: explicit, wrapped with context. No silent failures.

---

## Boundaries
- Module boundary = test boundary. If two packages cannot be tested apart, they are one module.
- Daemon owns all privileged operations (Firecracker, networking, storage). The CLI is a thin client.
- No shared databases between processes. Data shared via the daemon API.
- Cross-service / external calls require an explicit client with timeout and retry-with-backoff.

---

## Anti-Patterns (refuse these)
- Bypassing the daemon API to manipulate Firecracker, networking, or storage from the CLI.
- Stale / EOL base images, kernels, or fixtures (see `.claude/CLAUDE.md` no-stale-images policy and incident record).
- "Temporary" workarounds without an expiry date and an owner.
- Secrets in env files, source control, or CI variables without rotation.
- Silent error handling or ignored error returns.
- Files in the wrong directory (executables outside `scripts/`, WIP docs outside `docs/building/`, tests outside `test/`).

---

## Decision Logging
Log to `.logs/decisions/architecture.jsonl`:
```json
{"id":"arch-001","date":"YYYY-MM-DD","decision":"...","rationale":"...","alternatives":"...","references":["https://..."]}
```

---

## Reference Architectures
When citing patterns, prefer primary sources:
- Firecracker documentation (firecracker-microvm) for microVM lifecycle and security model.
- Official vendor documentation (AWS Well-Architected, Azure Architecture Center, GCP).
- NIST SP 800-series for security architecture; OWASP for application security patterns.
Cite the exact URL in `.logs/references/architecture.jsonl`.
