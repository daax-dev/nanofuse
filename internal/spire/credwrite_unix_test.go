//go:build unix

package spire

import (
	"os"
	"path/filepath"
	"testing"
)

// TestRemoveCredential_RejectsSymlinkedParent verifies the fd-anchored,
// no-follow removal refuses to unlink through a parent directory that has been
// swapped to a symlink, so a path-swap during a rotation failure cannot redirect
// the removal — matching the anti-symlink posture of the write path.
func TestRemoveCredential_RejectsSymlinkedParent(t *testing.T) {
	base := t.TempDir()
	realDir := filepath.Join(base, "real")
	if err := os.MkdirAll(realDir, svidDirMode); err != nil {
		t.Fatalf("mkdir real: %v", err)
	}
	name := "svid.json"
	realPath := filepath.Join(realDir, name)
	if err := os.WriteFile(realPath, []byte("cred"), svidFileMode); err != nil {
		t.Fatalf("write cred: %v", err)
	}

	// linkDir is a symlink standing in for a parent swapped underneath us.
	linkDir := filepath.Join(base, "link")
	if err := os.Symlink(realDir, linkDir); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	// Removal through the symlinked parent must be rejected (O_NOFOLLOW -> ELOOP).
	if err := removeCredential(linkDir, name); err == nil {
		t.Fatal("removeCredential must reject a symlinked parent directory")
	}
	// The real credential must be untouched by the rejected removal.
	if _, statErr := os.Stat(realPath); statErr != nil {
		t.Fatalf("credential behind a symlinked parent must not be removed: %v", statErr)
	}

	// Removing through the real (non-symlink) directory still works.
	if err := removeCredential(realDir, name); err != nil {
		t.Fatalf("removeCredential on the real directory: %v", err)
	}
	if _, statErr := os.Stat(realPath); !os.IsNotExist(statErr) {
		t.Fatal("credential must be removed via the real directory")
	}
}
