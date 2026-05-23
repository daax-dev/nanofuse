# Workflow

## Planning
- A plan is required for any non-trivial change.
- Trivial: typo fix, single-line config update, obvious rename. Everything else requires a plan.
- This repo is spec-driven. Non-trivial features flow spec -> plan -> implement (jp-spec-kit / SpecKit; see `.claude/CLAUDE.md` and `.specify/`). Specs are technology-agnostic (WHAT/WHY, not HOW).
- Write the plan down — in the PR description, the Backlog.md task, or `.logs/decisions/`. Plans held only in chat do not count.
- Present trade-offs as facts: option, cost, risk, reversibility. The operator decides; the agent executes.
- Do not start coding until the plan is approved.

---

## Execution Discipline
- State assumptions that affect implementation. If the request has multiple plausible readings, ask before editing.
- Smallest change that satisfies the verified goal. No speculative features, abstractions, or config.
- Touch only what the task requires. No adjacent cleanup or drive-by refactors. Every changed line traces to the request or its validation.
- Remove only the orphans your change created; leave pre-existing dead code (mention it, don't delete).
- Define a verifiable goal before coding. Add or update tests when behavior changes.
- Keep files in their correct directory: `scripts/` = executables only, `docs/` = user docs, `docs/building/` = WIP/implementation docs, `test/` = tests/fixtures, `examples/{name}/` = self-contained examples. See `.claude/CLAUDE.md`.

---

## Work Intake
Tasks originate from (check in this order):
1. Backlog.md — the `backlog/` directory is the canonical task system (`backlog task list --plain`, `backlog task create`, `backlog task edit`). Tasks reference the matching spec under `.specify/features/{branch}/`.
2. GitHub Issues / PR review threads on `daax-dev/nanofuse`.
3. Direct request from operator.

Identify the source before starting. If the same task appears in multiple systems, ask which is canonical.

---

## Model Selection
- Match model capability to task complexity. Do not waste large models on small tasks.
- Code with one model; validate with a model from a **different provider where possible** (e.g., produced by Claude/Anthropic, validated by Codex/OpenAI, or vice versa). Prefer cross-provider; a different model from the same provider is the fallback; same model is last resort. Record both — producer and validator — in the PR description, and note if cross-provider was not possible.
- Call out when a task requires a paid API call. State the cost estimate before incurring it.

---

## Communication
- Report blockers immediately. No silent workarounds.
- Surface uncertainty. State confidence level. No claims of certainty without a validated primary source.
- Objective language. No first-person pronouns. No apologies.

---

## Definition of Done
A task is done only when:
- [ ] Full local gate passes: `mage ci` (clean -> build -> `go vet` + golangci-lint -> `go test -v -race -coverprofile=coverage.out ./...` -> security check). Tests-only fast path: `mage test`.
- [ ] Formatter and linter pass with no errors (`go fmt ./...`, golangci-lint via `.golangci.yml`).
- [ ] PR opened with problem statement, approach, and test evidence.
- [ ] Non-trivial decisions logged in `.logs/decisions/` per `.claude/history.md`.
- [ ] Validation pass by a separate model — cross-provider (Claude <-> Codex) where possible — recorded in the PR description as `Validation:` producer model + validator model + verdict (note if cross-provider was not possible).
- [ ] Backlog.md task updated to done with a link to the PR/commit.
- [ ] New code maintains coverage; project rule is >= 80% for new code (per `.claude/CLAUDE.md`; not currently enforced as a CI gate — Codecov upload is non-blocking).
