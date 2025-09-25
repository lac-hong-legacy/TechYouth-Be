package dto

// Character DTOs
type CharacterResponse struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Era          string   `json:"era"`
	Dynasty      string   `json:"dynasty"`
	Rarity       string   `json:"rarity"`
	BirthYear    *int     `json:"birth_year"`
	DeathYear    *int     `json:"death_year"`
	Description  string   `json:"description"`
	FamousQuote  string   `json:"famous_quote"`
	Achievements []string `json:"achievements"`
	ImageURL     string   `json:"image_url"`
	IsUnlocked   bool     `json:"is_unlocked"`
	LessonCount  int      `json:"lesson_count"`
}

type CharacterCollectionResponse struct {
	Characters []CharacterResponse `json:"characters"`
	Total      int                 `json:"total"`
	Unlocked   int                 `json:"unlocked"`
}

// Lesson DTOs
type QuestionResponse struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Question string                 `json:"question"`
	Options  []string               `json:"options,omitempty"`
	Points   int                    `json:"points"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type LessonResponse struct {
	ID          string `json:"id"`
	CharacterID string `json:"character_id"`
	Title       string `json:"title"`
	Order       int    `json:"order"`
	Story       string `json:"story"`

	// Media Content
	VideoURL      string `json:"video_url"` // Video with embedded voice-over
	SubtitleURL   string `json:"subtitle_url"`
	ThumbnailURL  string `json:"thumbnail_url"`
	VideoDuration int    `json:"video_duration"`
	CanSkipAfter  int    `json:"can_skip_after"`
	HasSubtitles  bool   `json:"has_subtitles"`

	Questions []QuestionResponse `json:"questions"`
	XPReward  int                `json:"xp_reward"`
	MinScore  int                `json:"min_score"`
	Character CharacterResponse  `json:"character"`
}

type LessonAccessRequest struct {
	LessonID string `json:"lesson_id" binding:"required"`
}

type LessonAccessResponse struct {
	CanAccess    bool   `json:"can_access"`
	Reason       string `json:"reason"`
	HeartsNeeded int    `json:"hearts_needed,omitempty"`
}

type ValidateLessonRequest struct {
	LessonID    string                 `json:"lesson_id" binding:"required"`
	UserAnswers map[string]interface{} `json:"user_answers" binding:"required"`
}

type ValidateLessonResponse struct {
	Score       int  `json:"score"`
	Passed      bool `json:"passed"`
	TotalPoints int  `json:"total_points"`
	MinScore    int  `json:"min_score"`
}

type SubmitQuestionAnswerRequest struct {
	LessonID   string      `json:"lesson_id" binding:"required"`
	QuestionID string      `json:"question_id" binding:"required"`
	Answer     interface{} `json:"answer" binding:"required"`
}

type SubmitQuestionAnswerResponse struct {
	Correct      bool `json:"correct"`
	Points       int  `json:"points"`
	TotalPoints  int  `json:"total_points"`
	EarnedPoints int  `json:"earned_points"`
	CurrentScore int  `json:"current_score"`
	Passed       bool `json:"passed"`
	CanStillPass bool `json:"can_still_pass"`
	PointsNeeded int  `json:"points_needed"`
}

type CheckLessonStatusRequest struct {
	LessonID string `json:"lesson_id" binding:"required"`
}

type CheckLessonStatusResponse struct {
	Score             int  `json:"score"`
	Passed            bool `json:"passed"`
	TotalPoints       int  `json:"total_points"`
	EarnedPoints      int  `json:"earned_points"`
	MinScore          int  `json:"min_score"`
	QuestionsTotal    int  `json:"questions_total"`
	QuestionsAnswered int  `json:"questions_answered"`
	CanStillPass      bool `json:"can_still_pass"`
	PointsNeeded      int  `json:"points_needed"`
	RemainingPoints   int  `json:"remaining_points"`
}

type CompleteLessonResponse struct {
	XPGained        int    `json:"xp_gained"`
	NewLevel        int    `json:"new_level"`
	LeveledUp       bool   `json:"leveled_up"`
	CharacterUnlock string `json:"character_unlock,omitempty"`
	SpiritEvolved   bool   `json:"spirit_evolved"`
}

// Timeline DTOs
type TimelineResponse struct {
	ID          string `json:"id"`
	Era         string `json:"era"`
	Dynasty     string `json:"dynasty"`
	StartYear   int    `json:"start_year"`
	EndYear     *int   `json:"end_year"`
	Order       int    `json:"order"`
	Description string `json:"description"`
	IsUnlocked  bool   `json:"is_unlocked"`
}

type TimelineCollectionResponse struct {
	Eras []TimelineEraResponse `json:"eras"`
}

type TimelineEraResponse struct {
	Era        string                    `json:"era"`
	Dynasties  []TimelineDynastyResponse `json:"dynasties"`
	IsUnlocked bool                      `json:"is_unlocked"`
	Progress   float64                   `json:"progress"`
}

type TimelineDynastyResponse struct {
	Dynasty    string              `json:"dynasty"`
	StartYear  int                 `json:"start_year"`
	EndYear    *int                `json:"end_year"`
	Characters []CharacterResponse `json:"characters"`
	IsUnlocked bool                `json:"is_unlocked"`
	Progress   float64             `json:"progress"`
}

// Search DTOs
type SearchRequest struct {
	Query   string `json:"query" form:"query"`
	Era     string `json:"era" form:"era"`
	Dynasty string `json:"dynasty" form:"dynasty"`
	Rarity  string `json:"rarity" form:"rarity"`
	Limit   int    `json:"limit" form:"limit"`
}

type SearchResponse struct {
	Characters []CharacterResponse `json:"characters"`
	Total      int                 `json:"total"`
}

// Lesson Creation DTOs
type CreateLessonRequest struct {
	CharacterID   string                  `json:"character_id" binding:"required"`
	Title         string                  `json:"title" binding:"required"`
	Order         int                     `json:"order" binding:"required"`
	Story         string                  `json:"story"`
	VideoURL      string                  `json:"video_url"`
	SubtitleURL   string                  `json:"subtitle_url"`
	ThumbnailURL  string                  `json:"thumbnail_url"`
	VideoDuration int                     `json:"video_duration"`
	CanSkipAfter  int                     `json:"can_skip_after"`
	HasSubtitles  bool                    `json:"has_subtitles"`
	Questions     []CreateQuestionRequest `json:"questions"`
	XPReward      int                     `json:"xp_reward"`
	MinScore      int                     `json:"min_score"`
}

type CreateQuestionRequest struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type" binding:"required"`
	Question string                 `json:"question" binding:"required"`
	Options  []string               `json:"options,omitempty"`
	Answer   interface{}            `json:"answer" binding:"required"`
	Points   int                    `json:"points" binding:"required"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}
