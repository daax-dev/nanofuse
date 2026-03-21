package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/jpoley/nanofuse/internal/types"
)

// handleListSnapshots lists snapshots for a VM
func (s *Server) handleListSnapshots(w http.ResponseWriter, r *http.Request, vmID string) {
	// Verify VM exists
	vm, err := s.db.GetVM(vmID)
	if err != nil {
		s.logger.Printf("ERROR: Failed to get VM: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, "Failed to get VM", nil)
		return
	}

	if vm == nil {
		types.WriteError(w, http.StatusNotFound, types.ErrVMNotFound,
			fmt.Sprintf("Virtual machine with ID '%s' does not exist", vmID),
			map[string]interface{}{"vm_id": vmID})
		return
	}

	// List snapshots
	snapshots, err := s.db.ListSnapshots(vmID)
	if err != nil {
		s.logger.Printf("ERROR: Failed to list snapshots: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, "Failed to list snapshots", nil)
		return
	}

	response := types.ListSnapshotsResponse{
		Snapshots: make([]types.Snapshot, 0, len(snapshots)),
		Total:     len(snapshots),
	}

	for _, snap := range snapshots {
		response.Snapshots = append(response.Snapshots, *snap)
	}

	writeJSON(w, http.StatusOK, response)
}

// handleCreateSnapshot creates a snapshot
func (s *Server) handleCreateSnapshot(w http.ResponseWriter, r *http.Request, vmID string) {
	var req types.CreateSnapshotRequest
	if r.ContentLength > 0 {
		if err := readJSON(r, &req); err != nil {
			types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest, "Invalid request body", nil)
			return
		}
	}

	// Get VM
	vm, err := s.db.GetVM(vmID)
	if err != nil {
		s.logger.Printf("ERROR: Failed to get VM: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, "Failed to get VM", nil)
		return
	}

	if vm == nil {
		types.WriteError(w, http.StatusNotFound, types.ErrVMNotFound,
			fmt.Sprintf("Virtual machine with ID '%s' does not exist", vmID),
			map[string]interface{}{"vm_id": vmID})
		return
	}

	// Validate VM state
	if vm.State != types.StateRunning && vm.State != types.StatePaused {
		types.WriteError(w, http.StatusConflict, types.ErrInvalidStateTransition,
			fmt.Sprintf("Cannot snapshot VM in state '%s'", vm.State),
			map[string]interface{}{"current_state": vm.State})
		return
	}

	// Acquire lock
	if err := s.db.AcquireLock(vm.ID, "snapshot"); err != nil {
		types.WriteError(w, http.StatusConflict, types.ErrVMLocked, "VM is locked by another operation", nil)
		return
	}
	defer func() {
		if err := s.db.ReleaseLock(vm.ID); err != nil {
			s.logger.Printf("WARN: Failed to release lock: %v", err)
		}
	}()

	// Generate snapshot ID
	snapshotID := fmt.Sprintf("snapshot-%s", time.Now().Format("20060102-150405"))
	if req.Name == "" {
		req.Name = snapshotID
	}

	// Create snapshot directory
	snapshotDir := filepath.Join(s.config.Storage.DataDir, "snapshots", vm.ID, snapshotID)
	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		s.logger.Printf("ERROR: Failed to create snapshot directory: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, "Failed to create snapshot directory", nil)
		return
	}

	memPath := filepath.Join(snapshotDir, "mem.snap")
	snapPath := filepath.Join(snapshotDir, "vm.snap")

	// Create snapshot via Firecracker
	if err := s.fcManager.CreateSnapshot(vm, snapPath, memPath); err != nil {
		s.logger.Printf("ERROR: Failed to create snapshot: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError,
			fmt.Sprintf("Failed to create snapshot: %v", err), nil)
		return
	}

	// Calculate snapshot size
	var totalSize int64
	if info, err := os.Stat(memPath); err == nil {
		totalSize += info.Size()
	}
	if info, err := os.Stat(snapPath); err == nil {
		totalSize += info.Size()
	}

	// Create snapshot record
	snapshot := &types.Snapshot{
		ID:               snapshotID,
		VMID:             vm.ID,
		Name:             req.Name,
		MemoryFilePath:   memPath,
		SnapshotFilePath: snapPath,
		SizeBytes:        totalSize,
		CreatedAt:        time.Now(),
	}

	if err := s.db.CreateSnapshot(snapshot); err != nil {
		s.logger.Printf("ERROR: Failed to save snapshot: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, "Failed to save snapshot", nil)
		return
	}

	s.logger.Printf("INFO: Created snapshot: %s for VM %s", snapshotID, vm.Name)
	writeJSON(w, http.StatusCreated, snapshot)
}

// handleGetSnapshot gets a specific snapshot
func (s *Server) handleGetSnapshot(w http.ResponseWriter, r *http.Request, snapshotID string) {
	snapshot, err := s.db.GetSnapshot(snapshotID)
	if err != nil {
		s.logger.Printf("ERROR: Failed to get snapshot: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, "Failed to get snapshot", nil)
		return
	}

	if snapshot == nil {
		types.WriteError(w, http.StatusNotFound, types.ErrSnapshotNotFound,
			fmt.Sprintf("Snapshot '%s' not found", snapshotID),
			map[string]interface{}{"snapshot_id": snapshotID})
		return
	}

	writeJSON(w, http.StatusOK, snapshot)
}

// handleDeleteSnapshot deletes a snapshot
func (s *Server) handleDeleteSnapshot(w http.ResponseWriter, r *http.Request, snapshotID string) {
	snapshot, err := s.db.GetSnapshot(snapshotID)
	if err != nil {
		s.logger.Printf("ERROR: Failed to get snapshot: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, "Failed to get snapshot", nil)
		return
	}

	if snapshot == nil {
		types.WriteError(w, http.StatusNotFound, types.ErrSnapshotNotFound,
			fmt.Sprintf("Snapshot '%s' not found", snapshotID),
			map[string]interface{}{"snapshot_id": snapshotID})
		return
	}

	// Delete snapshot files
	snapshotDir := filepath.Dir(snapshot.MemoryFilePath)
	if err := os.RemoveAll(snapshotDir); err != nil {
		s.logger.Printf("WARN: Failed to delete snapshot files: %v", err)
	}

	// Delete from database
	if err := s.db.DeleteSnapshot(snapshotID); err != nil {
		s.logger.Printf("ERROR: Failed to delete snapshot: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, "Failed to delete snapshot", nil)
		return
	}

	s.logger.Printf("INFO: Deleted snapshot: %s", snapshotID)
	w.WriteHeader(http.StatusNoContent)
}

// ============================================================================
// Go 1.22+ Path Parameter Wrappers
// ============================================================================

// handleListSnapshotsByPath handles GET /vms/{id}/snapshots using path parameters
func (s *Server) handleListSnapshotsByPath(w http.ResponseWriter, r *http.Request) {
	vmID := r.PathValue("id")
	if vmID == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest, "VM ID is required", nil)
		return
	}
	s.handleListSnapshots(w, r, vmID)
}

// handleCreateSnapshotByPath handles POST /vms/{id}/snapshots using path parameters
func (s *Server) handleCreateSnapshotByPath(w http.ResponseWriter, r *http.Request) {
	vmID := r.PathValue("id")
	if vmID == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest, "VM ID is required", nil)
		return
	}
	s.handleCreateSnapshot(w, r, vmID)
}

// handleGetSnapshotByPath handles GET /snapshots/{id} using path parameters
func (s *Server) handleGetSnapshotByPath(w http.ResponseWriter, r *http.Request) {
	snapshotID := r.PathValue("id")
	if snapshotID == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest, "Snapshot ID is required", nil)
		return
	}
	s.handleGetSnapshot(w, r, snapshotID)
}

// handleDeleteSnapshotByPath handles DELETE /snapshots/{id} using path parameters
func (s *Server) handleDeleteSnapshotByPath(w http.ResponseWriter, r *http.Request) {
	snapshotID := r.PathValue("id")
	if snapshotID == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest, "Snapshot ID is required", nil)
		return
	}
	s.handleDeleteSnapshot(w, r, snapshotID)
}
