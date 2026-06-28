package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// withIsolationFlags sets the package-global isolation flags for the duration of
// a test and restores them afterward.
func withIsolationFlags(t *testing.T, dir string, requireRoot, strict bool) {
	t.Helper()
	prevDir, prevRoot, prevStrict := isolationSecretsDir, isolationRequireRoot, isolationStrict
	isolationSecretsDir, isolationRequireRoot, isolationStrict = dir, requireRoot, strict
	t.Cleanup(func() {
		isolationSecretsDir, isolationRequireRoot, isolationStrict = prevDir, prevRoot, prevStrict
	})
}

func TestRunIsolationVerifyRejectsEmptySecretsDir(t *testing.T) {
	withIsolationFlags(t, "   ", false, false)
	cmd := &cobra.Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)
	if err := runIsolationVerify(cmd, nil); err == nil {
		t.Fatal("expected an error for an empty --secrets-dir, got nil")
	} else if !strings.Contains(err.Error(), "--secrets-dir must not be empty") {
		t.Errorf("error = %v, want it to reject the empty --secrets-dir", err)
	}
}

func TestRunIsolationVerifyPassesOnGoodStore(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "secrets")
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	withIsolationFlags(t, dir, false, true)
	cmd := &cobra.Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)
	if err := runIsolationVerify(cmd, nil); err != nil {
		t.Fatalf("runIsolationVerify on a 0700 store = %v, want nil", err)
	}
	if !strings.Contains(out.String(), "credential isolation: PASS") {
		t.Errorf("output missing the PASS status line; got: %s", out.String())
	}
}
