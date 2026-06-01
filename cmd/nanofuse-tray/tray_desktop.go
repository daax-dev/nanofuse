//go:build darwin || windows

package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/daax-dev/nanofuse/internal/client"
	"github.com/daax-dev/nanofuse/internal/trayapp"
	"github.com/getlantern/systray"
)

const (
	maxMenuRows        = 25
	confirmationWindow = 10 * time.Second
)

type trayUI struct {
	ctx       context.Context
	cfg       trayapp.Config
	api       trayapp.API
	mu        sync.Mutex
	status    *trayapp.Status
	selected  string
	imageRef  string
	pending   trayapp.VMAction
	pendingAt time.Time
	refreshID uint64

	endpointItem *systray.MenuItem
	statusItem   *systray.MenuItem
	runtimeItem  *systray.MenuItem
	selectedItem *systray.MenuItem
	imageItem    *systray.MenuItem
	refreshItem  *systray.MenuItem
	createItem   *systray.MenuItem
	startItem    *systray.MenuItem
	stopItem     *systray.MenuItem
	killItem     *systray.MenuItem
	deleteItem   *systray.MenuItem
	vmItems      []*systray.MenuItem
	imageItems   []*systray.MenuItem
	quitItem     *systray.MenuItem
}

func runTray(ctx context.Context, cfg trayapp.Config) error {
	ui := &trayUI{
		ctx: ctx,
		cfg: cfg,
		api: cfg.NewClient(),
	}
	ready := func() { ui.onReady() }
	exit := func() {}

	go func() {
		<-ctx.Done()
		systray.Quit()
	}()

	systray.Run(ready, exit)
	return nil
}

func withTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		timeout = trayapp.DefaultTimeout
	}
	return context.WithTimeout(ctx, timeout)
}

func (ui *trayUI) onReady() {
	systray.SetTitle("NF")
	systray.SetTooltip("Nanofuse")

	ui.endpointItem = systray.AddMenuItem("Endpoint: "+ui.cfg.Endpoint(), "Configured nanofused API endpoint")
	ui.endpointItem.Disable()
	ui.statusItem = systray.AddMenuItem("Status: checking", "Daemon health")
	ui.statusItem.Disable()
	ui.runtimeItem = systray.AddMenuItem("Runtime: checking", "Runtime capabilities")
	ui.runtimeItem.Disable()
	systray.AddSeparator()

	ui.refreshItem = systray.AddMenuItem("Refresh", "Refresh daemon status, VMs, and images")
	ui.selectedItem = systray.AddMenuItem("Selected VM: none", "Currently selected VM")
	ui.selectedItem.Disable()
	ui.imageItem = systray.AddMenuItem("Selected Image: none", "Image used for new VM launches")
	ui.imageItem.Disable()
	ui.createItem = systray.AddMenuItem("Create and Start VM From Image", "Create and start a new VM from the selected image through the API")
	ui.startItem = systray.AddMenuItem("Start Selected VM", "Start the selected VM through the API")
	ui.stopItem = systray.AddMenuItem("Stop Selected VM", "Stop the selected VM through the API")
	ui.killItem = systray.AddMenuItem("Kill Selected VM", "Confirm, then force kill the selected VM through the API")
	ui.deleteItem = systray.AddMenuItem("Delete Selected VM", "Confirm, then delete the selected VM through the API")
	systray.AddSeparator()

	vmHeader := systray.AddMenuItem("VMs", "Known VMs")
	vmHeader.Disable()
	for i := 0; i < maxMenuRows; i++ {
		item := systray.AddMenuItem(fmt.Sprintf("VM slot %d", i+1), "Select VM")
		item.Hide()
		ui.vmItems = append(ui.vmItems, item)
	}
	systray.AddSeparator()

	imageHeader := systray.AddMenuItem("Images", "Cached images")
	imageHeader.Disable()
	for i := 0; i < maxMenuRows; i++ {
		item := systray.AddMenuItem(fmt.Sprintf("Image slot %d", i+1), "Cached image")
		item.Hide()
		item.Disable()
		ui.imageItems = append(ui.imageItems, item)
	}
	systray.AddSeparator()

	ui.quitItem = systray.AddMenuItem("Quit", "Quit Nanofuse tray")

	ui.updateActionState()
	ui.listen()
	ui.refresh()
}

func (ui *trayUI) listen() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ui.ctx.Done():
				systray.Quit()
				return
			case <-ticker.C:
				ui.refresh()
			case <-ui.refreshItem.ClickedCh:
				ui.refresh()
			case <-ui.createItem.ClickedCh:
				ui.launchSelectedImage()
			case <-ui.startItem.ClickedCh:
				ui.runAction(trayapp.VMActionStart)
			case <-ui.stopItem.ClickedCh:
				ui.runAction(trayapp.VMActionStop)
			case <-ui.killItem.ClickedCh:
				ui.runAction(trayapp.VMActionKill)
			case <-ui.deleteItem.ClickedCh:
				ui.runAction(trayapp.VMActionDelete)
			case <-ui.quitItem.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()

	for idx, item := range ui.vmItems {
		go func(index int, menuItem *systray.MenuItem) {
			for {
				select {
				case <-ui.ctx.Done():
					return
				case <-menuItem.ClickedCh:
					ui.selectVM(index)
				}
			}
		}(idx, item)
	}

	for idx, item := range ui.imageItems {
		go func(index int, menuItem *systray.MenuItem) {
			for {
				select {
				case <-ui.ctx.Done():
					return
				case <-menuItem.ClickedCh:
					ui.selectImage(index)
				}
			}
		}(idx, item)
	}
}

func (ui *trayUI) refresh() {
	ui.mu.Lock()
	ui.refreshID++
	refreshID := ui.refreshID
	ui.mu.Unlock()

	go func() {
		ctx, cancel := withTimeout(ui.ctx, ui.cfg.Timeout)
		defer cancel()

		status, err := trayapp.CollectStatus(ctx, ui.api, ui.cfg.Endpoint())
		ui.mu.Lock()
		defer ui.mu.Unlock()
		if refreshID != ui.refreshID {
			return
		}
		ui.status = status
		ui.pending = ""
		ui.pendingAt = time.Time{}

		if err != nil {
			ui.statusItem.SetTitle("Status: error")
			ui.runtimeItem.SetTitle(limitTitle("Runtime: " + trayapp.RuntimeSummary(status)))
			ui.updateVMItems(nil)
			ui.updateImageItems(nil)
			ui.updateActionStateLocked()
			return
		}

		ui.statusItem.SetTitle(limitTitle("Status: " + status.Health.Status))
		ui.runtimeItem.SetTitle(limitTitle("Runtime: " + trayapp.RuntimeSummary(status)))
		ui.updateVMItems(status.VMs)
		ui.updateImageItems(status.Images)
		ui.updateActionStateLocked()
	}()
}

func (ui *trayUI) updateVMItems(vms []client.VM) {
	for idx, item := range ui.vmItems {
		if idx >= len(vms) {
			item.Hide()
			continue
		}
		vm := vms[idx]
		item.SetTitle(limitTitle(displayVMRow(vm)))
		item.SetTooltip(displayVMTooltip(vm))
		item.Show()
		item.Enable()
	}
}

func (ui *trayUI) updateImageItems(images []client.Image) {
	for idx, item := range ui.imageItems {
		if idx >= len(images) {
			item.Hide()
			continue
		}
		image := images[idx]
		item.SetTitle(limitTitle(displayImageName(image)))
		item.SetTooltip(image.Digest)
		item.Show()
		item.Enable()
	}
}

func (ui *trayUI) selectVM(index int) {
	ui.mu.Lock()
	defer ui.mu.Unlock()

	if ui.status == nil || index >= len(ui.status.VMs) {
		return
	}
	vm := ui.status.VMs[index]
	ui.selected = vm.ID
	ui.pending = ""
	ui.pendingAt = time.Time{}
	ui.selectedItem.SetTitle(limitTitle("Selected VM: " + displayVMRow(vm)))
	ui.updateActionStateLocked()
}

func (ui *trayUI) selectImage(index int) {
	ui.mu.Lock()
	defer ui.mu.Unlock()

	if ui.status == nil || index >= len(ui.status.Images) {
		return
	}
	image := ui.status.Images[index]
	ui.imageRef = imageReference(image)
	ui.imageItem.SetTitle(limitTitle("Selected Image: " + displayImageName(image)))
	ui.updateActionStateLocked()
}

func (ui *trayUI) launchSelectedImage() {
	ui.mu.Lock()
	imageRef := ui.imageRef
	if imageRef == "" {
		ui.statusItem.SetTitle("Status: select an image first")
		ui.mu.Unlock()
		return
	}
	if !trayapp.VMActionReady(ui.status) {
		ui.statusItem.SetTitle("Status: runtime unavailable")
		ui.runtimeItem.SetTitle(limitTitle("Runtime: " + trayapp.RuntimeSummary(ui.status)))
		ui.updateActionStateLocked()
		ui.mu.Unlock()
		return
	}
	ui.createItem.Disable()
	ui.statusItem.SetTitle("Status: launching VM")
	ui.mu.Unlock()

	go func() {
		ctx, cancel := withTimeout(ui.ctx, ui.cfg.Timeout)
		defer cancel()
		vm, err := trayapp.LaunchVMFromImage(ctx, ui.api, imageRef)
		if err != nil {
			ui.statusItem.SetTitle("Status: launch failed")
			ui.runtimeItem.SetTitle(limitTitle(err.Error()))
			ui.updateActionState()
			return
		}
		ui.mu.Lock()
		if vm != nil && vm.ID != "" {
			ui.selected = vm.ID
			ui.selectedItem.SetTitle(limitTitle("Selected VM: " + displayVMRow(*vm)))
		}
		ui.mu.Unlock()
		ui.statusItem.SetTitle("Status: VM launched")
		ui.refresh()
	}()
}

func (ui *trayUI) runAction(action trayapp.VMAction) {
	ui.mu.Lock()
	selected := ui.selected
	if selected == "" {
		ui.statusItem.SetTitle("Status: select a VM first")
		ui.mu.Unlock()
		return
	}
	if !trayapp.VMActionReady(ui.status) {
		ui.statusItem.SetTitle("Status: runtime unavailable")
		ui.runtimeItem.SetTitle(limitTitle("Runtime: " + trayapp.RuntimeSummary(ui.status)))
		ui.updateActionStateLocked()
		ui.mu.Unlock()
		return
	}
	selectedVM := ui.selectedVMLocked()
	if !trayapp.VMActionAllowed(selectedVM, action) {
		ui.statusItem.SetTitle(limitTitle(fmt.Sprintf("Status: %s unavailable for VM state %s", action, selectedVMState(selectedVM))))
		ui.updateActionStateLocked()
		ui.mu.Unlock()
		return
	}
	if needsConfirmation(action) {
		pendingAt, confirmed := ui.confirmActionLocked(action)
		if !confirmed {
			ui.mu.Unlock()
			ui.expirePending(action, pendingAt)
			return
		}
	}
	ui.pending = ""
	ui.pendingAt = time.Time{}
	ui.updateActionStateLocked()
	ui.mu.Unlock()

	go func() {
		ctx, cancel := withTimeout(ui.ctx, ui.cfg.Timeout)
		defer cancel()
		if _, err := trayapp.ExecuteVMAction(ctx, ui.api, action, selected); err != nil {
			ui.statusItem.SetTitle(limitTitle(fmt.Sprintf("Status: %s failed", action)))
			ui.runtimeItem.SetTitle(limitTitle(err.Error()))
			return
		}
		ui.statusItem.SetTitle(limitTitle(fmt.Sprintf("Status: %s sent", action)))
		ui.refresh()
	}()
}

func needsConfirmation(action trayapp.VMAction) bool {
	return action == trayapp.VMActionKill || action == trayapp.VMActionDelete
}

func (ui *trayUI) confirmActionLocked(action trayapp.VMAction) (time.Time, bool) {
	if ui.pending == action && time.Since(ui.pendingAt) <= confirmationWindow {
		return ui.pendingAt, true
	}
	pendingAt := time.Now()
	ui.pending = action
	ui.pendingAt = pendingAt
	ui.setPendingTitleLocked(action)
	return pendingAt, false
}

func (ui *trayUI) expirePending(action trayapp.VMAction, pendingAt time.Time) {
	go func() {
		timer := time.NewTimer(confirmationWindow)
		defer timer.Stop()

		select {
		case <-ui.ctx.Done():
			return
		case <-timer.C:
		}

		ui.mu.Lock()
		defer ui.mu.Unlock()
		if ui.pending == action && ui.pendingAt.Equal(pendingAt) {
			ui.pending = ""
			ui.pendingAt = time.Time{}
			ui.statusItem.SetTitle("Status: confirmation expired")
			ui.updateActionStateLocked()
		}
	}()
}

func (ui *trayUI) setPendingTitleLocked(action trayapp.VMAction) {
	switch action {
	case trayapp.VMActionKill:
		ui.killItem.SetTitle("Confirm Kill Selected VM")
	case trayapp.VMActionDelete:
		ui.deleteItem.SetTitle("Confirm Delete Selected VM")
	}
}

func (ui *trayUI) updateActionState() {
	ui.mu.Lock()
	defer ui.mu.Unlock()
	ui.updateActionStateLocked()
}

func (ui *trayUI) updateActionStateLocked() {
	ui.createItem.SetTitle("Create and Start VM From Image")
	ui.startItem.SetTitle("Start Selected VM")
	ui.stopItem.SetTitle("Stop Selected VM")
	ui.killItem.SetTitle("Kill Selected VM")
	ui.deleteItem.SetTitle("Delete Selected VM")

	if ui.imageRef == "" || !trayapp.VMActionReady(ui.status) {
		ui.createItem.Disable()
	} else {
		ui.createItem.Enable()
	}

	selectedVM := ui.selectedVMLocked()
	if selectedVM == nil {
		ui.selected = ""
		ui.selectedItem.SetTitle("Selected VM: none")
	}

	if selectedVM == nil || !trayapp.VMActionReady(ui.status) {
		ui.startItem.Disable()
		ui.stopItem.Disable()
		ui.killItem.Disable()
		ui.deleteItem.Disable()
		return
	}
	ui.selectedItem.SetTitle(limitTitle("Selected VM: " + displayVMRow(*selectedVM)))
	setMenuEnabled(ui.startItem, trayapp.VMActionAllowed(selectedVM, trayapp.VMActionStart))
	setMenuEnabled(ui.stopItem, trayapp.VMActionAllowed(selectedVM, trayapp.VMActionStop))
	setMenuEnabled(ui.killItem, trayapp.VMActionAllowed(selectedVM, trayapp.VMActionKill))
	setMenuEnabled(ui.deleteItem, trayapp.VMActionAllowed(selectedVM, trayapp.VMActionDelete))
}

func (ui *trayUI) selectedVMLocked() *client.VM {
	if ui.status == nil || ui.selected == "" {
		return nil
	}
	for idx := range ui.status.VMs {
		if ui.status.VMs[idx].ID == ui.selected {
			return &ui.status.VMs[idx]
		}
	}
	return nil
}

func selectedVMState(vm *client.VM) string {
	if vm == nil || strings.TrimSpace(vm.State) == "" {
		return "unknown"
	}
	return vm.State
}

func setMenuEnabled(item *systray.MenuItem, enabled bool) {
	if enabled {
		item.Enable()
		return
	}
	item.Disable()
}

func displayVMName(vm client.VM) string {
	if vm.Name != "" {
		return vm.Name
	}
	if vm.ID != "" {
		return vm.ID
	}
	return "unnamed"
}

func displayVMRow(vm client.VM) string {
	label := fmt.Sprintf("%s [%s]", displayVMName(vm), vm.State)
	if ports := portSummary(vm.Config.Network.PortForwards); ports != "" {
		label += " " + ports
	}
	return label
}

func displayVMTooltip(vm client.VM) string {
	parts := []string{}
	if vm.ID != "" {
		parts = append(parts, "id="+vm.ID)
	}
	if vm.Image != "" {
		parts = append(parts, "image="+vm.Image)
	}
	if vm.Runtime != nil && vm.Runtime.ExternalID != "" {
		parts = append(parts, "runtime="+vm.Runtime.ExternalID)
	}
	if ports := portSummary(vm.Config.Network.PortForwards); ports != "" {
		parts = append(parts, "ports="+ports)
	}
	if len(parts) == 0 {
		return "VM"
	}
	return strings.Join(parts, " ")
}

func portSummary(portForwards []client.PortForward) string {
	if len(portForwards) == 0 {
		return ""
	}
	parts := make([]string, 0, len(portForwards))
	for _, pf := range portForwards {
		protocol := pf.Protocol
		if protocol == "" {
			protocol = "tcp"
		}
		parts = append(parts, fmt.Sprintf(":%d->%d/%s", pf.HostPort, pf.VMPort, protocol))
	}
	return strings.Join(parts, ",")
}

func displayImageName(image client.Image) string {
	if len(image.Tags) > 0 {
		return strings.Join(image.Tags, ",")
	}
	if image.Digest != "" {
		return image.Digest
	}
	return "untagged image"
}

func imageReference(image client.Image) string {
	if len(image.Tags) > 0 {
		return image.Tags[0]
	}
	return image.Digest
}

func limitTitle(value string) string {
	const maxLen = 72
	value = strings.ReplaceAll(value, "\n", " ")
	if len(value) <= maxLen {
		return value
	}
	return value[:maxLen-3] + "..."
}
