package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/daax-dev/nanofuse/internal/client"
	"github.com/daax-dev/nanofuse/internal/clierrors"
	"github.com/daax-dev/nanofuse/internal/inspect"
	"github.com/daax-dev/nanofuse/internal/output"
	"github.com/daax-dev/nanofuse/internal/validate"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var (
	// Global flags
	cfgFile    string
	apiSocket  string
	apiURL     string
	debug      bool
	jsonOutput bool
	noColor    bool
	timeout    time.Duration

	// Version info (set via ldflags)
	version   = "0.1.0"
	commit    = "dev"
	buildDate = "unknown"

	// API client
	apiClient *client.Client
	formatter *output.Formatter
)

const (
	// DefaultImageRegistry is the default GHCR registry for nanofuse images
	DefaultImageRegistry = "ghcr.io/daax-dev/nanofuse"
	// DefaultBaseImage is the default base image name
	DefaultBaseImage = "base"
	// DefaultImageTag is the default image tag
	DefaultImageTag = "latest"
	// DefaultAPISocketPath is the default nanofused Unix socket.
	DefaultAPISocketPath = "/var/run/nanofused.sock"
)

func main() {
	// Handle interrupts gracefully
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		// Check for client.ClientError (API errors with exit codes)
		if cerr, ok := err.(*client.ClientError); ok {
			fmt.Fprintf(os.Stderr, "Error: %s\n", cerr.Error())
			os.Exit(cerr.ExitCode())
		}
		// Check for clierrors.CLIError (user-friendly errors with exit codes)
		if cliErr, ok := err.(*clierrors.CLIError); ok {
			cliErr.Format(cliUseColor())
			os.Exit(cliErr.ExitCode)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:           "nanofuse",
	Short:         "NanoFuse - Firecracker microVM manager",
	SilenceUsage:  true,
	SilenceErrors: true,
	Long: `NanoFuse is a command-line tool for managing Firecracker-based microVMs.
It provides simple commands for VM lifecycle management, snapshots, and image handling.`,
	Version: version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip client setup for commands that don't need an API connection.
		// Layer commands operate on local files only.
		skipCommands := map[string]bool{
			"completion": true,
			"version":    true,
			"layer":      true,
		}
		if skipCommands[cmd.Name()] {
			return nil
		}
		// Also skip if parent command is in the skip list (e.g., "layer create")
		if cmd.Parent() != nil && skipCommands[cmd.Parent().Name()] {
			return nil
		}

		if err := applyClientEnvironment(); err != nil {
			return err
		}

		// Determine color usage
		useColor := cliUseColor()

		// Create formatter
		format := "table"
		if jsonOutput {
			format = "json"
		}
		formatter = output.NewFormatter(format, useColor)

		// Create API client
		if apiURL != "" {
			apiClient = client.NewTCPClient(apiURL, timeout, debug)
		} else {
			if apiSocket == "" {
				apiSocket = DefaultAPISocketPath
			}
			apiClient = client.NewClient(apiSocket, timeout, debug)
		}

		return nil
	},
}

func applyClientEnvironment() error {
	if cfgFile == "" {
		cfgFile = os.Getenv("NANOFUSE_CONFIG")
	}
	if apiURL == "" {
		apiURL = os.Getenv("NANOFUSE_API_URL")
	}
	if apiSocket == "" {
		apiSocket = os.Getenv("NANOFUSE_API_SOCKET")
	}
	if timeout == 30*time.Second {
		if value := os.Getenv("NANOFUSE_TIMEOUT"); value != "" {
			parsed, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("parse NANOFUSE_TIMEOUT: %w", err)
			}
			timeout = parsed
		}
	}
	if !debug && envTruthy(os.Getenv("NANOFUSE_DEBUG")) {
		debug = true
	}
	if !jsonOutput && strings.EqualFold(os.Getenv("NANOFUSE_OUTPUT"), "json") {
		jsonOutput = true
	}
	if !noColor && envTruthy(os.Getenv("NANOFUSE_NO_COLOR")) {
		noColor = true
	}

	return nil
}

func envTruthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func cliUseColor() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return !noColor && isTerminal()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.config/nanofuse/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&apiSocket, "api-socket", "", "API Unix socket path (default: /var/run/nanofused.sock)")
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "", "API URL for remote access (e.g., http://localhost:8080)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug output")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output in JSON format")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().DurationVar(&timeout, "timeout", 30*time.Second, "API request timeout")

	// Add subcommands
	rootCmd.AddCommand(healthCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(imageCmd)
	rootCmd.AddCommand(vmCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(layerCmd)

	// Customize version output
	rootCmd.SetVersionTemplate(`{{.Version}}
`)
}

// Health command
var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check API daemon health",
	RunE: func(cmd *cobra.Command, args []string) error {
		health, err := apiClient.Health(cmd.Context())
		if err != nil {
			return handleAPIError(err, "check API health")
		}

		return formatter.PrintHealth(health)
	},
}

// Version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		if jsonOutput {
			fmt.Printf(`{
  "cli_version": "%s",
  "git_commit": "%s",
  "built_at": "%s",
  "go_version": "go1.22",
  "platform": "linux/amd64"
}
`, version, commit, buildDate)
		} else {
			fmt.Printf("CLI Version:  %s\n", version)
			// Try to get API version
			if apiClient != nil {
				if health, err := apiClient.Health(context.Background()); err == nil {
					fmt.Printf("API Version:  %s\n", health.Version)
				}
			}
			fmt.Printf("Git Commit:   %s\n", commit)
			fmt.Printf("Built:        %s\n", buildDate)
			fmt.Printf("Go Version:   go1.22\n")
			fmt.Printf("Platform:     linux/amd64\n")
		}
	},
}

// Image commands
var imageCmd = &cobra.Command{
	Use:   "image",
	Short: "Manage VM images",
}

var (
	pullImageTag   string
	pullUseDefault bool
)

var imagePullCmd = &cobra.Command{
	Use:   "pull [image-ref]",
	Short: "Pull an image from registry",
	Long: `Pull an image from a container registry.

Examples:
  # Pull a specific image
  nanofuse image pull ghcr.io/user/image:v1.0

  # Pull the default nanofuse base image (latest)
  nanofuse image pull --default

  # Pull a specific version of the default base image
  nanofuse image pull --default --tag v1.0.0`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var imageRef string

		// Handle default flag
		if pullUseDefault {
			imageRef = fmt.Sprintf("%s/%s:%s", DefaultImageRegistry, DefaultBaseImage, pullImageTag)
			fmt.Printf("Using default image: %s\n", imageRef)
		} else if len(args) == 0 {
			return fmt.Errorf("image reference required (or use --default flag)")
		} else {
			imageRef = args[0]
		}

		// Start pull
		job, err := apiClient.PullImage(cmd.Context(), imageRef)
		if err != nil {
			return handleAPIErrorWithResource(err, "pull image", imageRef)
		}

		fmt.Printf("Pulling %s...\n", imageRef)

		// Poll for progress
		var bar *progressbar.ProgressBar
		for {
			job, err = apiClient.GetPullJob(cmd.Context(), job.ID)
			if err != nil {
				return handleAPIErrorWithResource(err, "check pull status", imageRef)
			}

			if job.State == "completed" {
				if bar != nil {
					_ = bar.Finish()
				}
				formatter.PrintSuccess("Pull complete!")
				fmt.Printf("\nDigest: %s\n", *job.ResultDigest)
				return nil
			}

			if job.State == "failed" {
				if bar != nil {
					_ = bar.Finish()
				}
				msg := "Pull failed"
				if job.Error != nil {
					msg = *job.Error
				}
				return fmt.Errorf("%s", msg)
			}

			if job.Progress != nil && !jsonOutput {
				if bar == nil {
					bar = output.NewProgressBar(job.Progress.TotalBytes, "Downloading")
				}
				_ = bar.Set64(job.Progress.CurrentBytes)
			}

			time.Sleep(500 * time.Millisecond)
		}
	},
}

var imageListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List cached images",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := apiClient.ListImages(cmd.Context())
		if err != nil {
			return handleAPIError(err, "list images")
		}

		return formatter.PrintImageList(resp.Images)
	},
}

var imageInspectCmd = &cobra.Command{
	Use:   "inspect <image-ref>",
	Short: "Show detailed image information",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		imageRef := args[0]

		img, err := apiClient.GetImage(cmd.Context(), imageRef)
		if err != nil {
			return handleAPIErrorWithResource(err, "get image", imageRef)
		}

		return formatter.PrintImage(img)
	},
}

var imageRemoveCmd = &cobra.Command{
	Use:     "remove <image-ref>",
	Aliases: []string{"rm"},
	Short:   "Remove a cached image",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		imageRef := args[0]

		if err := apiClient.DeleteImage(cmd.Context(), imageRef); err != nil {
			return handleAPIErrorWithResource(err, "remove image", imageRef)
		}

		formatter.PrintSuccess(fmt.Sprintf("Removed image: %s", imageRef))
		return nil
	},
}

var imageLabelsCmd = &cobra.Command{
	Use:   "labels <image-ref>",
	Short: "Show image labels (OCI and NanoFuse metadata)",
	Long: `Display all labels attached to an image, including:
  - OCI standard labels (org.opencontainers.image.*)
  - NanoFuse-specific labels (com.nanofuse.*)
  - Other custom labels

Labels contain metadata like version, build date, source repository, and image configuration.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		imageRef := args[0]

		img, err := apiClient.GetImage(cmd.Context(), imageRef)
		if err != nil {
			return handleAPIErrorWithResource(err, "get image labels", imageRef)
		}

		if jsonOutput {
			return formatter.PrintJSON(img.Labels)
		}

		if len(img.Labels) == 0 {
			fmt.Println("No labels found for this image")
			return nil
		}

		fmt.Printf("Image: %s\n", imageRef)
		fmt.Printf("Digest: %s\n\n", img.Digest)

		// Print OCI labels
		fmt.Println("OCI Standard Labels:")
		hasOCI := false
		for k, v := range img.Labels {
			if strings.HasPrefix(k, "org.opencontainers.image.") {
				fmt.Printf("  %-35s %s\n", k, v)
				hasOCI = true
			}
		}
		if !hasOCI {
			fmt.Println("  (none)")
		}

		// Print NanoFuse labels
		fmt.Println("\nNanoFuse Labels:")
		hasNanoFuse := false
		for k, v := range img.Labels {
			if strings.HasPrefix(k, "com.nanofuse.") {
				fmt.Printf("  %-35s %s\n", k, v)
				hasNanoFuse = true
			}
		}
		if !hasNanoFuse {
			fmt.Println("  (none)")
		}

		// Print other labels
		var otherLabels []string
		for k := range img.Labels {
			if !strings.HasPrefix(k, "org.opencontainers.image.") && !strings.HasPrefix(k, "com.nanofuse.") {
				otherLabels = append(otherLabels, k)
			}
		}
		if len(otherLabels) > 0 {
			fmt.Println("\nOther Labels:")
			for _, k := range otherLabels {
				fmt.Printf("  %-35s %s\n", k, img.Labels[k])
			}
		}

		return nil
	},
}

// imageInspectFileCmd inspects a local ext4 image file for layer metadata
var (
	inspectShowLayers bool
)

var imageInspectFileCmd = &cobra.Command{
	Use:   "inspect-file <path>",
	Short: "Inspect a local ext4 image file for layer metadata",
	Long: `Inspect a local ext4 rootfs image file and display layer metadata.

This command reads the build manifest from /etc/nanofuse/build-manifest.json
inside the image and displays information about:
  - Image name and build timestamp
  - Kernel version and cmdline
  - Layer information (name, version, digest, type)
  - Total image size

For images without NanoFuse layer metadata (legacy images), basic file
information will be displayed.

Examples:
  # Inspect an image file
  nanofuse image inspect-file ./build/rootfs.ext4

  # Show layer details
  nanofuse image inspect-file ./build/rootfs.ext4 --layers

  # Output as JSON
  nanofuse image inspect-file ./build/rootfs.ext4 --json

Note: This command requires either root privileges (for mounting) or
the debugfs tool (e2fsprogs package) for non-root access.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		imagePath := args[0]

		// Create a temporary work directory
		workDir, err := os.MkdirTemp("", "nanofuse-inspect-")
		if err != nil {
			return fmt.Errorf("failed to create temp directory: %w", err)
		}
		defer os.RemoveAll(workDir)

		// Create inspector
		inspector := inspect.NewInspector(workDir, debug)

		// Inspect the image
		result, err := inspector.InspectImage(cmd.Context(), imagePath)
		if err != nil {
			return fmt.Errorf("failed to inspect image: %w", err)
		}

		// Output result
		if jsonOutput {
			return inspect.FormatJSON(result, os.Stdout)
		}

		return inspect.FormatText(result, inspectShowLayers, os.Stdout)
	},
}

var (
	validateReportPath string
	validateStrict     bool
)

var imageValidateCmd = &cobra.Command{
	Use:   "validate <image-path>",
	Short: "Validate an ext4 filesystem image",
	Long: `Validate an ext4 filesystem image for NanoFuse compatibility.

This command mounts the image (requires sudo) and performs validation checks:
  - Filesystem integrity (e2fsck)
  - Required directories (/etc, /usr, /var, etc.)
  - Layer metadata (/etc/nanofuse/build-manifest.json)
  - Systemd configuration
  - SSH configuration

Examples:
  # Validate an image (text output)
  nanofuse image validate ./build/rootfs.ext4

  # Validate with JSON report output
  nanofuse image validate ./build/rootfs.ext4 --report validation.json

  # Strict mode (warnings treated as failures)
  nanofuse image validate ./build/rootfs.ext4 --strict`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		imagePath := args[0]

		// Resolve to absolute path
		absPath, err := filepath.Abs(imagePath)
		if err != nil {
			return fmt.Errorf("cannot resolve path: %w", err)
		}

		// Check if image exists
		if _, err := os.Stat(absPath); err != nil {
			return fmt.Errorf("image not found: %s", absPath)
		}

		// Create temporary mount point
		mountPoint, err := os.MkdirTemp("", "nanofuse-validate-*")
		if err != nil {
			return fmt.Errorf("cannot create mount point: %w", err)
		}
		defer os.RemoveAll(mountPoint)

		// Mount the image (requires sudo)
		fmt.Printf("Mounting %s...\n", absPath)
		mountCmd := exec.Command("sudo", "mount", "-o", "loop,ro", absPath, mountPoint)
		if output, err := mountCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to mount image: %s\n%s", err, string(output))
		}
		defer func() {
			// Unmount when done
			umountCmd := exec.Command("sudo", "umount", mountPoint)
			_ = umountCmd.Run()
		}()

		fmt.Println("Running validation checks...")
		fmt.Println()

		// Create validator and run checks
		validator := validate.NewValidator(absPath, validateStrict)
		result, err := validator.Validate(mountPoint)
		if err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}

		// Determine color usage
		useColor := !noColor && isTerminal()

		// Output report
		if validateReportPath != "" {
			// Write JSON report to file
			reportData, err := validate.GenerateJSONReport(result)
			if err != nil {
				return fmt.Errorf("failed to generate JSON report: %w", err)
			}
			if err := os.WriteFile(validateReportPath, reportData, 0600); err != nil {
				return fmt.Errorf("failed to write report: %w", err)
			}
			fmt.Printf("Report written to: %s\n\n", validateReportPath)
		}

		if jsonOutput {
			// Print JSON to stdout
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			if err := enc.Encode(result); err != nil {
				return fmt.Errorf("failed to encode JSON: %w", err)
			}
		} else {
			// Print text report
			report := validate.GenerateTextReport(result, useColor)
			fmt.Print(report)
		}

		// Return error if validation failed (sets exit code to 1)
		if !result.Passed {
			return &clierrors.CLIError{
				Message:  "Image validation failed",
				ExitCode: 1,
			}
		}

		return nil
	},
}

func init() {
	// Image pull flags
	imagePullCmd.Flags().BoolVar(&pullUseDefault, "default", false, "pull the default nanofuse base image")
	imagePullCmd.Flags().StringVar(&pullImageTag, "tag", DefaultImageTag, "image tag to pull (used with --default)")

	// Image inspect-file flags
	imageInspectFileCmd.Flags().BoolVar(&inspectShowLayers, "layers", false, "show detailed layer information")

	// Image validate flags
	imageValidateCmd.Flags().StringVar(&validateReportPath, "report", "", "write JSON report to file")
	imageValidateCmd.Flags().BoolVar(&validateStrict, "strict", false, "treat warnings as failures")

	imageCmd.AddCommand(imagePullCmd)
	imageCmd.AddCommand(imageListCmd)
	imageCmd.AddCommand(imageInspectCmd)
	imageCmd.AddCommand(imageInspectFileCmd)
	imageCmd.AddCommand(imageLabelsCmd)
	imageCmd.AddCommand(imageRemoveCmd)
	imageCmd.AddCommand(imageValidateCmd)
}

// VM commands
var vmCmd = &cobra.Command{
	Use:   "vm",
	Short: "Manage virtual machines",
}

var (
	vmName         string
	vmVCPUs        int
	vmMemory       int
	vmNetwork      string
	vmBridge       string
	vmKernelArgs   string
	vmPortForwards []string
	vmSSHKey       string
)

var vmCreateCmd = &cobra.Command{
	Use:   "create <image-ref> [name]",
	Short: "Create a new VM",
	Long: `Create a new VM from an image.

Image reference can be:
  - Full reference: ghcr.io/user/image:tag
  - Default shorthand: default, default:v1.0, base, base:latest

Examples:
  nanofuse vm create default my-vm
  nanofuse vm create default:v1.0.0 my-vm
  nanofuse vm create ghcr.io/user/custom:latest my-vm`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		imageRef := resolveImageRef(args[0])

		name := vmName
		if len(args) > 1 {
			name = args[1]
		}

		// Parse port forwards
		var portForwards []client.PortForward
		if len(vmPortForwards) > 0 {
			var err error
			portForwards, err = parsePortForwards(vmPortForwards)
			if err != nil {
				return fmt.Errorf("invalid port forward: %w", err)
			}
		}

		// Read SSH key if provided
		var sshKeyEncoded string
		if vmSSHKey != "" {
			var err error
			sshKeyEncoded, err = readSSHPublicKey(vmSSHKey)
			if err != nil {
				return fmt.Errorf("SSH key error: %w", err)
			}
		}

		req := &client.CreateVMRequest{
			Name:  name,
			Image: imageRef,
			Config: client.VMConfig{
				VCPUs:        vmVCPUs,
				MemoryMiB:    vmMemory,
				KernelArgs:   vmKernelArgs,
				SSHPublicKey: sshKeyEncoded,
				Network: client.NetworkConfig{
					Mode:         vmNetwork,
					BridgeName:   nil,
					PortForwards: portForwards,
				},
			},
		}

		if vmBridge != "" {
			req.Config.Network.BridgeName = &vmBridge
		}

		vm, err := apiClient.CreateVM(cmd.Context(), req)
		if err != nil {
			return handleAPIErrorWithResource(err, "create VM", imageRef)
		}

		formatter.PrintSuccess(fmt.Sprintf("Created VM: %s", vm.Name))
		fmt.Printf("ID:    %s\n", vm.ID)
		fmt.Printf("State: %s\n", vm.State)
		fmt.Printf("Image: %s\n", vm.Image)
		if sshKeyEncoded != "" {
			fmt.Printf("SSH:   key will be injected on boot\n")
		}
		if len(portForwards) > 0 {
			fmt.Printf("\nPort Forwards:\n")
			for _, pf := range portForwards {
				fmt.Printf("  host:%d -> vm:%d (%s)\n", pf.HostPort, pf.VMPort, pf.Protocol)
			}
		}
		fmt.Printf("\nUse 'nanofuse vm start %s' to start the VM\n", vm.Name)

		return nil
	},
}

var vmRunCmd = &cobra.Command{
	Use:   "run <image-ref> [name]",
	Short: "Create and start a VM",
	Long: `Create and start a VM from an image.

Image reference can be:
  - Full reference: ghcr.io/user/image:tag
  - Default shorthand: default, default:v1.0, base, base:latest

Examples:
  nanofuse vm run default my-vm
  nanofuse vm run default:v1.0.0 my-vm
  nanofuse vm run ghcr.io/user/custom:latest my-vm`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		imageRef := resolveImageRef(args[0])

		name := vmName
		if len(args) > 1 {
			name = args[1]
		}

		// Parse port forwards
		var portForwards []client.PortForward
		if len(vmPortForwards) > 0 {
			var err error
			portForwards, err = parsePortForwards(vmPortForwards)
			if err != nil {
				return fmt.Errorf("invalid port forward: %w", err)
			}
		}

		// Read SSH key if provided
		var sshKeyEncoded string
		if vmSSHKey != "" {
			var err error
			sshKeyEncoded, err = readSSHPublicKey(vmSSHKey)
			if err != nil {
				return fmt.Errorf("SSH key error: %w", err)
			}
		}

		req := &client.CreateVMRequest{
			Name:  name,
			Image: imageRef,
			Config: client.VMConfig{
				VCPUs:        vmVCPUs,
				MemoryMiB:    vmMemory,
				KernelArgs:   vmKernelArgs,
				SSHPublicKey: sshKeyEncoded,
				Network: client.NetworkConfig{
					Mode:         vmNetwork,
					PortForwards: portForwards,
				},
			},
		}

		if vmBridge != "" {
			req.Config.Network.BridgeName = &vmBridge
		}

		vm, err := apiClient.CreateVM(cmd.Context(), req)
		if err != nil {
			return handleAPIErrorWithResource(err, "create VM", imageRef)
		}

		formatter.PrintSuccess(fmt.Sprintf("Created VM: %s", vm.Name))
		fmt.Println("Starting VM...")

		vm, err = apiClient.StartVM(cmd.Context(), vm.ID)
		if err != nil {
			return handleAPIErrorWithResource(err, "start VM", vm.ID)
		}

		formatter.PrintSuccess("VM started successfully!")
		fmt.Printf("\nID:     %s\n", vm.ID)
		fmt.Printf("State:  %s\n", vm.State)
		if vm.Runtime != nil && vm.Runtime.NetworkInfo.GuestIP != "" {
			fmt.Printf("IP:     %s\n", vm.Runtime.NetworkInfo.GuestIP)
			if sshKeyEncoded != "" {
				fmt.Printf("\nSSH: ssh root@%s\n", vm.Runtime.NetworkInfo.GuestIP)
			}
		}
		if len(portForwards) > 0 {
			fmt.Printf("\nPort Forwards:\n")
			for _, pf := range portForwards {
				fmt.Printf("  localhost:%d -> vm:%d (%s)\n", pf.HostPort, pf.VMPort, pf.Protocol)
			}
		}

		return nil
	},
}

var vmListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all VMs",
	RunE: func(cmd *cobra.Command, args []string) error {
		stateFilter, _ := cmd.Flags().GetString("filter")

		resp, err := apiClient.ListVMs(cmd.Context(), stateFilter)
		if err != nil {
			return handleAPIError(err, "list VMs")
		}

		return formatter.PrintVMList(resp.VMs)
	},
}

var vmStatusCmd = &cobra.Command{
	Use:     "status <vm-id>",
	Aliases: []string{"inspect", "info"},
	Short:   "Show detailed VM status",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmID := args[0]

		vm, err := apiClient.GetVM(cmd.Context(), vmID)
		if err != nil {
			return handleAPIErrorWithResource(err, "get VM status", vmID)
		}

		return formatter.PrintVM(vm)
	},
}

var vmStartCmd = &cobra.Command{
	Use:   "start <vm-id>",
	Short: "Start a VM",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmID := args[0]

		fmt.Printf("Starting VM: %s\n", vmID)
		vm, err := apiClient.StartVM(cmd.Context(), vmID)
		if err != nil {
			return handleAPIErrorWithResource(err, "start VM", vmID)
		}

		formatter.PrintSuccess("VM started successfully!")
		fmt.Printf("State: %s\n", vm.State)
		if vm.Runtime != nil && vm.Runtime.NetworkInfo.GuestIP != "" {
			fmt.Printf("IP:    %s\n", vm.Runtime.NetworkInfo.GuestIP)
		}

		return nil
	},
}

var vmStopCmd = &cobra.Command{
	Use:   "stop <vm-id>",
	Short: "Stop a VM gracefully",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmID := args[0]
		stopTimeout, _ := cmd.Flags().GetInt("timeout")

		fmt.Printf("Stopping VM: %s\n", vmID)
		fmt.Println("Sending shutdown signal...")

		vm, err := apiClient.StopVM(cmd.Context(), vmID, stopTimeout)
		if err != nil {
			return handleAPIErrorWithResource(err, "stop VM", vmID)
		}

		formatter.PrintSuccess("VM stopped successfully!")
		fmt.Printf("State: %s\n", vm.State)

		return nil
	},
}

var vmKillCmd = &cobra.Command{
	Use:   "kill <vm-id>",
	Short: "Force kill a VM",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmID := args[0]

		fmt.Printf("Force killing VM: %s\n", vmID)

		vm, err := apiClient.KillVM(cmd.Context(), vmID)
		if err != nil {
			return handleAPIErrorWithResource(err, "kill VM", vmID)
		}

		formatter.PrintSuccess("VM killed!")
		fmt.Printf("State: %s\n", vm.State)

		return nil
	},
}

var vmRestartCmd = &cobra.Command{
	Use:   "restart <vm-id>",
	Short: "Restart a VM",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmID := args[0]

		fmt.Printf("Restarting VM: %s\n", vmID)
		fmt.Println("Stopping VM...")

		_, err := apiClient.StopVM(cmd.Context(), vmID, 30)
		if err != nil {
			return handleAPIErrorWithResource(err, "stop VM (during restart)", vmID)
		}

		formatter.PrintSuccess("VM stopped!")
		fmt.Println("Starting VM...")

		vm, err := apiClient.StartVM(cmd.Context(), vmID)
		if err != nil {
			return handleAPIErrorWithResource(err, "start VM (during restart)", vmID)
		}

		formatter.PrintSuccess("VM started successfully!")
		fmt.Printf("State: %s\n", vm.State)

		return nil
	},
}

var vmPauseCmd = &cobra.Command{
	Use:   "pause <vm-id>",
	Short: "Pause a running VM",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmID := args[0]

		fmt.Printf("Pausing VM: %s\n", vmID)

		vm, err := apiClient.PauseVM(cmd.Context(), vmID)
		if err != nil {
			return handleAPIErrorWithResource(err, "pause VM", vmID)
		}

		formatter.PrintSuccess("VM paused successfully!")
		fmt.Printf("State: %s\n", vm.State)

		return nil
	},
}

var vmResumeCmd = &cobra.Command{
	Use:   "resume <vm-id>",
	Short: "Resume a paused VM or from snapshot",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmID := args[0]
		snapshotID, _ := cmd.Flags().GetString("from-snapshot")

		if snapshotID != "" {
			fmt.Printf("Resuming VM: %s from snapshot: %s\n", vmID, snapshotID)
		} else {
			fmt.Printf("Resuming VM: %s\n", vmID)
		}

		vm, err := apiClient.ResumeVM(cmd.Context(), vmID, snapshotID)
		if err != nil {
			return handleAPIErrorWithResource(err, "resume VM", vmID)
		}

		formatter.PrintSuccess("VM resumed successfully!")
		fmt.Printf("State: %s\n", vm.State)

		return nil
	},
}

var vmDeleteCmd = &cobra.Command{
	Use:     "delete <vm-id>",
	Aliases: []string{"rm"},
	Short:   "Delete a VM",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmID := args[0]
		force, _ := cmd.Flags().GetBool("force")

		if !force {
			fmt.Printf("Delete VM '%s'?\n", vmID)
			fmt.Println("This will stop the VM and remove all state. Snapshots will be preserved.")
			fmt.Print("Continue? [y/N]: ")

			var response string
			if _, err := fmt.Scanln(&response); err != nil {
				// Treat error as cancellation
				fmt.Println("Cancelled")
				return nil
			}
			if response != "y" && response != "Y" {
				fmt.Println("Cancelled")
				return nil
			}
		}

		fmt.Println("Stopping VM...")
		fmt.Println("Removing VM state...")

		if err := apiClient.DeleteVM(cmd.Context(), vmID); err != nil {
			return handleAPIErrorWithResource(err, "delete VM", vmID)
		}

		formatter.PrintSuccess(fmt.Sprintf("Deleted VM: %s", vmID))

		return nil
	},
}

// streamVMLogs streams VM logs in real-time.
// NOTE: This implementation fetches the entire log on each 500ms poll, which is
// inefficient for large log files. A future optimization could implement server-side
// offset tracking, use a more efficient streaming protocol, or make the polling
// interval configurable to trade off responsiveness against load.
func streamVMLogs(ctx context.Context, vmID string, initialTail int) error {
	// Get the current logs once and use them both for initial output and sizing
	logs, err := apiClient.GetVMLogs(ctx, vmID, 0)
	if err != nil {
		return handleAPIErrorWithResource(err, "get VM logs", vmID)
	}

	// Print initial logs based on the requested tail
	if initialTail > 0 && len(logs) > 0 {
		// Trim trailing newlines and check if anything remains
		trimmed := strings.TrimRight(logs, "\n")
		if trimmed != "" {
			lines := strings.Split(trimmed, "\n")
			if len(lines) > initialTail {
				// Join the last N lines and add trailing newline for consistency
				fmt.Print(strings.Join(lines[len(lines)-initialTail:], "\n") + "\n")
			} else {
				fmt.Print(logs)
			}
		}
	} else if len(logs) > 0 {
		fmt.Print(logs)
	}

	// Start following from the current end of the log
	lastSize := int64(len(logs))

	// Poll for new logs every 500ms
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			// Get all logs to check if new content appeared
			currentLogs, err := apiClient.GetVMLogs(ctx, vmID, 0)
			if err != nil {
				// VM was deleted or no longer exists - that's okay for follow mode
				if cerr, ok := err.(*client.ClientError); ok && cerr.Code == "VM_NOT_FOUND" {
					return nil
				}
				// For other errors, keep trying
				continue
			}

			currentSize := int64(len(currentLogs))
			// Handle log truncation/rotation - if size decreased, treat current content as fresh
			if currentSize < lastSize {
				// On truncation/rotation, print the current logs as new (if any)
				if currentSize > 0 {
					fmt.Print(currentLogs)
				}
				lastSize = currentSize
				continue
			}
			if currentSize > lastSize {
				// New content appeared - print only the new part
				newContent := currentLogs[lastSize:]
				fmt.Print(newContent)
				lastSize = currentSize
			}
		}
	}
}

var vmLogsCmd = &cobra.Command{
	Use:   "logs <vm-id>",
	Short: "Show VM console logs",
	Long: `Show VM console logs from the serial console.

Examples:
  # Show all logs
  nanofuse vm logs my-vm

  # Show last 20 lines
  nanofuse vm logs my-vm --tail 20
  nanofuse vm logs my-vm -n 20

  # Follow logs in real-time
  nanofuse vm logs my-vm --follow
  nanofuse vm logs my-vm -f`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmID := args[0]
		follow, _ := cmd.Flags().GetBool("follow")
		tail, _ := cmd.Flags().GetInt("tail")
		lines, _ := cmd.Flags().GetInt("lines")

		// If --lines was explicitly set and --tail was not, use --lines value.
		// Otherwise, use --tail value (which may be default 0 or user-provided).
		if cmd.Flags().Changed("lines") && !cmd.Flags().Changed("tail") {
			tail = lines
		}

		if follow {
			// Stream logs in real-time
			return streamVMLogs(cmd.Context(), vmID, tail)
		}

		// Get logs once
		logs, err := apiClient.GetVMLogs(cmd.Context(), vmID, tail)
		if err != nil {
			return handleAPIErrorWithResource(err, "get VM logs", vmID)
		}

		fmt.Print(logs)

		return nil
	},
}

func init() {
	// VM create flags
	vmCreateCmd.Flags().StringVar(&vmName, "name", "", "VM name")
	vmCreateCmd.Flags().IntVar(&vmVCPUs, "vcpus", 2, "number of vCPUs")
	vmCreateCmd.Flags().IntVar(&vmMemory, "memory", 512, "memory in MiB")
	vmCreateCmd.Flags().StringVar(&vmNetwork, "network", "nat", "network mode (nat|bridged|none)")
	vmCreateCmd.Flags().StringVar(&vmBridge, "bridge", "", "bridge name (required if network=bridged)")
	vmCreateCmd.Flags().StringVar(&vmKernelArgs, "kernel-args", "", "kernel arguments")
	vmCreateCmd.Flags().StringArrayVarP(&vmPortForwards, "port-forward", "p", nil, "port forward (format: hostPort:vmPort[/protocol], e.g., 8080:80 or 53:53/udp)")
	vmCreateCmd.Flags().StringVar(&vmSSHKey, "ssh-key", "", "path to SSH public key for VM access")

	// VM run flags (same as create)
	vmRunCmd.Flags().StringVar(&vmName, "name", "", "VM name")
	vmRunCmd.Flags().IntVar(&vmVCPUs, "vcpus", 2, "number of vCPUs")
	vmRunCmd.Flags().IntVar(&vmMemory, "memory", 512, "memory in MiB")
	vmRunCmd.Flags().StringVar(&vmNetwork, "network", "nat", "network mode (nat|bridged|none)")
	vmRunCmd.Flags().StringVar(&vmBridge, "bridge", "", "bridge name (required if network=bridged)")
	vmRunCmd.Flags().StringVar(&vmKernelArgs, "kernel-args", "", "kernel arguments")
	vmRunCmd.Flags().StringArrayVarP(&vmPortForwards, "port-forward", "p", nil, "port forward (format: hostPort:vmPort[/protocol], e.g., 8080:80 or 53:53/udp)")
	vmRunCmd.Flags().StringVar(&vmSSHKey, "ssh-key", "", "path to SSH public key for VM access")

	// VM list flags
	vmListCmd.Flags().String("filter", "", "filter by state")

	// VM stop flags
	vmStopCmd.Flags().Int("timeout", 30, "graceful shutdown timeout in seconds")

	// VM resume flags
	vmResumeCmd.Flags().String("from-snapshot", "", "snapshot ID to resume from")

	// VM delete flags
	vmDeleteCmd.Flags().BoolP("force", "f", false, "force delete without confirmation")

	// VM logs flags
	vmLogsCmd.Flags().Int("tail", 0, "show last N lines")
	vmLogsCmd.Flags().IntP("lines", "n", 0, "show last N lines (alias for --tail)")
	vmLogsCmd.Flags().BoolP("follow", "f", false, "stream logs in real-time")

	// Add VM subcommands
	vmCmd.AddCommand(vmCreateCmd)
	vmCmd.AddCommand(vmRunCmd)
	vmCmd.AddCommand(vmListCmd)
	vmCmd.AddCommand(vmStatusCmd)
	vmCmd.AddCommand(vmStartCmd)
	vmCmd.AddCommand(vmStopCmd)
	vmCmd.AddCommand(vmKillCmd)
	vmCmd.AddCommand(vmRestartCmd)
	vmCmd.AddCommand(vmPauseCmd)
	vmCmd.AddCommand(vmResumeCmd)
	vmCmd.AddCommand(vmDeleteCmd)
	vmCmd.AddCommand(vmLogsCmd)

	// Snapshot commands as subcommand of vm
	vmCmd.AddCommand(snapshotCmd)
}

// Snapshot commands
var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Manage VM snapshots",
}

var snapshotCreateCmd = &cobra.Command{
	Use:   "create <vm-id> [name]",
	Short: "Create a VM snapshot",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmID := args[0]

		var name string
		if len(args) > 1 {
			name = args[1]
		}

		req := &client.CreateSnapshotRequest{
			Name: name,
		}

		fmt.Printf("Creating snapshot for VM: %s\n", vmID)

		snapshot, err := apiClient.CreateSnapshot(cmd.Context(), vmID, req)
		if err != nil {
			return handleAPIErrorWithResource(err, "create snapshot", vmID)
		}

		formatter.PrintSuccess("Snapshot created successfully!")
		fmt.Printf("\nID:      %s\n", snapshot.ID)
		if snapshot.Name != "" {
			fmt.Printf("Name:    %s\n", snapshot.Name)
		}
		fmt.Printf("Size:    %s\n", formatBytes(snapshot.SizeBytes))
		fmt.Printf("Created: %s\n", snapshot.CreatedAt.Format("2006-01-02 15:04:05 MST"))
		fmt.Printf("\nUse 'nanofuse vm resume %s --from-snapshot %s' to resume from this snapshot\n", vmID, snapshot.ID)

		return nil
	},
}

var snapshotListCmd = &cobra.Command{
	Use:     "list <vm-id>",
	Aliases: []string{"ls"},
	Short:   "List VM snapshots",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmID := args[0]

		resp, err := apiClient.ListSnapshots(cmd.Context(), vmID)
		if err != nil {
			return handleAPIErrorWithResource(err, "list snapshots", vmID)
		}

		return formatter.PrintSnapshotList(resp.Snapshots)
	},
}

var snapshotInspectCmd = &cobra.Command{
	Use:   "inspect <snapshot-id>",
	Short: "Show detailed snapshot information",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		snapshotID := args[0]

		snapshot, err := apiClient.GetSnapshot(cmd.Context(), snapshotID)
		if err != nil {
			return handleAPIErrorWithResource(err, "get snapshot", snapshotID)
		}

		return formatter.PrintSnapshot(snapshot)
	},
}

var snapshotDeleteCmd = &cobra.Command{
	Use:     "delete <snapshot-id>",
	Aliases: []string{"rm"},
	Short:   "Delete a snapshot",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		snapshotID := args[0]
		force, _ := cmd.Flags().GetBool("force")

		if !force {
			fmt.Printf("Delete snapshot '%s'?\n", snapshotID)
			fmt.Print("Continue? [y/N]: ")

			var response string
			if _, err := fmt.Scanln(&response); err != nil {
				// Treat error as cancellation
				fmt.Println("Cancelled")
				return nil
			}
			if response != "y" && response != "Y" {
				fmt.Println("Cancelled")
				return nil
			}
		}

		if err := apiClient.DeleteSnapshot(cmd.Context(), snapshotID); err != nil {
			return handleAPIErrorWithResource(err, "delete snapshot", snapshotID)
		}

		formatter.PrintSuccess(fmt.Sprintf("Deleted snapshot: %s", snapshotID))

		return nil
	},
}

func init() {
	snapshotDeleteCmd.Flags().BoolP("force", "f", false, "force delete without confirmation")

	snapshotCmd.AddCommand(snapshotCreateCmd)
	snapshotCmd.AddCommand(snapshotListCmd)
	snapshotCmd.AddCommand(snapshotInspectCmd)
	snapshotCmd.AddCommand(snapshotDeleteCmd)
}

// Config commands
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
}

var configViewCmd = &cobra.Command{
	Use:   "view",
	Short: "Show current configuration",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Configuration Sources:")
		fmt.Printf("  Config File:  ~/.config/nanofuse/config.yaml\n")
		fmt.Printf("  Environment:  NANOFUSE_* variables\n")
		fmt.Println()
		fmt.Println("Merged Configuration:")
		fmt.Printf("  API Socket:        %s\n", apiSocket)
		if apiURL != "" {
			fmt.Printf("  API URL:           %s\n", apiURL)
		}
		fmt.Printf("  API Timeout:       %v\n", timeout)
		fmt.Printf("  Default vCPUs:     2\n")
		fmt.Printf("  Default Memory:    512 MB\n")
		fmt.Printf("  Network Mode:      nat\n")
		fmt.Printf("  Output Format:     %s\n", map[bool]string{true: "json", false: "table"}[jsonOutput])
		fmt.Printf("  Color:             %s\n", map[bool]string{true: "never", false: "auto"}[noColor])
	},
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize default configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("get home directory: %w", err)
		}

		configPath := home + "/.config/nanofuse/config.yaml"

		// Check if exists
		if _, err := os.Stat(configPath); err == nil {
			return fmt.Errorf("config file already exists: %s", configPath)
		}

		// Create directory
		configDir := home + "/.config/nanofuse"
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("create config directory: %w", err)
		}

		// Write default config
		content := `# NanoFuse CLI Configuration

api:
  socket: /var/run/nanofused.sock
  timeout: 30s

defaults:
  vcpus: 2
  memory_mib: 512
  network_mode: nat

output:
  format: table
  color: auto
`

		if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
			return fmt.Errorf("write config file: %w", err)
		}

		formatter.PrintSuccess(fmt.Sprintf("Created config file: %s", configPath))
		fmt.Println("\nEdit the file to customize settings:")
		fmt.Printf("  vi %s\n", configPath)

		return nil
	},
}

func init() {
	configCmd.AddCommand(configViewCmd)
	configCmd.AddCommand(configInitCmd)
}

// Completion command
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate shell completion script for nanofuse.

To load completions:

Bash:
  $ source <(nanofuse completion bash)
  $ nanofuse completion bash > /etc/bash_completion.d/nanofuse

Zsh:
  $ nanofuse completion zsh > "${fpath[1]}/_nanofuse"

Fish:
  $ nanofuse completion fish > ~/.config/fish/completions/nanofuse.fish

PowerShell:
  PS> nanofuse completion powershell | Out-String | Invoke-Expression
`,
	Args:               cobra.ExactArgs(1),
	ValidArgs:          []string{"bash", "zsh", "fish", "powershell"},
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletion(os.Stdout)
		default:
			return fmt.Errorf("unsupported shell: %s", args[0])
		}
	},
}

// Helper functions

// resolveImageRef resolves image reference, handling special shorthands
func resolveImageRef(imageRef string) string {
	// If the image ref is just "default" or "base", expand to full default image
	switch imageRef {
	case "default", "base":
		return fmt.Sprintf("%s/%s:%s", DefaultImageRegistry, DefaultBaseImage, DefaultImageTag)
	case "default:latest", "base:latest":
		return fmt.Sprintf("%s/%s:latest", DefaultImageRegistry, DefaultBaseImage)
	}

	// If it starts with "default:" or "base:", expand with tag
	if strings.HasPrefix(imageRef, "default:") {
		tag := strings.TrimPrefix(imageRef, "default:")
		return fmt.Sprintf("%s/%s:%s", DefaultImageRegistry, DefaultBaseImage, tag)
	}
	if strings.HasPrefix(imageRef, "base:") {
		tag := strings.TrimPrefix(imageRef, "base:")
		return fmt.Sprintf("%s/%s:%s", DefaultImageRegistry, DefaultBaseImage, tag)
	}

	// Otherwise return as-is
	return imageRef
}

// handleAPIError converts an API error to a user-friendly CLIError and formats it.
// The operation parameter describes what was being attempted (e.g., "start VM", "pull image").
// This is a convenience wrapper that does not take a resource identifier and calls
// handleAPIErrorWithResource with an empty resource. If you have a specific resource
// (e.g., VM ID, image ref), use handleAPIErrorWithResource instead.
func handleAPIError(err error, operation string) error {
	return handleAPIErrorWithResource(err, operation, "")
}

// handleAPIErrorWithResource converts an API error to a user-friendly CLIError with resource context.
// The resource parameter identifies the resource being operated on (e.g., VM ID, image ref).
// For connection errors, it uses the selected API endpoint for context.
//
// Returns: Always returns an error whose concrete type is *clierrors.CLIError for consistent error handling.
// Callers can type-assert to *clierrors.CLIError and use cliErr.ExitCode for process exit codes.
func handleAPIErrorWithResource(err error, operation string, resource string) error {
	// For connection errors, use the selected TCP URL or Unix socket path.
	// For other errors, use the resource (VM ID, image ref, etc.)
	resourceOrSocket := resource
	if clierrors.IsConnectionError(err.Error()) {
		if apiURL != "" {
			resourceOrSocket = apiURL
		} else {
			resourceOrSocket = apiSocket
		}
		if resourceOrSocket == "" {
			resourceOrSocket = "/run/nanofused.sock"
		}
	}

	// Convert to CLIError with appropriate context
	cliErr := clierrors.FromError(err, operation, resourceOrSocket)

	// Always return CLIError for consistent type handling
	return cliErr
}

func isTerminal() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// readSSHPublicKey reads an SSH public key file and returns it base64 encoded
func readSSHPublicKey(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read SSH key file: %w", err)
	}

	key := strings.TrimSpace(string(data))
	if key == "" {
		return "", fmt.Errorf("SSH key file is empty")
	}

	// Validate it looks like a public key
	if !strings.HasPrefix(key, "ssh-") && !strings.HasPrefix(key, "ecdsa-") {
		return "", fmt.Errorf("invalid SSH public key format: must start with 'ssh-' or 'ecdsa-'")
	}

	return base64.StdEncoding.EncodeToString([]byte(key)), nil
}

// parsePortForwards parses port forward specifications
// Format: "hostPort:vmPort[/protocol]" where protocol defaults to "tcp"
// Examples: "8080:80", "8443:443/tcp", "53:53/udp"
func parsePortForwards(specs []string) ([]client.PortForward, error) {
	forwards := make([]client.PortForward, 0, len(specs))

	for _, spec := range specs {
		parts := strings.Split(spec, "/")
		portSpec := parts[0]
		protocol := "tcp" // default

		if len(parts) > 1 {
			protocol = parts[1]
		}

		if protocol != "tcp" && protocol != "udp" {
			return nil, fmt.Errorf("invalid protocol in %q: must be tcp or udp", spec)
		}

		portParts := strings.Split(portSpec, ":")
		if len(portParts) != 2 {
			return nil, fmt.Errorf("invalid port forward format %q: expected hostPort:vmPort[/protocol]", spec)
		}

		var hostPort, vmPort int
		if _, err := fmt.Sscanf(portParts[0], "%d", &hostPort); err != nil {
			return nil, fmt.Errorf("invalid host port in %q: %w", spec, err)
		}
		if _, err := fmt.Sscanf(portParts[1], "%d", &vmPort); err != nil {
			return nil, fmt.Errorf("invalid VM port in %q: %w", spec, err)
		}

		if hostPort < 1 || hostPort > 65535 {
			return nil, fmt.Errorf("invalid host port %d in %q: must be 1-65535", hostPort, spec)
		}
		if vmPort < 1 || vmPort > 65535 {
			return nil, fmt.Errorf("invalid VM port %d in %q: must be 1-65535", vmPort, spec)
		}

		forwards = append(forwards, client.PortForward{
			HostPort: hostPort,
			VMPort:   vmPort,
			Protocol: protocol,
		})
	}

	return forwards, nil
}
