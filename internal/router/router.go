package router

import (
	"encoding/json" 
	"errors"
	"net/http"
	"strconv"		
	"strings"		

	"gorm.io/gorm"

	"github.com/Jannnesi/Palvelinohjelmointi-GO-Group-2/internal/domain"
	"github.com/Jannnesi/Palvelinohjelmointi-GO-Group-2/internal/logger"
)

// Router handles HTTP routing
type Router struct { // Define the Router struct
	mux    *http.ServeMux // HTTP request multiplexer
	logger *logger.Logger // Logger for logging messages
	db     *gorm.DB       // Database connection
}

// New creates a new router instance
func New(log *logger.Logger, db *gorm.DB) *Router { // Initialize the router
	r := &Router{ 				// Create a new Router instance
		mux:    http.NewServeMux(), // Initialize the HTTP request multiplexer
		logger: log,		 	// Assign the provided logger
		db:     db,			// Assign the provided database connection
	}
	r.setupRoutes()		// Set up the application routes
	return r
}

// setupRoutes configures all application routes
func (r *Router) setupRoutes() {
	r.mux.HandleFunc("/", r.rootHandler)                   // List available endpoints
	r.mux.HandleFunc("/health", r.healthHandler)           // Health check endpoint
	r.mux.HandleFunc("/timeentries", r.timeEntriesHandler) // GET/POST
	r.mux.HandleFunc("/timeentries/", r.timeEntryByIDHandler) // GET/PUT/DELETE by ID
}

// rootHandler lists all available API endpoints
func (r *Router) rootHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	endpoints := map[string]string{
		"GET /":                    "List all available API endpoints",
		"GET /health":              "Check service health",
		"GET /timeentries":         "Get all time entries",
		"POST /timeentries":        "Create a new time entry",
		"GET /timeentries/{id}":    "Get a single time entry",
		"PUT /timeentries/{id}":    "Update an existing time entry",
		"DELETE /timeentries/{id}": "Remove a time entry",
	}
	// Serve index page at /
	r.mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, "./frontend/index.html")
	})

	// API endpoints
	r.mux.HandleFunc("/timeentries", r.timeEntriesHandler)
	r.mux.HandleFunc("/health", r.healthHandler)
	r.mux.HandleFunc("/login", r.loginHandler)

	// Serve all other frontend files (dashboard, JS, CSS)
	fs := http.FileServer(http.Dir("./frontend"))
	r.mux.Handle("/frontend/", http.StripPrefix("/frontend/", fs))
}

// healthHandler handles health check requests
func (r *Router) healthHandler(w http.ResponseWriter, req *http.Request) { // Health check endpoint
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	}); err != nil {
		r.logger.Error("Failed to encode health response: " + err.Error())
	}
}

// timeEntriesHandler routes list/create operations based on the HTTP method
func (r *Router) timeEntriesHandler(w http.ResponseWriter, req *http.Request) {
	// Fan out to the correct CRUD helper based on the HTTP verb
	switch req.Method {
	case http.MethodGet:
		r.listTimeEntries(w)
	case http.MethodPost:
		r.createTimeEntry(w, req)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listTimeEntries returns all time entries
func (r *Router) listTimeEntries(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
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

// createTimeEntry creates a new time entry
func (r *Router) createTimeEntry(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var entry domain.TimeEntry
	if err := json.NewDecoder(req.Body).Decode(&entry); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		r.logger.Error("JSON decode error: " + err.Error())
		return
	}
	if result := r.db.Create(&entry); result.Error != nil {
		http.Error(w, "Database insert failed", http.StatusInternalServerError)
		r.logger.Error("DB insert error: " + result.Error.Error())
		return
	}
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(entry); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		r.logger.Error("JSON encode error: " + err.Error())
	}
}

// timeEntryByIDHandler handles GET, PUT, DELETE for a specific time entry by ID
func (r *Router) timeEntryByIDHandler(w http.ResponseWriter, req *http.Request) {
	// Strip the prefix manually because we're on the default mux (no path params)
	idStr := strings.TrimPrefix(req.URL.Path, "/timeentries/")
	if idStr == "" {
		http.Error(w, "Time entry ID is required", http.StatusBadRequest)
		return
	}

	entryID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid time entry ID", http.StatusBadRequest)
		return
	}

	switch req.Method {
	case http.MethodGet:
		r.getTimeEntryByID(w, uint(entryID))
	case http.MethodPut:
		r.updateTimeEntry(w, req, uint(entryID))
	case http.MethodDelete:
		r.deleteTimeEntry(w, uint(entryID))
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getTimeEntryByID returns a single time entry
func (r *Router) getTimeEntryByID(w http.ResponseWriter, id uint) {
	w.Header().Set("Content-Type", "application/json")

	var entry domain.TimeEntry
	if err := r.db.First(&entry, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Time entry not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database query failed", http.StatusInternalServerError)
		}
		r.logger.Error("DB query error: " + err.Error())
		return
	}

	if err := json.NewEncoder(w).Encode(entry); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		r.logger.Error("JSON encode error: " + err.Error())
	}
}

// updateTimeEntry updates an existing time entry by ID
func (r *Router) updateTimeEntry(w http.ResponseWriter, req *http.Request, id uint) {
	w.Header().Set("Content-Type", "application/json")

	var payload domain.TimeEntry
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		r.logger.Error("JSON decode error: " + err.Error())
		return
	}

	// Fetch existing row to ensure we 404 correctly and retain missing fields
	var existing domain.TimeEntry
	if err := r.db.First(&existing, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Time entry not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database query failed", http.StatusInternalServerError)
		}
		r.logger.Error("DB query error: " + err.Error())
		return
	}

	payload.ID = id
	if err := r.db.Model(&existing).Updates(payload).Error; err != nil {
		http.Error(w, "Database update failed", http.StatusInternalServerError)
		r.logger.Error("DB update error: " + err.Error())
		return
	}

	if err := r.db.First(&existing, id).Error; err != nil {
		http.Error(w, "Database query failed", http.StatusInternalServerError)
		r.logger.Error("DB query error: " + err.Error())
		return
	}

	if err := json.NewEncoder(w).Encode(existing); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		r.logger.Error("JSON encode error: " + err.Error())
	}
}

// deleteTimeEntry removes a time entry by ID
func (r *Router) deleteTimeEntry(w http.ResponseWriter, id uint) {
	w.Header().Set("Content-Type", "application/json")

	result := r.db.Delete(&domain.TimeEntry{}, id)
	if result.Error != nil {
		http.Error(w, "Database delete failed", http.StatusInternalServerError)
		r.logger.Error("DB delete error: " + result.Error.Error())
		return
	}
	if result.RowsAffected == 0 {
		http.Error(w, "Time entry not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
