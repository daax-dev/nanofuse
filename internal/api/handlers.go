package api

import (
	"encoding/json"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/daax-dev/nanofuse/internal/types"
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
	if s.config != nil {
		socketPath = s.config.API.Socket
		tcpBind = s.config.API.TCPBind
		firecrackerBinary = s.config.Firecracker.BinaryPath
	}

	kvmExists, kvmReadWrite, kvmErr := inspectKVMDevice("/dev/kvm")
	firecrackerAvailable := executableAvailable(firecrackerBinary)
	nativeRuntime := runtime.GOOS == "linux" && kvmReadWrite && firecrackerAvailable

	message := "Linux KVM and Firecracker are available for local microVM execution"
	if !nativeRuntime {
		message = "Nanofuse microVM execution requires a Linux host with read/write /dev/kvm and a Firecracker binary; use this daemon as the runtime host and connect to it over the API from macOS or Windows"
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
			NativeRuntime:        nativeRuntime,
			FirecrackerBinary:    firecrackerBinary,
			FirecrackerAvailable: firecrackerAvailable,
			RootRequired:         true,
			NetworkSetupRequired: true,
			Message:              message,
		},
		API: types.APITransportCapabilities{
			UnixSocket: socketPath,
			TCPBind:    tcpBind,
		},
	}
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
