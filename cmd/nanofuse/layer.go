package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/daax-dev/nanofuse/internal/layer"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var layerCmd = &cobra.Command{
	Use:   "layer",
	Short: "Manage layer definitions",
	Long: `Manage NanoFuse layer definitions.

Layers are the building blocks of NanoFuse microVM images. Use these commands
to create new layer scaffolds and validate existing layer configurations.`,
}

var (
	layerCreateType        string
	layerCreateOutput      string
	layerCreateDescription string
	layerCreateVersion     string
	layerCreateForce       bool
)

var layerCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new layer scaffold",
	Long: `Create a new layer directory structure with template files.

This command generates:
  - layer.yaml (template with required fields)
  - rootfs/ directory
  - hooks/ directory with template scripts
  - tests/ directory

Examples:
  # Create a feature layer in current directory
  nanofuse layer create my-feature --type feature

  # Create a runtime layer in a specific location
  nanofuse layer create python-app --type application --output ./layers

  # Create with custom description
  nanofuse layer create security-tools --type feature --description "Security scanning tools"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Convert type string to LayerType
		layerType := layer.LayerType(layerCreateType)
		if !layer.IsValidLayerType(layerCreateType) {
			validTypes := make([]string, len(layer.ValidLayerTypes))
			for i, t := range layer.ValidLayerTypes {
				validTypes[i] = string(t)
			}
			return fmt.Errorf("invalid layer type %q: must be one of %s", layerCreateType, strings.Join(validTypes, ", "))
		}

		opts := layer.ScaffoldOptions{
			Name:        name,
			Type:        layerType,
			OutputDir:   layerCreateOutput,
			Description: layerCreateDescription,
			Version:     layerCreateVersion,
			Force:       layerCreateForce,
		}

		result, err := layer.ScaffoldLayer(opts)
		if err != nil {
			return fmt.Errorf("create layer: %w", err)
		}

		// Print success message (use consistent fmt.Printf for all output)
		fmt.Printf("Created layer: %s\n", name)
		fmt.Printf("\nLocation: %s\n", result.LayerDir)
		fmt.Printf("\nCreated files:\n")
		for _, f := range result.CreatedFiles {
			relPath, err := filepath.Rel(result.LayerDir, f)
			if err != nil {
				// Fall back to absolute path if relative path fails
				relPath = f
			}
			fmt.Printf("  - %s\n", relPath)
		}

		fmt.Printf("\nNext steps:\n")
		fmt.Printf("  1. Edit %s/layer.yaml to customize dependencies and config\n", name)
		fmt.Printf("  2. Add your rootfs contents to %s/rootfs/\n", name)
		fmt.Printf("  3. Customize %s/hooks/post-install.sh\n", name)
		fmt.Printf("  4. Validate with: nanofuse layer validate %s\n", result.LayerDir)

		return nil
	},
}

var (
	layerValidateStrict bool
)

var layerValidateCmd = &cobra.Command{
	Use:   "validate <path>",
	Short: "Validate a layer configuration",
	Long: `Validate a layer directory structure and configuration.

This command checks:
  - layer.yaml exists and is valid YAML
  - Required fields are present (name, version, type)
  - rootfs/ directory exists
  - Hooks are executable
  - No invalid file permissions

Examples:
  # Validate a layer
  nanofuse layer validate ./layers/my-feature

  # Validate with strict checks
  nanofuse layer validate ./layers/my-feature --strict`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		layerPath := args[0]

		// Convert to absolute path
		absPath, err := filepath.Abs(layerPath)
		if err != nil {
			return fmt.Errorf("resolve path: %w", err)
		}

		opts := layer.ValidateOptions{
			Strict: layerValidateStrict,
		}

		result := layer.ValidateLayer(absPath, opts)

		// Print results
		printValidationResult(result)

		// Return error if validation failed
		if !result.Valid {
			return fmt.Errorf("validation failed with %d error(s)", len(result.Errors()))
		}

		return nil
	},
}

func printValidationResult(result *layer.ValidationResult) {
	// Determine if we should use colors
	useColor := !noColor && isTerminal()

	// Color helpers
	red := color.New(color.FgRed)
	yellow := color.New(color.FgYellow)
	green := color.New(color.FgGreen)
	cyan := color.New(color.FgCyan)

	if !useColor {
		red.DisableColor()
		yellow.DisableColor()
		green.DisableColor()
		cyan.DisableColor()
	}

	fmt.Printf("Validating layer: %s\n\n", result.LayerPath)

	// Print spec info if parsed successfully
	if result.Spec != nil {
		fmt.Printf("Layer: %s\n", result.Spec.Name)
		fmt.Printf("Type:  %s\n", result.Spec.Type)
		fmt.Printf("Version: %s\n", result.Spec.Version)
		fmt.Println()
	}

	// Group issues by severity using ValidationResult helper methods
	errors := result.Errors()
	warnings := result.Warnings()
	infos := result.Info()

	// Print errors
	if len(errors) > 0 {
		red.Fprintf(os.Stdout, "Errors (%d):\n", len(errors))
		for _, issue := range errors {
			printIssue(issue, red, cyan, useColor)
		}
		fmt.Println()
	}

	// Print warnings
	if len(warnings) > 0 {
		yellow.Fprintf(os.Stdout, "Warnings (%d):\n", len(warnings))
		for _, issue := range warnings {
			printIssue(issue, yellow, cyan, useColor)
		}
		fmt.Println()
	}

	// Print info (only in verbose/strict mode or if there are issues)
	if len(infos) > 0 && layerValidateStrict {
		cyan.Fprintf(os.Stdout, "Info (%d):\n", len(infos))
		for _, issue := range infos {
			printIssue(issue, cyan, cyan, useColor)
		}
		fmt.Println()
	}

	// Print summary
	if result.Valid {
		if len(warnings) == 0 && len(infos) == 0 {
			green.Fprintf(os.Stdout, "Validation passed!\n")
		} else {
			green.Fprintf(os.Stdout, "Validation passed with %d warning(s)\n", len(warnings))
		}
	} else {
		red.Fprintf(os.Stdout, "Validation failed with %d error(s)\n", len(errors))
	}
}

func printIssue(issue layer.ValidationIssue, severityColor, suggestionColor *color.Color, useColor bool) {
	// Print field if present
	if issue.Field != "" {
		fmt.Printf("  [%s] ", issue.Field)
	} else {
		fmt.Print("  ")
	}

	// Print message
	severityColor.Fprintf(os.Stdout, "%s\n", issue.Message)

	// Print suggestion if present
	if issue.Suggestion != "" {
		fmt.Print("    ")
		if useColor {
			suggestionColor.Fprintf(os.Stdout, "Suggestion: %s\n", issue.Suggestion)
		} else {
			fmt.Printf("Suggestion: %s\n", issue.Suggestion)
		}
	}
}

func init() {
	// layer create flags
	layerCreateCmd.Flags().StringVarP(&layerCreateType, "type", "t", "feature", "layer type (base, runtime, feature, application)")
	layerCreateCmd.Flags().StringVarP(&layerCreateOutput, "output", "o", "", "output directory (default: current directory)")
	layerCreateCmd.Flags().StringVarP(&layerCreateDescription, "description", "d", "", "layer description")
	layerCreateCmd.Flags().StringVar(&layerCreateVersion, "version", "0.1.0", "initial version")
	layerCreateCmd.Flags().BoolVarP(&layerCreateForce, "force", "f", false, "overwrite existing layer directory")

	// layer validate flags
	layerValidateCmd.Flags().BoolVar(&layerValidateStrict, "strict", false, "enable strict validation checks")

	// Add subcommands
	layerCmd.AddCommand(layerCreateCmd)
	layerCmd.AddCommand(layerValidateCmd)
}
