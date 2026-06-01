//go:build windows

package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func promptImageReference(ctx context.Context) (string, bool, error) {
	command := `Add-Type -AssemblyName Microsoft.VisualBasic; ` +
		`$ref = [Microsoft.VisualBasic.Interaction]::InputBox('Container image reference:', 'New Nanofuse MicroVM', 'docker.io/library/alpine:3.20'); ` +
		`if ($ref) { Write-Output $ref }`
	output, err := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", command).CombinedOutput()
	if err != nil {
		return "", false, fmt.Errorf("run image prompt: %w: %s", err, strings.TrimSpace(string(output)))
	}

	imageRef := strings.TrimSpace(string(output))
	if imageRef == "" {
		return "", false, nil
	}
	return imageRef, true, nil
}
