//go:build windows

package main

import (
	_ "embed"

	"github.com/getlantern/systray"
)

//go:embed assets/icon.ico
var trayIcon []byte

// setTrayIcon applies the Windows notification-area icon.
func setTrayIcon() {
	if len(trayIcon) > 0 {
		systray.SetIcon(trayIcon)
	}
}
