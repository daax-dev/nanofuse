<!-- CLAUDE.md and AGENTS.md share the Operator Preferences and Hard Guardrails below. Keep them in sync. -->

# AGENTS.md

Entry point for OpenAI Codex and compatible agents.

---

## Project
Name: Nanofuse
Purpose: Firecracker-based microVM platform for running untrusted code in secure, isolated sandboxes (self-hosted, E2B-style). Ships a `nanofuse` CLI and a `nanofused` API daemon.

---

## Operator Preferences
<!-- Operator-specific. Revise or replace when applying to a different operator. -->
- State facts only. No sugarcoating.
- Surface problems, blockers, and risks immediately.
- Consult before one-way-door decisions and before any architectural change.
- Never guess. If validation is not possible, say so explicitly.
- Objective language. No first-person pronouns. No apologies or hedges.

---

## Hard Guardrails (always apply)
- Plan before any non-trivial change. Write the plan down. Wait for approval.
- Never commit or merge directly to `main`.
- Never commit secrets, tokens, keys, or `.env` files with live values.
- No destructive git (`reset --hard`, force-push, branch delete) without explicit operator approval.
- Never overwrite uncommitted user changes. Inspect existing patterns before editing.
- Run formatter, linter, and tests after changes (`mage ci`). If that is not possible, state exactly why.
- Log non-trivial decisions to `.logs/decisions/<topic>.jsonl`.
- Repo-local instructions override these template defaults.

---

## Required Reading
`.claude/workflow.md` — planning and definition of done — applies to every task. Read it before starting work.

Read the matching file **before** you:
- write or edit code → `.claude/language.md` (Go formatting, linting, testing)
- make an architectural or cross-boundary decision → `.claude/architecture.md`
- touch dependencies, runtime, or infrastructure → `.claude/stack.md`
- perform branch / PR / commit / merge operations → `.claude/sourcecontrol.md`
- write a decision or reference log entry → `.claude/history.md`

---

## Repository Map
- `cmd/nanofuse/` — CLI entry point (cobra). `cmd/nanofused/` — API daemon entry point. `cmd/extract-test/` — test helper.
- `internal/` — implementation packages: `api`, `builder`, `client`, `clierrors`, `config`, `firecracker`, `inspect`, `layer`, `layerbuild`, `logging`, `network`, `output`, `policy`, `recording`, `registry`, `spire`, `storage`, `types`, `validate`.
- `api/` — `openapi.yaml` REST contract + README. `layers/` — microVM image layer definitions (base-os, runtimes, tools).
- `magefile.go` — build/test/lint/CI automation (mage). `config.dev.yaml` — local dev daemon config.
- `.github/workflows/` — CI (`ci.yaml`), release (`release.yaml`), PR comment (`pr-comment.yaml`).
- `backlog/` — Backlog.md task system (source of work intake). `docs/` — user + developer docs; `docs/building/` — WIP/implementation docs.
- `.claude/CLAUDE.md` — extended repo-specific guidance (spec-driven dev, file-organization, no-stale-images). `.specify/` — jp-spec-kit artifacts and constitution.
- `test/` — integration and e2e tests. `systemd/`, `nanofused.service` — daemon service units.
