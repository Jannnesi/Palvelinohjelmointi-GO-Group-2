package domain

import "time"

// TimeEntry represents a time tracking entry
type TimeEntry struct {
	ID          uint       `gorm:"primaryKey"`
	UserID      uint       `json:"UserID" binding:"required" gorm:"foreignKey:UserID"`
	User        User       `json:"User"`
	Project     string     `json:"Project" binding:"required"`
	Description string     `json:"Description" binding:"required"`
	StartTime   time.Time  `json:"StartTime" binding:"required"`
	EndTime     time.Time  `json:"EndTime" binding:"required"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
