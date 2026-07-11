//go:build !linux

package snapshotstore

// fsyncDir is a no-op on non-Linux platforms. Directory fsync is unsupported
// there (e.g. macOS returns EINVAL), and the rename-durability guarantee it
// provides is only needed by the Linux daemon. This keeps FSBlob usable in
// dev/test builds on other platforms instead of failing every Put after the
// rename has already succeeded.
func fsyncDir(string) error { return nil }
