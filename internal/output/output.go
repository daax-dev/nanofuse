package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/daax-dev/nanofuse/internal/client"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/schollz/progressbar/v3"
)

// Formatter handles output formatting
type Formatter struct {
	format   string // table, json
	useColor bool
}

// NewFormatter creates a new formatter
func NewFormatter(format string, useColor bool) *Formatter {
	// Disable color globally if needed
	if !useColor {
		color.NoColor = true
	}
	return &Formatter{
		format:   format,
		useColor: useColor,
	}
}

// PrintJSON outputs JSON
func (f *Formatter) PrintJSON(v interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func newPlainTable(headers []string) *tablewriter.Table {
	table := tablewriter.NewTable(
		os.Stdout,
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
			Borders: tw.Border{
				Left:   tw.Off,
				Right:  tw.Off,
				Top:    tw.Off,
				Bottom: tw.Off,
			},
			Settings: tw.Settings{
				Separators: tw.Separators{
					ShowHeader:     tw.On,
					ShowFooter:     tw.Off,
					BetweenRows:    tw.Off,
					BetweenColumns: tw.Off,
				},
				Lines: tw.Lines{
					ShowTop:        tw.Off,
					ShowBottom:     tw.Off,
					ShowHeaderLine: tw.Off,
					ShowFooterLine: tw.Off,
				},
			},
		})),
		tablewriter.WithHeaderAlignment(tw.AlignLeft),
		tablewriter.WithRowAlignment(tw.AlignLeft),
	)
	table.Header(headers)
	return table
}

// PrintVMList prints VM list
func (f *Formatter) PrintVMList(vms []client.VM) error {
	if f.format == "json" {
		return f.PrintJSON(map[string]interface{}{
			"vms":   vms,
			"total": len(vms),
		})
	}

	table := newPlainTable([]string{"ID", "Name", "State", "Image", "Ports", "CPU", "Mem", "Uptime"})

	for _, vm := range vms {
		id := vm.ID
		if len(id) > 8 {
			id = id[:8]
		}

		image := vm.Image
		if len(image) > 40 {
			image = image[:37] + "..."
		}

		state := f.colorizeState(vm.State)
		memory := fmt.Sprintf("%dM", vm.Config.MemoryMiB)

		uptime := "-"
		if vm.UptimeSeconds != nil && *vm.UptimeSeconds > 0 {
			uptime = formatDuration(time.Duration(*vm.UptimeSeconds) * time.Second)
		}

		if err := table.Append([]string{
			id,
			vm.Name,
			state,
			image,
			formatPortForwards(vm.Config.Network.PortForwards),
			fmt.Sprintf("%d", vm.Config.VCPUs),
			memory,
			uptime,
		}); err != nil {
			return err
		}
	}

	return table.Render()
}

// PrintVM prints single VM details
func (f *Formatter) PrintVM(vm *client.VM) error {
	if f.format == "json" {
		return f.PrintJSON(vm)
	}

	fmt.Printf("ID:              %s\n", vm.ID)
	fmt.Printf("Name:            %s\n", vm.Name)
	fmt.Printf("State:           %s\n", f.colorizeState(vm.State))
	fmt.Printf("Image:           %s\n", vm.Image)
	fmt.Printf("Architecture:    %s\n", vm.Architecture)
	fmt.Println()
	fmt.Println("Configuration:")
	fmt.Printf("  vCPUs:         %d\n", vm.Config.VCPUs)
	fmt.Printf("  Memory:        %d MB\n", vm.Config.MemoryMiB)
	fmt.Printf("  Network:       %s\n", vm.Config.Network.Mode)
	if len(vm.Config.Network.PortForwards) > 0 {
		fmt.Println("  Port Forwards:")
		for _, pf := range vm.Config.Network.PortForwards {
			protocol := pf.Protocol
			if protocol == "" {
				protocol = "tcp"
			}
			fmt.Printf("    127.0.0.1:%d -> vm:%d/%s\n", pf.HostPort, pf.VMPort, protocol)
		}
	}
	if vm.Config.KernelArgs != "" {
		fmt.Printf("  Kernel Args:   %s\n", vm.Config.KernelArgs)
	}
	printVMMounts(vm.Config.Mounts)
	printVMSecrets(vm.Config.Secrets)

	if vm.Runtime != nil {
		fmt.Println()
		fmt.Println("Runtime:")
		if vm.Runtime.Driver != "" {
			fmt.Printf("  Driver:        %s\n", vm.Runtime.Driver)
		}
		if vm.Runtime.ExternalID != "" {
			fmt.Printf("  Runtime ID:    %s\n", vm.Runtime.ExternalID)
		}
		if vm.Runtime.PID != 0 {
			fmt.Printf("  PID:           %d\n", vm.Runtime.PID)
		}
		if vm.UptimeSeconds != nil {
			uptime := formatDuration(time.Duration(*vm.UptimeSeconds) * time.Second)
			fmt.Printf("  Uptime:        %s\n", uptime)
		}
		if vm.Runtime.NetworkInfo.GuestIP != "" {
			fmt.Printf("  IP Address:    %s\n", vm.Runtime.NetworkInfo.GuestIP)
			fmt.Printf("  Gateway:       %s\n", vm.Runtime.NetworkInfo.Gateway)
			fmt.Printf("  TAP Device:    %s\n", vm.Runtime.NetworkInfo.TAPDevice)
		}
		if vm.Runtime.ConsolePath != "" {
			fmt.Printf("  Console:       %s\n", vm.Runtime.ConsolePath)
		}
	}

	fmt.Println()
	fmt.Println("Timestamps:")
	fmt.Printf("  Created:       %s\n", formatTime(vm.CreatedAt))
	fmt.Printf("  Last Updated:  %s\n", formatTime(vm.UpdatedAt))

	return nil
}

// printVMMounts renders the operator-visible mount inventory for a VM.
func printVMMounts(mounts []client.Mount) {
	if len(mounts) == 0 {
		return
	}
	fmt.Println("  Mounts:")
	for _, m := range mounts {
		mode := "rw"
		if m.ReadOnly {
			mode = "ro"
		}
		typ := m.Type
		if typ == "" {
			typ = "bind"
		}
		source := m.Source
		if source == "" {
			source = "-"
		}
		fmt.Printf("    %s %s -> %s (%s)\n", typ, source, m.Target, mode)
	}
}

// printVMSecrets renders the secret-reference inventory for a VM (never values).
func printVMSecrets(secrets []client.SecretRef) {
	if len(secrets) == 0 {
		return
	}
	fmt.Println("  Secret Refs:")
	for _, s := range secrets {
		typ := s.Type
		if typ == "" {
			typ = "env"
		}
		target := s.Target
		if target == "" {
			target = s.Name
		}
		source := s.Source
		if source == "" {
			source = "-"
		}
		fmt.Printf("    %s (%s) source=%s target=%s\n", s.Name, typ, source, target)
	}
}

func formatPortForwards(portForwards []client.PortForward) string {
	if len(portForwards) == 0 {
		return "-"
	}

	parts := make([]string, 0, len(portForwards))
	for _, pf := range portForwards {
		protocol := pf.Protocol
		if protocol == "" {
			protocol = "tcp"
		}
		parts = append(parts, fmt.Sprintf("127.0.0.1:%d->%d/%s", pf.HostPort, pf.VMPort, protocol))
	}
	return strings.Join(parts, ",")
}

// PrintImageList prints image list
func (f *Formatter) PrintImageList(images []client.Image) error {
	if f.format == "json" {
		return f.PrintJSON(map[string]interface{}{
			"images": images,
			"total":  len(images),
		})
	}

	table := newPlainTable([]string{"Digest", "Tags", "Size", "Pulled"})

	for _, img := range images {
		digest := img.Digest
		if len(digest) > 16 {
			digest = digest[:16] + "..."
		}

		tags := strings.Join(img.Tags, ", ")
		if len(tags) > 50 {
			tags = tags[:47] + "..."
		}

		size := formatBytes(img.SizeBytes)
		pulled := formatTimeAgo(img.PulledAt)

		if err := table.Append([]string{digest, tags, size, pulled}); err != nil {
			return err
		}
	}

	return table.Render()
}

// PrintImage prints single image details
func (f *Formatter) PrintImage(img *client.Image) error {
	if f.format == "json" {
		return f.PrintJSON(img)
	}

	fmt.Printf("Digest:        %s\n", img.Digest)
	fmt.Printf("Tags:          %s\n", strings.Join(img.Tags, ", "))
	fmt.Printf("Architecture:  %s\n", img.Architecture)
	fmt.Printf("Size:          %s\n", formatBytes(img.SizeBytes))
	if img.KernelVersion != "" {
		fmt.Printf("Kernel:        %s\n", img.KernelVersion)
	}
	fmt.Printf("Rootfs:        %s\n", img.RootfsPath)
	fmt.Printf("Kernel Path:   %s\n", img.KernelPath)
	fmt.Printf("Pulled:        %s\n", formatTime(img.PulledAt))

	// Print labels if available
	if len(img.Labels) > 0 {
		fmt.Printf("\nLabels:\n")
		// Print OCI labels first
		for k, v := range img.Labels {
			if strings.HasPrefix(k, "org.opencontainers.image.") {
				fmt.Printf("  %-30s %s\n", k, v)
			}
		}
		// Then NanoFuse labels
		for k, v := range img.Labels {
			if strings.HasPrefix(k, "com.nanofuse.") {
				fmt.Printf("  %-30s %s\n", k, v)
			}
		}
		// Then other labels
		for k, v := range img.Labels {
			if !strings.HasPrefix(k, "org.opencontainers.image.") && !strings.HasPrefix(k, "com.nanofuse.") {
				fmt.Printf("  %-30s %s\n", k, v)
			}
		}
	}

	return nil
}

// PrintSnapshotList prints snapshot list
func (f *Formatter) PrintSnapshotList(snapshots []client.Snapshot) error {
	if f.format == "json" {
		return f.PrintJSON(map[string]interface{}{
			"snapshots": snapshots,
			"total":     len(snapshots),
		})
	}

	table := newPlainTable([]string{"ID", "Name", "Size", "Created"})

	for _, snap := range snapshots {
		name := snap.Name
		if name == "" {
			name = "-"
		}

		size := formatBytes(snap.SizeBytes)
		created := formatTimeAgo(snap.CreatedAt)

		if err := table.Append([]string{snap.ID, name, size, created}); err != nil {
			return err
		}
	}

	return table.Render()
}

// PrintSnapshot prints single snapshot details
func (f *Formatter) PrintSnapshot(snap *client.Snapshot) error {
	if f.format == "json" {
		return f.PrintJSON(snap)
	}

	fmt.Printf("ID:           %s\n", snap.ID)
	fmt.Printf("VM ID:        %s\n", snap.VMID)
	if snap.Name != "" {
		fmt.Printf("Name:         %s\n", snap.Name)
	}
	fmt.Printf("Size:         %s\n", formatBytes(snap.SizeBytes))
	fmt.Printf("Created:      %s\n", formatTime(snap.CreatedAt))
	fmt.Println()
	fmt.Println("Files:")
	fmt.Printf("  Memory:     %s\n", snap.MemoryFilePath)
	fmt.Printf("  Device:     %s\n", snap.SnapshotFilePath)

	return nil
}

// PrintHealth prints health status
func (f *Formatter) PrintHealth(health *client.HealthResponse) error {
	if f.format == "json" {
		return f.PrintJSON(health)
	}

	fmt.Printf("Status:   %s\n", f.colorizeHealth(health.Status))
	fmt.Printf("Version:  %s\n", health.Version)
	uptime := formatDuration(time.Duration(health.UptimeSeconds) * time.Second)
	fmt.Printf("Uptime:   %s\n", uptime)

	return nil
}

// PrintSuccess prints a success message
func (f *Formatter) PrintSuccess(msg string) {
	if f.useColor {
		color.Green("✓ %s", msg)
	} else {
		fmt.Println(msg)
	}
}

// PrintError prints an error message
func (f *Formatter) PrintError(msg string) {
	if f.useColor {
		color.Red("✗ %s", msg)
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
	}
}

// PrintHint prints a hint message
func (f *Formatter) PrintHint(msg string) {
	if f.useColor {
		color.Yellow("Hint: %s", msg)
	} else {
		fmt.Fprintf(os.Stderr, "Hint: %s\n", msg)
	}
}

// NewProgressBar creates a progress bar
func NewProgressBar(max int64, description string) *progressbar.ProgressBar {
	return progressbar.NewOptions64(
		max,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(40),
		progressbar.OptionThrottle(100*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
	)
}

// Helper functions

func (f *Formatter) colorizeState(state string) string {
	if !f.useColor {
		return state
	}

	switch state {
	case "running":
		return color.GreenString(state)
	case "paused":
		return color.YellowString(state)
	case "stopped", "created":
		return color.HiBlackString(state)
	case "failed":
		return color.RedString(state)
	case "starting", "stopping", "pausing", "resuming":
		return color.BlueString(state)
	default:
		return state
	}
}

func (f *Formatter) colorizeHealth(status string) string {
	if !f.useColor {
		return status
	}

	if status == "healthy" {
		return color.GreenString(status)
	}
	return color.RedString(status)
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

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0fm", d.Minutes())
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	if hours < 24 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	days := hours / 24
	hours = hours % 24
	return fmt.Sprintf("%dd %dh", days, hours)
}

func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05 MST")
}

func formatTimeAgo(t time.Time) string {
	duration := time.Since(t)
	if duration < time.Minute {
		return "just now"
	}
	if duration < time.Hour {
		mins := int(duration.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	}
	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}
	days := int(duration.Hours() / 24)
	if days == 1 {
		return "1 day ago"
	}
	if days < 30 {
		return fmt.Sprintf("%d days ago", days)
	}
	return formatTime(t)
}
