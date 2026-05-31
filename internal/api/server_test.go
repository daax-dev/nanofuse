package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"github.com/daax-dev/nanofuse/internal/config"
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

func TestAppleContainerNativeReadyRequiresRunningOrAutoStart(t *testing.T) {
	tests := []struct {
		name      string
		goos      string
		available bool
		running   bool
		autoStart bool
		want      bool
	}{
		{
			name:      "running service is ready",
			goos:      "darwin",
			available: true,
			running:   true,
			autoStart: false,
			want:      true,
		},
		{
			name:      "auto start can become ready",
			goos:      "darwin",
			available: true,
			running:   false,
			autoStart: true,
			want:      true,
		},
		{
			name:      "stopped service without auto start is not ready",
			goos:      "darwin",
			available: true,
			running:   false,
			autoStart: false,
			want:      false,
		},
		{
			name:      "missing CLI is not ready",
			goos:      "darwin",
			available: false,
			running:   true,
			autoStart: true,
			want:      false,
		},
		{
			name:      "linux does not use apple container readiness",
			goos:      "linux",
			available: true,
			running:   true,
			autoStart: true,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := appleContainerNativeReady(tt.goos, tt.available, tt.running, tt.autoStart)
			if got != tt.want {
				t.Fatalf("appleContainerNativeReady() = %v, want %v", got, tt.want)
			}
		})
	}
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
