<!-- CLAUDE.md and AGENTS.md share the Operator Preferences and Hard Guardrails below. Keep them in sync. -->

# CLAUDE.md

## Project
Name: Nanofuse
Purpose: Firecracker-based microVM platform for running untrusted code in secure, isolated sandboxes (self-hosted, E2B-style).
Goal: Production-ready control plane — `nanofuse` CLI + `nanofused` daemon — managing microVM lifecycle, networking, images, and snapshots with hardware-level isolation and sub-second boot times. Currently Alpha (~60% of Phase 1 core infrastructure).

---

## Operator Preferences
<!-- Operator-specific. Revise or replace when applying to a different operator. -->
- State facts only. No sugarcoating.
- Surface problems, blockers, and risks immediately.
- Consult before one-way-door decisions and before any architectural change.
- Never answer from a guess. Validate claims against primary sources. If validation is not possible, say so explicitly.
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
`.claude/workflow.md` is always loaded (see include below) — planning and definition of done apply to every task.

Read the matching file **before** you:
- write or edit code → `.claude/language.md` (formatting, linting, testing for Go)
- make an architectural or cross-boundary decision → `.claude/architecture.md`
- touch dependencies, runtime, or infrastructure → `.claude/stack.md`
- perform branch / PR / commit / merge operations → `.claude/sourcecontrol.md`
- write a decision or reference log entry → `.claude/history.md`

## Additional Repo Context
This repo predates this template and carries its own deep guidance. Treat the following as authoritative repo context (not superseded by this template):
- `.claude/CLAUDE.md` — spec-driven development (Flowspec), `backlog.md` task management, file-organization rules, no-stale-images policy, error-handling and TDD requirements.
- `.flowspec/memory/constitution.md` and `.claude/constitution.md` — project constitution. On conflict, the constitution wins.
- `docs/GOALS.md`, `ROADMAP.md`, `README.md` — mission, roadmap, and component overview.

@.claude/workflow.md
