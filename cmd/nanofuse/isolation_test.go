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

func TestRunIsolationVerifyAbsentStoreNonStrictExitsZero(t *testing.T) {
	absent := filepath.Join(t.TempDir(), "no-such-store")
	withIsolationFlags(t, absent, false, false) // non-strict
	cmd := &cobra.Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)
	// Lenient: an absent store under non-strict is a skip, not a failure.
	if err := runIsolationVerify(cmd, nil); err != nil {
		t.Fatalf("non-strict + absent store should exit 0, got error: %v", err)
	}
	if !strings.Contains(out.String(), "credential isolation: NOT VERIFIED") {
		t.Errorf("output should report NOT VERIFIED; got: %s", out.String())
	}
}

func TestRunIsolationVerifyAbsentStoreStrictFails(t *testing.T) {
	absent := filepath.Join(t.TempDir(), "no-such-store")
	withIsolationFlags(t, absent, false, true) // strict
	cmd := &cobra.Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)
	if err := runIsolationVerify(cmd, nil); err == nil {
		t.Fatal("strict + absent store should return an error, got nil")
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

func TestRunIsolationVerifyRejectsWhitespacePaddedSecretsDir(t *testing.T) {
	// A padded value is a distinct path on Unix (" /x" is relative), so it is a
	// usage error: non-zero (returns err) with no status line printed.
	for _, dir := range []string{" /var/run/secrets/daax", "/var/run/secrets/daax "} {
		t.Run(dir, func(t *testing.T) {
			withIsolationFlags(t, dir, false, false)
			cmd := &cobra.Command{}
			var out bytes.Buffer
			cmd.SetOut(&out)
			err := runIsolationVerify(cmd, nil)
			if err == nil {
				t.Fatalf("expected an error for whitespace-padded --secrets-dir %q, got nil", dir)
			}
			if !strings.Contains(err.Error(), "whitespace") {
				t.Errorf("error = %v, want whitespace rejection for %q", err, dir)
			}
			if out.String() != "" {
				t.Errorf("a usage error must not print a status line; got: %s", out.String())
			}
		})
	}
}

func TestIsolationCommandsSkipAPIClientSetup(t *testing.T) {
	// A bad NANOFUSE_TIMEOUT makes applyClientEnvironment fail. The local-only
	// isolation command (and its verify subcommand) must skip API client setup,
	// so PersistentPreRunE returns nil despite the invalid env. If a regression
	// removed "isolation" from the skip set, this pre-run would surface the error.
	t.Setenv("NANOFUSE_TIMEOUT", "not-a-duration")
	for _, cmd := range []*cobra.Command{isolationCmd, isolationVerifyCmd} {
		if err := rootCmd.PersistentPreRunE(cmd, nil); err != nil {
			t.Errorf("PersistentPreRunE(%q) = %v, want nil (isolation must skip API client setup)", cmd.Name(), err)
		}
	}
}
