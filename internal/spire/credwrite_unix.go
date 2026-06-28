//go:build unix

package spire

import (
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
	if st.Mode&unix.S_IFDIR == 0 {
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
	return nil
}
