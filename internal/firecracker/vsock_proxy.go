package firecracker

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

// VsockProxy proxies connections from Firecracker vsock to a target Unix socket.
// This enables VMs to access host services (like SPIRE agent) via virtio-vsock.
//
// Architecture:
//   - Firecracker creates a vsock device with a host-side Unix socket
//   - When VM connects to CID 2 (host) on the configured port, Firecracker
//     establishes a Unix socket connection to us
//   - We forward this connection to the target socket (e.g., SPIRE agent)
type VsockProxy struct {
	vsockPath    string       // Path to Firecracker's vsock Unix socket
	targetSocket string       // Target socket to forward connections to
	port         uint32       // Port number (for logging/validation)
	listener     net.Listener // Listener on vsock path
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	mu           sync.Mutex
	running      bool
}

// NewVsockProxy creates a new vsock proxy.
// vsockPath: Path where Firecracker exposes the vsock host socket
// targetSocket: Path to the target Unix socket (e.g., SPIRE agent socket)
// port: The port number the VM will connect to (for documentation/logging)
func NewVsockProxy(vsockPath, targetSocket string, port uint32) (*VsockProxy, error) {
	if vsockPath == "" {
		return nil, fmt.Errorf("vsock path is required")
	}
	if targetSocket == "" {
		return nil, fmt.Errorf("target socket is required")
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &VsockProxy{
		vsockPath:    vsockPath,
		targetSocket: targetSocket,
		port:         port,
		ctx:          ctx,
		cancel:       cancel,
	}, nil
}

// Start begins listening for connections on the vsock socket.
// Connections are proxied to the target socket.
func (p *VsockProxy) Start() error {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return fmt.Errorf("proxy already running")
	}

	// Remove existing socket file if present
	if err := os.Remove(p.vsockPath); err != nil && !os.IsNotExist(err) {
		p.mu.Unlock()
		return fmt.Errorf("failed to remove existing socket: %w", err)
	}

	// Create listener on vsock path
	listener, err := net.Listen("unix", p.vsockPath)
	if err != nil {
		p.mu.Unlock()
		return fmt.Errorf("failed to listen on vsock socket: %w", err)
	}

	p.listener = listener
	p.running = true
	p.mu.Unlock()

	// Start accept loop in goroutine
	p.wg.Add(1)
	go p.acceptLoop()

	log.Printf("INFO: Vsock proxy started: %s -> %s (port %d)", p.vsockPath, p.targetSocket, p.port)
	return nil
}

// Stop gracefully stops the proxy.
func (p *VsockProxy) Stop() {
	p.mu.Lock()
	if !p.running {
		p.mu.Unlock()
		return
	}
	p.running = false
	p.mu.Unlock()

	// Signal shutdown
	p.cancel()

	// Close listener to unblock Accept()
	if p.listener != nil {
		p.listener.Close()
	}

	// Wait for goroutines to finish (with timeout)
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Printf("INFO: Vsock proxy stopped cleanly: %s", p.vsockPath)
	case <-time.After(5 * time.Second):
		log.Printf("WARN: Vsock proxy stop timed out: %s", p.vsockPath)
	}

	// Clean up socket file
	os.Remove(p.vsockPath)
}

// acceptLoop accepts connections and spawns handlers.
func (p *VsockProxy) acceptLoop() {
	defer p.wg.Done()

	for {
		conn, err := p.listener.Accept()
		if err != nil {
			select {
			case <-p.ctx.Done():
				// Normal shutdown
				return
			default:
				log.Printf("WARN: Vsock accept error: %v", err)
				continue
			}
		}

		// Handle connection in goroutine
		p.wg.Add(1)
		go p.handleConnection(conn)
	}
}

// handleConnection proxies a single connection to the target socket.
func (p *VsockProxy) handleConnection(vsockConn net.Conn) {
	defer p.wg.Done()
	defer vsockConn.Close()

	// Connect to target socket (SPIRE agent)
	targetConn, err := net.Dial("unix", p.targetSocket)
	if err != nil {
		log.Printf("ERROR: Failed to connect to target socket %s: %v", p.targetSocket, err)
		return
	}
	defer targetConn.Close()

	// Set up bidirectional copy with context awareness
	errChan := make(chan error, 2)

	// Copy vsock -> target
	go func() {
		_, err := io.Copy(targetConn, vsockConn)
		errChan <- err
	}()

	// Copy target -> vsock
	go func() {
		_, err := io.Copy(vsockConn, targetConn)
		errChan <- err
	}()

	// Wait for either direction to complete or context cancellation
	select {
	case <-p.ctx.Done():
		// Shutdown requested - connections will be closed by defers
		return
	case err := <-errChan:
		if err != nil && err != io.EOF {
			log.Printf("DEBUG: Vsock proxy connection ended: %v", err)
		}
	}
}

// IsRunning returns true if the proxy is currently running.
func (p *VsockProxy) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.running
}
