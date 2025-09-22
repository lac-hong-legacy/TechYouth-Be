package dto

import "time"

// User Profile DTOs
type UpdateProfileRequest struct {
	Username  string `json:"username"`
	BirthYear int    `json:"birth_year"`
}

type UserProfileResponse struct {
	UserID      string    `json:"user_id"`
	Email       string    `json:"email"`
	Username    string    `json:"username,omitempty"`
	BirthYear   int       `json:"birth_year,omitempty"`
	JoinedAt    time.Time `json:"joined_at"`
	LastLoginAt time.Time `json:"last_login_at"`
}

// Progress DTOs
type UserProgressResponse struct {
	UserID             string                `json:"user_id"`
	Hearts             int                   `json:"hearts"`
	MaxHearts          int                   `json:"max_hearts"`
	XP                 int                   `json:"xp"`
	Level              int                   `json:"level"`
	XPToNextLevel      int                   `json:"xp_to_next_level"`
	CompletedLessons   []string              `json:"completed_lessons"`
	UnlockedCharacters []string              `json:"unlocked_characters"`
	Streak             int                   `json:"streak"`
	TotalPlayTime      int                   `json:"total_play_time"`
	LastHeartReset     *time.Time            `json:"last_heart_reset"`
	LastActivity       *time.Time            `json:"last_activity"`
	Spirit             SpiritResponse        `json:"spirit"`
	Achievements       []AchievementResponse `json:"recent_achievements"`
}

type SpiritResponse struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Stage    int    `json:"stage"`
	XP       int    `json:"xp"`
	XPToNext int    `json:"xp_to_next"`
	Name     string `json:"name"`
	ImageURL string `json:"image_url"`
}

type AchievementResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	BadgeURL    string     `json:"badge_url"`
	Category    string     `json:"category"`
	XPReward    int        `json:"xp_reward"`
	UnlockedAt  *time.Time `json:"unlocked_at,omitempty"`
}

// Leaderboard DTOs
type LeaderboardRequest struct {
	Period string `json:"period" form:"period"` // weekly, monthly, all_time
	Limit  int    `json:"limit" form:"limit"`
}

type LeaderboardResponse struct {
	Period      string                    `json:"period"`
	CurrentUser LeaderboardUserResponse   `json:"current_user"`
	TopUsers    []LeaderboardUserResponse `json:"top_users"`
}

type LeaderboardUserResponse struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	Level       int    `json:"level"`
	XP          int    `json:"xp"`
	Rank        int    `json:"rank"`
	SpiritType  string `json:"spirit_type"`
	SpiritStage int    `json:"spirit_stage"`
}

// Statistics DTOs
type UserStatsResponse struct {
	UserID             string                 `json:"user_id"`
	Level              int                    `json:"level"`
	XP                 int                    `json:"xp"`
	Hearts             int                    `json:"hearts"`
	Streak             int                    `json:"streak"`
	TotalPlayTime      int                    `json:"total_play_time"`
	CompletedLessons   int                    `json:"completed_lessons"`
	UnlockedCharacters int                    `json:"unlocked_characters"`
	Achievements       int                    `json:"achievements"`
	Rank               int                    `json:"rank"`
	SpiritStage        int                    `json:"spirit_stage"`
	SpiritType         string                 `json:"spirit_type"`
	WeeklyStats        map[string]interface{} `json:"weekly_stats"`
	MonthlyStats       map[string]interface{} `json:"monthly_stats"`
}

// Heart Management DTOs
type AddHeartsRequest struct {
	Source string `json:"source"` // "ad", "purchase", "daily_reset"
	Amount int    `json:"amount"`
}

type HeartStatusResponse struct {
	Hearts          int        `json:"hearts"`
	MaxHearts       int        `json:"max_hearts"`
	NextResetTime   *time.Time `json:"next_reset_time"`
	CanWatchAd      bool       `json:"can_watch_ad"`
	AdsWatchedToday int        `json:"ads_watched_today"`
}

// Collection DTOs
type CollectionResponse struct {
	Characters   CharacterCollectionResponse `json:"characters"`
	Achievements []AchievementResponse       `json:"achievements"`
	Stats        CollectionStatsResponse     `json:"stats"`
}

type CollectionStatsResponse struct {
	TotalCharacters    int            `json:"total_characters"`
	UnlockedCharacters int            `json:"unlocked_characters"`
	CompletionRate     float64        `json:"completion_rate"`
	RarityBreakdown    map[string]int `json:"rarity_breakdown"`
	DynastyBreakdown   map[string]int `json:"dynasty_breakdown"`
}

// Social DTOs
type ShareRequest struct {
	Type    string `json:"type"` // "achievement", "character_unlock", "level_up"
	Content string `json:"content"`
	ItemID  string `json:"item_id"`
}

type ShareResponse struct {
	ShareURL   string   `json:"share_url"`
	ShareImage string   `json:"share_image"`
	ShareText  string   `json:"share_text"`
	Platforms  []string `json:"platforms"`
}
