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

func TestSanitizeInline(t *testing.T) {
	// Single-line sanitizer: ESC/BEL/CR AND newline/tab are all stripped, so a
	// crafted value cannot inject escape sequences or spoof extra lines.
	in := "host\x1b[31mRED\x07\r\n fake line\ttab"
	got := sanitizeInline(in)
	if strings.ContainsAny(got, "\x1b\x07\r\n\t") {
		t.Errorf("inline sanitize must drop all control chars incl newline/tab: %q", got)
	}
	if got != "host[31mRED fake linetab" {
		t.Errorf("unexpected inline output: %q", got)
	}
}

func TestSanitizeBlock(t *testing.T) {
	// Block sanitizer: newlines preserved (YAML structure); ESC and tabs dropped
	// (tabs enable layout-spoofing and YAML indents with spaces).
	in := "image: a\x1b[31mb\ttab\nnetwork:\n  mode: nat\n"
	got := sanitizeBlock(in)
	if strings.ContainsRune(got, '\x1b') {
		t.Errorf("ESC should be stripped: %q", got)
	}
	if strings.ContainsRune(got, '\t') {
		t.Errorf("tab should be stripped: %q", got)
	}
	if strings.Count(got, "\n") != 3 {
		t.Errorf("newlines must be preserved for YAML: %q", got)
	}
}

func TestCliUseColorHonorsNanofuseNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("NANOFUSE_NO_COLOR", "1")
	if cliUseColor() {
		t.Error("cliUseColor() must be false when NANOFUSE_NO_COLOR is set")
	}
}

func TestSanitizersDropUnicodeWhitespace(t *testing.T) {
	// U+2028 line sep, U+2029 para sep, U+00A0 NBSP must be removed by both.
	in := "a b c d"
	if got := sanitizeInline(in); got != "abcd" {
		t.Errorf("sanitizeInline(%q) = %q, want abcd", in, got)
	}
	if got := sanitizeBlock(in); got != "abcd" {
		t.Errorf("sanitizeBlock(%q) = %q, want abcd", in, got)
	}
	if sanitizeInline("a b") != "a b" {
		t.Error("sanitizeInline should keep plain space")
	}
	if sanitizeBlock("a\nb c") != "a\nb c" {
		t.Error("sanitizeBlock should keep newline and space")
	}
}
