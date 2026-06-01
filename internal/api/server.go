package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	goruntime "runtime"
	"syscall"
	"time"

	"github.com/daax-dev/nanofuse/internal/applecontainer"
	"github.com/daax-dev/nanofuse/internal/config"
	"github.com/daax-dev/nanofuse/internal/firecracker"
	"github.com/daax-dev/nanofuse/internal/logging"
	"github.com/daax-dev/nanofuse/internal/network"
	"github.com/daax-dev/nanofuse/internal/recording"
	"github.com/daax-dev/nanofuse/internal/registry"
	"github.com/daax-dev/nanofuse/internal/spire"
	"github.com/daax-dev/nanofuse/internal/storage"
	"github.com/daax-dev/nanofuse/internal/vmm"
)

// Server represents the API server
type Server struct {
	config           *config.Config
	db               *storage.DB
	runtimeManager   vmm.Manager
	registryClient   *registry.Client
	ipam             *network.IPAM
	logger           *logging.Logger
	startTime        time.Time
	recordingStorage RecordingStorageInterface
	spireService     *spire.Service
}

// loadExistingAllocations loads IP allocations from existing VMs in the database
// This prevents IP conflicts after daemon restart
func loadExistingAllocations(db *storage.DB, ipam *network.IPAM, logger *logging.Logger) error {
	vms, err := db.ListVMs("")
	if err != nil {
		return fmt.Errorf("failed to list VMs: %w", err)
	}

	if len(vms) == 0 {
		logger.Info("No existing VMs found, starting with fresh IP pool")
		return nil
	}

	// Build allocation map from existing VMs
	allocations := make(map[string]string)
	for _, vm := range vms {
		if vm.Config.Network.IPAddress != "" {
			allocations[vm.ID] = vm.Config.Network.IPAddress
			logger.Info("Restored IP allocation: VM %s (%s) -> %s",
				vm.Name, vm.ID[:8], vm.Config.Network.IPAddress)
		}
	}

	if len(allocations) > 0 {
		ipam.LoadAllocations(allocations)
		logger.Info("Loaded %d existing IP allocations from database", len(allocations))
	}

	return nil
}

// initializeInfrastructure initializes database and managers
func initializeInfrastructure(cfg *config.Config, logger *logging.Logger) (*storage.DB, vmm.Manager, *registry.Client, error) {
	// Create data directory
	if err := os.MkdirAll(cfg.Storage.DataDir, 0755); err != nil {
		return nil, nil, nil, err
	}

	// Initialize database
	db, err := storage.New(cfg.Storage.Database)
	if err != nil {
		return nil, nil, nil, err
	}

	logger.Info("Database initialized: %s", cfg.Storage.Database)

	runtimeManager, err := newRuntimeManager(cfg, logger)
	if err != nil {
		return nil, nil, nil, err
	}

	// Configure SPIRE vsock proxy if enabled
	if cfg.SPIRE.Enabled && cfg.SPIRE.VsockCID >= 3 {
		if fcManager, ok := runtimeManager.(*firecracker.Manager); ok {
			if err := fcManager.SetSPIREConfig(&firecracker.SPIREProxyConfig{
				Enabled:     true,
				VsockCID:    cfg.SPIRE.VsockCID,
				VsockPort:   cfg.SPIRE.VsockPort,
				AgentSocket: cfg.SPIRE.AgentSocket,
			}); err != nil {
				return nil, nil, nil, fmt.Errorf("invalid SPIRE config: %w", err)
			}
			logger.Info("SPIRE vsock proxy enabled: CID=%d, Port=%d, AgentSocket=%s",
				cfg.SPIRE.VsockCID, cfg.SPIRE.VsockPort, cfg.SPIRE.AgentSocket)
		} else {
			logger.Warn("SPIRE vsock proxy requested but runtime %q does not support Firecracker vsock", selectedRuntimeDriver(cfg))
		}
	}

	// Initialize registry client with logging and timeouts
	registryLogger := logger.WithComponent("registry")
	registryClient := registry.NewClient(cfg.Storage.DataDir, registryLogger, registry.ClientOptions{
		PullTimeout:  time.Duration(cfg.Registry.PullTimeoutSecs) * time.Second,
		LayerTimeout: time.Duration(cfg.Registry.LayerTimeoutSecs) * time.Second,
	})

	return db, runtimeManager, registryClient, nil
}

func selectedRuntimeDriver(cfg *config.Config) string {
	if cfg == nil || cfg.Runtime.Driver == "auto" || cfg.Runtime.Driver == "" {
		if goruntime.GOOS == "darwin" {
			return applecontainer.DriverName
		}
		return "firecracker"
	}
	return cfg.Runtime.Driver
}

func newRuntimeManager(cfg *config.Config, logger *logging.Logger) (vmm.Manager, error) {
	driver := selectedRuntimeDriver(cfg)
	switch driver {
	case "firecracker":
		logger.Info("Runtime driver: firecracker")
		return firecracker.NewManager(cfg.Firecracker.BinaryPath, cfg.Storage.DataDir), nil
	case applecontainer.DriverName:
		logger.Info("Runtime driver: apple_container")
		return applecontainer.NewManager(cfg.Runtime.AppleContainer, cfg.Storage.DataDir), nil
	default:
		return nil, fmt.Errorf("unsupported runtime driver %q", driver)
	}
}

// setupNetworkInfrastructure sets up bridge and NAT
func setupNetworkInfrastructure(logger *logging.Logger) error {
	logger.Info("Setting up network infrastructure...")

	if err := network.SetupBridge(); err != nil {
		logger.Error("Failed to setup bridge: %v", err)
		logger.Info("HINT: Daemon must run as root or with CAP_NET_ADMIN capability")
		return err
	}
	logger.Info("✓ Bridge configured: nanofuse0")

	primaryIface, err := network.GetPrimaryInterface()
	if err != nil {
		logger.Warn("Could not detect primary interface: %v", err)
		primaryIface = "eth0" // Fallback
	}
	logger.Info("✓ Detected primary interface: %s", primaryIface)

	if err := network.SetupNAT(primaryIface); err != nil {
		logger.Error("Failed to setup NAT: %v", err)
		return err
	}
	logger.Info("✓ NAT configured (VMs will have internet access)")
	logger.Info("Network infrastructure ready")

	return nil
}

// setupHTTPRouter configures HTTP routes using Go 1.22+ method-aware patterns
func setupHTTPRouter(server *Server) *http.ServeMux {
	mux := http.NewServeMux()

	// Health endpoint
	mux.HandleFunc("GET /health", server.handleHealth)
	mux.HandleFunc("GET /capabilities", server.handleCapabilities)

	// VM collection endpoints
	mux.HandleFunc("GET /vms", server.handleListVMs)
	mux.HandleFunc("POST /vms", server.handleCreateVM)

	// VM instance endpoints - use path parameters
	mux.HandleFunc("GET /vms/{id}", server.handleGetVMByPath)
	mux.HandleFunc("DELETE /vms/{id}", server.handleDeleteVMByPath)

	// VM action endpoints
	mux.HandleFunc("POST /vms/{id}/start", server.handleVMStartByPath)
	mux.HandleFunc("POST /vms/{id}/stop", server.handleVMStopByPath)
	mux.HandleFunc("POST /vms/{id}/kill", server.handleVMKillByPath)
	mux.HandleFunc("POST /vms/{id}/pause", server.handleVMPauseByPath)
	mux.HandleFunc("POST /vms/{id}/resume", server.handleVMResumeByPath)
	mux.HandleFunc("GET /vms/{id}/logs", server.handleVMLogsByPath)
	mux.HandleFunc("POST /vms/{id}/exec", server.handleVMExecByPath)

	// VM snapshot endpoints
	mux.HandleFunc("GET /vms/{id}/snapshots", server.handleListSnapshotsByPath)
	mux.HandleFunc("POST /vms/{id}/snapshots", server.handleCreateSnapshotByPath)

	// VM backup endpoints
	mux.HandleFunc("GET /vms/{id}/backups", server.handleListBackupsByPath)
	mux.HandleFunc("POST /vms/{id}/backups", server.handleCreateBackupByPath)

	// Image collection endpoints
	mux.HandleFunc("GET /images", server.handleImages)
	mux.HandleFunc("POST /images/pull", server.handleImagePull)

	// Image job endpoints
	mux.HandleFunc("GET /images/jobs/{id}", server.handleImageJobByPath)

	// Image instance endpoints
	mux.HandleFunc("GET /images/{digest}", server.handleGetImageByPath)
	mux.HandleFunc("DELETE /images/{digest}", server.handleDeleteImageByPath)

	// Snapshot instance endpoints
	mux.HandleFunc("GET /snapshots/{id}", server.handleGetSnapshotByPath)
	mux.HandleFunc("DELETE /snapshots/{id}", server.handleDeleteSnapshotByPath)

	// Backup instance endpoints
	mux.HandleFunc("GET /backups/{id}", server.handleGetBackupByPath)
	mux.HandleFunc("DELETE /backups/{id}", server.handleDeleteBackupByPath)
	mux.HandleFunc("POST /backups/{id}/restore", server.handleRestoreBackupByPath)

	// Recording endpoints
	mux.HandleFunc("GET /recordings", server.handleListRecordings)
	mux.HandleFunc("GET /recordings/{id}", server.handleGetRecording)
	mux.HandleFunc("GET /recordings/{id}/events", server.handleGetRecordingEvents)
	mux.HandleFunc("POST /recordings/{id}/finalize", server.handleFinalizeRecording)
	mux.HandleFunc("DELETE /recordings/{id}", server.handleDeleteRecording)

	// VM recording endpoints
	mux.HandleFunc("GET /vms/{id}/recordings", server.handleListVMRecordings)

	return mux
}

// setupListeners creates TCP and/or Unix socket listeners
func setupListeners(cfg *config.Config, logger *logging.Logger) ([]net.Listener, error) {
	var listeners []net.Listener

	// Create Unix socket listener if configured
	if cfg.API.Socket != "" {
		socketPath := cfg.API.Socket

		// Remove existing socket
		if _, err := os.Stat(socketPath); err == nil {
			if err := os.Remove(socketPath); err != nil {
				return nil, fmt.Errorf("failed to remove existing socket: %w", err)
			}
		}

		listener, err := net.Listen("unix", socketPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create unix socket: %w", err)
		}

		// Set socket permissions to allow group access
		if err := os.Chmod(socketPath, 0666); err != nil {
			logger.Warn("Failed to set socket permissions: %v", err)
		}

		listeners = append(listeners, listener)
		logger.Info("Listening on Unix socket: %s", socketPath)
	}

	// Create TCP listener if configured
	if cfg.API.TCPBind != "" {
		listener, err := net.Listen("tcp", cfg.API.TCPBind)
		if err != nil {
			return nil, fmt.Errorf("failed to create TCP listener: %w", err)
		}

		listeners = append(listeners, listener)
		logger.Info("Listening on TCP: %s", cfg.API.TCPBind)
	}

	// Ensure at least one listener is configured
	if len(listeners) == 0 {
		return nil, fmt.Errorf("no listeners configured: set api.socket or api.tcp_bind in config")
	}

	return listeners, nil
}

// setupGracefulShutdown configures graceful shutdown handler
func setupGracefulShutdown(httpServer *http.Server, logger *logging.Logger) {
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		logger.Info("Shutdown signal received, stopping server...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(ctx); err != nil {
			logger.Error("Server shutdown error: %v", err)
		}
	}()
}

// StartWithOverrides starts the API server with CLI flag overrides
func StartWithOverrides(configPath, tcpBind, unixSocket string) error {
	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	// Apply CLI overrides
	if tcpBind != "" {
		cfg.API.TCPBind = tcpBind
	}
	if unixSocket != "" {
		cfg.API.Socket = unixSocket
	}

	if err := cfg.Validate(); err != nil {
		return err
	}

	return startServer(cfg)
}

// Start starts the API server
func Start(configPath string) error {
	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	if err := cfg.Validate(); err != nil {
		return err
	}

	return startServer(cfg)
}

// startServer contains the main server startup logic
func startServer(cfg *config.Config) error {
	// Initialize logger with file output
	logger, err := logging.New(logging.Config{
		Level:     cfg.Logging.Level,
		FilePath:  cfg.Logging.FilePath,
		Component: "nanofused",
	})
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer logger.Close()

	logger.Info("Starting NanoFuse API Daemon v0.1.0")
	if cfg.Logging.FilePath != "" {
		logger.Info("Logging to file: %s", cfg.Logging.FilePath)
	}

	// Initialize infrastructure
	db, runtimeManager, registryClient, err := initializeInfrastructure(cfg, logger)
	if err != nil {
		return err
	}
	defer db.Close()

	// Setup network when host networking is managed by nanofused.
	if cfg.Network.Setup && selectedRuntimeDriver(cfg) == "firecracker" {
		if err := setupNetworkInfrastructure(logger); err != nil {
			return err
		}
	} else if selectedRuntimeDriver(cfg) != "firecracker" {
		logger.Warn("Network infrastructure setup skipped for runtime %q; runtime-managed networking is used", selectedRuntimeDriver(cfg))
	} else {
		logger.Warn("Network infrastructure setup disabled by config; only network.mode=none VMs can start without preconfigured networking")
	}

	// Initialize IPAM
	ipam := network.NewIPAM()

	// Load existing IP allocations from database to prevent conflicts after restart
	if err := loadExistingAllocations(db, ipam, logger); err != nil {
		logger.Warn("Failed to load existing IP allocations: %v", err)
		// Continue anyway - this is not fatal, but may cause IP conflicts
	}

	logger.Info("IP address pool initialized (172.16.0.10-254, %d allocated, %d available)",
		ipam.GetAllocatedCount(), ipam.GetAvailableCount())

	// Initialize recording storage
	recordingsPath := filepath.Join(cfg.Storage.DataDir, "recordings")
	recordingStorage, err := recording.NewLocalStorage(
		recording.WithBasePath(recordingsPath),
		recording.WithRetentionDays(30),
	)
	if err != nil {
		logger.Warn("Failed to initialize recording storage: %v", err)
		// Continue without recording - not fatal
	} else {
		logger.Info("Recording storage initialized: %s", recordingsPath)
	}

	// Create server
	// Initialize SPIRE service for workload registration
	spireService := spire.NewService(&cfg.SPIRE)
	if spireService.IsEnabled() {
		logger.Info("SPIRE workload registration enabled (trust domain: %s)", cfg.SPIRE.TrustDomain)
	}

	server := &Server{
		config:           cfg,
		db:               db,
		runtimeManager:   runtimeManager,
		registryClient:   registryClient,
		ipam:             ipam,
		logger:           logger,
		startTime:        time.Now(),
		recordingStorage: recordingStorage,
		spireService:     spireService,
	}

	// Set up process exit handler to reap zombies and update VM state
	// This is critical for preventing zombie processes when VMs exit
	runtimeManager.SetProcessExitHandler(server.handleVMProcessExit)

	// Setup HTTP routes and server
	mux := setupHTTPRouter(server)
	handler := loggingMiddleware(logger)(mux)
	var authTLSConfig *tls.Config
	if cfg.Auth.Enabled && cfg.API.TCPBind != "" {
		authTLSConfig, err = BuildAuthTLSConfig(&cfg.Auth)
		if err != nil {
			return err
		}
		logger.Info("TCP API mTLS auth enabled")
	}

	// Setup listeners (can have multiple: Unix socket + TCP)
	listeners, err := setupListeners(cfg, logger)
	if err != nil {
		return err
	}

	// Create HTTP server for each listener
	errChan := make(chan error, len(listeners))
	for i, listener := range listeners {
		listenerHandler := handler
		serveListener := listener
		if cfg.Auth.Enabled && listener.Addr().Network() == "tcp" {
			serveListener = tls.NewListener(listener, authTLSConfig)
			listenerHandler = loggingMiddleware(logger)(MTLSIdentityMiddleware(logger, mux))
		}
		httpServer := &http.Server{
			Handler:           listenerHandler,
			ReadHeaderTimeout: 10 * time.Second,
		}

		// Setup graceful shutdown for this server
		setupGracefulShutdown(httpServer, logger)

		// Start server in goroutine (except last one)
		if i < len(listeners)-1 {
			go func(l net.Listener, srv *http.Server) {
				logger.Info("Starting server on listener: %s", l.Addr())
				errChan <- srv.Serve(l)
			}(serveListener, httpServer)
		} else {
			// Start last server in main goroutine
			logger.Info("NanoFuse API Daemon started successfully")
			return httpServer.Serve(serveListener)
		}
	}

	// If we had multiple listeners, wait for first error
	return <-errChan
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(logger *logging.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)
			logger.Info("%s %s - %d (%v)", r.Method, r.URL.Path, wrapped.statusCode, duration)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}
