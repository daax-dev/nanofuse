// Package spire provides SPIRE workload registration integration for nanofuse.
// It enables automatic SPIFFE identity creation for microVMs.
package spire

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/daax-dev/nanofuse/internal/config"
)

// Service provides SPIRE workload registration operations.
type Service struct {
	cfg     *config.SPIREConfig
	enabled bool
}

// NewService creates a new SPIRE service.
func NewService(cfg *config.SPIREConfig) *Service {
	return &Service{
		cfg:     cfg,
		enabled: cfg.Enabled,
	}
}

// IsEnabled returns true if SPIRE integration is enabled.
func (s *Service) IsEnabled() bool {
	return s.enabled
}

// WorkloadEntry represents a SPIRE workload entry.
type WorkloadEntry struct {
	SpiffeID    string
	ParentID    string
	Selectors   []string
	TTL         int
	OwnerUserID string
	GroupID     string
	WorkloadID  string
}

// BuildSPIFFEID builds a D025-compliant SPIFFE ID for a microVM.
// Format: spiffe://poley.dev/g/{group}/u/{user}/w/microvm/{vm-id}
func (s *Service) BuildSPIFFEID(groupID, ownerUserID, vmID string) string {
	workloadType := s.cfg.WorkloadType
	if workloadType == "" {
		workloadType = "microvm"
	}
	return fmt.Sprintf("spiffe://%s/g/%s/u/%s/w/%s/%s",
		s.cfg.TrustDomain, groupID, ownerUserID, workloadType, vmID)
}

// safeIDPattern validates identifiers contain only safe characters (alphanumeric, hyphen, underscore)
var safeIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// ValidateIdentityParams validates the owner and group parameters.
func (s *Service) ValidateIdentityParams(groupID, ownerUserID string) error {
	if groupID == "" {
		return fmt.Errorf("group_id is required for SPIFFE identity")
	}
	if !safeIDPattern.MatchString(groupID) {
		return fmt.Errorf("group_id contains invalid characters: %s", groupID)
	}

	if ownerUserID == "" {
		return fmt.Errorf("owner_user_id is required for SPIFFE identity")
	}
	if !safeIDPattern.MatchString(ownerUserID) {
		return fmt.Errorf("owner_user_id contains invalid characters: %s", ownerUserID)
	}

	return nil
}

// validateVMID validates a VM identifier to prevent command injection via selectors.
func validateVMID(vmID string) error {
	if vmID == "" {
		return fmt.Errorf("vm_id is required")
	}
	if !safeIDPattern.MatchString(vmID) {
		return fmt.Errorf("vm_id contains invalid characters: %s", vmID)
	}
	return nil
}

// RegisterWorkload creates a SPIRE workload entry for a microVM.
// This uses the spire-server CLI via docker exec for now.
// TODO: Consider using SPIRE Go API directly for better performance.
func (s *Service) RegisterWorkload(ctx context.Context, entry *WorkloadEntry) error {
	if !s.enabled {
		return nil
	}

	slog.Info("Creating SPIRE workload entry",
		slog.String("spiffe_id", entry.SpiffeID),
		slog.String("owner", entry.OwnerUserID),
		slog.String("group", entry.GroupID),
	)

	// Get container name from config (defaults to "spire-server")
	containerName := s.cfg.ContainerName
	if containerName == "" {
		containerName = "spire-server"
	}

	// Build the command arguments as a slice (prevents shell injection)
	cmdArgs := []string{
		"exec", containerName,
		"/opt/spire/bin/spire-server", "entry", "create",
		"-spiffeID", entry.SpiffeID,
		"-parentID", entry.ParentID,
		"-ttl", fmt.Sprintf("%d", entry.TTL),
	}

	// Add selectors
	for _, sel := range entry.Selectors {
		cmdArgs = append(cmdArgs, "-selector", sel)
	}

	// Execute docker command
	cmd := exec.CommandContext(ctx, "docker", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if entry already exists (idempotent)
		if strings.Contains(string(output), "already exists") {
			slog.Debug("SPIRE entry already exists", slog.String("spiffe_id", entry.SpiffeID))
			return nil
		}
		return fmt.Errorf("failed to create SPIRE entry: %w: %s", err, string(output))
	}

	slog.Info("SPIRE workload entry created", slog.String("spiffe_id", entry.SpiffeID))
	return nil
}

// UnregisterWorkload deletes a SPIRE workload entry.
func (s *Service) UnregisterWorkload(ctx context.Context, spiffeID string) error {
	if !s.enabled {
		return nil
	}

	slog.Info("Deleting SPIRE workload entry", slog.String("spiffe_id", spiffeID))

	// Get container name from config (defaults to "spire-server")
	containerName := s.cfg.ContainerName
	if containerName == "" {
		containerName = "spire-server"
	}

	// First, find the entry ID by SPIFFE ID
	findArgs := []string{
		"exec", containerName,
		"/opt/spire/bin/spire-server", "entry", "show",
		"-spiffeID", spiffeID,
	}

	findCmd := exec.CommandContext(ctx, "docker", findArgs...)
	output, err := findCmd.CombinedOutput()
	if err != nil {
		// Entry might not exist, which is fine
		slog.Debug("SPIRE entry not found (may not exist)", slog.String("spiffe_id", spiffeID))
		return nil
	}

	// Parse entry ID from output (format: "Entry ID         : <id>")
	entryIDPattern := regexp.MustCompile(`Entry ID\s*:\s*([a-f0-9-]+)`)
	matches := entryIDPattern.FindStringSubmatch(string(output))
	if len(matches) < 2 {
		slog.Debug("No entry ID found in output", slog.String("spiffe_id", spiffeID))
		return nil
	}
	entryID := matches[1]

	// Delete the entry
	deleteArgs := []string{
		"exec", containerName,
		"/opt/spire/bin/spire-server", "entry", "delete",
		"-entryID", entryID,
	}

	deleteCmd := exec.CommandContext(ctx, "docker", deleteArgs...)
	if _, err := deleteCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to delete SPIRE entry: %w", err)
	}

	slog.Info("SPIRE workload entry deleted", slog.String("spiffe_id", spiffeID))
	return nil
}

// CreateVMWorkloadEntry creates a complete workload entry for a VM.
func (s *Service) CreateVMWorkloadEntry(ctx context.Context, vmID, groupID, ownerUserID string) (string, error) {
	if !s.enabled {
		return "", nil
	}

	// Validate parameters (including vmID to prevent command injection via selectors)
	if err := s.ValidateIdentityParams(groupID, ownerUserID); err != nil {
		return "", err
	}
	if err := validateVMID(vmID); err != nil {
		return "", err
	}

	// Build SPIFFE ID
	spiffeID := s.BuildSPIFFEID(groupID, ownerUserID, vmID)

	// Create entry with default selectors
	entry := &WorkloadEntry{
		SpiffeID:    spiffeID,
		ParentID:    s.cfg.ParentID,
		Selectors:   []string{fmt.Sprintf("docker:label:vm_id:%s", vmID)},
		TTL:         s.cfg.DefaultTTL,
		OwnerUserID: ownerUserID,
		GroupID:     groupID,
		WorkloadID:  vmID,
	}

	// Add timeout for registration
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := s.RegisterWorkload(ctx, entry); err != nil {
		return "", err
	}

	return spiffeID, nil
}

// DeleteVMWorkloadEntry deletes the workload entry for a VM.
func (s *Service) DeleteVMWorkloadEntry(ctx context.Context, spiffeID string) error {
	if !s.enabled || spiffeID == "" {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return s.UnregisterWorkload(ctx, spiffeID)
}
