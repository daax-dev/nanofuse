package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP Metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "todo_app_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "todo_app_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// gRPC Metrics
	GRPCRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "todo_app_grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method", "status"},
	)

	GRPCRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "todo_app_grpc_request_duration_seconds",
			Help:    "gRPC request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	// Business Metrics
	TodosCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "todo_app_todos_created_total",
			Help: "Total number of todos created",
		},
	)

	TodosCompletedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "todo_app_todos_completed_total",
			Help: "Total number of todos completed",
		},
	)

	TodosDeletedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "todo_app_todos_deleted_total",
			Help: "Total number of todos deleted",
		},
	)

	TodosActiveGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "todo_app_todos_active",
			Help: "Number of active (incomplete) todos",
		},
	)

	// Database Metrics
	DBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "todo_app_db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	DBErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "todo_app_db_errors_total",
			Help: "Total number of database errors",
		},
		[]string{"operation"},
	)
)
