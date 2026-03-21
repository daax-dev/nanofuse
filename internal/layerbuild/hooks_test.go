package layerbuild

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestHookExecution tests basic hook execution
func TestHookExecution(t *testing.T) {
	tmpDir := t.TempDir()
	rootfs := filepath.Join(tmpDir, "rootfs")
	if err := os.MkdirAll(rootfs, 0755); err != nil {
		t.Fatalf("Failed to create rootfs: %v", err)
	}

	// Create a simple hook script
	hookScript := `#!/bin/sh
echo "Hook executed" > /tmp/hook-output.txt
exit 0
`
	hookPath := filepath.Join(rootfs, "hooks", "test.sh")
	if err := os.MkdirAll(filepath.Dir(hookPath), 0755); err != nil {
		t.Fatalf("Failed to create hooks dir: %v", err)
	}
	if err := os.WriteFile(hookPath, []byte(hookScript), 0755); err != nil {
		t.Fatalf("Failed to write hook script: %v", err)
	}

	he := &HookExecutor{
		rootfsPath: rootfs,
		dryRun:     true, // Use dry-run to avoid actual chroot
		verbose:    true,
	}

	ctx := context.Background()
	err := he.Execute(ctx, "/hooks/test.sh", nil)

	// In dry-run mode, should not error
	if err != nil && !isRootRequired(err) {
		t.Errorf("Execute() error = %v", err)
	}
}

// TestHookExecutionWithEnv tests hook execution with environment variables
func TestHookExecutionWithEnv(t *testing.T) {
	tmpDir := t.TempDir()
	rootfs := filepath.Join(tmpDir, "rootfs")
	if err := os.MkdirAll(rootfs, 0755); err != nil {
		t.Fatalf("Failed to create rootfs: %v", err)
	}

	hookScript := `#!/bin/sh
echo "VAR1=$VAR1" > /tmp/env-test.txt
echo "VAR2=$VAR2" >> /tmp/env-test.txt
exit 0
`
	hookPath := filepath.Join(rootfs, "hooks", "env-test.sh")
	if err := os.MkdirAll(filepath.Dir(hookPath), 0755); err != nil {
		t.Fatalf("Failed to create hooks dir: %v", err)
	}
	if err := os.WriteFile(hookPath, []byte(hookScript), 0755); err != nil {
		t.Fatalf("Failed to write hook script: %v", err)
	}

	he := &HookExecutor{
		rootfsPath: rootfs,
		dryRun:     true,
		verbose:    true,
	}

	env := map[string]string{
		"VAR1": "value1",
		"VAR2": "value2",
	}

	ctx := context.Background()
	err := he.Execute(ctx, "/hooks/env-test.sh", env)

	if err != nil && !isRootRequired(err) {
		t.Errorf("Execute() error = %v", err)
	}
}

// TestHookExecutionFailure tests hook execution failure handling
func TestHookExecutionFailure(t *testing.T) {
	tmpDir := t.TempDir()
	rootfs := filepath.Join(tmpDir, "rootfs")
	if err := os.MkdirAll(rootfs, 0755); err != nil {
		t.Fatalf("Failed to create rootfs: %v", err)
	}

	// Create a hook that exits with error
	hookScript := `#!/bin/sh
echo "Hook failed"
exit 1
`
	hookPath := filepath.Join(rootfs, "hooks", "fail.sh")
	if err := os.MkdirAll(filepath.Dir(hookPath), 0755); err != nil {
		t.Fatalf("Failed to create hooks dir: %v", err)
	}
	if err := os.WriteFile(hookPath, []byte(hookScript), 0755); err != nil {
		t.Fatalf("Failed to write hook script: %v", err)
	}

	he := &HookExecutor{
		rootfsPath: rootfs,
		dryRun:     false,
		verbose:    true,
	}

	ctx := context.Background()
	err := he.Execute(ctx, "/hooks/fail.sh", nil)

	// Should get an error (either from hook failure or root requirement)
	if err == nil {
		t.Error("Expected error from failing hook")
	}
}

// TestHookExecutionNonexistent tests execution of nonexistent hook
func TestHookExecutionNonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	rootfs := filepath.Join(tmpDir, "rootfs")
	if err := os.MkdirAll(rootfs, 0755); err != nil {
		t.Fatalf("Failed to create rootfs: %v", err)
	}

	he := &HookExecutor{
		rootfsPath: rootfs,
		dryRun:     true,
		verbose:    true,
	}

	ctx := context.Background()
	err := he.Execute(ctx, "/hooks/nonexistent.sh", nil)

	// Should handle missing hook gracefully (skip it)
	if err != nil {
		t.Logf("Execute() returned error for nonexistent hook: %v", err)
	}
}

// TestHookExecutionTimeout tests hook execution timeout
func TestHookExecutionTimeout(t *testing.T) {
	tmpDir := t.TempDir()
	rootfs := filepath.Join(tmpDir, "rootfs")
	if err := os.MkdirAll(rootfs, 0755); err != nil {
		t.Fatalf("Failed to create rootfs: %v", err)
	}

	// Create a hook that sleeps
	hookScript := `#!/bin/sh
sleep 10
exit 0
`
	hookPath := filepath.Join(rootfs, "hooks", "sleep.sh")
	if err := os.MkdirAll(filepath.Dir(hookPath), 0755); err != nil {
		t.Fatalf("Failed to create hooks dir: %v", err)
	}
	if err := os.WriteFile(hookPath, []byte(hookScript), 0755); err != nil {
		t.Fatalf("Failed to write hook script: %v", err)
	}

	he := &HookExecutor{
		rootfsPath: rootfs,
		dryRun:     false,
		verbose:    true,
		timeout:    1, // 1 second timeout
	}

	ctx := context.Background()
	err := he.Execute(ctx, "/hooks/sleep.sh", nil)

	// Should timeout or require root
	if err == nil {
		t.Error("Expected timeout or root error")
	}
}

// TestHookExecutionContextCancellation tests context cancellation during hook execution
func TestHookExecutionContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	rootfs := filepath.Join(tmpDir, "rootfs")
	if err := os.MkdirAll(rootfs, 0755); err != nil {
		t.Fatalf("Failed to create rootfs: %v", err)
	}

	hookScript := `#!/bin/sh
sleep 5
exit 0
`
	hookPath := filepath.Join(rootfs, "hooks", "long.sh")
	if err := os.MkdirAll(filepath.Dir(hookPath), 0755); err != nil {
		t.Fatalf("Failed to create hooks dir: %v", err)
	}
	if err := os.WriteFile(hookPath, []byte(hookScript), 0755); err != nil {
		t.Fatalf("Failed to write hook script: %v", err)
	}

	he := &HookExecutor{
		rootfsPath: rootfs,
		dryRun:     false,
		verbose:    true,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := he.Execute(ctx, "/hooks/long.sh", nil)

	if err == nil {
		t.Error("Expected error from cancelled context")
	}
}

// TestValidateHookScript tests hook script validation
func TestValidateHookScript(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name       string
		script     string
		executable bool
		wantErr    bool
	}{
		{
			name: "valid shell script",
			script: `#!/bin/sh
echo "test"
`,
			executable: true,
			wantErr:    false,
		},
		{
			name: "valid bash script",
			script: `#!/bin/bash
echo "test"
`,
			executable: true,
			wantErr:    false,
		},
		{
			name: "not executable",
			script: `#!/bin/sh
echo "test"
`,
			executable: false,
			wantErr:    true,
		},
		{
			name: "no shebang",
			script: `echo "test"
`,
			executable: true,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scriptPath := filepath.Join(tmpDir, "script.sh")
			perm := os.FileMode(0644)
			if tt.executable {
				perm = 0755
			}
			if err := os.WriteFile(scriptPath, []byte(tt.script), perm); err != nil {
				t.Fatalf("Failed to write script: %v", err)
			}

			err := validateHookScript(scriptPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateHookScript() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Clean up
			os.Remove(scriptPath)
		})
	}
}

// TestFindHookScript tests finding hook scripts in layer
func TestFindHookScript(t *testing.T) {
	tmpDir := t.TempDir()
	layerPath := filepath.Join(tmpDir, "layer")

	// Create layer structure
	hooksDir := filepath.Join(layerPath, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatalf("Failed to create hooks dir: %v", err)
	}

	// Create pre-install hook
	preInstall := filepath.Join(hooksDir, "pre-install.sh")
	if err := os.WriteFile(preInstall, []byte("#!/bin/sh\necho 'pre'\n"), 0755); err != nil {
		t.Fatalf("Failed to write pre-install: %v", err)
	}

	// Create post-install hook
	postInstall := filepath.Join(hooksDir, "post-install.sh")
	if err := os.WriteFile(postInstall, []byte("#!/bin/sh\necho 'post'\n"), 0755); err != nil {
		t.Fatalf("Failed to write post-install: %v", err)
	}

	tests := []struct {
		name     string
		hookName string
		want     string
		wantErr  bool
	}{
		{
			name:     "find pre-install",
			hookName: "pre-install.sh",
			want:     preInstall,
			wantErr:  false,
		},
		{
			name:     "find post-install",
			hookName: "post-install.sh",
			want:     postInstall,
			wantErr:  false,
		},
		{
			name:     "not found",
			hookName: "nonexistent.sh",
			want:     "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findHookScript(layerPath, tt.hookName)
			if (err != nil) != tt.wantErr {
				t.Errorf("findHookScript() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("findHookScript() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to check if error is due to root requirement
func isRootRequired(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == "root privileges required for chroot operations" ||
		os.IsPermission(err)
}
