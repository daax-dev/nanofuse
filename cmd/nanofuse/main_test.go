package main

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/daax-dev/nanofuse/internal/clierrors"
)

// Basic smoke test to ensure package builds
func TestPackageBuilds(t *testing.T) {
	// If this test runs, the package compiled successfully
	t.Log("nanofuse CLI package builds successfully")
}

// TestVMLogsCommandExists verifies the vm logs command exists
func TestVMLogsCommandExists(t *testing.T) {
	if vmLogsCmd == nil {
		t.Fatal("vmLogsCmd is nil")
	}

	if vmLogsCmd.Use != "logs <vm-id>" {
		t.Errorf("Expected Use to be 'logs <vm-id>', got '%s'", vmLogsCmd.Use)
	}
}

// TestVMLogsCommandFlags verifies required flags exist
func TestVMLogsCommandFlags(t *testing.T) {
	flags := vmLogsCmd.Flags()

	// Check --tail flag exists
	tailFlag := flags.Lookup("tail")
	if tailFlag == nil {
		t.Fatal("--tail flag not found")
	}

	// Check --lines/-n flag exists
	linesFlag := flags.Lookup("lines")
	if linesFlag == nil {
		t.Fatal("--lines flag not found")
	}
	if linesFlag.Shorthand != "n" {
		t.Errorf("Expected shorthand 'n' for --lines, got '%s'", linesFlag.Shorthand)
	}

	// Check --follow/-f flag exists
	followFlag := flags.Lookup("follow")
	if followFlag == nil {
		t.Fatal("--follow flag not found")
	}
	if followFlag.Shorthand != "f" {
		t.Errorf("Expected shorthand 'f' for --follow, got '%s'", followFlag.Shorthand)
	}
}

func TestApplyClientEnvironment(t *testing.T) {
	resetCLIStateForTest(t)
	t.Setenv("NANOFUSE_API_URL", "http://127.0.0.1:18080")
	t.Setenv("NANOFUSE_API_SOCKET", "/tmp/nanofused.sock")
	t.Setenv("NANOFUSE_TIMEOUT", "5s")
	t.Setenv("NANOFUSE_DEBUG", "true")
	t.Setenv("NANOFUSE_OUTPUT", "json")
	t.Setenv("NANOFUSE_NO_COLOR", "1")

	if err := applyClientEnvironment(); err != nil {
		t.Fatalf("applyClientEnvironment() error = %v", err)
	}

	if apiURL != "http://127.0.0.1:18080" {
		t.Errorf("expected API URL from env, got %q", apiURL)
	}
	if apiSocket != "/tmp/nanofused.sock" {
		t.Errorf("expected API socket from env, got %q", apiSocket)
	}
	if timeout != 5*time.Second {
		t.Errorf("expected timeout 5s, got %s", timeout)
	}
	if !debug {
		t.Error("expected debug from env")
	}
	if !jsonOutput {
		t.Error("expected json output from env")
	}
	if !noColor {
		t.Error("expected no-color from env")
	}
}

func TestApplyClientEnvironmentDoesNotOverrideExplicitValues(t *testing.T) {
	resetCLIStateForTest(t)
	apiURL = "http://explicit:8080"
	apiSocket = "/explicit.sock"
	timeout = 2 * time.Second
	t.Setenv("NANOFUSE_API_URL", "http://env:8080")
	t.Setenv("NANOFUSE_API_SOCKET", "/env.sock")
	t.Setenv("NANOFUSE_TIMEOUT", "5s")

	if err := applyClientEnvironment(); err != nil {
		t.Fatalf("applyClientEnvironment() error = %v", err)
	}

	if apiURL != "http://explicit:8080" {
		t.Errorf("expected explicit API URL to remain, got %q", apiURL)
	}
	if apiSocket != "/explicit.sock" {
		t.Errorf("expected explicit API socket to remain, got %q", apiSocket)
	}
	if timeout != 2*time.Second {
		t.Errorf("expected explicit timeout to remain, got %s", timeout)
	}
}

func TestApplyClientEnvironmentRejectsInvalidTimeout(t *testing.T) {
	resetCLIStateForTest(t)
	t.Setenv("NANOFUSE_TIMEOUT", "not-a-duration")

	if err := applyClientEnvironment(); err == nil {
		t.Fatal("expected invalid timeout error")
	}
}

func TestHandleAPIErrorUsesAPIURLForTCPConnection(t *testing.T) {
	resetCLIStateForTest(t)
	apiURL = "http://127.0.0.1:18080"

	err := handleAPIError(errors.New("Get \"http://127.0.0.1:18080/health\": dial tcp 127.0.0.1:18080: connect: connection refused"), "check API health")
	cliErr, ok := err.(*clierrors.CLIError)
	if !ok {
		t.Fatalf("handleAPIError() = %T, want *clierrors.CLIError", err)
	}
	if cliErr.Context == nil {
		t.Fatal("expected error context")
	}
	if cliErr.Context.Resource != apiURL {
		t.Fatalf("Resource = %q, want %q", cliErr.Context.Resource, apiURL)
	}
	if !strings.Contains(cliErr.Suggestion, "SSH tunnel") {
		t.Fatalf("Suggestion = %q, want SSH tunnel guidance", cliErr.Suggestion)
	}
}

func TestHandleAPIErrorUsesDefaultSocketWhenEndpointUnset(t *testing.T) {
	resetCLIStateForTest(t)

	err := handleAPIError(errors.New("dial unix /var/run/nanofused.sock: connect: no such file or directory"), "check API health")
	cliErr, ok := err.(*clierrors.CLIError)
	if !ok {
		t.Fatalf("handleAPIError() = %T, want *clierrors.CLIError", err)
	}
	if cliErr.Context == nil {
		t.Fatal("expected error context")
	}
	if cliErr.Context.Resource != DefaultAPISocketPath {
		t.Fatalf("Resource = %q, want %q", cliErr.Context.Resource, DefaultAPISocketPath)
	}
}

func TestRootCommandDoesNotSilenceUsageByDefault(t *testing.T) {
	resetCLIStateForTest(t)

	if rootCmd.SilenceUsage {
		t.Fatal("root command should not silence Cobra usage by default")
	}
	if rootCmd.SilenceErrors {
		t.Fatal("root command should not silence Cobra errors by default")
	}
}

func TestFormattedCLIErrorSilencesCobraOutput(t *testing.T) {
	resetCLIStateForTest(t)

	err := formattedCLIError(&clierrors.CLIError{Message: "formatted failure", ExitCode: 1})
	if err == nil {
		t.Fatal("formattedCLIError() returned nil")
	}
	if !rootCmd.SilenceUsage {
		t.Fatal("formatted CLI errors should silence Cobra usage")
	}
	if !rootCmd.SilenceErrors {
		t.Fatal("formatted CLI errors should silence Cobra errors")
	}
}

func resetCLIStateForTest(t *testing.T) {
	t.Helper()

	oldCfgFile := cfgFile
	oldAPISocket := apiSocket
	oldAPIURL := apiURL
	oldDebug := debug
	oldJSONOutput := jsonOutput
	oldNoColor := noColor
	oldTimeout := timeout
	oldSilenceUsage := rootCmd.SilenceUsage
	oldSilenceErrors := rootCmd.SilenceErrors

	cfgFile = ""
	apiSocket = ""
	apiURL = ""
	debug = false
	jsonOutput = false
	noColor = false
	timeout = 30 * time.Second
	rootCmd.SilenceUsage = false
	rootCmd.SilenceErrors = false

	t.Cleanup(func() {
		cfgFile = oldCfgFile
		apiSocket = oldAPISocket
		apiURL = oldAPIURL
		debug = oldDebug
		jsonOutput = oldJSONOutput
		noColor = oldNoColor
		timeout = oldTimeout
		rootCmd.SilenceUsage = oldSilenceUsage
		rootCmd.SilenceErrors = oldSilenceErrors
	})
}
