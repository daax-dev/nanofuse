package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_Health(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		resp := HealthResponse{
			Status:        "healthy",
			Version:       "0.1.0",
			UptimeSeconds: 3600,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	client := NewTCPClient(server.URL, 5*time.Second, false)

	// Test health endpoint
	ctx := context.Background()
	health, err := client.Health(ctx)
	if err != nil {
		t.Fatalf("Health() error = %v", err)
	}

	if health.Status != "healthy" {
		t.Errorf("expected status 'healthy', got '%s'", health.Status)
	}

	if health.Version != "0.1.0" {
		t.Errorf("expected version '0.1.0', got '%s'", health.Version)
	}

	if health.UptimeSeconds != 3600 {
		t.Errorf("expected uptime 3600, got %d", health.UptimeSeconds)
	}
}

func TestClient_Capabilities(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/capabilities" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		resp := CapabilitiesResponse{
			Status:  "ok",
			Version: "0.1.0",
			Host: HostCapabilities{
				OS:           "linux",
				Arch:         "amd64",
				KVMDevice:    "/dev/kvm",
				KVMExists:    true,
				KVMReadWrite: true,
			},
			Runtime: RuntimeCapabilities{
				NativeRuntime:        true,
				Driver:               "firecracker",
				FirecrackerBinary:    "/usr/local/bin/firecracker",
				FirecrackerAvailable: true,
				RootRequired:         true,
				NetworkSetupRequired: true,
			},
			API: APITransportCapabilities{
				UnixSocket: "/var/run/nanofused.sock",
				TCPBind:    "0.0.0.0:8080",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewTCPClient(server.URL, 5*time.Second, false)

	capabilities, err := client.Capabilities(context.Background())
	if err != nil {
		t.Fatalf("Capabilities() error = %v", err)
	}

	if capabilities.Status != "ok" {
		t.Errorf("expected status 'ok', got '%s'", capabilities.Status)
	}

	if !capabilities.Runtime.NativeRuntime {
		t.Error("expected native runtime to be true")
	}

	if capabilities.Runtime.Driver != "firecracker" {
		t.Errorf("expected runtime driver firecracker, got %q", capabilities.Runtime.Driver)
	}

	if capabilities.API.TCPBind != "0.0.0.0:8080" {
		t.Errorf("expected TCP bind '0.0.0.0:8080', got '%s'", capabilities.API.TCPBind)
	}
}

func TestRuntimeCapabilitiesJSONIncludesRequiredAndFalseDiagnostics(t *testing.T) {
	payload, err := json.Marshal(RuntimeCapabilities{
		Driver:  "firecracker",
		Message: "test",
	})
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	var got map[string]interface{}
	if err := json.Unmarshal(payload, &got); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if got["driver"] != "firecracker" {
		t.Fatalf("driver = %#v, want firecracker", got["driver"])
	}
	for _, key := range []string{
		"apple_container_available",
		"apple_container_running",
		"virtualization_framework_supported",
	} {
		if _, ok := got[key]; !ok {
			t.Fatalf("%s omitted from JSON payload %s", key, payload)
		}
		if _, ok := got[key].(bool); !ok {
			t.Fatalf("%s = %#v, want bool", key, got[key])
		}
	}
}

func TestClient_ListVMs(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/vms" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		resp := ListVMsResponse{
			VMs: []VM{
				{
					ID:    "vm-123",
					Name:  "test-vm",
					State: "running",
					Image: "ghcr.io/test/image:latest",
				},
			},
			Total: 1,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	client := NewTCPClient(server.URL, 5*time.Second, false)

	// Test list VMs
	ctx := context.Background()
	result, err := client.ListVMs(ctx, "")
	if err != nil {
		t.Fatalf("ListVMs() error = %v", err)
	}

	if result.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Total)
	}

	if len(result.VMs) != 1 {
		t.Errorf("expected 1 VM, got %d", len(result.VMs))
	}

	vm := result.VMs[0]
	if vm.ID != "vm-123" {
		t.Errorf("expected VM ID 'vm-123', got '%s'", vm.ID)
	}

	if vm.Name != "test-vm" {
		t.Errorf("expected VM name 'test-vm', got '%s'", vm.Name)
	}

	if vm.State != "running" {
		t.Errorf("expected VM state 'running', got '%s'", vm.State)
	}
}

func TestClient_CreateVM(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/vms" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("expected POST method, got %s", r.Method)
		}

		var req CreateVMRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}

		w.WriteHeader(http.StatusCreated)
		resp := VM{
			ID:     "vm-456",
			Name:   req.Name,
			State:  "created",
			Image:  req.Image,
			Config: req.Config,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	client := NewTCPClient(server.URL, 5*time.Second, false)

	// Test create VM
	ctx := context.Background()
	req := &CreateVMRequest{
		Name:  "new-vm",
		Image: "ghcr.io/test/image:latest",
		Config: VMConfig{
			VCPUs:     2,
			MemoryMiB: 512,
			Network: NetworkConfig{
				Mode: "nat",
			},
		},
	}

	vm, err := client.CreateVM(ctx, req)
	if err != nil {
		t.Fatalf("CreateVM() error = %v", err)
	}

	if vm.Name != "new-vm" {
		t.Errorf("expected VM name 'new-vm', got '%s'", vm.Name)
	}

	if vm.State != "created" {
		t.Errorf("expected VM state 'created', got '%s'", vm.State)
	}

	if vm.Config.VCPUs != 2 {
		t.Errorf("expected 2 vCPUs, got %d", vm.Config.VCPUs)
	}

	if vm.Config.MemoryMiB != 512 {
		t.Errorf("expected 512 MiB memory, got %d", vm.Config.MemoryMiB)
	}
}

func TestClient_ExecVM(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/vms/vm-123/exec" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST method, got %s", r.Method)
		}

		var req VMExecRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if len(req.Command) != 2 || req.Command[0] != "uname" || req.Command[1] != "-a" {
			t.Errorf("unexpected command: %#v", req.Command)
		}

		resp := VMExecResult{
			Command:   req.Command,
			ExitCode:  0,
			Stdout:    "Linux\n",
			RuntimeID: "nf-test",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewTCPClient(server.URL, 5*time.Second, false)
	result, err := client.ExecVM(context.Background(), "vm-123", &VMExecRequest{
		Command: []string{"uname", "-a"},
	})
	if err != nil {
		t.Fatalf("ExecVM() error = %v", err)
	}
	if result.Stdout != "Linux\n" {
		t.Fatalf("stdout = %q, want Linux", result.Stdout)
	}
	if result.RuntimeID != "nf-test" {
		t.Fatalf("runtime id = %q, want nf-test", result.RuntimeID)
	}
}

func TestClient_ErrorHandling(t *testing.T) {
	// Create mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		resp := APIError{
			Error: ErrorDetails{
				Code:    "VM_NOT_FOUND",
				Message: "Virtual machine not found",
				Details: map[string]interface{}{
					"vm_id": "vm-999",
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	client := NewTCPClient(server.URL, 5*time.Second, false)

	// Test error handling
	ctx := context.Background()
	_, err := client.GetVM(ctx, "vm-999")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	clientErr, ok := err.(*ClientError)
	if !ok {
		t.Fatalf("expected ClientError, got %T", err)
	}

	if clientErr.StatusCode != http.StatusNotFound {
		t.Errorf("expected status code 404, got %d", clientErr.StatusCode)
	}

	if clientErr.Code != "VM_NOT_FOUND" {
		t.Errorf("expected error code 'VM_NOT_FOUND', got '%s'", clientErr.Code)
	}

	if clientErr.Message != "Virtual machine not found" {
		t.Errorf("unexpected error message: %s", clientErr.Message)
	}

	// Test exit code mapping
	if clientErr.ExitCode() != 4 {
		t.Errorf("expected exit code 4, got %d", clientErr.ExitCode())
	}
}
