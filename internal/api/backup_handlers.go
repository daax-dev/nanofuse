package api

import (
	"fmt"
	"net/http"

	"github.com/daax-dev/nanofuse/internal/types"
)

// handleListBackups lists backups for a VM
func (s *Server) handleListBackups(w http.ResponseWriter, r *http.Request, vmID string) {
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

	// TODO: Implement S3 backup listing
	s.logger.Printf("WARN: Backup listing not yet implemented for VM %s", vm.Name)
	types.WriteError(w, http.StatusNotImplemented, types.ErrInternalError, "Backup listing not yet implemented", nil)
}

// handleCreateBackup creates a backup
func (s *Server) handleCreateBackup(w http.ResponseWriter, r *http.Request, vmID string) {
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

	// Validate VM state - can only backup running or paused VMs
	if vm.State != types.StateRunning && vm.State != types.StatePaused {
		types.WriteError(w, http.StatusConflict, types.ErrInvalidStateTransition,
			fmt.Sprintf("Cannot backup VM in state '%s'", vm.State),
			map[string]interface{}{"current_state": vm.State})
		return
	}

	// TODO: Implement S3 backup creation
	// 1. Create snapshot
	// 2. Compress snapshot files
	// 3. Upload to S3
	// 4. Return backup job status
	s.logger.Printf("WARN: Backup creation not yet implemented for VM %s", vm.Name)
	types.WriteError(w, http.StatusNotImplemented, types.ErrInternalError, "Backup creation not yet implemented", nil)
}

// handleGetBackup gets a specific backup
func (s *Server) handleGetBackup(w http.ResponseWriter, r *http.Request, backupID string) {
	// TODO: Implement S3 backup retrieval
	s.logger.Printf("WARN: Get backup not yet implemented: %s", backupID)
	types.WriteError(w, http.StatusNotImplemented, types.ErrInternalError, "Get backup not yet implemented", nil)
}

// handleDeleteBackup deletes a backup
func (s *Server) handleDeleteBackup(w http.ResponseWriter, r *http.Request, backupID string) {
	// TODO: Implement S3 backup deletion
	s.logger.Printf("WARN: Delete backup not yet implemented: %s", backupID)
	types.WriteError(w, http.StatusNotImplemented, types.ErrInternalError, "Delete backup not yet implemented", nil)
}

// handleRestoreBackup restores a backup
func (s *Server) handleRestoreBackup(w http.ResponseWriter, r *http.Request, backupID string) {
	// TODO: Implement backup restore
	// 1. Download backup from S3
	// 2. Decompress
	// 3. Create new VM from backup
	// 4. Restore state
	// 5. Return restore job status
	s.logger.Printf("WARN: Restore backup not yet implemented: %s", backupID)
	types.WriteError(w, http.StatusNotImplemented, types.ErrInternalError, "Restore backup not yet implemented", nil)
}

// ============================================================================
// Go 1.22+ Path Parameter Wrappers
// ============================================================================

// handleListBackupsByPath handles GET /vms/{id}/backups using path parameters
func (s *Server) handleListBackupsByPath(w http.ResponseWriter, r *http.Request) {
	vmID := r.PathValue("id")
	if vmID == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest, "VM ID is required", nil)
		return
	}
	s.handleListBackups(w, r, vmID)
}

// handleCreateBackupByPath handles POST /vms/{id}/backups using path parameters
func (s *Server) handleCreateBackupByPath(w http.ResponseWriter, r *http.Request) {
	vmID := r.PathValue("id")
	if vmID == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest, "VM ID is required", nil)
		return
	}
	s.handleCreateBackup(w, r, vmID)
}

// handleGetBackupByPath handles GET /backups/{id} using path parameters
func (s *Server) handleGetBackupByPath(w http.ResponseWriter, r *http.Request) {
	backupID := r.PathValue("id")
	if backupID == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest, "Backup ID is required", nil)
		return
	}
	s.handleGetBackup(w, r, backupID)
}

// handleDeleteBackupByPath handles DELETE /backups/{id} using path parameters
func (s *Server) handleDeleteBackupByPath(w http.ResponseWriter, r *http.Request) {
	backupID := r.PathValue("id")
	if backupID == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest, "Backup ID is required", nil)
		return
	}
	s.handleDeleteBackup(w, r, backupID)
}

// handleRestoreBackupByPath handles POST /backups/{id}/restore using path parameters
func (s *Server) handleRestoreBackupByPath(w http.ResponseWriter, r *http.Request) {
	backupID := r.PathValue("id")
	if backupID == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest, "Backup ID is required", nil)
		return
	}
	s.handleRestoreBackup(w, r, backupID)
}
