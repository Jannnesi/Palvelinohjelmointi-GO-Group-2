package database

import (
	"log"
	"time"

	"github.com/Jannnesi/Palvelinohjelmointi-GO-Group-2/internal/domain"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// Connect opens a SQLite database, runs migrations, and seeds initial data
func Connect() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("worklogger.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Keep schema in sync at startup; sqlite + gorm makes this cheap
	if err := db.AutoMigrate(&domain.User{}, &domain.TimeEntry{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	var count int64
	if err := db.Model(&domain.TimeEntry{}).Count(&count).Error; err != nil {
		log.Printf("failed to check time entry count: %v", err)
	} else if count == 0 {
		// Drop in a single entry so the UI has data on fresh installs
		seed := domain.TimeEntry{
			UserID:      1,
			Description: "Hardcoded test entry",
			StartTime:   time.Now().Add(-2 * time.Hour),
			EndTime:     time.Now(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := db.Create(&seed).Error; err != nil {
			log.Printf("failed to seed database: %v", err)
		}
	}

	log.Println("Database connected, migrated, and seeded.")
	return db
}
