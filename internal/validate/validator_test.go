package validate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNewValidator(t *testing.T) {
	v := NewValidator("/path/to/image.ext4", false)
	if v == nil {
		t.Fatal("NewValidator returned nil")
	}
	if v.imagePath != "/path/to/image.ext4" {
		t.Errorf("imagePath = %q, want %q", v.imagePath, "/path/to/image.ext4")
	}
	if v.strict {
		t.Error("strict = true, want false")
	}

	v2 := NewValidator("/path/to/image.ext4", true)
	if !v2.strict {
		t.Error("strict = false, want true")
	}
}

func TestCheckRequiredDirectories(t *testing.T) {
	// Create a temp directory structure
	tmpDir := t.TempDir()

	// Create required directories
	requiredDirs := []string{
		"etc",
		"usr",
		"var",
		"home",
		"root",
		"tmp",
		"dev",
		"proc",
		"sys",
		"run",
	}

	for _, dir := range requiredDirs {
		if err := os.MkdirAll(filepath.Join(tmpDir, dir), 0755); err != nil {
			t.Fatalf("Failed to create dir %s: %v", dir, err)
		}
	}

	v := NewValidator(tmpDir, false)
	v.mountPath = tmpDir
	v.checkRequiredDirectories()

	// Should have passed
	found := false
	for _, check := range v.checks {
		if check.Name == "required_directories" {
			found = true
			if check.Status != "passed" {
				t.Errorf("required_directories status = %q, want 'passed'", check.Status)
			}
		}
	}
	if !found {
		t.Error("required_directories check not found")
	}
}

func TestCheckRequiredDirectoriesMissing(t *testing.T) {
	// Create a temp directory with missing directories
	tmpDir := t.TempDir()

	// Only create some directories
	for _, dir := range []string{"etc", "usr", "var"} {
		if err := os.MkdirAll(filepath.Join(tmpDir, dir), 0755); err != nil {
			t.Fatalf("Failed to create dir %s: %v", dir, err)
		}
	}

	v := NewValidator(tmpDir, false)
	v.mountPath = tmpDir
	v.checkRequiredDirectories()

	// Should have failed
	found := false
	for _, check := range v.checks {
		if check.Name == "required_directories" {
			found = true
			if check.Status != "failed" {
				t.Errorf("required_directories status = %q, want 'failed'", check.Status)
			}
			if check.Severity != SeverityError {
				t.Errorf("required_directories severity = %q, want 'error'", check.Severity)
			}
		}
	}
	if !found {
		t.Error("required_directories check not found")
	}
}

func TestCheckLayerMetadata(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nanofuse metadata directory and manifest
	metadataDir := filepath.Join(tmpDir, "etc", "nanofuse")
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		t.Fatalf("Failed to create metadata dir: %v", err)
	}

	manifest := BuildManifest{
		Version:   "1.0",
		Name:      "test-image",
		BuildDate: "2024-12-30T00:00:00Z",
		Layers: []LayerInfo{
			{Name: "base", Version: "1.0", Type: "base"},
		},
	}

	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("Failed to marshal manifest: %v", err)
	}

	if err := os.WriteFile(filepath.Join(metadataDir, "build-manifest.json"), data, 0644); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	v := NewValidator(tmpDir, false)
	v.mountPath = tmpDir
	v.checkLayerMetadata()

	// Should have passed
	found := false
	for _, check := range v.checks {
		if check.Name == "layer_metadata" {
			found = true
			if check.Status != "passed" {
				t.Errorf("layer_metadata status = %q, want 'passed', message: %s", check.Status, check.Message)
			}
		}
	}
	if !found {
		t.Error("layer_metadata check not found")
	}
}

func TestCheckLayerMetadataMissing(t *testing.T) {
	tmpDir := t.TempDir()

	v := NewValidator(tmpDir, false)
	v.mountPath = tmpDir
	v.checkLayerMetadata()

	// Should have failed
	found := false
	for _, check := range v.checks {
		if check.Name == "layer_metadata" {
			found = true
			if check.Status != "failed" {
				t.Errorf("layer_metadata status = %q, want 'failed'", check.Status)
			}
		}
	}
	if !found {
		t.Error("layer_metadata check not found")
	}
}

func TestCheckLayerMetadataInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nanofuse metadata directory with invalid JSON
	metadataDir := filepath.Join(tmpDir, "etc", "nanofuse")
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		t.Fatalf("Failed to create metadata dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(metadataDir, "build-manifest.json"), []byte("{invalid json"), 0644); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	v := NewValidator(tmpDir, false)
	v.mountPath = tmpDir
	v.checkLayerMetadata()

	// Should have failed
	found := false
	for _, check := range v.checks {
		if check.Name == "layer_metadata" {
			found = true
			if check.Status != "failed" {
				t.Errorf("layer_metadata status = %q, want 'failed'", check.Status)
			}
			if check.Severity != SeverityError {
				t.Errorf("layer_metadata severity = %q, want 'error'", check.Severity)
			}
		}
	}
	if !found {
		t.Error("layer_metadata check not found")
	}
}

func TestCheckSystemdConfiguration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create systemd directory structure
	systemdDir := filepath.Join(tmpDir, "etc", "systemd", "system")
	if err := os.MkdirAll(systemdDir, 0755); err != nil {
		t.Fatalf("Failed to create systemd dir: %v", err)
	}

	// Create lib/systemd/system for multi-user.target
	libSystemdDir := filepath.Join(tmpDir, "lib", "systemd", "system")
	if err := os.MkdirAll(libSystemdDir, 0755); err != nil {
		t.Fatalf("Failed to create lib systemd dir: %v", err)
	}

	// Create multi-user.target
	if err := os.WriteFile(filepath.Join(libSystemdDir, "multi-user.target"), []byte("[Unit]\nDescription=Multi-User System"), 0644); err != nil {
		t.Fatalf("Failed to write multi-user.target: %v", err)
	}

	// Create serial-getty service
	if err := os.WriteFile(filepath.Join(libSystemdDir, "serial-getty@.service"), []byte("[Unit]\nDescription=Serial Getty"), 0644); err != nil {
		t.Fatalf("Failed to write serial-getty: %v", err)
	}

	v := NewValidator(tmpDir, false)
	v.mountPath = tmpDir
	v.checkSystemdConfiguration()

	// Should have passed for directory and target
	for _, check := range v.checks {
		if check.Name == "systemd_directory" && check.Status != "passed" {
			t.Errorf("systemd_directory status = %q, want 'passed'", check.Status)
		}
		if check.Name == "systemd_target" && check.Status != "passed" {
			t.Errorf("systemd_target status = %q, want 'passed'", check.Status)
		}
		if check.Name == "systemd_serial_console" && check.Status != "passed" {
			t.Errorf("systemd_serial_console status = %q, want 'passed'", check.Status)
		}
	}
}

func TestCheckSystemdNotPresent(t *testing.T) {
	tmpDir := t.TempDir()

	v := NewValidator(tmpDir, false)
	v.mountPath = tmpDir
	v.checkSystemdConfiguration()

	// Should be skipped (no systemd)
	found := false
	for _, check := range v.checks {
		if check.Name == "systemd_directory" {
			found = true
			if check.Status != "skipped" {
				t.Errorf("systemd_directory status = %q, want 'skipped'", check.Status)
			}
		}
	}
	if !found {
		t.Error("systemd_directory check not found")
	}
}

func TestCheckSSHConfiguration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create SSH directory and config
	sshDir := filepath.Join(tmpDir, "etc", "ssh")
	if err := os.MkdirAll(sshDir, 0755); err != nil {
		t.Fatalf("Failed to create ssh dir: %v", err)
	}

	// Create sshd_config
	sshdConfig := `# SSH Server Configuration
Port 22
PermitRootLogin prohibit-password
PasswordAuthentication no
PubkeyAuthentication yes
AuthorizedKeysFile .ssh/authorized_keys
`
	if err := os.WriteFile(filepath.Join(sshDir, "sshd_config"), []byte(sshdConfig), 0644); err != nil {
		t.Fatalf("Failed to write sshd_config: %v", err)
	}

	// Create a host key
	if err := os.WriteFile(filepath.Join(sshDir, "ssh_host_ed25519_key"), []byte("fake-key"), 0600); err != nil {
		t.Fatalf("Failed to write host key: %v", err)
	}

	// Create sshd binary to indicate SSH is installed
	sbinDir := filepath.Join(tmpDir, "usr", "sbin")
	if err := os.MkdirAll(sbinDir, 0755); err != nil {
		t.Fatalf("Failed to create sbin dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sbinDir, "sshd"), []byte("fake-binary"), 0755); err != nil {
		t.Fatalf("Failed to write sshd: %v", err)
	}

	v := NewValidator(tmpDir, false)
	v.mountPath = tmpDir
	v.checkSSHConfiguration()

	// Should have passed for config and host keys
	for _, check := range v.checks {
		if check.Name == "ssh_config" && check.Status != "passed" {
			t.Errorf("ssh_config status = %q, want 'passed', message: %s", check.Status, check.Message)
		}
		if check.Name == "ssh_host_keys" && check.Status != "passed" {
			t.Errorf("ssh_host_keys status = %q, want 'passed'", check.Status)
		}
	}
}

func TestCheckSSHNotInstalled(t *testing.T) {
	tmpDir := t.TempDir()

	v := NewValidator(tmpDir, false)
	v.mountPath = tmpDir
	v.checkSSHConfiguration()

	// Should be skipped (no SSH installed)
	found := false
	for _, check := range v.checks {
		if check.Name == "ssh_installed" {
			found = true
			if check.Status != "skipped" {
				t.Errorf("ssh_installed status = %q, want 'skipped'", check.Status)
			}
		}
	}
	if !found {
		t.Error("ssh_installed check not found")
	}
}

func TestValidationResult(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a minimal valid structure
	for _, dir := range []string{"etc", "usr", "var", "home", "root", "tmp", "dev", "proc", "sys", "run"} {
		if err := os.MkdirAll(filepath.Join(tmpDir, dir), 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
	}

	// Create a valid manifest
	metadataDir := filepath.Join(tmpDir, "etc", "nanofuse")
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		t.Fatalf("Failed to create metadata dir: %v", err)
	}

	manifest := BuildManifest{
		Version:   "1.0",
		Name:      "test-image",
		BuildDate: "2024-12-30T00:00:00Z",
	}
	data, _ := json.Marshal(manifest)
	if err := os.WriteFile(filepath.Join(metadataDir, "build-manifest.json"), data, 0644); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	// Create a valid ext4 image file for the validator
	imageFile := filepath.Join(tmpDir, "test.ext4")
	if err := os.WriteFile(imageFile, []byte("fake image"), 0644); err != nil {
		t.Fatalf("Failed to write image file: %v", err)
	}

	v := NewValidator(imageFile, false)
	result, err := v.Validate(tmpDir)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	if result.ImagePath != imageFile {
		t.Errorf("ImagePath = %q, want %q", result.ImagePath, imageFile)
	}

	if result.TotalChecks == 0 {
		t.Error("TotalChecks = 0, want > 0")
	}

	if result.Manifest == nil {
		t.Error("Manifest is nil, expected to be loaded")
	}
}

func TestGenerateTextReport(t *testing.T) {
	result := &ValidationResult{
		ImagePath:     "/path/to/image.ext4",
		Passed:        true,
		TotalChecks:   5,
		PassedChecks:  4,
		FailedChecks:  0,
		WarningChecks: 1,
		SkippedChecks: 0,
		Checks: []ValidationCheck{
			{Name: "test_check", Category: "filesystem", Status: "passed", Message: "Test passed"},
		},
	}

	report := GenerateTextReport(result, false)

	if report == "" {
		t.Error("GenerateTextReport returned empty string")
	}

	// Check for expected content
	if !contains(report, "NanoFuse Image Validation Report") {
		t.Error("Report missing header")
	}
	if !contains(report, "/path/to/image.ext4") {
		t.Error("Report missing image path")
	}
	if !contains(report, "PASSED") {
		t.Error("Report missing status")
	}
	if !contains(report, "[FILESYSTEM]") {
		t.Error("Report missing filesystem section")
	}
}

func TestGenerateJSONReport(t *testing.T) {
	result := &ValidationResult{
		ImagePath:    "/path/to/image.ext4",
		Passed:       true,
		TotalChecks:  5,
		PassedChecks: 5,
		Checks: []ValidationCheck{
			{Name: "test_check", Category: "filesystem", Status: "passed", Message: "Test passed"},
		},
	}

	data, err := GenerateJSONReport(result)
	if err != nil {
		t.Fatalf("GenerateJSONReport failed: %v", err)
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Generated JSON is invalid: %v", err)
	}

	// Check for expected fields
	if _, ok := parsed["image_path"]; !ok {
		t.Error("JSON missing image_path field")
	}
	if _, ok := parsed["passed"]; !ok {
		t.Error("JSON missing passed field")
	}
	if _, ok := parsed["checks"]; !ok {
		t.Error("JSON missing checks field")
	}
}

func TestStrictMode(t *testing.T) {
	tmpDir := t.TempDir()

	// Create minimal structure
	for _, dir := range []string{"etc", "usr", "var", "home", "root", "tmp", "dev", "proc", "sys", "run"} {
		if err := os.MkdirAll(filepath.Join(tmpDir, dir), 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
	}

	// Create image file
	imageFile := filepath.Join(tmpDir, "test.ext4")
	if err := os.WriteFile(imageFile, []byte("fake image"), 0644); err != nil {
		t.Fatalf("Failed to write image file: %v", err)
	}

	// Non-strict mode - should pass with warnings
	v1 := NewValidator(imageFile, false)
	result1, _ := v1.Validate(tmpDir)

	// Strict mode - warnings count as failures
	v2 := NewValidator(imageFile, true)
	result2, _ := v2.Validate(tmpDir)

	// The strict result should have the same checks but may have different passed status
	if result1.TotalChecks != result2.TotalChecks {
		t.Errorf("TotalChecks mismatch: non-strict=%d, strict=%d", result1.TotalChecks, result2.TotalChecks)
	}
}

func TestFileExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create a directory
	testDir := filepath.Join(tmpDir, "testdir")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}

	if !fileExists(testFile) {
		t.Error("fileExists returned false for existing file")
	}

	if fileExists(testDir) {
		t.Error("fileExists returned true for directory")
	}

	if fileExists(filepath.Join(tmpDir, "nonexistent")) {
		t.Error("fileExists returned true for nonexistent file")
	}
}

func TestSymlinkExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file and symlink
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	testLink := filepath.Join(tmpDir, "test-link")
	if err := os.Symlink(testFile, testLink); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	if !symlinkExists(testLink) {
		t.Error("symlinkExists returned false for existing symlink")
	}

	if symlinkExists(testFile) {
		t.Error("symlinkExists returned true for regular file")
	}

	if symlinkExists(filepath.Join(tmpDir, "nonexistent")) {
		t.Error("symlinkExists returned true for nonexistent path")
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
