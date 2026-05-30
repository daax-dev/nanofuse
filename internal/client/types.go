package client

import "time"

// VM represents a virtual machine
type VM struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	State         string     `json:"state"`
	Image         string     `json:"image"`
	ImageDigest   string     `json:"image_digest"`
	Architecture  string     `json:"architecture"`
	Config        VMConfig   `json:"config"`
	Runtime       *VMRuntime `json:"runtime,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	UptimeSeconds *int       `json:"uptime_seconds,omitempty"`
	LockedBy      *string    `json:"locked_by,omitempty"`
	LockedAt      *time.Time `json:"locked_at,omitempty"`
}

// VMConfig represents VM configuration
type VMConfig struct {
	VCPUs        int           `json:"vcpus"`
	MemoryMiB    int           `json:"memory_mib"`
	KernelArgs   string        `json:"kernel_args"`
	SSHPublicKey string        `json:"ssh_public_key,omitempty"` // Base64-encoded SSH public key
	Network      NetworkConfig `json:"network"`
	Disks        []DiskConfig  `json:"disks,omitempty"`
}

// NetworkConfig represents network configuration
type NetworkConfig struct {
	Mode         string        `json:"mode"`
	TAPDevice    string        `json:"tap_device,omitempty"`
	MACAddress   string        `json:"mac_address,omitempty"`
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
type EgressPolicy struct {
	Enabled       bool         `json:"enabled"`
	DefaultAction string       `json:"default_action,omitempty"`
	AllowDNS      bool         `json:"allow_dns,omitempty"`
	ProxyOnly     bool         `json:"proxy_only,omitempty"`
	Proxy         *EgressProxy `json:"proxy,omitempty"`
	AllowRules    []EgressRule `json:"allow_rules,omitempty"`
}

// EgressProxy is the host-controlled proxy endpoint a VM may reach.
type EgressProxy struct {
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol,omitempty"`
}

// EgressRule allows one outbound L3/L4 destination.
type EgressRule struct {
	CIDR        string `json:"cidr"`
	Protocol    string `json:"protocol"`
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
	PID         int         `json:"pid"`
	SocketPath  string      `json:"socket_path"`
	ConsolePath string      `json:"console_path"`
	NetworkInfo NetworkInfo `json:"network_info,omitempty"`
}

// NetworkInfo represents runtime network information
type NetworkInfo struct {
	TAPDevice string `json:"tap_device"`
	HostIP    string `json:"host_ip"`
	GuestIP   string `json:"guest_ip"`
	Gateway   string `json:"gateway"`
}

// Image represents a cached image
type Image struct {
	Digest        string            `json:"digest"`
	Tags          []string          `json:"tags"`
	Architecture  string            `json:"architecture"`
	SizeBytes     int64             `json:"size_bytes"`
	KernelVersion string            `json:"kernel_version,omitempty"`
	RootfsPath    string            `json:"rootfs_path"`
	KernelPath    string            `json:"kernel_path"`
	PulledAt      time.Time         `json:"pulled_at"`
	Labels        map[string]string `json:"labels,omitempty"`
}

// Snapshot represents a VM snapshot
type Snapshot struct {
	ID               string    `json:"id"`
	VMID             string    `json:"vm_id"`
	Name             string    `json:"name,omitempty"`
	MemoryFilePath   string    `json:"memory_file_path"`
	SnapshotFilePath string    `json:"snapshot_file_path"`
	SizeBytes        int64     `json:"size_bytes"`
	CreatedAt        time.Time `json:"created_at"`
}

// ImagePullJob represents an image pull operation
type ImagePullJob struct {
	ID           string        `json:"id"`
	ImageRef     string        `json:"image_ref"`
	State        string        `json:"state"`
	Progress     *PullProgress `json:"progress,omitempty"`
	Error        *string       `json:"error,omitempty"`
	ResultDigest *string       `json:"result_digest,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
	CompletedAt  *time.Time    `json:"completed_at,omitempty"`
}

// PullProgress represents image pull progress
type PullProgress struct {
	CurrentBytes int64 `json:"current_bytes"`
	TotalBytes   int64 `json:"total_bytes"`
	Percentage   int   `json:"percentage"`
}

// HealthResponse represents API health status
type HealthResponse struct {
	Status        string `json:"status"`
	Version       string `json:"version"`
	UptimeSeconds int    `json:"uptime_seconds"`
}

// CapabilitiesResponse describes daemon runtime capabilities.
type CapabilitiesResponse struct {
	Status  string                   `json:"status"`
	Version string                   `json:"version"`
	Host    HostCapabilities         `json:"host"`
	Runtime RuntimeCapabilities      `json:"runtime"`
	API     APITransportCapabilities `json:"api"`
}

// HostCapabilities describes host-level platform support.
type HostCapabilities struct {
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	KVMDevice    string `json:"kvm_device"`
	KVMExists    bool   `json:"kvm_exists"`
	KVMReadWrite bool   `json:"kvm_read_write"`
	KVMError     string `json:"kvm_error,omitempty"`
}

// RuntimeCapabilities describes the microVM runtime available to nanofused.
type RuntimeCapabilities struct {
	NativeRuntime        bool   `json:"native_runtime"`
	FirecrackerBinary    string `json:"firecracker_binary"`
	FirecrackerAvailable bool   `json:"firecracker_available"`
	RootRequired         bool   `json:"root_required"`
	NetworkSetupRequired bool   `json:"network_setup_required"`
	Message              string `json:"message"`
}

// APITransportCapabilities describes how clients can reach the daemon.
type APITransportCapabilities struct {
	UnixSocket string `json:"unix_socket,omitempty"`
	TCPBind    string `json:"tcp_bind,omitempty"`
}

// APIError represents an API error response
type APIError struct {
	Error ErrorDetails `json:"error"`
}

// ErrorDetails provides error information
type ErrorDetails struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// CreateVMRequest represents VM creation request
type CreateVMRequest struct {
	Name   string   `json:"name,omitempty"`
	Image  string   `json:"image"`
	Config VMConfig `json:"config"`
}

// CreateSnapshotRequest represents snapshot creation request
type CreateSnapshotRequest struct {
	Name string `json:"name,omitempty"`
}

// ResumeRequest represents VM resume request
type ResumeRequest struct {
	SnapshotID string `json:"snapshot_id,omitempty"`
}

// StopRequest represents VM stop request
type StopRequest struct {
	TimeoutSeconds int `json:"timeout_seconds,omitempty"`
}

// PullImageRequest represents image pull request
type PullImageRequest struct {
	ImageRef string `json:"image_ref"`
}

// PullImageResponse represents the initial response from pull request
type PullImageResponse struct {
	JobID     string `json:"job_id"`
	ImageRef  string `json:"image_ref"`
	State     string `json:"state"`
	StatusURL string `json:"status_url"`
}

// ListVMsResponse represents VM list response
type ListVMsResponse struct {
	VMs   []VM `json:"vms"`
	Total int  `json:"total"`
}

// ListImagesResponse represents image list response
type ListImagesResponse struct {
	Images []Image `json:"images"`
	Total  int     `json:"total"`
}

// ListSnapshotsResponse represents snapshot list response
type ListSnapshotsResponse struct {
	Snapshots []Snapshot `json:"snapshots"`
	Total     int        `json:"total"`
}
