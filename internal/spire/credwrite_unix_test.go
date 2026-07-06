//go:build unix

package spire

import (
	"os"
	"path/filepath"
	"testing"
)

// TestWriteCredentialAtomic_HappyPath verifies the fd-anchored write persists
// the exact bytes at mode 0400 and, after the parent-directory fsync added for
// crash durability, the credential is readable at the expected path.
func TestWriteCredentialAtomic_HappyPath(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, svidDirMode); err != nil {
		t.Fatalf("chmod dir: %v", err)
	}
	name := "svid.json"
	data := []byte(`{"spiffe":"test"}`)

	if err := writeCredentialAtomic(dir, name, data); err != nil {
		t.Fatalf("writeCredentialAtomic: %v", err)
	}

	dest := filepath.Join(dir, name)
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read credential: %v", err)
	}
	if string(got) != string(data) {
		t.Fatalf("credential contents = %q, want %q", got, data)
	}
	info, err := os.Stat(dest)
	if err != nil {
		t.Fatalf("stat credential: %v", err)
	}
	if perm := info.Mode().Perm(); perm != svidFileMode {
		t.Fatalf("credential mode = %#o, want %#o", perm, svidFileMode)
	}
	// No temp file should linger after a successful atomic write.
	if _, err := os.Stat(filepath.Join(dir, "."+name+".tmp")); !os.IsNotExist(err) {
		t.Fatalf("temp file must not remain after rename: stat err = %v", err)
	}
}

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

// TestRemoveCredential_RejectsInsecureDirPerms verifies removeCredential applies
// the same directory permission validation as writeCredentialAtomic: if the
// parent directory became group/other-accessible between write and removal, the
// removal is refused (rather than proceeding on an insecure directory) and the
// credential is left in place — matching the write path's posture.
func TestRemoveCredential_RejectsInsecureDirPerms(t *testing.T) {
	dir := t.TempDir()
	name := "svid.json"
	credPath := filepath.Join(dir, name)
	if err := os.WriteFile(credPath, []byte("cred"), svidFileMode); err != nil {
		t.Fatalf("write cred: %v", err)
	}

	// The write path requires the directory be group/other-inaccessible; a write
	// into a secure dir succeeds.
	if err := os.Chmod(dir, svidDirMode); err != nil {
		t.Fatalf("chmod secure dir: %v", err)
	}
	if err := writeCredentialAtomic(dir, "probe.json", []byte("x")); err != nil {
		t.Fatalf("writeCredentialAtomic into secure dir: %v", err)
	}

	// Simulate the directory being loosened to group/other-accessible after the
	// credential was written.
	if err := os.Chmod(dir, 0o755); err != nil {
		t.Fatalf("chmod insecure dir: %v", err)
	}

	// Removal must now be rejected with the same insecure-permissions check the
	// write path enforces, and the credential must remain on disk.
	if err := removeCredential(dir, name); err == nil {
		t.Fatal("removeCredential must reject a group/other-accessible directory")
	}
	if _, statErr := os.Stat(credPath); statErr != nil {
		t.Fatalf("credential must not be removed from an insecure directory: %v", statErr)
	}

	// Restoring secure permissions allows removal to proceed.
	if err := os.Chmod(dir, svidDirMode); err != nil {
		t.Fatalf("restore secure dir: %v", err)
	}
	if err := removeCredential(dir, name); err != nil {
		t.Fatalf("removeCredential on a secure directory: %v", err)
	}
	if _, statErr := os.Stat(credPath); !os.IsNotExist(statErr) {
		t.Fatal("credential must be removed once the directory is secure again")
	}
}
