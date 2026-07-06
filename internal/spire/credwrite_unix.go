//go:build unix

package spire

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

// writeCredentialAtomic writes data to dir/name with mode 0400 using
// directory-fd-anchored operations. The directory is opened O_NOFOLLOW (so a
// symlinked directory is rejected, not followed), its mode/owner are checked via
// fstat on that fd, and the temp create + rename happen relative to the fd
// (openat/renameat). This removes the TOCTOU window of path-based checks: the
// directory cannot be swapped or redirected between validation and write.
func writeCredentialAtomic(dir, name string, data []byte) error {
	dirFD, err := unix.Open(dir, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		return fmt.Errorf("open SVID directory %q (no-follow): %w", dir, err)
	}
	defer func() { _ = unix.Close(dirFD) }()

	var st unix.Stat_t
	if err := unix.Fstat(dirFD, &st); err != nil {
		return fmt.Errorf("fstat SVID directory %q: %w", dir, err)
	}
	if st.Mode&unix.S_IFMT != unix.S_IFDIR {
		return fmt.Errorf("SVID directory path %q is not a directory", dir)
	}
	if st.Mode&0o077 != 0 {
		return fmt.Errorf("SVID directory %q has insecure permissions %#o (must be group/other-inaccessible)", dir, st.Mode&0o777)
	}
	if euid := os.Geteuid(); euid >= 0 && st.Uid != uint32(euid) && st.Uid != 0 { //nolint:gosec // euid >= 0 guarded
		return fmt.Errorf("SVID directory %q is owned by uid %d, not the current user (%d) or root", dir, st.Uid, euid)
	}

	tmp := "." + name + ".tmp"
	_ = unix.Unlinkat(dirFD, tmp, 0) // clear any stale temp from a prior crash

	fd, err := unix.Openat(dirFD, tmp, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_NOFOLLOW|unix.O_CLOEXEC, uint32(svidFileMode))
	if err != nil {
		return fmt.Errorf("create temp SVID file: %w", err)
	}
	f := os.NewFile(uintptr(fd), tmp)
	cleanup := func() { _ = unix.Unlinkat(dirFD, tmp, 0) }

	// Enforce exact mode regardless of umask.
	if err := f.Chmod(svidFileMode); err != nil {
		_ = f.Close()
		cleanup()
		return fmt.Errorf("chmod temp SVID file: %w", err)
	}
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		cleanup()
		return fmt.Errorf("write temp SVID file: %w", err)
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		cleanup()
		return fmt.Errorf("sync temp SVID file: %w", err)
	}
	if err := f.Close(); err != nil {
		cleanup()
		return fmt.Errorf("close temp SVID file: %w", err)
	}
	if err := unix.Renameat(dirFD, tmp, dirFD, name); err != nil {
		cleanup()
		return fmt.Errorf("rename SVID file into place: %w", err)
	}
	// Durability: the rename created/replaced a directory entry. fsync the parent
	// directory (via the fd already anchored above) so the entry survives a
	// crash/power-loss — without it a successful return could still lose or roll
	// back the credential on reboot. The file entry now points at persisted data
	// (temp file was fsync'd before the rename), so no cleanup on failure: the
	// credential is in place, only its durability could not be confirmed.
	if err := unix.Fsync(dirFD); err != nil {
		return fmt.Errorf("fsync SVID directory after rename: %w", err)
	}
	return nil
}

// removeCredential deletes dir/name using the same directory-fd-anchored,
// no-follow posture as writeCredentialAtomic. The directory is opened
// O_NOFOLLOW (a directory swapped to a symlink is rejected with ELOOP, not
// followed) and the entry is removed via unlinkat relative to that fd, so a
// parent redirected between write and removal cannot cause the wrong path to be
// unlinked. A missing directory or entry is treated as success — the goal state
// (no credential on disk) is already met. This closes the same TOCTOU window the
// write path closes; no residual path-swap limitation on unix.
func removeCredential(dir, name string) error {
	dirFD, err := unix.Open(dir, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		if errors.Is(err, unix.ENOENT) {
			return nil // directory (and therefore the credential) is already gone
		}
		return fmt.Errorf("open SVID directory %q (no-follow): %w", dir, err)
	}
	defer func() { _ = unix.Close(dirFD) }()

	var st unix.Stat_t
	if err := unix.Fstat(dirFD, &st); err != nil {
		return fmt.Errorf("fstat SVID directory %q: %w", dir, err)
	}
	if st.Mode&unix.S_IFMT != unix.S_IFDIR {
		return fmt.Errorf("SVID directory path %q is not a directory", dir)
	}
	if err := unix.Unlinkat(dirFD, name, 0); err != nil && !errors.Is(err, unix.ENOENT) {
		return fmt.Errorf("unlink SVID document %q: %w", name, err)
	}
	// Fsync the directory so the entry removal is durable: without it a
	// crash/power-loss could resurrect the removed (expired) credential after
	// reboot even though removal returned success.
	if err := unix.Fsync(dirFD); err != nil {
		return fmt.Errorf("fsync SVID directory %q after removal: %w", dir, err)
	}
	return nil
}
