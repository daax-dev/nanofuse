# Source Control

---

## Repository
- Host: GitHub — `github.com/daax-dev/nanofuse` (the `origin` remote).
- Identity note: the canonical org is `daax-dev`. The Go module path (`github.com/daax-dev/nanofuse`), `README.md` badges / install URLs, GHCR image references, and repo metadata are all aligned to `daax-dev/nanofuse`. A prior three-way drift (`jpoley`, `peregrinesummit`, `daax-dev`) was resolved in favor of `daax-dev`; see `.logs/decisions.jsonl`.
- Default branch: `main`.
- All work lands via PR. No direct commits to `main`.

---

## Branch Naming
- Feature: `feature/<short-topic>`
- Bug fix: `fix/<short-topic>`
- Docs: `docs/<short-topic>`
- Chore / tooling: `chore/<short-topic>`
- Claude Code sessions: harness-assigned name (e.g., `claude/<task>-<id>`). Do not rename mid-session.
- Lowercase, hyphen-separated. Keep names short.

---

## Commits
- Imperative mood, present tense: "add X", not "added X" or "adds X".
- Subject line <= 72 characters.
- Body explains the **why**. The diff shows the what.
- One logical change per commit. Mixed-purpose commits get rejected at review.
- Do not amend a commit that has already been pushed unless explicitly asked.
- A `[release]` token in a commit message triggers an auto image release in CI — use deliberately.

---

## Pull Requests
- Open a PR as soon as the branch has a meaningful commit. Draft is fine.
- PR title = leading commit subject line.
- PRs are immutable after creation: do not push additional commits to a PR head branch, do not amend the pushed commits, and do not edit the PR title/body/metadata. If any change is needed after PR creation, close that PR and create a fresh branch with a fresh PR.
- Install the repo-local push guard before PR work: `git config core.hooksPath .githooks`. The `.githooks/pre-push` hook calls `scripts/guard-pr-updates.sh` and rejects pushes to any branch that already has a GitHub PR in any state.
- PR body must include:
  - Problem statement.
  - Approach taken and alternatives considered.
  - Test evidence (commands run, output — at minimum `mage ci`).
  - Which model produced and which model validated (if AI-assisted).
- Never break the build — PRs that fail CI are rejected.
- Never merge your own PR unless explicitly authorized by the operator. Only humans merge to `main`.
- `pr-comment.yaml` posts automated PR feedback; address it before requesting human review.

---

## Worktrees
- Long-running parallel work uses `git worktree` rather than branch-switching in place.
- Worktree paths live outside the primary checkout (e.g., `/tmp/<repo>-<branch>`).
- Worktrees are disposable. Clean them up when the branch lands (`git worktree remove`).

---

## What Never Gets Committed
- Secrets, tokens, keys, connection strings, registry credentials.
- `.env` files with live values; live daemon configs with real paths/keys.
- Generated build output (`bin/`, `coverage.out`, `coverage.html`) — already in `.gitignore`.
- IDE / OS noise (`.DS_Store`, `Thumbs.db`).

---

## Destructive Operations
- Force-push to a shared branch requires explicit operator authorization.
- `git reset --hard`, branch deletion, and history rewrites require confirmation when recovery is uncertain.
- Treat destructive git operations as high-risk: pause, verify the target, get confirmation.

---

## Tags and Releases
- Tag scheme: semver `v*` (e.g., `v0.1.0`) for binary releases (`release.yaml`); `image-v*` (e.g., `image-v1.0.0`) for microVM/base image releases (`ci.yaml`).
- Release notes: GitHub releases via `softprops/action-gh-release`; image releases can be auto-tagged by CI on `[release]` commits. See `.github/RELEASE_PROCESS.md` and `.github/VERSIONING_SUMMARY.md`.
