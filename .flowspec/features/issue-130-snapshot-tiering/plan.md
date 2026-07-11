# Implementation Plan: Object-Storage Snapshot Tiering

**Issue:** daax-dev/nanofuse#130
**Spec:** ./spec.md
**Branch:** feat/issue-130-snapshot-tiering
**Module:** github.com/daax-dev/nanofuse (go 1.25)

## Architecture

Two layered seams keep the integrity/compression logic backend-agnostic and make
a real object store a drop-in:

```
handleCreateSnapshot ──opt-in──► snapshotstore.Store
                                     (TieredStore)
                                  ├─ zstd compress + sha256 + manifest-last
                                  └─ snapshotstore.Blob   ◄── swap point
                                        └─ FSBlob (hermetic, this increment)
                                        └─ S3Blob (DEFERRED)
```

### New package: `internal/snapshotstore`

**`Blob`** — low-level object backend (the S3 drop-in seam):
```go
type Blob interface {
    Put(ctx context.Context, key string, r io.Reader) error
    Get(ctx context.Context, key string) (io.ReadCloser, error)
    List(ctx context.Context, prefix string) ([]string, error)
    Exists(ctx context.Context, key string) (bool, error)
}
```
`FSBlob` implements it against a root directory with atomic temp-file+rename
writes and prefix-walk listing. `S3Blob` is a documented TODO.

**`Store`** — snapshot-level tiering API:
```go
type Store interface {
    Put(ctx context.Context, id string, files []SourceFile, rt RuntimeVersions) (*Manifest, error)
    Get(ctx context.Context, id, destDir string) (*Manifest, error)
    List(ctx context.Context) ([]string, error)
    Manifest(ctx context.Context, id string) (*Manifest, error)
}
type SourceFile struct { Name, Role, Path string } // Name = logical restore filename
```
`TieredStore{blob, level, parallelism}` is the sole implementation.

**Manifest (versioned, JSON, commit marker):**
```go
const ManifestSchemaVersion = 1
type Manifest struct {
    SchemaVersion int
    SnapshotID    string
    CreatedAt     time.Time
    Compression   string          // "zstd"
    Runtime       RuntimeVersions // firecracker/kernel/nanofuse/snapshot_api
    Files         []FileEntry     // name, role, key, size, compressed_size, digest(sha256)
}
```

### Behavior

- **Put**: for each `SourceFile`, stream source → sha256 (tee) → zstd encoder →
  `Blob.Put(id + "/" + name + ".zst")`, recording uncompressed size + compressed
  size + digest. Files upload concurrently, bounded by `parallelism`
  (`errgroup` with `SetLimit`). After **all** succeed, marshal the manifest and
  `Blob.Put(id + "/manifest.json")` — **last**. Any file error aborts before the
  manifest is written, so no partial snapshot is ever committed.
- **Get**: read manifest first (`ErrManifestNotFound` if absent → "incomplete or
  absent"); reject `SchemaVersion > ManifestSchemaVersion`
  (`ErrUnsupportedManifestVersion`); for each entry validate the logical name
  against traversal, download → zstd-decode → `io.LimitReader(size+1)` and reject
  overruns (`ErrSizeMismatch`), verify sha256 (`ErrDigestMismatch`), write
  atomically (temp+rename) into `destDir`. Files download concurrently.
- **List**: enumerate keys ending in `/manifest.json`; only committed snapshots
  appear. **Manifest**: fetch+decode `id/manifest.json`.

### Safety guards
- Path traversal: `validateName` rejects empty, `/`, `\`, `..`, absolute — restore
  filename must be a plain base name.
- Decompression bomb: bounded `LimitReader` + exact declared-size check.
- Torn write: atomic temp+rename in both `FSBlob.Put` and restore output.
- Version pinning: firecracker version probed from the binary
  (`firecracker.BinaryVersion` → runs `--version`, parses); nanofuse from the
  build version string; recorded in the manifest.

### Wiring (opt-in, additive, non-regressing)
- `config.SnapshotStoreConfig{ Backend, Path, Compression, Parallelism }` added to
  `Config`; `DefaultConfig` = disabled (`Backend: ""`); `Validate` requires
  `Path` when `Backend == "filesystem"`.
- `cmd/nanofused` constructs the store best-effort (mirrors the existing
  `recordingStorage` init) and sets a new `Server.snapshotStore` field (nil when
  disabled).
- `handleCreateSnapshot`: after the local snapshot + DB record succeed, if
  `snapshotStore != nil`, tier `vm.snap` + `mem.snap` with probed runtime
  versions. Tiering is logged; a tiering failure is surfaced via log + a
  `Tiered`/`TierError` note without discarding the already-valid local snapshot.
  When no store is configured the code path is identical to today.
- `firecracker.BinaryVersion(path)` helper added for manifest version-pinning.

### New dependency
- `golang.org/x/sync/errgroup` — already in `go.sum` (indirect); promoted to direct.
- No AWS SDK added (keeps the tree light; S3 backend deferred).

## Test plan (bulk of value — all hermetic, no network/creds)
1. `FSBlob` put/get/list/exists/overwrite/missing-key.
2. zstd round-trip correctness (via Store put/get byte-identity).
3. `Store` Put→Manifest→Get round-trip; digests + byte-identity (SC-1, SC-2).
4. Manifest-last: failing `Blob` on a data file ⇒ no manifest; `Get` ⇒
   `ErrManifestNotFound` (SC-3).
5. Corrupt stored file ⇒ digest/decompress error (SC-4).
6. Cross-node simulation: Put via store A, Get via store B on same backend root,
   different destDir (SC-5).
7. Unsupported manifest schema version ⇒ error (SC-6).
8. Path-traversal name in manifest ⇒ rejected, nothing escapes destDir (SC-7).
9. Size-mismatch / decompression-bomb guard (SC-8).
10. `List` skips partial (manifest-less) snapshots.
11. Config defaults disabled; filesystem backend requires path.
12. `firecracker.BinaryVersion` parses a stub binary; missing binary errors.
13. Handler wiring: `handleCreateSnapshot` with a store tiers both files
    (manifest present in blob root) using a fake runtime manager + FSBlob.

## Verification gates
`go build ./...`, `go vet ./...`, `gofmt -l`, `go test -race ./...`, `mage test`,
golangci-lint (watch cyclomatic complexity — keep Put/Get helpers small).

## Explicitly deferred (documented in PR)
- Real S3/GCS `Blob` + credential/bucket wiring.
- Live-VM resume of a tiered snapshot (blocked on missing local
  resume-from-snapshot: no Firecracker `/snapshot/load` exists yet).
- Live-Firecracker tiering e2e in a KVM sandbox.
