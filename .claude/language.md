# Language Conventions

Entries marked "none observed; confirm with operator" are gaps — treat as a question for the operator, not a guess.

For each active language, this file records:
1. Pinned version and how it is pinned.
2. Formatter and config location.
3. Linter and config location.
4. Type checker and strictness level.
5. Test framework and coverage threshold.
6. Any style rules that override the formatter's defaults.

---

## Active Languages

### Go (primary)
- Version: 1.25 — CI pins `GO_VERSION: '1.25'` (`.github/workflows/ci.yaml`); `go.mod` declares `go 1.25.0`. Module path `github.com/daax-dev/nanofuse`.
- Package manager: Go modules (`go.mod` / `go.sum`), `modules-download-mode: readonly`. No vendoring. `go.sum` is committed; Dependabot manages bumps.
- Formatter: gofmt via `go fmt ./...` (run by `mage lint`). No hand-formatting.
- Linter: golangci-lint, config `.golangci.yml` — enables `bodyclose`, `gocyclo` (min-complexity 20), `gosec` (G104 excluded), `misspell`, `prealloc`, `revive` (exported-symbol-doc rule disabled). `revive` skipped under `cmd/`; `errcheck`/`gosec` skipped in `_test.go`. Excludes `third_party`, `builtin`, `examples`. Plus `go vet ./...`.
- Type checker: `go vet` + golangci-lint static analysis (staticcheck `all` minus `QF*`).
- Tests: stdlib `testing` with `-race` (table-driven where practical); gdt (`github.com/gdt-dev/gdt`) for declarative scenario tests (`mage testgdt`, or the per-suite targets `mage testgdtbuild`, `mage testgdtcli`, `mage testgdtapi`, `mage testgdte2e`). Integration tests under build tag `integration` in `test/integration` (`mage testintegration`, needs sudo for networking).
  - Unit: `mage test` -> `go test -v -race -coverprofile=coverage.out ./...`.
- Coverage threshold: project rule >= 80% for new code (per `.claude/CLAUDE.md`). Not enforced as a CI gate — Codecov upload uses `fail_ci_if_error: false`.
- Error handling: explicit on every path. No silent failures, no ignored error returns, no generic messages. Wrap with context (`fmt.Errorf("...: %w", err)`).

### Shell (bash)
- Used for build/install/fixture scripts under `scripts/` and `mage` shell-outs.
- Version target: bash 5.x.
- Linter: none observed; confirm with operator (shellcheck is referenced in a PRD spec and the `.flowspec` pre-commit template, but is not wired into CI or `mage`).
- Style: Use `set -euo pipefail` in new or modified scripts where compatible. Quote all expansions. No `eval`.

---

## Cross-Cutting Rules
- No language rule overrides the formatter. Fix the config, not the code.
- Generated code lives under a path excluded by the linter/formatter (`third_party`, `builtin`). Never edit generated files by hand.
- Lockfiles (`go.sum`) are committed. Updating a dependency is a deliberate change — call it out in the PR.
- Run `mage ci` before declaring work done (build + `go vet` + golangci-lint + race tests + security check).
- File-organization discipline applies: executables in `scripts/`, WIP docs in `docs/building/`, tests in `test/`. See `.claude/CLAUDE.md`.
