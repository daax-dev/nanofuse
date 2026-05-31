package config

import (
	"fmt"
	"os"
	"runtime"

	"gopkg.in/yaml.v3"
)

// Config represents the daemon configuration
type Config struct {
	API         APIConfig         `yaml:"api"`
	Storage     StorageConfig     `yaml:"storage"`
	Runtime     RuntimeConfig     `yaml:"runtime"`
	Firecracker FirecrackerConfig `yaml:"firecracker"`
	Limits      LimitsConfig      `yaml:"limits"`
	Registry    RegistryConfig    `yaml:"registry"`
	Logging     LoggingConfig     `yaml:"logging"`
	Network     NetworkConfig     `yaml:"network"`
	SPIRE       SPIREConfig       `yaml:"spire"`
	Auth        AuthConfig        `yaml:"auth"`
}

// RuntimeConfig selects the local microVM runtime backend.
type RuntimeConfig struct {
	// Driver may be "auto", "firecracker", or "apple_container".
	Driver         string                      `yaml:"driver"`
	AppleContainer AppleContainerRuntimeConfig `yaml:"apple_container,omitempty"`
}

// AppleContainerRuntimeConfig configures the macOS Apple container backend.
type AppleContainerRuntimeConfig struct {
	BinaryPath     string `yaml:"binary_path,omitempty"`
	AutoStart      bool   `yaml:"auto_start"`
	DefaultCommand string `yaml:"default_command,omitempty"`
}

// AuthConfig holds authentication configuration for the daemon.
type AuthConfig struct {
	// Enabled requires mTLS client identity on TCP API listeners. Unix socket
	// listeners continue to rely on filesystem permissions.
	Enabled bool `yaml:"enabled"`

	// TLSCertFile is the daemon certificate presented on TCP listeners.
	TLSCertFile string `yaml:"tls_cert_file,omitempty"`

	// TLSKeyFile is the daemon private key presented on TCP listeners.
	TLSKeyFile string `yaml:"tls_key_file,omitempty"`

	// ClientCAFile is the CA bundle used to verify client certificates.
	ClientCAFile string `yaml:"client_ca_file,omitempty"`
}

// SPIREConfig represents SPIRE integration configuration
type SPIREConfig struct {
	Enabled       bool   `yaml:"enabled"`        // Enable SPIRE workload registration
	ServerSocket  string `yaml:"server_socket"`  // Path to SPIRE server API socket
	TrustDomain   string `yaml:"trust_domain"`   // SPIFFE trust domain (e.g., poley.dev)
	ParentID      string `yaml:"parent_id"`      // Parent SPIFFE ID for workload entries
	WorkloadType  string `yaml:"workload_type"`  // Workload type segment (default: microvm)
	DefaultTTL    int    `yaml:"default_ttl"`    // Default SVID TTL in seconds
	VsockCID      uint32 `yaml:"vsock_cid"`      // Vsock CID for SPIRE agent proxy (0 = disabled)
	VsockPort     uint32 `yaml:"vsock_port"`     // Vsock port for SPIRE agent proxy
	AgentSocket   string `yaml:"agent_socket"`   // Path to SPIRE agent workload API socket (for vsock proxy)
	ContainerName string `yaml:"container_name"` // Docker container name for SPIRE server (default: spire-server)
}

// APIConfig represents API configuration
type APIConfig struct {
	Socket      string `yaml:"socket"`
	SocketMode  string `yaml:"socket_mode"`
	SocketGroup string `yaml:"socket_group"`
	TCPBind     string `yaml:"tcp_bind,omitempty"`
}

// StorageConfig represents storage configuration
type StorageConfig struct {
	DataDir  string `yaml:"data_dir"`
	Database string `yaml:"database"`
}

// FirecrackerConfig represents Firecracker configuration
type FirecrackerConfig struct {
	BinaryPath string `yaml:"binary_path"`
	JailerPath string `yaml:"jailer_path,omitempty"`
}

// LimitsConfig represents resource limits
type LimitsConfig struct {
	MaxVMs                int `yaml:"max_vms"`
	MaxTotalMemoryMiB     int `yaml:"max_total_memory_mib"`
	MaxVCPUsPerVM         int `yaml:"max_vcpus_per_vm"`
	MaxMemoryPerVMMiB     int `yaml:"max_memory_per_vm_mib"`
	MaxSnapshotStorageGiB int `yaml:"max_snapshot_storage_gib"`
}

// RegistryConfig represents registry configuration
type RegistryConfig struct {
	AuthConfigPath   string              `yaml:"auth_config_path"`
	Auth             map[string]AuthInfo `yaml:"auth,omitempty"`
	PullTimeoutSecs  int                 `yaml:"pull_timeout_secs"`  // Timeout for image pull operations
	LayerTimeoutSecs int                 `yaml:"layer_timeout_secs"` // Timeout per layer download
}

// AuthInfo represents registry authentication
type AuthInfo struct {
	Username string `yaml:"username"`
	Token    string `yaml:"token"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level                string `yaml:"level"`
	Format               string `yaml:"format"`
	FilePath             string `yaml:"file_path"` // Path to log file (empty = no file logging)
	ConsoleLogMaxSizeMB  int    `yaml:"console_log_max_size_mb"`
	ConsoleLogMaxBackups int    `yaml:"console_log_max_backups"`
}

// NetworkConfig represents daemon-managed host network setup.
type NetworkConfig struct {
	Setup bool `yaml:"setup"` // Create bridge/NAT at daemon startup.
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		API: APIConfig{
			Socket:      "/var/run/nanofused.sock",
			SocketMode:  "0660",
			SocketGroup: "nanofuse",
		},
		Storage: StorageConfig{
			DataDir:  "/var/lib/nanofuse",
			Database: "/var/lib/nanofuse/nanofuse.db",
		},
		Runtime: RuntimeConfig{
			Driver: "auto",
			AppleContainer: AppleContainerRuntimeConfig{
				BinaryPath:     "/usr/local/bin/container",
				AutoStart:      true,
				DefaultCommand: "sleep infinity",
			},
		},
		Firecracker: FirecrackerConfig{
			BinaryPath: "/usr/local/bin/firecracker",
		},
		Limits: LimitsConfig{
			MaxVMs:                50,
			MaxTotalMemoryMiB:     32768,
			MaxVCPUsPerVM:         8,
			MaxMemoryPerVMMiB:     8192,
			MaxSnapshotStorageGiB: 100,
		},
		Registry: RegistryConfig{
			AuthConfigPath:   "/root/.docker/config.json",
			PullTimeoutSecs:  600, // 10 minutes total for pull
			LayerTimeoutSecs: 300, // 5 minutes per layer
		},
		Logging: LoggingConfig{
			Level:                "debug",
			Format:               "text",
			FilePath:             "/var/log/nanofuse/nanofused.log",
			ConsoleLogMaxSizeMB:  10,
			ConsoleLogMaxBackups: 3,
		},
		Network: NetworkConfig{
			Setup: true,
		},
		SPIRE: SPIREConfig{
			Enabled:       false,                                // Disabled by default
			ServerSocket:  "/tmp/spire-server/private/api.sock", // Default SPIRE server socket
			TrustDomain:   "poley.dev",                          // Default trust domain
			ParentID:      "spiffe://poley.dev/agent/local",     // Default parent SPIFFE ID
			WorkloadType:  "microvm",                            // Default workload type
			DefaultTTL:    3600,                                 // 1 hour
			VsockCID:      0,                                    // Disabled by default
			VsockPort:     8307,                                 // Default SPIRE agent vsock port
			AgentSocket:   "/run/spire/sockets/agent.sock",      // Default SPIRE agent socket
			ContainerName: "spire-server",                       // Default Docker container name
		},
		Auth: AuthConfig{
			Enabled: false,
		},
	}
}

// Load loads configuration from a file
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	// If no path provided, return defaults
	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // Return defaults if file doesn't exist
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Storage.DataDir == "" {
		return fmt.Errorf("storage.data_dir is required")
	}
	if c.Storage.Database == "" {
		return fmt.Errorf("storage.database is required")
	}
	if err := c.validateRuntime(); err != nil {
		return err
	}
	if c.Limits.MaxVMs <= 0 {
		return fmt.Errorf("limits.max_vms must be positive")
	}
	if c.Limits.MaxTotalMemoryMiB <= 0 {
		return fmt.Errorf("limits.max_total_memory_mib must be positive")
	}
	return c.validateAuth()
}

func (c *Config) validateRuntime() error {
	if c.Runtime.Driver == "" {
		c.Runtime.Driver = "auto"
	}
	driver, err := runtimeDriverForHost(c.Runtime.Driver, runtime.GOOS)
	if err != nil {
		return err
	}
	if driver == "firecracker" {
		if c.Firecracker.BinaryPath == "" {
			return fmt.Errorf("firecracker.binary_path is required")
		}
	}
	if driver == "apple_container" {
		if c.Runtime.AppleContainer.BinaryPath == "" {
			c.Runtime.AppleContainer.BinaryPath = "/usr/local/bin/container"
		}
		if c.Runtime.AppleContainer.DefaultCommand == "" {
			c.Runtime.AppleContainer.DefaultCommand = "sleep infinity"
		}
	}
	return nil
}

func runtimeDriverForHost(driver, goos string) (string, error) {
	if driver == "" {
		driver = "auto"
	}
	switch driver {
	case "auto":
		switch goos {
		case "darwin":
			return "apple_container", nil
		case "linux":
			return "firecracker", nil
		default:
			return "", fmt.Errorf("runtime.driver auto does not support host OS %q; use linux with firecracker or darwin with apple_container", goos)
		}
	case "firecracker":
		if goos != "linux" {
			return "", fmt.Errorf("runtime.driver firecracker requires linux host, got %q", goos)
		}
		return "firecracker", nil
	case "apple_container":
		if goos != "darwin" {
			return "", fmt.Errorf("runtime.driver apple_container requires darwin host, got %q", goos)
		}
		return "apple_container", nil
	default:
		return "", fmt.Errorf("runtime.driver must be one of auto, firecracker, apple_container")
	}
}

func (c *Config) validateAuth() error {
	if !c.Auth.Enabled || c.API.TCPBind == "" {
		return nil
	}
	if c.Auth.TLSCertFile == "" {
		return fmt.Errorf("auth.tls_cert_file is required when auth.enabled is true for TCP listeners")
	}
	if c.Auth.TLSKeyFile == "" {
		return fmt.Errorf("auth.tls_key_file is required when auth.enabled is true for TCP listeners")
	}
	if c.Auth.ClientCAFile == "" {
		return fmt.Errorf("auth.client_ca_file is required when auth.enabled is true for TCP listeners")
	}
	return nil
}
