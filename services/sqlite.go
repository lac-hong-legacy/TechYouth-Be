package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lac-hong-legacy/TechYouth-Be/dto"
	"github.com/lac-hong-legacy/TechYouth-Be/model"

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

// Id returns Service ID
func (ds SqliteService) Id() string {
	return SQLITE_SVC
}

// Db Access to raw SqliteService db
func (ds SqliteService) Db() *gorm.DB {
	return ds.db
}

// Configure the service
func (ds *SqliteService) Configure(ctx *context.Context) error {
	ds.database = os.Getenv("DB_NAME")

	return ds.DefaultService.Configure(ctx)
}

// Start the service and open connection to the database
// Migrate any tables that have changed since last runtime
func (ds *SqliteService) Start() (err error) {
	ds.db, err = gorm.Open(sqlite.Open(ds.database), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		return err
	}

	models := []interface{}{
		&model.User{},
		&model.GuestSession{},
		&model.GuestProgress{},
		&model.GuestLessonAttempt{},
		&model.RateLimit{},
		&model.RateLimitConfig{},

		// New content models
		&model.Character{},
		&model.Lesson{},
		&model.Timeline{},

		// New user models
		&model.UserProgress{},
		&model.Spirit{},
		&model.Achievement{},
		&model.UserAchievement{},
		&model.UserLessonAttempt{},
	}

	err = ds.db.AutoMigrate(models...)
	if err != nil {
		log.Printf("Failed to migrate database: %v", err)
		return err
	}

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

func (ds *SqliteService) CreateUser(user dto.RegisterRequest) (*model.User, error) {
	if _, err := ds.GetUserByEmail(user.Email); err == nil {
		return nil, errors.New("email already exists")
	}

	if _, err := ds.GetUserByUsername(user.Username); err == nil {
		return nil, errors.New("username already exists")
	}

	userModel := model.User{
		ID:       uuid.New().String(),
		Email:    user.Email,
		Username: user.Username,
		Password: user.Password,
	}

	if err := ds.db.Create(&userModel).Error; err != nil {
		return nil, err
	}
	return &userModel, nil
}

func (ds *SqliteService) GetSessionByDeviceID(deviceID string) (*model.GuestSession, error) {
	var session model.GuestSession
	if err := ds.db.Where("device_id = ?", deviceID).First(&session).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return &session, nil
}

func (ds *SqliteService) CreateSession(session *model.GuestSession) (*model.GuestSession, error) {
	session.ID = uuid.New().String()
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
	progress.ID = uuid.New().String()
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
	attempt.ID = uuid.New().String()
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
		rateLimit.ID = uuid.New().String()
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
		character.ID = uuid.New().String()
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
		lesson.ID = uuid.New().String()
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
		timeline.ID = uuid.New().String()
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
		progress.ID = uuid.New().String()
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
		spirit.ID = uuid.New().String()
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
		achievement.ID = uuid.New().String()
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
		userAchievement.ID = uuid.New().String()
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

func (ds *SqliteService) GetUserStats(userID string) (map[string]interface{}, error) {
	progress, err := ds.GetUserProgress(userID)
	if err != nil {
		return nil, err
	}

	spirit, err := ds.GetUserSpirit(userID)
	if err != nil {
		return nil, err
	}

	var completedLessons []string
	if err := json.Unmarshal(progress.CompletedLessons, &completedLessons); err != nil {
		return nil, err
	}

	var unlockedCharacters []string
	if err := json.Unmarshal(progress.UnlockedCharacters, &unlockedCharacters); err != nil {
		return nil, err
	}

	achievements, err := ds.GetUserAchievements(userID)
	if err != nil {
		return nil, err
	}

	rank, err := ds.GetUserRank(userID)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"user_id":             userID,
		"level":               progress.Level,
		"xp":                  progress.XP,
		"hearts":              progress.Hearts,
		"streak":              progress.Streak,
		"total_play_time":     progress.TotalPlayTime,
		"completed_lessons":   len(completedLessons),
		"unlocked_characters": len(unlockedCharacters),
		"achievements":        len(achievements),
		"rank":                rank,
		"spirit_stage":        spirit.Stage,
		"spirit_type":         spirit.Type,
	}

	return stats, nil
}

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
	attempt.ID = uuid.New().String()
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

// ==================== SEEDING METHODS ====================
// SeedCharacters seeds the database with Vietnamese historical characters
func (s *SqliteService) SeedCharacters() error {
	characters := s.getHistoricalCharacters()

	for _, character := range characters {
		// Check if character already exists
		var existingChar model.Character
		if err := s.db.Where("id = ?", character.ID).First(&existingChar).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Character doesn't exist, create it
				if err := s.db.Create(&character).Error; err != nil {
					log.Printf("Error creating character %s: %v", character.Name, err)
					return err
				}
				log.Printf("Created character: %s", character.Name)
			} else {
				log.Printf("Error checking character %s: %v", character.Name, err)
				return err
			}
		} else {
			log.Printf("Character %s already exists, skipping", character.Name)
		}
	}

	log.Println("Character seeding completed successfully")
	return nil
}

// getHistoricalCharacters returns the list of 20 Vietnamese historical characters
func (s *SqliteService) getHistoricalCharacters() []model.Character {
	now := time.Now()

	characters := []model.Character{
		{
			ID:          "char_hung_vuong",
			Name:        "Hùng Vương",
			Dynasty:     "Văn Lang",
			Rarity:      "Legendary",
			BirthYear:   intPtr(-2879),
			DeathYear:   intPtr(-258),
			Description: "Legendary founder of the first Vietnamese state, Văn Lang. Known as the first king of Vietnam, he established the foundation of Vietnamese civilization along the Red River delta.",
			FamousQuote: "Con rồng cháu tiên",
			Achievements: jsonArray([]string{
				"Founded the Văn Lang kingdom",
				"Established the Hùng dynasty",
				"Created the foundation of Vietnamese culture",
				"Unified the Lạc Việt tribes",
			}),
			ImageURL:   "/assets/characters/hung_vuong.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_ly_thai_to",
			Name:        "Lý Thái Tổ",
			Dynasty:     "Lý",
			Rarity:      "Legendary",
			BirthYear:   intPtr(974),
			DeathYear:   intPtr(1028),
			Description: "Founder of the Lý dynasty and builder of Thăng Long (modern Hanoi). He moved the capital from Hoa Lư to Thăng Long, establishing a golden age of Vietnamese culture and Buddhism.",
			FamousQuote: "Thăng Long hữu đức khí",
			Achievements: jsonArray([]string{
				"Founded the Lý dynasty in 1009",
				"Established Thăng Long as the capital",
				"Promoted Buddhism and education",
				"Unified and strengthened Vietnam",
			}),
			ImageURL:   "/assets/characters/ly_thai_to.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_tran_hung_dao",
			Name:        "Trần Hưng Đạo",
			Dynasty:     "Trần",
			Rarity:      "Legendary",
			BirthYear:   intPtr(1228),
			DeathYear:   intPtr(1300),
			Description: "Greatest military strategist in Vietnamese history. Successfully defended Vietnam against three Mongol invasions, using guerrilla tactics and superior knowledge of local terrain.",
			FamousQuote: "Thà chết vì tổ quốc chứ không sống làm nô lệ",
			Achievements: jsonArray([]string{
				"Defeated three Mongol invasions (1258, 1285, 1287-1288)",
				"Developed innovative guerrilla warfare tactics",
				"Protected Vietnamese independence",
				"Wrote military treatises on strategy",
			}),
			ImageURL:   "/assets/characters/tran_hung_dao.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_le_loi",
			Name:        "Lê Lợi",
			Dynasty:     "Lê",
			Rarity:      "Legendary",
			BirthYear:   intPtr(1385),
			DeathYear:   intPtr(1433),
			Description: "Hero who liberated Vietnam from Chinese Ming occupation. Founded the Lê dynasty after a ten-year resistance war, becoming Emperor Lê Thái Tổ.",
			FamousQuote: "Nam quốc sơn hà Nam đế cư",
			Achievements: jsonArray([]string{
				"Led successful rebellion against Ming China (1418-1428)",
				"Founded the Later Lê dynasty",
				"Liberated Vietnam from foreign rule",
				"Established the Proclamation of Victory over the Wu",
			}),
			ImageURL:   "/assets/characters/le_loi.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_nguyen_trai",
			Name:        "Nguyễn Trãi",
			Dynasty:     "Lê",
			Rarity:      "Rare",
			BirthYear:   intPtr(1380),
			DeathYear:   intPtr(1442),
			Description: "Brilliant strategist, poet, and diplomat who served Lê Lợi. Known for his literary works and the famous 'Bình Ngô Đại Cáo' proclamation.",
			FamousQuote: "Dĩ nhân thắng bạo, dĩ nghĩa thắng phi nghĩa",
			Achievements: jsonArray([]string{
				"Authored Bình Ngô Đại Cáo (Great Proclamation of Victory)",
				"Chief advisor to Lê Lợi during independence war",
				"Renowned poet and literary figure",
				"Pioneered Vietnamese diplomatic writing",
			}),
			ImageURL:   "/assets/characters/nguyen_trai.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_nguyen_hue",
			Name:        "Nguyễn Huệ (Quang Trung)",
			Dynasty:     "Tây Sơn",
			Rarity:      "Legendary",
			BirthYear:   intPtr(1753),
			DeathYear:   intPtr(1792),
			Description: "Emperor Quang Trung, leader of the Tây Sơn rebellion. Defeated the Qing invasion and initiated important reforms including promoting Vietnamese language and culture.",
			FamousQuote: "Bắc định Thanh quân, nam bình Chúa Nguyễn",
			Achievements: jsonArray([]string{
				"Led the Tây Sơn uprising",
				"Defeated Qing army at Đống Đa (1789)",
				"Promoted Vietnamese language in education",
				"Implemented progressive social reforms",
			}),
			ImageURL:   "/assets/characters/nguyen_hue.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_hai_ba_trung",
			Name:        "Hai Bà Trưng",
			Dynasty:     "Trưng",
			Rarity:      "Legendary",
			BirthYear:   intPtr(12),
			DeathYear:   intPtr(43),
			Description: "The Trưng Sisters - Trưng Trắc and Trưng Nhị - led the first major rebellion against Chinese domination. They established an independent kingdom for three years.",
			FamousQuote: "Trước để trừ giặc, sau để cảnh báo hậu thế",
			Achievements: jsonArray([]string{
				"Led successful rebellion against Chinese Han dynasty",
				"Established independent Vietnamese kingdom (40-43 AD)",
				"Symbol of Vietnamese women's strength",
				"First recorded Vietnamese rulers after Chinese occupation",
			}),
			ImageURL:   "/assets/characters/hai_ba_trung.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_ly_thuong_kiet",
			Name:        "Lý Thường Kiệt",
			Dynasty:     "Lý",
			Rarity:      "Rare",
			BirthYear:   intPtr(1019),
			DeathYear:   intPtr(1105),
			Description: "Great general of the Lý dynasty who successfully defended Vietnam against Song China. Famous for his military innovations and the poem 'Nam quốc sơn hà'.",
			FamousQuote: "Nam quốc sơn hà Nam đế cư, Tiệt nhiên định phận tại thiên thư",
			Achievements: jsonArray([]string{
				"Defeated Song Chinese invasion",
				"Authored the famous patriotic poem 'Nam quốc sơn hà'",
				"Served as regent during Lý Nhân Tông's minority",
				"Strengthened Vietnamese borders",
			}),
			ImageURL:   "/assets/characters/ly_thuong_kiet.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_ngo_quyen",
			Name:        "Ngô Quyền",
			Dynasty:     "Ngô",
			Rarity:      "Rare",
			BirthYear:   intPtr(897),
			DeathYear:   intPtr(944),
			Description: "Founder of the Ngô dynasty who ended nearly 1000 years of Chinese domination. Won the decisive Battle of Bạch Đằng River against the Southern Han fleet.",
			FamousQuote: "Đại Việt độc lập",
			Achievements: jsonArray([]string{
				"Won Battle of Bạch Đằng River (938)",
				"Ended Chinese domination after 1000 years",
				"Founded the Ngô dynasty",
				"Established Vietnamese independence",
			}),
			ImageURL:   "/assets/characters/ngo_quyen.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_dinh_bo_linh",
			Name:        "Đinh Bộ Lĩnh",
			Dynasty:     "Đinh",
			Rarity:      "Rare",
			BirthYear:   intPtr(924),
			DeathYear:   intPtr(979),
			Description: "First emperor of unified Vietnam, known as Đinh Tiên Hoàng. Established the Đại Cồ Việt kingdom and brought stability after years of chaos.",
			FamousQuote: "Đại Cồ Việt Hoàng Đế",
			Achievements: jsonArray([]string{
				"Unified Vietnam under Đại Cồ Việt kingdom",
				"First emperor of independent Vietnam",
				"Established capital at Hoa Lư",
				"Created organized government structure",
			}),
			ImageURL:   "/assets/characters/dinh_bo_linh.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_le_thanh_tong",
			Name:        "Lê Thánh Tông",
			Dynasty:     "Lê",
			Rarity:      "Legendary",
			BirthYear:   intPtr(1442),
			DeathYear:   intPtr(1497),
			Description: "Greatest emperor of the Lê dynasty, known for territorial expansion and legal reforms. Created the Hồng Đức legal code and promoted literature and arts.",
			FamousQuote: "Minh đức tân dân",
			Achievements: jsonArray([]string{
				"Created the comprehensive Hồng Đức legal code",
				"Expanded territory to Champa kingdom",
				"Promoted Confucian education and civil service",
				"Golden age of Vietnamese literature and arts",
			}),
			ImageURL:   "/assets/characters/le_thanh_tong.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_ho_quy_ly",
			Name:        "Hồ Quý Ly",
			Dynasty:     "Hồ",
			Rarity:      "Common",
			BirthYear:   intPtr(1336),
			DeathYear:   intPtr(1407),
			Description: "Controversial reformer who founded the short-lived Hồ dynasty. Known for progressive reforms but also for the events that led to Ming Chinese invasion.",
			FamousQuote: "Cải cách duy tân",
			Achievements: jsonArray([]string{
				"Implemented land redistribution reforms",
				"Promoted paper currency system",
				"Advanced agricultural techniques",
				"Founded Hồ dynasty (1400-1407)",
			}),
			ImageURL:   "/assets/characters/ho_quy_ly.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_mac_dang_dung",
			Name:        "Mạc Đăng Dung",
			Dynasty:     "Mạc",
			Rarity:      "Common",
			BirthYear:   intPtr(1483),
			DeathYear:   intPtr(1541),
			Description: "Founder of the Mạc dynasty who seized power from the Lê dynasty. Known for his military skills but his legitimacy was often disputed.",
			FamousQuote: "Thiên mệnh tại ta",
			Achievements: jsonArray([]string{
				"Founded the Mạc dynasty",
				"Skilled military commander",
				"Controlled northern Vietnam",
				"Established diplomatic relations with China",
			}),
			ImageURL:   "/assets/characters/mac_dang_dung.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_nguyen_anh",
			Name:        "Nguyễn Ánh (Gia Long)",
			Dynasty:     "Nguyễn",
			Rarity:      "Rare",
			BirthYear:   intPtr(1762),
			DeathYear:   intPtr(1820),
			Description: "Founded the Nguyễn dynasty and unified Vietnam under Emperor Gia Long. Established Huế as the capital and created the modern borders of Vietnam.",
			FamousQuote: "Thống nhất giang sơn",
			Achievements: jsonArray([]string{
				"Unified Vietnam from north to south",
				"Founded the Nguyễn dynasty (1802-1945)",
				"Established Huế as imperial capital",
				"Created modern Vietnamese territorial unity",
			}),
			ImageURL:   "/assets/characters/nguyen_anh.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_phan_boi_chau",
			Name:        "Phan Bội Châu",
			Dynasty:     "Cận đại",
			Rarity:      "Rare",
			BirthYear:   intPtr(1867),
			DeathYear:   intPtr(1940),
			Description: "Pioneering nationalist and independence activist. Led early resistance movements against French colonial rule and promoted modern education and reform.",
			FamousQuote: "Việt Nam vong quốc sử",
			Achievements: jsonArray([]string{
				"Founded Việt Nam Duy Tân Hội",
				"Led Đông Du movement to Japan",
				"Wrote influential nationalist literature",
				"Pioneer of Vietnamese independence movement",
			}),
			ImageURL:   "/assets/characters/phan_boi_chau.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_phan_chu_trinh",
			Name:        "Phan Châu Trinh",
			Dynasty:     "Cận đại",
			Rarity:      "Common",
			BirthYear:   intPtr(1872),
			DeathYear:   intPtr(1926),
			Description: "Reformist scholar who advocated for modernization and education. Promoted peaceful resistance and cultural reform during French colonial period.",
			FamousQuote: "Dân trí là gốc của mọi việc",
			Achievements: jsonArray([]string{
				"Advocated for educational reform",
				"Promoted peaceful resistance methods",
				"Founded modern Vietnamese journalism",
				"Influenced intellectual awakening movement",
			}),
			ImageURL:   "/assets/characters/phan_chu_trinh.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_nguyen_du",
			Name:        "Nguyễn Du",
			Dynasty:     "Nguyễn",
			Rarity:      "Rare",
			BirthYear:   intPtr(1765),
			DeathYear:   intPtr(1820),
			Description: "Vietnam's greatest poet, author of the epic poem 'Truyện Kiều' (The Tale of Kiều). His work represents the pinnacle of Vietnamese classical literature.",
			FamousQuote: "Trăm năm trong cõi người ta, chữ tài chữ mệnh khéo là ghét nhau",
			Achievements: jsonArray([]string{
				"Authored Truyện Kiều, masterpiece of Vietnamese literature",
				"Master of Nôm script poetry",
				"Influenced Vietnamese cultural identity",
				"UNESCO recognized his contribution to world literature",
			}),
			ImageURL:   "/assets/characters/nguyen_du.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_ba_trieu",
			Name:        "Bà Triệu",
			Dynasty:     "Ngô",
			Rarity:      "Rare",
			BirthYear:   intPtr(225),
			DeathYear:   intPtr(248),
			Description: "Female warrior who led a rebellion against Wu Chinese occupation in the 3rd century. Known for her courage and determination in fighting foreign domination.",
			FamousQuote: "Tôi muốn cưỡi cơn gió mạnh, đạp sóng dữ, chém cá kình ở biển Đông",
			Achievements: jsonArray([]string{
				"Led rebellion against Chinese Wu dynasty",
				"Symbol of Vietnamese women's heroism",
				"Fought for Vietnamese independence in 3rd century",
				"Inspired future generations of patriots",
			}),
			ImageURL:   "/assets/characters/ba_trieu.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_vo_thi_sau",
			Name:        "Võ Thị Sáu",
			Dynasty:     "Cận đại",
			Rarity:      "Common",
			BirthYear:   intPtr(1933),
			DeathYear:   intPtr(1952),
			Description: "Young revolutionary who became a symbol of resistance during the French colonial period. Executed at age 19 for her anti-colonial activities.",
			FamousQuote: "Tôi chết vì Tổ quốc, lòng tôi vui sướng lắm",
			Achievements: jsonArray([]string{
				"Active member of resistance movement",
				"Symbol of young Vietnamese patriotism",
				"Sacrificed life for independence cause",
				"Inspired anti-colonial resistance",
			}),
			ImageURL:   "/assets/characters/vo_thi_sau.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:          "char_ho_chi_minh",
			Name:        "Hồ Chí Minh",
			Dynasty:     "Cận đại",
			Rarity:      "Legendary",
			BirthYear:   intPtr(1890),
			DeathYear:   intPtr(1969),
			Description: "Founding father of modern Vietnam and leader of independence movement. Led the country's struggle against French colonialism and American intervention.",
			FamousQuote: "Không có gì quý hơn độc lập tự do",
			Achievements: jsonArray([]string{
				"Founded Democratic Republic of Vietnam",
				"Led independence movement against France",
				"Declared Vietnamese independence (1945)",
				"Father of modern Vietnamese nation",
			}),
			ImageURL:   "/assets/characters/ho_chi_minh.jpg",
			IsUnlocked: false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}

	return characters
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func jsonArray(items []string) json.RawMessage {
	data, _ := json.Marshal(items)
	return json.RawMessage(data)
}

// SeedLessons seeds the database with lessons for historical characters
func (s *SqliteService) SeedLessons() error {
	lessons := s.getHistoricalLessons()

	for _, lesson := range lessons {
		// Check if lesson already exists
		var existingLesson model.Lesson
		if err := s.db.Where("id = ?", lesson.ID).First(&existingLesson).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Lesson doesn't exist, create it
				if err := s.db.Create(&lesson).Error; err != nil {
					log.Printf("Error creating lesson %s: %v", lesson.Title, err)
					return err
				}
				log.Printf("Created lesson: %s", lesson.Title)
			} else {
				log.Printf("Error checking lesson %s: %v", lesson.Title, err)
				return err
			}
		} else {
			log.Printf("Lesson %s already exists, skipping", lesson.Title)
		}
	}

	log.Println("Lesson seeding completed successfully")
	return nil
}

// getHistoricalLessons returns sample lessons for key historical characters
func (s *SqliteService) getHistoricalLessons() []model.Lesson {
	now := time.Now()

	lessons := []model.Lesson{
		// Hùng Vương lessons
		{
			ID:           "lesson_hung_vuong_1",
			CharacterID:  "char_hung_vuong",
			Title:        "The Birth of Văn Lang",
			Order:        1,
			Story:        "Long ago, in the misty mountains of northern Vietnam, lived the Dragon King Lạc Long Quân and the Fairy Mother Âu Cơ. From their union came 100 sons, and the eldest became the first Hùng King, founding the ancient kingdom of Văn Lang...",
			VoiceOverURL: "/assets/audio/hung_vuong_1.mp3",
			Questions: createQuestions([]QuestionData{
				{
					ID:       "q_hv1_1",
					Type:     "multiple_choice",
					Question: "Who founded the first Vietnamese kingdom of Văn Lang?",
					Options:  []string{"Hùng Vương", "Lý Thái Tổ", "Ngô Quyền", "Đinh Bộ Lĩnh"},
					Answer:   "Hùng Vương",
					Points:   10,
				},
				{
					ID:       "q_hv1_2",
					Type:     "fill_blank",
					Question: "The ancient kingdom founded by Hùng Vương was called _____.",
					Answer:   "Văn Lang",
					Points:   15,
				},
			}),
			XPReward:  50,
			MinScore:  70,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:           "lesson_hung_vuong_2",
			CharacterID:  "char_hung_vuong",
			Title:        "The Lạc Việt People",
			Order:        2,
			Story:        "The Hùng Kings ruled over the Lạc Việt people, who were skilled in bronze-making, rice cultivation, and boat building. They created a sophisticated society along the Red River delta, laying the foundation for Vietnamese civilization...",
			VoiceOverURL: "/assets/audio/hung_vuong_2.mp3",
			Questions: createQuestions([]QuestionData{
				{
					ID:       "q_hv2_1",
					Type:     "multiple_choice",
					Question: "The Hùng Kings ruled over which ancient Vietnamese people?",
					Options:  []string{"Cham people", "Lạc Việt", "Kinh people", "Hmong people"},
					Answer:   "Lạc Việt",
					Points:   10,
				},
			}),
			XPReward:  50,
			MinScore:  70,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},

		// Trần Hưng Đạo lessons
		{
			ID:           "lesson_tran_hung_dao_1",
			CharacterID:  "char_tran_hung_dao",
			Title:        "The Mongol Threat",
			Order:        1,
			Story:        "In the 13th century, the mighty Mongol Empire had conquered China and was turning its attention to Đại Việt. Kublai Khan's armies seemed unstoppable, but Prince Trần Quốc Tuấn, known as Trần Hưng Đạo, would prove them wrong...",
			VoiceOverURL: "/assets/audio/tran_hung_dao_1.mp3",
			Questions: createQuestions([]QuestionData{
				{
					ID:       "q_thd1_1",
					Type:     "multiple_choice",
					Question: "Which empire threatened Đại Việt in the 13th century?",
					Options:  []string{"Chinese Ming", "Mongol Empire", "Khmer Empire", "Cham Kingdom"},
					Answer:   "Mongol Empire",
					Points:   10,
				},
				{
					ID:       "q_thd1_2",
					Type:     "multiple_choice",
					Question: "Who was the leader of the Mongol Empire during the invasions of Vietnam?",
					Options:  []string{"Genghis Khan", "Ögedei Khan", "Kublai Khan", "Möngke Khan"},
					Answer:   "Kublai Khan",
					Points:   15,
				},
			}),
			XPReward:  75,
			MinScore:  70,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:           "lesson_tran_hung_dao_2",
			CharacterID:  "char_tran_hung_dao",
			Title:        "Victory at Bạch Đằng",
			Order:        2,
			Story:        "Using ancient tactics and superior knowledge of local waters, Trần Hưng Đạo planted iron-tipped wooden stakes in the Bạch Đằng River. When the Mongol fleet attacked during high tide, the Vietnamese lured them forward, then retreated as the tide fell, leaving the enemy ships impaled...",
			VoiceOverURL: "/assets/audio/tran_hung_dao_2.mp3",
			Questions: createQuestions([]QuestionData{
				{
					ID:       "q_thd2_1",
					Type:     "multiple_choice",
					Question: "What strategy did Trần Hưng Đạo use at Bạch Đằng River?",
					Options:  []string{"Direct naval battle", "Iron stakes and tide tactics", "Land siege", "Cavalry charge"},
					Answer:   "Iron stakes and tide tactics",
					Points:   15,
				},
				{
					ID:       "q_thd2_2",
					Type:     "drag_drop",
					Question: "Order the steps of the Bạch Đằng strategy:",
					Options:  []string{"Plant iron stakes", "Lure enemy fleet", "Wait for low tide", "Victory"},
					Answer:   []string{"Plant iron stakes", "Lure enemy fleet", "Wait for low tide", "Victory"},
					Points:   20,
				},
			}),
			XPReward:  100,
			MinScore:  75,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},

		// Lê Lợi lessons
		{
			ID:           "lesson_le_loi_1",
			CharacterID:  "char_le_loi",
			Title:        "The Ming Occupation",
			Order:        1,
			Story:        "After the fall of the Hồ dynasty, the Chinese Ming forces occupied Vietnam for 20 years. The people suffered under harsh rule and cultural suppression. But in the mountains of Thanh Hóa, a landlord named Lê Lợi began gathering patriots to resist...",
			VoiceOverURL: "/assets/audio/le_loi_1.mp3",
			Questions: createQuestions([]QuestionData{
				{
					ID:       "q_ll1_1",
					Type:     "multiple_choice",
					Question: "Which Chinese dynasty occupied Vietnam before Lê Lợi's rebellion?",
					Options:  []string{"Tang Dynasty", "Song Dynasty", "Ming Dynasty", "Qing Dynasty"},
					Answer:   "Ming Dynasty",
					Points:   10,
				},
				{
					ID:       "q_ll1_2",
					Type:     "fill_blank",
					Question: "Lê Lợi started his rebellion in the province of _____.",
					Answer:   "Thanh Hóa",
					Points:   15,
				},
			}),
			XPReward:  60,
			MinScore:  70,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:           "lesson_le_loi_2",
			CharacterID:  "char_le_loi",
			Title:        "The Legend of the Returned Sword",
			Order:        2,
			Story:        "Legend tells that Lê Lợi received a magical sword from the Dragon King to drive out the invaders. After victory, while boating on a lake in Thăng Long, a golden turtle surfaced and reclaimed the sword. The lake was thereafter called Hồ Hoàn Kiếm - Lake of the Returned Sword...",
			VoiceOverURL: "/assets/audio/le_loi_2.mp3",
			Questions: createQuestions([]QuestionData{
				{
					ID:       "q_ll2_1",
					Type:     "multiple_choice",
					Question: "What does 'Hồ Hoàn Kiếm' mean in English?",
					Options:  []string{"Dragon Lake", "Victory Lake", "Lake of the Returned Sword", "Golden Turtle Lake"},
					Answer:   "Lake of the Returned Sword",
					Points:   10,
				},
				{
					ID:       "q_ll2_2",
					Type:     "multiple_choice",
					Question: "Who gave Lê Lợi the magical sword according to legend?",
					Options:  []string{"Buddha", "Dragon King", "Jade Emperor", "Mountain Spirit"},
					Answer:   "Dragon King",
					Points:   15,
				},
			}),
			XPReward:  75,
			MinScore:  70,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},

		// Hai Bà Trưng lessons
		{
			ID:           "lesson_hai_ba_trung_1",
			CharacterID:  "char_hai_ba_trung",
			Title:        "Sisters of Resistance",
			Order:        1,
			Story:        "In 40 AD, when Chinese Han officials became increasingly oppressive, two noble sisters from Mê Linh decided to act. Trưng Trắc and Trưng Nhị, trained in martial arts and military strategy, could no longer watch their people suffer...",
			VoiceOverURL: "/assets/audio/hai_ba_trung_1.mp3",
			Questions: createQuestions([]QuestionData{
				{
					ID:       "q_hbt1_1",
					Type:     "multiple_choice",
					Question: "The Trưng Sisters led their rebellion in which year?",
					Options:  []string{"30 AD", "40 AD", "50 AD", "60 AD"},
					Answer:   "40 AD",
					Points:   10,
				},
				{
					ID:       "q_hbt1_2",
					Type:     "fill_blank",
					Question: "The Trưng Sisters were from _____ district.",
					Answer:   "Mê Linh",
					Points:   15,
				},
			}),
			XPReward:  65,
			MinScore:  70,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},

		// Ngô Quyền lessons
		{
			ID:           "lesson_ngo_quyen_1",
			CharacterID:  "char_ngo_quyen",
			Title:        "The End of a Thousand Years",
			Order:        1,
			Story:        "For nearly a thousand years, Vietnam had been under Chinese rule. But in 938 AD, a military commander named Ngô Quyền would change the course of history forever at the Bạch Đằng River...",
			VoiceOverURL: "/assets/audio/ngo_quyen_1.mp3",
			Questions: createQuestions([]QuestionData{
				{
					ID:       "q_nq1_1",
					Type:     "multiple_choice",
					Question: "How long had Vietnam been under Chinese rule before Ngô Quyền's victory?",
					Options:  []string{"500 years", "800 years", "Nearly 1000 years", "1200 years"},
					Answer:   "Nearly 1000 years",
					Points:   15,
				},
				{
					ID:       "q_nq1_2",
					Type:     "multiple_choice",
					Question: "Where did Ngô Quyền win his decisive victory?",
					Options:  []string{"Red River", "Bạch Đằng River", "Perfume River", "Mekong River"},
					Answer:   "Bạch Đằng River",
					Points:   10,
				},
			}),
			XPReward:  80,
			MinScore:  75,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},

		// Lý Thái Tổ lessons
		{
			ID:           "lesson_ly_thai_to_1",
			CharacterID:  "char_ly_thai_to",
			Title:        "The Move to Thăng Long",
			Order:        1,
			Story:        "In 1010, Emperor Lý Thái Tổ made a momentous decision - to move the capital from Hoa Lư to the site of modern-day Hanoi. He named it Thăng Long, meaning 'Rising Dragon', after witnessing a golden dragon ascending from the Red River...",
			VoiceOverURL: "/assets/audio/ly_thai_to_1.mp3",
			Questions: createQuestions([]QuestionData{
				{
					ID:       "q_ltt1_1",
					Type:     "multiple_choice",
					Question: "What does 'Thăng Long' mean?",
					Options:  []string{"Golden City", "Rising Dragon", "Royal Capital", "Sacred Land"},
					Answer:   "Rising Dragon",
					Points:   10,
				},
				{
					ID:       "q_ltt1_2",
					Type:     "multiple_choice",
					Question: "When did Lý Thái Tổ move the capital to Thăng Long?",
					Options:  []string{"1009", "1010", "1011", "1012"},
					Answer:   "1010",
					Points:   15,
				},
			}),
			XPReward:  70,
			MinScore:  70,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},

		// Hồ Chí Minh lessons
		{
			ID:           "lesson_ho_chi_minh_1",
			CharacterID:  "char_ho_chi_minh",
			Title:        "The Young Patriot",
			Order:        1,
			Story:        "Born as Nguyễn Sinh Cung in 1890, the future leader of Vietnam grew up witnessing French colonial oppression. As a young man, he left Vietnam on a French steamship, beginning a journey that would take him around the world and shape his revolutionary ideals...",
			VoiceOverURL: "/assets/audio/ho_chi_minh_1.mp3",
			Questions: createQuestions([]QuestionData{
				{
					ID:       "q_hcm1_1",
					Type:     "multiple_choice",
					Question: "What was Hồ Chí Minh's birth name?",
					Options:  []string{"Nguyễn Tất Thành", "Nguyễn Sinh Cung", "Nguyễn Ái Quốc", "Phan Bội Châu"},
					Answer:   "Nguyễn Sinh Cung",
					Points:   15,
				},
				{
					ID:       "q_hcm1_2",
					Type:     "fill_blank",
					Question: "Hồ Chí Minh was born in the year _____.",
					Answer:   "1890",
					Points:   10,
				},
			}),
			XPReward:  85,
			MinScore:  70,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:           "lesson_ho_chi_minh_2",
			CharacterID:  "char_ho_chi_minh",
			Title:        "Declaration of Independence",
			Order:        2,
			Story:        "On September 2, 1945, in Ba Đình Square, Hồ Chí Minh read the Declaration of Independence, establishing the Democratic Republic of Vietnam. His opening words quoted the American Declaration of Independence: 'All men are created equal...'",
			VoiceOverURL: "/assets/audio/ho_chi_minh_2.mp3",
			Questions: createQuestions([]QuestionData{
				{
					ID:       "q_hcm2_1",
					Type:     "multiple_choice",
					Question: "When did Hồ Chí Minh declare Vietnamese independence?",
					Options:  []string{"August 19, 1945", "September 2, 1945", "October 10, 1945", "December 19, 1946"},
					Answer:   "September 2, 1945",
					Points:   15,
				},
				{
					ID:       "q_hcm2_2",
					Type:     "multiple_choice",
					Question: "Where did Hồ Chí Minh read the Declaration of Independence?",
					Options:  []string{"Hoan Kiem Lake", "Ba Đình Square", "Presidential Palace", "National Assembly"},
					Answer:   "Ba Đình Square",
					Points:   10,
				},
			}),
			XPReward:  100,
			MinScore:  75,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	return lessons
}

// QuestionData represents question data for easier creation
type QuestionData struct {
	ID       string
	Type     string
	Question string
	Options  []string
	Answer   interface{}
	Points   int
	Metadata map[string]interface{}
}

// createQuestions converts QuestionData to JSON format
func createQuestions(questionsData []QuestionData) json.RawMessage {
	var questions []model.Question

	for _, qData := range questionsData {
		question := model.Question{
			ID:       qData.ID,
			Type:     qData.Type,
			Question: qData.Question,
			Options:  qData.Options,
			Answer:   qData.Answer,
			Points:   qData.Points,
			Metadata: qData.Metadata,
		}
		questions = append(questions, question)
	}

	data, _ := json.Marshal(questions)
	return json.RawMessage(data)
}

// SeedTimelines seeds the database with Vietnamese historical periods
func (s *SqliteService) SeedTimelines() error {
	timelines := s.getHistoricalTimelines()

	for _, timeline := range timelines {
		// Check if timeline already exists
		var existingTimeline model.Timeline
		if err := s.db.Where("id = ?", timeline.ID).First(&existingTimeline).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Timeline doesn't exist, create it
				if err := s.db.Create(&timeline).Error; err != nil {
					log.Printf("Error creating timeline %s: %v", timeline.Era, err)
					return err
				}
				log.Printf("Created timeline: %s", timeline.Era)
			} else {
				log.Printf("Error checking timeline %s: %v", timeline.Era, err)
				return err
			}
		} else {
			log.Printf("Timeline %s already exists, skipping", timeline.Era)
		}
	}

	log.Println("Timeline seeding completed successfully")
	return nil
}

// getHistoricalTimelines returns the Vietnamese historical periods
func (s *SqliteService) getHistoricalTimelines() []model.Timeline {
	now := time.Now()

	timelines := []model.Timeline{
		{
			ID:          "timeline_van_lang",
			Era:         "Văn Lang",
			StartYear:   -2879,
			EndYear:     intPtr(-258),
			Description: "The legendary first Vietnamese kingdom founded by Hùng Vương. Period of Lạc Việt civilization and bronze age culture.",
			KeyEvents: jsonArray([]string{
				"Foundation of Văn Lang kingdom",
				"Establishment of Hùng dynasty",
				"Development of bronze age culture",
				"Rice cultivation advancement",
			}),
			CharacterIds: jsonArray([]string{"char_hung_vuong"}),
			ImageURL:     "/assets/timeline/van_lang.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_au_lac",
			Era:         "Âu Lạc",
			StartYear:   -257,
			EndYear:     intPtr(-207),
			Description: "Kingdom established by An Dương Vương after conquering Văn Lang. Known for the construction of Cổ Loa citadel.",
			KeyEvents: jsonArray([]string{
				"An Dương Vương conquers Văn Lang",
				"Establishment of Âu Lạc kingdom",
				"Construction of Cổ Loa citadel",
				"Introduction of crossbow technology",
			}),
			CharacterIds: jsonArray([]string{}),
			ImageURL:     "/assets/timeline/au_lac.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_bac_thuoc",
			Era:         "Bắc thuộc",
			StartYear:   -111,
			EndYear:     intPtr(938),
			Description: "Period of Chinese domination lasting over 1000 years, interrupted by several major rebellions including the Trưng Sisters and Bà Triệu.",
			KeyEvents: jsonArray([]string{
				"Chinese Han conquest (-111)",
				"Trưng Sisters rebellion (40-43 AD)",
				"Bà Triệu rebellion (248 AD)",
				"Various resistance movements",
				"Cultural and technological exchanges",
			}),
			CharacterIds: jsonArray([]string{"char_hai_ba_trung", "char_ba_trieu"}),
			ImageURL:     "/assets/timeline/bac_thuoc.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_ngo",
			Era:         "Ngô",
			StartYear:   939,
			EndYear:     intPtr(965),
			Description: "First independent Vietnamese dynasty after Chinese rule, founded by Ngô Quyền following his victory at Bạch Đằng River.",
			KeyEvents: jsonArray([]string{
				"Battle of Bạch Đằng River (938)",
				"End of Chinese domination",
				"Establishment of Ngô dynasty",
				"Capital at Cổ Loa",
			}),
			CharacterIds: jsonArray([]string{"char_ngo_quyen"}),
			ImageURL:     "/assets/timeline/ngo.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_dinh_le",
			Era:         "Đinh - Tiền Lê",
			StartYear:   968,
			EndYear:     intPtr(1009),
			Description: "Period of the Đinh and Early Lê dynasties, establishing the foundation of independent Vietnamese state structure.",
			KeyEvents: jsonArray([]string{
				"Đinh Bộ Lĩnh unifies Vietnam",
				"Establishment of Đại Cồ Việt",
				"Capital moved to Hoa Lư",
				"Early Lê dynasty",
			}),
			CharacterIds: jsonArray([]string{"char_dinh_bo_linh"}),
			ImageURL:     "/assets/timeline/dinh_le.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_ly",
			Era:         "Lý",
			StartYear:   1009,
			EndYear:     intPtr(1225),
			Description: "Golden age of Vietnamese culture and Buddhism. Capital moved to Thăng Long (Hanoi). Period of territorial expansion and cultural development.",
			KeyEvents: jsonArray([]string{
				"Lý Thái Tổ founds dynasty (1009)",
				"Capital moved to Thăng Long (1010)",
				"Temple of Literature established",
				"Victory over Song China",
				"Buddhist cultural flourishing",
			}),
			CharacterIds: jsonArray([]string{"char_ly_thai_to", "char_ly_thuong_kiet"}),
			ImageURL:     "/assets/timeline/ly.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_tran",
			Era:         "Trần",
			StartYear:   1225,
			EndYear:     intPtr(1400),
			Description: "Dynasty famous for successfully repelling three Mongol invasions under the leadership of Trần Hưng Đạo. Period of military innovation and cultural development.",
			KeyEvents: jsonArray([]string{
				"Trần dynasty established (1225)",
				"First Mongol invasion repelled (1258)",
				"Second Mongol invasion defeated (1285)",
				"Third Mongol invasion crushed (1287-1288)",
				"Development of guerrilla warfare tactics",
			}),
			CharacterIds: jsonArray([]string{"char_tran_hung_dao"}),
			ImageURL:     "/assets/timeline/tran.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_ho",
			Era:         "Hồ",
			StartYear:   1400,
			EndYear:     intPtr(1407),
			Description: "Short-lived dynasty known for progressive reforms but ended with Chinese Ming occupation due to internal conflicts.",
			KeyEvents: jsonArray([]string{
				"Hồ Quý Ly seizes power (1400)",
				"Land and monetary reforms",
				"Internal rebellions",
				"Chinese Ming invasion (1407)",
			}),
			CharacterIds: jsonArray([]string{"char_ho_quy_ly"}),
			ImageURL:     "/assets/timeline/ho.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_ming_occupation",
			Era:         "Minh chiếm đóng",
			StartYear:   1407,
			EndYear:     intPtr(1428),
			Description: "Period of Chinese Ming occupation. Vietnamese resistance led by Lê Lợi culminated in independence and the founding of the Lê dynasty.",
			KeyEvents: jsonArray([]string{
				"Chinese Ming occupation begins (1407)",
				"Cultural suppression policies",
				"Lê Lợi begins resistance (1418)",
				"Lam Sơn uprising",
				"Liberation of Vietnam (1428)",
			}),
			CharacterIds: jsonArray([]string{"char_le_loi", "char_nguyen_trai"}),
			ImageURL:     "/assets/timeline/ming_occupation.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_later_le",
			Era:         "Hậu Lê",
			StartYear:   1428,
			EndYear:     intPtr(1788),
			Description: "Longest dynasty in Vietnamese history. Golden age under Lê Thánh Tông with territorial expansion and legal codification.",
			KeyEvents: jsonArray([]string{
				"Lê dynasty established (1428)",
				"Hồng Đức golden age (1470-1497)",
				"Conquest of Champa territories",
				"Hồng Đức legal code",
				"Division into Trịnh-Nguyễn period",
			}),
			CharacterIds: jsonArray([]string{"char_le_thanh_tong"}),
			ImageURL:     "/assets/timeline/later_le.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_mac",
			Era:         "Mạc",
			StartYear:   1527,
			EndYear:     intPtr(1677),
			Description: "Dynasty that controlled northern Vietnam during the divided period, competing with the restored Lê dynasty.",
			KeyEvents: jsonArray([]string{
				"Mạc Đăng Dung seizes power (1527)",
				"Control of northern territories",
				"Conflict with restored Lê dynasty",
				"Gradual territorial losses",
			}),
			CharacterIds: jsonArray([]string{"char_mac_dang_dung"}),
			ImageURL:     "/assets/timeline/mac.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_tay_son",
			Era:         "Tây Sơn",
			StartYear:   1778,
			EndYear:     intPtr(1802),
			Description: "Revolutionary period led by the Nguyễn brothers. Quang Trung defeated Chinese Qing invasion and implemented progressive reforms.",
			KeyEvents: jsonArray([]string{
				"Tây Sơn uprising begins (1778)",
				"Defeat of Trịnh and Nguyễn lords",
				"Quang Trung becomes emperor (1788)",
				"Victory over Qing invasion (1789)",
				"Social and educational reforms",
			}),
			CharacterIds: jsonArray([]string{"char_nguyen_hue"}),
			ImageURL:     "/assets/timeline/tay_son.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_nguyen",
			Era:         "Nguyễn",
			StartYear:   1802,
			EndYear:     intPtr(1945),
			Description: "Final Vietnamese imperial dynasty. Period of unification, modernization attempts, and eventual French colonization.",
			KeyEvents: jsonArray([]string{
				"Gia Long unifies Vietnam (1802)",
				"Capital established at Huế",
				"French colonial period begins (1858)",
				"Can Vuong resistance movements",
				"End of imperial rule (1945)",
			}),
			CharacterIds: jsonArray([]string{"char_nguyen_anh", "char_nguyen_du"}),
			ImageURL:     "/assets/timeline/nguyen.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_modern",
			Era:         "Cận đại",
			StartYear:   1858,
			EndYear:     intPtr(1975),
			Description: "Period of French colonization, independence movements, and wars of liberation culminating in reunification.",
			KeyEvents: jsonArray([]string{
				"French conquest begins (1858)",
				"Nationalist movements emerge",
				"World War II and Japanese occupation",
				"Declaration of Independence (1945)",
				"First Indochina War (1946-1954)",
				"Vietnam War (1955-1975)",
				"Reunification of Vietnam (1975)",
			}),
			CharacterIds: jsonArray([]string{
				"char_phan_boi_chau",
				"char_phan_chu_trinh",
				"char_vo_thi_sau",
				"char_ho_chi_minh",
			}),
			ImageURL:  "/assets/timeline/modern.jpg",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	return timelines
}
