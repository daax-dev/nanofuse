package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"
)

// Client represents the API client
type Client struct {
	httpClient *http.Client
	baseURL    string
	debug      bool
}

// NewClient creates a new API client
func NewClient(socketPath string, timeout time.Duration, debug bool) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
					var d net.Dialer
					return d.DialContext(ctx, "unix", socketPath)
				},
			},
		},
		baseURL: "http://unix",
		debug:   debug,
	}
}

// NewTCPClient creates a new TCP-based API client
func NewTCPClient(url string, timeout time.Duration, debug bool) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL: url,
		debug:   debug,
	}
}

// Health checks API health
func (c *Client) Health(ctx context.Context) (*HealthResponse, error) {
	var resp HealthResponse
	if err := c.get(ctx, "/health", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Capabilities returns daemon runtime capabilities.
func (c *Client) Capabilities(ctx context.Context) (*CapabilitiesResponse, error) {
	var resp CapabilitiesResponse
	if err := c.get(ctx, "/capabilities", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateVM creates a new VM
func (c *Client) CreateVM(ctx context.Context, req *CreateVMRequest) (*VM, error) {
	var vm VM
	if err := c.post(ctx, "/vms", req, &vm); err != nil {
		return nil, err
	}
	return &vm, nil
}

// ListVMs lists all VMs
func (c *Client) ListVMs(ctx context.Context, state string) (*ListVMsResponse, error) {
	path := "/vms"
	if state != "" {
		path += "?state=" + state
	}
	var resp ListVMsResponse
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetVM gets VM details
func (c *Client) GetVM(ctx context.Context, id string) (*VM, error) {
	var vm VM
	if err := c.get(ctx, "/vms/"+id, &vm); err != nil {
		return nil, err
	}
	return &vm, nil
}

// DeleteVM deletes a VM
func (c *Client) DeleteVM(ctx context.Context, id string) error {
	return c.delete(ctx, "/vms/"+id)
}

// StartVM starts a VM
func (c *Client) StartVM(ctx context.Context, id string) (*VM, error) {
	var vm VM
	if err := c.post(ctx, "/vms/"+id+"/start", nil, &vm); err != nil {
		return nil, err
	}
	return &vm, nil
}

// StopVM stops a VM
func (c *Client) StopVM(ctx context.Context, id string, timeoutSeconds int) (*VM, error) {
	req := &StopRequest{TimeoutSeconds: timeoutSeconds}
	var vm VM
	if err := c.post(ctx, "/vms/"+id+"/stop", req, &vm); err != nil {
		return nil, err
	}
	return &vm, nil
}

// KillVM force kills a VM
func (c *Client) KillVM(ctx context.Context, id string) (*VM, error) {
	var vm VM
	if err := c.post(ctx, "/vms/"+id+"/kill", nil, &vm); err != nil {
		return nil, err
	}
	return &vm, nil
}

// PauseVM pauses a VM
func (c *Client) PauseVM(ctx context.Context, id string) (*VM, error) {
	var vm VM
	if err := c.post(ctx, "/vms/"+id+"/pause", nil, &vm); err != nil {
		return nil, err
	}
	return &vm, nil
}

// ResumeVM resumes a VM
func (c *Client) ResumeVM(ctx context.Context, id string, snapshotID string) (*VM, error) {
	req := &ResumeRequest{SnapshotID: snapshotID}
	var vm VM
	if err := c.post(ctx, "/vms/"+id+"/resume", req, &vm); err != nil {
		return nil, err
	}
	return &vm, nil
}

// GetVMLogs gets VM console logs
func (c *Client) GetVMLogs(ctx context.Context, id string, tail int) (string, error) {
	path := "/vms/" + id + "/logs"
	if tail > 0 {
		path += fmt.Sprintf("?tail=%d", tail)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+path, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", c.handleError(resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	return string(body), nil
}

// ExecVM executes a command in a running VM when the runtime supports it.
func (c *Client) ExecVM(ctx context.Context, id string, req *VMExecRequest) (*VMExecResult, error) {
	var result VMExecResult
	if err := c.post(ctx, "/vms/"+id+"/exec", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateSnapshot creates a VM snapshot
func (c *Client) CreateSnapshot(ctx context.Context, vmID string, req *CreateSnapshotRequest) (*Snapshot, error) {
	var snapshot Snapshot
	if err := c.post(ctx, "/vms/"+vmID+"/snapshots", req, &snapshot); err != nil {
		return nil, err
	}
	return &snapshot, nil
}

// ListSnapshots lists VM snapshots
func (c *Client) ListSnapshots(ctx context.Context, vmID string) (*ListSnapshotsResponse, error) {
	var resp ListSnapshotsResponse
	if err := c.get(ctx, "/vms/"+vmID+"/snapshots", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetSnapshot gets snapshot details
func (c *Client) GetSnapshot(ctx context.Context, snapshotID string) (*Snapshot, error) {
	var snapshot Snapshot
	if err := c.get(ctx, "/snapshots/"+snapshotID, &snapshot); err != nil {
		return nil, err
	}
	return &snapshot, nil
}

// DeleteSnapshot deletes a snapshot
func (c *Client) DeleteSnapshot(ctx context.Context, snapshotID string) error {
	return c.delete(ctx, "/snapshots/"+snapshotID)
}

// ListImages lists cached images
func (c *Client) ListImages(ctx context.Context) (*ListImagesResponse, error) {
	var resp ListImagesResponse
	if err := c.get(ctx, "/images", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetImage gets image details
func (c *Client) GetImage(ctx context.Context, digest string) (*Image, error) {
	var image Image
	if err := c.get(ctx, "/images/"+digest, &image); err != nil {
		return nil, err
	}
	return &image, nil
}

// DeleteImage deletes an image
func (c *Client) DeleteImage(ctx context.Context, digest string) error {
	return c.delete(ctx, "/images/"+digest)
}

// PullImage initiates an image pull
func (c *Client) PullImage(ctx context.Context, imageRef string) (*ImagePullJob, error) {
	req := &PullImageRequest{ImageRef: imageRef}
	var resp PullImageResponse
	if err := c.postWithStatus(ctx, "/images/pull", req, &resp, http.StatusAccepted); err != nil {
		return nil, err
	}
	// Convert PullImageResponse to ImagePullJob
	return &ImagePullJob{
		ID:        resp.JobID,
		ImageRef:  resp.ImageRef,
		State:     resp.State,
		CreatedAt: time.Now(), // Approximate - we don't get this from initial response
	}, nil
}

// GetPullJob gets image pull job status
func (c *Client) GetPullJob(ctx context.Context, jobID string) (*ImagePullJob, error) {
	var job ImagePullJob
	if err := c.get(ctx, "/images/jobs/"+jobID, &job); err != nil {
		return nil, err
	}
	return &job, nil
}

// Helper methods

func (c *Client) get(ctx context.Context, path string, result interface{}) error {
	if c.debug {
		fmt.Fprintf(os.Stderr, "DEBUG: GET %s\n", path)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", "nanofuse-cli/0.1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if c.debug {
		fmt.Fprintf(os.Stderr, "DEBUG: Response: %d %s\n", resp.StatusCode, resp.Status)
	}

	if resp.StatusCode != http.StatusOK {
		return c.handleError(resp)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

func (c *Client) post(ctx context.Context, path string, body, result interface{}) error {
	return c.postWithStatus(ctx, path, body, result, http.StatusOK)
}

func (c *Client) postWithStatus(ctx context.Context, path string, body, result interface{}, expectedStatus int) error {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewReader(data)

		if c.debug {
			fmt.Fprintf(os.Stderr, "DEBUG: POST %s\n", path)
			fmt.Fprintf(os.Stderr, "DEBUG: Request: %s\n", string(data))
		}
	} else {
		if c.debug {
			fmt.Fprintf(os.Stderr, "DEBUG: POST %s (no body)\n", path)
		}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "nanofuse-cli/0.1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if c.debug {
		fmt.Fprintf(os.Stderr, "DEBUG: Response: %d %s\n", resp.StatusCode, resp.Status)
	}

	if resp.StatusCode != expectedStatus && resp.StatusCode != http.StatusCreated {
		return c.handleError(resp)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}

func (c *Client) delete(ctx context.Context, path string) error {
	if c.debug {
		fmt.Fprintf(os.Stderr, "DEBUG: DELETE %s\n", path)
	}

	req, err := http.NewRequestWithContext(ctx, "DELETE", c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", "nanofuse-cli/0.1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if c.debug {
		fmt.Fprintf(os.Stderr, "DEBUG: Response: %d %s\n", resp.StatusCode, resp.Status)
	}

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return c.handleError(resp)
	}

	return nil
}

func (c *Client) handleError(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("HTTP %d (failed to read error body)", resp.StatusCode)
	}

	if c.debug {
		fmt.Fprintf(os.Stderr, "DEBUG: Error body: %s\n", string(body))
	}

	var apiErr APIError
	if err := json.Unmarshal(body, &apiErr); err != nil {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return &ClientError{
		StatusCode: resp.StatusCode,
		Code:       apiErr.Error.Code,
		Message:    apiErr.Error.Message,
		Details:    apiErr.Error.Details,
	}
}

// ClientError represents a structured API error
type ClientError struct {
	StatusCode int
	Code       string
	Message    string
	Details    map[string]interface{}
}

func (e *ClientError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// ExitCode maps error codes to CLI exit codes
func (e *ClientError) ExitCode() int {
	switch e.StatusCode {
	case 400, 422:
		return 2 // Validation error
	case 404:
		return 4 // Resource not found
	case 409:
		return 5 // Operation conflict
	case 503:
		return 3 // API unreachable
	default:
		return 1 // General error
	}
}
