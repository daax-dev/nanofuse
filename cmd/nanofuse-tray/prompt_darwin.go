//go:build darwin

package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func promptImageReference(ctx context.Context) (string, bool, error) {
	script := `set dialogResult to display dialog "Container image reference:" default answer "docker.io/library/alpine:3.20" with title "New Nanofuse MicroVM" buttons {"Cancel", "Launch"} default button "Launch" cancel button "Cancel"
return text returned of dialogResult`
	output, err := exec.CommandContext(ctx, "osascript", "-e", script).CombinedOutput()
	if err != nil {
		text := strings.TrimSpace(string(output))
		if strings.Contains(text, "User canceled") {
			return "", false, nil
		}
		return "", false, fmt.Errorf("run image prompt: %w: %s", err, text)
	}

	imageRef := strings.TrimSpace(string(output))
	if imageRef == "" {
		return "", false, nil
	}
	return imageRef, true, nil
}
