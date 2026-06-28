//go:build unix

package credisolation

import (
	"os"
	"syscall"
)

// statOwner returns the POSIX owner uid/gid of info. ok is false when the
// platform does not expose POSIX ownership through the stat result.
func statOwner(info os.FileInfo) (uid, gid int, ok bool) {
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok || st == nil {
		return 0, 0, false
	}
	return int(st.Uid), int(st.Gid), true
}
