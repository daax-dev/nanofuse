package recording

import (
	"context"
	"testing"
	"time"
)

func TestNewReceiverDefaults(t *testing.T) {
	r := NewReceiver()

	if r.port != DefaultVsockPort {
		t.Errorf("port = %d, want %d", r.port, DefaultVsockPort)
	}
	if r.bufferSize != 1000 {
		t.Errorf("bufferSize = %d, want 1000", r.bufferSize)
	}
	if r.storage != nil {
		t.Error("storage should be nil by default")
	}
	if r.handler != nil {
		t.Error("handler should be nil by default")
	}
	if r.connections == nil {
		t.Error("connections map should be initialized")
	}
	if r.sessions == nil {
		t.Error("sessions map should be initialized")
	}
}

func TestReceiverWithOptions(t *testing.T) {
	mockStorage := &mockStorage{}
	mockHandler := func(ctx context.Context, vmCID uint32, events <-chan *Event) {
		// Handler for testing
	}

	r := NewReceiver(
		WithPort(100),
		WithBufferSize(500),
		WithStorage(mockStorage),
		WithHandler(mockHandler),
	)

	if r.port != 100 {
		t.Errorf("port = %d, want 100", r.port)
	}
	if r.bufferSize != 500 {
		t.Errorf("bufferSize = %d, want 500", r.bufferSize)
	}
	if r.storage != mockStorage {
		t.Error("storage not set correctly")
	}
	if r.handler == nil {
		t.Error("handler should be set")
	}
}

func TestReceiverGetSession(t *testing.T) {
	r := NewReceiver()

	// Add a session
	session := &Session{
		ID:        "test-session-123",
		VMID:      "vm-123",
		CID:       3,
		StartTime: time.Now(),
		State:     SessionStateActive,
		Events:    100,
		Bytes:     5000,
	}
	r.sessions[session.ID] = session

	// Get existing session
	got, ok := r.GetSession("test-session-123")
	if !ok {
		t.Error("GetSession returned false for existing session")
	}
	if got.ID != session.ID {
		t.Errorf("ID = %q, want %q", got.ID, session.ID)
	}
	if got.VMID != session.VMID {
		t.Errorf("VMID = %q, want %q", got.VMID, session.VMID)
	}
	if got.Events != session.Events {
		t.Errorf("Events = %d, want %d", got.Events, session.Events)
	}

	// Get non-existent session
	_, ok = r.GetSession("non-existent")
	if ok {
		t.Error("GetSession returned true for non-existent session")
	}
}

func TestReceiverGetActiveConnections(t *testing.T) {
	r := NewReceiver()

	if got := r.GetActiveConnections(); got != 0 {
		t.Errorf("GetActiveConnections() = %d, want 0", got)
	}

	// Add mock connections
	r.connections[3] = &vmConnection{cid: 3}
	r.connections[4] = &vmConnection{cid: 4}

	if got := r.GetActiveConnections(); got != 2 {
		t.Errorf("GetActiveConnections() = %d, want 2", got)
	}
}

func TestSessionStates(t *testing.T) {
	states := []SessionState{
		SessionStateActive,
		SessionStateCompleted,
		SessionStateFailed,
	}

	for _, state := range states {
		if state == "" {
			t.Errorf("SessionState should not be empty string")
		}
	}

	// Verify values
	if SessionStateActive != "active" {
		t.Error("SessionStateActive != 'active'")
	}
	if SessionStateCompleted != "completed" {
		t.Error("SessionStateCompleted != 'completed'")
	}
	if SessionStateFailed != "failed" {
		t.Error("SessionStateFailed != 'failed'")
	}
}

func TestReceiverSessionCopyIsolation(t *testing.T) {
	r := NewReceiver()

	// Add a session
	session := &Session{
		ID:     "test-isolation",
		Events: 50,
	}
	r.sessions[session.ID] = session

	// Get session copy
	got, _ := r.GetSession("test-isolation")

	// Modify the copy
	got.Events = 999

	// Original should be unchanged
	if r.sessions["test-isolation"].Events != 50 {
		t.Error("Modifying returned session affected the original")
	}
}

// mockStorage implements Storage interface for testing
type mockStorage struct {
	events    [][]*Event
	finalized []string
}

func (m *mockStorage) Write(ctx context.Context, events []*Event) error {
	m.events = append(m.events, events)
	return nil
}

func (m *mockStorage) Finalize(ctx context.Context, sessionID string) error {
	m.finalized = append(m.finalized, sessionID)
	return nil
}

func (m *mockStorage) GetPlaybackURL(sessionID string) (string, error) {
	return "https://example.com/playback/" + sessionID, nil
}

// Note: Integration tests with actual vsock would require running inside
// a VM environment or using the host kernel's vsock support.
// Those tests should be in a separate file with build tags:
// //go:build integration

func TestMockStorageInterface(t *testing.T) {
	// Verify mockStorage implements Storage
	var _ Storage = (*mockStorage)(nil)

	m := &mockStorage{}

	// Test Write
	events := []*Event{
		{VMID: "vm1", Type: EventSessionStart},
		{VMID: "vm1", Type: EventTerminalOutput},
	}
	if err := m.Write(context.Background(), events); err != nil {
		t.Errorf("Write failed: %v", err)
	}
	if len(m.events) != 1 {
		t.Errorf("events batches = %d, want 1", len(m.events))
	}

	// Test Finalize
	if err := m.Finalize(context.Background(), "session-123"); err != nil {
		t.Errorf("Finalize failed: %v", err)
	}
	if len(m.finalized) != 1 || m.finalized[0] != "session-123" {
		t.Errorf("finalized = %v, want [session-123]", m.finalized)
	}

	// Test GetPlaybackURL
	url, err := m.GetPlaybackURL("session-456")
	if err != nil {
		t.Errorf("GetPlaybackURL failed: %v", err)
	}
	if url != "https://example.com/playback/session-456" {
		t.Errorf("URL = %q, unexpected", url)
	}
}
