# Gateway Runtime Layer Dockerfile
# Provides the Go-based gateway binary and runtime dependencies
#
# Build: docker build -t nanofuse-layer-gateway-runtime .
# Export: docker export $(docker create nanofuse-layer-gateway-runtime) | tar -C rootfs -xf -

FROM golang:1.23-bookworm AS builder

WORKDIR /build

# Copy gateway source (placeholder - will be built from main repo)
# In production, copy actual gateway source here
RUN mkdir -p /build/cmd/nanofuse-gateway

# Create placeholder main.go for gateway binary
RUN cat > /build/cmd/nanofuse-gateway/main.go << 'EOF'
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_requests_total",
			Help: "Total number of requests handled by the gateway",
		},
		[]string{"method", "path", "status"},
	)
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gateway_request_duration_seconds",
			Help:    "Request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

func init() {
	prometheus.MustRegister(requestsTotal)
	prometheus.MustRegister(requestDuration)
}

func main() {
	listenPort := flag.Int("listen-port", 8080, "Main listener port")
	metricsPort := flag.Int("metrics-port", 9090, "Prometheus metrics port")
	healthPort := flag.Int("health-port", 8081, "Health check port")
	configPath := flag.String("config", "/etc/nanofuse/gateway.yaml", "Config file path")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	slog.Info("Starting nanofuse-gateway",
		"listen_port", *listenPort,
		"metrics_port", *metricsPort,
		"health_port", *healthPort,
		"config", *configPath,
	)

	// Health check server
	healthMux := http.NewServeMux()
	healthMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		data, err := json.Marshal(map[string]string{"status": "healthy"})
		if err != nil {
			// Return JSON error response for consistency with success responses.
			slog.Error("failed to encode health response", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := w.Write([]byte(`{"error":"Internal Server Error"}`)); err != nil {
				slog.Error("failed to write health error response", "error", err, "partial_response", true)
			}
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(data); err != nil {
			// Write failed after explicit 200 OK. Client receives incomplete response.
			// Nothing more we can do here since headers are already sent.
			slog.Error("failed to write health response", "error", err)
		}
	})
	healthMux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		data, err := json.Marshal(map[string]string{"status": "ready"})
		if err != nil {
			// Return JSON error response for consistency with success responses.
			slog.Error("failed to encode ready response", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := w.Write([]byte(`{"error":"Internal Server Error"}`)); err != nil {
				slog.Error("failed to write ready error response", "error", err, "partial_response", true)
			}
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(data); err != nil {
			// Write failed after explicit 200 OK. Client receives incomplete response.
			// Nothing more we can do here since headers are already sent.
			slog.Error("failed to write ready response", "error", err)
		}
	})

	// Metrics server
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())

	// Main gateway server (placeholder implementation)
	// Note: This handler tracks and records actual HTTP status codes; a real
	// implementation may use middleware for more advanced routing and metrics.
	gatewayMux := http.NewServeMux()
	gatewayMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Pre-encode response to buffer so we can detect errors before sending headers.
		// statusCode is set based on the outcome: 200 OK for success, 500 for errors.
		response := map[string]string{
			"message": "nanofuse-gateway",
			"version": "1.0.0",
		}
		data, err := json.Marshal(response)

		var statusCode int
		if err != nil {
			// Return JSON error response for consistency with success responses.
			slog.Error("failed to encode gateway response", "error", err)
			statusCode = http.StatusInternalServerError
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			if _, err := w.Write([]byte(`{"error":"Internal Server Error"}`)); err != nil {
				slog.Error("failed to write gateway error response", "error", err, "partial_response", true)
			}
		} else {
			statusCode = http.StatusOK
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			if _, err := w.Write(data); err != nil {
				slog.Error("failed to write gateway response", "error", err, "partial_response", true)
				// After WriteHeader(200), if Write fails the client receives an incomplete
				// response body. The HTTP status code cannot be changed at this point.
				// We keep statusCode as 200 so metrics reflect the actual wire status.
				// The "partial_response=true" log field helps identify these cases in logs.
			}
		}

		// Record metrics with actual status code
		duration := time.Since(start).Seconds()
		requestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
		requestsTotal.WithLabelValues(r.Method, r.URL.Path, fmt.Sprintf("%d", statusCode)).Inc()
	})

	// Start servers
	healthServer := &http.Server{Addr: fmt.Sprintf(":%d", *healthPort), Handler: healthMux}
	metricsServer := &http.Server{Addr: fmt.Sprintf(":%d", *metricsPort), Handler: metricsMux}
	gatewayServer := &http.Server{Addr: fmt.Sprintf(":%d", *listenPort), Handler: gatewayMux}

	go func() {
		slog.Info("Starting health server", "port", *healthPort)
		if err := healthServer.ListenAndServe(); err != http.ErrServerClosed {
			slog.Error("Health server error", "error", err)
		}
	}()

	go func() {
		slog.Info("Starting metrics server", "port", *metricsPort)
		if err := metricsServer.ListenAndServe(); err != http.ErrServerClosed {
			slog.Error("Metrics server error", "error", err)
		}
	}()

	go func() {
		slog.Info("Starting gateway server", "port", *listenPort)
		if err := gatewayServer.ListenAndServe(); err != http.ErrServerClosed {
			slog.Error("Gateway server error", "error", err)
		}
	}()

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	slog.Info("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := gatewayServer.Shutdown(ctx); err != nil {
		slog.Error("Gateway server shutdown error", "error", err)
	}
	if err := metricsServer.Shutdown(ctx); err != nil {
		slog.Error("Metrics server shutdown error", "error", err)
	}
	if err := healthServer.Shutdown(ctx); err != nil {
		slog.Error("Health server shutdown error", "error", err)
	}

	slog.Info("Shutdown complete")
}
EOF

# Initialize module and download dependencies with pinned versions
# Pin to specific version for supply-chain security (prevents floating remote code)
ARG PROMETHEUS_CLIENT_VERSION=v1.20.5
RUN cd /build && \
    go mod init nanofuse-gateway && \
    go get github.com/prometheus/client_golang/prometheus@${PROMETHEUS_CLIENT_VERSION} && \
    go get github.com/prometheus/client_golang/prometheus/promhttp@${PROMETHEUS_CLIENT_VERSION}

# Build the gateway binary
RUN cd /build && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /nanofuse-gateway ./cmd/nanofuse-gateway

# Runtime image
FROM ubuntu:24.04

# Copy binary
COPY --from=builder /nanofuse-gateway /usr/local/bin/nanofuse-gateway

# Create directories
RUN mkdir -p /etc/nanofuse /var/log/nanofuse/gateway

# Create default configuration
RUN cat > /etc/nanofuse/gateway.yaml << 'EOF'
# NanoFuse Gateway Configuration

server:
  listen_port: 8080
  metrics_port: 9090
  health_port: 8081

routing:
  upstream_timeout: 30s
  max_connections: 10000
  enable_access_log: true

tls:
  enabled: false
  cert_file: ""
  key_file: ""

logging:
  level: info
  format: json
EOF
