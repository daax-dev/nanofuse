package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/daax-dev/nanofuse/internal/gondolin"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// sanitizeInline removes anything that could inject terminal escapes or spoof
// layout from a value printed on a single line: all control characters AND all
// whitespace other than the plain ASCII space. That covers non-control unicode
// whitespace such as U+2028/U+2029 (line/paragraph separators) and U+00A0 (NBSP)
// which a control-only filter would miss. Used for host names, paths, etc.
func sanitizeInline(s string) string {
	return strings.Map(func(r rune) rune {
		if r == ' ' {
			return r
		}
		if unicode.IsControl(r) || unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
}

// sanitizeBlock is like sanitizeInline but PRESERVES the plain newline, for
// multi-line output whose line structure is meaningful — e.g. the rendered YAML
// spec (yaml.Marshal indents with spaces and escapes control chars in values).
// It still drops tabs and non-newline unicode whitespace (incl. U+2028/U+2029)
// so only genuine "\n" line breaks remain. Defense in depth against injection.
func sanitizeBlock(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '\n' || r == ' ' {
			return r
		}
		if unicode.IsControl(r) || unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
}

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert external sandbox definitions to nanofuse specs",
	Long: `Convert external agent-sandbox definitions into nanofuse specs.

These commands operate on local files only and do not contact the API daemon.`,
}

var (
	convertAllowLossy    bool
	convertResolveEgress bool
	convertOutput        string
)

var convertGondolinCmd = &cobra.Command{
	Use:   "gondolin <file>",
	Short: "Convert a gondolin mirror sandbox to a nanofuse spec",
	Long: `Convert a nanofuse-authored "gondolin mirror" sandbox YAML into a nanofuse
spec, and report every gondolin feature that has no faithful nanofuse equivalent.

Reframe: the gondolin project (earendil-works/gondolin) has NO declarative
sandbox config file. Its sandbox is defined imperatively via gondolin bash/exec
CLI flags and a TypeScript VM.create() API, so there is nothing to parse from
gondolin itself. This command instead parses a nanofuse-authored YAML that
mirrors gondolin's flag surface (image, allow-host, host-secret,
mount-hostfs/memfs, ssh-allow-host, tcp-map, dns, env, cwd, vmm, rootfs-size,
plus nanofuse resource hints) and emits a nanofuse spec together with an
explicit divergence report.

By default the command FAILS CLOSED: if any gondolin feature has no faithful
nanofuse equivalent, it prints the divergence report and exits non-zero rather
than silently dropping it. Pass --allow-lossy to drop those features (loudly)
and proceed.

An L7 HTTP allowlist (allow-host) is always degraded safely to a locked-down
default-deny egress policy with a warning; pass --resolve-egress to opt in to
point-in-time hostname->CIDR resolution for literal hostnames.

Examples:
  # Inspect divergences (fails closed if features are unrepresentable)
  nanofuse convert gondolin sandbox.yaml

  # Drop unrepresentable features and write the spec
  nanofuse convert gondolin sandbox.yaml --allow-lossy -o nanofuse-spec.yaml`,
	Args: cobra.ExactArgs(1),
	// Fail-closed is an expected outcome, not a usage error; do not dump usage.
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("read gondolin mirror: %s", sanitizeInline(err.Error()))
		}

		sb, err := gondolin.Parse(data)
		if err != nil {
			// Parse errors can quote untrusted YAML keys/values (incl. control
			// chars); sanitize before returning so Cobra prints a safe message.
			return fmt.Errorf("%s", sanitizeInline(err.Error()))
		}

		opts := gondolin.Options{
			AllowLossy:    convertAllowLossy,
			ResolveEgress: convertResolveEgress,
		}
		if convertResolveEgress {
			resolver := &net.Resolver{}
			opts.Resolver = func(host string) ([]string, error) {
				// Bound each lookup and honour command cancellation (net.LookupHost
				// cannot be canceled or timed out).
				lookupCtx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
				defer cancel()
				return resolver.LookupHost(lookupCtx, host)
			}
		}

		req, divs, convErr := gondolin.Convert(sb, opts)

		// Always print the divergence report so the operator sees the full
		// picture, whether or not the conversion failed closed. Skip it for a pure
		// input-validation error (no divergences) so we don't print a misleading
		// "no divergences" line alongside a validation failure.
		if convErr == nil || len(divs) > 0 {
			printDivergences(divs)
		}

		if convErr != nil {
			// Conversion errors (e.g. an unknown dns mode) can echo untrusted
			// input values; sanitize before returning.
			return fmt.Errorf("%s", sanitizeInline(convErr.Error()))
		}

		specYAML, err := gondolin.RenderSpecYAML(req)
		if err != nil {
			return fmt.Errorf("render nanofuse spec: %w", err)
		}

		if convertOutput != "" {
			if err := os.WriteFile(convertOutput, specYAML, 0o600); err != nil {
				return fmt.Errorf("write nanofuse spec: %s", sanitizeInline(err.Error()))
			}
			fmt.Printf("Wrote nanofuse spec to %s\n", sanitizeInline(convertOutput))
			return nil
		}

		fmt.Printf("\n# nanofuse spec\n")
		// yaml.Marshal already escapes control characters in string values, but
		// sanitize the terminal-bound output as defense in depth against escape
		// injection from the untrusted input file (the -o file path writes raw).
		fmt.Print(sanitizeBlock(string(specYAML)))
		return nil
	},
}

func printDivergences(divs []gondolin.Divergence) {
	useColor := cliUseColor()
	red := color.New(color.FgRed)
	yellow := color.New(color.FgYellow)
	cyan := color.New(color.FgCyan)
	if !useColor {
		red.DisableColor()
		yellow.DisableColor()
		cyan.DisableColor()
	}

	if len(divs) == 0 {
		fmt.Println("Divergence report: no divergences (all fields map cleanly).")
		return
	}

	var blocking, warn, info int
	for _, d := range divs {
		switch d.Severity {
		case gondolin.SeverityBlocking:
			blocking++
		case gondolin.SeverityWarn:
			warn++
		case gondolin.SeverityInfo:
			info++
		}
	}

	fmt.Printf("Divergence report (%d blocking, %d warning, %d info):\n\n", blocking, warn, info)
	for _, d := range divs {
		var c *color.Color
		switch d.Severity {
		case gondolin.SeverityBlocking:
			c = red
		case gondolin.SeverityWarn:
			c = yellow
		default:
			c = cyan
		}
		c.Fprintf(os.Stdout, "  [%s] %s\n", d.Severity, sanitizeInline(d.Feature))
		fmt.Printf("      %s\n", sanitizeInline(d.Detail))
	}
}

func init() {
	convertGondolinCmd.Flags().BoolVar(&convertAllowLossy, "allow-lossy", false,
		"drop gondolin features with no faithful nanofuse equivalent and proceed (default: fail closed)")
	convertGondolinCmd.Flags().BoolVar(&convertResolveEgress, "resolve-egress", false,
		"opt in to point-in-time resolution of literal allow-host names to /32 egress rules")
	convertGondolinCmd.Flags().StringVarP(&convertOutput, "output", "o", "",
		"write the nanofuse spec to a file instead of stdout")

	convertCmd.AddCommand(convertGondolinCmd)
}
