package types

import "time"

// VMState represents the state of a VM
type VMState string

const (
	StateCreated  VMState = "created"
	StateStarting VMState = "starting"
	StateRunning  VMState = "running"
	StateStopping VMState = "stopping"
	StateStopped  VMState = "stopped"
	StatePausing  VMState = "pausing"
	StatePaused   VMState = "paused"
	StateResuming VMState = "resuming"
	StateFailed   VMState = "failed"
)

// VM represents a virtual machine instance
type VM struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	State        VMState    `json:"state"`
	Image        string     `json:"image"`
	ImageDigest  string     `json:"image_digest"`
	Architecture string     `json:"architecture"`
	Config       VMConfig   `json:"config"`
	Runtime      *VMRuntime `json:"runtime,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	LockedBy     *string    `json:"locked_by,omitempty"`
	LockedAt     *time.Time `json:"locked_at,omitempty"`

	// SPIFFE human attribution (D025 format)
	OwnerUserID string `json:"owner_user_id,omitempty"` // Human owner for SPIFFE ID /u/ segment
	GroupID     string `json:"group_id,omitempty"`      // Group for SPIFFE ID /g/ segment
	SpiffeID    string `json:"spiffe_id,omitempty"`     // Generated SPIFFE ID if auto-registered
}

// VMConfig represents VM configuration
type VMConfig struct {
	VCPUs        int           `json:"vcpus"`
	MemoryMiB    int           `json:"memory_mib"`
	KernelArgs   string        `json:"kernel_args"`
	SSHPublicKey string        `json:"ssh_public_key,omitempty"` // Base64-encoded SSH public key
	Network      NetworkConfig `json:"network"`
	Disks        []DiskConfig  `json:"disks"`
}

// NetworkConfig represents network configuration
type NetworkConfig struct {
	Mode         string        `json:"mode"` // "nat", "bridged", "none"
	TapDevice    string        `json:"tap_device"`
	MACAddress   string        `json:"mac_address"`
	BridgeName   *string       `json:"bridge_name,omitempty"`
	IPAddress    string        `json:"ip_address,omitempty"`
	Gateway      string        `json:"gateway,omitempty"`
	Netmask      string        `json:"netmask,omitempty"`
	PortForwards []PortForward `json:"port_forwards,omitempty"`
	EgressPolicy *EgressPolicy `json:"egress_policy,omitempty"`
}

// PortForward represents a port forwarding rule from host to VM
type PortForward struct {
	HostPort int    `json:"host_port"`
	VMPort   int    `json:"vm_port"`
	Protocol string `json:"protocol"` // "tcp" or "udp"
}

// EgressPolicy controls outbound VM network access.
// When enabled with default_action unset, the daemon treats it as "deny".
type EgressPolicy struct {
	Enabled       bool         `json:"enabled"`
	DefaultAction string       `json:"default_action,omitempty"` // "deny" or "allow"
	AllowDNS      bool         `json:"allow_dns,omitempty"`
	ProxyOnly     bool         `json:"proxy_only,omitempty"`
	Proxy         *EgressProxy `json:"proxy,omitempty"`
	AllowRules    []EgressRule `json:"allow_rules,omitempty"`
}

// EgressProxy is the host-controlled proxy endpoint a VM may reach.
type EgressProxy struct {
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol,omitempty"` // "tcp" or "udp"; defaults to "tcp"
}

// EgressRule allows one outbound L3/L4 destination.
type EgressRule struct {
	CIDR        string `json:"cidr"`
	Protocol    string `json:"protocol"` // "tcp" or "udp"
	Port        int    `json:"port"`
	Description string `json:"description,omitempty"`
}

// DiskConfig represents disk configuration
type DiskConfig struct {
	DriveID      string `json:"drive_id"`
	PathOnHost   string `json:"path_on_host"`
	IsReadOnly   bool   `json:"is_read_only"`
	IsRootDevice bool   `json:"is_root_device"`
}

// VMRuntime represents VM runtime information
type VMRuntime struct {
	PID         int                 `json:"pid"`
	Driver      string              `json:"driver,omitempty"`
	ExternalID  string              `json:"external_id,omitempty"`
	SocketPath  string              `json:"socket_path,omitempty"`
	ConsolePath string              `json:"console_path"`
	NetworkInfo *NetworkRuntimeInfo `json:"network_info,omitempty"`
}

// NetworkRuntimeInfo represents network runtime information
type NetworkRuntimeInfo struct {
	TapDevice string `json:"tap_device"`
	HostIP    string `json:"host_ip"`
	GuestIP   string `json:"guest_ip"`
	Gateway   string `json:"gateway"`
}

// CreateVMRequest represents a request to create a VM
type CreateVMRequest struct {
	Name   string           `json:"name,omitempty"`
	Image  string           `json:"image"`
	Config *VMConfigRequest `json:"config,omitempty"`

	// SPIFFE human attribution (D025 format) - optional
	// When provided, nanofuse auto-registers a SPIRE workload entry
	// SPIFFE ID format: spiffe://poley.dev/g/{GroupID}/u/{OwnerUserID}/w/microvm/{vm-id}
	OwnerUserID        string `json:"owner_user_id,omitempty"`        // Human owner for SPIFFE ID /u/ segment
	GroupID            string `json:"group_id,omitempty"`             // Group for SPIFFE ID /g/ segment
	AutoRegisterSPIFFE *bool  `json:"auto_register_spiffe,omitempty"` // Auto-register SPIRE workload entry (default: true if owner/group provided)
}

// VMConfigRequest represents VM config in create request
type VMConfigRequest struct {
	VCPUs        *int                  `json:"vcpus,omitempty"`
	MemoryMiB    *int                  `json:"memory_mib,omitempty"`
	KernelArgs   *string               `json:"kernel_args,omitempty"`
	SSHPublicKey *string               `json:"ssh_public_key,omitempty"` // Base64-encoded SSH public key
	Network      *NetworkConfigRequest `json:"network,omitempty"`
}

// NetworkConfigRequest represents network config in create request
type NetworkConfigRequest struct {
	Mode         *string        `json:"mode,omitempty"`
	BridgeName   *string        `json:"bridge_name,omitempty"`
	MACAddress   *string        `json:"mac_address,omitempty"`
	PortForwards *[]PortForward `json:"port_forwards,omitempty"`
	EgressPolicy *EgressPolicy  `json:"egress_policy,omitempty"`
}

// StopVMRequest represents a request to stop a VM
type StopVMRequest struct {
	TimeoutSeconds *int `json:"timeout_seconds,omitempty"`
}

// ResumeVMRequest represents a request to resume a VM
type ResumeVMRequest struct {
	SnapshotID *string `json:"snapshot_id,omitempty"`
}

// ListVMsResponse represents a list of VMs
type ListVMsResponse struct {
	VMs   []VMListItem `json:"vms"`
	Total int          `json:"total"`
}

// VMListItem represents a VM in list view
type VMListItem struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	State         VMState   `json:"state"`
	Image         string    `json:"image"`
	ImageDigest   string    `json:"image_digest"`
	Architecture  string    `json:"architecture"`
	Config        VMConfig  `json:"config"`
	CreatedAt     time.Time `json:"created_at"`
	UptimeSeconds *int      `json:"uptime_seconds,omitempty"`

	// SPIFFE human attribution (D025 format)
	OwnerUserID string `json:"owner_user_id,omitempty"`
	GroupID     string `json:"group_id,omitempty"`
	SpiffeID    string `json:"spiffe_id,omitempty"`
}
