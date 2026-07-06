package snapshotstore

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/klauspost/compress/zstd"
	"golang.org/x/sync/errgroup"
)

// maxManifestBytes bounds how much of a manifest object is read, guarding against
// a hostile or corrupt manifest blob.
const maxManifestBytes = 8 << 20 // 8 MiB

// defaultParallelism bounds concurrent per-file uploads/downloads.
const defaultParallelism = 4

// Store is the snapshot-level tiering API. Implementations move a snapshot's
// files to a durable, portable tier and back, attaching a version-pinned
// manifest that is written last as a commit marker.
type Store interface {
	// Put compresses and stores the given files under id, then writes the
	// manifest last. It returns the manifest on success. If any file fails, no
	// manifest is written (the snapshot is never partially committed).
	Put(ctx context.Context, id string, files []SourceFile, rt RuntimeVersions) (*Manifest, error)
	// Get downloads, decompresses, and integrity-verifies every file of snapshot
	// id into destDir, returning the manifest. Files that fail verification are
	// not written.
	Get(ctx context.Context, id, destDir string) (*Manifest, error)
	// List returns the ids of committed snapshots (those with a manifest).
	List(ctx context.Context) ([]string, error)
	// Manifest returns the manifest for id without downloading data files.
	Manifest(ctx context.Context, id string) (*Manifest, error)
}

// Options configures a TieredStore.
type Options struct {
	// Parallelism bounds concurrent per-file transfers (default 4).
	Parallelism int
	// Level is the zstd encoder level (default SpeedBetterCompression).
	Level zstd.EncoderLevel
}

// TieredStore is the Store implementation: it layers zstd compression, SHA-256
// integrity, and manifest-last semantics over any Blob backend.
type TieredStore struct {
	blob        Blob
	parallelism int
	level       zstd.EncoderLevel
}

// NewTieredStore builds a TieredStore over blob.
func NewTieredStore(blob Blob, opts Options) *TieredStore {
	p := opts.Parallelism
	if p <= 0 {
		p = defaultParallelism
	}
	level := opts.Level
	if level == 0 {
		level = zstd.SpeedBetterCompression
	}
	return &TieredStore{blob: blob, parallelism: p, level: level}
}

// compile-time check.
var _ Store = (*TieredStore)(nil)

func objectKey(id, name string) string { return id + "/" + name }

// Put implements Store.
func (s *TieredStore) Put(ctx context.Context, id string, files []SourceFile, rt RuntimeVersions) (*Manifest, error) {
	if err := validateSnapshotID(id); err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("snapshotstore: no files to store for snapshot %q", id)
	}
	seenNames := make(map[string]struct{}, len(files))
	for _, f := range files {
		if err := validateName(f.Name); err != nil {
			return nil, err
		}
		// Names map 1:1 to backend object keys; duplicates would race and
		// overwrite the same object and corrupt the manifest.
		if _, dup := seenNames[f.Name]; dup {
			return nil, fmt.Errorf("%w: %q in snapshot %q", ErrDuplicateFileName, f.Name, id)
		}
		seenNames[f.Name] = struct{}{}
	}

	entries := make([]FileEntry, len(files))
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(s.parallelism)
	for i, f := range files {
		i, f := i, f
		g.Go(func() error {
			fe, err := s.putFile(gctx, id, f)
			if err != nil {
				return err
			}
			entries[i] = fe
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		// No manifest is written: the snapshot stays uncommitted (invisible to
		// List/Get) even though some data objects may have landed.
		return nil, fmt.Errorf("snapshotstore: tier snapshot %q: %w", id, err)
	}

	manifest := &Manifest{
		SchemaVersion: ManifestSchemaVersion,
		SnapshotID:    id,
		CreatedAt:     time.Now().UTC(),
		Compression:   CompressionZstd,
		Runtime:       rt,
		Files:         entries,
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("snapshotstore: marshal manifest for %q: %w", id, err)
	}
	// Commit marker: the manifest is written strictly last.
	if err := s.blob.Put(ctx, objectKey(id, manifestObject), bytes.NewReader(data)); err != nil {
		return nil, fmt.Errorf("snapshotstore: write manifest for %q: %w", id, err)
	}
	return manifest, nil
}

// countingWriter counts bytes written to the wrapped writer.
type countingWriter struct {
	w io.Writer
	n int64
}

func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	c.n += int64(n)
	return n, err
}

// putFile streams one source file through SHA-256 (uncompressed) and zstd into
// the blob, returning its manifest entry.
func (s *TieredStore) putFile(ctx context.Context, id string, sf SourceFile) (FileEntry, error) {
	src, err := os.Open(sf.Path) // #nosec G304 -- caller-supplied local snapshot path.
	if err != nil {
		return FileEntry{}, fmt.Errorf("open source %q: %w", sf.Path, err)
	}
	defer func() { _ = src.Close() }()

	key := objectKey(id, sf.Name+".zst")
	hasher := sha256.New()
	pr, pw := io.Pipe()
	counter := &countingWriter{w: pw}
	enc, err := zstd.NewWriter(counter, zstd.WithEncoderLevel(s.level))
	if err != nil {
		// The copy goroutine below is not started yet, so close the pipe ends we
		// already created rather than leak them on this early return.
		_ = pw.Close()
		_ = pr.Close()
		return FileEntry{}, fmt.Errorf("create zstd encoder: %w", err)
	}

	var uncompressed int64
	go func() {
		n, copyErr := io.Copy(io.MultiWriter(hasher, enc), src)
		uncompressed = n
		if closeErr := enc.Close(); copyErr == nil {
			copyErr = closeErr
		}
		// Propagate any error to the blob reader so Put unblocks and fails.
		_ = pw.CloseWithError(copyErr)
	}()

	if err := s.blob.Put(ctx, key, pr); err != nil {
		// Unblock the writer goroutine if it is still copying.
		_ = pr.CloseWithError(err)
		return FileEntry{}, fmt.Errorf("upload %q: %w", key, err)
	}

	return FileEntry{
		Name:           sf.Name,
		Role:           sf.Role,
		Key:            key,
		Size:           uncompressed,
		CompressedSize: counter.n,
		Digest:         hex.EncodeToString(hasher.Sum(nil)),
	}, nil
}

// Manifest implements Store.
func (s *TieredStore) Manifest(ctx context.Context, id string) (*Manifest, error) {
	if err := validateSnapshotID(id); err != nil {
		return nil, err
	}
	rc, err := s.blob.Get(ctx, objectKey(id, manifestObject))
	if err != nil {
		if errors.Is(err, ErrObjectNotFound) {
			return nil, fmt.Errorf("%w: %s", ErrManifestNotFound, id)
		}
		return nil, fmt.Errorf("snapshotstore: read manifest for %q: %w", id, err)
	}
	defer func() { _ = rc.Close() }()

	data, err := io.ReadAll(io.LimitReader(rc, maxManifestBytes))
	if err != nil {
		return nil, fmt.Errorf("snapshotstore: read manifest for %q: %w", id, err)
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("snapshotstore: parse manifest for %q: %w", id, err)
	}
	if m.SchemaVersion > ManifestSchemaVersion {
		return nil, fmt.Errorf("%w: manifest %q is v%d, this build supports up to v%d",
			ErrUnsupportedManifestVersion, id, m.SchemaVersion, ManifestSchemaVersion)
	}
	// The manifest self-identifies which snapshot it describes; reject it if that
	// does not match the id it was fetched under. A mismatch means the object was
	// corrupted, swapped, or tampered with, so its pinned digests/sizes cannot be
	// trusted to describe the requested snapshot.
	if m.SnapshotID != id {
		return nil, fmt.Errorf("%w: requested %q but manifest reports %q",
			ErrManifestIDMismatch, id, m.SnapshotID)
	}
	return &m, nil
}

// List implements Store.
func (s *TieredStore) List(ctx context.Context) ([]string, error) {
	keys, err := s.blob.List(ctx, "")
	if err != nil {
		return nil, err
	}
	suffix := "/" + manifestObject
	ids := make([]string, 0, len(keys))
	for _, k := range keys {
		if !strings.HasSuffix(k, suffix) {
			continue
		}
		id := strings.TrimSuffix(k, suffix)
		if id == "" || strings.Contains(id, "/") {
			continue // only top-level snapshot prefixes
		}
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids, nil
}

// Get implements Store.
func (s *TieredStore) Get(ctx context.Context, id, destDir string) (*Manifest, error) {
	if err := validateSnapshotID(id); err != nil {
		return nil, err
	}
	if destDir == "" {
		return nil, fmt.Errorf("snapshotstore: destination directory is required")
	}
	manifest, err := s.Manifest(ctx, id)
	if err != nil {
		return nil, err
	}
	// The manifest is untrusted input on restore. Reject a tampered file list
	// (duplicate names, negative sizes, unsafe names) up front, before any
	// concurrent download acts on it.
	if err := validateManifestFiles(manifest.Files); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(destDir, 0o750); err != nil {
		return nil, fmt.Errorf("snapshotstore: create restore dir: %w", err)
	}

	// Restore into a private staging dir and promote the fully-verified set into
	// destDir only after every file succeeds. A mid-restore failure then leaves
	// destDir untouched instead of partially populated. staging is a subdir of
	// destDir, so the final promotion renames stay on one filesystem.
	staging, err := os.MkdirTemp(destDir, ".restore-*")
	if err != nil {
		return nil, fmt.Errorf("snapshotstore: create restore staging dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(staging) }()

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(s.parallelism)
	for _, fe := range manifest.Files {
		fe := fe
		g.Go(func() error { return s.getFile(gctx, id, staging, fe) })
	}
	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("snapshotstore: restore snapshot %q: %w", id, err)
	}

	if err := promoteRestored(staging, destDir, manifest.Files); err != nil {
		return nil, fmt.Errorf("snapshotstore: commit restored snapshot %q: %w", id, err)
	}
	return manifest, nil
}

// promoteRestored moves every verified file from the staging dir into destDir.
// It runs only after all files have been downloaded and integrity-checked, so
// destDir goes from empty to fully populated with a series of same-filesystem
// renames. If a rename fails partway (an exceptional local FS fault), the files
// already promoted are removed so destDir is not left holding a partial
// snapshot. Names are already validated, keeping the joins within destDir.
func promoteRestored(staging, destDir string, files []FileEntry) error {
	for i, fe := range files {
		from := filepath.Join(staging, fe.Name)
		to := filepath.Join(destDir, fe.Name)
		if err := os.Rename(from, to); err != nil {
			// Roll back the already-promoted files so a partial promotion does not
			// leave destDir holding an incomplete snapshot.
			for _, done := range files[:i] {
				_ = os.Remove(filepath.Join(destDir, done.Name))
			}
			return fmt.Errorf("commit restored file %q: %w", fe.Name, err)
		}
	}
	return nil
}

// getFile downloads, decompresses, verifies, and atomically writes one file into
// dir (the restore staging dir). The caller promotes dir's contents into the
// real destination only after every file succeeds.
func (s *TieredStore) getFile(ctx context.Context, id, dir string, fe FileEntry) error {
	// Path-traversal guard: the manifest is untrusted input on restore.
	if err := validateName(fe.Name); err != nil {
		return err
	}
	// Derive the object key deterministically from the snapshot id and the
	// validated file name rather than trusting the manifest-provided Key: a
	// tampered manifest must not be able to point the download at an arbitrary
	// backend object.
	key := objectKey(id, fe.Name+".zst")
	rc, err := s.blob.Get(ctx, key)
	if err != nil {
		return fmt.Errorf("download %q: %w", key, err)
	}
	defer func() { _ = rc.Close() }()

	dec, err := zstd.NewReader(rc)
	if err != nil {
		return fmt.Errorf("create zstd reader for %q: %w", key, err)
	}
	defer dec.Close()

	// Bound decompressed output to one byte over the declared size so an
	// over-large (bomb / torn) file is detected rather than streamed to disk.
	limited := io.LimitReader(dec, fe.Size+1)
	hasher := sha256.New()

	tmp, err := os.CreateTemp(dir, ".tmp-"+fe.Name+"-*")
	if err != nil {
		return fmt.Errorf("create temp restore file: %w", err)
	}
	tmpName := tmp.Name()
	committed := false
	defer func() {
		if !committed {
			_ = tmp.Close()
			_ = os.Remove(tmpName)
		}
	}()

	n, err := io.Copy(io.MultiWriter(tmp, hasher), limited)
	if err != nil {
		return fmt.Errorf("decompress %q: %w", key, err)
	}
	if n != fe.Size {
		return fmt.Errorf("%w: %q expected %d bytes, got %d", ErrSizeMismatch, fe.Name, fe.Size, n)
	}
	if got := hex.EncodeToString(hasher.Sum(nil)); got != fe.Digest {
		return fmt.Errorf("%w: %q expected %s, got %s", ErrDigestMismatch, fe.Name, fe.Digest, got)
	}
	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("sync restore file %q: %w", fe.Name, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close restore file %q: %w", fe.Name, err)
	}
	if err := os.Rename(tmpName, filepath.Join(dir, fe.Name)); err != nil {
		return fmt.Errorf("commit restore file %q: %w", fe.Name, err)
	}
	committed = true
	return nil
}
