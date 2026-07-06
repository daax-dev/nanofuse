package snapshotstore

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ErrObjectNotFound is returned by a Blob when a key does not exist.
var ErrObjectNotFound = errors.New("snapshotstore: object not found")

// Blob is the low-level object backend behind a Store. It is a small, generic
// key/value blob API so that a concrete object store (an S3/GCS implementation)
// is a drop-in replacement for the local filesystem backend used here.
//
// Keys use forward slashes as separators regardless of backend.
type Blob interface {
	// Put stores r under key, overwriting any existing object atomically.
	Put(ctx context.Context, key string, r io.Reader) error
	// Get opens the object at key for reading. It returns ErrObjectNotFound if
	// the key does not exist. The caller must Close the returned reader.
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	// List returns all keys with the given prefix (no wildcard semantics).
	List(ctx context.Context, prefix string) ([]string, error)
	// Exists reports whether key exists.
	Exists(ctx context.Context, key string) (bool, error)
}

// FSBlob is a filesystem-backed Blob. It stores each object as a file under a
// root directory, using the object key as a relative path. Writes are atomic
// (temp file + rename) so an interrupted Put never leaves a torn object visible.
//
// FSBlob exercises the entire Store/manifest/compression/digest path without a
// network or credentials, and doubles as a "local object store" tier. The real
// S3/GCS backend is a separate Blob implementation (deferred, see plan.md).
type FSBlob struct {
	root string
}

// NewFSBlob creates a filesystem blob rooted at dir, creating it if needed.
func NewFSBlob(dir string) (*FSBlob, error) {
	if dir == "" {
		return nil, fmt.Errorf("snapshotstore: filesystem blob root is required")
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("snapshotstore: resolve blob root: %w", err)
	}
	if err := os.MkdirAll(abs, 0o750); err != nil {
		return nil, fmt.Errorf("snapshotstore: create blob root: %w", err)
	}
	return &FSBlob{root: abs}, nil
}

// keyPath maps an object key to an on-disk path, rejecting any key that would
// escape the root (defense in depth; Store already validates ids and names).
func (b *FSBlob) keyPath(key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("snapshotstore: empty object key")
	}
	clean := filepath.Clean("/" + filepath.FromSlash(key))
	full := filepath.Join(b.root, clean)
	rel, err := filepath.Rel(b.root, full)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("snapshotstore: object key %q escapes blob root", key)
	}
	return full, nil
}

// Put implements Blob.
func (b *FSBlob) Put(ctx context.Context, key string, r io.Reader) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	path, err := b.keyPath(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("snapshotstore: create object dir: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-"+filepath.Base(path)+"-*")
	if err != nil {
		return fmt.Errorf("snapshotstore: create temp object: %w", err)
	}
	tmpName := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = tmp.Close()
			_ = os.Remove(tmpName)
		}
	}()
	if _, err := io.Copy(tmp, r); err != nil {
		return fmt.Errorf("snapshotstore: write object %q: %w", key, err)
	}
	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("snapshotstore: sync object %q: %w", key, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("snapshotstore: close object %q: %w", key, err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("snapshotstore: commit object %q: %w", key, err)
	}
	cleanup = false
	return nil
}

// Get implements Blob.
func (b *FSBlob) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	path, err := b.keyPath(key)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path) // #nosec G304 -- path is confined to blob root by keyPath.
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("%w: %s", ErrObjectNotFound, key)
		}
		return nil, fmt.Errorf("snapshotstore: open object %q: %w", key, err)
	}
	return f, nil
}

// List implements Blob.
func (b *FSBlob) List(ctx context.Context, prefix string) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var keys []string
	walkErr := filepath.WalkDir(b.root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		// Skip in-flight temp files from interrupted Puts.
		if strings.HasPrefix(d.Name(), ".tmp-") {
			return nil
		}
		rel, err := filepath.Rel(b.root, path)
		if err != nil {
			return err
		}
		key := filepath.ToSlash(rel)
		if strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
		return nil
	})
	if walkErr != nil {
		if errors.Is(walkErr, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("snapshotstore: list objects: %w", walkErr)
	}
	return keys, nil
}

// Exists implements Blob.
func (b *FSBlob) Exists(ctx context.Context, key string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	path, err := b.keyPath(key)
	if err != nil {
		return false, err
	}
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("snapshotstore: stat object %q: %w", key, err)
	}
	return true, nil
}
