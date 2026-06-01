package api

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/daax-dev/nanofuse/internal/applecontainer"
	"github.com/daax-dev/nanofuse/internal/types"
	"github.com/daax-dev/nanofuse/internal/vmm"
)

const (
	appleContainerSystemStatusTimeout = 2 * time.Second
	appleVirtualizationProbeTimeout   = 2 * time.Second
)

var (
	appleContainerSystemStatusCommand = exec.CommandContext
	appleVirtualizationSupportCommand = exec.CommandContext
)

// handleRoot returns a browser-readable daemon status page (GET /).
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	capabilities := s.capabilitiesResponse()
	vms, vmErr := s.rootVMs()
	images, imageErr := s.rootImages()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, renderRootPage(capabilities, vms, vmErr, images, imageErr))
}

func (s *Server) rootVMs() ([]*types.VM, error) {
	if s.db == nil {
		return nil, nil
	}
	return s.db.ListVMs("")
}

func (s *Server) rootImages() ([]*types.Image, error) {
	var images []*types.Image
	if s.db != nil {
		dbImages, err := s.db.ListImages()
		if err != nil {
			return nil, err
		}
		images = dbImages
	}

	if provider, ok := s.runtimeManager.(vmm.ImageProvider); ok {
		runtimeImages, err := provider.ListImages()
		if err != nil {
			return images, err
		}
		images = mergeRuntimeImages(images, runtimeImages)
	}

	return images, nil
}

func renderRootPage(
	capabilities types.CapabilitiesResponse,
	vms []*types.VM,
	vmErr error,
	images []*types.Image,
	imageErr error,
) string {
	var b strings.Builder
	b.WriteString(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Nanofuse</title>
<style>
:root { color-scheme: light dark; font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; }
body { margin: 0; padding: 32px; background: Canvas; color: CanvasText; }
main { max-width: 1120px; margin: 0 auto; }
h1 { margin: 0 0 8px; font-size: 32px; }
h2 { margin-top: 32px; font-size: 20px; }
a { color: LinkText; }
code, pre { font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; }
pre { overflow: auto; padding: 16px; border: 1px solid color-mix(in srgb, CanvasText 20%, Canvas); border-radius: 6px; }
table { width: 100%; border-collapse: collapse; margin-top: 12px; }
th, td { text-align: left; padding: 10px 12px; border-bottom: 1px solid color-mix(in srgb, CanvasText 16%, Canvas); vertical-align: top; }
th { font-size: 12px; text-transform: uppercase; letter-spacing: 0; }
.muted { color: color-mix(in srgb, CanvasText 62%, Canvas); }
.grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 12px; margin-top: 20px; }
.panel { border: 1px solid color-mix(in srgb, CanvasText 18%, Canvas); border-radius: 6px; padding: 14px; }
.label { display: block; font-size: 12px; text-transform: uppercase; letter-spacing: 0; color: color-mix(in srgb, CanvasText 62%, Canvas); }
.value { display: block; margin-top: 4px; overflow-wrap: anywhere; }
</style>
</head>
<body>
<main>
<h1>Nanofuse</h1>
<p class="muted">Local microVM API daemon status and launch inventory.</p>
<p><a href="/health">Health</a> | <a href="/capabilities">Capabilities</a> | <a href="/vms">VMs JSON</a> | <a href="/images">Images JSON</a></p>
`)

	writeRootRuntime(&b, capabilities)
	writeRootVMs(&b, vms, vmErr)
	writeRootImages(&b, images, imageErr)
	writeRootCommands(&b)

	b.WriteString(`</main>
</body>
</html>
`)
	return b.String()
}

func writeRootRuntime(b *strings.Builder, capabilities types.CapabilitiesResponse) {
	fmt.Fprintf(b, `<h2>Runtime</h2>
<div class="grid">
<div class="panel"><span class="label">Driver</span><span class="value">%s</span></div>
<div class="panel"><span class="label">Native Runtime</span><span class="value">%t</span></div>
<div class="panel"><span class="label">Apple Container</span><span class="value">installed=%t running=%t virtualization=%t</span></div>
<div class="panel"><span class="label">Firecracker</span><span class="value">installed=%t kvm=%t</span></div>
</div>
<p class="muted">%s</p>
`,
		html.EscapeString(capabilities.Runtime.Driver),
		capabilities.Runtime.NativeRuntime,
		capabilities.Runtime.AppleContainerAvailable,
		capabilities.Runtime.AppleContainerRunning,
		capabilities.Runtime.VirtualizationFrameworkSupported,
		capabilities.Runtime.FirecrackerAvailable,
		capabilities.Host.KVMReadWrite,
		html.EscapeString(capabilities.Runtime.Message),
	)
}

func writeRootVMs(b *strings.Builder, vms []*types.VM, vmErr error) {
	b.WriteString(`<h2>VMs</h2>
`)
	if vmErr != nil {
		fmt.Fprintf(b, `<p>VM inventory unavailable: %s</p>
`, html.EscapeString(vmErr.Error()))
		return
	}
	if len(vms) == 0 {
		b.WriteString(`<p>No VMs.</p>
`)
		return
	}

	b.WriteString(`<table>
<thead><tr><th>Name</th><th>ID</th><th>State</th><th>Image</th><th>Ports</th></tr></thead>
<tbody>
`)
	for _, vm := range vms {
		fmt.Fprintf(b, "<tr><td>%s</td><td><code>%s</code></td><td>%s</td><td>%s</td><td>%s</td></tr>\n",
			html.EscapeString(rootVMName(vm)),
			html.EscapeString(vm.ID),
			html.EscapeString(string(vm.State)),
			html.EscapeString(vm.Image),
			html.EscapeString(rootPortSummary(vm.Config.Network.PortForwards)),
		)
	}
	b.WriteString(`</tbody>
</table>
`)
}

func writeRootImages(b *strings.Builder, images []*types.Image, imageErr error) {
	b.WriteString(`<h2>Images</h2>
`)
	if imageErr != nil {
		fmt.Fprintf(b, `<p>Runtime image inventory partially unavailable: %s</p>
`, html.EscapeString(imageErr.Error()))
	}
	if len(images) == 0 {
		b.WriteString(`<p>No cached or runtime images.</p>
`)
		return
	}

	b.WriteString(`<table>
<thead><tr><th>Reference</th><th>Digest</th><th>Architecture</th></tr></thead>
<tbody>
`)
	for _, image := range images {
		fmt.Fprintf(b, "<tr><td>%s</td><td><code>%s</code></td><td>%s</td></tr>\n",
			html.EscapeString(rootImageName(image)),
			html.EscapeString(image.Digest),
			html.EscapeString(image.Architecture),
		)
	}
	b.WriteString(`</tbody>
</table>
`)
}

func writeRootCommands(b *strings.Builder) {
	b.WriteString(`<h2>CLI</h2>
<pre>export NANOFUSE_API_URL=http://127.0.0.1:18080
bin/nanofuse vm list
bin/nanofuse vm ports
bin/nanofuse vm exec &lt;vm-id&gt; -- sh -lc 'cat /etc/os-release'</pre>
`)
}

func rootVMName(vm *types.VM) string {
	if vm == nil {
		return ""
	}
	if strings.TrimSpace(vm.Name) != "" {
		return vm.Name
	}
	return vm.ID
}

func rootImageName(image *types.Image) string {
	if image == nil {
		return ""
	}
	if len(image.Tags) > 0 {
		return strings.Join(image.Tags, ", ")
	}
	if image.Digest != "" {
		return image.Digest
	}
	return "untagged image"
}

func rootPortSummary(portForwards []types.PortForward) string {
	if len(portForwards) == 0 {
		return "none exposed"
	}
	parts := make([]string, 0, len(portForwards))
	for _, pf := range portForwards {
		protocol := pf.Protocol
		if protocol == "" {
			protocol = "tcp"
		}
		parts = append(parts, fmt.Sprintf("127.0.0.1:%d -> vm:%d/%s", pf.HostPort, pf.VMPort, protocol))
	}
	return strings.Join(parts, ", ")
}

// handleHealth handles the health check endpoint (GET /health)
// Method validation is handled by the router using Go 1.22+ patterns
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(s.startTime).Seconds()

	response := types.HealthResponse{
		Status:        "healthy",
		Version:       "0.1.0",
		UptimeSeconds: int64(uptime),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Printf("ERROR: Failed to encode health response: %v", err)
	}
}

// handleCapabilities handles the capability endpoint (GET /capabilities).
func (s *Server) handleCapabilities(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.capabilitiesResponse())
}

func (s *Server) capabilitiesResponse() types.CapabilitiesResponse {
	socketPath := ""
	tcpBind := ""
	firecrackerBinary := ""
	driver := selectedRuntimeDriver(s.config)
	appleContainerBinary := ""
	appleContainerAutoStart := false
	if s.config != nil {
		socketPath = s.config.API.Socket
		tcpBind = s.config.API.TCPBind
		firecrackerBinary = s.config.Firecracker.BinaryPath
		appleContainerBinary = s.config.Runtime.AppleContainer.BinaryPath
		appleContainerAutoStart = s.config.Runtime.AppleContainer.AutoStart
	}

	kvmExists, kvmReadWrite, kvmErr := inspectKVMDevice("/dev/kvm")
	firecrackerAvailable := executableAvailable(firecrackerBinary)
	appleContainerAvailable := executableAvailable(appleContainerBinary)
	appleContainerRunning := false
	if driver == applecontainer.DriverName && appleContainerAvailable {
		appleContainerRunning = appleContainerSystemRunning(appleContainerBinary)
	}
	virtualizationSupported := appleVirtualizationFrameworkSupported(runtime.GOOS)
	appleContainerReady := driver == applecontainer.DriverName &&
		appleContainerNativeReady(runtime.GOOS, appleContainerAvailable, virtualizationSupported, appleContainerRunning, appleContainerAutoStart)
	nativeRuntime := (driver == "firecracker" && runtime.GOOS == "linux" && kvmReadWrite && firecrackerAvailable) ||
		appleContainerReady

	message := "Linux KVM and Firecracker are available for local microVM execution"
	if driver == applecontainer.DriverName && appleContainerReady {
		message = "Apple container and Virtualization.framework are available for local macOS Linux microVM execution"
		if !appleContainerRunning {
			message = "Apple container is installed for local macOS Linux microVM execution; service will be started on demand"
		}
	}
	if !nativeRuntime {
		message = "Nanofuse microVM execution requires Linux/KVM with Firecracker or macOS with Apple container and Virtualization.framework"
		if driver == applecontainer.DriverName && runtime.GOOS == "darwin" && appleContainerAvailable && !appleContainerRunning && !appleContainerAutoStart {
			message = "Apple container is installed but services are stopped and runtime.apple_container.auto_start is false"
		}
	}

	return types.CapabilitiesResponse{
		Status:  "ok",
		Version: "0.1.0",
		Host: types.HostCapabilities{
			OS:           runtime.GOOS,
			Arch:         runtime.GOARCH,
			KVMDevice:    "/dev/kvm",
			KVMExists:    kvmExists,
			KVMReadWrite: kvmReadWrite,
			KVMError:     kvmErr,
		},
		Runtime: types.RuntimeCapabilities{
			NativeRuntime:                    nativeRuntime,
			Driver:                           driver,
			FirecrackerBinary:                firecrackerBinary,
			FirecrackerAvailable:             firecrackerAvailable,
			AppleContainerBinary:             appleContainerBinary,
			AppleContainerAvailable:          appleContainerAvailable,
			AppleContainerRunning:            appleContainerRunning,
			VirtualizationFrameworkSupported: virtualizationSupported,
			RootRequired:                     driver == "firecracker",
			NetworkSetupRequired:             driver == "firecracker",
			Message:                          message,
		},
		API: types.APITransportCapabilities{
			UnixSocket: socketPath,
			TCPBind:    tcpBind,
		},
	}
}

func appleContainerNativeReady(goos string, available, virtualizationSupported, running, autoStart bool) bool {
	return goos == "darwin" && available && virtualizationSupported && (running || autoStart)
}

func appleVirtualizationFrameworkSupported(goos string) bool {
	if goos != "darwin" {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), appleVirtualizationProbeTimeout)
	defer cancel()

	cmd := appleVirtualizationSupportCommand(ctx, "sysctl", "-n", "kern.hv_support")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "1"
}

func appleContainerSystemRunning(path string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), appleContainerSystemStatusTimeout)
	defer cancel()

	cmd := appleContainerSystemStatusCommand(ctx, path, "system", "status")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), "apiserver is running")
}

func inspectKVMDevice(path string) (bool, bool, string) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, false, "not found"
		}
		return false, false, err.Error()
	}

	file, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return true, false, err.Error()
	}
	_ = file.Close()

	return true, true, ""
}

func executableAvailable(path string) bool {
	if path == "" {
		return false
	}

	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}

	return info.Mode()&0111 != 0
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	// Encoding errors are logged but we can't return an error after WriteHeader
	_ = json.NewEncoder(w).Encode(data)
}

// readJSON reads a JSON request body
func readJSON(r *http.Request, v interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(v)
}
