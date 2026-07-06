//go:build !unix

package spire

import (
	"fmt"
	"os"
	"path/filepath"
)

// writeCredentialAtomic is the portable fallback for non-unix platforms. It
// rejects a symlinked or non-directory target directory (best effort; POSIX
// ownership and fd-anchoring semantics are not available here) and performs an
// atomic temp-then-rename write with the file set to mode 0400. The SVID mount
// target is a Linux guest path; the fd-anchored implementation in
// credwrite_unix.go is the production path.
func writeCredentialAtomic(dir, name string, data []byte) error {
	info, err := os.Lstat(dir)
	if err != nil {
		return fmt.Errorf("lstat SVID directory %q: %w", dir, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("SVID directory %q is a symlink; refusing to write credential through it", dir)
	}
	if !info.IsDir() {
		return fmt.Errorf("SVID directory path %q is not a directory", dir)
	}

	tmp, err := os.CreateTemp(dir, "."+name+"-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp SVID file: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }() // no-op after a successful rename

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp SVID file: %w", err)
	}
	if err := tmp.Chmod(svidFileMode); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("chmod temp SVID file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp SVID file: %w", err)
	}
	dest := filepath.Join(dir, name)
	if err := os.Rename(tmpName, dest); err != nil {
		return fmt.Errorf("rename SVID file into place: %w", err)
	}
	return nil
}

// removeCredential deletes dir/name for non-unix platforms, mirroring the
// anti-symlink posture of writeCredentialAtomic above: it Lstat-checks the
// parent and refuses to remove through a symlinked directory. A missing
// directory or entry is success — the goal state (no credential on disk) is met.
// Residual limitation: without fd-anchoring there is a TOCTOU window between the
// Lstat and the os.Remove, so this is best-effort on non-unix. The fd-anchored
// implementation in credwrite_unix.go is the production (Linux guest) path.
func removeCredential(dir, name string) error {
	info, err := os.Lstat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // directory (and therefore the credential) is already gone
		}
		return fmt.Errorf("lstat SVID directory %q: %w", dir, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("SVID directory %q is a symlink; refusing to remove credential through it", dir)
	}
	if !info.IsDir() {
		return fmt.Errorf("SVID directory path %q is not a directory", dir)
	}
	dest := filepath.Join(dir, name)
	if err := os.Remove(dest); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove SVID document %q: %w", dest, err)
	}
	return nil
}
