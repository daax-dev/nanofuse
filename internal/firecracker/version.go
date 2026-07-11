package firecracker

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// BinaryVersion returns the version string reported by the Firecracker binary at
// path (via `firecracker --version`). It is used to pin the exact runtime
// version into a snapshot manifest so a restore on another node can reproduce or
// validate the runtime that produced the snapshot.
//
// The first output line is returned with a leading "Firecracker " prefix
// stripped when present, e.g. "v1.7.0". An empty path or a failing invocation
// returns an error; callers should treat version-pinning as best-effort.
func BinaryVersion(binaryPath string) (string, error) {
	if binaryPath == "" {
		return "", fmt.Errorf("firecracker binary path is required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// CombinedOutput (not Output) so a failure carries the binary's own message
	// (permission denied, missing loader, wrong arch, …) in the error, which is
	// far more useful operationally than a bare exec error.
	out, err := exec.CommandContext(ctx, binaryPath, "--version").CombinedOutput() // #nosec G204 -- path is operator-configured.
	if err != nil {
		if detail := strings.TrimSpace(string(out)); detail != "" {
			return "", fmt.Errorf("query firecracker version: %w: %s", err, detail)
		}
		return "", fmt.Errorf("query firecracker version: %w", err)
	}
	s := string(out)
	if idx := strings.IndexByte(s, '\n'); idx >= 0 {
		s = s[:idx]
	}
	version := strings.TrimSpace(s)
	version = strings.TrimPrefix(version, "Firecracker ")
	version = strings.TrimSpace(version)
	if version == "" {
		return "", fmt.Errorf("firecracker reported an empty version")
	}
	return version, nil
}
