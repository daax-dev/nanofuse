# Plan: gondolin-style sandbox conversion (issue #18, reframed)

## Approach

- New package `internal/gondolin`:
  - `Sandbox` struct mirroring gondolin's flag surface; `Parse` with
    `yaml.Decoder.KnownFields(true)`, empty-input and multi-document guards.
  - `Convert(*Sandbox, Options) (*client.CreateVMRequest, []Divergence, error)`:
    maps faithful fields, builds an egress policy for `allow_host`/`dns`, and
    emits one blocking divergence per unrepresentable feature.
  - Severity model: `info` (disclosed defaults), `warn` (safe degrade / lossy
    drop), `blocking` (fails closed unless `AllowLossy`).
  - `RenderSpecYAML` renders the nanofuse spec via presentation structs with
    stable YAML keys (deterministic for golden tests).
- New cobra command `nanofuse convert gondolin <file> [--allow-lossy]
  [--resolve-egress] [-o out]`; wired into `cmd/nanofuse/main.go` and added to
  the `PersistentPreRunE` skip set (local-only, no API client).

## Mapping decisions

- `image` -> `Image` (clean; required).
- `resources` -> `VCPUs`/`MemoryMiB`; absent -> nanofuse defaults (2 / 512),
  disclosed as an `info` divergence (gondolin has no resource model).
- `allow_host` (L7) -> default-drop `EgressPolicy` + `warn` (safe degrade).
  Opt-in `--resolve-egress` resolves literal hostnames to /32 TCP/443 rules via
  an injectable resolver; wildcards/paths remain dropped and reported.
- `vmm`, `cwd`, `env`, `host_secret`, `mount_hostfs`, `mount_memfs`,
  `ssh_allow_host`, `tcp_map`, `dns`, `rootfs_size` -> no faithful equivalent ->
  `blocking`. Fail closed by default; `--allow-lossy` downgrades to `warn` and
  drops loudly. Never silently translated (avoids false equivalence / false
  security).

## Testing

Table-driven tests in `internal/gondolin`: clean fields, defaulted/partial
resources, allow-host drop-and-warn, resolve-egress with fake resolver, each
unrepresentable feature (fail-closed + allow-lossy), no-silent-drop coverage,
determinism, golden rendered spec, and adversarial parse inputs (empty,
unknown key, wrong type, multi-document, malformed).

## Validation

- `go build ./...`, `go vet`, `gofmt`, `go test -race ./...`, `mage test`.
- Adversarial cross-model review (gemini) on the diff.
- No end-to-end run: no gondolin runtime and no nanofuse spec-file importer.
