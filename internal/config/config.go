package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the daemon configuration
type Config struct {
	API         APIConfig         `yaml:"api"`
	Storage     StorageConfig     `yaml:"storage"`
	Firecracker FirecrackerConfig `yaml:"firecracker"`
	Limits      LimitsConfig      `yaml:"limits"`
	Registry    RegistryConfig    `yaml:"registry"`
	Logging     LoggingConfig     `yaml:"logging"`
	SPIRE       SPIREConfig       `yaml:"spire"`
	Auth        AuthConfig        `yaml:"auth"`
}

// AuthConfig holds authentication configuration for the daemon.
// When DevStaticKeys is false (production), only SPIFFE SVIDs are accepted.
// Set DEV_STATIC_KEYS=true in the environment to also allow static API key auth.
type AuthConfig struct {
	// Enabled turns the auth middleware on. When false all requests pass through.
	Enabled bool `yaml:"enabled"`

	// StaticAPIKeys lists API keys permitted when running in dev mode only.
	// Production MUST leave this empty; the middleware enforces the restriction.
	StaticAPIKeys []string `yaml:"static_api_keys,omitempty"`

	// SVIDRotation controls the automatic rotation of SPIFFE SVIDs.
	SVIDRotation SVIDRotationConfig `yaml:"svid_rotation"`
}

// SVIDRotationConfig controls SVID rotation behaviour (issue #4).
type SVIDRotationConfig struct {
	// MaxTTLSeconds is the maximum SVID lifetime (default 3600 = 60 min).
	MaxTTLSeconds int `yaml:"max_ttl_seconds"`

	// PreRefreshSeconds is how many seconds before expiry to trigger rotation
	// (default 900 = 15 min).
	PreRefreshSeconds int `yaml:"pre_refresh_seconds"`

	// GracePeriodSeconds is how long the old SVID remains valid after the new
	// one has been issued (default 300 = 5 min).
	GracePeriodSeconds int `yaml:"grace_period_seconds"`

	// StaleAlertSeconds is how long to wait for the agent to pick up the new
	// SVID before emitting a warning alert (default 300 = 5 min).
	StaleAlertSeconds int `yaml:"stale_alert_seconds"`
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
			Enabled: false, // Disabled by default; enable in production
			SVIDRotation: SVIDRotationConfig{
				MaxTTLSeconds:      3600, // 60 minutes
				PreRefreshSeconds:  900,  // 15 minutes before expiry
				GracePeriodSeconds: 300,  // 5-minute grace window
				StaleAlertSeconds:  300,  // alert if agent hasn't picked up within 5 min
			},
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

	data, err := os.ReadFile(path) //nolint:gosec // path is from daemon config, not user input
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
	if c.Firecracker.BinaryPath == "" {
		return fmt.Errorf("firecracker.binary_path is required")
	}
	if c.Limits.MaxVMs <= 0 {
		return fmt.Errorf("limits.max_vms must be positive")
	}
	if c.Limits.MaxTotalMemoryMiB <= 0 {
		return fmt.Errorf("limits.max_total_memory_mib must be positive")
	}
	return nil
}
