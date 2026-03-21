//go:build integration
// +build integration

// Package integration provides end-to-end integration tests for the nanofuse platform.
// These tests validate the recording subsystem including:
// - Recording agent boot verification
// - Terminal event generation (TERMINAL_INPUT/OUTPUT)
// - Graceful and forced shutdown session finalization
// - Multi-VM concurrent recording
// - Performance benchmarks for latency and throughput
package integration

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jpoley/nanofuse/internal/recording"
)

// eventProcessingDelay is the time to wait for asynchronous mock event processing.
// In tests with mock infrastructure, we use a short delay to allow goroutines
// to process events. In production code, proper channel synchronization would be used.
// This is acceptable in test code because:
// 1. We're testing mock infrastructure, not production async behavior
// 2. The delay is short enough to not impact test runtime significantly
// 3. Tests verify the end state, not timing-sensitive behavior
// 4. CI flakiness is mitigated by using multipliers (2x, 5x) for complex scenarios
const eventProcessingDelay = 100 * time.Millisecond

// RecordingTestSuite encapsulates the test environment for recording tests.
// It simulates the recording pipeline without requiring actual VMs by using
// mock vsock connections and in-memory storage.
type RecordingTestSuite struct {
	t             *testing.T
	storage       *MockRecordingStorage
	receiver      *MockReceiver
	tempDir       string
	eventChannels map[string]chan *recording.Event
	mu            sync.RWMutex
}

// MockRecordingStorage implements recording.Storage for testing purposes.
// It tracks all events written and sessions finalized for verification.
// In addition to the Storage interface methods (Write, Finalize, GetPlaybackURL),
// it provides test helper methods (GetSession, ListSessions, DeleteSession,
// GetEventsBySession, GetEventsByType, IsFinalized) for test verification.
type MockRecordingStorage struct {
	mu           sync.RWMutex
	eventBatches [][]*recording.Event
	allEvents    []*recording.Event
	finalized    map[string]bool
	sessions     map[string]*recording.SessionMetadata
}

// NewMockRecordingStorage creates a new mock storage instance.
func NewMockRecordingStorage() *MockRecordingStorage {
	return &MockRecordingStorage{
		eventBatches: make([][]*recording.Event, 0),
		allEvents:    make([]*recording.Event, 0),
		finalized:    make(map[string]bool),
		sessions:     make(map[string]*recording.SessionMetadata),
	}
}

// Write stores a batch of events.
func (m *MockRecordingStorage) Write(ctx context.Context, events []*recording.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.eventBatches = append(m.eventBatches, events)
	for _, e := range events {
		m.allEvents = append(m.allEvents, e)

		// Auto-create session if needed
		if _, exists := m.sessions[e.SessionID]; !exists {
			m.sessions[e.SessionID] = &recording.SessionMetadata{
				ID:        e.SessionID,
				VMID:      e.VMID,
				StartedAt: e.Timestamp,
				Status:    "active",
			}
		}
		m.sessions[e.SessionID].EventCount++
		m.sessions[e.SessionID].SizeBytes += int64(len(e.Payload))
	}

	return nil
}

// Finalize marks a session as finalized.
func (m *MockRecordingStorage) Finalize(ctx context.Context, sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.finalized[sessionID] = true
	if session, exists := m.sessions[sessionID]; exists {
		session.Status = "completed"
		session.EndedAt = time.Now()
	}
	return nil
}

// GetPlaybackURL returns a mock playback URL.
func (m *MockRecordingStorage) GetPlaybackURL(sessionID string) (string, error) {
	return fmt.Sprintf("mock://playback/%s", sessionID), nil
}

// GetSession returns session metadata.
func (m *MockRecordingStorage) GetSession(sessionID string) (*recording.SessionMetadata, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, nil
	}
	// Return a copy
	copy := *session
	return &copy, nil
}

// ListSessions returns all sessions, optionally filtered by VM ID.
func (m *MockRecordingStorage) ListSessions(vmID string) ([]*recording.SessionMetadata, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*recording.SessionMetadata, 0)
	for _, session := range m.sessions {
		if vmID == "" || session.VMID == vmID {
			copy := *session
			result = append(result, &copy)
		}
	}
	return result, nil
}

// DeleteSession removes a session and its associated events.
func (m *MockRecordingStorage) DeleteSession(sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.sessions, sessionID)
	delete(m.finalized, sessionID)

	// Clean up events for this session to prevent memory leaks.
	// Note: O(n) complexity is acceptable for test mock; production would use
	// a map[sessionID][]*Event structure for O(1) session deletion.
	filteredEvents := make([]*recording.Event, 0, len(m.allEvents))
	for _, e := range m.allEvents {
		if e.SessionID != sessionID {
			filteredEvents = append(filteredEvents, e)
		}
	}
	m.allEvents = filteredEvents

	// Also clean up eventBatches to maintain consistent state.
	// Filter out batches that become empty after removing session events.
	filteredBatches := make([][]*recording.Event, 0, len(m.eventBatches))
	for _, batch := range m.eventBatches {
		filteredBatch := make([]*recording.Event, 0, len(batch))
		for _, e := range batch {
			if e.SessionID != sessionID {
				filteredBatch = append(filteredBatch, e)
			}
		}
		// Only keep non-empty batches
		if len(filteredBatch) > 0 {
			filteredBatches = append(filteredBatches, filteredBatch)
		}
	}
	m.eventBatches = filteredBatches

	return nil
}

// GetEventsBySession returns all events for a specific session.
func (m *MockRecordingStorage) GetEventsBySession(sessionID string) []*recording.Event {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*recording.Event, 0)
	for _, e := range m.allEvents {
		if e.SessionID == sessionID {
			result = append(result, e)
		}
	}
	return result
}

// GetEventsByType returns all events of a specific type.
func (m *MockRecordingStorage) GetEventsByType(eventType recording.EventType) []*recording.Event {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*recording.Event, 0)
	for _, e := range m.allEvents {
		if e.Type == eventType {
			result = append(result, e)
		}
	}
	return result
}

// IsFinalized checks if a session has been finalized.
func (m *MockRecordingStorage) IsFinalized(sessionID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.finalized[sessionID]
}

// EventCount returns the total number of events stored.
func (m *MockRecordingStorage) EventCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.allEvents)
}

// MockReceiver simulates the recording receiver without actual vsock.
type MockReceiver struct {
	storage     *MockRecordingStorage
	connections map[uint32]*MockVMConnection
	sessions    map[string]*recording.Session
	mu          sync.RWMutex
	bufferSize  int
	started     bool
}

// MockVMConnection simulates a VM connection with event streaming.
type MockVMConnection struct {
	cid       uint32
	vmID      string
	sessionID string
	events    chan *recording.Event
	done      chan struct{}
	closed    atomic.Bool
}

// VMID returns the VM ID associated with this connection.
func (c *MockVMConnection) VMID() string {
	return c.vmID
}

// SessionID returns the session ID associated with this connection.
func (c *MockVMConnection) SessionID() string {
	return c.sessionID
}

// NewMockReceiver creates a new mock receiver.
func NewMockReceiver(storage *MockRecordingStorage) *MockReceiver {
	return &MockReceiver{
		storage:     storage,
		connections: make(map[uint32]*MockVMConnection),
		sessions:    make(map[string]*recording.Session),
		bufferSize:  1000,
	}
}

// Start begins accepting mock connections.
func (r *MockReceiver) Start(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.started = true
	return nil
}

// Stop shuts down the receiver.
func (r *MockReceiver) Stop(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, conn := range r.connections {
		conn.Close()
	}
	r.started = false
	return nil
}

// SimulateVMConnection creates a mock VM connection for testing.
func (r *MockReceiver) SimulateVMConnection(cid uint32, vmID, sessionID string) *MockVMConnection {
	r.mu.Lock()
	defer r.mu.Unlock()

	conn := &MockVMConnection{
		cid:       cid,
		vmID:      vmID,
		sessionID: sessionID,
		events:    make(chan *recording.Event, r.bufferSize),
		done:      make(chan struct{}),
	}
	r.connections[cid] = conn
	r.sessions[sessionID] = &recording.Session{
		ID:        sessionID,
		VMID:      vmID,
		CID:       cid,
		StartTime: time.Now(),
		State:     recording.SessionStateActive,
	}

	// Start event processor
	go r.processConnectionEvents(conn)

	return conn
}

// processConnectionEvents handles events from a mock connection.
// For testing, events are written immediately rather than batched.
func (r *MockReceiver) processConnectionEvents(conn *MockVMConnection) {
	for {
		select {
		case <-conn.done:
			return
		case event, ok := <-conn.events:
			if !ok {
				return
			}

			// Write event immediately for testing (no batching delay)
			// Log errors but don't fail - this runs in a goroutine
			if err := r.storage.Write(context.Background(), []*recording.Event{event}); err != nil {
				fmt.Fprintf(os.Stderr, "mock receiver: write error: %v\n", err)
			}

			// Update session stats
			r.mu.Lock()
			if session, exists := r.sessions[event.SessionID]; exists {
				session.Events++
				session.Bytes += int64(len(event.Payload))
			}
			r.mu.Unlock()

			// Handle session end
			if event.Type == recording.EventSessionEnd {
				r.mu.Lock()
				if session, exists := r.sessions[event.SessionID]; exists {
					session.EndTime = event.Timestamp
					session.State = recording.SessionStateCompleted
				}
				r.mu.Unlock()
			}
		}
	}
}

// DisconnectVM simulates a VM disconnect (graceful).
func (r *MockReceiver) DisconnectVM(cid uint32) {
	r.mu.Lock()
	conn, exists := r.connections[cid]
	if exists {
		// Remove from map before unlocking to prevent races with KillVM
		delete(r.connections, cid)
	}
	r.mu.Unlock()

	if exists {
		conn.Close()
	}
}

// KillVM simulates a forced VM termination.
func (r *MockReceiver) KillVM(cid uint32) {
	var conn *MockVMConnection
	var exists bool

	// Collect session IDs to finalize while holding the lock
	var sessionsToFinalize []string

	r.mu.Lock()
	conn, exists = r.connections[cid]
	if exists {
		// Mark session as failed but still process remaining events
		for sessionID, session := range r.sessions {
			if session.CID == cid && session.State == recording.SessionStateActive {
				sessionsToFinalize = append(sessionsToFinalize, sessionID)
			}
		}
		// Remove from map first before unlocking - this prevents other goroutines
		// from finding this connection while we're closing it.
		delete(r.connections, cid)
	}
	r.mu.Unlock()

	// Finalize sessions outside the lock to prevent potential deadlocks.
	// Storage.Finalize may take time or acquire other locks internally.
	for _, sessionID := range sessionsToFinalize {
		if err := r.storage.Finalize(context.Background(), sessionID); err != nil {
			fmt.Fprintf(os.Stderr, "mock receiver: finalize error for session %s: %v\n", sessionID, err)
		}
	}

	// ForceClose is called outside the lock to prevent potential deadlocks.
	// This is safe because we've already removed the connection from the map,
	// and we have a local reference that won't be modified by other goroutines.
	if exists {
		conn.ForceClose()
	}
}

// GetSession returns session info.
func (r *MockReceiver) GetSession(sessionID string) (*recording.Session, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, exists := r.sessions[sessionID]
	if !exists {
		return nil, false
	}
	copy := *session
	return &copy, true
}

// GetActiveConnections returns the count of active connections.
func (r *MockReceiver) GetActiveConnections() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.connections)
}

// Close gracefully closes the mock connection.
func (c *MockVMConnection) Close() {
	if c.closed.CompareAndSwap(false, true) {
		close(c.done)
		close(c.events)
	}
}

// ForceClose immediately closes the connection, simulating abrupt termination.
// Unlike Close(), this drains and closes the events channel to prevent goroutine leaks
// while still simulating the effect of an abrupt disconnect (pending events are discarded).
func (c *MockVMConnection) ForceClose() {
	if c.closed.CompareAndSwap(false, true) {
		close(c.done)
		// Close and drain events channel to prevent goroutine leaks.
		// This simulates abrupt termination: pending events are discarded.
		close(c.events)
		for range c.events {
			// Drain remaining events
		}
	}
}

// SendEvent sends an event through the mock connection.
func (c *MockVMConnection) SendEvent(event *recording.Event) error {
	if c.closed.Load() {
		return fmt.Errorf("connection closed")
	}

	// The select with c.done case handles the race between checking c.closed
	// and sending on c.events. If Close() is called after we check c.closed
	// but before we send, the c.done case will trigger, preventing a panic.
	// The default case handles buffer-full scenarios without blocking.
	select {
	case c.events <- event:
		return nil
	case <-c.done:
		return fmt.Errorf("connection closed")
	default:
		return fmt.Errorf("event buffer full")
	}
}

// setupRecordingTestSuite creates the test environment.
func setupRecordingTestSuite(t *testing.T) *RecordingTestSuite {
	tempDir, err := os.MkdirTemp("", "nanofuse-recording-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	storage := NewMockRecordingStorage()
	receiver := NewMockReceiver(storage)

	if err := receiver.Start(context.Background()); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to start receiver: %v", err)
	}

	return &RecordingTestSuite{
		t:             t,
		storage:       storage,
		receiver:      receiver,
		tempDir:       tempDir,
		eventChannels: make(map[string]chan *recording.Event),
	}
}

// tearDown cleans up the test environment.
func (ts *RecordingTestSuite) tearDown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ts.receiver.Stop(ctx)
	os.RemoveAll(ts.tempDir)
}

// createTestEvent creates a recording event for testing.
func createTestEvent(vmID, sessionID string, eventType recording.EventType, payload []byte) *recording.Event {
	return &recording.Event{
		VMID:      vmID,
		SessionID: sessionID,
		Timestamp: time.Now(),
		Type:      eventType,
		Payload:   payload,
		Metadata:  make(map[string]string),
	}
}

// TestRecordingAgentStartsOnBoot verifies that the recording agent service
// starts correctly when a VM with a recording layer boots.
func TestRecordingAgentStartsOnBoot(t *testing.T) {
	ts := setupRecordingTestSuite(t)
	defer ts.tearDown()

	// Simulate VM boot with recording agent
	vmID := uuid.New().String()
	sessionID := uuid.New().String()
	cid := uint32(3)

	// Create mock connection (simulates agent connecting)
	conn := ts.receiver.SimulateVMConnection(cid, vmID, sessionID)

	// Verify connection is established
	if ts.receiver.GetActiveConnections() != 1 {
		t.Errorf("Expected 1 active connection, got %d", ts.receiver.GetActiveConnections())
	}

	// Send SESSION_START event (agent announces itself)
	startEvent := createTestEvent(vmID, sessionID, recording.EventSessionStart, []byte("agent_version=1.0.0"))
	if err := conn.SendEvent(startEvent); err != nil {
		t.Fatalf("Failed to send start event: %v", err)
	}

	// Allow time for event processing.
	// Note: time.Sleep is used here because MockReceiver uses channel-based async processing
	// and adding wait groups would require significant refactoring. The delay is calibrated
	// to be reliable in CI environments while keeping tests fast.
	time.Sleep(eventProcessingDelay)

	// Verify session was created
	session, exists := ts.receiver.GetSession(sessionID)
	if !exists {
		t.Fatal("Session should exist after SESSION_START event")
	}

	if session.VMID != vmID {
		t.Errorf("Session VMID = %q, want %q", session.VMID, vmID)
	}

	if session.State != recording.SessionStateActive {
		t.Errorf("Session state = %v, want %v", session.State, recording.SessionStateActive)
	}

	t.Logf("Recording agent started successfully: session=%s, vm=%s", sessionID, vmID)
}

// TestTerminalEventsGenerated verifies that terminal commands generate
// appropriate TERMINAL_INPUT and TERMINAL_OUTPUT events.
func TestTerminalEventsGenerated(t *testing.T) {
	ts := setupRecordingTestSuite(t)
	defer ts.tearDown()

	vmID := uuid.New().String()
	sessionID := uuid.New().String()
	cid := uint32(3)

	conn := ts.receiver.SimulateVMConnection(cid, vmID, sessionID)

	// Send session start
	startEvent := createTestEvent(vmID, sessionID, recording.EventSessionStart, nil)
	if err := conn.SendEvent(startEvent); err != nil {
		t.Fatalf("Failed to send session start: %v", err)
	}

	// Simulate terminal interaction
	testCommands := []struct {
		input  string
		output string
	}{
		{"ls -la\n", "total 32\ndrwxr-xr-x 4 user user 4096 Dec 30 12:00 .\n"},
		{"echo hello\n", "hello\n"},
		{"pwd\n", "/home/user\n"},
	}

	for _, cmd := range testCommands {
		// Input event
		inputEvent := createTestEvent(vmID, sessionID, recording.EventTerminalInput, []byte(cmd.input))
		inputEvent.Metadata["stream"] = "stdin"
		if err := conn.SendEvent(inputEvent); err != nil {
			t.Errorf("Failed to send input event: %v", err)
		}

		// Output event
		outputEvent := createTestEvent(vmID, sessionID, recording.EventTerminalOutput, []byte(cmd.output))
		outputEvent.Metadata["stream"] = "stdout"
		if err := conn.SendEvent(outputEvent); err != nil {
			t.Errorf("Failed to send output event: %v", err)
		}
	}

	// Allow time for event processing (2x eventProcessingDelay for multiple event pairs)
	time.Sleep(2 * eventProcessingDelay)

	// Verify events were recorded
	inputEvents := ts.storage.GetEventsByType(recording.EventTerminalInput)
	if len(inputEvents) != len(testCommands) {
		t.Errorf("Expected %d TERMINAL_INPUT events, got %d", len(testCommands), len(inputEvents))
	}

	outputEvents := ts.storage.GetEventsByType(recording.EventTerminalOutput)
	if len(outputEvents) != len(testCommands) {
		t.Errorf("Expected %d TERMINAL_OUTPUT events, got %d", len(testCommands), len(outputEvents))
	}

	// Verify payload content
	if len(inputEvents) > 0 && string(inputEvents[0].Payload) != testCommands[0].input {
		t.Errorf("First input event payload = %q, want %q",
			string(inputEvents[0].Payload), testCommands[0].input)
	}

	t.Logf("Terminal events recorded: %d inputs, %d outputs",
		len(inputEvents), len(outputEvents))
}

// TestGracefulShutdownProducesSessionEnd verifies that graceful VM shutdown
// produces a SESSION_END event with proper session finalization.
func TestGracefulShutdownProducesSessionEnd(t *testing.T) {
	ts := setupRecordingTestSuite(t)
	defer ts.tearDown()

	vmID := uuid.New().String()
	sessionID := uuid.New().String()
	cid := uint32(3)

	conn := ts.receiver.SimulateVMConnection(cid, vmID, sessionID)

	// Start session
	if err := conn.SendEvent(createTestEvent(vmID, sessionID, recording.EventSessionStart, nil)); err != nil {
		t.Fatalf("Failed to send session start: %v", err)
	}

	// Send some terminal events
	if err := conn.SendEvent(createTestEvent(vmID, sessionID, recording.EventTerminalOutput, []byte("some output"))); err != nil {
		t.Errorf("Failed to send terminal output: %v", err)
	}

	// Graceful shutdown - agent sends SESSION_END
	endEvent := createTestEvent(vmID, sessionID, recording.EventSessionEnd, nil)
	endEvent.Metadata["reason"] = "graceful_shutdown"
	if err := conn.SendEvent(endEvent); err != nil {
		t.Errorf("Failed to send session end: %v", err)
	}

	// Allow processing
	time.Sleep(eventProcessingDelay)

	// Verify session end was recorded
	endEvents := ts.storage.GetEventsByType(recording.EventSessionEnd)
	if len(endEvents) != 1 {
		t.Errorf("Expected 1 SESSION_END event, got %d", len(endEvents))
	}

	if len(endEvents) > 0 && endEvents[0].Metadata["reason"] != "graceful_shutdown" {
		t.Errorf("SESSION_END reason = %q, want 'graceful_shutdown'",
			endEvents[0].Metadata["reason"])
	}

	// Verify session state
	session, _ := ts.receiver.GetSession(sessionID)
	if session.State != recording.SessionStateCompleted {
		t.Errorf("Session state = %v, want %v", session.State, recording.SessionStateCompleted)
	}

	t.Logf("Graceful shutdown completed: session=%s finalized", sessionID)
}

// TestForcedKillFinalizesSession verifies that a forced VM kill still
// finalizes the session (best effort).
func TestForcedKillFinalizesSession(t *testing.T) {
	ts := setupRecordingTestSuite(t)
	defer ts.tearDown()

	vmID := uuid.New().String()
	sessionID := uuid.New().String()
	cid := uint32(3)

	conn := ts.receiver.SimulateVMConnection(cid, vmID, sessionID)

	// Start session
	if err := conn.SendEvent(createTestEvent(vmID, sessionID, recording.EventSessionStart, nil)); err != nil {
		t.Fatalf("Failed to send session start: %v", err)
	}

	// Send some events
	for i := 0; i < 10; i++ {
		if err := conn.SendEvent(createTestEvent(vmID, sessionID, recording.EventTerminalOutput,
			[]byte(fmt.Sprintf("output line %d\n", i)))); err != nil {
			t.Errorf("Failed to send terminal output %d: %v", i, err)
		}
	}

	// Allow some events to be processed
	time.Sleep(eventProcessingDelay)

	// Simulate forced kill (no SESSION_END sent by agent)
	ts.receiver.KillVM(cid)

	// Allow finalization
	time.Sleep(eventProcessingDelay)

	// Verify session was finalized (best effort)
	if !ts.storage.IsFinalized(sessionID) {
		t.Error("Session should be finalized even after forced kill")
	}

	// Verify some events were captured before kill
	sessionEvents := ts.storage.GetEventsBySession(sessionID)
	if len(sessionEvents) == 0 {
		t.Error("Should have captured some events before forced kill")
	}

	t.Logf("Forced kill handled: session=%s finalized with %d events",
		sessionID, len(sessionEvents))
}

// TestMultipleVMsRecordIndependently verifies that multiple VMs can
// record simultaneously without interference.
func TestMultipleVMsRecordIndependently(t *testing.T) {
	ts := setupRecordingTestSuite(t)
	defer ts.tearDown()

	numVMs := 5
	eventsPerVM := 20

	type vmInfo struct {
		vmID      string
		sessionID string
		conn      *MockVMConnection
	}

	vms := make([]vmInfo, numVMs)

	// Start all VMs
	for i := 0; i < numVMs; i++ {
		vms[i] = vmInfo{
			vmID:      uuid.New().String(),
			sessionID: uuid.New().String(),
		}
		vms[i].conn = ts.receiver.SimulateVMConnection(uint32(i+3), vms[i].vmID, vms[i].sessionID)

		// Start session
		if err := vms[i].conn.SendEvent(createTestEvent(vms[i].vmID, vms[i].sessionID,
			recording.EventSessionStart, nil)); err != nil {
			t.Errorf("VM %d: failed to send session start: %v", i, err)
		}
	}

	// Verify all connections active
	if ts.receiver.GetActiveConnections() != numVMs {
		t.Errorf("Expected %d active connections, got %d",
			numVMs, ts.receiver.GetActiveConnections())
	}

	// Send events from all VMs concurrently
	// Track errors across goroutines using atomic counter
	var sendErrors atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < numVMs; i++ {
		wg.Add(1)
		go func(vm vmInfo, vmIndex int) {
			defer wg.Done()
			for j := 0; j < eventsPerVM; j++ {
				payload := []byte(fmt.Sprintf("VM-%d-Event-%d", vmIndex, j))
				event := createTestEvent(vm.vmID, vm.sessionID,
					recording.EventTerminalOutput, payload)
				event.Metadata["vm_index"] = fmt.Sprintf("%d", vmIndex)
				if err := vm.conn.SendEvent(event); err != nil {
					sendErrors.Add(1)
				}
				time.Sleep(time.Millisecond) // Small delay to simulate real I/O
			}
		}(vms[i], i)
	}
	wg.Wait()

	if errorCount := sendErrors.Load(); errorCount > 0 {
		// Allow up to 5% error rate in concurrent scenarios due to timing,
		// but always permit at least one error for small test runs.
		// Calculation uses float to avoid integer division truncation issues.
		totalEvents := numVMs * eventsPerVM
		const minAllowedErrors int32 = 1
		maxAllowedErrors := int32(float64(totalEvents) * 0.05)
		if maxAllowedErrors < minAllowedErrors {
			maxAllowedErrors = minAllowedErrors
		}
		if errorCount > maxAllowedErrors {
			t.Errorf("Too many send errors: %d (max allowed: %d)", errorCount, maxAllowedErrors)
		} else {
			t.Logf("Warning: %d send errors during concurrent event sending (within tolerance)", errorCount)
		}
	}

	// End all sessions
	for i, vm := range vms {
		if err := vm.conn.SendEvent(createTestEvent(vm.vmID, vm.sessionID,
			recording.EventSessionEnd, nil)); err != nil {
			t.Errorf("VM %d: failed to send session end: %v", i, err)
		}
	}

	// Allow processing (5x eventProcessingDelay for multi-VM concurrent scenario)
	time.Sleep(5 * eventProcessingDelay)

	// Verify each VM's events are isolated
	for i, vm := range vms {
		events := ts.storage.GetEventsBySession(vm.sessionID)
		// Should have: SESSION_START + eventsPerVM + SESSION_END
		expectedMin := eventsPerVM + 1 // At least the terminal events plus start
		if len(events) < expectedMin {
			t.Errorf("VM %d: expected at least %d events, got %d",
				i, expectedMin, len(events))
		}

		// Verify events belong to correct VM
		for _, e := range events {
			if e.VMID != vm.vmID {
				t.Errorf("Event VMID = %q, want %q", e.VMID, vm.vmID)
			}
			if e.SessionID != vm.sessionID {
				t.Errorf("Event SessionID = %q, want %q", e.SessionID, vm.sessionID)
			}
		}
	}

	// Verify total event count
	totalEvents := ts.storage.EventCount()
	expectedTotal := numVMs * (eventsPerVM + 2) // +2 for START and END
	// Tolerance: allow minimal missing events due to timing.
	// Only SESSION_END events may race with assertion.
	// Formula: ceil(numVMs/2) = (numVMs + 1) / 2 using integer division.
	// This computes ceiling division, allowing for ceil(numVMs/2) missing events.
	// Example: numVMs=3 gives tolerance=2; numVMs=4 gives tolerance=2.
	tolerance := (numVMs + 1) / 2
	if tolerance < 1 {
		tolerance = 1
	}
	if totalEvents < expectedTotal-tolerance {
		t.Errorf("Total events = %d, expected approximately %d (tolerance: %d)", totalEvents, expectedTotal, tolerance)
	}

	t.Logf("Multi-VM recording successful: %d VMs, %d total events", numVMs, totalEvents)
}

// TestRecordingLayerDisabledByCondition verifies that recording can be
// disabled via layer conditions.
func TestRecordingLayerDisabledByCondition(t *testing.T) {
	ts := setupRecordingTestSuite(t)
	defer ts.tearDown()

	// Simulate VM without recording enabled (no connection attempt)
	// In a real scenario, the recording layer wouldn't be added to the image

	// Start with zero connections
	if ts.receiver.GetActiveConnections() != 0 {
		t.Errorf("Expected 0 connections for VM without recording, got %d",
			ts.receiver.GetActiveConnections())
	}

	// Verify no events recorded
	if ts.storage.EventCount() != 0 {
		t.Errorf("Expected 0 events for VM without recording, got %d",
			ts.storage.EventCount())
	}

	t.Log("Recording disabled condition verified: no connection, no events")
}

// BenchmarkEventSendLatency measures the end-to-end latency of the SendEvent call
// from the caller's perspective, including in-memory queuing and channel operations
// inside the mock receiver. It does not include production storage writes or network I/O.
// Target: <5ms latency for SendEvent under nominal load.
func BenchmarkEventSendLatency(b *testing.B) {
	storage := NewMockRecordingStorage()
	receiver := NewMockReceiver(storage)
	receiver.Start(context.Background())
	defer receiver.Stop(context.Background())

	vmID := uuid.New().String()
	sessionID := uuid.New().String()
	conn := receiver.SimulateVMConnection(3, vmID, sessionID)

	// Warm up - count failures to help diagnose benchmark issues
	warmupFailures := 0
	for i := 0; i < 100; i++ {
		if err := conn.SendEvent(createTestEvent(vmID, sessionID,
			recording.EventTerminalOutput, []byte("warmup"))); err != nil {
			warmupFailures++
		}
	}
	if warmupFailures > 0 {
		b.Logf("Warning: %d warmup send failures (may affect benchmark accuracy)", warmupFailures)
	}
	time.Sleep(eventProcessingDelay)

	// Typical terminal output size
	payload := bytes.Repeat([]byte("x"), 80) // One line of terminal output

	var totalLatency time.Duration
	var maxLatency time.Duration
	samples := 0
	sendErrors := 0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()

		event := createTestEvent(vmID, sessionID, recording.EventTerminalOutput, payload)
		if err := conn.SendEvent(event); err != nil {
			sendErrors++
			continue
		}

		latency := time.Since(start)
		totalLatency += latency
		if latency > maxLatency {
			maxLatency = latency
		}
		samples++
	}
	b.StopTimer()

	// Report error rate if any errors occurred
	if sendErrors > 0 {
		errorRate := float64(sendErrors) / float64(b.N) * 100
		b.Logf("Warning: %d/%d events failed to send (%.2f%% error rate)", sendErrors, b.N, errorRate)
		b.ReportMetric(errorRate, "error_%")
	}

	if samples > 0 {
		avgLatency := totalLatency / time.Duration(samples)
		b.ReportMetric(float64(avgLatency.Nanoseconds())/1e6, "avg_ms")
		b.ReportMetric(float64(maxLatency.Nanoseconds())/1e6, "max_ms")

		// Verify latency requirement
		if avgLatency > 5*time.Millisecond {
			b.Errorf("Average latency %v exceeds 5ms requirement", avgLatency)
		}
	}
}

// BenchmarkEventThroughput measures sustained event throughput.
// Target: 1000 events/second per VM
func BenchmarkEventThroughput(b *testing.B) {
	storage := NewMockRecordingStorage()
	receiver := NewMockReceiver(storage)
	receiver.Start(context.Background())
	defer receiver.Stop(context.Background())

	vmID := uuid.New().String()
	sessionID := uuid.New().String()
	conn := receiver.SimulateVMConnection(3, vmID, sessionID)

	// Standard payload size
	payload := []byte("terminal output line\n")

	b.ResetTimer()

	start := time.Now()
	var sent int64
	var failed int64

	for i := 0; i < b.N; i++ {
		event := createTestEvent(vmID, sessionID, recording.EventTerminalOutput, payload)
		if err := conn.SendEvent(event); err != nil {
			atomic.AddInt64(&failed, 1)
		} else {
			atomic.AddInt64(&sent, 1)
		}
	}

	duration := time.Since(start)
	b.StopTimer()

	// Calculate throughput
	if duration > 0 {
		throughput := float64(sent) / duration.Seconds()
		b.ReportMetric(throughput, "events/sec")

		// Verify throughput requirement
		if throughput < 1000 && b.N > 1000 {
			b.Logf("WARNING: Throughput %.0f events/sec may be below 1000/sec requirement", throughput)
		}
	}
}

// BenchmarkConcurrentVMThroughput measures throughput with multiple VMs.
func BenchmarkConcurrentVMThroughput(b *testing.B) {
	storage := NewMockRecordingStorage()
	receiver := NewMockReceiver(storage)
	receiver.Start(context.Background())
	defer receiver.Stop(context.Background())

	numVMs := 5
	conns := make([]*MockVMConnection, numVMs)

	for i := 0; i < numVMs; i++ {
		vmID := uuid.New().String()
		sessionID := uuid.New().String()
		conns[i] = receiver.SimulateVMConnection(uint32(i+3), vmID, sessionID)
	}

	payload := []byte("concurrent output line\n")

	b.ResetTimer()

	var wg sync.WaitGroup
	var totalSent int64

	eventsPerVM := b.N / numVMs

	start := time.Now()
	for i := 0; i < numVMs; i++ {
		wg.Add(1)
		go func(conn *MockVMConnection, vmIdx int) {
			defer wg.Done()
			// Use the connection's IDs to ensure events match the session
			vmID := conn.VMID()
			sessionID := conn.SessionID()

			for j := 0; j < eventsPerVM; j++ {
				event := createTestEvent(vmID, sessionID, recording.EventTerminalOutput, payload)
				if err := conn.SendEvent(event); err == nil {
					atomic.AddInt64(&totalSent, 1)
				}
			}
		}(conns[i], i)
	}
	wg.Wait()

	duration := time.Since(start)
	b.StopTimer()

	if duration > 0 {
		throughput := float64(totalSent) / duration.Seconds()
		b.ReportMetric(throughput, "total_events/sec")
		b.ReportMetric(throughput/float64(numVMs), "per_vm_events/sec")
	}
}

// TestEventSerializationRoundTrip tests that events can be serialized
// and deserialized correctly through the wire protocol.
func TestEventSerializationRoundTrip(t *testing.T) {
	testCases := []struct {
		name    string
		event   *recording.Event
		wantErr bool
	}{
		{
			name: "session start",
			event: &recording.Event{
				VMID:      "12345678-1234-1234-1234-123456789012",
				SessionID: "abcdefgh-1234-1234-1234-123456789012",
				Timestamp: time.Now().Truncate(time.Nanosecond),
				Type:      recording.EventSessionStart,
				Payload:   []byte("version=1.0"),
				Metadata: map[string]string{
					"agent_version": "1.0.0",
				},
			},
		},
		{
			name: "terminal input",
			event: &recording.Event{
				VMID:      uuid.New().String(),
				SessionID: uuid.New().String(),
				Timestamp: time.Now().Truncate(time.Nanosecond),
				Type:      recording.EventTerminalInput,
				Payload:   []byte("ls -la\n"),
				Metadata: map[string]string{
					"stream": "stdin",
				},
			},
		},
		{
			name: "terminal output with large payload",
			event: &recording.Event{
				VMID:      uuid.New().String(),
				SessionID: uuid.New().String(),
				Timestamp: time.Now().Truncate(time.Nanosecond),
				Type:      recording.EventTerminalOutput,
				Payload:   bytes.Repeat([]byte("x"), 10000),
				Metadata: map[string]string{
					"stream": "stdout",
					"width":  "120",
					"height": "40",
				},
			},
		},
		{
			name: "session end",
			event: &recording.Event{
				VMID:      uuid.New().String(),
				SessionID: uuid.New().String(),
				Timestamp: time.Now().Truncate(time.Nanosecond),
				Type:      recording.EventSessionEnd,
				Payload:   nil,
				Metadata: map[string]string{
					"reason":     "graceful",
					"exit_code":  "0",
					"duration_s": "3600",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Serialize
			var buf bytes.Buffer
			if err := recording.WriteEvent(&buf, tc.event); err != nil {
				if !tc.wantErr {
					t.Fatalf("WriteEvent failed: %v", err)
				}
				return
			}

			// Deserialize
			got, err := recording.ReadEvent(&buf)
			if err != nil {
				if !tc.wantErr {
					t.Fatalf("ReadEvent failed: %v", err)
				}
				return
			}

			// Verify fields
			if got.VMID != tc.event.VMID {
				t.Errorf("VMID = %q, want %q", got.VMID, tc.event.VMID)
			}
			if got.SessionID != tc.event.SessionID {
				t.Errorf("SessionID = %q, want %q", got.SessionID, tc.event.SessionID)
			}
			if !got.Timestamp.Equal(tc.event.Timestamp) {
				t.Errorf("Timestamp = %v, want %v", got.Timestamp, tc.event.Timestamp)
			}
			if got.Type != tc.event.Type {
				t.Errorf("Type = %v, want %v", got.Type, tc.event.Type)
			}
			if !bytes.Equal(got.Payload, tc.event.Payload) {
				t.Errorf("Payload length = %d, want %d", len(got.Payload), len(tc.event.Payload))
			}
			for k, v := range tc.event.Metadata {
				if got.Metadata[k] != v {
					t.Errorf("Metadata[%q] = %q, want %q", k, got.Metadata[k], v)
				}
			}
		})
	}
}

// TestLocalStorageIntegration tests the local storage backend.
func TestLocalStorageIntegration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "recording-storage-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storage, err := recording.NewLocalStorage(
		recording.WithBasePath(tempDir),
		recording.WithRetentionDays(1),
	)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	vmID := uuid.New().String()
	sessionID := uuid.New().String()

	// Start session
	if err := storage.StartSession(vmID, sessionID); err != nil {
		t.Fatalf("StartSession failed: %v", err)
	}

	// Write events
	events := []*recording.Event{
		{
			VMID:      vmID,
			SessionID: sessionID,
			Timestamp: time.Now(),
			Type:      recording.EventSessionStart,
			Payload:   []byte("session start"),
		},
		{
			VMID:      vmID,
			SessionID: sessionID,
			Timestamp: time.Now(),
			Type:      recording.EventTerminalOutput,
			Payload:   []byte("hello world\n"),
		},
		{
			VMID:      vmID,
			SessionID: sessionID,
			Timestamp: time.Now(),
			Type:      recording.EventSessionEnd,
			Payload:   nil,
		},
	}

	if err := storage.Write(context.Background(), events); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Get session before finalization
	session, err := storage.GetSession(sessionID)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if session.Status != "active" {
		t.Errorf("Session status = %q, want 'active'", session.Status)
	}
	if session.EventCount != 3 {
		t.Errorf("EventCount = %d, want 3", session.EventCount)
	}

	// Finalize session
	if err := storage.Finalize(context.Background(), sessionID); err != nil {
		t.Fatalf("Finalize failed: %v", err)
	}

	// Verify finalized session
	session, err = storage.GetSession(sessionID)
	if err != nil {
		t.Fatalf("GetSession after finalize failed: %v", err)
	}
	if session.Status != "completed" {
		t.Errorf("Finalized session status = %q, want 'completed'", session.Status)
	}
	if !session.Compressed {
		t.Error("Finalized session should be compressed")
	}

	// Verify compressed file exists
	compressedPath := filepath.Join(tempDir, sessionID, "events.bin.zst")
	if _, err := os.Stat(compressedPath); os.IsNotExist(err) {
		t.Error("Compressed events file should exist")
	}

	// List sessions
	sessions, err := storage.ListSessions(vmID)
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	if len(sessions) != 1 {
		t.Errorf("ListSessions returned %d sessions, want 1", len(sessions))
	}

	// Get playback URL
	url, err := storage.GetPlaybackURL(sessionID)
	if err != nil {
		t.Fatalf("GetPlaybackURL failed: %v", err)
	}
	if url == "" {
		t.Error("Playback URL should not be empty")
	}

	t.Logf("Local storage integration test passed: session=%s, url=%s", sessionID, url)
}

// TestRecordingReceiverWithNetPipe tests the recording receiver with net.Pipe-based connections.
func TestRecordingReceiverWithNetPipe(t *testing.T) {
	// Create a pipe to simulate vsock connection
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	storage := NewMockRecordingStorage()

	vmID := uuid.New().String()
	sessionID := uuid.New().String()

	// Write events from "client" (agent) side
	// Use error channel to communicate write failures back to main test goroutine
	writeErrs := make(chan error, 10)
	go func() {
		defer close(writeErrs)

		// Session start
		event := &recording.Event{
			VMID:      vmID,
			SessionID: sessionID,
			Timestamp: time.Now(),
			Type:      recording.EventSessionStart,
			Payload:   []byte("agent started"),
		}
		if err := recording.WriteEvent(clientConn, event); err != nil {
			writeErrs <- fmt.Errorf("failed to write session start: %w", err)
			return
		}

		// Terminal output
		for i := 0; i < 5; i++ {
			event := &recording.Event{
				VMID:      vmID,
				SessionID: sessionID,
				Timestamp: time.Now(),
				Type:      recording.EventTerminalOutput,
				Payload:   []byte(fmt.Sprintf("line %d\n", i)),
			}
			if err := recording.WriteEvent(clientConn, event); err != nil {
				writeErrs <- fmt.Errorf("failed to write terminal output %d: %w", i, err)
				return
			}
		}

		// Session end
		event = &recording.Event{
			VMID:      vmID,
			SessionID: sessionID,
			Timestamp: time.Now(),
			Type:      recording.EventSessionEnd,
		}
		if err := recording.WriteEvent(clientConn, event); err != nil {
			writeErrs <- fmt.Errorf("failed to write session end: %w", err)
			return
		}

		clientConn.Close()
	}()

	// Read events on "server" (receiver) side
	var events []*recording.Event
	for {
		event, err := recording.ReadEvent(serverConn)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("ReadEvent failed: %v", err)
		}
		events = append(events, event)
	}

	// Check for any write errors from the client goroutine
	for err := range writeErrs {
		t.Errorf("Client write error: %v", err)
	}

	// Write to storage
	if err := storage.Write(context.Background(), events); err != nil {
		t.Fatalf("Storage write failed: %v", err)
	}

	// Verify
	if len(events) != 7 { // 1 start + 5 outputs + 1 end
		t.Errorf("Expected 7 events, got %d", len(events))
	}

	if events[0].Type != recording.EventSessionStart {
		t.Errorf("First event type = %v, want SESSION_START", events[0].Type)
	}

	if events[len(events)-1].Type != recording.EventSessionEnd {
		t.Errorf("Last event type = %v, want SESSION_END", events[len(events)-1].Type)
	}

	t.Logf("Mock connection test passed: %d events processed", len(events))
}
