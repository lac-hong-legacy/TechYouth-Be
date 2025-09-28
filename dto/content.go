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
	LessonID string `json:"lesson_id" validate:"required"`
}

func (l LessonAccessRequest) Validate() error {
	return GetValidator().Struct(l)
}

type LessonAccessResponse struct {
	CanAccess    bool   `json:"can_access"`
	Reason       string `json:"reason"`
	HeartsNeeded int    `json:"hearts_needed,omitempty"`
}

type ValidateLessonRequest struct {
	LessonID    string                 `json:"lesson_id" validate:"required"`
	UserAnswers map[string]interface{} `json:"user_answers" validate:"required"`
}

func (v ValidateLessonRequest) Validate() error {
	return GetValidator().Struct(v)
}

type ValidateLessonResponse struct {
	Score       int  `json:"score"`
	Passed      bool `json:"passed"`
	TotalPoints int  `json:"total_points"`
	MinScore    int  `json:"min_score"`
}

type SubmitQuestionAnswerRequest struct {
	LessonID   string      `json:"lesson_id" validate:"required"`
	QuestionID string      `json:"question_id" validate:"required"`
	Answer     interface{} `json:"answer" validate:"required"`
}

func (s SubmitQuestionAnswerRequest) Validate() error {
	return GetValidator().Struct(s)
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
	LessonID string `json:"lesson_id" validate:"required"`
}

func (c CheckLessonStatusRequest) Validate() error {
	return GetValidator().Struct(c)
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
	Query   string `json:"query" form:"query" validate:"omitempty,min=1,max=100"`
	Era     string `json:"era" form:"era" validate:"omitempty"`
	Dynasty string `json:"dynasty" form:"dynasty" validate:"omitempty"`
	Rarity  string `json:"rarity" form:"rarity" validate:"omitempty,oneof=common rare epic legendary"`
	Limit   int    `json:"limit" form:"limit" validate:"omitempty,min=1,max=100"`
}

func (s SearchRequest) Validate() error {
	return GetValidator().Struct(s)
}

type SearchResponse struct {
	Characters []CharacterResponse `json:"characters"`
	Total      int                 `json:"total"`
}

// Lesson Creation DTOs
type CreateLessonRequest struct {
	CharacterID   string                  `json:"character_id" validate:"required"`
	Title         string                  `json:"title" validate:"required,min=1,max=200"`
	Order         int                     `json:"order" validate:"required,min=1"`
	Story         string                  `json:"story" validate:"omitempty,max=5000"`
	VideoURL      string                  `json:"video_url" validate:"omitempty,url"`
	SubtitleURL   string                  `json:"subtitle_url" validate:"omitempty,url"`
	ThumbnailURL  string                  `json:"thumbnail_url" validate:"omitempty,url"`
	VideoDuration int                     `json:"video_duration" validate:"omitempty,min=1"`
	CanSkipAfter  int                     `json:"can_skip_after" validate:"omitempty,min=0"`
	HasSubtitles  bool                    `json:"has_subtitles"`
	Questions     []CreateQuestionRequest `json:"questions" validate:"omitempty,dive"`
	XPReward      int                     `json:"xp_reward" validate:"omitempty,min=1,max=1000"`
	MinScore      int                     `json:"min_score" validate:"omitempty,min=0,max=100"`
}

func (c CreateLessonRequest) Validate() error {
	return GetValidator().Struct(c)
}

type CreateQuestionRequest struct {
	ID       string                 `json:"id" validate:"omitempty"`
	Type     string                 `json:"type" validate:"required,oneof=multiple_choice true_false fill_blank matching"`
	Question string                 `json:"question" validate:"required,min=1,max=1000"`
	Options  []string               `json:"options,omitempty" validate:"omitempty,dive,min=1,max=200"`
	Answer   interface{}            `json:"answer" validate:"required"`
	Points   int                    `json:"points" validate:"required,min=1,max=100"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

func (c CreateQuestionRequest) Validate() error {
	return GetValidator().Struct(c)
}
