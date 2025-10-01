package model

import "time"

type RateLimit struct {
	ID           string     `json:"id" gorm:"primaryKey;type:text;not null"`
	Identifier   string     `json:"identifier" gorm:"not null;index;size:255"`
	EndpointType string     `json:"endpoint_type" gorm:"not null;size:50"`
	RequestCount int        `json:"request_count" gorm:"default:0;not null"`
	WindowStart  time.Time  `json:"window_start" gorm:"not null"`
	BlockedUntil *time.Time `json:"blocked_until,omitempty" gorm:"index"`
	CreatedAt    time.Time  `json:"created_at" gorm:"not null"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"not null"`
}

type RateLimitConfig struct {
	ID           string    `json:"id" gorm:"primaryKey;type:text;not null"`
	EndpointType string    `json:"endpoint_type" gorm:"uniqueIndex;not null;size:50"`
	Limit        int       `json:"limit" gorm:"not null"`
	WindowSize   int       `json:"window_size" gorm:"not null"` // seconds
	BlockTime    int       `json:"block_time" gorm:"not null"`  // seconds
	Description  string    `json:"description" gorm:"type:text"`
	IsActive     bool      `json:"is_active" gorm:"default:true;not null"`
	CreatedAt    time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"not null"`
}
