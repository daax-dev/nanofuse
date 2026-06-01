package api

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/daax-dev/nanofuse/internal/applecontainer"
	"github.com/daax-dev/nanofuse/internal/types"
)

const (
	appleContainerSystemStatusTimeout = 2 * time.Second
	appleVirtualizationProbeTimeout   = 2 * time.Second
)

var (
	appleContainerSystemStatusCommand = exec.CommandContext
	appleVirtualizationSupportCommand = exec.CommandContext
)

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
