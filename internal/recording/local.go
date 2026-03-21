package recording

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/klauspost/compress/zstd"
)

// LocalStorage stores recording sessions to the local filesystem.
// Each session is stored in its own directory with:
// - events.bin: Raw binary event stream
// - events.bin.zst: Compressed event stream (after finalization)
// - metadata.json: Session metadata
type LocalStorage struct {
	basePath      string
	retentionDays int

	mu       sync.RWMutex
	sessions map[string]*localSession
	encoder  *zstd.Encoder
}

// localSession tracks an active session's state
type localSession struct {
	metadata *SessionMetadata
	file     *os.File
	path     string
}

// SessionMetadata contains information about a recording session
type SessionMetadata struct {
	ID         string    `json:"id"`
	VMID       string    `json:"vm_id"`
	StartedAt  time.Time `json:"started_at"`
	EndedAt    time.Time `json:"ended_at,omitempty"`
	EventCount int64     `json:"event_count"`
	SizeBytes  int64     `json:"size_bytes"`
	Status     string    `json:"status"` // active, completed, failed
	Compressed bool      `json:"compressed"`
}

// LocalStorageOption configures a LocalStorage instance
type LocalStorageOption func(*LocalStorage)

// WithBasePath sets the base path for recording storage
func WithBasePath(path string) LocalStorageOption {
	return func(l *LocalStorage) {
		l.basePath = path
	}
}

// WithRetentionDays sets the retention policy
func WithRetentionDays(days int) LocalStorageOption {
	return func(l *LocalStorage) {
		l.retentionDays = days
	}
}

// NewLocalStorage creates a new local filesystem storage backend
func NewLocalStorage(opts ...LocalStorageOption) (*LocalStorage, error) {
	l := &LocalStorage{
		basePath:      "/var/lib/nanofuse/recordings",
		retentionDays: 30,
		sessions:      make(map[string]*localSession),
	}

	for _, opt := range opts {
		opt(l)
	}

	// Create base directory
	if err := os.MkdirAll(l.basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create recording directory: %w", err)
	}

	// Create zstd encoder (reusable)
	encoder, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedDefault))
	if err != nil {
		return nil, fmt.Errorf("failed to create zstd encoder: %w", err)
	}
	l.encoder = encoder

	return l, nil
}

// StartSession creates a new recording session
func (l *LocalStorage) StartSession(vmID, sessionID string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Check if session already exists
	if _, exists := l.sessions[sessionID]; exists {
		return fmt.Errorf("session %s already exists", sessionID)
	}

	// Create session directory
	sessionPath := filepath.Join(l.basePath, sessionID)
	if err := os.MkdirAll(sessionPath, 0755); err != nil {
		return fmt.Errorf("failed to create session directory: %w", err)
	}

	// Create events file
	eventsPath := filepath.Join(sessionPath, "events.bin")
	file, err := os.OpenFile(eventsPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("failed to create events file: %w", err)
	}

	// Create metadata
	metadata := &SessionMetadata{
		ID:        sessionID,
		VMID:      vmID,
		StartedAt: time.Now(),
		Status:    "active",
	}

	// Write initial metadata
	if err := l.writeMetadata(sessionPath, metadata); err != nil {
		file.Close()
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	// Track session
	l.sessions[sessionID] = &localSession{
		metadata: metadata,
		file:     file,
		path:     sessionPath,
	}

	return nil
}

// Write writes a batch of events to storage
func (l *LocalStorage) Write(ctx context.Context, events []*Event) error {
	if len(events) == 0 {
		return nil
	}

	// Group events by session
	bySession := make(map[string][]*Event)
	for _, e := range events {
		bySession[e.SessionID] = append(bySession[e.SessionID], e)
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	for sessionID, sessionEvents := range bySession {
		session, exists := l.sessions[sessionID]
		if !exists {
			// Auto-create session if it doesn't exist
			vmID := sessionEvents[0].VMID
			l.mu.Unlock()
			if err := l.StartSession(vmID, sessionID); err != nil {
				l.mu.Lock()
				return fmt.Errorf("failed to auto-create session %s: %w", sessionID, err)
			}
			l.mu.Lock()
			session = l.sessions[sessionID]
		}

		// Write events to file
		for _, e := range sessionEvents {
			if err := WriteEvent(session.file, e); err != nil {
				return fmt.Errorf("failed to write event: %w", err)
			}
			session.metadata.EventCount++
			session.metadata.SizeBytes += int64(len(e.Payload) + HeaderSize)
		}
	}

	return nil
}

// Finalize completes a recording session
func (l *LocalStorage) Finalize(ctx context.Context, sessionID string) error {
	l.mu.Lock()
	session, exists := l.sessions[sessionID]
	if !exists {
		l.mu.Unlock()
		return fmt.Errorf("session %s not found", sessionID)
	}

	// Close the events file
	if session.file != nil {
		_ = session.file.Sync() // Best effort sync before close
		session.file.Close()
		session.file = nil
	}

	sessionPath := session.path
	metadata := session.metadata
	l.mu.Unlock()

	// Compress the events file
	eventsPath := filepath.Join(sessionPath, "events.bin")
	compressedPath := filepath.Join(sessionPath, "events.bin.zst")

	if err := l.compressFile(eventsPath, compressedPath); err != nil {
		// Mark as failed but don't return error
		l.mu.Lock()
		metadata.Status = "failed"
		_ = l.writeMetadata(sessionPath, metadata) // Best effort, already returning error
		delete(l.sessions, sessionID)
		l.mu.Unlock()
		return fmt.Errorf("failed to compress session: %w", err)
	}

	// Remove uncompressed file
	os.Remove(eventsPath)

	// Update metadata
	l.mu.Lock()
	defer l.mu.Unlock()

	metadata.EndedAt = time.Now()
	metadata.Status = "completed"
	metadata.Compressed = true

	// Get compressed file size
	if info, err := os.Stat(compressedPath); err == nil {
		metadata.SizeBytes = info.Size()
	}

	if err := l.writeMetadata(sessionPath, metadata); err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}

	// Remove from active sessions
	delete(l.sessions, sessionID)

	return nil
}

// GetSession returns metadata for a session
func (l *LocalStorage) GetSession(sessionID string) (*SessionMetadata, error) {
	// Check active sessions first
	l.mu.RLock()
	if session, exists := l.sessions[sessionID]; exists {
		// Return a copy
		meta := *session.metadata
		l.mu.RUnlock()
		return &meta, nil
	}
	l.mu.RUnlock()

	// Check on disk
	sessionPath := filepath.Join(l.basePath, sessionID)
	return l.readMetadata(sessionPath)
}

// ListSessions returns all sessions for a VM
func (l *LocalStorage) ListSessions(vmID string) ([]*SessionMetadata, error) {
	// Walk the base directory
	entries, err := os.ReadDir(l.basePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []*SessionMetadata{}, nil
		}
		return nil, fmt.Errorf("failed to read recordings directory: %w", err)
	}

	// Pre-allocate with estimated capacity
	sessions := make([]*SessionMetadata, 0, len(entries))

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		sessionPath := filepath.Join(l.basePath, entry.Name())
		metadata, err := l.readMetadata(sessionPath)
		if err != nil {
			continue // Skip invalid sessions
		}

		// Filter by VM ID if specified
		if vmID != "" && metadata.VMID != vmID {
			continue
		}

		sessions = append(sessions, metadata)
	}

	return sessions, nil
}

// DeleteSession removes a session and its data
func (l *LocalStorage) DeleteSession(sessionID string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Close file if active
	if session, exists := l.sessions[sessionID]; exists {
		if session.file != nil {
			session.file.Close()
		}
		delete(l.sessions, sessionID)
	}

	// Remove directory
	sessionPath := filepath.Join(l.basePath, sessionID)
	if err := os.RemoveAll(sessionPath); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// GetPlaybackURL returns a URL for playing back a session
// For local storage, this returns a file:// URL
func (l *LocalStorage) GetPlaybackURL(sessionID string) (string, error) {
	sessionPath := filepath.Join(l.basePath, sessionID)

	// Check if session exists
	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		return "", fmt.Errorf("session %s not found", sessionID)
	}

	// Check for compressed file first
	compressedPath := filepath.Join(sessionPath, "events.bin.zst")
	if _, err := os.Stat(compressedPath); err == nil {
		return "file://" + compressedPath, nil
	}

	// Fall back to uncompressed
	eventsPath := filepath.Join(sessionPath, "events.bin")
	return "file://" + eventsPath, nil
}

// CleanupExpired removes sessions older than the retention period
func (l *LocalStorage) CleanupExpired(ctx context.Context) (int, error) {
	if l.retentionDays <= 0 {
		return 0, nil // Retention disabled
	}

	cutoff := time.Now().AddDate(0, 0, -l.retentionDays)
	var deleted int

	entries, err := os.ReadDir(l.basePath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to read recordings directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check context
		select {
		case <-ctx.Done():
			return deleted, ctx.Err()
		default:
		}

		sessionPath := filepath.Join(l.basePath, entry.Name())
		metadata, err := l.readMetadata(sessionPath)
		if err != nil {
			continue
		}

		// Only delete completed sessions
		if metadata.Status != "completed" {
			continue
		}

		// Check if expired
		if metadata.EndedAt.Before(cutoff) {
			if err := os.RemoveAll(sessionPath); err == nil {
				deleted++
			}
		}
	}

	return deleted, nil
}

// Close closes the storage backend
func (l *LocalStorage) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Close all active session files
	for _, session := range l.sessions {
		if session.file != nil {
			_ = session.file.Sync() // Best effort sync before close
			session.file.Close()
		}
	}
	l.sessions = make(map[string]*localSession)

	// Close encoder
	if l.encoder != nil {
		l.encoder.Close()
	}

	return nil
}

// writeMetadata writes session metadata to disk
func (l *LocalStorage) writeMetadata(sessionPath string, metadata *SessionMetadata) error {
	metadataPath := filepath.Join(sessionPath, "metadata.json")
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(metadataPath, data, 0600)
}

// readMetadata reads session metadata from disk
func (l *LocalStorage) readMetadata(sessionPath string) (*SessionMetadata, error) {
	metadataPath := filepath.Join(sessionPath, "metadata.json")
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, err
	}

	var metadata SessionMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

// compressFile compresses a file using zstd
func (l *LocalStorage) compressFile(src, dst string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Create encoder writing to destination
	encoder, err := zstd.NewWriter(dstFile, zstd.WithEncoderLevel(zstd.SpeedBetterCompression))
	if err != nil {
		return fmt.Errorf("failed to create encoder: %w", err)
	}

	// Copy and compress
	if _, err := io.Copy(encoder, srcFile); err != nil {
		encoder.Close()
		return fmt.Errorf("failed to compress: %w", err)
	}

	// Close encoder to flush
	if err := encoder.Close(); err != nil {
		return fmt.Errorf("failed to finalize compression: %w", err)
	}

	return nil
}

// Verify LocalStorage implements Storage interface
var _ Storage = (*LocalStorage)(nil)
