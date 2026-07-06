//go:build darwin

package main

import (
	_ "embed"

	"github.com/getlantern/systray"
)

//go:embed assets/icon_mac.png
var trayIcon []byte

// setTrayIcon applies the macOS menu-bar icon as a template image so the OS
// tints it for light/dark menu bars.
func setTrayIcon() {
	if len(trayIcon) > 0 {
		systray.SetTemplateIcon(trayIcon, trayIcon)
	}
}
