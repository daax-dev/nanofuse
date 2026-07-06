// Package snapshotstore provides a pluggable, portable tier for Firecracker VM
// snapshots. A snapshot's data files are compressed and stored behind a Blob
// backend together with a version-pinned, self-describing manifest that is
// written last as a commit marker. The manifest lets any node discover, verify,
// and restore a snapshot using only the object tier — the basis for
// cross-node portability and durability (issue #130).
package snapshotstore

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ManifestSchemaVersion is the manifest format version understood by this build.
// A reader refuses a manifest whose SchemaVersion is greater than this, so that
// snapshots produced by a newer nanofuse are never silently mis-restored.
const ManifestSchemaVersion = 1

// manifestObject is the reserved object name (per snapshot prefix) that holds the
// manifest. It is written last; its presence marks a snapshot as committed.
const manifestObject = "manifest.json"

// CompressionZstd is the only compression codec currently emitted.
const CompressionZstd = "zstd"

// Errors surfaced by a Store. They are sentinels so callers can branch on the
// failure class (missing vs. corrupt vs. version) rather than string-matching.
var (
	// ErrManifestNotFound means no committed manifest exists for the snapshot id.
	// Because the manifest is written last, this also covers the "upload was
	// interrupted / partial" case: an incomplete snapshot has no manifest.
	ErrManifestNotFound = errors.New("snapshotstore: snapshot manifest not found (absent or incomplete)")

	// ErrUnsupportedManifestVersion means the manifest was written by a newer
	// format than this build understands.
	ErrUnsupportedManifestVersion = errors.New("snapshotstore: unsupported manifest schema version")

	// ErrDigestMismatch means a downloaded file's content digest did not match
	// the digest pinned in the manifest.
	ErrDigestMismatch = errors.New("snapshotstore: file digest mismatch")

	// ErrSizeMismatch means a downloaded file's decompressed size did not match
	// the size pinned in the manifest (torn write or decompression bomb).
	ErrSizeMismatch = errors.New("snapshotstore: file size mismatch")

	// ErrUnsafeName means a manifest file name is not a safe base name and would
	// escape the restore destination (path traversal).
	ErrUnsafeName = errors.New("snapshotstore: unsafe file name in manifest")

	// ErrInvalidSnapshotID means a snapshot id is empty or not a safe segment.
	ErrInvalidSnapshotID = errors.New("snapshotstore: invalid snapshot id")
)

// RuntimeVersions pins the runtime binaries a restore must reproduce. Recorded
// at snapshot time so a restore on another node can validate/select the exact
// runtime that produced the memory + VM state.
type RuntimeVersions struct {
	Firecracker string `json:"firecracker,omitempty"`
	Kernel      string `json:"kernel,omitempty"`
	Nanofuse    string `json:"nanofuse,omitempty"`
	// SnapshotAPI is the Firecracker snapshot format version (e.g. the value sent
	// on /snapshot/create), which governs cross-version restore compatibility.
	SnapshotAPI string `json:"snapshot_api,omitempty"`
}

// FileEntry describes one stored snapshot file.
type FileEntry struct {
	// Name is the logical file name and the exact base name written on restore.
	// It must be a safe base name (validated) — never a path.
	Name string `json:"name"`
	// Role is a semantic tag, e.g. "vmstate" or "memory".
	Role string `json:"role"`
	// Key is the backend object key the compressed bytes live under.
	Key string `json:"key"`
	// Size is the uncompressed size in bytes.
	Size int64 `json:"size"`
	// CompressedSize is the stored (compressed) size in bytes.
	CompressedSize int64 `json:"compressed_size"`
	// Digest is the lowercase hex SHA-256 of the uncompressed content.
	Digest string `json:"digest"`
}

// Manifest is the self-describing, version-pinned description of a snapshot.
type Manifest struct {
	SchemaVersion int             `json:"schema_version"`
	SnapshotID    string          `json:"snapshot_id"`
	CreatedAt     time.Time       `json:"created_at"`
	Compression   string          `json:"compression"`
	Runtime       RuntimeVersions `json:"runtime"`
	Files         []FileEntry     `json:"files"`
}

// SourceFile is a local file to be tiered.
type SourceFile struct {
	// Name is the logical name; it becomes the restored file's base name.
	Name string
	// Role is a semantic tag, e.g. "vmstate" or "memory".
	Role string
	// Path is the local filesystem path to read from.
	Path string
}

// validateName ensures a manifest/source file name is a plain, safe base name
// that cannot escape a destination directory on restore.
func validateName(name string) error {
	if name == "" {
		return fmt.Errorf("%w: empty", ErrUnsafeName)
	}
	if name == "." || name == ".." {
		return fmt.Errorf("%w: %q", ErrUnsafeName, name)
	}
	if strings.ContainsAny(name, "/\\") {
		return fmt.Errorf("%w: %q contains a path separator", ErrUnsafeName, name)
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("%w: %q contains %q", ErrUnsafeName, name, "..")
	}
	// Reject the reserved manifest object name to avoid a file colliding with the
	// commit marker.
	if name == manifestObject {
		return fmt.Errorf("%w: %q is reserved", ErrUnsafeName, name)
	}
	return nil
}

// validateSnapshotID ensures the snapshot id is a safe single path segment used
// as the object prefix.
func validateSnapshotID(id string) error {
	if id == "" {
		return fmt.Errorf("%w: empty", ErrInvalidSnapshotID)
	}
	if id == "." || id == ".." {
		return fmt.Errorf("%w: %q", ErrInvalidSnapshotID, id)
	}
	if strings.ContainsAny(id, "/\\") {
		return fmt.Errorf("%w: %q contains a path separator", ErrInvalidSnapshotID, id)
	}
	if strings.Contains(id, "..") {
		return fmt.Errorf("%w: %q contains %q", ErrInvalidSnapshotID, id, "..")
	}
	return nil
}
