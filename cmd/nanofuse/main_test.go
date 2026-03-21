package main

import (
	"testing"
)

// Basic smoke test to ensure package builds
func TestPackageBuilds(t *testing.T) {
	// If this test runs, the package compiled successfully
	t.Log("nanofuse CLI package builds successfully")
}

// TestVMLogsCommandExists verifies the vm logs command exists
func TestVMLogsCommandExists(t *testing.T) {
	if vmLogsCmd == nil {
		t.Fatal("vmLogsCmd is nil")
	}

	if vmLogsCmd.Use != "logs <vm-id>" {
		t.Errorf("Expected Use to be 'logs <vm-id>', got '%s'", vmLogsCmd.Use)
	}
}

// TestVMLogsCommandFlags verifies required flags exist
func TestVMLogsCommandFlags(t *testing.T) {
	flags := vmLogsCmd.Flags()

	// Check --tail flag exists
	tailFlag := flags.Lookup("tail")
	if tailFlag == nil {
		t.Fatal("--tail flag not found")
	}

	// Check --lines/-n flag exists
	linesFlag := flags.Lookup("lines")
	if linesFlag == nil {
		t.Fatal("--lines flag not found")
	}
	if linesFlag.Shorthand != "n" {
		t.Errorf("Expected shorthand 'n' for --lines, got '%s'", linesFlag.Shorthand)
	}

	// Check --follow/-f flag exists
	followFlag := flags.Lookup("follow")
	if followFlag == nil {
		t.Fatal("--follow flag not found")
	}
	if followFlag.Shorthand != "f" {
		t.Errorf("Expected shorthand 'f' for --follow, got '%s'", followFlag.Shorthand)
	}
}
