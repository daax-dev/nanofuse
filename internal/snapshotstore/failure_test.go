package snapshotstore

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBlobContextCancelled(t *testing.T) {
	blob, _ := NewFSBlob(t.TempDir())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := blob.Put(ctx, "k", bytes.NewReader([]byte("x"))); !errors.Is(err, context.Canceled) {
		t.Errorf("Put cancelled = %v, want context.Canceled", err)
	}
	if _, err := blob.Get(ctx, "k"); !errors.Is(err, context.Canceled) {
		t.Errorf("Get cancelled = %v, want context.Canceled", err)
	}
	if _, err := blob.List(ctx, ""); !errors.Is(err, context.Canceled) {
		t.Errorf("List cancelled = %v, want context.Canceled", err)
	}
	if _, err := blob.Exists(ctx, "k"); !errors.Is(err, context.Canceled) {
		t.Errorf("Exists cancelled = %v, want context.Canceled", err)
	}
}

func TestPutMissingSourceFile(t *testing.T) {
	store, _ := newStore(t)
	files := []SourceFile{{Name: "vm.snap", Role: "vmstate", Path: filepath.Join(t.TempDir(), "absent")}}
	if _, err := store.Put(context.Background(), "snap-x", files, RuntimeVersions{}); err == nil {
		t.Fatal("Put with missing source file should fail")
	}
}

func TestManifestCorruptJSON(t *testing.T) {
	store, root := newStore(t)
	id := "snap-badjson"
	if err := os.MkdirAll(filepath.Join(root, id), 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, id, manifestObject), []byte("{not json"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := store.Manifest(context.Background(), id); err == nil {
		t.Fatal("Manifest on corrupt JSON should fail")
	}
}

func TestGetMissingDataObject(t *testing.T) {
	store, root := newStore(t)
	ctx := context.Background()

	// A committed manifest that references a data object which does not exist.
	m := Manifest{
		SchemaVersion: ManifestSchemaVersion,
		SnapshotID:    "snap-gap",
		Compression:   CompressionZstd,
		Files:         []FileEntry{{Name: "vm.snap", Key: "snap-gap/vm.snap.zst", Size: 4, Digest: "00"}},
	}
	data, _ := json.Marshal(m)
	if err := os.MkdirAll(filepath.Join(root, "snap-gap"), 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "snap-gap", manifestObject), data, 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	if _, err := store.Get(ctx, "snap-gap", t.TempDir()); err == nil {
		t.Fatal("Get with missing data object should fail")
	}
}

// writeManifest writes m directly to the backend under id, bypassing Put so a
// hostile/tampered manifest can be crafted for restore-side tests.
func writeManifest(t *testing.T, root, id string, m Manifest) {
	t.Helper()
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, id), 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, id, manifestObject), data, 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
}

func TestGetRejectsDuplicateManifestNames(t *testing.T) {
	// An untrusted manifest naming the same file twice must be rejected up front:
	// concurrent restores would otherwise race to write the same destination file.
	store, root := newStore(t)
	id := "snap-dupman"
	writeManifest(t, root, id, Manifest{
		SchemaVersion: ManifestSchemaVersion,
		SnapshotID:    id,
		Compression:   CompressionZstd,
		Files: []FileEntry{
			{Name: "vm.snap", Key: id + "/vm.snap.zst", Size: 4, Digest: "00"},
			{Name: "vm.snap", Key: id + "/vm.snap.zst", Size: 4, Digest: "00"},
		},
	})
	if _, err := store.Get(context.Background(), id, t.TempDir()); !errors.Is(err, ErrDuplicateFileName) {
		t.Fatalf("Get duplicate-name manifest = %v, want ErrDuplicateFileName", err)
	}
}

func TestGetRejectsNegativeSize(t *testing.T) {
	// A negative declared Size is nonsensical and defeats the LimitReader bomb
	// guard, so restore must refuse it before downloading anything.
	store, root := newStore(t)
	id := "snap-negsize"
	writeManifest(t, root, id, Manifest{
		SchemaVersion: ManifestSchemaVersion,
		SnapshotID:    id,
		Compression:   CompressionZstd,
		Files:         []FileEntry{{Name: "vm.snap", Key: id + "/vm.snap.zst", Size: -1, Digest: "00"}},
	})
	if _, err := store.Get(context.Background(), id, t.TempDir()); !errors.Is(err, ErrNegativeFileSize) {
		t.Fatalf("Get negative-size manifest = %v, want ErrNegativeFileSize", err)
	}
}

func TestGetPartialRestoreLeavesDestClean(t *testing.T) {
	// When one file of a multi-file snapshot fails mid-restore, destDir must be
	// left empty: the already-verified sibling is staged, not committed, so a
	// caller reusing destDir never observes a partial snapshot.
	store, root := newStore(t)
	ctx := context.Background()
	src := t.TempDir()

	id := "snap-partial-restore"
	goodData := bytes.Repeat([]byte("good-content-"), 4096)
	badData := bytes.Repeat([]byte("bad-content-"), 4096)
	files := []SourceFile{
		{Name: "good.snap", Role: "vmstate", Path: writeFile(t, src, "good.snap", goodData)},
		{Name: "bad.snap", Role: "memory", Path: writeFile(t, src, "bad.snap", badData)},
	}
	if _, err := store.Put(ctx, id, files, RuntimeVersions{}); err != nil {
		t.Fatalf("Put: %v", err)
	}

	// Corrupt one stored object so its restore fails integrity verification.
	badObj := filepath.Join(root, id, "bad.snap.zst")
	raw, err := os.ReadFile(badObj)
	if err != nil {
		t.Fatalf("read object: %v", err)
	}
	raw[len(raw)/2] ^= 0xFF
	if err := os.WriteFile(badObj, raw, 0o600); err != nil {
		t.Fatalf("corrupt object: %v", err)
	}

	dest := t.TempDir()
	if _, err := store.Get(ctx, id, dest); err == nil {
		t.Fatal("expected Get to fail when a file fails restore")
	}
	// Neither the failed file nor its verified sibling may be left in destDir, and
	// no staging dir may remain.
	entries, err := os.ReadDir(dest)
	if err != nil {
		t.Fatalf("read dest: %v", err)
	}
	if len(entries) != 0 {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Fatalf("destDir not clean after failed restore: %v", names)
	}
}

func TestGetShortFileFailsSizeCheck(t *testing.T) {
	store, root := newStore(t)
	ctx := context.Background()

	id := "snap-short"
	data := []byte("exactly-this")
	files := []SourceFile{{Name: "vm.snap", Role: "vmstate", Path: writeFile(t, t.TempDir(), "vm.snap", data)}}
	m, err := store.Put(ctx, id, files, RuntimeVersions{})
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
	// Overstate the declared size: the decompressed stream is shorter than
	// declared, which must be rejected as a size mismatch.
	m.Files[0].Size = int64(len(data)) + 100
	out, _ := json.Marshal(m)
	if err := os.WriteFile(filepath.Join(root, id, manifestObject), out, 0o600); err != nil {
		t.Fatalf("rewrite manifest: %v", err)
	}
	if _, err := store.Get(ctx, id, t.TempDir()); !errors.Is(err, ErrSizeMismatch) {
		t.Fatalf("Get short = %v, want ErrSizeMismatch", err)
	}
}

func TestGetRejectsTooLargeSize(t *testing.T) {
	// An absurd Size (near math.MaxInt64) is rejected up front so downstream
	// Size+1 arithmetic cannot overflow on a tampered manifest.
	store, root := newStore(t)
	id := "snap-huge"
	writeManifest(t, root, id, Manifest{
		SchemaVersion: ManifestSchemaVersion,
		SnapshotID:    id,
		Compression:   CompressionZstd,
		Files:         []FileEntry{{Name: "vm.snap", Key: id + "/vm.snap.zst", Size: math.MaxInt64, Digest: "00"}},
	})
	if _, err := store.Get(context.Background(), id, t.TempDir()); !errors.Is(err, ErrFileSizeTooLarge) {
		t.Fatalf("Get huge-size manifest = %v, want ErrFileSizeTooLarge", err)
	}
}

func TestPromoteRestoredRefusesToClobber(t *testing.T) {
	// A restore must not overwrite a pre-existing file in destDir (rollback could
	// then delete a file we did not create).
	staging := t.TempDir()
	dest := t.TempDir()
	if err := os.WriteFile(filepath.Join(staging, "vm.snap"), []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dest, "vm.snap"), []byte("preexisting"), 0o600); err != nil {
		t.Fatal(err)
	}
	err := promoteRestored(staging, dest, []FileEntry{{Name: "vm.snap"}})
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("promoteRestored over existing file = %v, want already-exists error", err)
	}
	// The pre-existing file must be untouched.
	got, _ := os.ReadFile(filepath.Join(dest, "vm.snap"))
	if string(got) != "preexisting" {
		t.Errorf("pre-existing file was modified: %q", got)
	}
}

func TestGetRejectsInvalidSchemaVersion(t *testing.T) {
	store, root := newStore(t)
	id := "snap-v0"
	writeManifest(t, root, id, Manifest{
		SchemaVersion: 0, // no v0 parser exists
		SnapshotID:    id,
		Compression:   CompressionZstd,
		Files:         []FileEntry{{Name: "vm.snap", Key: id + "/vm.snap.zst", Size: 1, Digest: "00"}},
	})
	if _, err := store.Get(context.Background(), id, t.TempDir()); !errors.Is(err, ErrUnsupportedManifestVersion) {
		t.Fatalf("Get schema v0 = %v, want ErrUnsupportedManifestVersion", err)
	}
}

func TestGetRejectsUnknownCompression(t *testing.T) {
	store, root := newStore(t)
	id := "snap-codec"
	writeManifest(t, root, id, Manifest{
		SchemaVersion: ManifestSchemaVersion,
		SnapshotID:    id,
		Compression:   "lz4-evil",
		Files:         []FileEntry{{Name: "vm.snap", Key: id + "/vm.snap.zst", Size: 1, Digest: "00"}},
	})
	if _, err := store.Get(context.Background(), id, t.TempDir()); !errors.Is(err, ErrUnsupportedCompression) {
		t.Fatalf("Get unknown codec = %v, want ErrUnsupportedCompression", err)
	}
}

func TestGetRejectsEmptyFileList(t *testing.T) {
	store, root := newStore(t)
	id := "snap-empty"
	writeManifest(t, root, id, Manifest{
		SchemaVersion: ManifestSchemaVersion,
		SnapshotID:    id,
		Compression:   CompressionZstd,
		Files:         nil,
	})
	if _, err := store.Get(context.Background(), id, t.TempDir()); !errors.Is(err, ErrEmptyManifest) {
		t.Fatalf("Get empty-file manifest = %v, want ErrEmptyManifest", err)
	}
}

func TestManifestValidatesFileList(t *testing.T) {
	store, root := newStore(t)
	id := "snap-manifest-empty"
	writeManifest(t, root, id, Manifest{
		SchemaVersion: ManifestSchemaVersion,
		SnapshotID:    id,
		Compression:   CompressionZstd,
		Files:         nil,
	})
	if _, err := store.Manifest(context.Background(), id); !errors.Is(err, ErrEmptyManifest) {
		t.Fatalf("Manifest() with empty file list = %v, want ErrEmptyManifest", err)
	}
}
