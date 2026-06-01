//go:build darwin || windows

package main

import (
	"context"
	"fmt"
	"strconv"
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
	pendingVM string
	pendingAt time.Time
	refreshID uint64

	endpointItem     *systray.MenuItem
	statusItem       *systray.MenuItem
	runtimeItem      *systray.MenuItem
	selectedItem     *systray.MenuItem
	imageItem        *systray.MenuItem
	refreshItem      *systray.MenuItem
	promptLaunchItem *systray.MenuItem
	createItem       *systray.MenuItem
	addImageItem     *systray.MenuItem
	vmItems          []vmMenuItems
	imageItems       []*systray.MenuItem
	quitItem         *systray.MenuItem
}

type vmMenuItems struct {
	root   *systray.MenuItem
	detail *systray.MenuItem
	start  *systray.MenuItem
	stop   *systray.MenuItem
	kill   *systray.MenuItem
	delete *systray.MenuItem
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
	ui.promptLaunchItem = systray.AddMenuItem("New MicroVM From Container...", "Enter an OCI image reference and launch it through the API")
	ui.createItem = systray.AddMenuItem("Launch Selected Cached Image", "Create and start a new VM from the selected cached image through the API")
	ui.addImageItem = systray.AddMenuItem("Add Image to List...", "Pull or resolve an OCI image reference through the API")
	systray.AddSeparator()

	vmHeader := systray.AddMenuItem("VMs", "Known VMs")
	vmHeader.Disable()
	for i := 0; i < maxMenuRows; i++ {
		root := systray.AddMenuItem(fmt.Sprintf("VM slot %d", i+1), "VM")
		items := vmMenuItems{
			root:   root,
			detail: root.AddSubMenuItem("Details", "VM details"),
			start:  root.AddSubMenuItem("Start", "Start this VM through the API"),
			stop:   root.AddSubMenuItem("Stop", "Stop this VM through the API"),
			kill:   root.AddSubMenuItem("Kill", "Confirm, then force kill this VM through the API"),
			delete: root.AddSubMenuItem("Delete", "Confirm, then delete this VM through the API"),
		}
		items.detail.Disable()
		hideVMMenu(items)
		ui.vmItems = append(ui.vmItems, items)
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
	go ui.listenMainEvents()

	for idx, items := range ui.vmItems {
		go ui.listenVMRow(idx, items)
	}

	for idx, item := range ui.imageItems {
		go ui.listenImageRow(idx, item)
	}
}

func (ui *trayUI) listenMainEvents() {
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
		case <-ui.promptLaunchItem.ClickedCh:
			ui.promptAndLaunchImage()
		case <-ui.createItem.ClickedCh:
			ui.launchSelectedImage()
		case <-ui.addImageItem.ClickedCh:
			ui.promptAndAddImage()
		case <-ui.quitItem.ClickedCh:
			systray.Quit()
			return
		}
	}
}

func (ui *trayUI) listenVMRow(index int, menuItems vmMenuItems) {
	for {
		select {
		case <-ui.ctx.Done():
			return
		case <-menuItems.root.ClickedCh:
			ui.selectVM(index)
		case <-menuItems.start.ClickedCh:
			ui.runActionForIndex(index, trayapp.VMActionStart)
		case <-menuItems.stop.ClickedCh:
			ui.runActionForIndex(index, trayapp.VMActionStop)
		case <-menuItems.kill.ClickedCh:
			ui.runActionForIndex(index, trayapp.VMActionKill)
		case <-menuItems.delete.ClickedCh:
			ui.runActionForIndex(index, trayapp.VMActionDelete)
		}
	}
}

func (ui *trayUI) listenImageRow(index int, menuItem *systray.MenuItem) {
	for {
		select {
		case <-ui.ctx.Done():
			return
		case <-menuItem.ClickedCh:
			ui.selectImage(index)
		}
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
		ui.pendingVM = ""
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
	ready := trayapp.VMActionReady(ui.status)
	for idx, items := range ui.vmItems {
		if idx >= len(vms) {
			hideVMMenu(items)
			continue
		}
		vm := vms[idx]
		items.root.SetTitle(limitTitle(displayVMRow(vm)))
		items.root.SetTooltip(displayVMTooltip(vm))
		items.detail.SetTitle(limitTitle("Details: " + displayVMTooltip(vm)))
		items.detail.SetTooltip(displayVMTooltip(vm))
		items.start.SetTitle("Start")
		items.stop.SetTitle("Stop")
		items.kill.SetTitle(actionTitle(ui.pending, ui.pendingVM, vm.ID, trayapp.VMActionKill, "Kill", "Confirm Kill"))
		items.delete.SetTitle(actionTitle(ui.pending, ui.pendingVM, vm.ID, trayapp.VMActionDelete, "Delete", "Confirm Delete"))
		showVMMenu(items)
		items.root.Enable()
		setMenuEnabled(items.start, ready && trayapp.VMActionAllowed(&vm, trayapp.VMActionStart))
		setMenuEnabled(items.stop, ready && trayapp.VMActionAllowed(&vm, trayapp.VMActionStop))
		setMenuEnabled(items.kill, ready && trayapp.VMActionAllowed(&vm, trayapp.VMActionKill))
		setMenuEnabled(items.delete, ready && trayapp.VMActionAllowed(&vm, trayapp.VMActionDelete))
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
	ui.pendingVM = ""
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
	ui.pending = ""
	ui.pendingVM = ""
	ui.pendingAt = time.Time{}
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
	ui.mu.Unlock()
	ui.launchImageRef(imageRef)
}

func (ui *trayUI) promptAndLaunchImage() {
	ui.mu.Lock()
	if !trayapp.VMActionReady(ui.status) {
		ui.statusItem.SetTitle("Status: runtime unavailable")
		ui.runtimeItem.SetTitle(limitTitle("Runtime: " + trayapp.RuntimeSummary(ui.status)))
		ui.updateActionStateLocked()
		ui.mu.Unlock()
		return
	}
	ui.promptLaunchItem.Disable()
	ui.statusItem.SetTitle("Status: waiting for image reference")
	ui.mu.Unlock()

	go func() {
		imageRef, ok, err := promptImageReference(ui.ctx)
		if err != nil {
			ui.statusItem.SetTitle("Status: image prompt failed")
			ui.runtimeItem.SetTitle(limitTitle(err.Error()))
			ui.updateActionState()
			return
		}
		if !ok {
			ui.statusItem.SetTitle("Status: launch canceled")
			ui.updateActionState()
			return
		}
		ui.launchImageRef(imageRef)
	}()
}

func (ui *trayUI) promptAndAddImage() {
	ui.mu.Lock()
	if !trayapp.VMActionReady(ui.status) {
		ui.statusItem.SetTitle("Status: runtime unavailable")
		ui.runtimeItem.SetTitle(limitTitle("Runtime: " + trayapp.RuntimeSummary(ui.status)))
		ui.updateActionStateLocked()
		ui.mu.Unlock()
		return
	}
	ui.addImageItem.Disable()
	ui.statusItem.SetTitle("Status: waiting for image reference")
	ui.mu.Unlock()

	go func() {
		imageRef, ok, err := promptImageReference(ui.ctx)
		if err != nil {
			ui.statusItem.SetTitle("Status: image prompt failed")
			ui.runtimeItem.SetTitle(limitTitle(err.Error()))
			ui.updateActionState()
			return
		}
		if !ok {
			ui.statusItem.SetTitle("Status: add image canceled")
			ui.updateActionState()
			return
		}
		ctx, cancel := withTimeout(ui.ctx, ui.cfg.Timeout)
		defer cancel()
		job, err := trayapp.AddImage(ctx, ui.api, imageRef)
		if err != nil {
			ui.statusItem.SetTitle("Status: add image failed")
			ui.runtimeItem.SetTitle(limitTitle(err.Error()))
			ui.updateActionState()
			return
		}
		ui.statusItem.SetTitle(limitTitle("Status: image pull started " + job.ID))
		ui.refresh()
	}()
}

func (ui *trayUI) launchImageRef(imageRef string) {
	ui.mu.Lock()
	if !trayapp.VMActionReady(ui.status) {
		ui.statusItem.SetTitle("Status: runtime unavailable")
		ui.runtimeItem.SetTitle(limitTitle("Runtime: " + trayapp.RuntimeSummary(ui.status)))
		ui.updateActionStateLocked()
		ui.mu.Unlock()
		return
	}
	ui.createItem.Disable()
	ui.promptLaunchItem.Disable()
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

func (ui *trayUI) runActionForIndex(index int, action trayapp.VMAction) {
	ui.mu.Lock()
	if ui.status == nil || index >= len(ui.status.VMs) {
		ui.statusItem.SetTitle("Status: VM row unavailable")
		ui.mu.Unlock()
		return
	}
	vm := ui.status.VMs[index]
	ui.selected = vm.ID
	ui.selectedItem.SetTitle(limitTitle("Selected VM: " + displayVMRow(vm)))
	if !trayapp.VMActionReady(ui.status) {
		ui.statusItem.SetTitle("Status: runtime unavailable")
		ui.runtimeItem.SetTitle(limitTitle("Runtime: " + trayapp.RuntimeSummary(ui.status)))
		ui.updateActionStateLocked()
		ui.mu.Unlock()
		return
	}
	if !trayapp.VMActionAllowed(&vm, action) {
		ui.statusItem.SetTitle(limitTitle(fmt.Sprintf("Status: %s unavailable for VM state %s", action, selectedVMState(&vm))))
		ui.updateActionStateLocked()
		ui.mu.Unlock()
		return
	}
	if needsConfirmation(action) {
		pendingAt, confirmed := ui.confirmActionLocked(action, vm.ID)
		if !confirmed {
			ui.updateActionStateLocked()
			ui.mu.Unlock()
			ui.expirePending(action, vm.ID, pendingAt)
			return
		}
	}
	ui.pending = ""
	ui.pendingVM = ""
	ui.pendingAt = time.Time{}
	ui.updateActionStateLocked()
	ui.mu.Unlock()

	go func() {
		ctx, cancel := withTimeout(ui.ctx, ui.cfg.Timeout)
		defer cancel()
		if _, err := trayapp.ExecuteVMAction(ctx, ui.api, action, vm.ID); err != nil {
			ui.statusItem.SetTitle(limitTitle(fmt.Sprintf("Status: %s failed", action)))
			ui.runtimeItem.SetTitle(limitTitle(err.Error()))
			return
		}
		ui.statusItem.SetTitle(limitTitle(fmt.Sprintf("Status: %s sent for %s", action, displayVMName(vm))))
		ui.refresh()
	}()
}

func needsConfirmation(action trayapp.VMAction) bool {
	return action == trayapp.VMActionKill || action == trayapp.VMActionDelete
}

func (ui *trayUI) confirmActionLocked(action trayapp.VMAction, vmID string) (time.Time, bool) {
	if ui.pending == action && ui.pendingVM == vmID && time.Since(ui.pendingAt) <= confirmationWindow {
		return ui.pendingAt, true
	}
	pendingAt := time.Now()
	ui.pending = action
	ui.pendingVM = vmID
	ui.pendingAt = pendingAt
	return pendingAt, false
}

func (ui *trayUI) expirePending(action trayapp.VMAction, vmID string, pendingAt time.Time) {
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
		if ui.pending == action && ui.pendingVM == vmID && ui.pendingAt.Equal(pendingAt) {
			ui.pending = ""
			ui.pendingVM = ""
			ui.pendingAt = time.Time{}
			ui.statusItem.SetTitle("Status: confirmation expired")
			ui.updateActionStateLocked()
		}
	}()
}

func (ui *trayUI) updateActionState() {
	ui.mu.Lock()
	defer ui.mu.Unlock()
	ui.updateActionStateLocked()
}

func (ui *trayUI) updateActionStateLocked() {
	ui.promptLaunchItem.SetTitle("New MicroVM From Container...")
	ui.createItem.SetTitle("Launch Selected Cached Image")
	ui.addImageItem.SetTitle("Add Image to List...")

	ready := trayapp.VMActionReady(ui.status)
	setMenuEnabled(ui.promptLaunchItem, ready)
	setMenuEnabled(ui.addImageItem, ready)

	if ui.imageRef == "" || !ready {
		ui.createItem.Disable()
	} else {
		ui.createItem.Enable()
	}

	selectedVM := ui.selectedVMLocked()
	if selectedVM == nil {
		ui.selected = ""
		ui.selectedItem.SetTitle("Selected VM: none")
	}

	if selectedVM == nil || !ready {
		if ui.status != nil {
			ui.updateVMItems(ui.status.VMs)
		}
		return
	}
	ui.selectedItem.SetTitle(limitTitle("Selected VM: " + displayVMRow(*selectedVM)))
	if ui.status != nil {
		ui.updateVMItems(ui.status.VMs)
	}
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

func hideVMMenu(items vmMenuItems) {
	for _, item := range []*systray.MenuItem{items.root, items.detail, items.start, items.stop, items.kill, items.delete} {
		if item != nil {
			item.Hide()
		}
	}
}

func showVMMenu(items vmMenuItems) {
	for _, item := range []*systray.MenuItem{items.root, items.detail, items.start, items.stop, items.kill, items.delete} {
		if item != nil {
			item.Show()
		}
	}
}

func actionTitle(pending trayapp.VMAction, pendingVM, vmID string, action trayapp.VMAction, normal, confirm string) string {
	if pending == action && pendingVM == vmID {
		return confirm
	}
	return normal
}

func displayVMName(vm client.VM) string {
	if vm.Name != "" {
		return vm.Name
	}
	if vm.ID != "" {
		return "vm-" + shortID(vm.ID)
	}
	return "unnamed"
}

func displayVMRow(vm client.VM) string {
	label := fmt.Sprintf("%s [%s]", displayVMName(vm), vm.State)
	if image := displayImageRefShort(vm.Image); image != "" {
		label += " " + image
	}
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
		hostPort := strconv.Itoa(pf.HostPort)
		if pf.HostPort == 0 {
			hostPort = "auto"
		}
		parts = append(parts, fmt.Sprintf(":%s->%d/%s", hostPort, pf.VMPort, protocol))
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

func displayImageRefShort(imageRef string) string {
	imageRef = strings.TrimSpace(imageRef)
	if imageRef == "" {
		return ""
	}
	if digestIndex := strings.LastIndex(imageRef, "@"); digestIndex >= 0 {
		imageRef = imageRef[:digestIndex]
	}
	if slashIndex := strings.LastIndex(imageRef, "/"); slashIndex >= 0 {
		imageRef = imageRef[slashIndex+1:]
	}
	return imageRef
}

func shortID(id string) string {
	id = strings.TrimSpace(id)
	if len(id) <= 12 {
		return id
	}
	return id[:12]
}

func limitTitle(value string) string {
	const maxLen = 72
	value = strings.ReplaceAll(value, "\n", " ")
	if len(value) <= maxLen {
		return value
	}
	return value[:maxLen-3] + "..."
}
