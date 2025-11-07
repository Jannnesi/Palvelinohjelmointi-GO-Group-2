package router

import (
	"encoding/json"
	"net/http"

	"gorm.io/gorm"

	"github.com/Jannnesi/Palvelinohjelmointi-GO-Group-2/internal/domain"
	"github.com/Jannnesi/Palvelinohjelmointi-GO-Group-2/internal/logger"
)

// Router handles HTTP routing
type Router struct {
	mux    *http.ServeMux
	logger *logger.Logger
	db     *gorm.DB
}

// New creates a new router instance
func New(log *logger.Logger, db *gorm.DB) *Router {
	r := &Router{
		mux:    http.NewServeMux(),
		logger: log,
		db:     db,
	}
	r.setupRoutes()
	return r
}

// setupRoutes configures all application routes
func (r *Router) setupRoutes() {
	r.mux.HandleFunc("/", r.rootHandler)
	r.mux.HandleFunc("/health", r.healthHandler)
	r.mux.HandleFunc("/timeentries", r.timeEntriesHandler)
}

// rootHandler lists all available API endpoints
func (r *Router) rootHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	endpoints := map[string]string{
		"GET /":            "List all available API endpoints",
		"GET /health":      "Check service health",
		"GET /timeentries": "Get all time entries",
	}

	if err := json.NewEncoder(w).Encode(endpoints); err != nil {
		r.logger.Error("Failed to encode endpoints response: " + err.Error())
	}
}

// healthHandler handles health check requests
func (r *Router) healthHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	}); err != nil {
		r.logger.Error("Failed to encode health response: " + err.Error())
	}
}

// timeEntriesHandler returns all time entries
func (r *Router) timeEntriesHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var entries []domain.TimeEntry
	if result := r.db.Find(&entries); result.Error != nil {
		http.Error(w, "Database query failed", http.StatusInternalServerError)
		r.logger.Error("DB query error: " + result.Error.Error())
		return
	}
	if err := json.NewEncoder(w).Encode(entries); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		r.logger.Error("JSON encode error: " + err.Error())
	}
}

// ServeHTTP implements http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
