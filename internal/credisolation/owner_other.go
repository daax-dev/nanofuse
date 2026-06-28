//go:build !unix

package credisolation

import "os"

// statOwner is a no-op on platforms that do not expose POSIX ownership. ok is
// false so callers skip the root:root assertion.
func statOwner(_ os.FileInfo) (uid, gid int, ok bool) {
	return 0, 0, false
}
