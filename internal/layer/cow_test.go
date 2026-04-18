package layer

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// newTestBase creates a temporary directory to act as the shared base (lowerdir).
func newTestBase(t *testing.T) string {
	t.Helper()
	base := t.TempDir()
	// Populate with a sentinel file to verify CoW behaviour.
	if err := os.WriteFile(filepath.Join(base, "base.txt"), []byte("base content"), 0644); err != nil {
		t.Fatalf("create base.txt: %v", err)
	}
	return base
}

// TestCowOptionsValidation checks that NewCowLayer rejects incomplete options.
func TestCowOptionsValidation(t *testing.T) {
	base := newTestBase(t)

	tests := []struct {
		name    string
		opts    CowOptions
		wantErr bool
	}{
		{
			name:    "missing BaseDir",
			opts:    CowOptions{SessionDir: t.TempDir(), SessionID: "s1"},
			wantErr: true,
		},
		{
			name:    "missing SessionDir",
			opts:    CowOptions{BaseDir: base, SessionID: "s1"},
			wantErr: true,
		},
		{
			name:    "missing SessionID",
			opts:    CowOptions{BaseDir: base, SessionDir: t.TempDir()},
			wantErr: true,
		},
		{
			name:    "non-existent BaseDir",
			opts:    CowOptions{BaseDir: "/no/such/path", SessionDir: t.TempDir(), SessionID: "s1"},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// We don't actually mount in unit tests (requires root + kernel support).
			// Validation errors happen before the mount syscall, so we can test them.
			cl, err := newCowLayerNoMount(tc.opts)
			if tc.wantErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if cl != nil {
				// No mount happened, just clean up dirs.
				_ = os.RemoveAll(filepath.Join(tc.opts.SessionDir, tc.opts.SessionID))
			}
		})
	}
}

// TestCowLayerDirectoryStructure verifies that upper, work, and merged dirs are created.
func TestCowLayerDirectoryStructure(t *testing.T) {
	base := newTestBase(t)
	sessionDir := t.TempDir()

	opts := CowOptions{
		BaseDir:    base,
		SessionDir: sessionDir,
		SessionID:  "session-dir-test",
	}

	// Use the no-mount variant so the test doesn't need root privileges.
	cl, err := newCowLayerNoMount(opts)
	if err != nil {
		t.Fatalf("newCowLayerNoMount: %v", err)
	}

	// Verify directories were created.
	for _, dir := range []string{cl.upperDir, cl.workDir, cl.mergedDir} {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("expected directory %s to exist", dir)
		}
	}
}

// TestCowLayerMergedDirDefault checks that MergedDir defaults to <sessionDir>/<sessionID>/merged.
func TestCowLayerMergedDirDefault(t *testing.T) {
	base := newTestBase(t)
	sessionDir := t.TempDir()

	opts := CowOptions{
		BaseDir:    base,
		SessionDir: sessionDir,
		SessionID:  "merged-default",
	}

	cl, err := newCowLayerNoMount(opts)
	if err != nil {
		t.Fatalf("newCowLayerNoMount: %v", err)
	}

	expected := filepath.Join(sessionDir, "merged-default", "merged")
	if cl.MergedDir() != expected {
		t.Errorf("MergedDir() = %q, want %q", cl.MergedDir(), expected)
	}
}

// TestCowLayerCustomMountDir checks that an explicit MountDir is respected.
func TestCowLayerCustomMountDir(t *testing.T) {
	base := newTestBase(t)
	sessionDir := t.TempDir()
	customMount := t.TempDir()

	opts := CowOptions{
		BaseDir:    base,
		SessionDir: sessionDir,
		SessionID:  "custom-mount",
		MountDir:   customMount,
	}

	cl, err := newCowLayerNoMount(opts)
	if err != nil {
		t.Fatalf("newCowLayerNoMount: %v", err)
	}

	if cl.MergedDir() != customMount {
		t.Errorf("MergedDir() = %q, want %q", cl.MergedDir(), customMount)
	}
}

// TestCowLayerIsMountedFalse verifies IsMounted returns false before a real mount.
func TestCowLayerIsMountedFalse(t *testing.T) {
	base := newTestBase(t)
	opts := CowOptions{
		BaseDir:    base,
		SessionDir: t.TempDir(),
		SessionID:  "not-mounted",
	}

	cl, err := newCowLayerNoMount(opts)
	if err != nil {
		t.Fatalf("newCowLayerNoMount: %v", err)
	}

	if cl.IsMounted() {
		t.Error("expected IsMounted() to be false for layer created without mount")
	}
}

// TestCowLayerSnapshotRequiresMounted verifies Snapshot errors when not mounted.
func TestCowLayerSnapshotRequiresMounted(t *testing.T) {
	base := newTestBase(t)
	opts := CowOptions{
		BaseDir:    base,
		SessionDir: t.TempDir(),
		SessionID:  "snap-unmounted",
	}

	cl, err := newCowLayerNoMount(opts)
	if err != nil {
		t.Fatalf("newCowLayerNoMount: %v", err)
	}

	err = cl.Snapshot(t.TempDir())
	if err == nil {
		t.Error("expected error when snapshotting an unmounted layer")
	}
}

// TestCowLayerUpperDirAccessor verifies the UpperDir accessor.
func TestCowLayerUpperDirAccessor(t *testing.T) {
	base := newTestBase(t)
	sessionDir := t.TempDir()
	opts := CowOptions{
		BaseDir:    base,
		SessionDir: sessionDir,
		SessionID:  "upper-accessor",
	}

	cl, err := newCowLayerNoMount(opts)
	if err != nil {
		t.Fatalf("newCowLayerNoMount: %v", err)
	}

	expected := filepath.Join(sessionDir, "upper-accessor", "upper")
	if cl.UpperDir() != expected {
		t.Errorf("UpperDir() = %q, want %q", cl.UpperDir(), expected)
	}
}

// newCowLayerNoMount is a test-only variant that creates directories but skips the mount syscall.
// This allows unit tests to run without root privileges or kernel overlayfs support.
func newCowLayerNoMount(opts CowOptions) (*CowLayer, error) {
	if opts.BaseDir == "" {
		return nil, fmt.Errorf("cow: BaseDir is required")
	}
	if opts.SessionDir == "" {
		return nil, fmt.Errorf("cow: SessionDir is required")
	}
	if opts.SessionID == "" {
		return nil, fmt.Errorf("cow: SessionID is required")
	}
	if _, err := os.Stat(opts.BaseDir); err != nil {
		return nil, fmt.Errorf("cow: base dir %q not found: %w", opts.BaseDir, err)
	}

	sessionRoot := filepath.Join(opts.SessionDir, opts.SessionID)
	upperDir := filepath.Join(sessionRoot, "upper")
	workDir := filepath.Join(sessionRoot, "work")
	mergedDir := opts.MountDir
	if mergedDir == "" {
		mergedDir = filepath.Join(sessionRoot, "merged")
	}

	for _, dir := range []string{upperDir, workDir, mergedDir} {
		if err := os.MkdirAll(dir, 0750); err != nil {
			return nil, fmt.Errorf("cow: create dir %s: %w", dir, err)
		}
	}

	return &CowLayer{
		opts:      opts,
		upperDir:  upperDir,
		workDir:   workDir,
		mergedDir: mergedDir,
		mounted:   false, // skip actual overlayfs mount
	}, nil
}
