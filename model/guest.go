package model

import (
	"encoding/json"
	"time"
)

type GuestSession struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	DeviceID     string    `json:"device_id" gorm:"not null"`
	SessionStart time.Time `json:"session_start" gorm:"not null"`
	LastActivity time.Time `json:"last_activity" gorm:"not null"`
	IsActive     bool      `json:"is_active" gorm:"not null"`
	CreatedAt    time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"not null"`
}

type GuestProgress struct {
	ID               string          `json:"id" gorm:"primaryKey"`
	GuestSessionID   string          `json:"guest_session_id" gorm:"not null"`
	Hearts           int             `json:"hearts" gorm:"not null"`
	MaxHearts        int             `json:"max_hearts" gorm:"not null"`
	XP               int             `json:"xp" gorm:"not null"`
	Level            int             `json:"level" gorm:"not null"`
	CompletedLessons json.RawMessage `json:"completed_lessons" gorm:"not null"`
	Streak           int             `json:"streak" gorm:"not null"`
	TotalPlayTime    int             `json:"total_play_time" gorm:"not null"` // in minutes
	AdsWatched       int             `json:"ads_watched" gorm:"not null"`
	CreatedAt        time.Time       `json:"created_at" gorm:"not null"`
	UpdatedAt        time.Time       `json:"updated_at" gorm:"not null"`
}

type GuestLessonAttempt struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	GuestSessionID string    `json:"guest_session_id" gorm:"not null"`
	LessonID       string    `json:"lesson_id" gorm:"not null"`
	IsCompleted    bool      `json:"is_completed" gorm:"not null"`
	Score          int       `json:"score" gorm:"not null"`
	TimeSpent      int       `json:"time_spent" gorm:"not null"` // in seconds
	AttemptsCount  int       `json:"attempts_count" gorm:"not null"`
	CreatedAt      time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"not null"`
}
