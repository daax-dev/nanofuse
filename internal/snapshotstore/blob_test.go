package snapshotstore

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestFSBlobRoundTrip(t *testing.T) {
	ctx := context.Background()
	blob, err := NewFSBlob(t.TempDir())
	if err != nil {
		t.Fatalf("NewFSBlob: %v", err)
	}

	if ok, _ := blob.Exists(ctx, "a/b.txt"); ok {
		t.Fatal("Exists should be false for absent key")
	}
	if _, err := blob.Get(ctx, "a/b.txt"); !errors.Is(err, ErrObjectNotFound) {
		t.Fatalf("Get absent = %v, want ErrObjectNotFound", err)
	}

	want := []byte("hello object store")
	if err := blob.Put(ctx, "a/b.txt", bytes.NewReader(want)); err != nil {
		t.Fatalf("Put: %v", err)
	}
	ok, err := blob.Exists(ctx, "a/b.txt")
	if err != nil || !ok {
		t.Fatalf("Exists after Put = %v, %v", ok, err)
	}
	rc, err := blob.Get(ctx, "a/b.txt")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	got, _ := io.ReadAll(rc)
	_ = rc.Close()
	if !bytes.Equal(got, want) {
		t.Fatalf("Get = %q, want %q", got, want)
	}

	// Overwrite is atomic and visible.
	if err := blob.Put(ctx, "a/b.txt", bytes.NewReader([]byte("v2"))); err != nil {
		t.Fatalf("overwrite Put: %v", err)
	}
	rc2, _ := blob.Get(ctx, "a/b.txt")
	got2, _ := io.ReadAll(rc2)
	_ = rc2.Close()
	if string(got2) != "v2" {
		t.Fatalf("after overwrite = %q, want v2", got2)
	}
}

func TestFSBlobList(t *testing.T) {
	ctx := context.Background()
	blob, _ := NewFSBlob(t.TempDir())
	keys := []string{"s1/vm.snap.zst", "s1/manifest.json", "s2/manifest.json", "other/x"}
	for _, k := range keys {
		if err := blob.Put(ctx, k, bytes.NewReader([]byte("x"))); err != nil {
			t.Fatalf("Put %s: %v", k, err)
		}
	}
	got, err := blob.List(ctx, "s1/")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	sort.Strings(got)
	want := []string{"s1/manifest.json", "s1/vm.snap.zst"}
	if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("List(s1/) = %v, want %v", got, want)
	}

	all, _ := blob.List(ctx, "")
	if len(all) != 4 {
		t.Fatalf("List(all) = %v, want 4 keys", all)
	}
}

func TestFSBlobConfinesEscape(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	blob, _ := NewFSBlob(root)
	// An escaping key is normalized (confined) to within the root rather than
	// escaping it — defense in depth beneath the Store's name validation.
	if err := blob.Put(ctx, "../../escape", bytes.NewReader([]byte("x"))); err != nil {
		t.Fatalf("Put: %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(filepath.Dir(root), "escape")); !os.IsNotExist(statErr) {
		t.Fatal("escaping key must not write outside blob root")
	}
}

func TestNewFSBlobRequiresRoot(t *testing.T) {
	if _, err := NewFSBlob(""); err == nil {
		t.Fatal("NewFSBlob(\"\") should fail")
	}
}

func TestFSBlobGetReaderHonorsCanceledContext(t *testing.T) {
	blob, err := NewFSBlob(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := blob.Put(ctx, "a/b", bytes.NewReader([]byte("payload"))); err != nil {
		t.Fatal(err)
	}
	cctx, cancel := context.WithCancel(ctx)
	rc, err := blob.Get(cctx, "a/b")
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()
	cancel() // cancel after opening; the reader must now refuse reads
	if _, rerr := rc.Read(make([]byte, 4)); !errors.Is(rerr, context.Canceled) {
		t.Errorf("Read after cancel = %v, want context.Canceled", rerr)
	}
}
