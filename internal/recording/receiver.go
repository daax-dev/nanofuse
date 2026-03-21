package recording

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/mdlayher/vsock"
)

// DefaultVsockPort is the default vsock port for recording
const DefaultVsockPort = 52

// ConnectionHandler is called when a new VM connection is established
type ConnectionHandler func(ctx context.Context, vmCID uint32, events <-chan *Event)

// Receiver listens for recording events from VMs via virtio-vsock.
// Each VM connection is handled in its own goroutine.
type Receiver struct {
	port     uint32
	listener net.Listener
	handler  ConnectionHandler
	storage  Storage

	// Connection management
	mu          sync.RWMutex
	connections map[uint32]*vmConnection // keyed by CID
	sessions    map[string]*Session      // keyed by session ID

	// Lifecycle
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	bufferSize int
}

// vmConnection represents an active connection from a VM
type vmConnection struct {
	cid       uint32
	conn      net.Conn
	sessionID string
	cancel    context.CancelFunc
}

// Session represents an active recording session
type Session struct {
	ID        string
	VMID      string
	CID       uint32
	StartTime time.Time
	EndTime   time.Time
	Events    int64
	Bytes     int64
	State     SessionState
}

// SessionState represents the state of a recording session
type SessionState string

const (
	SessionStateActive    SessionState = "active"
	SessionStateCompleted SessionState = "completed"
	SessionStateFailed    SessionState = "failed"
)

// ReceiverOption configures a Receiver
type ReceiverOption func(*Receiver)

// WithPort sets the vsock port to listen on
func WithPort(port uint32) ReceiverOption {
	return func(r *Receiver) {
		r.port = port
	}
}

// WithStorage sets the storage backend for recordings
func WithStorage(storage Storage) ReceiverOption {
	return func(r *Receiver) {
		r.storage = storage
	}
}

// WithBufferSize sets the event buffer size per connection
func WithBufferSize(size int) ReceiverOption {
	return func(r *Receiver) {
		r.bufferSize = size
	}
}

// WithHandler sets a custom connection handler
func WithHandler(handler ConnectionHandler) ReceiverOption {
	return func(r *Receiver) {
		r.handler = handler
	}
}

// NewReceiver creates a new recording receiver
func NewReceiver(opts ...ReceiverOption) *Receiver {
	r := &Receiver{
		port:        DefaultVsockPort,
		connections: make(map[uint32]*vmConnection),
		sessions:    make(map[string]*Session),
		bufferSize:  1000, // Default buffer size
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// Start begins listening for VM connections.
// This is a non-blocking call; use Stop() to stop the receiver.
func (r *Receiver) Start(ctx context.Context) error {
	r.ctx, r.cancel = context.WithCancel(ctx)

	// Create vsock listener
	// CID 2 is the host, port is our configured port
	listener, err := vsock.Listen(r.port, nil)
	if err != nil {
		return fmt.Errorf("failed to create vsock listener on port %d: %w", r.port, err)
	}
	r.listener = listener

	log.Printf("INFO: Recording receiver listening on vsock port %d", r.port)

	// Start accepting connections
	r.wg.Add(1)
	go r.acceptLoop()

	return nil
}

// Stop gracefully shuts down the receiver
func (r *Receiver) Stop(ctx context.Context) error {
	if r.cancel != nil {
		r.cancel()
	}

	// Close listener to unblock Accept
	if r.listener != nil {
		r.listener.Close()
	}

	// Close all active connections
	r.mu.Lock()
	for _, conn := range r.connections {
		if conn.cancel != nil {
			conn.cancel()
		}
		if conn.conn != nil {
			conn.conn.Close()
		}
	}
	r.mu.Unlock()

	// Wait for all goroutines to finish
	done := make(chan struct{})
	go func() {
		r.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Printf("INFO: Recording receiver stopped")
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// GetSession returns information about a session
func (r *Receiver) GetSession(sessionID string) (*Session, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	session, ok := r.sessions[sessionID]
	if !ok {
		return nil, false
	}
	// Return a copy to prevent concurrent modification
	sessionCopy := *session
	return &sessionCopy, true
}

// GetActiveConnections returns the number of active VM connections
func (r *Receiver) GetActiveConnections() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.connections)
}

// acceptLoop accepts incoming vsock connections
func (r *Receiver) acceptLoop() {
	defer r.wg.Done()

	for {
		conn, err := r.listener.Accept()
		if err != nil {
			// Check if we're shutting down
			if r.ctx.Err() != nil {
				return
			}
			// Log error but continue accepting
			log.Printf("WARN: Failed to accept vsock connection: %v", err)
			continue
		}

		// Extract CID from vsock connection
		vsockConn, ok := conn.(*vsock.Conn)
		if !ok {
			log.Printf("WARN: Connection is not vsock, closing")
			conn.Close()
			continue
		}

		cid := vsockConn.RemoteAddr().(*vsock.Addr).ContextID
		log.Printf("INFO: Accepted recording connection from CID %d", cid)

		// Handle connection in goroutine
		r.wg.Add(1)
		go r.handleConnection(conn, cid)
	}
}

// handleConnection handles a single VM connection
func (r *Receiver) handleConnection(conn net.Conn, cid uint32) {
	defer r.wg.Done()
	defer conn.Close()

	// Create connection context
	connCtx, connCancel := context.WithCancel(r.ctx)
	defer connCancel()

	// Register connection
	vmConn := &vmConnection{
		cid:    cid,
		conn:   conn,
		cancel: connCancel,
	}
	r.mu.Lock()
	r.connections[cid] = vmConn
	r.mu.Unlock()

	// Ensure cleanup
	defer func() {
		r.mu.Lock()
		delete(r.connections, cid)
		r.mu.Unlock()
	}()

	// Create event channel
	events := make(chan *Event, r.bufferSize)
	defer close(events)

	// If custom handler is set, call it
	if r.handler != nil {
		go r.handler(connCtx, cid, events)
	}

	// Start event processor (writes to storage)
	var processorWg sync.WaitGroup
	processorWg.Add(1)
	go func() {
		defer processorWg.Done()
		r.processEvents(connCtx, cid, events)
	}()

	// Read events from connection
	r.readEvents(connCtx, conn, cid, vmConn, events)

	// Wait for processor to finish
	processorWg.Wait()

	log.Printf("INFO: Connection from CID %d closed", cid)
}

// readEvents reads events from the connection and sends them to the channel
func (r *Receiver) readEvents(ctx context.Context, conn net.Conn, cid uint32, vmConn *vmConnection, events chan<- *Event) {
	for {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Set read deadline to allow periodic context checks
		_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))

		// Read event
		event, err := ReadEvent(conn)
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Printf("INFO: CID %d: Connection closed by peer", cid)
				return
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue // Timeout is expected, check context and retry
			}
			if ctx.Err() != nil {
				return // Context cancelled
			}
			log.Printf("WARN: CID %d: Error reading event: %v", cid, err)
			return
		}

		// Track session
		if event.Type == EventSessionStart {
			vmConn.sessionID = event.SessionID
			r.mu.Lock()
			r.sessions[event.SessionID] = &Session{
				ID:        event.SessionID,
				VMID:      event.VMID,
				CID:       cid,
				StartTime: event.Timestamp,
				State:     SessionStateActive,
			}
			r.mu.Unlock()
			log.Printf("INFO: CID %d: Session started: %s", cid, event.SessionID)
		}

		// Send event to channel (non-blocking with overflow handling)
		select {
		case events <- event:
			// Update session stats
			r.mu.Lock()
			if session, ok := r.sessions[event.SessionID]; ok {
				session.Events++
				session.Bytes += int64(len(event.Payload))
			}
			r.mu.Unlock()
		default:
			log.Printf("WARN: CID %d: Event buffer full, dropping event", cid)
		}

		// Handle session end
		if event.Type == EventSessionEnd {
			r.mu.Lock()
			if session, ok := r.sessions[event.SessionID]; ok {
				session.EndTime = event.Timestamp
				session.State = SessionStateCompleted
			}
			r.mu.Unlock()
			log.Printf("INFO: CID %d: Session ended: %s", cid, event.SessionID)
		}
	}
}

// processEvents processes events and writes them to storage
func (r *Receiver) processEvents(ctx context.Context, cid uint32, events <-chan *Event) {
	// Buffer for batch writes
	batch := make([]*Event, 0, 100)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		if r.storage != nil {
			if err := r.storage.Write(ctx, batch); err != nil {
				log.Printf("WARN: CID %d: Failed to write events: %v", cid, err)
			}
		}
		batch = batch[:0]
	}

	for {
		select {
		case <-ctx.Done():
			flush()
			return
		case <-ticker.C:
			flush()
		case event, ok := <-events:
			if !ok {
				flush()
				return
			}
			batch = append(batch, event)
			// Flush if batch is full
			if len(batch) >= 100 {
				flush()
			}
		}
	}
}

// Storage interface for recording storage backends
type Storage interface {
	// Write writes a batch of events to storage
	Write(ctx context.Context, events []*Event) error
	// Finalize finalizes a recording session
	Finalize(ctx context.Context, sessionID string) error
	// GetPlaybackURL returns a URL for playing back a session
	GetPlaybackURL(sessionID string) (string, error)
}
