# Feature Specification: Object-Storage Snapshot Tiering + Cross-Node Portability

**Issue:** daax-dev/nanofuse#130
**Status:** Foundational increment (write-path wired; restore primitive delivered; live-VM resume + real object-storage backend explicitly deferred)

## Why (Problem)

nanofuse takes real Firecracker VM snapshots, but writes them only to **local
filesystem paths**. A snapshot is therefore stranded on the node that produced
it: there is no durability if that node's disk is lost, and no portability to
resume the session on a different host. This blocks any workload-scheduling
scheme that needs a paused session to come back on whichever worker is free.

## What (Outcomes)

Snapshots become **portable, durable, and self-describing** without changing the
existing local-disk default behavior:

1. Snapshot artifacts (VM state file + memory file) can be moved to a durable
   object tier, compressed, in addition to the local copy.
2. Each tiered snapshot carries a **self-describing, version-pinned manifest**
   that lists every file with its size and content digest, plus the runtime
   versions required to reproduce the restore. The manifest is the **commit
   marker**: it is written only after every data file is durably stored, so a
   reader never observes a partially uploaded snapshot as valid.
3. A restore reader on **any node** can discover a tiered snapshot, read its
   manifest, download and decompress its files concurrently, and verify every
   file's integrity before the bytes are handed to a sandbox — using only the
   object tier (no dependency on the origin node's local database).
4. The object tier is reached through a **pluggable interface** so the concrete
   backend (a hermetic filesystem-backed object store now; S3 later) is a
   drop-in, and so the existing local-disk flow is never on the critical path
   for callers that do not opt in.

## Scope

**In scope (this increment)**
- A snapshot-store interface with `Put` / `Get` / `List` / `Manifest`, and a
  lower-level blob backend seam behind it.
- A version-pinned JSON manifest type (file list + sizes + SHA-256 digests +
  runtime versions), written last.
- zstd compression on upload, decompression on download, using the maintained
  `github.com/klauspost/compress/zstd` library already vendored.
- A filesystem-backed blob backend that exercises the entire interface,
  manifest, compression, and digest path hermetically (no network, no creds).
- Opt-in wiring: snapshot **create** additionally tiers to the store when a store
  is configured; local disk stays the default and is unaffected when it is not.
- A restore **primitive** (`Get`) that stages a tiered snapshot back to local
  disk with full integrity verification, including a cross-node retrieval
  simulation.

**Out of scope (explicitly deferred, with reason)**
- A real S3/GCS backend + credential/bucket wiring. Deferred: the value at risk
  is the interface and integrity semantics, which are fully exercised by the
  hermetic backend; the S3 backend is a mechanical `Blob` implementation.
- Live-VM resume of a tiered snapshot end-to-end. Deferred because
  resume-from-snapshot is **not implemented even for local disk** in the current
  codebase (`Manager.Resume` and the resume handler ignore the snapshot; there
  is no Firecracker `/snapshot/load` call). Wiring `Get` into a live resume is
  blocked on that pre-existing gap, which is a separate feature.
- Parallel *upload* fan-out across many files is bounded but the current
  snapshot has exactly two files; the design supports N via a concurrency limit.

## Success Criteria (measurable / verifiable)

- SC-1 A snapshot put through the store and retrieved into a *different*
  destination directory yields byte-identical files (verified by digest and by
  content comparison in tests).
- SC-2 The manifest records, for every file: logical name, role, object key,
  uncompressed size, compressed size, and SHA-256 digest — plus firecracker /
  kernel / nanofuse runtime versions.
- SC-3 Manifest-last is enforced: if any data file fails to store, no manifest
  is written and `Get`/`Manifest` report the snapshot as absent/incomplete
  (tested by injecting a failing backend).
- SC-4 Corruption is caught: a tampered stored file causes `Get` to fail with a
  digest-mismatch (or decompression) error and does not surface corrupt bytes.
- SC-5 A restore reader constructed against the same backend from a *separate*
  store instance (cross-node simulation) retrieves and verifies the snapshot
  with no access to the producer's local state.
- SC-6 A manifest whose schema version exceeds the supported version is rejected
  with a clear, actionable error (version-mismatch safety).
- SC-7 A manifest naming a file with a path-traversal component
  (`..`, separators, absolute) is rejected on restore; nothing is written
  outside the destination directory.
- SC-8 A file whose decompressed size exceeds its manifest-declared size is
  rejected (decompression-bomb / size guard), not written.
- SC-9 The existing local-disk snapshot create path is unchanged when no store
  is configured (default), and the full test suite plus `mage test` pass.
