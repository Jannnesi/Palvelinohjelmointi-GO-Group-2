package domain

import "time"

// TimeEntry represents a time tracking entry
type TimeEntry struct {
	ID          uint `gorm:"primaryKey"`
	UserID      uint
	Project     string
	Description string
	StartTime   time.Time
	EndTime     time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
