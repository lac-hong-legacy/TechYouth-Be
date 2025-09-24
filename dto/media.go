package dto

// Media Upload DTOs
type MediaUploadResponse struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	FileName string `json:"file_name"`
	FileType string `json:"file_type"`
	FileSize int64  `json:"file_size"`
}

type MediaAssetResponse struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	Duration int    `json:"duration,omitempty"` // seconds
	FileSize int64  `json:"file_size"`
	Width    int    `json:"width,omitempty"`
	Height   int    `json:"height,omitempty"`
}

type LessonMediaResponse struct {
	LessonID string                         `json:"lesson_id"`
	Media    map[string]*MediaAssetResponse `json:"media"` // key: video, subtitle, thumbnail
}

// Batch Upload DTOs
type BatchMediaUploadRequest struct {
	LessonID string `json:"lesson_id" binding:"required"`
	// Files will be handled separately in multipart form
}

type BatchMediaUploadResponse struct {
	LessonID      string                `json:"lesson_id"`
	UploadedFiles []MediaUploadResponse `json:"uploaded_files"`
	Errors        []string              `json:"errors,omitempty"`
}

// Media Processing DTOs
type MediaProcessingStatus struct {
	MediaAssetID string `json:"media_asset_id"`
	Status       string `json:"status"`   // pending, processing, completed, failed
	Progress     int    `json:"progress"` // 0-100
	Message      string `json:"message,omitempty"`
}

// Video Metadata DTOs
type VideoMetadata struct {
	Duration  int     `json:"duration"`   // seconds
	Width     int     `json:"width"`      // pixels
	Height    int     `json:"height"`     // pixels
	Bitrate   int     `json:"bitrate"`    // kbps
	FrameRate float64 `json:"frame_rate"` // fps
	Format    string  `json:"format"`     // mp4, mov, etc.
	Size      int64   `json:"size"`       // bytes
}

// Subtitle DTOs
type SubtitleCue struct {
	StartTime string `json:"start_time"` // "00:00:10.500"
	EndTime   string `json:"end_time"`   // "00:00:13.000"
	Text      string `json:"text"`
}

type SubtitleTrack struct {
	Language string        `json:"language"` // "vi", "en"
	Label    string        `json:"label"`    // "Tiếng Việt", "English"
	Cues     []SubtitleCue `json:"cues"`
}

// Enhanced Lesson Response with Media
type EnhancedLessonResponse struct {
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

	// Media Assets (detailed)
	MediaAssets *LessonMediaResponse `json:"media_assets,omitempty"`
}

// Content Management DTOs
type ContentPackageInfo struct {
	TotalVideos     int    `json:"total_videos"`
	TotalSubtitles  int    `json:"total_subtitles"`
	TotalThumbnails int    `json:"total_thumbnails"`
	TotalSize       int64  `json:"total_size_bytes"`
	PackageVersion  string `json:"package_version"`
}

type MediaValidationResult struct {
	IsValid    bool     `json:"is_valid"`
	Errors     []string `json:"errors,omitempty"`
	Warnings   []string `json:"warnings,omitempty"`
	Duration   int      `json:"duration,omitempty"`
	Resolution string   `json:"resolution,omitempty"`
	FileSize   int64    `json:"file_size"`
}
