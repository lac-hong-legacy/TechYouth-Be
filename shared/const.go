package shared

const (
	UserID = "user_id"

	RarityCommon    = "common"
	RarityRare      = "rare"
	RarityLegendary = "legendary"

	QuestionTypeMultipleChoice = "multiple_choice"
	QuestionTypeDragDrop       = "drag_drop"
	QuestionTypeFillBlank      = "fill_blank"
	QuestionTypeConnect        = "connect"

	CacheKeyPrefix    = "techyouth:"
	CacheKeyUser      = CacheKeyPrefix + "user:"
	CacheKeyAuth      = CacheKeyPrefix + "auth:"
	CacheKeyContent   = CacheKeyPrefix + "content:"
	CacheKeySession   = CacheKeyPrefix + "session:"
	CacheKeyRateLimit = CacheKeyPrefix + "rate_limit:"
	CacheKeyGuest     = CacheKeyPrefix + "guest:"

	DefaultCacheTTL   = 3600
	AuthCacheTTL      = 1800
	SessionCacheTTL   = 7200
	RateLimitCacheTTL = 60

	MediaTypeVideo           = "video"
	MediaTypeSubtitle        = "subtitle"
	MediaTypeThumbnail       = "thumbnail"
	MediaTypeAudio           = "audio"
	MediaTypeBackgroundMusic = "background_music"
	MediaTypeVoiceOver       = "voice_over"
	MediaTypeAnimation       = "animation"
	MediaTypeIllustration    = "illustration"
)
