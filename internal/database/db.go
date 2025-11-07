package database

import (
	"log"
	"time"

	"github.com/Jannnesi/Palvelinohjelmointi-GO-Group-2/internal/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Connect opens a SQLite database and performs auto migration
func Connect() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("worklogger.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// 🧠 Seed one entry for testing
	seed := domain.TimeEntry{
		UserID:      1,
		Description: "Hardcoded test entry",
		StartTime:   time.Now().Add(-2 * time.Hour),
		EndTime:     time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	db.Create(&seed) // insert into the database

	log.Println("Database connected, migrated, and seeded.")
	return db
}
