package rest

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/daax-dev/nanofuse/examples/todo-app/backend/internal/domain"
	"github.com/daax-dev/nanofuse/examples/todo-app/backend/internal/observability"
	"github.com/daax-dev/nanofuse/examples/todo-app/backend/internal/storage"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

type Server struct {
	service   *domain.TodoService
	router    *chi.Mux
	staticDir string
}

func NewServer(service *domain.TodoService, staticDir string) *Server {
	s := &Server{
		service:   service,
		router:    chi.NewRouter(),
		staticDir: staticDir,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

func (s *Server) setupMiddleware() {
	// Standard middleware
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Timeout(60 * time.Second))

	// Logging middleware
	s.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)

			duration := time.Since(start)

			observability.HTTPRequestsTotal.WithLabelValues(
				r.Method,
				r.URL.Path,
				strconv.Itoa(ww.Status()),
			).Inc()

			observability.HTTPRequestDuration.WithLabelValues(
				r.Method,
				r.URL.Path,
			).Observe(duration.Seconds())

			observability.Info("HTTP request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", ww.Status()),
				zap.Duration("duration", duration),
				zap.String("request_id", middleware.GetReqID(r.Context())),
			)
		})
	})

	// CORS middleware
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
}

func (s *Server) setupRoutes() {
	// Health endpoints
	s.router.Get("/health", s.handleHealth)
	s.router.Get("/ready", s.handleReady)

	// Metrics endpoint
	s.router.Handle("/metrics", promhttp.Handler())

	// API routes
	s.router.Route("/api/v1", func(r chi.Router) {
		r.Get("/todos", s.handleListTodos)
		r.Post("/todos", s.handleCreateTodo)
		r.Get("/todos/{id}", s.handleGetTodo)
		r.Put("/todos/{id}", s.handleUpdateTodo)
		r.Delete("/todos/{id}", s.handleDeleteTodo)
	})

	// Serve static files if directory is configured
	if s.staticDir != "" {
		// Serve static files at root
		// API routes are defined first, so they take precedence
		s.router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			// Check if the requested file exists
			filePath := r.URL.Path
			if filePath == "/" {
				filePath = "/index.html"
			}

			// Try to open the file to check if it exists
			f, err := http.Dir(s.staticDir).Open(filePath)
			if err != nil {
				// File not found, serve index.html for SPA routing
				http.ServeFile(w, r, s.staticDir+"/index.html")
				return
			}
			f.Close()

			// Serve the actual file
			http.ServeFile(w, r, s.staticDir+filePath)
		})

		observability.Info("Static file serving enabled", zap.String("dir", s.staticDir))
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// Health check handlers
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if err := s.service.HealthCheck(); err != nil {
		respondError(w, http.StatusServiceUnavailable, "database unhealthy", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	if err := s.service.HealthCheck(); err != nil {
		respondError(w, http.StatusServiceUnavailable, "not ready", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"status": "ready",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// Todo handlers
func (s *Server) handleListTodos(w http.ResponseWriter, r *http.Request) {
	req := domain.ListTodosRequest{
		Limit:  20,
		Offset: 0,
	}

	// Parse query parameters
	if limit := r.URL.Query().Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			req.Limit = l
		}
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			req.Offset = o
		}
	}

	if completed := r.URL.Query().Get("completed"); completed != "" {
		if c, err := strconv.ParseBool(completed); err == nil {
			req.Completed = &c
		}
	}

	if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
		req.SortBy = sortBy
	}

	if sortOrder := r.URL.Query().Get("sort_order"); sortOrder != "" {
		req.SortOrder = sortOrder
	}

	// Validate request
	if err := validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request", err)
		return
	}

	todos, total, err := s.service.List(req)
	if err != nil {
		observability.Error("Failed to list todos", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to list todos", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"todos":  todos,
		"total":  total,
		"limit":  req.Limit,
		"offset": req.Offset,
	})
}

func (s *Server) handleCreateTodo(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	if err := validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, "validation failed", err)
		return
	}

	todo, err := s.service.Create(req)
	if err != nil {
		observability.Error("Failed to create todo", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to create todo", err)
		return
	}

	observability.TodosCreatedTotal.Inc()
	if !todo.Completed {
		observability.TodosActiveGauge.Inc()
	}

	respondJSON(w, http.StatusCreated, todo)
}

func (s *Server) handleGetTodo(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid todo ID", err)
		return
	}

	todo, err := s.service.GetByID(id)
	if err != nil {
		if err == storage.ErrNotFound {
			respondError(w, http.StatusNotFound, "todo not found", err)
			return
		}
		observability.Error("Failed to get todo", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to get todo", err)
		return
	}

	respondJSON(w, http.StatusOK, todo)
}

func (s *Server) handleUpdateTodo(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid todo ID", err)
		return
	}

	var req domain.UpdateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	if err := validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, "validation failed", err)
		return
	}

	// Check if marking as completed
	if req.Completed != nil {
		oldTodo, err := s.service.GetByID(id)
		if err == nil && !oldTodo.Completed && *req.Completed {
			observability.TodosCompletedTotal.Inc()
			observability.TodosActiveGauge.Dec()
		} else if err == nil && oldTodo.Completed && !*req.Completed {
			observability.TodosActiveGauge.Inc()
		}
	}

	todo, err := s.service.Update(id, req)
	if err != nil {
		if err == storage.ErrNotFound {
			respondError(w, http.StatusNotFound, "todo not found", err)
			return
		}
		observability.Error("Failed to update todo", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to update todo", err)
		return
	}

	respondJSON(w, http.StatusOK, todo)
}

func (s *Server) handleDeleteTodo(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid todo ID", err)
		return
	}

	// Get todo first to check if it was active
	todo, err := s.service.GetByID(id)
	if err == nil && !todo.Completed {
		observability.TodosActiveGauge.Dec()
	}

	err = s.service.Delete(id)
	if err != nil {
		if err == storage.ErrNotFound {
			respondError(w, http.StatusNotFound, "todo not found", err)
			return
		}
		observability.Error("Failed to delete todo", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to delete todo", err)
		return
	}

	observability.TodosDeletedTotal.Inc()

	w.WriteHeader(http.StatusNoContent)
}

// Helper functions
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string, err error) {
	observability.Debug("Error response",
		zap.Int("status", status),
		zap.String("message", message),
		zap.Error(err),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   message,
		"details": err.Error(),
	})
}
