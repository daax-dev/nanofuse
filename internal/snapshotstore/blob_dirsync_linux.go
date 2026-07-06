//go:build linux

package snapshotstore

import "os"

// fsyncDir flushes a directory's metadata so a rename into it is durable across
// a crash/power loss. This is required on Linux (the nanofuse daemon's platform)
// to preserve the manifest-last commit ordering of the snapshot store.
func fsyncDir(dir string) error {
	d, err := os.Open(dir) // #nosec G304 -- dir is derived from a keyPath confined to the blob root.
	if err != nil {
		return err
	}
	if err := d.Sync(); err != nil {
		_ = d.Close()
		return err
	}
	return d.Close()
}
