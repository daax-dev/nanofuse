package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/jpoley/nanofuse/examples/todo-app/backend/internal/api/rest"
	"github.com/jpoley/nanofuse/examples/todo-app/backend/internal/domain"
	"github.com/jpoley/nanofuse/examples/todo-app/backend/internal/observability"
	"github.com/jpoley/nanofuse/examples/todo-app/backend/internal/storage"
)

var (
	httpPort  = flag.Int("http-port", 8080, "HTTP server port")
	grpcPort  = flag.Int("grpc-port", 9090, "gRPC server port")
	dbPath    = flag.String("db-path", "/data/todos.db", "DuckDB database path")
	staticDir = flag.String("static-dir", "", "Directory to serve static files from (optional)")
	debug     = flag.Bool("debug", false, "Enable debug logging")
)

func main() {
	flag.Parse()

	// Initialize logger
	if err := observability.InitLogger(*debug); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer observability.Sync()

	observability.Info("Starting todo app server",
		zap.Int("http_port", *httpPort),
		zap.Int("grpc_port", *grpcPort),
		zap.String("db_path", *dbPath),
		zap.String("static_dir", *staticDir),
		zap.Bool("debug", *debug),
	)

	// Initialize database
	repo, err := storage.NewDuckDBRepository(*dbPath)
	if err != nil {
		observability.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer repo.Close()

	observability.Info("Database initialized successfully")

	// Initialize service
	service := domain.NewTodoService(repo)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", *httpPort),
		Handler:      rest.NewServer(service, *staticDir),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Register health check service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	// TODO: Register TodoService gRPC implementation
	// For now, we'll just have the health check

	// Start HTTP server
	go func() {
		observability.Info("Starting HTTP server", zap.Int("port", *httpPort))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			observability.Fatal("HTTP server failed", zap.Error(err))
		}
	}()

	// Start gRPC server
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *grpcPort))
		if err != nil {
			observability.Fatal("Failed to listen for gRPC", zap.Error(err))
		}

		observability.Info("Starting gRPC server", zap.Int("port", *grpcPort))
		if err := grpcServer.Serve(lis); err != nil {
			observability.Fatal("gRPC server failed", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	observability.Info("Shutting down servers...")

	// Shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		observability.Error("HTTP server shutdown failed", zap.Error(err))
	}

	// Shutdown gRPC server
	grpcServer.GracefulStop()

	observability.Info("Servers shut down successfully")
}
