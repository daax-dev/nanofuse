package snapshotstore

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeFile writes data to a fresh file in dir and returns its path.
func writeFile(t *testing.T, dir, name string, data []byte) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, data, 0o600); err != nil {
		t.Fatalf("write %s: %v", p, err)
	}
	return p
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// newStore returns a TieredStore over a fresh FSBlob rooted in a temp dir, plus
// the blob root directory.
func newStore(t *testing.T) (*TieredStore, string) {
	t.Helper()
	root := t.TempDir()
	blob, err := NewFSBlob(root)
	if err != nil {
		t.Fatalf("NewFSBlob: %v", err)
	}
	return NewTieredStore(blob, Options{}), root
}

func TestPutGetRoundTrip(t *testing.T) {
	store, _ := newStore(t)
	src := t.TempDir()

	vmData := bytes.Repeat([]byte("vmstate-content-"), 4096) // compressible
	memData := make([]byte, 1<<20)
	if _, err := rand.Read(memData); err != nil { // incompressible
		t.Fatalf("rand: %v", err)
	}
	vmPath := writeFile(t, src, "vm.snap", vmData)
	memPath := writeFile(t, src, "mem.snap", memData)

	rt := RuntimeVersions{Firecracker: "v1.7.0", Kernel: "6.1.0", Nanofuse: "0.1.0"}
	files := []SourceFile{
		{Name: "vm.snap", Role: "vmstate", Path: vmPath},
		{Name: "mem.snap", Role: "memory", Path: memPath},
	}

	ctx := context.Background()
	manifest, err := store.Put(ctx, "snap-1", files, rt)
	if err != nil {
		t.Fatalf("Put: %v", err)
	}

	// SC-2: manifest content.
	if manifest.SchemaVersion != ManifestSchemaVersion {
		t.Errorf("schema version = %d, want %d", manifest.SchemaVersion, ManifestSchemaVersion)
	}
	if manifest.Compression != CompressionZstd {
		t.Errorf("compression = %q, want %q", manifest.Compression, CompressionZstd)
	}
	if manifest.Runtime != rt {
		t.Errorf("runtime = %+v, want %+v", manifest.Runtime, rt)
	}
	if len(manifest.Files) != 2 {
		t.Fatalf("files = %d, want 2", len(manifest.Files))
	}
	byName := map[string]FileEntry{}
	for _, fe := range manifest.Files {
		byName[fe.Name] = fe
	}
	if byName["vm.snap"].Digest != sha256Hex(vmData) {
		t.Errorf("vm.snap digest mismatch in manifest")
	}
	if byName["vm.snap"].Size != int64(len(vmData)) {
		t.Errorf("vm.snap size = %d, want %d", byName["vm.snap"].Size, len(vmData))
	}
	if byName["vm.snap"].CompressedSize <= 0 {
		t.Errorf("vm.snap compressed size not recorded")
	}
	// Compressible data should actually shrink.
	if byName["vm.snap"].CompressedSize >= byName["vm.snap"].Size {
		t.Errorf("expected compression to shrink compressible vm.snap: %d >= %d",
			byName["vm.snap"].CompressedSize, byName["vm.snap"].Size)
	}

	// SC-1: Get into a fresh dir yields byte-identical files.
	dest := t.TempDir()
	got, err := store.Get(ctx, "snap-1", dest)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.SnapshotID != "snap-1" {
		t.Errorf("restored manifest id = %q", got.SnapshotID)
	}
	assertFileEqual(t, filepath.Join(dest, "vm.snap"), vmData)
	assertFileEqual(t, filepath.Join(dest, "mem.snap"), memData)
}

func assertFileEqual(t *testing.T, path string, want []byte) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("file %s content mismatch (got %d bytes, want %d)", path, len(got), len(want))
	}
}

func TestGetMissingManifest(t *testing.T) {
	store, _ := newStore(t)
	ctx := context.Background()

	if _, err := store.Get(ctx, "does-not-exist", t.TempDir()); !errors.Is(err, ErrManifestNotFound) {
		t.Fatalf("Get missing = %v, want ErrManifestNotFound", err)
	}
	if _, err := store.Manifest(ctx, "does-not-exist"); !errors.Is(err, ErrManifestNotFound) {
		t.Fatalf("Manifest missing = %v, want ErrManifestNotFound", err)
	}
}

// failingBlob wraps a Blob and fails Put for keys containing failKey.
type failingBlob struct {
	Blob
	failKey string
}

func (f *failingBlob) Put(ctx context.Context, key string, r io.Reader) error {
	if f.failKey != "" && bytes.Contains([]byte(key), []byte(f.failKey)) {
		// Drain so the upstream writer goroutine unblocks, then fail.
		_, _ = io.Copy(io.Discard, r)
		return errors.New("injected put failure")
	}
	return f.Blob.Put(ctx, key, r)
}

func TestManifestWrittenLast(t *testing.T) {
	root := t.TempDir()
	inner, err := NewFSBlob(root)
	if err != nil {
		t.Fatalf("NewFSBlob: %v", err)
	}
	blob := &failingBlob{Blob: inner, failKey: "mem.snap"}
	store := NewTieredStore(blob, Options{Parallelism: 1})

	src := t.TempDir()
	files := []SourceFile{
		{Name: "vm.snap", Role: "vmstate", Path: writeFile(t, src, "vm.snap", []byte("ok"))},
		{Name: "mem.snap", Role: "memory", Path: writeFile(t, src, "mem.snap", []byte("boom"))},
	}

	ctx := context.Background()
	if _, err := store.Put(ctx, "snap-partial", files, RuntimeVersions{}); err == nil {
		t.Fatal("expected Put to fail when a data file fails")
	}

	// SC-3: no manifest was written, so the snapshot is not committed.
	exists, err := inner.Exists(ctx, objectKey("snap-partial", manifestObject))
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if exists {
		t.Fatal("manifest must not be written when a data file fails")
	}
	if _, err := store.Get(ctx, "snap-partial", t.TempDir()); !errors.Is(err, ErrManifestNotFound) {
		t.Fatalf("Get partial = %v, want ErrManifestNotFound", err)
	}
	if ids, _ := store.List(ctx); len(ids) != 0 {
		t.Fatalf("List = %v, want empty (partial snapshot must not appear)", ids)
	}
}

func TestGetDetectsCorruption(t *testing.T) {
	store, root := newStore(t)
	src := t.TempDir()
	data := bytes.Repeat([]byte("payload"), 1000)
	files := []SourceFile{{Name: "vm.snap", Role: "vmstate", Path: writeFile(t, src, "vm.snap", data)}}

	ctx := context.Background()
	if _, err := store.Put(ctx, "snap-c", files, RuntimeVersions{}); err != nil {
		t.Fatalf("Put: %v", err)
	}

	// Corrupt the stored compressed object.
	objPath := filepath.Join(root, "snap-c", "vm.snap.zst")
	blobBytes, err := os.ReadFile(objPath)
	if err != nil {
		t.Fatalf("read object: %v", err)
	}
	blobBytes[len(blobBytes)/2] ^= 0xFF
	if err := os.WriteFile(objPath, blobBytes, 0o600); err != nil {
		t.Fatalf("corrupt object: %v", err)
	}

	// SC-4: restore must fail and not surface corrupt bytes.
	dest := t.TempDir()
	if _, err := store.Get(ctx, "snap-c", dest); err == nil {
		t.Fatal("expected Get to fail on corrupt object")
	}
	if _, statErr := os.Stat(filepath.Join(dest, "vm.snap")); !os.IsNotExist(statErr) {
		t.Fatal("corrupt file must not be written to destination")
	}
}

func TestManifestRejectsIDMismatch(t *testing.T) {
	// A manifest whose recorded snapshot_id differs from the id it is fetched
	// under (corrupted/swapped/tampered object) must be rejected, not returned.
	store, root := newStore(t)
	src := t.TempDir()
	data := bytes.Repeat([]byte("payload"), 128)
	files := []SourceFile{{Name: "vm.snap", Role: "vmstate", Path: writeFile(t, src, "vm.snap", data)}}

	ctx := context.Background()
	if _, err := store.Put(ctx, "snap-real", files, RuntimeVersions{}); err != nil {
		t.Fatalf("Put: %v", err)
	}

	// Tamper only the snapshot_id field inside the stored manifest, leaving the
	// object at its original key.
	manPath := filepath.Join(root, "snap-real", manifestObject)
	raw, err := os.ReadFile(manPath)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	var m Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal manifest: %v", err)
	}
	m.SnapshotID = "snap-imposter"
	tampered, err := json.Marshal(&m)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	if err := os.WriteFile(manPath, tampered, 0o600); err != nil {
		t.Fatalf("write tampered manifest: %v", err)
	}

	if _, err := store.Manifest(ctx, "snap-real"); !errors.Is(err, ErrManifestIDMismatch) {
		t.Fatalf("Manifest err = %v, want ErrManifestIDMismatch", err)
	}
	// Get, which resolves the manifest first, must also reject it.
	if _, err := store.Get(ctx, "snap-real", t.TempDir()); !errors.Is(err, ErrManifestIDMismatch) {
		t.Fatalf("Get err = %v, want ErrManifestIDMismatch", err)
	}
}

func TestCrossNodeRestore(t *testing.T) {
	// SC-5: producer store on host A, restore via a separate store instance on
	// host B, sharing only the object backend (a shared bucket/directory).
	root := t.TempDir()
	blobA, _ := NewFSBlob(root)
	storeA := NewTieredStore(blobA, Options{})

	src := t.TempDir()
	data := bytes.Repeat([]byte("cross-node-"), 5000)
	files := []SourceFile{{Name: "vm.snap", Role: "vmstate", Path: writeFile(t, src, "vm.snap", data)}}

	ctx := context.Background()
	if _, err := storeA.Put(ctx, "session-42", files, RuntimeVersions{Firecracker: "v1.7.0"}); err != nil {
		t.Fatalf("host A Put: %v", err)
	}

	// Host B: a brand new store over the same backend, no shared in-memory state.
	blobB, _ := NewFSBlob(root)
	storeB := NewTieredStore(blobB, Options{})

	ids, err := storeB.List(ctx)
	if err != nil {
		t.Fatalf("host B List: %v", err)
	}
	if len(ids) != 1 || ids[0] != "session-42" {
		t.Fatalf("host B List = %v, want [session-42]", ids)
	}
	m, err := storeB.Manifest(ctx, "session-42")
	if err != nil {
		t.Fatalf("host B Manifest: %v", err)
	}
	if m.Runtime.Firecracker != "v1.7.0" {
		t.Errorf("host B sees firecracker = %q, want v1.7.0", m.Runtime.Firecracker)
	}
	// Manifest() must normalize each Key to the canonical *compressed* object key
	// (<id>/<name>.zst) — the key objects are actually stored/read under — so a
	// caller using Manifest() directly sees keys that resolve in the backend.
	for _, fe := range m.Files {
		want := "session-42/" + fe.Name + ".zst"
		if fe.Key != want {
			t.Errorf("manifest Key for %q = %q, want canonical compressed key %q", fe.Name, fe.Key, want)
		}
	}
	dest := t.TempDir()
	if _, err := storeB.Get(ctx, "session-42", dest); err != nil {
		t.Fatalf("host B Get: %v", err)
	}
	assertFileEqual(t, filepath.Join(dest, "vm.snap"), data)
}

func TestUnsupportedManifestVersion(t *testing.T) {
	store, root := newStore(t)
	ctx := context.Background()

	// Write a manifest with a future schema version directly to the backend.
	future := Manifest{SchemaVersion: ManifestSchemaVersion + 1, SnapshotID: "snap-future"}
	data, _ := json.Marshal(future)
	if err := os.MkdirAll(filepath.Join(root, "snap-future"), 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "snap-future", manifestObject), data, 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	// SC-6.
	if _, err := store.Manifest(ctx, "snap-future"); !errors.Is(err, ErrUnsupportedManifestVersion) {
		t.Fatalf("Manifest future = %v, want ErrUnsupportedManifestVersion", err)
	}
	if _, err := store.Get(ctx, "snap-future", t.TempDir()); !errors.Is(err, ErrUnsupportedManifestVersion) {
		t.Fatalf("Get future = %v, want ErrUnsupportedManifestVersion", err)
	}
}

func TestGetRejectsPathTraversal(t *testing.T) {
	store, root := newStore(t)
	ctx := context.Background()

	// Hand-craft a snapshot whose manifest names a traversal file. The data
	// object is stored under a safe key; only the manifest Name is hostile.
	id := "snap-evil"
	payload := []byte("evil")
	// Reuse the store's own put path for a legitimately-stored object, then
	// rewrite the manifest to point its restore Name at a traversal path.
	files := []SourceFile{{Name: "vm.snap", Role: "vmstate", Path: writeFile(t, t.TempDir(), "vm.snap", payload)}}
	m, err := store.Put(ctx, id, files, RuntimeVersions{})
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
	m.Files[0].Name = "../escaped"
	data, _ := json.Marshal(m)
	if err := os.WriteFile(filepath.Join(root, id, manifestObject), data, 0o600); err != nil {
		t.Fatalf("rewrite manifest: %v", err)
	}

	// SC-7.
	dest := t.TempDir()
	_, err = store.Get(ctx, id, dest)
	if !errors.Is(err, ErrUnsafeName) {
		t.Fatalf("Get traversal = %v, want ErrUnsafeName", err)
	}
	if _, statErr := os.Stat(filepath.Join(filepath.Dir(dest), "escaped")); !os.IsNotExist(statErr) {
		t.Fatal("path traversal must not write outside destination")
	}
}

func TestGetRejectsSizeMismatch(t *testing.T) {
	store, root := newStore(t)
	ctx := context.Background()

	id := "snap-bomb"
	data := bytes.Repeat([]byte("A"), 10000)
	files := []SourceFile{{Name: "vm.snap", Role: "vmstate", Path: writeFile(t, t.TempDir(), "vm.snap", data)}}
	m, err := store.Put(ctx, id, files, RuntimeVersions{})
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
	// Understate the declared size so the actual decompressed stream overruns it
	// (the decompression-bomb / torn-write guard).
	m.Files[0].Size = 5
	out, _ := json.Marshal(m)
	if err := os.WriteFile(filepath.Join(root, id, manifestObject), out, 0o600); err != nil {
		t.Fatalf("rewrite manifest: %v", err)
	}

	// SC-8.
	dest := t.TempDir()
	if _, err := store.Get(ctx, id, dest); !errors.Is(err, ErrSizeMismatch) {
		t.Fatalf("Get oversize = %v, want ErrSizeMismatch", err)
	}
	if _, statErr := os.Stat(filepath.Join(dest, "vm.snap")); !os.IsNotExist(statErr) {
		t.Fatal("size-mismatched file must not be committed")
	}
}

func TestValidateName(t *testing.T) {
	bad := []string{"", ".", "..", "a/b", "a\\b", "../x", "x/..", manifestObject}
	for _, n := range bad {
		if err := validateName(n); err == nil {
			t.Errorf("validateName(%q) = nil, want error", n)
		}
	}
	for _, n := range []string{"vm.snap", "mem.snap", "a.b.c"} {
		if err := validateName(n); err != nil {
			t.Errorf("validateName(%q) = %v, want nil", n, err)
		}
	}
}

func TestPutRejectsBadID(t *testing.T) {
	store, _ := newStore(t)
	ctx := context.Background()
	for _, id := range []string{"", "..", "a/b"} {
		if _, err := store.Put(ctx, id, []SourceFile{{Name: "x", Path: "x"}}, RuntimeVersions{}); !errors.Is(err, ErrInvalidSnapshotID) {
			t.Errorf("Put id=%q err = %v, want ErrInvalidSnapshotID", id, err)
		}
	}
}

func TestEmptyFileRoundTrip(t *testing.T) {
	store, _ := newStore(t)
	ctx := context.Background()
	src := t.TempDir()
	files := []SourceFile{{Name: "empty.snap", Role: "vmstate", Path: writeFile(t, src, "empty.snap", nil)}}
	if _, err := store.Put(ctx, "snap-empty", files, RuntimeVersions{}); err != nil {
		t.Fatalf("Put empty: %v", err)
	}
	dest := t.TempDir()
	if _, err := store.Get(ctx, "snap-empty", dest); err != nil {
		t.Fatalf("Get empty: %v", err)
	}
	assertFileEqual(t, filepath.Join(dest, "empty.snap"), []byte{})
}

// A tampered manifest Key must not redirect the download: getFile derives the
// object key deterministically from the snapshot id + validated file name.
func TestGetIgnoresTamperedManifestKey(t *testing.T) {
	ctx := context.Background()
	store, root := newStore(t)
	src := t.TempDir()
	data := []byte("real vmstate payload")
	files := []SourceFile{{Name: "vm.snap", Role: "vmstate", Path: writeFile(t, src, "vm.snap", data)}}
	if _, err := store.Put(ctx, "snap-t", files, RuntimeVersions{}); err != nil {
		t.Fatalf("Put: %v", err)
	}

	// Tamper the manifest so the file entry's Key points at an attacker object.
	mpath := filepath.Join(root, "snap-t", manifestObject)
	raw, err := os.ReadFile(mpath)
	if err != nil {
		t.Fatal(err)
	}
	var m Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	m.Files[0].Key = "snap-t/../evil"
	out, _ := json.MarshalIndent(&m, "", "  ")
	if err := os.WriteFile(mpath, out, 0o600); err != nil {
		t.Fatal(err)
	}

	dest := t.TempDir()
	if _, err := store.Get(ctx, "snap-t", dest); err != nil {
		t.Fatalf("Get should still succeed using the derived key: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dest, "vm.snap"))
	if err != nil {
		t.Fatalf("restored file missing: %v", err)
	}
	if string(got) != string(data) {
		t.Errorf("restored content = %q, want %q (tampered key must be ignored)", got, data)
	}
}

func TestPutRejectsDuplicateNames(t *testing.T) {
	ctx := context.Background()
	store, _ := newStore(t)
	dir := t.TempDir()
	p := writeFile(t, dir, "dup", []byte("x"))
	files := []SourceFile{
		{Name: "dup", Role: "a", Path: p},
		{Name: "dup", Role: "b", Path: p},
	}
	if _, err := store.Put(ctx, "snap-dup", files, RuntimeVersions{}); err == nil {
		t.Fatal("expected error for duplicate file names, got nil")
	} else if !strings.Contains(err.Error(), "duplicate file name") {
		t.Errorf("error = %v, want duplicate-name rejection", err)
	}
}
