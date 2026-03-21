package api

import (
	"context"
	"net/http"
	"strconv"

	"github.com/jpoley/nanofuse/internal/recording"
	"github.com/jpoley/nanofuse/internal/types"
)

// RecordingSessionResponse represents a recording session in API responses
type RecordingSessionResponse struct {
	ID         string `json:"id"`
	VMID       string `json:"vm_id"`
	StartedAt  string `json:"started_at"`
	EndedAt    string `json:"ended_at,omitempty"`
	EventCount int64  `json:"event_count"`
	SizeBytes  int64  `json:"size_bytes"`
	Status     string `json:"status"`
	Compressed bool   `json:"compressed"`
}

// ListRecordingsResponse is the response for listing recordings
type ListRecordingsResponse struct {
	Recordings []RecordingSessionResponse `json:"recordings"`
	Total      int                        `json:"total"`
}

// handleListRecordings handles GET /recordings
func (s *Server) handleListRecordings(w http.ResponseWriter, r *http.Request) {
	if s.recordingStorage == nil {
		types.WriteError(w, http.StatusServiceUnavailable, types.ErrInternalError,
			"Recording storage not configured", nil)
		return
	}

	vmID := r.URL.Query().Get("vm_id")

	sessions, err := s.recordingStorage.ListSessions(vmID)
	if err != nil {
		s.logger.Error("Failed to list recordings: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError,
			"Failed to list recordings", nil)
		return
	}

	recordings := make([]RecordingSessionResponse, 0, len(sessions))
	for _, session := range sessions {
		rec := RecordingSessionResponse{
			ID:         session.ID,
			VMID:       session.VMID,
			StartedAt:  session.StartedAt.Format("2006-01-02T15:04:05Z"),
			EventCount: session.EventCount,
			SizeBytes:  session.SizeBytes,
			Status:     session.Status,
			Compressed: session.Compressed,
		}
		if !session.EndedAt.IsZero() {
			rec.EndedAt = session.EndedAt.Format("2006-01-02T15:04:05Z")
		}
		recordings = append(recordings, rec)
	}

	response := ListRecordingsResponse{
		Recordings: recordings,
		Total:      len(recordings),
	}

	writeJSON(w, http.StatusOK, response)
}

// handleGetRecording handles GET /recordings/{id}
func (s *Server) handleGetRecording(w http.ResponseWriter, r *http.Request) {
	if s.recordingStorage == nil {
		types.WriteError(w, http.StatusServiceUnavailable, types.ErrInternalError,
			"Recording storage not configured", nil)
		return
	}

	sessionID := r.PathValue("id")
	if sessionID == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest,
			"Session ID is required", nil)
		return
	}

	session, err := s.recordingStorage.GetSession(sessionID)
	if err != nil {
		s.logger.Error("Failed to get recording: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError,
			"Failed to get recording", nil)
		return
	}

	if session == nil {
		types.WriteError(w, http.StatusNotFound, types.ErrRecordingNotFound,
			"Recording not found", nil)
		return
	}

	response := RecordingSessionResponse{
		ID:         session.ID,
		VMID:       session.VMID,
		StartedAt:  session.StartedAt.Format("2006-01-02T15:04:05Z"),
		EventCount: session.EventCount,
		SizeBytes:  session.SizeBytes,
		Status:     session.Status,
		Compressed: session.Compressed,
	}
	if !session.EndedAt.IsZero() {
		response.EndedAt = session.EndedAt.Format("2006-01-02T15:04:05Z")
	}

	writeJSON(w, http.StatusOK, response)
}

// RecordingEventResponse represents a recording event in API responses
type RecordingEventResponse struct {
	Timestamp string            `json:"timestamp"`
	Type      string            `json:"type"`
	Payload   string            `json:"payload,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// ListEventsResponse is the response for listing recording events
type ListEventsResponse struct {
	Events []RecordingEventResponse `json:"events"`
	Offset int                      `json:"offset"`
	Limit  int                      `json:"limit"`
	Total  int                      `json:"total"`
}

// handleGetRecordingEvents handles GET /recordings/{id}/events
func (s *Server) handleGetRecordingEvents(w http.ResponseWriter, r *http.Request) {
	if s.recordingStorage == nil {
		types.WriteError(w, http.StatusServiceUnavailable, types.ErrInternalError,
			"Recording storage not configured", nil)
		return
	}

	sessionID := r.PathValue("id")
	if sessionID == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest,
			"Session ID is required", nil)
		return
	}

	// Parse pagination parameters
	offset := 0
	limit := 100

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 1000 {
			limit = parsed
		}
	}

	// Verify session exists
	session, err := s.recordingStorage.GetSession(sessionID)
	if err != nil {
		s.logger.Error("Failed to get recording: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError,
			"Failed to get recording", nil)
		return
	}

	if session == nil {
		types.WriteError(w, http.StatusNotFound, types.ErrRecordingNotFound,
			"Recording not found", nil)
		return
	}

	// Get playback URL and read events
	// Note: For a full implementation, we'd need to read and parse the events file
	// For now, return metadata about the session with a note that event streaming
	// would require additional implementation
	response := ListEventsResponse{
		Events: []RecordingEventResponse{},
		Offset: offset,
		Limit:  limit,
		Total:  int(session.EventCount),
	}

	// TODO: Implement event reading from storage file
	// This requires decompressing zstd files and parsing the binary format
	// For now, we return empty events with the correct total count

	writeJSON(w, http.StatusOK, response)
}

// handleFinalizeRecording handles POST /recordings/{id}/finalize
func (s *Server) handleFinalizeRecording(w http.ResponseWriter, r *http.Request) {
	if s.recordingStorage == nil {
		types.WriteError(w, http.StatusServiceUnavailable, types.ErrInternalError,
			"Recording storage not configured", nil)
		return
	}

	sessionID := r.PathValue("id")
	if sessionID == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest,
			"Session ID is required", nil)
		return
	}

	// Verify session exists and is active
	session, err := s.recordingStorage.GetSession(sessionID)
	if err != nil {
		s.logger.Error("Failed to get recording: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError,
			"Failed to get recording", nil)
		return
	}

	if session == nil {
		types.WriteError(w, http.StatusNotFound, types.ErrRecordingNotFound,
			"Recording not found", nil)
		return
	}

	if session.Status != "active" {
		types.WriteError(w, http.StatusConflict, types.ErrInvalidRequest,
			"Recording is already finalized", nil)
		return
	}

	// Finalize the session
	ctx := context.Background()
	if err := s.recordingStorage.Finalize(ctx, sessionID); err != nil {
		s.logger.Error("Failed to finalize recording: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError,
			"Failed to finalize recording", nil)
		return
	}

	// Get updated session
	session, _ = s.recordingStorage.GetSession(sessionID)

	response := RecordingSessionResponse{
		ID:         session.ID,
		VMID:       session.VMID,
		StartedAt:  session.StartedAt.Format("2006-01-02T15:04:05Z"),
		EventCount: session.EventCount,
		SizeBytes:  session.SizeBytes,
		Status:     session.Status,
		Compressed: session.Compressed,
	}
	if !session.EndedAt.IsZero() {
		response.EndedAt = session.EndedAt.Format("2006-01-02T15:04:05Z")
	}

	writeJSON(w, http.StatusOK, response)
}

// handleDeleteRecording handles DELETE /recordings/{id}
func (s *Server) handleDeleteRecording(w http.ResponseWriter, r *http.Request) {
	if s.recordingStorage == nil {
		types.WriteError(w, http.StatusServiceUnavailable, types.ErrInternalError,
			"Recording storage not configured", nil)
		return
	}

	sessionID := r.PathValue("id")
	if sessionID == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest,
			"Session ID is required", nil)
		return
	}

	// Verify session exists
	session, err := s.recordingStorage.GetSession(sessionID)
	if err != nil {
		s.logger.Error("Failed to get recording: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError,
			"Failed to get recording", nil)
		return
	}

	if session == nil {
		types.WriteError(w, http.StatusNotFound, types.ErrRecordingNotFound,
			"Recording not found", nil)
		return
	}

	// Delete the session
	if err := s.recordingStorage.DeleteSession(sessionID); err != nil {
		s.logger.Error("Failed to delete recording: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError,
			"Failed to delete recording", nil)
		return
	}

	s.logger.Info("Deleted recording session: %s", sessionID)
	w.WriteHeader(http.StatusNoContent)
}

// handleListVMRecordings handles GET /vms/{id}/recordings
func (s *Server) handleListVMRecordings(w http.ResponseWriter, r *http.Request) {
	if s.recordingStorage == nil {
		types.WriteError(w, http.StatusServiceUnavailable, types.ErrInternalError,
			"Recording storage not configured", nil)
		return
	}

	vmID := r.PathValue("id")
	if vmID == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest,
			"VM ID is required", nil)
		return
	}

	// Verify VM exists
	vm, err := s.db.GetVM(vmID)
	if err != nil {
		s.logger.Error("Failed to get VM: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError,
			"Failed to get VM", nil)
		return
	}

	if vm == nil {
		types.WriteError(w, http.StatusNotFound, types.ErrVMNotFound,
			"VM not found", nil)
		return
	}

	// List recordings for this VM
	sessions, err := s.recordingStorage.ListSessions(vm.ID)
	if err != nil {
		s.logger.Error("Failed to list recordings: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError,
			"Failed to list recordings", nil)
		return
	}

	recordings := make([]RecordingSessionResponse, 0, len(sessions))
	for _, session := range sessions {
		rec := RecordingSessionResponse{
			ID:         session.ID,
			VMID:       session.VMID,
			StartedAt:  session.StartedAt.Format("2006-01-02T15:04:05Z"),
			EventCount: session.EventCount,
			SizeBytes:  session.SizeBytes,
			Status:     session.Status,
			Compressed: session.Compressed,
		}
		if !session.EndedAt.IsZero() {
			rec.EndedAt = session.EndedAt.Format("2006-01-02T15:04:05Z")
		}
		recordings = append(recordings, rec)
	}

	response := ListRecordingsResponse{
		Recordings: recordings,
		Total:      len(recordings),
	}

	writeJSON(w, http.StatusOK, response)
}

// RecordingStorageInterface defines what the server needs from recording storage
// This matches the LocalStorage methods we need
type RecordingStorageInterface interface {
	ListSessions(vmID string) ([]*recording.SessionMetadata, error)
	GetSession(sessionID string) (*recording.SessionMetadata, error)
	Finalize(ctx context.Context, sessionID string) error
	DeleteSession(sessionID string) error
}
