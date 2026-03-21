package recording

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewLocalStorage(t *testing.T) {
	tmpDir := t.TempDir()

	storage, err := NewLocalStorage(
		WithBasePath(tmpDir),
		WithRetentionDays(7),
	)
	if err != nil {
		t.Fatalf("NewLocalStorage failed: %v", err)
	}
	defer storage.Close()

	// Verify directory was created
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Error("Base directory was not created")
	}
}

func TestLocalStorageStartSession(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewLocalStorage(WithBasePath(tmpDir))
	if err != nil {
		t.Fatalf("NewLocalStorage failed: %v", err)
	}
	defer storage.Close()

	// Start a session
	err = storage.StartSession("vm-123", "session-456")
	if err != nil {
		t.Fatalf("StartSession failed: %v", err)
	}

	// Verify session directory exists
	sessionDir := filepath.Join(tmpDir, "session-456")
	if _, err := os.Stat(sessionDir); os.IsNotExist(err) {
		t.Error("Session directory was not created")
	}

	// Verify events.bin exists
	eventsFile := filepath.Join(sessionDir, "events.bin")
	if _, err := os.Stat(eventsFile); os.IsNotExist(err) {
		t.Error("Events file was not created")
	}

	// Verify metadata.json exists
	metadataFile := filepath.Join(sessionDir, "metadata.json")
	if _, err := os.Stat(metadataFile); os.IsNotExist(err) {
		t.Error("Metadata file was not created")
	}

	// Starting same session again should fail
	err = storage.StartSession("vm-123", "session-456")
	if err == nil {
		t.Error("Expected error when starting duplicate session")
	}
}

func TestLocalStorageWriteEvents(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewLocalStorage(WithBasePath(tmpDir))
	if err != nil {
		t.Fatalf("NewLocalStorage failed: %v", err)
	}
	defer storage.Close()

	// Start session
	if err := storage.StartSession("vm-write", "session-write"); err != nil {
		t.Fatalf("StartSession failed: %v", err)
	}

	// Write events
	events := []*Event{
		{
			VMID:      "vm-write",
			SessionID: "session-write",
			Timestamp: time.Now(),
			Type:      EventSessionStart,
			Payload:   []byte("start"),
		},
		{
			VMID:      "vm-write",
			SessionID: "session-write",
			Timestamp: time.Now(),
			Type:      EventTerminalOutput,
			Payload:   []byte("Hello, World!\n"),
		},
	}

	if err := storage.Write(context.Background(), events); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Verify events were written
	eventsFile := filepath.Join(tmpDir, "session-write", "events.bin")
	info, err := os.Stat(eventsFile)
	if err != nil {
		t.Fatalf("Failed to stat events file: %v", err)
	}
	if info.Size() == 0 {
		t.Error("Events file is empty")
	}

	// Get session and verify event count
	session, err := storage.GetSession("session-write")
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if session.EventCount != 2 {
		t.Errorf("EventCount = %d, want 2", session.EventCount)
	}
}

func TestLocalStorageAutoCreateSession(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewLocalStorage(WithBasePath(tmpDir))
	if err != nil {
		t.Fatalf("NewLocalStorage failed: %v", err)
	}
	defer storage.Close()

	// Write events without starting session first
	events := []*Event{
		{
			VMID:      "vm-auto",
			SessionID: "session-auto",
			Timestamp: time.Now(),
			Type:      EventSessionStart,
			Payload:   []byte("auto-start"),
		},
	}

	if err := storage.Write(context.Background(), events); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Session should have been auto-created
	session, err := storage.GetSession("session-auto")
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if session == nil {
		t.Fatal("Session was not auto-created")
	}
	if session.VMID != "vm-auto" {
		t.Errorf("VMID = %q, want %q", session.VMID, "vm-auto")
	}
}

func TestLocalStorageFinalize(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewLocalStorage(WithBasePath(tmpDir))
	if err != nil {
		t.Fatalf("NewLocalStorage failed: %v", err)
	}
	defer storage.Close()

	// Start session and write events
	if err := storage.StartSession("vm-finalize", "session-finalize"); err != nil {
		t.Fatalf("StartSession failed: %v", err)
	}

	events := []*Event{
		{
			VMID:      "vm-finalize",
			SessionID: "session-finalize",
			Timestamp: time.Now(),
			Type:      EventTerminalOutput,
			Payload:   []byte("test output"),
		},
	}
	if err := storage.Write(context.Background(), events); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Finalize
	if err := storage.Finalize(context.Background(), "session-finalize"); err != nil {
		t.Fatalf("Finalize failed: %v", err)
	}

	// Verify compressed file exists
	compressedFile := filepath.Join(tmpDir, "session-finalize", "events.bin.zst")
	if _, err := os.Stat(compressedFile); os.IsNotExist(err) {
		t.Error("Compressed file was not created")
	}

	// Verify uncompressed file was removed
	eventsFile := filepath.Join(tmpDir, "session-finalize", "events.bin")
	if _, err := os.Stat(eventsFile); !os.IsNotExist(err) {
		t.Error("Uncompressed file should have been removed")
	}

	// Verify session metadata
	session, err := storage.GetSession("session-finalize")
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if session.Status != "completed" {
		t.Errorf("Status = %q, want %q", session.Status, "completed")
	}
	if !session.Compressed {
		t.Error("Compressed should be true")
	}
	if session.EndedAt.IsZero() {
		t.Error("EndedAt should be set")
	}
}

func TestLocalStorageListSessions(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewLocalStorage(WithBasePath(tmpDir))
	if err != nil {
		t.Fatalf("NewLocalStorage failed: %v", err)
	}
	defer storage.Close()

	// Create multiple sessions
	sessions := []struct {
		vmID      string
		sessionID string
	}{
		{"vm-list-1", "session-list-1"},
		{"vm-list-1", "session-list-2"},
		{"vm-list-2", "session-list-3"},
	}

	for _, s := range sessions {
		if err := storage.StartSession(s.vmID, s.sessionID); err != nil {
			t.Fatalf("StartSession failed: %v", err)
		}
	}

	// List all sessions
	allSessions, err := storage.ListSessions("")
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	if len(allSessions) != 3 {
		t.Errorf("ListSessions returned %d sessions, want 3", len(allSessions))
	}

	// List sessions for specific VM
	vm1Sessions, err := storage.ListSessions("vm-list-1")
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	if len(vm1Sessions) != 2 {
		t.Errorf("ListSessions for vm-list-1 returned %d sessions, want 2", len(vm1Sessions))
	}
}

func TestLocalStorageDeleteSession(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewLocalStorage(WithBasePath(tmpDir))
	if err != nil {
		t.Fatalf("NewLocalStorage failed: %v", err)
	}
	defer storage.Close()

	// Create and delete session
	if err := storage.StartSession("vm-delete", "session-delete"); err != nil {
		t.Fatalf("StartSession failed: %v", err)
	}

	if err := storage.DeleteSession("session-delete"); err != nil {
		t.Fatalf("DeleteSession failed: %v", err)
	}

	// Verify session is gone
	sessionDir := filepath.Join(tmpDir, "session-delete")
	if _, err := os.Stat(sessionDir); !os.IsNotExist(err) {
		t.Error("Session directory should have been deleted")
	}

	// GetSession should return nil
	session, err := storage.GetSession("session-delete")
	if err == nil && session != nil {
		t.Error("GetSession should return error or nil for deleted session")
	}
}

func TestLocalStorageGetPlaybackURL(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewLocalStorage(WithBasePath(tmpDir))
	if err != nil {
		t.Fatalf("NewLocalStorage failed: %v", err)
	}
	defer storage.Close()

	// Create session
	if err := storage.StartSession("vm-url", "session-url"); err != nil {
		t.Fatalf("StartSession failed: %v", err)
	}

	// Get playback URL for active session
	url, err := storage.GetPlaybackURL("session-url")
	if err != nil {
		t.Fatalf("GetPlaybackURL failed: %v", err)
	}
	if url != "file://"+filepath.Join(tmpDir, "session-url", "events.bin") {
		t.Errorf("Unexpected URL: %s", url)
	}

	// Get URL for non-existent session
	_, err = storage.GetPlaybackURL("session-nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent session")
	}
}

func TestLocalStorageCleanupExpired(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewLocalStorage(
		WithBasePath(tmpDir),
		WithRetentionDays(1),
	)
	if err != nil {
		t.Fatalf("NewLocalStorage failed: %v", err)
	}
	defer storage.Close()

	// Create session and finalize it
	if err := storage.StartSession("vm-expire", "session-expire"); err != nil {
		t.Fatalf("StartSession failed: %v", err)
	}

	events := []*Event{{
		VMID:      "vm-expire",
		SessionID: "session-expire",
		Timestamp: time.Now(),
		Type:      EventSessionEnd,
	}}
	if err := storage.Write(context.Background(), events); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if err := storage.Finalize(context.Background(), "session-expire"); err != nil {
		t.Fatalf("Finalize failed: %v", err)
	}

	// Manually set the ended_at time to 2 days ago by updating metadata
	sessionPath := filepath.Join(tmpDir, "session-expire")
	metadata, _ := storage.readMetadata(sessionPath)
	metadata.EndedAt = time.Now().AddDate(0, 0, -2)
	storage.writeMetadata(sessionPath, metadata)

	// Run cleanup
	deleted, err := storage.CleanupExpired(context.Background())
	if err != nil {
		t.Fatalf("CleanupExpired failed: %v", err)
	}
	if deleted != 1 {
		t.Errorf("CleanupExpired deleted %d sessions, want 1", deleted)
	}

	// Verify session is gone
	if _, err := os.Stat(sessionPath); !os.IsNotExist(err) {
		t.Error("Expired session should have been deleted")
	}
}

func TestLocalStorageClose(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewLocalStorage(WithBasePath(tmpDir))
	if err != nil {
		t.Fatalf("NewLocalStorage failed: %v", err)
	}

	// Start a session
	if err := storage.StartSession("vm-close", "session-close"); err != nil {
		t.Fatalf("StartSession failed: %v", err)
	}

	// Close should not error
	if err := storage.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// After close, operations should fail gracefully or handle the closed state
	// The implementation should be resilient to operations after close
}
