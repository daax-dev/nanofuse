package trayapp

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/daax-dev/nanofuse/internal/client"
)

type fakeAPI struct {
	calls     []string
	errAt     string
	createReq *client.CreateVMRequest
}

func (f *fakeAPI) Health(context.Context) (*client.HealthResponse, error) {
	f.calls = append(f.calls, "health")
	if f.errAt == "health" {
		return nil, errors.New("health failed")
	}
	return &client.HealthResponse{Status: "healthy", Version: "test", UptimeSeconds: 12}, nil
}

func (f *fakeAPI) Capabilities(context.Context) (*client.CapabilitiesResponse, error) {
	f.calls = append(f.calls, "capabilities")
	if f.errAt == "capabilities" {
		return nil, errors.New("capabilities failed")
	}
	return &client.CapabilitiesResponse{
		Status:  "ok",
		Version: "test",
		Runtime: client.RuntimeCapabilities{
			NativeRuntime:        true,
			FirecrackerAvailable: true,
			Message:              "ready",
		},
	}, nil
}

func (f *fakeAPI) ListVMs(context.Context, string) (*client.ListVMsResponse, error) {
	f.calls = append(f.calls, "vms")
	if f.errAt == "vms" {
		return nil, errors.New("vms failed")
	}
	return &client.ListVMsResponse{
		VMs: []client.VM{{ID: "vm-1", Name: "agent", State: "running", Image: "nanofuse-ci:latest"}},
	}, nil
}

func (f *fakeAPI) ListImages(context.Context) (*client.ListImagesResponse, error) {
	f.calls = append(f.calls, "images")
	if f.errAt == "images" {
		return nil, errors.New("images failed")
	}
	return &client.ListImagesResponse{
		Images: []client.Image{{Digest: "sha256:abc", Tags: []string{"nanofuse-ci:latest"}}},
	}, nil
}

func (f *fakeAPI) CreateVM(_ context.Context, req *client.CreateVMRequest) (*client.VM, error) {
	f.calls = append(f.calls, "create:"+req.Image)
	f.createReq = req
	return &client.VM{ID: "vm-created", State: "created", Image: req.Image}, nil
}

func (f *fakeAPI) StartVM(context.Context, string) (*client.VM, error) {
	f.calls = append(f.calls, "start")
	return &client.VM{ID: "vm-1", State: "running"}, nil
}

func (f *fakeAPI) StopVM(context.Context, string, int) (*client.VM, error) {
	f.calls = append(f.calls, "stop")
	return &client.VM{ID: "vm-1", State: "stopped"}, nil
}

func (f *fakeAPI) KillVM(context.Context, string) (*client.VM, error) {
	f.calls = append(f.calls, "kill")
	return &client.VM{ID: "vm-1", State: "killed"}, nil
}

func (f *fakeAPI) DeleteVM(context.Context, string) error {
	f.calls = append(f.calls, "delete")
	return nil
}

func TestCollectStatusCallsRequiredAPIEndpoints(t *testing.T) {
	api := &fakeAPI{}

	status, err := CollectStatus(context.Background(), api, "http://127.0.0.1:18080")
	if err != nil {
		t.Fatalf("CollectStatus() error = %v", err)
	}

	wantCalls := []string{"health", "capabilities", "vms", "images"}
	if !reflect.DeepEqual(api.calls, wantCalls) {
		t.Fatalf("calls = %v, want %v", api.calls, wantCalls)
	}
	if status.Endpoint != "http://127.0.0.1:18080" {
		t.Fatalf("endpoint = %q", status.Endpoint)
	}
	if status.Health.Status != "healthy" {
		t.Fatalf("health status = %q", status.Health.Status)
	}
	if len(status.VMs) != 1 || status.VMs[0].Name != "agent" {
		t.Fatalf("VMs = %#v", status.VMs)
	}
	if len(status.Images) != 1 || len(status.Images[0].Tags) != 1 {
		t.Fatalf("images = %#v", status.Images)
	}
}

func TestCollectStatusReturnsPartialStatusOnError(t *testing.T) {
	api := &fakeAPI{errAt: "capabilities"}

	status, err := CollectStatus(context.Background(), api, "unix:///var/run/nanofused.sock")
	if err == nil {
		t.Fatal("CollectStatus() error = nil")
	}
	if status == nil {
		t.Fatal("status = nil")
	}
	if status.Health == nil {
		t.Fatal("expected health to be populated before capabilities failure")
	}
	if status.Error == "" {
		t.Fatal("expected status error")
	}
	wantCalls := []string{"health", "capabilities"}
	if !reflect.DeepEqual(api.calls, wantCalls) {
		t.Fatalf("calls = %v, want %v", api.calls, wantCalls)
	}
}

func TestExecuteVMAction(t *testing.T) {
	tests := []struct {
		name   string
		action VMAction
		call   string
	}{
		{name: "start", action: VMActionStart, call: "start"},
		{name: "stop", action: VMActionStop, call: "stop"},
		{name: "kill", action: VMActionKill, call: "kill"},
		{name: "delete", action: VMActionDelete, call: "delete"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &fakeAPI{}
			_, err := ExecuteVMAction(context.Background(), api, tt.action, "vm-1")
			if err != nil {
				t.Fatalf("ExecuteVMAction() error = %v", err)
			}
			if len(api.calls) != 1 || api.calls[0] != tt.call {
				t.Fatalf("calls = %v, want %s", api.calls, tt.call)
			}
		})
	}
}

func TestExecuteVMActionRejectsMissingVMID(t *testing.T) {
	_, err := ExecuteVMAction(context.Background(), &fakeAPI{}, VMActionStart, " ")
	if err == nil {
		t.Fatal("ExecuteVMAction() error = nil")
	}
}

func TestLaunchVMFromImageCreatesAndStartsVM(t *testing.T) {
	api := &fakeAPI{}

	vm, err := LaunchVMFromImage(context.Background(), api, "nanofuse-ci:latest")
	if err != nil {
		t.Fatalf("LaunchVMFromImage() error = %v", err)
	}
	wantCalls := []string{"create:nanofuse-ci:latest", "start"}
	if !reflect.DeepEqual(api.calls, wantCalls) {
		t.Fatalf("calls = %v, want %v", api.calls, wantCalls)
	}
	if vm.State != "running" {
		t.Fatalf("state = %q, want running", vm.State)
	}
	if api.createReq.Config.VCPUs != DefaultVCPUs || api.createReq.Config.MemoryMiB != DefaultMemoryMiB {
		t.Fatalf("config = %#v, want default resources", api.createReq.Config)
	}
	if api.createReq.Config.Network.Mode != DefaultNetworkMode {
		t.Fatalf("network mode = %q, want %q", api.createReq.Config.Network.Mode, DefaultNetworkMode)
	}
}

func TestLaunchVMFromImageRejectsMissingImage(t *testing.T) {
	_, err := LaunchVMFromImage(context.Background(), &fakeAPI{}, " ")
	if err == nil {
		t.Fatal("LaunchVMFromImage() error = nil")
	}
}

func TestVMActionReady(t *testing.T) {
	readyStatus := &Status{
		Capabilities: &client.CapabilitiesResponse{
			Runtime: client.RuntimeCapabilities{NativeRuntime: true},
		},
	}

	tests := []struct {
		name   string
		status *Status
		want   bool
	}{
		{name: "ready", status: readyStatus, want: true},
		{name: "nil status", status: nil, want: false},
		{name: "status error", status: &Status{Error: "health failed", Capabilities: readyStatus.Capabilities}, want: false},
		{name: "missing capabilities", status: &Status{}, want: false},
		{name: "native runtime false", status: &Status{Capabilities: &client.CapabilitiesResponse{}}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := VMActionReady(tt.status); got != tt.want {
				t.Fatalf("VMActionReady() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigFromEnv(t *testing.T) {
	t.Setenv("NANOFUSE_TRAY_API_URL", "http://127.0.0.1:18080")
	t.Setenv("NANOFUSE_TRAY_TIMEOUT", "250ms")
	t.Setenv("NANOFUSE_TRAY_DEBUG", "true")

	cfg := ConfigFromEnv()
	if cfg.APIURL != "http://127.0.0.1:18080" {
		t.Fatalf("APIURL = %q", cfg.APIURL)
	}
	if cfg.Endpoint() != "http://127.0.0.1:18080" {
		t.Fatalf("Endpoint() = %q", cfg.Endpoint())
	}
	if cfg.Timeout != 250*time.Millisecond {
		t.Fatalf("Timeout = %v", cfg.Timeout)
	}
	if !cfg.Debug {
		t.Fatal("Debug = false")
	}
}

func TestConfigDefaultsToUnixSocket(t *testing.T) {
	cfg := (Config{}).Normalize()
	if cfg.Endpoint() != "unix://"+DefaultAPISocketPath {
		t.Fatalf("Endpoint() = %q", cfg.Endpoint())
	}
	if cfg.Timeout != DefaultTimeout {
		t.Fatalf("Timeout = %v", cfg.Timeout)
	}
}
