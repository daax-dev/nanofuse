package layerbuild

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Common hook script names
const (
	PreInstallHook  = "pre-install.sh"
	PostInstallHook = "post-install.sh"
)

// Default hook execution timeout (5 minutes)
const defaultHookTimeout = 5 * time.Minute

var (
	// errRequiresRoot is returned when chroot operations require root privileges
	errRequiresRoot = errors.New("root privileges required for chroot operations")

	// errHookTimeout is returned when a hook exceeds its execution timeout
	errHookTimeout = errors.New("hook execution timeout")

	// errHookFailed is returned when a hook exits with non-zero status
	errHookFailed = errors.New("hook execution failed")
)

// HookExecutor executes hook scripts in a chroot environment.
type HookExecutor struct {
	rootfsPath string
	dryRun     bool
	verbose    bool
	timeout    int // timeout in seconds, 0 = default
}

// NewHookExecutor creates a new hook executor for the given rootfs path.
func NewHookExecutor(rootfsPath string, dryRun, verbose bool) *HookExecutor {
	return &HookExecutor{
		rootfsPath: rootfsPath,
		dryRun:     dryRun,
		verbose:    verbose,
		timeout:    0, // use default
	}
}

// Execute runs a hook script in the chroot environment.
// The scriptPath is relative to the rootfs (e.g., "/hooks/pre-install.sh").
// Environment variables can be passed via the env map.
func (he *HookExecutor) Execute(ctx context.Context, scriptPath string, env map[string]string) error {
	// Check if hook exists in rootfs
	fullPath := filepath.Join(he.rootfsPath, scriptPath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		if he.verbose {
			fmt.Printf("Hook %s not found, skipping\n", scriptPath)
		}
		return nil // Not an error - hooks are optional
	}

	// Validate hook script
	if err := validateHookScript(fullPath); err != nil {
		return fmt.Errorf("invalid hook script %s: %w", scriptPath, err)
	}

	if he.verbose {
		fmt.Printf("Executing hook: %s\n", scriptPath)
	}

	if he.dryRun {
		if he.verbose {
			fmt.Printf("[DRY-RUN] Would execute hook: chroot %s %s\n", he.rootfsPath, scriptPath)
		}
		return nil
	}

	// Check if we're running as root
	if os.Geteuid() != 0 {
		return errRequiresRoot
	}

	// Determine timeout
	timeout := defaultHookTimeout
	if he.timeout > 0 {
		timeout = time.Duration(he.timeout) * time.Second
	}

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Build command: chroot <rootfs> <script>
	// #nosec G204 -- rootfsPath and scriptPath are validated internal paths from layer extraction
	cmd := exec.CommandContext(execCtx, "chroot", he.rootfsPath, scriptPath)

	// Set environment variables
	cmd.Env = buildHookEnv(env)

	// Capture output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start hook: %w", err)
	}

	// Stream output if verbose
	if he.verbose {
		go streamOutput(stdout, "[hook stdout]")
		go streamOutput(stderr, "[hook stderr]")
	}

	// Wait for completion
	err = cmd.Wait()

	// Check for timeout
	if execCtx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("%w: %s exceeded %v", errHookTimeout, scriptPath, timeout)
	}

	// Check for cancellation
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Check for hook failure
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("%w: %s exited with status %d", errHookFailed, scriptPath, exitErr.ExitCode())
		}
		return fmt.Errorf("hook execution error: %w", err)
	}

	if he.verbose {
		fmt.Printf("Hook %s completed successfully\n", scriptPath)
	}

	return nil
}

// ExecuteLayerHooks executes pre-install and post-install hooks for a layer.
// The layerPath should be the extracted layer directory containing a hooks/ subdirectory.
func (he *HookExecutor) ExecuteLayerHooks(ctx context.Context, layerPath, layerName string, env map[string]string, phase string) error {
	var hookName string
	switch phase {
	case "pre-install":
		hookName = PreInstallHook
	case "post-install":
		hookName = PostInstallHook
	default:
		return fmt.Errorf("invalid hook phase: %s", phase)
	}

	// Find hook in layer
	hookPath, err := findHookScript(layerPath, hookName)
	if err != nil {
		// Hook not found - this is OK, hooks are optional
		if he.verbose {
			fmt.Printf("Layer %s has no %s hook\n", layerName, hookName)
		}
		return nil
	}

	// Copy hook to rootfs hooks directory
	rootfsHookPath := filepath.Join("/hooks", layerName, hookName)
	fullRootfsHookPath := filepath.Join(he.rootfsPath, rootfsHookPath)

	if err := os.MkdirAll(filepath.Dir(fullRootfsHookPath), 0755); err != nil {
		return fmt.Errorf("failed to create hooks directory: %w", err)
	}

	if !he.dryRun {
		if err := copyFile(hookPath, fullRootfsHookPath); err != nil {
			return fmt.Errorf("failed to copy hook to rootfs: %w", err)
		}

		// Make executable
		if err := os.Chmod(fullRootfsHookPath, 0755); err != nil {
			return fmt.Errorf("failed to set hook permissions: %w", err)
		}
	}

	// Execute the hook
	return he.Execute(ctx, rootfsHookPath, env)
}

// validateHookScript validates that a hook script is executable and has a shebang.
func validateHookScript(path string) error {
	// Check if file is executable
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.Mode()&0111 == 0 {
		return fmt.Errorf("script is not executable")
	}

	// Read first line to check for shebang
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return fmt.Errorf("empty script file")
	}

	firstLine := scanner.Text()
	if !strings.HasPrefix(firstLine, "#!") {
		return fmt.Errorf("script missing shebang (#!) line")
	}

	return nil
}

// findHookScript searches for a hook script in a layer directory.
func findHookScript(layerPath, hookName string) (string, error) {
	// Check in hooks/ subdirectory
	hookPath := filepath.Join(layerPath, "hooks", hookName)
	if _, err := os.Stat(hookPath); err == nil {
		return hookPath, nil
	}

	// Check in root of layer
	hookPath = filepath.Join(layerPath, hookName)
	if _, err := os.Stat(hookPath); err == nil {
		return hookPath, nil
	}

	return "", fmt.Errorf("hook %s not found in layer %s", hookName, layerPath)
}

// buildHookEnv builds the environment variable list for hook execution.
func buildHookEnv(env map[string]string) []string {
	// Start with minimal safe environment - preallocate with capacity for 3 base vars + custom env
	baseEnv := make([]string, 0, 3+len(env))
	baseEnv = append(baseEnv,
		"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		"HOME=/root",
		"TERM=xterm",
	)

	// Add custom environment variables
	for key, value := range env {
		baseEnv = append(baseEnv, fmt.Sprintf("%s=%s", key, value))
	}

	return baseEnv
}

// streamOutput reads from a pipe and prints lines with a prefix.
func streamOutput(pipe interface{ Read([]byte) (int, error) }, prefix string) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		fmt.Printf("%s %s\n", prefix, scanner.Text())
	}
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := dstFile.ReadFrom(srcFile); err != nil {
		return err
	}

	// Copy permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())
}
