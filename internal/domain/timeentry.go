package domain

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
