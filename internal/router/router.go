package router

import (
	"encoding/json"
	"net/http"

	"github.com/Jannnesi/Palvelinohjelmointi-GO-Group-2/internal/logger"
)

// Router handles HTTP routing
type Router struct {
	mux    *http.ServeMux
	logger *logger.Logger
}

// New creates a new router instance
func New(log *logger.Logger) *Router {
	r := &Router{
		mux:    http.NewServeMux(),
		logger: log,
	}
	r.setupRoutes()
	return r
}

// setupRoutes configures all application routes
func (r *Router) setupRoutes() {
	r.mux.HandleFunc("/health", r.healthHandler)
}

// healthHandler handles health check requests
func (r *Router) healthHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

// ServeHTTP implements http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
