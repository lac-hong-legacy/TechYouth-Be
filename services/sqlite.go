package services

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lac-hong-legacy/ven_api/dto"
	"github.com/lac-hong-legacy/ven_api/model"
	"golang.org/x/crypto/bcrypt"

	"github.com/alphabatem/common/context"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type SqliteService struct {
	context.DefaultService
	db *gorm.DB

	database string
}

const SQLITE_SVC = "sqlite_svc"

func (ds SqliteService) Id() string {
	return SQLITE_SVC
}

func (ds SqliteService) Db() *gorm.DB {
	return ds.db
}

func (ds *SqliteService) Configure(ctx *context.Context) error {
	ds.database = os.Getenv("DB_NAME")

	return ds.DefaultService.Configure(ctx)
}

func (ds *SqliteService) Start() (err error) {
	ds.db, err = gorm.Open(sqlite.Open(ds.database), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		return err
	}

	models := []interface{}{
		// Existing models
		&model.User{},
		&model.GuestSession{},
		&model.GuestProgress{},
		&model.GuestLessonAttempt{},
		&model.RateLimit{},
		&model.RateLimitConfig{},

		// Content models
		&model.Character{},
		&model.Lesson{},
		&model.Timeline{},

		// User progress models
		&model.UserProgress{},
		&model.Spirit{},
		&model.Achievement{},
		&model.UserAchievement{},
		&model.UserLessonAttempt{},

		// New authentication models
		&model.UserSession{},
		&model.AuthAuditLog{},
		&model.PasswordResetCode{},
		&model.BlacklistedToken{},
		&model.TrustedDevice{},
		&model.LoginAttempt{},
	}

	err = ds.db.AutoMigrate(models...)
	if err != nil {
		log.Printf("Failed to migrate database: %v", err)
		return err
	}

	err = ds.seedInitialData()
	if err != nil {
		log.Printf("Failed to seed initial data: %v", err)
		return err
	}

	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for range ticker.C {
			err := ds.CleanupExpiredData()
			if err != nil {
				log.Printf("Failed to cleanup expired data: %v", err)
			}
		}
	}()

	log.Println("Database connected and migrated successfully")
	return nil
}

func (ds *SqliteService) Shutdown() {
}

func (ds *SqliteService) HandleError(err error) error {
	if err == nil {
		return nil
	}

	var statusCode int
	var errorType string

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		statusCode = http.StatusNotFound // 404
		errorType = "NOT_FOUND"
	case errors.Is(err, gorm.ErrDuplicatedKey):
		statusCode = http.StatusConflict // 409
		errorType = "CONFLICT"
	case errors.Is(err, gorm.ErrForeignKeyViolated):
		statusCode = http.StatusBadRequest // 400
		errorType = "FOREIGN_KEY_VIOLATION"
	case errors.Is(err, gorm.ErrInvalidTransaction):
		statusCode = http.StatusInternalServerError // 500
		errorType = "TRANSACTION_ERROR"
	default:
		// Check for SQLite-specific errors
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			statusCode = http.StatusConflict // 409
			errorType = "UNIQUE_CONSTRAINT"
		} else if strings.Contains(err.Error(), "no such table") {
			statusCode = http.StatusInternalServerError // 500
			errorType = "SCHEMA_ERROR"
		} else {
			statusCode = http.StatusInternalServerError // 500
			errorType = "INTERNAL_ERROR"
		}
	}

	logEntry := log.WithFields(log.Fields{
		"status_code": statusCode,
		"error_type":  errorType,
		"error":       err.Error(),
	})

	if statusCode >= 500 {
		logEntry.Error("Database error occurred")
	} else {
		logEntry.Warn("Database operation failed")
	}

	return fmt.Errorf("%s: %w", errorType, err)
}

func (ds *SqliteService) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	if err := ds.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (ds *SqliteService) GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	if err := ds.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (ds *SqliteService) GetUserByEmailOrUsername(emailOrUsername string) (*model.User, error) {
	var user model.User
	if err := ds.db.Where("email = ? OR username = ?", emailOrUsername, emailOrUsername).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (ds *SqliteService) GetSessionByDeviceID(deviceID string) (*model.GuestSession, error) {
	var session model.GuestSession
	if err := ds.db.Where("device_id = ?", deviceID).First(&session).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return &session, nil
}

func (ds *SqliteService) CreateSession(session *model.GuestSession) (*model.GuestSession, error) {
	id, _ := uuid.NewV7()
	session.ID = id.String()
	if err := ds.db.Create(session).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return session, nil
}

func (ds *SqliteService) UpdateSession(session *model.GuestSession) error {
	if err := ds.db.Save(session).Error; err != nil {
		return ds.HandleError(err)
	}
	return nil
}

func (ds *SqliteService) GetProgress(sessionID string) (*model.GuestProgress, error) {
	var progress model.GuestProgress
	if err := ds.db.Where("guest_session_id = ?", sessionID).First(&progress).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return &progress, nil
}

func (ds *SqliteService) CreateProgress(progress *model.GuestProgress) (*model.GuestProgress, error) {
	id, _ := uuid.NewV7()
	progress.ID = id.String()
	if err := ds.db.Create(progress).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return progress, nil
}

func (ds *SqliteService) UpdateProgress(progress *model.GuestProgress) error {
	if err := ds.db.Save(progress).Error; err != nil {
		return ds.HandleError(err)
	}
	return nil
}

func (ds *SqliteService) CreateLessonAttempt(attempt *model.GuestLessonAttempt) error {
	id, _ := uuid.NewV7()
	attempt.ID = id.String()
	if err := ds.db.Create(attempt).Error; err != nil {
		return ds.HandleError(err)
	}
	return nil
}

func DeactivateSession(sessionID string) error {
	// Placeholder for session deactivation logic
	return nil
}

func (s *SqliteService) GetRateLimit(identifier, endpointType string) (*model.RateLimit, error) {
	var rateLimit model.RateLimit

	err := s.db.Where("identifier = ? AND endpoint_type = ?", identifier, endpointType).First(&rateLimit).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &rateLimit, nil
}

func (s *SqliteService) SaveRateLimit(rateLimit *model.RateLimit) error {
	// Generate ID if not set
	if rateLimit.ID == "" {
		id, _ := uuid.NewV7()
		rateLimit.ID = id.String()
	}

	// Set timestamps if not set
	now := time.Now()
	if rateLimit.CreatedAt.IsZero() {
		rateLimit.CreatedAt = now
	}
	rateLimit.UpdatedAt = now

	// Use GORM's Save method which will INSERT or UPDATE based on primary key
	if err := s.db.Save(rateLimit).Error; err != nil {
		return err
	}
	return nil
}

func (s *SqliteService) UpdateRateLimit(rateLimit *model.RateLimit) error {
	// Update specific fields using GORM's Updates method
	err := s.db.Model(rateLimit).Where("id = ?", rateLimit.ID).Updates(map[string]interface{}{
		"request_count": rateLimit.RequestCount,
		"blocked_until": rateLimit.BlockedUntil,
		"updated_at":    rateLimit.UpdatedAt,
	}).Error

	return err
}

// Cleanup old rate limit records
func (s *SqliteService) CleanupOldRecords() error {
	// Remove records older than 7 days and not currently blocked
	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	now := time.Now()

	err := s.db.Where("created_at < ? AND (blocked_until IS NULL OR blocked_until < ?)", cutoff, now).
		Delete(&model.RateLimit{}).Error

	return err
}

// ==================== CHARACTER METHODS ====================

func (ds *SqliteService) CreateCharacter(character *model.Character) (*model.Character, error) {
	if character.ID == "" {
		id, _ := uuid.NewV7()
		character.ID = id.String()
	}
	character.CreatedAt = time.Now()
	character.UpdatedAt = time.Now()

	if err := ds.db.Create(character).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return character, nil
}

func (ds *SqliteService) GetCharacter(id string) (*model.Character, error) {
	var character model.Character
	if err := ds.db.Where("id = ?", id).First(&character).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return &character, nil
}

func (ds *SqliteService) GetCharactersByDynasty(dynasty string) ([]model.Character, error) {
	var characters []model.Character
	query := ds.db.Model(&model.Character{})

	if dynasty != "" {
		query = query.Where("dynasty = ?", dynasty)
	}

	if err := query.Find(&characters).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return characters, nil
}

func (ds *SqliteService) GetCharactersByRarity(rarity string) ([]model.Character, error) {
	var characters []model.Character
	if err := ds.db.Where("rarity = ?", rarity).Find(&characters).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return characters, nil
}

func (ds *SqliteService) UpdateCharacter(character *model.Character) error {
	character.UpdatedAt = time.Now()
	if err := ds.db.Save(character).Error; err != nil {
		return ds.HandleError(err)
	}
	return nil
}

// ==================== LESSON METHODS ====================

func (ds *SqliteService) CreateLesson(lesson *model.Lesson) (*model.Lesson, error) {
	if lesson.ID == "" {
		id, _ := uuid.NewV7()
		lesson.ID = id.String()
	}
	lesson.CreatedAt = time.Now()
	lesson.UpdatedAt = time.Now()

	if err := ds.db.Create(lesson).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return lesson, nil
}

func (ds *SqliteService) GetLesson(id string) (*model.Lesson, error) {
	var lesson model.Lesson
	if err := ds.db.Preload("Character").Where("id = ?", id).First(&lesson).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return &lesson, nil
}

func (ds *SqliteService) GetLessonsByCharacter(characterID string) ([]model.Lesson, error) {
	var lessons []model.Lesson
	if err := ds.db.Preload("Character").Where("character_id = ? AND is_active = ?", characterID, true).
		Order("\"order\" ASC").Find(&lessons).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return lessons, nil
}

func (ds *SqliteService) UpdateLesson(lesson *model.Lesson) error {
	lesson.UpdatedAt = time.Now()
	if err := ds.db.Save(lesson).Error; err != nil {
		return ds.HandleError(err)
	}
	return nil
}

// ==================== TIMELINE METHODS ====================

func (ds *SqliteService) CreateTimeline(timeline *model.Timeline) (*model.Timeline, error) {
	if timeline.ID == "" {
		id, _ := uuid.NewV7()
		timeline.ID = id.String()
	}
	timeline.CreatedAt = time.Now()
	timeline.UpdatedAt = time.Now()

	if err := ds.db.Create(timeline).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return timeline, nil
}

func (ds *SqliteService) GetTimeline() ([]model.Timeline, error) {
	var timelines []model.Timeline
	if err := ds.db.Order("\"order\" ASC").Find(&timelines).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return timelines, nil
}

func (ds *SqliteService) GetTimelineByEra(era string) ([]model.Timeline, error) {
	var timelines []model.Timeline
	if err := ds.db.Where("era = ?", era).Order("\"order\" ASC").Find(&timelines).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return timelines, nil
}

// ==================== USER PROGRESS METHODS ====================

func (ds *SqliteService) CreateUserProgress(progress *model.UserProgress) (*model.UserProgress, error) {
	if progress.ID == "" {
		id, _ := uuid.NewV7()
		progress.ID = id.String()
	}
	progress.CreatedAt = time.Now()
	progress.UpdatedAt = time.Now()

	if err := ds.db.Create(progress).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return progress, nil
}

func (ds *SqliteService) GetUserProgress(userID string) (*model.UserProgress, error) {
	var progress model.UserProgress
	if err := ds.db.Where("user_id = ?", userID).First(&progress).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return &progress, nil
}

func (ds *SqliteService) UpdateUserProgress(progress *model.UserProgress) error {
	progress.UpdatedAt = time.Now()
	if err := ds.db.Save(progress).Error; err != nil {
		return ds.HandleError(err)
	}
	return nil
}

func (ds *SqliteService) GetUsersForHeartReset(since time.Time) ([]model.UserProgress, error) {
	var users []model.UserProgress
	if err := ds.db.Where("last_heart_reset < ? OR last_heart_reset IS NULL", since).
		Find(&users).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return users, nil
}

// ==================== SPIRIT METHODS ====================

func (ds *SqliteService) CreateSpirit(spirit *model.Spirit) (*model.Spirit, error) {
	if spirit.ID == "" {
		id, _ := uuid.NewV7()
		spirit.ID = id.String()
	}
	spirit.CreatedAt = time.Now()
	spirit.UpdatedAt = time.Now()

	if err := ds.db.Create(spirit).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return spirit, nil
}

func (ds *SqliteService) GetUserSpirit(userID string) (*model.Spirit, error) {
	var spirit model.Spirit
	if err := ds.db.Where("user_id = ?", userID).First(&spirit).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return &spirit, nil
}

func (ds *SqliteService) UpdateSpirit(spirit *model.Spirit) error {
	spirit.UpdatedAt = time.Now()
	if err := ds.db.Save(spirit).Error; err != nil {
		return ds.HandleError(err)
	}
	return nil
}

// ==================== ACHIEVEMENT METHODS ====================

func (ds *SqliteService) CreateAchievement(achievement *model.Achievement) (*model.Achievement, error) {
	if achievement.ID == "" {
		id, _ := uuid.NewV7()
		achievement.ID = id.String()
	}
	achievement.CreatedAt = time.Now()
	achievement.UpdatedAt = time.Now()

	if err := ds.db.Create(achievement).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return achievement, nil
}

func (ds *SqliteService) GetActiveAchievements() ([]model.Achievement, error) {
	var achievements []model.Achievement
	if err := ds.db.Where("is_active = ?", true).Find(&achievements).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return achievements, nil
}

func (ds *SqliteService) CreateUserAchievement(userAchievement *model.UserAchievement) error {
	if userAchievement.ID == "" {
		id, _ := uuid.NewV7()
		userAchievement.ID = id.String()
	}
	userAchievement.CreatedAt = time.Now()
	userAchievement.UnlockedAt = time.Now()

	if err := ds.db.Create(userAchievement).Error; err != nil {
		return ds.HandleError(err)
	}
	return nil
}

func (ds *SqliteService) GetUserAchievements(userID string) ([]model.UserAchievement, error) {
	var userAchievements []model.UserAchievement
	if err := ds.db.Preload("Achievement").Where("user_id = ?", userID).
		Find(&userAchievements).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return userAchievements, nil
}

// ==================== LEADERBOARD METHODS ====================

func (ds *SqliteService) GetWeeklyLeaderboard(limit int) ([]model.UserProgress, error) {
	var users []model.UserProgress
	weekAgo := time.Now().AddDate(0, 0, -7)

	if err := ds.db.Where("updated_at >= ?", weekAgo).
		Order("xp DESC").Limit(limit).Find(&users).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return users, nil
}

func (ds *SqliteService) GetMonthlyLeaderboard(limit int) ([]model.UserProgress, error) {
	var users []model.UserProgress
	monthAgo := time.Now().AddDate(0, -1, 0)

	if err := ds.db.Where("updated_at >= ?", monthAgo).
		Order("xp DESC").Limit(limit).Find(&users).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return users, nil
}

func (ds *SqliteService) GetAllTimeLeaderboard(limit int) ([]model.UserProgress, error) {
	var users []model.UserProgress
	if err := ds.db.Order("xp DESC").Limit(limit).Find(&users).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return users, nil
}

func (ds *SqliteService) GetUserRank(userID string) (int, error) {
	var rank int64
	userProgress, err := ds.GetUserProgress(userID)
	if err != nil {
		return 0, err
	}

	if err := ds.db.Model(&model.UserProgress{}).
		Where("xp > ?", userProgress.XP).Count(&rank).Error; err != nil {
		return 0, ds.HandleError(err)
	}

	return int(rank + 1), nil // +1 because rank is 0-indexed
}

// ==================== CONTENT SEARCH AND FILTERING ====================

func (ds *SqliteService) SearchCharacters(query string, era string, dynasty string, rarity string, limit int) ([]model.Character, error) {
	var characters []model.Character
	dbQuery := ds.db.Model(&model.Character{})

	if query != "" {
		dbQuery = dbQuery.Where("name LIKE ? OR description LIKE ?", "%"+query+"%", "%"+query+"%")
	}

	if era != "" {
		dbQuery = dbQuery.Where("era = ?", era)
	}

	if dynasty != "" {
		dbQuery = dbQuery.Where("dynasty = ?", dynasty)
	}

	if rarity != "" {
		dbQuery = dbQuery.Where("rarity = ?", rarity)
	}

	if limit > 0 {
		dbQuery = dbQuery.Limit(limit)
	}

	if err := dbQuery.Find(&characters).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return characters, nil
}

// ==================== ANALYTICS METHODS ====================

func (ds *SqliteService) HasUserUnlockedAchievement(userID, achievementID string) (bool, error) {
	var count int64
	if err := ds.db.Model(&model.UserAchievement{}).
		Where("user_id = ? AND achievement_id = ?", userID, achievementID).
		Count(&count).Error; err != nil {
		return false, ds.HandleError(err)
	}
	return count > 0, nil
}

// Add these methods to the existing sqlite_service_extensions

// ==================== MISSING USER METHODS ====================

func (ds *SqliteService) GetUser(userID string) (*model.User, error) {
	var user model.User
	if err := ds.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return &user, nil
}

func (ds *SqliteService) UpdateUser(user *model.User) error {
	user.UpdatedAt = time.Now()
	if err := ds.db.Save(user).Error; err != nil {
		return ds.HandleError(err)
	}
	return nil
}

// ==================== ACHIEVEMENT PRELOAD METHODS ====================

// Update the existing GetUserAchievements method to include Achievement preloading
func (ds *SqliteService) GetUserAchievementsWithDetails(userID string) ([]struct {
	UserAchievement model.UserAchievement
	Achievement     model.Achievement
}, error) {
	var results []struct {
		UserAchievement model.UserAchievement
		Achievement     model.Achievement
	}

	if err := ds.db.Table("user_achievements").
		Select("user_achievements.*, achievements.*").
		Joins("LEFT JOIN achievements ON user_achievements.achievement_id = achievements.id").
		Where("user_achievements.user_id = ?", userID).
		Scan(&results).Error; err != nil {
		return nil, ds.HandleError(err)
	}

	return results, nil
}

// ==================== ADVANCED QUERY METHODS ====================

func (ds *SqliteService) GetUserProgressWithJoins(userID string) (*model.UserProgress, error) {
	var progress model.UserProgress
	if err := ds.db.Where("user_id = ?", userID).First(&progress).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return &progress, nil
}

func (ds *SqliteService) GetCharactersWithLessonCount() ([]struct {
	Character   model.Character
	LessonCount int64
}, error) {
	var results []struct {
		Character   model.Character
		LessonCount int64
	}

	if err := ds.db.Table("characters").
		Select("characters.*, COUNT(lessons.id) as lesson_count").
		Joins("LEFT JOIN lessons ON characters.id = lessons.character_id").
		Group("characters.id").
		Scan(&results).Error; err != nil {
		return nil, ds.HandleError(err)
	}

	return results, nil
}

// ==================== BATCH OPERATIONS ====================

func (ds *SqliteService) BatchUpdateCharacterUnlockStatus(characterIDs []string, unlocked bool) error {
	if len(characterIDs) == 0 {
		return nil
	}

	if err := ds.db.Model(&model.Character{}).
		Where("id IN ?", characterIDs).
		Update("is_unlocked", unlocked).Error; err != nil {
		return ds.HandleError(err)
	}
	return nil
}

func (ds *SqliteService) GetMultipleCharacters(characterIDs []string) ([]model.Character, error) {
	var characters []model.Character
	if err := ds.db.Where("id IN ?", characterIDs).Find(&characters).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return characters, nil
}

// ==================== LESSON ATTEMPT TRACKING ====================

func (ds *SqliteService) CreateUserLessonAttempt(attempt *model.UserLessonAttempt) error {
	id, _ := uuid.NewV7()
	attempt.ID = id.String()
	attempt.CreatedAt = time.Now()
	attempt.UpdatedAt = time.Now()

	if err := ds.db.Create(attempt).Error; err != nil {
		return ds.HandleError(err)
	}
	return nil
}

func (ds *SqliteService) GetUserLessonAttempts(userID, lessonID string) ([]model.UserLessonAttempt, error) {
	var attempts []model.UserLessonAttempt
	if err := ds.db.Where("user_id = ? AND lesson_id = ?", userID, lessonID).
		Order("created_at DESC").Find(&attempts).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return attempts, nil
}

// ==================== ANALYTICS HELPERS ====================

func (ds *SqliteService) GetDailyActiveUsers(date time.Time) (int64, error) {
	var count int64
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	if err := ds.db.Model(&model.UserProgress{}).
		Where("last_activity_date >= ? AND last_activity_date < ?", startOfDay, endOfDay).
		Count(&count).Error; err != nil {
		return 0, ds.HandleError(err)
	}
	return count, nil
}

func (ds *SqliteService) GetLessonCompletionStats() (map[string]interface{}, error) {
	var totalLessons int64
	var totalCompletions int64

	// Get total lessons
	if err := ds.db.Model(&model.Lesson{}).Where("is_active = ?", true).Count(&totalLessons).Error; err != nil {
		return nil, ds.HandleError(err)
	}

	// Get total completions (this is a simplified count)
	if err := ds.db.Model(&model.UserProgress{}).Count(&totalCompletions).Error; err != nil {
		return nil, ds.HandleError(err)
	}

	stats := map[string]interface{}{
		"total_lessons":           totalLessons,
		"total_completions":       totalCompletions,
		"average_completion_rate": 0.0, // TODO: Calculate proper completion rate
	}

	return stats, nil
}

// ==================== MEDIA ASSET METHODS ====================

func (ds *SqliteService) CreateMediaAsset(asset *model.MediaAsset) error {
	if asset.ID == "" {
		id, _ := uuid.NewV7()
		asset.ID = id.String()
	}
	asset.CreatedAt = time.Now()
	asset.UpdatedAt = time.Now()

	if err := ds.db.Create(asset).Error; err != nil {
		return ds.HandleError(err)
	}
	return nil
}

func (ds *SqliteService) GetMediaAsset(id string) (*model.MediaAsset, error) {
	var asset model.MediaAsset
	if err := ds.db.Where("id = ?", id).First(&asset).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return &asset, nil
}

func (ds *SqliteService) UpdateMediaAsset(asset *model.MediaAsset) error {
	asset.UpdatedAt = time.Now()
	if err := ds.db.Save(asset).Error; err != nil {
		return ds.HandleError(err)
	}
	return nil
}

func (ds *SqliteService) DeleteMediaAsset(id string) error {
	// Delete related lesson media records first
	if err := ds.db.Where("media_asset_id = ?", id).Delete(&model.LessonMedia{}).Error; err != nil {
		return ds.HandleError(err)
	}

	// Delete the media asset
	if err := ds.db.Where("id = ?", id).Delete(&model.MediaAsset{}).Error; err != nil {
		return ds.HandleError(err)
	}
	return nil
}

func (ds *SqliteService) GetMediaAssetsByType(fileType string) ([]model.MediaAsset, error) {
	var assets []model.MediaAsset
	if err := ds.db.Where("file_type = ?", fileType).Find(&assets).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return assets, nil
}

func (ds *SqliteService) GetUnprocessedMediaAssets() ([]model.MediaAsset, error) {
	var assets []model.MediaAsset
	if err := ds.db.Where("is_processed = ?", false).Find(&assets).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return assets, nil
}

// ==================== LESSON MEDIA METHODS ====================

func (ds *SqliteService) CreateLessonMedia(lessonMedia *model.LessonMedia) error {
	if lessonMedia.ID == "" {
		id, _ := uuid.NewV7()
		lessonMedia.ID = id.String()
	}
	lessonMedia.CreatedAt = time.Now()

	if err := ds.db.Create(lessonMedia).Error; err != nil {
		return ds.HandleError(err)
	}
	return nil
}

func (ds *SqliteService) GetLessonMediaAssets(lessonID string) ([]model.LessonMedia, error) {
	var lessonMedia []model.LessonMedia
	if err := ds.db.Where("lesson_id = ? AND is_active = ?", lessonID, true).
		Preload("MediaAsset").
		Find(&lessonMedia).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return lessonMedia, nil
}

func (ds *SqliteService) GetLessonMediaByType(lessonID, mediaType string) (*model.LessonMedia, error) {
	var lessonMedia model.LessonMedia
	if err := ds.db.Where("lesson_id = ? AND media_type = ? AND is_active = ?", lessonID, mediaType, true).
		Preload("MediaAsset").
		First(&lessonMedia).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return &lessonMedia, nil
}

func (ds *SqliteService) UpdateLessonMedia(lessonMedia *model.LessonMedia) error {
	if err := ds.db.Save(lessonMedia).Error; err != nil {
		return ds.HandleError(err)
	}
	return nil
}

func (ds *SqliteService) DeleteLessonMedia(id string) error {
	if err := ds.db.Where("id = ?", id).Delete(&model.LessonMedia{}).Error; err != nil {
		return ds.HandleError(err)
	}
	return nil
}

func (ds *SqliteService) DeactivateLessonMediaByType(lessonID, mediaType string) error {
	if err := ds.db.Model(&model.LessonMedia{}).
		Where("lesson_id = ? AND media_type = ?", lessonID, mediaType).
		Update("is_active", false).Error; err != nil {
		return ds.HandleError(err)
	}
	return nil
}

// ==================== USER QUESTION ANSWER METHODS ====================

func (ds *SqliteService) SaveUserQuestionAnswer(answer *model.UserQuestionAnswer) error {
	if answer.ID == "" {
		id, _ := uuid.NewV7()
		answer.ID = id.String()
	}
	answer.CreatedAt = time.Now()
	answer.UpdatedAt = time.Now()

	// Check if answer already exists for this user/lesson/question
	var existing model.UserQuestionAnswer
	err := ds.db.Where("user_id = ? AND lesson_id = ? AND question_id = ?",
		answer.UserID, answer.LessonID, answer.QuestionID).First(&existing).Error

	if err == nil {
		// Update existing answer
		existing.Answer = answer.Answer
		existing.IsCorrect = answer.IsCorrect
		existing.Points = answer.Points
		existing.UpdatedAt = time.Now()
		return ds.db.Save(&existing).Error
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create new answer
		return ds.db.Create(answer).Error
	}

	return ds.HandleError(err)
}

func (ds *SqliteService) GetUserQuestionAnswers(userID, lessonID string) ([]model.UserQuestionAnswer, error) {
	var answers []model.UserQuestionAnswer
	if err := ds.db.Where("user_id = ? AND lesson_id = ?", userID, lessonID).
		Find(&answers).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return answers, nil
}

func (ds *SqliteService) GetUserQuestionAnswer(userID, lessonID, questionID string) (*model.UserQuestionAnswer, error) {
	var answer model.UserQuestionAnswer
	if err := ds.db.Where("user_id = ? AND lesson_id = ? AND question_id = ?",
		userID, lessonID, questionID).First(&answer).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return &answer, nil
}

func (ds *SqliteService) DeleteUserQuestionAnswers(userID, lessonID string) error {
	if err := ds.db.Where("user_id = ? AND lesson_id = ?", userID, lessonID).
		Delete(&model.UserQuestionAnswer{}).Error; err != nil {
		return ds.HandleError(err)
	}
	return nil
}

// ==================== ENHANCED LESSON METHODS WITH MEDIA ====================

func (ds *SqliteService) GetLessonWithMedia(id string) (*model.Lesson, []model.LessonMedia, error) {
	var lesson model.Lesson
	if err := ds.db.Where("id = ?", id).Preload("Character").First(&lesson).Error; err != nil {
		return nil, nil, ds.HandleError(err)
	}

	mediaAssets, err := ds.GetLessonMediaAssets(id)
	if err != nil {
		return &lesson, nil, err
	}

	return &lesson, mediaAssets, nil
}

func (ds *SqliteService) GetLessonsWithMediaByCharacter(characterID string) ([]model.Lesson, map[string][]model.LessonMedia, error) {
	var lessons []model.Lesson
	if err := ds.db.Where("character_id = ? AND is_active = ?", characterID, true).
		Order("\"order\" ASC").
		Find(&lessons).Error; err != nil {
		return nil, nil, ds.HandleError(err)
	}

	// Get media for all lessons
	mediaMap := make(map[string][]model.LessonMedia)
	for _, lesson := range lessons {
		media, err := ds.GetLessonMediaAssets(lesson.ID)
		if err != nil {
			log.Printf("Failed to get media for lesson %s: %v", lesson.ID, err)
			continue
		}
		mediaMap[lesson.ID] = media
	}

	return lessons, mediaMap, nil
}

// ==================== MEDIA STATISTICS METHODS ====================

func (ds *SqliteService) GetMediaStatistics() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Count total media assets by type
	var videoCount, subtitleCount, thumbnailCount int64

	ds.db.Model(&model.MediaAsset{}).Where("file_type = ?", "video").Count(&videoCount)
	ds.db.Model(&model.MediaAsset{}).Where("file_type = ?", "subtitle").Count(&subtitleCount)
	ds.db.Model(&model.MediaAsset{}).Where("file_type = ?", "thumbnail").Count(&thumbnailCount)

	stats["total_videos"] = videoCount
	stats["total_subtitles"] = subtitleCount
	stats["total_thumbnails"] = thumbnailCount

	// Calculate total storage used
	var totalSize int64
	ds.db.Model(&model.MediaAsset{}).Select("COALESCE(SUM(file_size), 0)").Scan(&totalSize)
	stats["total_storage_bytes"] = totalSize
	stats["total_storage_mb"] = totalSize / (1024 * 1024)

	// Count lessons with media
	var lessonsWithVideo, lessonsWithSubtitles int64
	ds.db.Model(&model.LessonMedia{}).
		Where("media_type = ? AND is_active = ?", "video", true).
		Count(&lessonsWithVideo)
	ds.db.Model(&model.LessonMedia{}).
		Where("media_type = ? AND is_active = ?", "subtitle", true).
		Count(&lessonsWithSubtitles)

	stats["lessons_with_video"] = lessonsWithVideo
	stats["lessons_with_subtitles"] = lessonsWithSubtitles

	// Count unprocessed media
	var unprocessedCount int64
	ds.db.Model(&model.MediaAsset{}).Where("is_processed = ?", false).Count(&unprocessedCount)
	stats["unprocessed_media"] = unprocessedCount

	return stats, nil
}

func (ds *SqliteService) GetLessonsWithoutMedia() ([]model.Lesson, error) {
	var lessons []model.Lesson

	// Find lessons that don't have any active media
	subQuery := ds.db.Model(&model.LessonMedia{}).
		Select("lesson_id").
		Where("is_active = ?", true)

	if err := ds.db.Where("id NOT IN (?)", subQuery).
		Preload("Character").
		Find(&lessons).Error; err != nil {
		return nil, ds.HandleError(err)
	}

	return lessons, nil
}

// ==================== BULK OPERATIONS ====================

func (ds *SqliteService) BulkCreateMediaAssets(assets []model.MediaAsset) error {
	if len(assets) == 0 {
		return nil
	}

	// Set IDs and timestamps
	now := time.Now()
	for i := range assets {
		if assets[i].ID == "" {
			id, _ := uuid.NewV7()
			assets[i].ID = id.String()
		}
		assets[i].CreatedAt = now
		assets[i].UpdatedAt = now
	}

	if err := ds.db.CreateInBatches(assets, 100).Error; err != nil {
		return ds.HandleError(err)
	}
	return nil
}

func (ds *SqliteService) BulkCreateLessonMedia(lessonMedia []model.LessonMedia) error {
	if len(lessonMedia) == 0 {
		return nil
	}

	// Set IDs and timestamps
	now := time.Now()
	for i := range lessonMedia {
		if lessonMedia[i].ID == "" {
			id, _ := uuid.NewV7()
			lessonMedia[i].ID = id.String()
		}
		lessonMedia[i].CreatedAt = now
	}

	if err := ds.db.CreateInBatches(lessonMedia, 100).Error; err != nil {
		return ds.HandleError(err)
	}
	return nil
}

// Add these methods to your existing SqliteService

// ==================== ENHANCED USER METHODS ====================

func (ds *SqliteService) CreateUser(req dto.RegisterRequest, verificationCode string) (*model.User, error) {
	codeExpiry := time.Now().Add(15 * time.Minute) // Code expires in 15 minutes
	user := &model.User{
		ID:                     uuid.New().String(),
		Username:               req.Username,
		Email:                  req.Email,
		Password:               req.Password,
		Role:                   model.RoleUser,
		IsActive:               true,
		EmailVerified:          false,
		VerificationCode:       verificationCode,
		VerificationCodeExpiry: &codeExpiry,
		FailedAttempts:         0,
		LoginNotifications:     true,
		SessionTimeout:         1440, // 24 hours
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}

	if err := ds.db.Create(user).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return user, nil
}

func (ds *SqliteService) GetUserByID(userID string) (*model.User, error) {
	var user model.User
	err := ds.db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, ds.HandleError(err)
	}
	return &user, nil
}

func (ds *SqliteService) GetUserByVerificationCode(email, code string) (*model.User, error) {
	var user model.User
	err := ds.db.Where("email = ? AND verification_code = ?", email, code).First(&user).Error
	if err != nil {
		return nil, ds.HandleError(err)
	}
	return &user, nil
}

func (ds *SqliteService) UpdateUserPassword(userID, hashedPassword string) error {
	now := time.Now()
	return ds.db.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"password":             hashedPassword,
		"last_password_change": &now,
		"updated_at":           now,
	}).Error
}

func (ds *SqliteService) UpdateLastLogin(userID, ip string) error {
	now := time.Now()
	return ds.db.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"last_login_at": &now,
		"last_login_ip": ip,
		"updated_at":    now,
	}).Error
}

func (ds *SqliteService) IncrementFailedAttempts(userID string) error {
	return ds.db.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"failed_attempts": gorm.Expr("failed_attempts + 1"),
		"updated_at":      time.Now(),
	}).Error
}

func (ds *SqliteService) ResetFailedAttempts(userID string) error {
	return ds.db.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"failed_attempts": 0,
		"locked_until":    nil,
		"updated_at":      time.Now(),
	}).Error
}

func (ds *SqliteService) LockAccount(userID string, lockUntil time.Time) error {
	return ds.db.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"locked_until": &lockUntil,
		"updated_at":   time.Now(),
	}).Error
}

func (ds *SqliteService) VerifyUserEmail(userID string) error {
	return ds.db.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"email_verified":           true,
		"verification_code":        nil,
		"verification_code_expiry": nil,
		"updated_at":               time.Now(),
	}).Error
}

func (ds *SqliteService) UpdateVerificationCode(userID, code string) error {
	codeExpiry := time.Now().Add(15 * time.Minute) // Code expires in 15 minutes
	return ds.db.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"verification_code":        code,
		"verification_code_expiry": &codeExpiry,
		"updated_at":               time.Now(),
	}).Error
}

func (ds *SqliteService) IsUsernameAvailable(username string) (bool, error) {
	var count int64
	err := ds.db.Model(&model.User{}).Where("LOWER(username) = LOWER(?) AND deleted_at IS NULL", username).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func (ds *SqliteService) IsEmailAvailable(email string) (bool, error) {
	var count int64
	err := ds.db.Model(&model.User{}).Where("LOWER(email) = LOWER(?) AND deleted_at IS NULL", email).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

// ==================== USER SESSION METHODS ====================

func (ds *SqliteService) CreateUserSession(session dto.UserSession) (string, error) {
	dbSession := &model.UserSession{
		ID:        uuid.New().String(),
		UserID:    session.UserID,
		TokenHash: session.TokenHash,
		DeviceID:  session.DeviceID,
		IP:        session.IP,
		UserAgent: session.UserAgent,
		CreatedAt: session.CreatedAt,
		LastUsed:  session.LastUsed,
		IsActive:  session.IsActive,
		ExpiresAt: session.CreatedAt.Add(7 * 24 * time.Hour), // 7 days
	}

	if err := ds.db.Create(dbSession).Error; err != nil {
		return "", ds.HandleError(err)
	}
	return dbSession.ID, nil
}

func (ds *SqliteService) GetActiveSession(userID, tokenHash string) (*model.UserSession, error) {
	var session model.UserSession
	err := ds.db.Where("user_id = ? AND token_hash = ? AND is_active = ? AND expires_at > ?",
		userID, tokenHash, true, time.Now()).First(&session).Error
	if err != nil {
		return nil, ds.HandleError(err)
	}
	return &session, nil
}

func (ds *SqliteService) UpdateSessionLastUsed(sessionID string) error {
	return ds.db.Model(&model.UserSession{}).Where("id = ?", sessionID).Update("last_used", time.Now()).Error
}

func (ds *SqliteService) UpdateSessionToken(sessionID, newTokenHash string) error {
	return ds.db.Model(&model.UserSession{}).Where("id = ?", sessionID).Updates(map[string]interface{}{
		"token_hash": newTokenHash,
		"last_used":  time.Now(),
	}).Error
}

func (ds *SqliteService) DeactivateSession(sessionID, userID string) error {
	return ds.db.Model(&model.UserSession{}).Where("id = ? AND user_id = ?", sessionID, userID).Updates(map[string]interface{}{
		"is_active": false,
		"last_used": time.Now(),
	}).Error
}

func (ds *SqliteService) DeactivateAllUserSessions(userID, exceptSessionID string) error {
	query := ds.db.Model(&model.UserSession{}).Where("user_id = ?", userID)
	if exceptSessionID != "" {
		query = query.Where("id != ?", exceptSessionID)
	}

	return query.Updates(map[string]interface{}{
		"is_active": false,
		"last_used": time.Now(),
	}).Error
}

func (ds *SqliteService) GetUserSessions(userID string) ([]model.UserSession, error) {
	var sessions []model.UserSession
	err := ds.db.Where("user_id = ? AND is_active = ?", userID, true).
		Order("last_used DESC").Find(&sessions).Error
	if err != nil {
		return nil, ds.HandleError(err)
	}
	return sessions, nil
}

func (ds *SqliteService) CleanupExpiredSessions() error {
	return ds.db.Model(&model.UserSession{}).
		Where("expires_at < ?", time.Now()).
		Update("is_active", false).Error
}

// ==================== PASSWORD RESET METHODS ====================

func (ds *SqliteService) CreatePasswordResetCode(userID, code string, expiresAt time.Time) error {
	resetToken := &model.PasswordResetCode{
		ID:        uuid.New().String(),
		UserID:    userID,
		Code:      code,
		ExpiresAt: expiresAt,
		Used:      false,
		CreatedAt: time.Now(),
	}

	return ds.db.Create(resetToken).Error
}

func (ds *SqliteService) GetPasswordResetCode(code string) (*model.PasswordResetCode, error) {
	var resetCode model.PasswordResetCode
	err := ds.db.Where("code = ? AND used = ?", code, false).First(&resetCode).Error
	if err != nil {
		return nil, ds.HandleError(err)
	}
	return &resetCode, nil
}

func (ds *SqliteService) InvalidatePasswordResetCode(code string) error {
	return ds.db.Model(&model.PasswordResetCode{}).Where("code = ?", code).Update("used", true).Error
}

func (ds *SqliteService) CleanupExpiredPasswordCodes() error {
	return ds.db.Where("expires_at < ?", time.Now()).Delete(&model.PasswordResetCode{}).Error
}

// ==================== TOKEN BLACKLIST METHODS ====================

func (ds *SqliteService) BlacklistToken(jti string, expiresAt time.Time) error {
	blacklistedToken := &model.BlacklistedToken{
		JTI:       jti,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	return ds.db.Create(blacklistedToken).Error
}

func (ds *SqliteService) IsTokenBlacklisted(jti string) bool {
	var count int64
	ds.db.Model(&model.BlacklistedToken{}).Where("jti = ? AND expires_at > ?", jti, time.Now()).Count(&count)
	return count > 0
}

func (ds *SqliteService) CleanupExpiredBlacklistedTokens() error {
	return ds.db.Where("expires_at < ?", time.Now()).Delete(&model.BlacklistedToken{}).Error
}

// ==================== AUDIT LOG METHODS ====================

func (ds *SqliteService) CreateAuthAuditLog(log dto.AuthAuditLog) error {
	auditLog := &model.AuthAuditLog{
		ID:        uuid.New().String(),
		UserID:    log.UserID,
		Action:    log.Action,
		IP:        log.IP,
		UserAgent: log.UserAgent,
		Timestamp: log.Timestamp,
		Success:   log.Success,
		Details:   log.Details,
	}

	return ds.db.Create(auditLog).Error
}

func (ds *SqliteService) GetUserAuditLogs(userID string, page, limit int) ([]model.AuthAuditLog, int64, error) {
	var logs []model.AuthAuditLog
	var total int64

	// Get total count
	ds.db.Model(&model.AuthAuditLog{}).Where("user_id = ?", userID).Count(&total)

	// Get paginated results
	offset := (page - 1) * limit
	err := ds.db.Where("user_id = ?", userID).
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error

	if err != nil {
		return nil, 0, ds.HandleError(err)
	}

	return logs, total, nil
}

func (ds *SqliteService) GetAuditLogs(page, limit int, userID, action string) ([]model.AuthAuditLog, int64, error) {
	var logs []model.AuthAuditLog
	var total int64

	query := ds.db.Model(&model.AuthAuditLog{})

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}

	// Get total count
	query.Count(&total)

	// Get paginated results
	offset := (page - 1) * limit
	err := query.Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error

	if err != nil {
		return nil, 0, ds.HandleError(err)
	}

	return logs, total, nil
}

func (ds *SqliteService) CleanupOldAuditLogs(olderThan time.Time) error {
	return ds.db.Where("timestamp < ?", olderThan).Delete(&model.AuthAuditLog{}).Error
}

// ==================== TRUSTED DEVICE METHODS ====================

func (ds *SqliteService) CreateTrustedDevice(device *model.TrustedDevice) error {
	device.ID = uuid.New().String()
	device.CreatedAt = time.Now()
	device.LastUsed = time.Now()

	return ds.db.Create(device).Error
}

func (ds *SqliteService) GetTrustedDevice(userID, deviceID string) (*model.TrustedDevice, error) {
	var device model.TrustedDevice
	err := ds.db.Where("user_id = ? AND device_id = ?", userID, deviceID).First(&device).Error
	if err != nil {
		return nil, ds.HandleError(err)
	}
	return &device, nil
}

func (ds *SqliteService) UpdateTrustedDevice(device *model.TrustedDevice) error {
	device.LastUsed = time.Now()
	return ds.db.Save(device).Error
}

func (ds *SqliteService) GetUserTrustedDevices(userID string) ([]model.TrustedDevice, error) {
	var devices []model.TrustedDevice
	err := ds.db.Where("user_id = ?", userID).Order("last_used DESC").Find(&devices).Error
	if err != nil {
		return nil, ds.HandleError(err)
	}
	return devices, nil
}

func (ds *SqliteService) RemoveTrustedDevice(userID, deviceID string) error {
	return ds.db.Where("user_id = ? AND device_id = ?", userID, deviceID).Delete(&model.TrustedDevice{}).Error
}

// ==================== LOGIN ATTEMPT METHODS ====================

func (ds *SqliteService) RecordLoginAttempt(ip, email, userAgent string, success bool) error {
	attempt := &model.LoginAttempt{
		ID:        uuid.New().String(),
		IP:        ip,
		Email:     email,
		Success:   success,
		Timestamp: time.Now(),
		UserAgent: userAgent,
	}

	return ds.db.Create(attempt).Error
}

func (ds *SqliteService) GetRecentLoginAttempts(ip string, since time.Time) ([]model.LoginAttempt, error) {
	var attempts []model.LoginAttempt
	err := ds.db.Where("ip = ? AND timestamp > ?", ip, since).
		Order("timestamp DESC").Find(&attempts).Error
	if err != nil {
		return nil, ds.HandleError(err)
	}
	return attempts, nil
}

func (ds *SqliteService) CleanupOldLoginAttempts(olderThan time.Time) error {
	return ds.db.Where("timestamp < ?", olderThan).Delete(&model.LoginAttempt{}).Error
}

// ==================== ADMIN USER MANAGEMENT ====================

func (ds *SqliteService) AdminGetUsers(page, limit int, search string) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := ds.db.Model(&model.User{}).Where("deleted_at IS NULL")

	if search != "" {
		searchPattern := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(username) LIKE ? OR LOWER(email) LIKE ?", searchPattern, searchPattern)
	}

	// Get total count
	query.Count(&total)

	// Get paginated results
	offset := (page - 1) * limit
	err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error

	if err != nil {
		return nil, 0, ds.HandleError(err)
	}

	return users, total, nil
}

func (ds *SqliteService) AdminUpdateUser(userID string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return ds.db.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error
}

func (ds *SqliteService) AdminDeleteUser(userID string) error {
	now := time.Now()
	return ds.db.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"deleted_at": &now,
		"is_active":  false,
		"updated_at": now,
	}).Error
}

// ==================== USER PROFILE & SECURITY METHODS ====================

func (ds *SqliteService) GetUserProfile(userID string) (*model.User, error) {
	var user model.User
	err := ds.db.Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error
	if err != nil {
		return nil, ds.HandleError(err)
	}
	return &user, nil
}

func (ds *SqliteService) UpdateUserProfile(userID string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return ds.db.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error
}

func (ds *SqliteService) GetSecuritySettings(userID string) (*dto.SecuritySettings, error) {
	var user model.User
	err := ds.db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, ds.HandleError(err)
	}

	settings := &dto.SecuritySettings{
		TwoFactorEnabled:     user.TwoFactorEnabled,
		BackupCodesGenerated: user.BackupCodes != "",
		LastPasswordChange:   user.LastPasswordChange,
		LoginNotifications:   user.LoginNotifications,
		SessionTimeout:       user.SessionTimeout,
	}

	return settings, nil
}

func (ds *SqliteService) UpdateSecuritySettings(userID string, settings dto.UpdateSecuritySettingsRequest) error {
	updates := make(map[string]interface{})
	updates["updated_at"] = time.Now()

	if settings.LoginNotifications != nil {
		updates["login_notifications"] = *settings.LoginNotifications
	}
	if settings.SessionTimeout != nil {
		updates["session_timeout"] = *settings.SessionTimeout
	}

	return ds.db.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error
}

// ==================== CLEANUP AND MAINTENANCE ====================

func (ds *SqliteService) CleanupExpiredData() error {
	now := time.Now()

	// Cleanup expired sessions
	ds.CleanupExpiredSessions()

	// Cleanup expired password reset codes
	ds.CleanupExpiredPasswordCodes()

	// Cleanup expired blacklisted tokens
	ds.CleanupExpiredBlacklistedTokens()

	// Cleanup old login attempts (keep last 30 days)
	ds.CleanupOldLoginAttempts(now.Add(-30 * 24 * time.Hour))

	// Cleanup old audit logs (keep last 90 days)
	ds.CleanupOldAuditLogs(now.Add(-90 * 24 * time.Hour))

	return nil
}

// ==================== STATISTICS METHODS ====================

func (ds *SqliteService) GetUserStats(userID string) (*dto.UserStats, error) {
	var user model.User
	err := ds.db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, ds.HandleError(err)
	}

	// Count active sessions
	var sessionCount int64
	ds.db.Model(&model.UserSession{}).Where("user_id = ? AND is_active = ? AND expires_at > ?",
		userID, true, time.Now()).Count(&sessionCount)

	// Count total logins from audit logs
	var loginCount int64
	ds.db.Model(&model.AuthAuditLog{}).Where("user_id = ? AND action = ? AND success = ?",
		userID, model.ActionLogin, true).Count(&loginCount)

	stats := &dto.UserStats{
		TotalLogins:        int(loginCount),
		FailedAttempts:     user.FailedAttempts,
		ActiveSessions:     int(sessionCount),
		LastPasswordChange: user.LastPasswordChange,
	}

	return stats, nil
}

// Seed initial data for the enhanced auth system
func (ds *SqliteService) seedInitialData() error {
	// Create default admin user if it doesn't exist
	err := ds.createDefaultAdmin()
	if err != nil {
		return err
	}

	// Create default rate limit configs
	err = ds.createDefaultRateLimitConfigs()
	if err != nil {
		return err
	}

	return nil
}

// Create default admin user
func (ds *SqliteService) createDefaultAdmin() error {
	var count int64
	ds.db.Model(&model.User{}).Where("role = ?", model.RoleAdmin).Count(&count)

	if count == 0 {
		// Hash default password (CHANGE THIS IN PRODUCTION!)
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), 12)
		if err != nil {
			return err
		}

		admin := &model.User{
			ID:                 "admin-" + time.Now().Format("20060102150405"),
			Username:           "admin",
			Email:              "admin@techyouth.com",
			Password:           string(hashedPassword),
			Role:               model.RoleAdmin,
			IsActive:           true,
			EmailVerified:      true,
			LoginNotifications: true,
			SessionTimeout:     1440,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}

		err = ds.db.Create(admin).Error
		if err != nil {
			log.Printf("Failed to create admin user: %v", err)
			return err
		}

		log.Println("Default admin user created - Username: admin, Password: admin123 (CHANGE THIS!)")
	}

	return nil
}

// Create default rate limit configurations
func (ds *SqliteService) createDefaultRateLimitConfigs() error {
	configs := []model.RateLimitConfig{
		{
			ID:           "login-config",
			EndpointType: "login",
			Limit:        10,
			WindowSize:   900,  // 15 minutes
			BlockTime:    1800, // 30 minutes
		},
		{
			ID:           "register-config",
			EndpointType: "register",
			Limit:        5,
			WindowSize:   900,  // 15 minutes
			BlockTime:    3600, // 1 hour
		},
		{
			ID:           "forgot-password-config",
			EndpointType: "forgot_password",
			Limit:        3,
			WindowSize:   900,  // 15 minutes
			BlockTime:    3600, // 1 hour
		},
		{
			ID:           "reset-password-config",
			EndpointType: "reset_password",
			Limit:        5,
			WindowSize:   900,  // 15 minutes
			BlockTime:    1800, // 30 minutes
		},
		{
			ID:           "refresh-config",
			EndpointType: "refresh",
			Limit:        20,
			WindowSize:   900, // 15 minutes
			BlockTime:    300, // 5 minutes
		},
		{
			ID:           "resend-verification-config",
			EndpointType: "resend_verification",
			Limit:        3,
			WindowSize:   300,  // 5 minutes
			BlockTime:    1800, // 30 minutes
		},
	}

	for _, config := range configs {
		var existing model.RateLimitConfig
		err := ds.db.Where("endpoint_type = ?", config.EndpointType).First(&existing).Error
		if err != nil {
			// Config doesn't exist, create it
			err = ds.db.Create(&config).Error
			if err != nil {
				log.Printf("Failed to create rate limit config for %s: %v", config.EndpointType, err)
			}
		}
	}

	return nil
}
