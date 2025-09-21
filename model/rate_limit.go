package model

import "time"

type RateLimit struct {
	ID           string     `json:"id" gorm:"column:id;type:varchar(255);primaryKey"`
	Identifier   string     `json:"identifier" gorm:"column:identifier;type:varchar(255);not null;index"`       // IP or device_id
	EndpointType string     `json:"endpoint_type" gorm:"column:endpoint_type;type:varchar(100);not null;index"` // guest_session, lesson_complete, etc
	RequestCount int        `json:"request_count" gorm:"column:request_count;type:int;default:0"`
	WindowStart  time.Time  `json:"window_start" gorm:"column:window_start;type:datetime;not null"`
	BlockedUntil *time.Time `json:"blocked_until" gorm:"column:blocked_until;type:datetime"`
	CreatedAt    time.Time  `json:"created_at" gorm:"column:created_at;type:datetime;autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"column:updated_at;type:datetime;autoUpdateTime"`
}

type RateLimitConfig struct {
	EndpointType string        `json:"endpoint_type" gorm:"column:endpoint_type;type:varchar(100);primaryKey"`
	MaxRequests  int           `json:"max_requests" gorm:"column:max_requests;type:int;not null"`
	WindowSize   time.Duration `json:"window_size" gorm:"column:window_size;type:bigint;not null"` // Store as nanoseconds
	BlockTime    time.Duration `json:"block_time" gorm:"column:block_time;type:bigint;not null"`   // Store as nanoseconds
	Description  string        `json:"description" gorm:"column:description;type:text"`
}
