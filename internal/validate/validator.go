// Package validate provides filesystem image validation for NanoFuse microVM images.
// It verifies filesystem integrity, required directories, layer metadata, systemd
// configuration, and SSH setup.
package validate

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ValidationSeverity represents the severity of a validation issue.
type ValidationSeverity string

const (
	// SeverityError indicates a critical issue that fails validation.
	SeverityError ValidationSeverity = "error"
	// SeverityWarning indicates a non-critical issue.
	SeverityWarning ValidationSeverity = "warning"
	// SeverityInfo indicates informational output.
	SeverityInfo ValidationSeverity = "info"
)

// ValidationCheck represents a single validation check result.
type ValidationCheck struct {
	Name     string             `json:"name"`
	Category string             `json:"category"`
	Status   string             `json:"status"` // "passed", "failed", "skipped"
	Severity ValidationSeverity `json:"severity,omitempty"`
	Message  string             `json:"message,omitempty"`
	Details  string             `json:"details,omitempty"`
}

// BuildManifest represents the build-manifest.json contents.
type BuildManifest struct {
	Version     string            `json:"version"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	BuildDate   string            `json:"build_date"`
	Layers      []LayerInfo       `json:"layers,omitempty"`
	Kernel      KernelInfo        `json:"kernel,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

// LayerInfo represents layer information in the manifest.
type LayerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	Type    string `json:"type,omitempty"`
	SHA256  string `json:"sha256,omitempty"`
}

// KernelInfo represents kernel information in the manifest.
type KernelInfo struct {
	Version string `json:"version,omitempty"`
	Cmdline string `json:"cmdline,omitempty"`
}

// ValidationResult contains the complete validation result.
type ValidationResult struct {
	ImagePath     string            `json:"image_path"`
	ValidationAt  time.Time         `json:"validated_at"`
	Passed        bool              `json:"passed"`
	TotalChecks   int               `json:"total_checks"`
	PassedChecks  int               `json:"passed_checks"`
	FailedChecks  int               `json:"failed_checks"`
	WarningChecks int               `json:"warning_checks"`
	SkippedChecks int               `json:"skipped_checks"`
	Checks        []ValidationCheck `json:"checks"`
	Manifest      *BuildManifest    `json:"manifest,omitempty"`
}

// Validator performs image validation.
type Validator struct {
	imagePath string
	mountPath string
	strict    bool
	checks    []ValidationCheck
}

// NewValidator creates a new Validator for the given image path.
func NewValidator(imagePath string, strict bool) *Validator {
	return &Validator{
		imagePath: imagePath,
		strict:    strict,
		checks:    make([]ValidationCheck, 0),
	}
}

// addCheck adds a validation check result.
func (v *Validator) addCheck(name, category, status string, severity ValidationSeverity, message, details string) {
	v.checks = append(v.checks, ValidationCheck{
		Name:     name,
		Category: category,
		Status:   status,
		Severity: severity,
		Message:  message,
		Details:  details,
	})
}

// Validate performs all validation checks on the image.
// It requires the image to be mounted at mountPath.
func (v *Validator) Validate(mountPath string) (*ValidationResult, error) {
	v.mountPath = mountPath
	v.checks = make([]ValidationCheck, 0)

	// Run all validation checks
	v.checkFilesystemIntegrity()
	v.checkRequiredDirectories()
	v.checkLayerMetadata()
	v.checkSystemdConfiguration()
	v.checkSSHConfiguration()

	// Build result
	result := &ValidationResult{
		ImagePath:    v.imagePath,
		ValidationAt: time.Now().UTC(),
		Checks:       v.checks,
	}

	// Count results
	for _, check := range v.checks {
		result.TotalChecks++
		switch check.Status {
		case "passed":
			result.PassedChecks++
		case "failed":
			result.FailedChecks++
		case "skipped":
			result.SkippedChecks++
		}
		if check.Severity == SeverityWarning {
			result.WarningChecks++
		}
	}

	// In strict mode, warnings also count as failures
	if v.strict {
		result.Passed = result.FailedChecks == 0 && result.WarningChecks == 0
	} else {
		result.Passed = result.FailedChecks == 0
	}

	// Try to load manifest for metadata
	result.Manifest = v.loadManifest()

	return result, nil
}

// checkFilesystemIntegrity verifies the filesystem using fsck.
func (v *Validator) checkFilesystemIntegrity() {
	// Check if image file exists
	info, err := os.Stat(v.imagePath)
	if err != nil {
		v.addCheck("filesystem_exists", "filesystem", "failed", SeverityError,
			"Image file does not exist", err.Error())
		return
	}

	// Check if it's a regular file
	if !info.Mode().IsRegular() {
		v.addCheck("filesystem_regular", "filesystem", "failed", SeverityError,
			"Image path is not a regular file", fmt.Sprintf("Mode: %v", info.Mode()))
		return
	}

	v.addCheck("filesystem_exists", "filesystem", "passed", "",
		"Image file exists", fmt.Sprintf("Size: %d bytes", info.Size()))

	// Run e2fsck to check filesystem integrity (dry-run mode)
	// #nosec G204 -- imagePath is validated user input for CLI image validation tool
	cmd := exec.Command("e2fsck", "-n", "-f", v.imagePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// e2fsck returns non-zero for errors OR if corrections are needed
		// Check the specific exit code
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode := exitErr.ExitCode()
			// Exit codes: 0=no errors, 1=errors corrected, 2=errors corrected+reboot needed
			// 4=errors left uncorrected, 8=operational error, 16=usage error, 32=cancelled
			if exitCode >= 4 {
				v.addCheck("filesystem_integrity", "filesystem", "failed", SeverityError,
					"Filesystem has uncorrectable errors", string(output))
				return
			}
			// Exit codes 1-3 mean corrections needed but filesystem is usable
			v.addCheck("filesystem_integrity", "filesystem", "passed", SeverityWarning,
				"Filesystem has minor issues that could be corrected", string(output))
			return
		}
		// Command not found or other error
		v.addCheck("filesystem_integrity", "filesystem", "skipped", SeverityWarning,
			"Could not run filesystem check", err.Error())
		return
	}

	v.addCheck("filesystem_integrity", "filesystem", "passed", "",
		"Filesystem integrity verified", "e2fsck reports no errors")
}

// checkRequiredDirectories verifies required directories exist in the mounted filesystem.
func (v *Validator) checkRequiredDirectories() {
	requiredDirs := []string{
		"/etc",
		"/usr",
		"/var",
		"/home",
		"/root",
		"/tmp",
		"/dev",
		"/proc",
		"/sys",
		"/run",
	}

	allPresent := true
	var missingDirs []string

	for _, dir := range requiredDirs {
		fullPath := filepath.Join(v.mountPath, dir)
		info, err := os.Stat(fullPath)
		if err != nil {
			allPresent = false
			missingDirs = append(missingDirs, dir)
			continue
		}
		if !info.IsDir() {
			allPresent = false
			missingDirs = append(missingDirs, dir+" (not a directory)")
		}
	}

	if allPresent {
		v.addCheck("required_directories", "structure", "passed", "",
			"All required directories present", strings.Join(requiredDirs, ", "))
	} else {
		v.addCheck("required_directories", "structure", "failed", SeverityError,
			"Missing required directories", strings.Join(missingDirs, ", "))
	}
}

// checkLayerMetadata verifies the NanoFuse layer metadata is present and valid.
func (v *Validator) checkLayerMetadata() {
	manifestPath := filepath.Join(v.mountPath, "etc", "nanofuse", "build-manifest.json")

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			v.addCheck("layer_metadata", "metadata", "failed", SeverityError,
				"Build manifest not found", fmt.Sprintf("Expected: %s", manifestPath))
		} else {
			v.addCheck("layer_metadata", "metadata", "failed", SeverityError,
				"Cannot read build manifest", err.Error())
		}
		return
	}

	var manifest BuildManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		v.addCheck("layer_metadata", "metadata", "failed", SeverityError,
			"Invalid JSON in build manifest", err.Error())
		return
	}

	// Validate required fields
	var issues []string
	if manifest.Version == "" {
		issues = append(issues, "missing version")
	}
	if manifest.Name == "" {
		issues = append(issues, "missing name")
	}
	if manifest.BuildDate == "" {
		issues = append(issues, "missing build_date")
	}

	if len(issues) > 0 {
		v.addCheck("layer_metadata", "metadata", "failed", SeverityWarning,
			"Build manifest has missing fields", strings.Join(issues, ", "))
		return
	}

	details := fmt.Sprintf("Name: %s, Version: %s, Built: %s", manifest.Name, manifest.Version, manifest.BuildDate)
	if len(manifest.Layers) > 0 {
		details += fmt.Sprintf(", Layers: %d", len(manifest.Layers))
	}

	v.addCheck("layer_metadata", "metadata", "passed", "",
		"Build manifest is valid", details)
}

// checkSystemdConfiguration verifies systemd is properly configured.
func (v *Validator) checkSystemdConfiguration() {
	// Check for systemd directory
	systemdPath := filepath.Join(v.mountPath, "etc", "systemd", "system")
	if _, err := os.Stat(systemdPath); err != nil {
		if os.IsNotExist(err) {
			v.addCheck("systemd_directory", "systemd", "skipped", SeverityInfo,
				"Systemd system directory not found", "Image may not use systemd")
			return
		}
		v.addCheck("systemd_directory", "systemd", "failed", SeverityError,
			"Cannot access systemd directory", err.Error())
		return
	}

	v.addCheck("systemd_directory", "systemd", "passed", "",
		"Systemd directory exists", systemdPath)

	// Check for default.target or multi-user.target
	defaultTarget := filepath.Join(v.mountPath, "etc", "systemd", "system", "default.target")
	multiUserTarget := filepath.Join(v.mountPath, "lib", "systemd", "system", "multi-user.target")

	hasDefault := fileExists(defaultTarget)
	hasMultiUser := fileExists(multiUserTarget)

	if hasDefault || hasMultiUser {
		targetInfo := ""
		if hasDefault {
			target, _ := os.Readlink(defaultTarget)
			targetInfo = fmt.Sprintf("default.target -> %s", target)
		} else {
			targetInfo = "multi-user.target present"
		}
		v.addCheck("systemd_target", "systemd", "passed", "",
			"Boot target is configured", targetInfo)
	} else {
		v.addCheck("systemd_target", "systemd", "failed", SeverityWarning,
			"No default boot target found", "Expected default.target or multi-user.target")
	}

	// Check for getty service (serial console)
	serialGettys := []string{
		filepath.Join(v.mountPath, "etc", "systemd", "system", "getty.target.wants", "serial-getty@ttyS0.service"),
		filepath.Join(v.mountPath, "lib", "systemd", "system", "serial-getty@.service"),
	}

	hasSerial := false
	for _, path := range serialGettys {
		if fileExists(path) || symlinkExists(path) {
			hasSerial = true
			break
		}
	}

	if hasSerial {
		v.addCheck("systemd_serial_console", "systemd", "passed", "",
			"Serial console getty is configured", "Required for Firecracker console access")
	} else {
		v.addCheck("systemd_serial_console", "systemd", "failed", SeverityWarning,
			"Serial console getty not found", "Console access may not work in Firecracker")
	}
}

// checkSSHConfiguration verifies SSH is properly configured.
func (v *Validator) checkSSHConfiguration() {
	sshdConfig := filepath.Join(v.mountPath, "etc", "ssh", "sshd_config")

	data, err := os.ReadFile(sshdConfig)
	if err != nil {
		if os.IsNotExist(err) {
			// Check if SSH is even installed
			sshdBinary := filepath.Join(v.mountPath, "usr", "sbin", "sshd")
			if !fileExists(sshdBinary) {
				v.addCheck("ssh_installed", "ssh", "skipped", SeverityInfo,
					"SSH is not installed", "Image may not require SSH access")
				return
			}
			v.addCheck("ssh_config", "ssh", "failed", SeverityWarning,
				"SSHD config file not found", "SSH is installed but not configured")
			return
		}
		v.addCheck("ssh_config", "ssh", "failed", SeverityError,
			"Cannot read sshd_config", err.Error())
		return
	}

	// Parse sshd_config for important settings
	content := string(data)
	issues := []string{}
	info := []string{}

	// Check for root login settings
	if strings.Contains(content, "PermitRootLogin yes") {
		info = append(info, "PermitRootLogin=yes")
	} else if strings.Contains(content, "PermitRootLogin prohibit-password") {
		info = append(info, "PermitRootLogin=prohibit-password (keys only)")
	} else if strings.Contains(content, "PermitRootLogin no") {
		issues = append(issues, "root login disabled")
	}

	// Check for password authentication
	if strings.Contains(content, "PasswordAuthentication no") {
		info = append(info, "password auth disabled (key-only)")
	} else if strings.Contains(content, "PasswordAuthentication yes") {
		if v.strict {
			issues = append(issues, "password authentication enabled (security risk)")
		} else {
			info = append(info, "password auth enabled")
		}
	}

	// Check for AuthorizedKeysFile setting
	if strings.Contains(content, "AuthorizedKeysFile") {
		info = append(info, "custom authorized_keys path configured")
	}

	// Check for port
	if strings.Contains(content, "Port ") {
		info = append(info, "custom port configured")
	}

	if len(issues) > 0 {
		severity := SeverityWarning
		if v.strict {
			severity = SeverityError
		}
		v.addCheck("ssh_config", "ssh", "failed", severity,
			"SSH configuration issues", strings.Join(issues, "; "))
		return
	}

	details := "sshd_config is valid"
	if len(info) > 0 {
		details = strings.Join(info, ", ")
	}
	v.addCheck("ssh_config", "ssh", "passed", "",
		"SSH configuration is valid", details)

	// Check for host keys
	hostKeyDir := filepath.Join(v.mountPath, "etc", "ssh")
	hostKeys := []string{
		"ssh_host_rsa_key",
		"ssh_host_ecdsa_key",
		"ssh_host_ed25519_key",
	}

	hasHostKey := false
	for _, key := range hostKeys {
		if fileExists(filepath.Join(hostKeyDir, key)) {
			hasHostKey = true
			break
		}
	}

	if hasHostKey {
		v.addCheck("ssh_host_keys", "ssh", "passed", "",
			"SSH host keys are present", "Required for SSH connections")
	} else {
		v.addCheck("ssh_host_keys", "ssh", "failed", SeverityWarning,
			"No SSH host keys found", "Keys may be generated on first boot")
	}
}

// loadManifest attempts to load the build manifest from the mounted filesystem.
func (v *Validator) loadManifest() *BuildManifest {
	manifestPath := filepath.Join(v.mountPath, "etc", "nanofuse", "build-manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil
	}

	var manifest BuildManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil
	}

	return &manifest
}

// fileExists checks if a regular file exists at the given path.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode().IsRegular()
}

// symlinkExists checks if a symlink exists at the given path.
func symlinkExists(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}

// GenerateTextReport generates a human-readable text report.
func GenerateTextReport(result *ValidationResult, useColor bool) string {
	var sb strings.Builder

	// Header
	sb.WriteString("NanoFuse Image Validation Report\n")
	sb.WriteString("================================\n\n")

	sb.WriteString(fmt.Sprintf("Image:      %s\n", result.ImagePath))
	sb.WriteString(fmt.Sprintf("Validated:  %s\n", result.ValidationAt.Format("2006-01-02 15:04:05 UTC")))

	if result.Manifest != nil {
		sb.WriteString(fmt.Sprintf("Name:       %s\n", result.Manifest.Name))
		sb.WriteString(fmt.Sprintf("Version:    %s\n", result.Manifest.Version))
	}
	sb.WriteString("\n")

	// Summary
	status := "PASSED"
	if !result.Passed {
		status = "FAILED"
	}
	sb.WriteString(fmt.Sprintf("Status: %s\n", status))
	sb.WriteString(fmt.Sprintf("Total Checks: %d | Passed: %d | Failed: %d | Warnings: %d | Skipped: %d\n\n",
		result.TotalChecks, result.PassedChecks, result.FailedChecks, result.WarningChecks, result.SkippedChecks))

	// Group checks by category
	categories := make(map[string][]ValidationCheck)
	for _, check := range result.Checks {
		categories[check.Category] = append(categories[check.Category], check)
	}

	// Print checks by category
	for _, category := range []string{"filesystem", "structure", "metadata", "systemd", "ssh"} {
		checks, ok := categories[category]
		if !ok {
			continue
		}

		sb.WriteString(fmt.Sprintf("[%s]\n", strings.ToUpper(category)))
		for _, check := range checks {
			var statusSymbol string
			switch check.Status {
			case "passed":
				statusSymbol = "[OK]"
				if useColor {
					statusSymbol = "\033[32m[OK]\033[0m"
				}
			case "failed":
				statusSymbol = "[FAIL]"
				if useColor {
					statusSymbol = "\033[31m[FAIL]\033[0m"
				}
			case "skipped":
				statusSymbol = "[SKIP]"
				if useColor {
					statusSymbol = "\033[33m[SKIP]\033[0m"
				}
			}

			sb.WriteString(fmt.Sprintf("  %s %s\n", statusSymbol, check.Message))
			if check.Details != "" {
				sb.WriteString(fmt.Sprintf("       %s\n", check.Details))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// GenerateJSONReport generates a JSON report.
func GenerateJSONReport(result *ValidationResult) ([]byte, error) {
	return json.MarshalIndent(result, "", "  ")
}
