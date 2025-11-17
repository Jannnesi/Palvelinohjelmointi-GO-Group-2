package database

import (
	"log"

	"github.com/Jannnesi/Palvelinohjelmointi-GO-Group-2/internal/domain"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

/*
import "time"

// TimeEntry represents a time tracking entry
type TimeEntry struct {
	ID          uint      `json:"id" validate:"omitempty"`
	UserID      uint      `json:"user_id" validate:"required"`
	Description string    `json:"description" validate:"required,min=1,max=500"`
	StartTime   time.Time `json:"start_time" validate:"required"`
	EndTime     time.Time `json:"end_time" validate:"required,gtfield=StartTime"`
	CreatedAt   time.Time `json:"created_at" validate:"omitempty"`
	UpdatedAt   time.Time `json:"updated_at" validate:"omitempty"`
}
*/

func Connect() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("worklogger.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Run auto migration to keep schema up to date
	err = db.AutoMigrate(&domain.User{}, &domain.TimeEntry{})
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	log.Println("✅ Database connected and migrated.")
	return db

}

func GetEntries(db *gorm.DB) ([]domain.TimeEntry, error) {
	var entries []domain.TimeEntry
	result := db.Find(&entries)
	return entries, result.Error
}

func AddEntry(db *gorm.DB, entry domain.TimeEntry) error {
	return db.Create(&entry).Error
}
