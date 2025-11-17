package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Jannnesi/Palvelinohjelmointi-GO-Group-2/internal/config"
	"github.com/Jannnesi/Palvelinohjelmointi-GO-Group-2/internal/database"
	"github.com/Jannnesi/Palvelinohjelmointi-GO-Group-2/internal/domain"
	"github.com/Jannnesi/Palvelinohjelmointi-GO-Group-2/internal/logger"
	"github.com/Jannnesi/Palvelinohjelmointi-GO-Group-2/internal/router"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	log := logger.New(cfg.LogLevel)
	log.Info("Starting server...")

	// Connect to SQLite database
	db := database.Connect()
	db.AutoMigrate(&domain.TimeEntry{})

	entry := domain.TimeEntry{
		UserID:      1,
		Description: "Hardcoded test entry",
		StartTime:   time.Now().Add(-2 * time.Hour),
		EndTime:     time.Now(),
	}

	err := database.AddEntry(db, entry)
	if err != nil {
		log.Error("Failed to add time entry: " + err.Error())
	} else {
		log.Info("Time entry added successfully")
	}

	// Create router
	r := router.New(log, db)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Info(fmt.Sprintf("Server listening on %s", addr))

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(fmt.Sprintf("Server failed to start: %v", err))
	}

}
