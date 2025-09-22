// model/content.go
package model

import (
	"encoding/json"
	"time"
)

// Character represents historical Vietnamese characters
type Character struct {
	ID           string          `json:"id" gorm:"primaryKey"`
	Name         string          `json:"name" gorm:"not null"`
	Dynasty      string          `json:"dynasty"`
	Rarity       string          `json:"rarity"` // Common, Rare, Legendary
	BirthYear    *int            `json:"birth_year"`
	DeathYear    *int            `json:"death_year"`
	Description  string          `json:"description" gorm:"type:text"`
	FamousQuote  string          `json:"famous_quote"`
	Achievements json.RawMessage `json:"achievements" gorm:"type:text"` // JSON array of achievements
	ImageURL     string          `json:"image_url"`
	IsUnlocked   bool            `json:"is_unlocked" gorm:"default:false"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// Lesson represents individual learning content
type Lesson struct {
	ID           string          `json:"id" gorm:"primaryKey"`
	CharacterID  string          `json:"character_id" gorm:"not null"`
	Title        string          `json:"title" gorm:"not null"`
	Order        int             `json:"order" gorm:"not null"` // Lesson order within character
	Story        string          `json:"story" gorm:"type:text"`
	VoiceOverURL string          `json:"voice_over_url"`
	Questions    json.RawMessage `json:"questions" gorm:"type:text"` // JSON array of questions
	XPReward     int             `json:"xp_reward" gorm:"default:50"`
	MinScore     int             `json:"min_score" gorm:"default:60"` // Minimum score to pass
	IsActive     bool            `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`

	// Relationship
	Character Character `json:"character" gorm:"foreignKey:CharacterID"`
}

// Question represents quiz questions within lessons
type Question struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"` // multiple_choice, drag_drop, fill_blank, connect
	Question string                 `json:"question"`
	Options  []string               `json:"options,omitempty"`
	Answer   interface{}            `json:"answer"`
	Points   int                    `json:"points"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Timeline represents the historical timeline structure
type Timeline struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Era         string    `json:"era"` // "Bac_Thuoc", "Doc_Lap", etc.
	Dynasty     string    `json:"dynasty"`
	StartYear   int       `json:"start_year"`
	EndYear     *int      `json:"end_year"`
	Order       int       `json:"order"`
	Description string    `json:"description"`
	IsUnlocked  bool      `json:"is_unlocked" gorm:"default:false"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserProgress represents registered user progress (different from guest)
type UserProgress struct {
	ID                 string          `json:"id" gorm:"primaryKey"`
	UserID             string          `json:"user_id" gorm:"not null"`
	Hearts             int             `json:"hearts" gorm:"default:5"`
	MaxHearts          int             `json:"max_hearts" gorm:"default:5"`
	XP                 int             `json:"xp" gorm:"default:0"`
	Level              int             `json:"level" gorm:"default:1"`
	CompletedLessons   json.RawMessage `json:"completed_lessons" gorm:"type:text"`
	UnlockedCharacters json.RawMessage `json:"unlocked_characters" gorm:"type:text"`
	Streak             int             `json:"streak" gorm:"default:0"`
	StreakFreezeUsed   bool            `json:"streak_freeze_used" gorm:"default:false"`
	TotalPlayTime      int             `json:"total_play_time" gorm:"default:0"` // in minutes
	LastHeartReset     *time.Time      `json:"last_heart_reset"`
	LastActivityDate   *time.Time      `json:"last_activity_date"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

// Achievement represents unlockable achievements
type Achievement struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	BadgeURL    string    `json:"badge_url"`
	Category    string    `json:"category"`  // learning, streak, collection, special
	Condition   string    `json:"condition"` // JSON describing unlock condition
	XPReward    int       `json:"xp_reward" gorm:"default:0"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserAchievement tracks which achievements users have unlocked
type UserAchievement struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	UserID        string    `json:"user_id" gorm:"not null"`
	AchievementID string    `json:"achievement_id" gorm:"not null"`
	UnlockedAt    time.Time `json:"unlocked_at"`
	CreatedAt     time.Time `json:"created_at"`

	// Relationship
	Achievement Achievement `json:"achievement" gorm:"foreignKey:AchievementID"`
}

// UserLessonAttempt tracks lesson attempts for registered users (different from guest)
type UserLessonAttempt struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	UserID        string    `json:"user_id" gorm:"not null"`
	LessonID      string    `json:"lesson_id" gorm:"not null"`
	IsCompleted   bool      `json:"is_completed" gorm:"not null"`
	Score         int       `json:"score" gorm:"not null"`
	TimeSpent     int       `json:"time_spent" gorm:"not null"` // in seconds
	AttemptsCount int       `json:"attempts_count" gorm:"not null"`
	CreatedAt     time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"not null"`

	// Relationship
	User   User   `json:"user" gorm:"foreignKey:UserID"`
	Lesson Lesson `json:"lesson" gorm:"foreignKey:LessonID"`
}

// Spirit/Linh Thu system
type Spirit struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id" gorm:"not null"`
	Type      string    `json:"type"`                   // Based on zodiac year
	Stage     int       `json:"stage" gorm:"default:1"` // 1-5 (egg to legendary)
	XP        int       `json:"xp" gorm:"default:0"`
	XPToNext  int       `json:"xp_to_next" gorm:"default:500"`
	Name      string    `json:"name"`
	ImageURL  string    `json:"image_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
