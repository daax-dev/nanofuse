package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/daax-dev/nanofuse/internal/types"
)

// handleHealth handles the health check endpoint (GET /health)
// Method validation is handled by the router using Go 1.22+ patterns
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(s.startTime).Seconds()

	response := types.HealthResponse{
		Status:        "healthy",
		Version:       "0.1.0",
		UptimeSeconds: int64(uptime),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Printf("ERROR: Failed to encode health response: %v", err)
	}
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	// Encoding errors are logged but we can't return an error after WriteHeader
	_ = json.NewEncoder(w).Encode(data)
}

// readJSON reads a JSON request body
func readJSON(r *http.Request, v interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(v)
}
