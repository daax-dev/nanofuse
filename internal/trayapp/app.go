package trayapp

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/daax-dev/nanofuse/internal/client"
)

const (
	DefaultAPISocketPath = "/var/run/nanofused.sock"
	DefaultTimeout       = 10 * time.Second
)

type Config struct {
	APIURL    string        `json:"api_url,omitempty"`
	APISocket string        `json:"api_socket,omitempty"`
	Timeout   time.Duration `json:"timeout"`
	Debug     bool          `json:"debug"`
}

type Status struct {
	Endpoint     string                       `json:"endpoint"`
	CheckedAt    time.Time                    `json:"checked_at"`
	Health       *client.HealthResponse       `json:"health,omitempty"`
	Capabilities *client.CapabilitiesResponse `json:"capabilities,omitempty"`
	VMs          []client.VM                  `json:"vms,omitempty"`
	Images       []client.Image               `json:"images,omitempty"`
	Error        string                       `json:"error,omitempty"`
}

type API interface {
	Health(context.Context) (*client.HealthResponse, error)
	Capabilities(context.Context) (*client.CapabilitiesResponse, error)
	CreateVM(context.Context, *client.CreateVMRequest) (*client.VM, error)
	ListVMs(context.Context, string) (*client.ListVMsResponse, error)
	ListImages(context.Context) (*client.ListImagesResponse, error)
	StartVM(context.Context, string) (*client.VM, error)
	StopVM(context.Context, string, int) (*client.VM, error)
	KillVM(context.Context, string) (*client.VM, error)
	DeleteVM(context.Context, string) error
}

type VMAction string

const (
	VMActionStart  VMAction = "start"
	VMActionStop   VMAction = "stop"
	VMActionKill   VMAction = "kill"
	VMActionDelete VMAction = "delete"
)

const (
	DefaultVCPUs           = 2
	DefaultMemoryMiB       = 512
	DefaultNetworkMode     = "nat"
	DefaultPublishedVMPort = 8080
)

func ConfigFromEnv() Config {
	cfg := Config{
		APIURL:    firstNonEmpty(os.Getenv("NANOFUSE_TRAY_API_URL"), os.Getenv("NANOFUSE_API_URL")),
		APISocket: firstNonEmpty(os.Getenv("NANOFUSE_TRAY_API_SOCKET"), os.Getenv("NANOFUSE_API_SOCKET")),
		Timeout:   DefaultTimeout,
		Debug:     truthy(firstNonEmpty(os.Getenv("NANOFUSE_TRAY_DEBUG"), os.Getenv("NANOFUSE_DEBUG"))),
	}
	if value := firstNonEmpty(os.Getenv("NANOFUSE_TRAY_TIMEOUT"), os.Getenv("NANOFUSE_TIMEOUT")); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			cfg.Timeout = parsed
		}
	}
	return cfg.Normalize()
}

func (c Config) Normalize() Config {
	if c.Timeout <= 0 {
		c.Timeout = DefaultTimeout
	}
	if c.APIURL == "" && c.APISocket == "" {
		c.APISocket = DefaultAPISocketPath
	}
	return c
}

func (c Config) Endpoint() string {
	cfg := c.Normalize()
	if cfg.APIURL != "" {
		return cfg.APIURL
	}
	return "unix://" + cfg.APISocket
}

func (c Config) NewClient() *client.Client {
	cfg := c.Normalize()
	if cfg.APIURL != "" {
		return client.NewTCPClient(cfg.APIURL, cfg.Timeout, cfg.Debug)
	}
	return client.NewClient(cfg.APISocket, cfg.Timeout, cfg.Debug)
}

func CollectStatus(ctx context.Context, api API, endpoint string) (*Status, error) {
	status := &Status{
		Endpoint:  endpoint,
		CheckedAt: time.Now().UTC(),
	}

	health, err := api.Health(ctx)
	if err != nil {
		status.Error = fmt.Sprintf("health: %v", err)
		return status, fmt.Errorf("health: %w", err)
	}
	status.Health = health

	capabilities, err := api.Capabilities(ctx)
	if err != nil {
		status.Error = fmt.Sprintf("capabilities: %v", err)
		return status, fmt.Errorf("capabilities: %w", err)
	}
	status.Capabilities = capabilities

	vms, err := api.ListVMs(ctx, "")
	if err != nil {
		status.Error = fmt.Sprintf("list VMs: %v", err)
		return status, fmt.Errorf("list VMs: %w", err)
	}
	if vms != nil {
		status.VMs = vms.VMs
	}

	images, err := api.ListImages(ctx)
	if err != nil {
		status.Error = fmt.Sprintf("list images: %v", err)
		return status, fmt.Errorf("list images: %w", err)
	}
	if images != nil {
		status.Images = images.Images
	}

	return status, nil
}

func ExecuteVMAction(ctx context.Context, api API, action VMAction, vmID string) (*client.VM, error) {
	if strings.TrimSpace(vmID) == "" {
		return nil, fmt.Errorf("vm id is required")
	}

	switch action {
	case VMActionStart:
		return api.StartVM(ctx, vmID)
	case VMActionStop:
		return api.StopVM(ctx, vmID, 30)
	case VMActionKill:
		return api.KillVM(ctx, vmID)
	case VMActionDelete:
		return nil, api.DeleteVM(ctx, vmID)
	default:
		return nil, fmt.Errorf("unsupported VM action %q", action)
	}
}

func LaunchVMFromImage(ctx context.Context, api API, imageRef string) (*client.VM, error) {
	imageRef = strings.TrimSpace(imageRef)
	if imageRef == "" {
		return nil, fmt.Errorf("image reference is required")
	}

	portForwards, err := defaultLaunchPortForwards()
	if err != nil {
		return nil, fmt.Errorf("allocate host port for image %q: %w", imageRef, err)
	}

	vm, err := api.CreateVM(ctx, &client.CreateVMRequest{
		Image: imageRef,
		Config: client.VMConfig{
			VCPUs:     DefaultVCPUs,
			MemoryMiB: DefaultMemoryMiB,
			Network: client.NetworkConfig{
				Mode:         DefaultNetworkMode,
				PortForwards: portForwards,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create VM from image %q: %w", imageRef, err)
	}
	if vm == nil || strings.TrimSpace(vm.ID) == "" {
		return nil, fmt.Errorf("create VM from image %q returned no VM id", imageRef)
	}

	started, err := api.StartVM(ctx, vm.ID)
	if err != nil {
		return nil, fmt.Errorf("start VM %q: %w", vm.ID, err)
	}
	return started, nil
}

func defaultLaunchPortForwards() ([]client.PortForward, error) {
	hostPort, err := availableLocalTCPPort()
	if err != nil {
		return nil, err
	}
	return []client.PortForward{
		{
			HostPort: hostPort,
			VMPort:   DefaultPublishedVMPort,
			Protocol: "tcp",
		},
	}, nil
}

func availableLocalTCPPort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	_, portText, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		return 0, err
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		return 0, err
	}
	return port, nil
}

func RuntimeSummary(status *Status) string {
	if status == nil {
		return "unknown"
	}
	if status.Error != "" {
		return "error: " + status.Error
	}
	if status.Capabilities == nil {
		return "capabilities unavailable"
	}
	if status.Capabilities.Runtime.NativeRuntime {
		return "runtime ready"
	}
	if status.Capabilities.Runtime.Message != "" {
		return status.Capabilities.Runtime.Message
	}
	return "runtime unavailable"
}

func VMActionReady(status *Status) bool {
	return status != nil &&
		status.Error == "" &&
		status.Capabilities != nil &&
		status.Capabilities.Runtime.NativeRuntime
}

func VMActionAllowed(vm *client.VM, action VMAction) bool {
	if vm == nil || strings.TrimSpace(vm.ID) == "" {
		return false
	}

	state := strings.ToLower(strings.TrimSpace(vm.State))
	switch action {
	case VMActionStart:
		return state == "created" || state == "stopped"
	case VMActionStop:
		return state == "running" || state == "paused"
	case VMActionKill:
		return vmHasRuntimeHandle(vm) && (state == "running" ||
			state == "starting" ||
			state == "stopping" ||
			state == "pausing" ||
			state == "paused" ||
			state == "resuming")
	case VMActionDelete:
		return true
	default:
		return false
	}
}

func vmHasRuntimeHandle(vm *client.VM) bool {
	return vm != nil &&
		vm.Runtime != nil &&
		(vm.Runtime.PID != 0 ||
			strings.TrimSpace(vm.Runtime.ExternalID) != "" ||
			strings.TrimSpace(vm.Runtime.SocketPath) != "")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func truthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
