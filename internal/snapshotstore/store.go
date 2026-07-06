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
	for _, f := range files {
		if err := validateName(f.Name); err != nil {
			return nil, err
		}
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
	if err := os.MkdirAll(destDir, 0o750); err != nil {
		return nil, fmt.Errorf("snapshotstore: create restore dir: %w", err)
	}

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(s.parallelism)
	for _, fe := range manifest.Files {
		fe := fe
		g.Go(func() error { return s.getFile(gctx, destDir, fe) })
	}
	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("snapshotstore: restore snapshot %q: %w", id, err)
	}
	return manifest, nil
}

// getFile downloads, decompresses, verifies, and atomically writes one file.
func (s *TieredStore) getFile(ctx context.Context, destDir string, fe FileEntry) error {
	// Path-traversal guard: the manifest is untrusted input on restore.
	if err := validateName(fe.Name); err != nil {
		return err
	}
	rc, err := s.blob.Get(ctx, fe.Key)
	if err != nil {
		return fmt.Errorf("download %q: %w", fe.Key, err)
	}
	defer func() { _ = rc.Close() }()

	dec, err := zstd.NewReader(rc)
	if err != nil {
		return fmt.Errorf("create zstd reader for %q: %w", fe.Key, err)
	}
	defer dec.Close()

	// Bound decompressed output to one byte over the declared size so an
	// over-large (bomb / torn) file is detected rather than streamed to disk.
	limited := io.LimitReader(dec, fe.Size+1)
	hasher := sha256.New()

	tmp, err := os.CreateTemp(destDir, ".tmp-"+fe.Name+"-*")
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
		return fmt.Errorf("decompress %q: %w", fe.Key, err)
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
	if err := os.Rename(tmpName, filepath.Join(destDir, fe.Name)); err != nil {
		return fmt.Errorf("commit restore file %q: %w", fe.Name, err)
	}
	committed = true
	return nil
}
