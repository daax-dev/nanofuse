package types

import "time"

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

// CreateSnapshotRequest represents a request to create a snapshot
type CreateSnapshotRequest struct {
	Name string `json:"name,omitempty"`
}

// ListSnapshotsResponse represents a list of snapshots
type ListSnapshotsResponse struct {
	Snapshots []Snapshot `json:"snapshots"`
	Total     int        `json:"total"`
}
