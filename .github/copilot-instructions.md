# Copilot Instructions

GitHub Copilot reads this file automatically. Rules here are enforced in every session.

---

## Project
Name: Nanofuse
Purpose: Firecracker-based microVM platform for running untrusted code in secure, isolated sandboxes. Ships a `nanofuse` CLI and a `nanofused` API daemon.

---

## Operator Preferences
<!-- Operator-specific. Revise or replace when applying to a different operator. -->
- State facts only. No sugarcoating.
- Surface problems, blockers, and risks immediately.
- Consult before one-way-door or architectural decisions.
- Never answer from a guess. Say so when a claim cannot be validated.
- Objective language. No first-person pronouns. No apologies.

---

## Planning
- A plan is required for any non-trivial change. Trivial = typo fix, single-line config update, obvious rename.
- Write the plan first. Present it. Wait for approval. Do not start coding until approved.
- Present options with trade-offs. The operator decides; the agent executes.
- This repo is spec-driven: features flow spec -> plan -> implement. See `.claude/CLAUDE.md` and `.specify/`. Track all work in `backlog/` (Backlog.md).

---

## Stack
- Runtime: Go 1.24 (CI pins `1.24`; `go.mod` declares `go 1.24.3`). Module path: `github.com/jpoley/nanofuse`.
- Build tool: mage (`magefile.go`). Package manager: Go modules (`go.mod` / `go.sum`, readonly mode).
- Test framework: `go test` (stdlib testing) with `-race`; gdt for declarative scenario tests.
- Persistence: SQLite (`mattn/go-sqlite3`) + local filesystem data dir. Backend: stdlib `net/http`. CLI: cobra.
- Target platform: Firecracker microVMs on Linux/KVM; daemon talks over a unix socket / vsock (`mdlayher/vsock`).
- CI: GitHub Actions (`.github/workflows/ci.yaml`, runs `mage ci`). Artifact registry: GHCR (`ghcr.io`).

---

## Code Conventions
- Run the formatter before committing (`go fmt ./...` / gofmt). No hand-formatted code.
- All tests must pass before declaring done: `mage test` (`go test -v -race -coverprofile=coverage.out ./...`). Full gate: `mage ci`.
- Linter: golangci-lint, config `.golangci.yml` (revive, gosec, gocyclo min-complexity 20, bodyclose, misspell, prealloc) plus `go vet`.
- Handle every error explicitly. No silent failures, no ignored error returns, no generic messages. Return actionable, context-rich errors.
- `go.sum` is committed. Updating dependencies is a deliberate change — note it in the PR. Dependabot manages module bumps.
- Never edit generated files by hand. Keep files in their correct directory (scripts/ = executables, docs/building/ = WIP docs, test/ = tests).
- No EOL/stale base images, kernels, or fixtures (see `.claude/CLAUDE.md` no-stale-images policy).

---

## Source Control
- Repo: `github.com/daax-dev/nanofuse`. Never commit directly to `main`. All work lands via PR.
- Branch naming: `feature/`, `fix/`, `docs/`, `chore/`.
- Commits: imperative mood, present tense. Subject <= 72 characters. Body explains **why**.
- PR body must include: problem statement, approach, alternatives considered, test evidence.
- Never merge your own PR unless explicitly authorized.
- Never commit secrets, tokens, keys, or `.env` files with live values.

---

## Architecture
- Module boundary = test boundary. If two packages cannot be tested apart, they are one module.
- Secrets go through the platform secret store / daemon config. Never in source control or committed env files.
- UTC everywhere internally. Local time is a presentation concern.
- "Temporary" workarounds without an expiry date and an owner are not acceptable.
- The `nanofused` daemon owns microVM lifecycle; the `nanofuse` CLI and SDKs talk to it over the REST API (`api/openapi.yaml`). Do not bypass that boundary.

---

## Definition of Done
A task is done only when:
- All tests pass (`mage ci`).
- Formatter and linter pass with no errors.
- PR opened with problem statement, approach, and test evidence.
- No `[FILL IN]` placeholders left in affected files.
- Decisions logged in `.logs/decisions/` if a non-trivial choice was made.
- Backlog.md task updated to done with a link to the PR/commit.
