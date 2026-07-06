package firecracker

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// writeStubBinary writes an executable shell stub that prints the given stdout.
func writeStubBinary(t *testing.T, dir, name, stdout string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("shell stub not supported on windows")
	}
	p := filepath.Join(dir, name)
	// Emit stdout literally via a quoted heredoc (no shell interpolation).
	content := "#!/bin/sh\ncat <<'EOF'\n" + stdout + "\nEOF\n"
	if err := os.WriteFile(p, []byte(content), 0o700); err != nil { // #nosec G306 -- test stub must be executable.
		t.Fatalf("write stub: %v", err)
	}
	return p
}

func TestBinaryVersionParsesFirecrackerPrefix(t *testing.T) {
	bin := writeStubBinary(t, t.TempDir(), "firecracker", "Firecracker v1.7.0")
	got, err := BinaryVersion(bin)
	if err != nil {
		t.Fatalf("BinaryVersion: %v", err)
	}
	if got != "v1.7.0" {
		t.Fatalf("version = %q, want v1.7.0", got)
	}
}

func TestBinaryVersionMultiLine(t *testing.T) {
	bin := writeStubBinary(t, t.TempDir(), "firecracker", "v1.8.0\nsupported api version: 1.0.0")
	got, err := BinaryVersion(bin)
	if err != nil {
		t.Fatalf("BinaryVersion: %v", err)
	}
	if got != "v1.8.0" {
		t.Fatalf("version = %q, want v1.8.0 (first line only)", got)
	}
}

func TestBinaryVersionErrors(t *testing.T) {
	if _, err := BinaryVersion(""); err == nil {
		t.Fatal("empty path should error")
	}
	if _, err := BinaryVersion(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Fatal("missing binary should error")
	}
}
