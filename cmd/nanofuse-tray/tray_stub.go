//go:build !darwin && !windows

package main

import (
	"context"
	"fmt"
	"runtime"

	"github.com/daax-dev/nanofuse/internal/trayapp"
)

func runTray(context.Context, trayapp.Config) error {
	return fmt.Errorf("nanofuse-tray GUI is implemented for macOS and Windows; %s supports --smoke only", runtime.GOOS)
}
