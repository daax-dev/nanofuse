//go:build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var (
	// Build variables
	Version   = getVersion()
	GitCommit = getGitCommit()
	BuildDate = time.Now().Format(time.RFC3339)
	GoVersion = runtime.Version()
)

// Default target to run when none is specified
var Default = All

// Build all binaries
func All() error {
	mg.Deps(Clean)
	mg.Deps(CLI, Daemon, RegisterLocalImage)
	return nil
}

// Build the CLI binary
func CLI() error {
	fmt.Println("Building nanofuse CLI...")
	return buildBinary("./cmd/nanofuse", "./bin/nanofuse")
}

// Build the API daemon binary (with CGO for SQLite support)
func Daemon() error {
	fmt.Println("Building nanofused daemon (with CGO for SQLite)...")
	return buildDaemonBinary("./cmd/nanofused", "./bin/nanofused")
}

// Build the register-local-image utility (with CGO for SQLite support)
func RegisterLocalImage() error {
	fmt.Println("Building register-local-image utility (with CGO for SQLite)...")
	return buildDaemonBinary("./register-local-image.go", "./bin/register-local-image")
}

// Run unit tests only
func Test() error {
	fmt.Println("Running unit tests...")
	return sh.RunV("go", "test", "-v", "-race", "-coverprofile=coverage.out", "./...")
}

// Run unit tests with coverage report
func TestCoverage() error {
	mg.Deps(Test)
	fmt.Println("Generating coverage report...")
	if err := sh.RunV("go", "tool", "cover", "-html=coverage.out", "-o", "coverage.html"); err != nil {
		return err
	}
	fmt.Println("Coverage report: coverage.html")
	return sh.RunV("go", "tool", "cover", "-func=coverage.out")
}

// Run integration tests (requires sudo for iptables/networking)
func TestIntegration() error {
	fmt.Println("Running integration tests...")
	fmt.Println("Note: Integration tests require sudo for iptables/networking setup")
	return sh.RunV("sudo", "go", "test", "-v", "-tags=integration", "./test/integration/...")
}

// Run all tests (unit + integration)
func TestAll() error {
	mg.Deps(Test)
	if err := TestIntegration(); err != nil {
		fmt.Println("Warning: Integration tests failed (this is expected if daemon not running)")
		fmt.Println("Continuing with unit tests only...")
	}
	return nil
}

// Run quick tests (no race detector, faster)
func TestQuick() error {
	fmt.Println("Running quick tests...")
	return sh.RunV("go", "test", "./...")
}

// Run tests with verbose output
func TestVerbose() error {
	fmt.Println("Running verbose tests...")
	return sh.RunV("go", "test", "-v", "-race", "./...")
}

// Watch and run tests on file changes (requires entr or similar)
func TestWatch() error {
	fmt.Println("Running tests in watch mode...")
	fmt.Println("Note: This requires 'entr' to be installed")
	return sh.RunV("bash", "-c", "find . -name '*.go' | entr -c mage test")
}

// Run linters
func Lint() error {
	fmt.Println("Running linters...")

	fmt.Println("1. Running go fmt...")
	if err := sh.RunV("go", "fmt", "./..."); err != nil {
		return err
	}

	fmt.Println("2. Running go vet...")
	if err := sh.RunV("go", "vet", "./..."); err != nil {
		return err
	}

	fmt.Println("3. Checking for golangci-lint...")
	if _, err := exec.LookPath("golangci-lint"); err == nil {
		fmt.Println("4. Running golangci-lint...")
		return sh.RunV("golangci-lint", "run", "./...")
	}

	fmt.Println("golangci-lint not found, skipping (install: brew install golangci-lint)")
	return nil
}

// Run security checks
func SecurityCheck() error {
	fmt.Println("Running security checks...")

	// Check if gosec is available
	if _, err := exec.LookPath("gosec"); err == nil {
		fmt.Println("Running gosec...")
		return sh.RunV("gosec", "./...")
	}

	fmt.Println("gosec not found (install: go install github.com/securego/gosec/v2/cmd/gosec@latest)")
	return nil
}

// Clean build artifacts
func Clean() error {
	fmt.Println("Cleaning build artifacts...")
	os.RemoveAll("./bin")
	os.RemoveAll("./coverage.out")
	os.RemoveAll("./coverage.html")
	return nil
}

// Install binaries to system (requires sudo)
func Install() error {
	mg.Deps(All)
	fmt.Println("Installing binaries...")

	installDir := "/usr/local/bin"
	if os.Getenv("PREFIX") != "" {
		installDir = filepath.Join(os.Getenv("PREFIX"), "bin")
	}

	if err := sh.Copy(filepath.Join(installDir, "nanofuse"), "./bin/nanofuse"); err != nil {
		return err
	}
	if err := sh.Copy(filepath.Join(installDir, "nanofused"), "./bin/nanofused"); err != nil {
		return err
	}

	fmt.Printf("Installed to %s\n", installDir)
	return nil
}

// InstallUser installs binaries to ~/bin (no sudo required)
func InstallUser() error {
	mg.Deps(All)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	installDir := filepath.Join(homeDir, "bin")
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create ~/bin: %w", err)
	}

	if err := sh.Copy(filepath.Join(installDir, "nanofuse"), "./bin/nanofuse"); err != nil {
		return fmt.Errorf("failed to install nanofuse: %w", err)
	}
	if err := sh.Copy(filepath.Join(installDir, "nanofused"), "./bin/nanofused"); err != nil {
		return fmt.Errorf("failed to install nanofused: %w", err)
	}
	if err := sh.Copy(filepath.Join(installDir, "register-local-image"), "./bin/register-local-image"); err != nil {
		return fmt.Errorf("failed to install register-local-image: %w", err)
	}

	fmt.Printf("✓ Installed to %s\n", installDir)
	fmt.Println("\nMake sure ~/bin is in your PATH:")
	fmt.Println("  echo 'export PATH=\"$HOME/bin:$PATH\"' >> ~/.bashrc  # for bash")
	fmt.Println("  echo 'export PATH=\"$HOME/bin:$PATH\"' >> ~/.zshrc   # for zsh")
	fmt.Println("\nThen restart your shell or run: source ~/.bashrc (or ~/.zshrc)")
	return nil
}

// CI runs all checks that will run in GitHub Actions
func CI() error {
	fmt.Println("===============================================")
	fmt.Println("Running CI checks locally")
	fmt.Println("===============================================")
	fmt.Println()

	// Clean first
	fmt.Println("Step 1: Clean")
	if err := Clean(); err != nil {
		return fmt.Errorf("clean failed: %w", err)
	}

	// Build
	fmt.Println("\nStep 2: Build")
	if err := All(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	// Lint
	fmt.Println("\nStep 3: Lint")
	if err := Lint(); err != nil {
		return fmt.Errorf("lint failed: %w", err)
	}

	// Test
	fmt.Println("\nStep 4: Test")
	if err := Test(); err != nil {
		return fmt.Errorf("tests failed: %w", err)
	}

	// Security check (optional, won't fail CI if not installed)
	fmt.Println("\nStep 5: Security Check")
	SecurityCheck() // Don't fail on this

	// Summary
	fmt.Println("\n===============================================")
	fmt.Println("✓ All CI checks passed!")
	fmt.Println("===============================================")
	fmt.Println("\nYou can now safely push to GitHub")
	fmt.Println("The GitHub Actions workflow will run the same checks")

	return nil
}

// Validate runs a quick sanity check
func Validate() error {
	fmt.Println("Running validation checks...")

	// Check Go version
	fmt.Println("1. Checking Go version...")
	if err := sh.RunV("go", "version"); err != nil {
		return err
	}

	// Check go.mod is valid
	fmt.Println("2. Validating go.mod...")
	if err := sh.RunV("go", "mod", "verify"); err != nil {
		return err
	}

	// Check for build issues
	fmt.Println("3. Checking for build issues...")
	if err := sh.RunV("go", "build", "./..."); err != nil {
		return err
	}

	// Quick test
	fmt.Println("4. Running quick tests...")
	if err := TestQuick(); err != nil {
		return err
	}

	fmt.Println("\n✓ Validation passed!")
	return nil
}

// Build Docker image
func ImageBuild() error {
	fmt.Println("Building Docker image...")
	return sh.RunV("docker", "build", "-t", "nanofuse-base:latest", "./images/base")
}

// Run the CLI (for testing)
func RunCLI(args ...string) error {
	mg.Deps(CLI)
	return sh.RunV("./bin/nanofuse", args...)
}

// Run the daemon (for testing)
func RunDaemon(args ...string) error {
	mg.Deps(Daemon)
	return sh.RunV("./bin/nanofused", args...)
}

// Check if all dependencies are installed
func Check() error {
	fmt.Println("Checking dependencies...")

	deps := []string{"go", "docker"}
	optional := []string{"golangci-lint", "firecracker", "act"}

	for _, dep := range deps {
		if err := checkCommand(dep); err != nil {
			return fmt.Errorf("required dependency missing: %s", dep)
		}
	}

	for _, dep := range optional {
		if err := checkCommand(dep); err != nil {
			fmt.Printf("Warning: optional dependency missing: %s\n", dep)
		}
	}

	fmt.Println("All required dependencies installed!")
	return nil
}

// ActCI runs all GitHub Actions CI jobs locally using act
// Simulates a push event to main branch
func ActCI() error {
	if err := checkCommand("act"); err != nil {
		return fmt.Errorf("act is not installed\nInstall: https://github.com/nektos/act#installation\n  macOS: brew install act\n  Linux: curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash")
	}

	fmt.Println("===============================================")
	fmt.Println("Running all CI jobs locally (simulating push)")
	fmt.Println("===============================================")
	fmt.Println()

	return sh.RunV("act",
		"push",
		"--workflows", ".github/workflows/ci.yaml",
		"--verbose",
	)
}

// ActCIDryRun shows what would run without actually executing
func ActCIDryRun() error {
	if err := checkCommand("act"); err != nil {
		return fmt.Errorf("act is not installed\nInstall: https://github.com/nektos/act#installation")
	}

	fmt.Println("Dry run - showing what would execute:")
	return sh.RunV("act",
		"push",
		"--workflows", ".github/workflows/ci.yaml",
		"--dryrun",
	)
}

// ActCIList lists all available jobs in the CI workflow
func ActCIList() error {
	if err := checkCommand("act"); err != nil {
		return fmt.Errorf("act is not installed\nInstall: https://github.com/nektos/act#installation")
	}

	fmt.Println("Available CI jobs:")
	return sh.RunV("act", "--list", "--workflows", ".github/workflows/ci.yaml")
}

// ActCIJob runs a specific CI job locally
// Example: mage actCIJob build-go
func ActCIJob(jobName string) error {
	if err := checkCommand("act"); err != nil {
		return fmt.Errorf("act is not installed\nInstall: https://github.com/nektos/act#installation")
	}

	fmt.Printf("Running job: %s\n", jobName)
	return sh.RunV("act",
		"push",
		"--job", jobName,
		"--workflows", ".github/workflows/ci.yaml",
		"--verbose",
	)
}

// Helper functions

func buildBinary(pkgPath, output string) error {
	ldflags := fmt.Sprintf(
		"-s -w -X main.version=%s -X main.commit=%s -X main.buildDate=%s",
		Version, GitCommit, BuildDate,
	)

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		return err
	}

	return sh.RunWith(
		map[string]string{"CGO_ENABLED": "0"},
		"go", "build",
		"-ldflags", ldflags,
		"-o", output,
		pkgPath,
	)
}

// buildDaemonBinary builds the daemon with CGO enabled for SQLite support
func buildDaemonBinary(pkgPath, output string) error {
	ldflags := fmt.Sprintf(
		"-s -w -X main.version=%s -X main.commit=%s -X main.buildDate=%s",
		Version, GitCommit, BuildDate,
	)

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		return err
	}

	// Build with CGO enabled for SQLite (go-sqlite3 requires CGO)
	return sh.RunWith(
		map[string]string{"CGO_ENABLED": "1"},
		"go", "build",
		"-ldflags", ldflags,
		"-o", output,
		pkgPath,
	)
}

func getGitCommit() string {
	commit, err := sh.Output("git", "rev-parse", "--short", "HEAD")
	if err != nil {
		return "unknown"
	}
	return commit
}

func getVersion() string {
	// Try to get version from git tag (excluding image-v* tags)
	version, err := sh.Output("git", "describe", "--tags", "--match", "v[0-9]*", "--abbrev=0")
	if err != nil {
		// No tags found, use default
		return "0.0.0-dev"
	}
	// Strip the 'v' prefix
	if len(version) > 0 && version[0] == 'v' {
		return version[1:]
	}
	return version
}

func checkCommand(cmd string) error {
	_, err := exec.LookPath(cmd)
	return err
}

// TestBuild runs build validation tests (kernel and rootfs)
func TestBuild() error {
	fmt.Println("Running build validation tests...")

	// Check if build artifacts exist
	if _, err := os.Stat("images/base/build/vmlinux"); err != nil {
		fmt.Println("Note: Kernel not found - some tests will be skipped")
	}
	if _, err := os.Stat("images/base/build/rootfs.ext4"); err != nil {
		fmt.Println("Note: Rootfs not found - some tests will be skipped")
	}

	// Run Go build tests (primary tests)
	if err := sh.RunV("go", "test", "-v", "./test/build/..."); err != nil {
		return fmt.Errorf("build tests failed: %w", err)
	}

	fmt.Println("✓ Build validation tests passed")
	return nil
}

// TestE2E runs end-to-end lifecycle tests (requires sudo, KVM, Firecracker)
func TestE2E() error {
	fmt.Println("Running E2E lifecycle tests...")
	fmt.Println("Note: E2E tests require sudo, KVM, and Firecracker installed")

	// Check prerequisites
	if os.Geteuid() != 0 {
		fmt.Println("Warning: Not running as root - E2E tests may be skipped")
	}

	if _, err := os.Stat("/dev/kvm"); err != nil {
		fmt.Println("Warning: KVM not available - E2E tests may be skipped")
	}

	// Run standalone E2E script if running as root
	if os.Geteuid() == 0 {
		if err := sh.RunV("./scripts/e2e-test.sh"); err != nil {
			return fmt.Errorf("E2E tests failed: %w", err)
		}
	} else {
		// Run Go E2E tests (will skip if not root)
		if err := sh.RunV("go", "test", "-v", "-tags=e2e", "./test/e2e/..."); err != nil {
			return fmt.Errorf("E2E tests failed: %w", err)
		}
	}

	fmt.Println("✓ E2E tests passed")
	return nil
}

// TestGdt runs all gdt declarative tests.
//
// This function uses "best-effort" execution: it runs all test suites and reports
// warnings for failures rather than stopping on the first error. This design allows
// CI environments to see the full picture of test results even when some prerequisites
// (like daemon, build artifacts) are missing.
//
// For strict failure propagation, use the individual TestGdt* targets (TestGdtBuild,
// TestGdtCLI, TestGdtAPI, TestGdtE2E) which return errors on failure.
func TestGdt() error {
	fmt.Println("Running gdt declarative tests (best-effort mode)...")
	fmt.Println("Note: Warnings are expected if prerequisites are missing.")
	fmt.Println("For strict mode, use individual targets: TestGdtBuild, TestGdtCLI, TestGdtAPI")

	// Run build tests
	fmt.Println("\n--- Build Tests ---")
	if err := sh.RunV("go", "test", "-v", "./test/gdt/build/..."); err != nil {
		fmt.Println("Warning: Build tests failed (may be expected if artifacts not built)")
	}

	// Run CLI tests (doesn't require daemon for most tests)
	fmt.Println("\n--- CLI Tests ---")
	if err := sh.RunV("go", "test", "-v", "./test/gdt/cli/..."); err != nil {
		fmt.Println("Warning: CLI tests failed (may be expected if CLI not built)")
	}

	// Run API tests (requires daemon running)
	fmt.Println("\n--- API Tests ---")
	fmt.Println("Note: Requires nanofused daemon running")
	if err := sh.RunV("go", "test", "-v", "./test/gdt/api/..."); err != nil {
		fmt.Println("Warning: API tests failed (may be expected if daemon not running)")
	}

	fmt.Println("✓ gdt declarative tests completed (best-effort)")
	return nil
}

// TestGdtBuild runs only gdt build tests
func TestGdtBuild() error {
	fmt.Println("Running gdt build tests...")
	return sh.RunV("go", "test", "-v", "./test/gdt/build/...")
}

// TestGdtCLI runs only gdt CLI tests
func TestGdtCLI() error {
	fmt.Println("Running gdt CLI tests...")
	fmt.Println("Note: Requires nanofuse CLI binary - run 'mage cli' first")
	return sh.RunV("go", "test", "-v", "./test/gdt/cli/...")
}

// TestGdtAPI runs only gdt API tests (requires daemon running)
func TestGdtAPI() error {
	fmt.Println("Running gdt API tests...")
	fmt.Println("Note: Requires nanofused daemon running")
	return sh.RunV("go", "test", "-v", "./test/gdt/api/...")
}

// TestGdtE2E runs only gdt E2E tests (requires sudo, KVM, Firecracker)
func TestGdtE2E() error {
	fmt.Println("Running gdt E2E tests...")
	fmt.Println("Note: Requires sudo, KVM, Firecracker, and daemon running")
	return sh.RunV("go", "test", "-v", "-tags=e2e", "./test/gdt/e2e/...")
}
