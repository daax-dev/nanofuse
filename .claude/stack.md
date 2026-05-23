# Stack

Entries marked "none observed; confirm with operator" are unverified in-repo — treat as a question for the operator, not a guess.
Only document what is confirmed and deployable today.

---

## Runtime
- Go 1.24 (CI env `GO_VERSION: '1.24'`; `go.mod` declares `go 1.24.3`).
- Module path: `github.com/daax-dev/nanofuse`, aligned with the git remote and `README.md` badges/install URLs. Canonical org is `daax-dev` — see sourcecontrol.md.
- Target platform: Linux host with KVM (`/dev/kvm`), x86_64. Workloads run as Firecracker microVMs.

## Frameworks
- Backend: Go stdlib `net/http` (REST API in `internal/api`; contract in `api/openapi.yaml`).
- Frontend: none.
- CLI: cobra (`github.com/spf13/cobra`), entry point `cmd/nanofuse`.
- Daemon: `cmd/nanofused`, runs as a systemd service (`nanofused.service`), exposes a unix socket and optional TCP bind.

## Persistence
- Primary: SQLite (`github.com/mattn/go-sqlite3`; `config.dev.yaml` -> `storage.database`).
- Cache: none.
- Search: none.
- Object storage: local filesystem (`storage.data_dir`); microVM images pulled from an OCI registry (`google/go-containerregistry`).

## Messaging / Eventing
- none (control is synchronous over the daemon API; host<->guest uses vsock `github.com/mdlayher/vsock`).

## Auth
- Daemon API has no application-layer auth today: `internal/api` exposes only request `loggingMiddleware` (`server.go`); no Authorization/API-key header checks, and no `internal/policy` package exists.
- Access control: unix socket file permissions (`daemon.socket_mode` default `0660`, `daemon.socket_group` default `nanofuse`); optional `daemon.tcp_bind` exposes the API over TCP with no built-in authentication — bind cautiously.
- SPIRE/SPIFFE: config scaffolding under `internal/spire` and `internal/config` (server/agent socket paths); not wired into API request authz.
- Registry auth via `~/.docker/config.json` (`registry.auth_config_path`).

## Observability
- Traces: none observed in-repo; `internal/recording` handles session recording, not distributed tracing. Confirm with operator if tracing is planned.
- Metrics: none observed in-repo; confirm with operator.
- Logs: structured logging via `internal/logging`; daemon log level/format configurable (`logging.level`, `logging.format`).

## Build / Package
- Go: go modules (`go.mod` / `go.sum`), `modules-download-mode: readonly`. No vendoring.
- Build automation: mage (`magefile.go`). Key targets: `cli`, `daemon`, `all`, `lint`, `test`, `testintegration`, `ci`, `validate`, `install`, `imagebuild`.
- CI: GitHub Actions — `.github/workflows/ci.yaml` (runs `mage ci`, Codecov upload, golangci-lint job, `mage validate`, govulncheck), `release.yaml` (on `v*` tags), `pr-comment.yaml`. Dependabot enabled (`.github/dependabot.yml`).
- Artifact registry: GHCR (`ghcr.io`) for microVM/base images; binaries published as GitHub release assets.

## Explicitly Not in Stack
List rejected tools and the reason. Prevents re-proposal.
- Stale / EOL base images, kernels, and test fixtures — banned. Use actively supported versions only (Ubuntu 24.04, kernel 6.1.x LTS; official Firecracker CI images, not 2021 quickstart images). See `.claude/CLAUDE.md` no-stale-images policy and incident record.
- No other explicitly banned tools documented in-repo; confirm with operator before adding new runtime dependencies.
