package router

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin" // Gin importattu
	"gorm.io/gorm"

	"github.com/Jannnesi/Palvelinohjelmointi-GO-Group-2/internal/domain"
	"github.com/Jannnesi/Palvelinohjelmointi-GO-Group-2/internal/logger"
)

// Router käyttää nyt Ginin Engineä standardikirjaston Muxin sijaan
type Router struct {
	engine *gin.Engine
	logger *logger.Logger
	db     *gorm.DB
}

// New creates a new router instance
func New(log *logger.Logger, db *gorm.DB) *Router {
	r := &Router{
		engine: gin.Default(), // Default sisältää valmiiksi loggerin ja recovery-middlewaren
		logger: log,
		db:     db,
	}
	r.setupRoutes()
	return r
}

// setupRoutes configures all application routes
func (r *Router) setupRoutes() {
	// Staattiset tiedostot on Ginissä todella helppoja
	r.engine.Static("/frontend", "./frontend")

	// Perusreitit
	r.engine.GET("/", r.rootHandler)
	r.engine.GET("/health", r.healthHandler)
	r.engine.POST("/login", r.loginHandler)

	// Ryhmitellään timeentries-reitit (API Group)
	// Tämä selkeyttää koodia
	api := r.engine.Group("/timeentries")
	{
		api.GET("", r.listTimeEntries)
		api.POST("", r.createTimeEntry)
		api.GET("/:id", r.getTimeEntryByID)    // Huomaa :id notaatio
		api.PUT("/:id", r.updateTimeEntry)     // Huomaa :id notaatio
		api.DELETE("/:id", r.deleteTimeEntry)  // Huomaa :id notaatio
	}
}

// ServeHTTP tarvitaan, jotta palvelin voidaan käynnistää main.go:ssa
// (Gin engine toteuttaa http.Handler-rajapinnan)
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.engine.ServeHTTP(w, req)
}

// --- HANDLERS ---

func (r *Router) rootHandler(c *gin.Context) {
	endpoints := map[string]string{
		"GET /":                  "List all available API endpoints",
		"GET /health":            "Check service health",
		"GET /timeentries":       "Get all time entries",
		"POST /timeentries":      "Create a new time entry",
		"GET /timeentries/:id":   "Get a single time entry",
		"PUT /timeentries/:id":   "Update an existing time entry",
		"DELETE /timeentries/:id": "Remove a time entry",
		"POST /login":            "Mock login for selecting worker/manager role",
	}
	// Gin hoitaa JSON-vastauksen ja otsikot automaattisesti
	c.JSON(http.StatusOK, endpoints)
}

func (r *Router) healthHandler(c *gin.Context) {
	// gin.H on lyhenne map[string]interface{}:lle
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// listTimeEntries handles GET /timeentries
func (r *Router) listTimeEntries(c *gin.Context) {
	var entries []domain.TimeEntry
	if err := r.db.
		Preload("User").
		Find(&entries).Error; err != nil {

		r.logger.Error("DB query error: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
		return
	}

	c.JSON(http.StatusOK, entries)
}

// createTimeEntry handles POST /timeentries
func (r *Router) createTimeEntry(c *gin.Context) {
	var entry domain.TimeEntry

	// ShouldBindJSON lukee Bodyn ja tarkistaa onko se validia JSONia
	if err := c.ShouldBindJSON(&entry); err != nil {
		r.logger.Error("JSON bind error: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if result := r.db.Create(&entry); result.Error != nil {
		r.logger.Error("DB insert error: " + result.Error.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database insert failed"})
		return
	}

	c.JSON(http.StatusCreated, entry)
}

// getTimeEntryByID handles GET /timeentries/:id
func (r *Router) getTimeEntryByID(c *gin.Context) {
	// Gin lukee parametrin suoraan URL:sta ":id"
	id := c.Param("id") 

	var entry domain.TimeEntry
	if err := r.db.First(&entry, id).Error; err != nil { // GORM osaa käyttää string-ID:tä suoraan tai sen voi muuttaa intiksi
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Time entry not found"})
		} else {
			r.logger.Error("DB query error: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
		}
		return
	}

	c.JSON(http.StatusOK, entry)
}

// updateTimeEntry handles PUT /timeentries/:id
func (r *Router) updateTimeEntry(c *gin.Context) {
	id := c.Param("id") // Saadaan ID URL:sta

	var payload domain.TimeEntry
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	var existing domain.TimeEntry
	if err := r.db.First(&existing, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Time entry not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	// Päivitetään tiedot
	// payload.ID = existing.ID // Varmistetaan ettei ID muutu, jos structissa on ID
	if err := r.db.Model(&existing).Updates(payload).Error; err != nil {
		r.logger.Error("DB update error: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}

    // Haetaan päivitetty olio (valinnainen, mutta hyvä käytäntö palauttaa tuore data)
    r.db.First(&existing, id)

	c.JSON(http.StatusOK, existing)
}

// deleteTimeEntry handles DELETE /timeentries/:id
func (r *Router) deleteTimeEntry(c *gin.Context) {
	id := c.Param("id")

	// Gormille voi antaa ID:n suoraan
	result := r.db.Delete(&domain.TimeEntry{}, id)
	
	if result.Error != nil {
		r.logger.Error("DB delete error: " + result.Error.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Delete failed"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Time entry not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

// loginHandler handles POST /login
func (r *Router) loginHandler(c *gin.Context) {
	var payload struct {
		Role string `json:"role" binding:"required"` // Gin voi tarkistaa onko kenttä pakollinen
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid body or missing role"})
		return
	}

	if payload.Role != "worker" && payload.Role != "manager" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"role":    payload.Role,
	})
}