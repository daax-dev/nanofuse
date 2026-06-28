package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

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
	Long: "Verify the invariants that keep each microVM's credentials isolated:\n" +
		"the credential store is root-only (0700), no host/shared mount targets it,\n" +
		"and each VM has a distinct SPIFFE identity.",
}

var isolationVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify credential isolation and report status",
	Long: "Run the host-side credential-isolation checks and print a status line.\n\n" +
		"On success the final line is exactly:\n\n  credential isolation: PASS\n\n" +
		"Run with --secrets-dir pointing at " + credisolation.GuestSecretsDir +
		" inside a guest (and --require-root as root) to verify the live store's\n" +
		"ownership and permissions.",
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

	opts := credisolation.HostCheckOptions{
		SecretsDir:  isolationSecretsDir,
		RequireRoot: isolationRequireRoot,
	}

	switch _, err := os.Stat(isolationSecretsDir); {
	case err == nil:
		opts.CheckDir = true
	case errors.Is(err, fs.ErrNotExist):
		// Genuinely absent: skip the perms check (or fail under --strict).
		if isolationStrict {
			return fmt.Errorf("credential store %s is absent: %w", isolationSecretsDir, err)
		}
		fmt.Fprintf(out, "  [skip] store-perms — %s not present on this host (use --strict to fail)\n",
			isolationSecretsDir)
	default:
		// Permission denied, I/O error, etc. — do not misreport as "absent";
		// a verifier that cannot read the store has not verified it.
		return fmt.Errorf("cannot stat credential store %s: %w", isolationSecretsDir, err)
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

	if !report.Pass() {
		if report.Subjects == 0 {
			return fmt.Errorf("nothing concrete verified: run inside a guest (or pass --secrets-dir " +
				"to a live store) so the 0700 store contract can be checked")
		}
		return fmt.Errorf("credential isolation verification failed")
	}
	return nil
}
