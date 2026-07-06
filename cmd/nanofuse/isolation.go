package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/daax-dev/nanofuse/internal/credisolation"
	"github.com/spf13/cobra"
)

var (
	isolationSecretsDir  string
	isolationRequireRoot bool
	isolationStrict      bool
)

var isolationCmd = &cobra.Command{
	Use:   "isolation",
	Short: "Inspect and verify per-microVM credential isolation",
	Long: "Verify the host-side invariants that keep each microVM's credentials\n" +
		"isolated: the credential store is root-only (0700, and root:root with\n" +
		"--require-root), and a policy self-check proves the mount guard denies\n" +
		"host/shared mounts over the store and admits a private tmpfs.",
}

var isolationVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify credential isolation and report status",
	Long: "Run the host-side credential-isolation checks and print a status line.\n\n" +
		"The final line is one of (process exit code in parentheses):\n" +
		"  credential isolation: PASS          all checks passed (exit 0)\n" +
		"  credential isolation: NOT VERIFIED  nothing was checked, e.g. the store\n" +
		"                                      is absent without --strict (exit 0)\n" +
		"  credential isolation: FAIL          a check failed (non-zero exit)\n\n" +
		"Run with --secrets-dir pointing at " + credisolation.GuestSecretsDir +
		" inside a guest (and --require-root as root) to verify the live store's\n" +
		"ownership and permissions. Use --strict to fail when the store is absent.",
	// A failed verification is a runtime result, not a usage mistake; do not
	// dump the command usage on error.
	SilenceUsage: true,
	RunE:         runIsolationVerify,
}

func init() {
	isolationVerifyCmd.Flags().StringVar(&isolationSecretsDir, "secrets-dir",
		credisolation.GuestSecretsDir, "credential store directory to verify")
	isolationVerifyCmd.Flags().BoolVar(&isolationRequireRoot, "require-root",
		false, "require root:root ownership of the credential store")
	isolationVerifyCmd.Flags().BoolVar(&isolationStrict, "strict",
		false, "fail if the credential store directory is absent")
	isolationCmd.AddCommand(isolationVerifyCmd)
	rootCmd.AddCommand(isolationCmd)
}

func runIsolationVerify(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()

	// Reject an empty or whitespace-only --secrets-dir (cobra permits it) rather
	// than silently treating it as an absent store and emitting empty-path
	// messages.
	if strings.TrimSpace(isolationSecretsDir) == "" {
		return fmt.Errorf("--secrets-dir must not be empty")
	}
	// Reject leading/trailing whitespace: on Unix " /x" and "/x " are paths
	// distinct from "/x" (and " /x" is relative), so an accidentally padded
	// value would silently verify the wrong location. Require the exact path.
	if isolationSecretsDir != strings.TrimSpace(isolationSecretsDir) {
		return fmt.Errorf("--secrets-dir %q has leading or trailing whitespace; pass the exact path", isolationSecretsDir)
	}

	opts := credisolation.HostCheckOptions{
		SecretsDir:  isolationSecretsDir,
		RequireRoot: isolationRequireRoot,
	}

	// Lstat (not Stat) so a symlinked store path is handed to the verifier,
	// which fails it, instead of being silently followed.
	switch _, err := os.Lstat(isolationSecretsDir); {
	case err == nil:
		opts.CheckDir = true
	case errors.Is(err, fs.ErrNotExist):
		// Genuinely absent: skip the perms check, or under --strict run it so the
		// store-perms check fails and a FAIL status line is printed (matching the
		// documented terminal states) rather than returning before any status line.
		if isolationStrict {
			opts.CheckDir = true
		} else {
			fmt.Fprintf(out, "  [skip] store-perms — %q not present on this host (use --strict to fail)\n",
				isolationSecretsDir)
		}
	default:
		// Permission denied, I/O error, etc. — do not misreport as "absent";
		// a verifier that cannot read the store has not verified it.
		return fmt.Errorf("cannot lstat credential store %q: %w", isolationSecretsDir, err)
	}

	report := credisolation.VerifyHost(opts)
	for _, res := range report.Results {
		status := "PASS"
		if !res.Pass {
			status = "FAIL"
		}
		fmt.Fprintf(out, "  [%s] %s — %s\n", status, res.Name, res.Detail)
	}
	fmt.Fprintln(out, report.StatusLine())

	// Exit non-zero only on an actual failed check. "Nothing verified" (the store
	// is legitimately absent and --strict was not set) is a lenient skip that
	// exits 0 — a CI health check on a host where the store is not provisioned
	// yet must not be treated as a failure. An absent store under --strict
	// already returned an error above.
	if report.HasFailure() {
		// Include the status line so stderr alone is actionable when stdout (the
		// per-check detail) is not captured.
		return fmt.Errorf("credential isolation verification failed: %s", report.StatusLine())
	}
	return nil
}
