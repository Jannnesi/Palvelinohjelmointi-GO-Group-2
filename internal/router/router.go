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
	// API routes first
	r.mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, "./frontend/index.html")
	})

	r.mux.HandleFunc("/timeentries", r.timeEntriesHandler)
	r.mux.HandleFunc("/health", r.healthHandler)
	r.mux.HandleFunc("/workerdashboard", r.workerdashboardHandler)

	// Serve static HTML from frontend/
	fs := http.FileServer(http.Dir("./frontend"))
	r.mux.Handle("/frontend/", http.StripPrefix("/frontend/", fs))

	r.mux.HandleFunc("/login", r.loginHandler)
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

func (r *Router) workerdashboardHandler(w http.ResponseWriter, req *http.Request) {
	http.ServeFile(w, req, "frontend/workerdashboard.html")
}

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

func (r *Router) loginHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var payload struct {
		Role string `json:"role"`
	}

	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if payload.Role != "worker" && payload.Role != "manager" {
		http.Error(w, "Invalid role", http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"role":    payload.Role,
	})
}

// ServeHTTP implements http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
