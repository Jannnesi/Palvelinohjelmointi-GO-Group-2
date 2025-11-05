package main

import (
	"fmt"
	"net/http"

	"github.com/Jannnesi/Palvelinohjelmointi-GO-Group-2/internal/config"
	"github.com/Jannnesi/Palvelinohjelmointi-GO-Group-2/internal/logger"
	"github.com/Jannnesi/Palvelinohjelmointi-GO-Group-2/internal/router"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	log := logger.New(cfg.LogLevel)
	log.Info("Starting server...")

	// Create router
	r := router.New(log)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Info(fmt.Sprintf("Server listening on %s", addr))

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(fmt.Sprintf("Server failed to start: %v", err))
	}
}
