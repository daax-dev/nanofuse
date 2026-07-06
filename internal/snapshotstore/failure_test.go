package snapshotstore

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
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
