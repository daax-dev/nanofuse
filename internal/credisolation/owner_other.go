//go:build !unix

package credisolation

import "os"

// statOwner is a no-op on platforms that do not expose POSIX ownership. ok is
// false, so a caller requiring root:root ownership fails closed (the assertion
// cannot be verified) rather than passing it silently.
func statOwner(_ os.FileInfo) (uid, gid int, ok bool) {
	return 0, 0, false
}
