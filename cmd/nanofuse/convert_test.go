package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// resetConvertFlags restores the convert command's package globals after a test.
func resetConvertFlags(t *testing.T) {
	t.Helper()
	prevLossy, prevResolve, prevOut := convertAllowLossy, convertResolveEgress, convertOutput
	t.Cleanup(func() {
		convertAllowLossy, convertResolveEgress, convertOutput = prevLossy, prevResolve, prevOut
	})
}

func TestConvertCommandRegistration(t *testing.T) {
	// convert gondolin must be wired as a subcommand.
	var gondolin *cobra.Command
	for _, c := range convertCmd.Commands() {
		if c.Name() == "gondolin" {
			gondolin = c
		}
	}
	if gondolin == nil {
		t.Fatal("convert gondolin subcommand not registered")
	}
	if gondolin.RunE == nil {
		t.Error("convert gondolin has no RunE")
	}
}

func TestConvertGondolin_FailsClosedOnBlockingFeature(t *testing.T) {
	resetConvertFlags(t)
	// env is a blocking divergence with no faithful nanofuse equivalent.
	yaml := "image: alpine\nenv:\n  API_KEY: secret\n"
	path := filepath.Join(t.TempDir(), "sb.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}
	convertAllowLossy = false
	err := convertGondolinCmd.RunE(convertGondolinCmd, []string{path})
	if err == nil {
		t.Fatal("expected fail-closed error for a blocking feature, got nil")
	}
	if !strings.Contains(err.Error(), "fail-closed") {
		t.Errorf("error = %v, want fail-closed", err)
	}
}

func TestConvertGondolin_AllowLossyWritesSpec(t *testing.T) {
	resetConvertFlags(t)
	yaml := "image: alpine\nenv:\n  API_KEY: secret\n"
	dir := t.TempDir()
	path := filepath.Join(dir, "sb.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "spec.yaml")
	convertAllowLossy = true
	convertOutput = out
	if err := convertGondolinCmd.RunE(convertGondolinCmd, []string{path}); err != nil {
		t.Fatalf("--allow-lossy should proceed, got: %v", err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("expected spec written to %s: %v", out, err)
	}
	if !strings.Contains(string(data), "image: alpine") {
		t.Errorf("rendered spec missing image; got:\n%s", data)
	}
}

func TestConvertGondolin_MissingFile(t *testing.T) {
	resetConvertFlags(t)
	err := convertGondolinCmd.RunE(convertGondolinCmd, []string{filepath.Join(t.TempDir(), "nope.yaml")})
	if err == nil {
		t.Fatal("expected an error for a missing file, got nil")
	}
	if !strings.Contains(err.Error(), "read gondolin mirror") {
		t.Errorf("error = %v, want read failure", err)
	}
}

func TestSanitizeForTerminal(t *testing.T) {
	// ANSI escape + bell + carriage return must be stripped; tab kept.
	in := "host\x1b[31mRED\x07\r end\tkeep"
	got := sanitizeForTerminal(in)
	if strings.ContainsAny(got, "\x1b\x07\r") {
		t.Errorf("control chars survived: %q", got)
	}
	if !strings.Contains(got, "\t") {
		t.Errorf("tab should be preserved: %q", got)
	}
	// The ESC/BEL/CR bytes are removed; the printable "[31m" text remains.
	if got != "host[31mRED end\tkeep" {
		t.Errorf("unexpected sanitized output: %q", got)
	}
}
