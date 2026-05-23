package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/daax-dev/nanofuse/internal/logging"
	"github.com/daax-dev/nanofuse/internal/recording"
)

// mockRecordingStorage implements RecordingStorageInterface for testing
type mockRecordingStorage struct {
	sessions    map[string]*recording.SessionMetadata
	listError   error
	getError    error
	finalizeErr error
	deleteErr   error
}

func newMockRecordingStorage() *mockRecordingStorage {
	return &mockRecordingStorage{
		sessions: make(map[string]*recording.SessionMetadata),
	}
}

func (m *mockRecordingStorage) ListSessions(vmID string) ([]*recording.SessionMetadata, error) {
	if m.listError != nil {
		return nil, m.listError
	}

	var result []*recording.SessionMetadata
	for _, s := range m.sessions {
		if vmID == "" || s.VMID == vmID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (m *mockRecordingStorage) GetSession(sessionID string) (*recording.SessionMetadata, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	return m.sessions[sessionID], nil
}

func (m *mockRecordingStorage) Finalize(ctx context.Context, sessionID string) error {
	if m.finalizeErr != nil {
		return m.finalizeErr
	}
	if s, ok := m.sessions[sessionID]; ok {
		s.Status = "completed"
		s.EndedAt = time.Now()
		s.Compressed = true
	}
	return nil
}

func (m *mockRecordingStorage) DeleteSession(sessionID string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.sessions, sessionID)
	return nil
}

func (m *mockRecordingStorage) addSession(s *recording.SessionMetadata) {
	m.sessions[s.ID] = s
}

func createTestServer(storage RecordingStorageInterface) *Server {
	logger, _ := logging.New(logging.Config{Level: "error"})
	return &Server{
		startTime:        time.Now(),
		logger:           logger,
		recordingStorage: storage,
	}
}

func TestHandleListRecordings(t *testing.T) {
	storage := newMockRecordingStorage()
	storage.addSession(&recording.SessionMetadata{
		ID:         "session-1",
		VMID:       "vm-1",
		StartedAt:  time.Now().Add(-time.Hour),
		EventCount: 100,
		SizeBytes:  5000,
		Status:     "active",
	})
	storage.addSession(&recording.SessionMetadata{
		ID:         "session-2",
		VMID:       "vm-1",
		StartedAt:  time.Now().Add(-2 * time.Hour),
		EndedAt:    time.Now().Add(-time.Hour),
		EventCount: 200,
		SizeBytes:  10000,
		Status:     "completed",
		Compressed: true,
	})

	server := createTestServer(storage)

	req := httptest.NewRequest(http.MethodGet, "/recordings", nil)
	w := httptest.NewRecorder()

	server.handleListRecordings(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response ListRecordingsResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Total != 2 {
		t.Errorf("Expected 2 recordings, got %d", response.Total)
	}
}

func TestHandleListRecordingsFilterByVM(t *testing.T) {
	storage := newMockRecordingStorage()
	storage.addSession(&recording.SessionMetadata{
		ID:     "session-1",
		VMID:   "vm-1",
		Status: "active",
	})
	storage.addSession(&recording.SessionMetadata{
		ID:     "session-2",
		VMID:   "vm-2",
		Status: "active",
	})

	server := createTestServer(storage)

	req := httptest.NewRequest(http.MethodGet, "/recordings?vm_id=vm-1", nil)
	w := httptest.NewRecorder()

	server.handleListRecordings(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response ListRecordingsResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Total != 1 {
		t.Errorf("Expected 1 recording for vm-1, got %d", response.Total)
	}
}

func TestHandleListRecordingsNoStorage(t *testing.T) {
	server := createTestServer(nil)

	req := httptest.NewRequest(http.MethodGet, "/recordings", nil)
	w := httptest.NewRecorder()

	server.handleListRecordings(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", w.Code)
	}
}

func TestHandleGetRecording(t *testing.T) {
	storage := newMockRecordingStorage()
	storage.addSession(&recording.SessionMetadata{
		ID:         "session-123",
		VMID:       "vm-456",
		StartedAt:  time.Now().Add(-time.Hour),
		EventCount: 150,
		SizeBytes:  7500,
		Status:     "active",
	})

	server := createTestServer(storage)
	mux := setupHTTPRouter(server)

	req := httptest.NewRequest(http.MethodGet, "/recordings/session-123", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response RecordingSessionResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.ID != "session-123" {
		t.Errorf("Expected ID 'session-123', got '%s'", response.ID)
	}
	if response.VMID != "vm-456" {
		t.Errorf("Expected VMID 'vm-456', got '%s'", response.VMID)
	}
	if response.EventCount != 150 {
		t.Errorf("Expected EventCount 150, got %d", response.EventCount)
	}
}

func TestHandleGetRecordingNotFound(t *testing.T) {
	storage := newMockRecordingStorage()
	server := createTestServer(storage)
	mux := setupHTTPRouter(server)

	req := httptest.NewRequest(http.MethodGet, "/recordings/nonexistent", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestHandleGetRecordingEvents(t *testing.T) {
	storage := newMockRecordingStorage()
	storage.addSession(&recording.SessionMetadata{
		ID:         "session-events",
		VMID:       "vm-1",
		EventCount: 500,
		Status:     "completed",
	})

	server := createTestServer(storage)
	mux := setupHTTPRouter(server)

	req := httptest.NewRequest(http.MethodGet, "/recordings/session-events/events?offset=0&limit=50", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response ListEventsResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Total != 500 {
		t.Errorf("Expected Total 500, got %d", response.Total)
	}
	if response.Offset != 0 {
		t.Errorf("Expected Offset 0, got %d", response.Offset)
	}
	if response.Limit != 50 {
		t.Errorf("Expected Limit 50, got %d", response.Limit)
	}
}

func TestHandleFinalizeRecording(t *testing.T) {
	storage := newMockRecordingStorage()
	storage.addSession(&recording.SessionMetadata{
		ID:     "session-finalize",
		VMID:   "vm-1",
		Status: "active",
	})

	server := createTestServer(storage)
	mux := setupHTTPRouter(server)

	req := httptest.NewRequest(http.MethodPost, "/recordings/session-finalize/finalize", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response RecordingSessionResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != "completed" {
		t.Errorf("Expected Status 'completed', got '%s'", response.Status)
	}
	if !response.Compressed {
		t.Error("Expected Compressed to be true")
	}
}

func TestHandleFinalizeRecordingAlreadyFinalized(t *testing.T) {
	storage := newMockRecordingStorage()
	storage.addSession(&recording.SessionMetadata{
		ID:     "session-done",
		VMID:   "vm-1",
		Status: "completed", // Already finalized
	})

	server := createTestServer(storage)
	mux := setupHTTPRouter(server)

	req := httptest.NewRequest(http.MethodPost, "/recordings/session-done/finalize", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status 409 (Conflict), got %d", w.Code)
	}
}

func TestHandleDeleteRecording(t *testing.T) {
	storage := newMockRecordingStorage()
	storage.addSession(&recording.SessionMetadata{
		ID:     "session-delete",
		VMID:   "vm-1",
		Status: "completed",
	})

	server := createTestServer(storage)
	mux := setupHTTPRouter(server)

	req := httptest.NewRequest(http.MethodDelete, "/recordings/session-delete", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}

	// Verify session is deleted
	if _, exists := storage.sessions["session-delete"]; exists {
		t.Error("Session should have been deleted")
	}
}

func TestHandleDeleteRecordingNotFound(t *testing.T) {
	storage := newMockRecordingStorage()
	server := createTestServer(storage)
	mux := setupHTTPRouter(server)

	req := httptest.NewRequest(http.MethodDelete, "/recordings/nonexistent", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestRecordingRoutesRegistered(t *testing.T) {
	server := createTestServer(newMockRecordingStorage())
	mux := setupHTTPRouter(server)

	routes := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/recordings"},
		{http.MethodGet, "/recordings/test-id"},
		{http.MethodGet, "/recordings/test-id/events"},
		{http.MethodPost, "/recordings/test-id/finalize"},
		{http.MethodDelete, "/recordings/test-id"},
	}

	for _, route := range routes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			// Should not be 404 (route exists) - might be 404 for "not found" resource
			// but that's different from 404 for "route not found"
			if w.Code == http.StatusMethodNotAllowed {
				t.Errorf("Route %s %s not registered", route.method, route.path)
			}
		})
	}
}
