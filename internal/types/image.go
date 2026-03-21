package types

import "time"

// Image represents a cached OCI image
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

// ImagePullJob represents an async image pull operation
type ImagePullJob struct {
	ID           string        `json:"id"`
	ImageRef     string        `json:"image_ref"`
	State        PullJobState  `json:"state"`
	Progress     *PullProgress `json:"progress,omitempty"`
	Error        *string       `json:"error,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
	CompletedAt  *time.Time    `json:"completed_at,omitempty"`
	ResultDigest *string       `json:"result_digest,omitempty"`
}

// PullJobState represents the state of a pull job
type PullJobState string

const (
	PullJobPending    PullJobState = "pending"
	PullJobInProgress PullJobState = "in_progress"
	PullJobCompleted  PullJobState = "completed"
	PullJobFailed     PullJobState = "failed"
)

// PullProgress represents progress of an image pull
type PullProgress struct {
	CurrentBytes int64 `json:"current_bytes"`
	TotalBytes   int64 `json:"total_bytes"`
	Percentage   int   `json:"percentage"`
}

// PullImageRequest represents a request to pull an image
type PullImageRequest struct {
	ImageRef string `json:"image_ref"`
}

// PullImageResponse represents the response from pull request
type PullImageResponse struct {
	JobID     string       `json:"job_id"`
	ImageRef  string       `json:"image_ref"`
	State     PullJobState `json:"state"`
	StatusURL string       `json:"status_url"`
}

// ListImagesResponse represents a list of images
type ListImagesResponse struct {
	Images []Image `json:"images"`
	Total  int     `json:"total"`
}
