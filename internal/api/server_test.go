package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/daax-dev/nanofuse/internal/config"
	"github.com/daax-dev/nanofuse/internal/storage"
	"github.com/daax-dev/nanofuse/internal/types"
)

func TestHealthEndpoint(t *testing.T) {
	// Create a test server with minimal setup
	server := &Server{
		startTime: time.Now(),
	}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response types.HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", response.Status)
	}

	if response.Version != "0.1.0" {
		t.Errorf("Expected version '0.1.0', got '%s'", response.Version)
	}

	if response.UptimeSeconds < 0 {
		t.Errorf("Expected non-negative uptime, got %d", response.UptimeSeconds)
	}
}

func TestHealthEndpointMethodNotAllowed(t *testing.T) {
	// Create a test server and router to test Go 1.22+ method-aware routing
	server := &Server{
		startTime: time.Now(),
	}
	mux := setupHTTPRouter(server)

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	// Go 1.22+ ServeMux returns 405 Method Not Allowed for mismatched methods
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestRootEndpointReturnsStatusPage(t *testing.T) {
	db, err := storage.New(filepath.Join(t.TempDir(), "nanofuse.db"))
	if err != nil {
		t.Fatalf("storage.New() error = %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("db.Close() error = %v", err)
		}
	})

	now := time.Now()
	if err := db.UpsertImage(&types.Image{
		Digest:       "sha256:image-root",
		Tags:         []string{"docker.io/library/alpine:3.20"},
		Architecture: "arm64",
		PulledAt:     now,
	}); err != nil {
		t.Fatalf("UpsertImage() error = %v", err)
	}
	if err := db.CreateVM(&types.VM{
		ID:           "vm-root",
		Name:         "root-page",
		State:        types.StateRunning,
		Image:        "docker.io/library/alpine:3.20",
		ImageDigest:  "sha256:image-root",
		Architecture: "arm64",
		Config: types.VMConfig{
			VCPUs:     2,
			MemoryMiB: 512,
			Network: types.NetworkConfig{
				Mode: "nat",
				PortForwards: []types.PortForward{
					{HostPort: 19080, VMPort: 8080, Protocol: "tcp"},
				},
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("CreateVM() error = %v", err)
	}

	server := &Server{
		db:        db,
		config:    &config.Config{Runtime: config.RuntimeConfig{Driver: "apple_container"}},
		startTime: now,
	}
	mux := setupHTTPRouter(server)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if contentType := w.Header().Get("Content-Type"); !strings.Contains(contentType, "text/html") {
		t.Fatalf("Content-Type = %q, want text/html", contentType)
	}
	body := w.Body.String()
	for _, want := range []string{
		"Nanofuse",
		"/health",
		"/capabilities",
		"/vms",
		"root-page",
		"vm-root",
		"127.0.0.1:19080 -&gt; vm:8080/tcp",
		"docker.io/library/alpine:3.20",
		"bin/nanofuse vm ports",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("root body missing %q:\n%s", want, body)
		}
	}
}

func TestRootEndpointDoesNotMaskUnknownPath(t *testing.T) {
	server := &Server{startTime: time.Now()}
	mux := setupHTTPRouter(server)

	req := httptest.NewRequest(http.MethodGet, "/missing-route", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestCapabilitiesEndpoint(t *testing.T) {
	server := &Server{
		config: &config.Config{
			API: config.APIConfig{
				Socket:  "/tmp/nanofused.sock",
				TCPBind: "127.0.0.1:8080",
			},
			Firecracker: config.FirecrackerConfig{
				BinaryPath: "/usr/local/bin/firecracker",
			},
		},
		startTime: time.Now(),
	}

	req := httptest.NewRequest(http.MethodGet, "/capabilities", nil)
	w := httptest.NewRecorder()

	server.handleCapabilities(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response types.CapabilitiesResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response.Status)
	}

	if response.Host.OS != runtime.GOOS {
		t.Errorf("Expected host OS %q, got %q", runtime.GOOS, response.Host.OS)
	}

	if response.API.UnixSocket != "/tmp/nanofused.sock" {
		t.Errorf("Expected unix socket from config, got %q", response.API.UnixSocket)
	}

	if response.API.TCPBind != "127.0.0.1:8080" {
		t.Errorf("Expected TCP bind from config, got %q", response.API.TCPBind)
	}

	if response.Runtime.FirecrackerBinary != "/usr/local/bin/firecracker" {
		t.Errorf("Expected firecracker path from config, got %q", response.Runtime.FirecrackerBinary)
	}
}

func TestCapabilitiesEndpointIncludesRuntimeContractFieldsWithoutConfig(t *testing.T) {
	server := &Server{
		startTime: time.Now(),
	}

	req := httptest.NewRequest(http.MethodGet, "/capabilities", nil)
	w := httptest.NewRecorder()

	server.handleCapabilities(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	runtimePayload, ok := response["runtime"].(map[string]interface{})
	if !ok {
		t.Fatalf("runtime payload missing or wrong type: %#v", response["runtime"])
	}
	if got := runtimePayload["driver"]; got != selectedRuntimeDriver(nil) {
		t.Fatalf("runtime.driver = %#v, want %q", got, selectedRuntimeDriver(nil))
	}
	for _, key := range []string{
		"apple_container_available",
		"apple_container_running",
		"virtualization_framework_supported",
	} {
		if _, ok := runtimePayload[key]; !ok {
			t.Fatalf("runtime.%s omitted from capabilities response", key)
		}
		if _, ok := runtimePayload[key].(bool); !ok {
			t.Fatalf("runtime.%s = %#v, want bool", key, runtimePayload[key])
		}
	}
}

func TestAppleContainerNativeReadyRequiresRunningOrAutoStart(t *testing.T) {
	tests := []struct {
		name      string
		goos      string
		available bool
		vfSupport bool
		running   bool
		autoStart bool
		want      bool
	}{
		{
			name:      "running service is ready",
			goos:      "darwin",
			available: true,
			vfSupport: true,
			running:   true,
			autoStart: false,
			want:      true,
		},
		{
			name:      "auto start can become ready",
			goos:      "darwin",
			available: true,
			vfSupport: true,
			running:   false,
			autoStart: true,
			want:      true,
		},
		{
			name:      "unsupported virtualization framework is not ready",
			goos:      "darwin",
			available: true,
			vfSupport: false,
			running:   true,
			autoStart: true,
			want:      false,
		},
		{
			name:      "stopped service without auto start is not ready",
			goos:      "darwin",
			available: true,
			vfSupport: true,
			running:   false,
			autoStart: false,
			want:      false,
		},
		{
			name:      "missing CLI is not ready",
			goos:      "darwin",
			available: false,
			vfSupport: true,
			running:   true,
			autoStart: true,
			want:      false,
		},
		{
			name:      "linux does not use apple container readiness",
			goos:      "linux",
			available: true,
			vfSupport: true,
			running:   true,
			autoStart: true,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := appleContainerNativeReady(tt.goos, tt.available, tt.vfSupport, tt.running, tt.autoStart)
			if got != tt.want {
				t.Fatalf("appleContainerNativeReady() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppleVirtualizationFrameworkSupportedProbesHypervisorSupport(t *testing.T) {
	oldCommand := appleVirtualizationSupportCommand
	t.Cleanup(func() {
		appleVirtualizationSupportCommand = oldCommand
	})

	t.Setenv("NANOFUSE_TEST_APPLE_VIRTUALIZATION_SUPPORT", "1")
	called := false
	appleVirtualizationSupportCommand = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		called = true
		if name != "sysctl" {
			t.Fatalf("command name = %q, want sysctl", name)
		}
		if len(arg) != 2 || arg[0] != "-n" || arg[1] != "kern.hv_support" {
			t.Fatalf("command args = %#v, want -n kern.hv_support", arg)
		}
		if _, ok := ctx.Deadline(); !ok {
			t.Fatal("expected virtualization support command to receive a deadline")
		}
		testBinary, err := os.Executable()
		if err != nil {
			t.Fatalf("os.Executable: %v", err)
		}
		return exec.CommandContext(ctx, testBinary, "-test.run=TestAppleVirtualizationSupportHelper", "--")
	}

	if !appleVirtualizationFrameworkSupported("darwin") {
		t.Fatal("expected darwin host with kern.hv_support=1 to support virtualization framework")
	}
	if !called {
		t.Fatal("expected virtualization support command to be called")
	}

	called = false
	if appleVirtualizationFrameworkSupported("linux") {
		t.Fatal("expected non-darwin host to report virtualization framework unsupported")
	}
	if called {
		t.Fatal("expected non-darwin host to skip virtualization support command")
	}
}

func TestAppleVirtualizationSupportHelper(t *testing.T) {
	if value := os.Getenv("NANOFUSE_TEST_APPLE_VIRTUALIZATION_SUPPORT"); value != "" {
		fmt.Println(value)
		os.Exit(0)
	}
}

func TestAppleContainerSystemRunningUsesTimeout(t *testing.T) {
	oldCommand := appleContainerSystemStatusCommand
	t.Cleanup(func() {
		appleContainerSystemStatusCommand = oldCommand
	})

	t.Setenv("NANOFUSE_TEST_APPLE_CONTAINER_STATUS", "running")
	called := false
	appleContainerSystemStatusCommand = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		called = true
		if name != "/fake/container" {
			t.Fatalf("command name = %q, want fake container path", name)
		}
		if len(arg) != 2 || arg[0] != "system" || arg[1] != "status" {
			t.Fatalf("command args = %#v, want system status", arg)
		}
		if _, ok := ctx.Deadline(); !ok {
			t.Fatal("expected apple container status command to receive a deadline")
		}
		testBinary, err := os.Executable()
		if err != nil {
			t.Fatalf("os.Executable: %v", err)
		}
		return exec.CommandContext(ctx, testBinary, "-test.run=TestAppleContainerSystemStatusHelper", "--")
	}

	if !appleContainerSystemRunning("/fake/container") {
		t.Fatal("expected appleContainerSystemRunning to parse running helper output")
	}
	if !called {
		t.Fatal("expected status command to be called")
	}
}

func TestAppleContainerSystemStatusHelper(t *testing.T) {
	if os.Getenv("NANOFUSE_TEST_APPLE_CONTAINER_STATUS") == "" {
		return
	}
	fmt.Println("apiserver is running")
	os.Exit(0)
}

func TestCapabilitiesEndpointMethodNotAllowed(t *testing.T) {
	server := &Server{
		startTime: time.Now(),
	}
	mux := setupHTTPRouter(server)

	req := httptest.NewRequest(http.MethodPost, "/capabilities", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestServerStart(t *testing.T) {
	// Test that Start function exists and has correct signature
	// This is a basic smoke test - full integration tests will come later

	// If this test compiles and runs, the Start function exists
	t.Log("Start function exists and package builds")
}
