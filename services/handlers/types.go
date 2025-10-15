package handlers

import (
	"mime/multipart"

	"github.com/gofiber/fiber/v2"
	"github.com/lac-hong-legacy/ven_api/dto"
	"github.com/lac-hong-legacy/ven_api/model"
)

type AuthServiceInterface interface {
	Register(req dto.RegisterRequest) (*dto.RegisterResponse, error)
	Login(req dto.LoginRequest, clientIP, userAgent string) (*dto.LoginResponse, error)
	RefreshToken(req dto.RefreshTokenRequest, clientIP, userAgent string) (*dto.LoginResponse, error)
	Logout(userID, sessionID, accessToken, clientIP, userAgent string) error
	LogoutAllDevices(userID, sessionID, accessToken, clientIP, userAgent string) error
	VerifyEmail(email, code string) error
	ResendVerificationEmail(email string) error
	ForgotPassword(email string) error
	ResetPassword(req dto.ResetPasswordRequest) error
	ChangePassword(userID string, req dto.ChangePasswordRequest) error
	GetUserDevices(userID string) ([]dto.DeviceInfo, error)
	UpdateDeviceTrust(userID, deviceID string, trust bool) error
	RemoveDevice(userID, deviceID string) error
	RequiredAuth() fiber.Handler
	RequireRole(role string) fiber.Handler
}

type JWTServiceInterface interface {
	ExtractTokenFromHeader(authHeader string) (string, error)
	VerifyJWTToken(token string) (string, error)
}

type UserServiceInterface interface {
	CheckUsernameAvailability(username string) (bool, error)
	GetUserProfile(userID string) (*dto.UserProfileResponse, error)
	UpdateUserProfile(userID string, req dto.UpdateProfileRequest) (*dto.UserProfileResponse, error)
	InitializeUserProfile(userID string, birthYear int) error
	GetUserProgress(userID string) (*dto.UserProgressResponse, error)
	GetUserCollection(userID string) (*dto.CollectionResponse, error)
	CheckLessonAccess(userID, lessonID string) (*dto.LessonAccessResponse, error)
	CompleteLesson(userID, lessonID string, score, timeSpent int) error
	GetHeartStatus(userID string) (*dto.HeartStatusResponse, error)
	AddHearts(userID, source string, amount int) (*dto.HeartStatusResponse, error)
	LoseHeart(userID string) (*dto.HeartStatusResponse, error)
	GetUserSessions(userID, currentSessionID string) (*dto.SessionListResponse, error)
	RevokeUserSession(userID, sessionID string) error
	GetSecuritySettings(userID string) (*dto.SecuritySettings, error)
	UpdateSecuritySettings(userID string, req dto.UpdateSecuritySettingsRequest) (*dto.SecuritySettings, error)
	GetUserAuditLogs(userID string, page, limit int) (*dto.AuditLogResponse, error)
	CreateShareContent(userID string, req dto.ShareRequest) (*dto.ShareResponse, error)
	GetWeeklyLeaderboard(limit int, userID string) (*dto.LeaderboardResponse, error)
	GetMonthlyLeaderboard(limit int, userID string) (*dto.LeaderboardResponse, error)
	GetAllTimeLeaderboard(limit int, userID string) (*dto.LeaderboardResponse, error)
	AdminGetUsers(page, limit int, search string) (*dto.AdminUserListResponse, error)
	AdminUpdateUser(userID string, req dto.AdminUpdateUserRequest) (*dto.AdminUserInfo, error)
	AdminDeleteUser(userID string) error
}

type GuestServiceInterface interface {
	CreateOrGetSession(deviceID string) (*model.GuestSession, error)
	CanAccessLesson(sessionID, lessonID string) (bool, string, error)
	CompleteLesson(sessionID, lessonID string, score, timeSpent int) error
	AddHeartsFromAd(sessionID string) error
	LoseHeart(sessionID string) error
}

type ContentServiceInterface interface {
	GetTimeline() (*dto.TimelineCollectionResponse, error)
	GetCharacters(dynasty, rarity string) (*dto.CharacterCollectionResponse, error)
	GetCharacterDetails(characterID string) (*dto.CharacterResponse, error)
	GetCharacterLessons(characterID string) ([]dto.LessonResponse, error)
	GetLessonContent(lessonID string) (*dto.LessonResponse, error)
	ValidateLessonAnswers(lessonID string, userAnswers map[string]interface{}) (*dto.ValidateLessonResponse, error)
	SearchContent(req dto.SearchRequest) (*dto.SearchResponse, error)
	SubmitQuestionAnswer(userID, lessonID, questionID string, answer interface{}) (*dto.SubmitQuestionAnswerResponse, error)
	CheckLessonStatus(userID, lessonID string) (*dto.CheckLessonStatusResponse, error)
	GetEras() ([]string, error)
	GetDynasties() ([]string, error)
	CreateCharacter(character *model.Character) (*dto.CharacterResponse, error)
	CreateLessonFromRequest(req dto.CreateLessonRequest) (*dto.LessonResponse, error)
	UpdateLessonScript(lessonID, script string) (*model.Lesson, error)
	GetLessonProductionStatus(lessonID string) (*dto.LessonProductionStatusResponse, error)
	MapLessonToResponse(lesson *model.Lesson) dto.LessonResponse
	MarkAudioUploaded(lessonID string) error
	MarkAnimationUploaded(lessonID string) error
	GetProgress(sessionID string) (*model.GuestProgress, error)
}

type MediaServiceInterface interface {
	UploadLessonSubtitle(lessonID string, file *multipart.FileHeader) (*dto.MediaUploadResponse, error)
	UploadThumbnail(lessonID string, file *multipart.FileHeader) (*dto.MediaUploadResponse, error)
	GetLessonMedia(lessonID string) (*dto.LessonMediaResponse, error)
	DeleteMediaAsset(assetID string) error
	UploadLessonAudio(lessonID string, file *multipart.FileHeader) (*dto.MediaUploadResponse, error)
	UploadLessonAnimation(lessonID string, file *multipart.FileHeader) (*dto.MediaUploadResponse, error)
	GetMediaStatistics() (map[string]interface{}, error)
}
