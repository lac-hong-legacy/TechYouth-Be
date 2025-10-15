package repositories

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lac-hong-legacy/ven_api/model"
	"gorm.io/gorm"
)

type ContentRepository struct {
	BaseRepository
}

func NewContentRepository(db *gorm.DB) *ContentRepository {
	return &ContentRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (ds *ContentRepository) GetProgress(sessionID string) (*model.GuestProgress, error) {
	var progress model.GuestProgress
	if err := ds.db.Where("guest_session_id = ?", sessionID).First(&progress).Error; err != nil {
		return nil, err
	}
	return &progress, nil
}

func (ds *ContentRepository) CreateProgress(progress *model.GuestProgress) (*model.GuestProgress, error) {
	id, _ := uuid.NewV7()
	progress.ID = id.String()
	if err := ds.db.Create(progress).Error; err != nil {
		return nil, err
	}
	return progress, nil
}

func (ds *ContentRepository) UpdateProgress(progress *model.GuestProgress) error {
	if err := ds.db.Save(progress).Error; err != nil {
		return err
	}
	return nil
}

func (ds *ContentRepository) CreateLessonAttempt(attempt *model.GuestLessonAttempt) error {
	id, _ := uuid.NewV7()
	attempt.ID = id.String()
	if err := ds.db.Create(attempt).Error; err != nil {
		return err
	}
	return nil
}

func (ds *ContentRepository) CreateCharacter(character *model.Character) (*model.Character, error) {
	if character.ID == "" {
		id, _ := uuid.NewV7()
		character.ID = id.String()
	}
	character.CreatedAt = time.Now()
	character.UpdatedAt = time.Now()

	if err := ds.db.Create(character).Error; err != nil {
		return nil, err
	}
	return character, nil
}

func (ds *ContentRepository) GetCharacter(id string) (*model.Character, error) {
	var character model.Character
	if err := ds.db.Where("id = ?", id).First(&character).Error; err != nil {
		return nil, err
	}
	return &character, nil
}

func (ds *ContentRepository) GetCharactersByDynasty(dynasty string) ([]model.Character, error) {
	var characters []model.Character
	query := ds.db.Model(&model.Character{})

	if dynasty != "" {
		query = query.Where("dynasty = ?", dynasty)
	}

	if err := query.Find(&characters).Error; err != nil {
		return nil, err
	}
	return characters, nil
}

func (ds *ContentRepository) GetCharactersByRarity(rarity string) ([]model.Character, error) {
	var characters []model.Character
	if err := ds.db.Where("rarity = ?", rarity).Find(&characters).Error; err != nil {
		return nil, err
	}
	return characters, nil
}

func (ds *ContentRepository) UpdateCharacter(character *model.Character) error {
	character.UpdatedAt = time.Now()
	if err := ds.db.Save(character).Error; err != nil {
		return err
	}
	return nil
}

// ==================== LESSON METHODS ====================

func (ds *ContentRepository) CreateLesson(lesson *model.Lesson) (*model.Lesson, error) {
	if lesson.ID == "" {
		id, _ := uuid.NewV7()
		lesson.ID = id.String()
	}
	lesson.CreatedAt = time.Now()
	lesson.UpdatedAt = time.Now()

	if err := ds.db.Create(lesson).Error; err != nil {
		return nil, err
	}
	return lesson, nil
}

func (ds *ContentRepository) GetLesson(id string) (*model.Lesson, error) {
	var lesson model.Lesson
	if err := ds.db.Preload("Character").Where("id = ?", id).First(&lesson).Error; err != nil {
		return nil, err
	}
	return &lesson, nil
}

func (ds *ContentRepository) GetLessonsByCharacter(characterID string) ([]model.Lesson, error) {
	var lessons []model.Lesson
	if err := ds.db.Preload("Character").Where("character_id = ? AND is_active = ?", characterID, true).
		Order("\"order\" ASC").Find(&lessons).Error; err != nil {
		return nil, err
	}
	return lessons, nil
}

func (ds *ContentRepository) UpdateLesson(lesson *model.Lesson) error {
	lesson.UpdatedAt = time.Now()
	if err := ds.db.Save(lesson).Error; err != nil {
		return err
	}
	return nil
}

// ==================== TIMELINE METHODS ====================

func (ds *ContentRepository) CreateTimeline(timeline *model.Timeline) (*model.Timeline, error) {
	if timeline.ID == "" {
		id, _ := uuid.NewV7()
		timeline.ID = id.String()
	}
	timeline.CreatedAt = time.Now()
	timeline.UpdatedAt = time.Now()

	if err := ds.db.Create(timeline).Error; err != nil {
		return nil, err
	}
	return timeline, nil
}

func (ds *ContentRepository) GetTimeline() ([]model.Timeline, error) {
	var timelines []model.Timeline
	if err := ds.db.Order("\"order\" ASC").Find(&timelines).Error; err != nil {
		return nil, err
	}
	return timelines, nil
}

func (ds *ContentRepository) GetTimelineByEra(era string) ([]model.Timeline, error) {
	var timelines []model.Timeline
	if err := ds.db.Where("era = ?", era).Order("\"order\" ASC").Find(&timelines).Error; err != nil {
		return nil, err
	}
	return timelines, nil
}

// ==================== USER PROGRESS METHODS ====================

func (ds *ContentRepository) CreateUserProgress(progress *model.UserProgress) (*model.UserProgress, error) {
	if progress.ID == "" {
		id, _ := uuid.NewV7()
		progress.ID = id.String()
	}
	progress.CreatedAt = time.Now()
	progress.UpdatedAt = time.Now()

	if err := ds.db.Create(progress).Error; err != nil {
		return nil, err
	}
	return progress, nil
}

func (ds *ContentRepository) GetUserProgress(userID string) (*model.UserProgress, error) {
	var progress model.UserProgress
	if err := ds.db.Where("user_id = ?", userID).First(&progress).Error; err != nil {
		return nil, err
	}
	return &progress, nil
}

func (ds *ContentRepository) UpdateUserProgress(progress *model.UserProgress) error {
	progress.UpdatedAt = time.Now()
	if err := ds.db.Save(progress).Error; err != nil {
		return err
	}
	return nil
}

func (ds *ContentRepository) GetUsersForHeartReset(since time.Time) ([]model.UserProgress, error) {
	var users []model.UserProgress
	if err := ds.db.Where("last_heart_reset < ? OR last_heart_reset IS NULL", since).
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// ==================== SPIRIT METHODS ====================

func (ds *ContentRepository) CreateSpirit(spirit *model.Spirit) (*model.Spirit, error) {
	if spirit.ID == "" {
		id, _ := uuid.NewV7()
		spirit.ID = id.String()
	}
	spirit.CreatedAt = time.Now()
	spirit.UpdatedAt = time.Now()

	if err := ds.db.Create(spirit).Error; err != nil {
		return nil, err
	}
	return spirit, nil
}

func (ds *ContentRepository) GetUserSpirit(userID string) (*model.Spirit, error) {
	var spirit model.Spirit
	if err := ds.db.Where("user_id = ?", userID).First(&spirit).Error; err != nil {
		return nil, err
	}
	return &spirit, nil
}

func (ds *ContentRepository) UpdateSpirit(spirit *model.Spirit) error {
	spirit.UpdatedAt = time.Now()
	if err := ds.db.Save(spirit).Error; err != nil {
		return err
	}
	return nil
}

// ==================== ACHIEVEMENT METHODS ====================

func (ds *ContentRepository) CreateAchievement(achievement *model.Achievement) (*model.Achievement, error) {
	if achievement.ID == "" {
		id, _ := uuid.NewV7()
		achievement.ID = id.String()
	}
	achievement.CreatedAt = time.Now()
	achievement.UpdatedAt = time.Now()

	if err := ds.db.Create(achievement).Error; err != nil {
		return nil, err
	}
	return achievement, nil
}

func (ds *ContentRepository) GetActiveAchievements() ([]model.Achievement, error) {
	var achievements []model.Achievement
	if err := ds.db.Where("is_active = ?", true).Find(&achievements).Error; err != nil {
		return nil, err
	}
	return achievements, nil
}

func (ds *ContentRepository) CreateUserAchievement(userAchievement *model.UserAchievement) error {
	if userAchievement.ID == "" {
		id, _ := uuid.NewV7()
		userAchievement.ID = id.String()
	}
	userAchievement.CreatedAt = time.Now()
	userAchievement.UnlockedAt = time.Now()

	if err := ds.db.Create(userAchievement).Error; err != nil {
		return err
	}
	return nil
}

func (ds *ContentRepository) GetUserAchievements(userID string) ([]model.UserAchievement, error) {
	var userAchievements []model.UserAchievement
	if err := ds.db.Preload("Achievement").Where("user_id = ?", userID).
		Find(&userAchievements).Error; err != nil {
		return nil, err
	}
	return userAchievements, nil
}

// ==================== LEADERBOARD METHODS ====================

func (ds *ContentRepository) GetWeeklyLeaderboard(limit int) ([]model.UserProgress, error) {
	var users []model.UserProgress
	weekAgo := time.Now().AddDate(0, 0, -7)

	if err := ds.db.Where("updated_at >= ?", weekAgo).
		Order("xp DESC").Limit(limit).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (ds *ContentRepository) GetMonthlyLeaderboard(limit int) ([]model.UserProgress, error) {
	var users []model.UserProgress
	monthAgo := time.Now().AddDate(0, -1, 0)

	if err := ds.db.Where("updated_at >= ?", monthAgo).
		Order("xp DESC").Limit(limit).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (ds *ContentRepository) GetAllTimeLeaderboard(limit int) ([]model.UserProgress, error) {
	var users []model.UserProgress
	if err := ds.db.Order("xp DESC").Limit(limit).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (ds *ContentRepository) GetUserRank(userID string) (int, error) {
	var rank int64
	userProgress, err := ds.GetUserProgress(userID)
	if err != nil {
		return 0, err
	}

	if err := ds.db.Model(&model.UserProgress{}).
		Where("xp > ?", userProgress.XP).Count(&rank).Error; err != nil {
		return 0, err
	}

	return int(rank + 1), nil // +1 because rank is 0-indexed
}

// ==================== CONTENT SEARCH AND FILTERING ====================

func (ds *ContentRepository) SearchCharacters(query string, era string, dynasty string, rarity string, limit int) ([]model.Character, error) {
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
		return nil, err
	}
	return characters, nil
}

func (ds *ContentRepository) SaveUserQuestionAnswer(answer *model.UserQuestionAnswer) error {
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

	return err
}

func (ds *ContentRepository) GetUserQuestionAnswers(userID, lessonID string) ([]model.UserQuestionAnswer, error) {
	var answers []model.UserQuestionAnswer
	if err := ds.db.Where("user_id = ? AND lesson_id = ?", userID, lessonID).
		Find(&answers).Error; err != nil {
		return nil, err
	}
	return answers, nil
}

func (ds *ContentRepository) GetUserQuestionAnswer(userID, lessonID, questionID string) (*model.UserQuestionAnswer, error) {
	var answer model.UserQuestionAnswer
	if err := ds.db.Where("user_id = ? AND lesson_id = ? AND question_id = ?",
		userID, lessonID, questionID).First(&answer).Error; err != nil {
		return nil, err
	}
	return &answer, nil
}

func (ds *ContentRepository) DeleteUserQuestionAnswers(userID, lessonID string) error {
	if err := ds.db.Where("user_id = ? AND lesson_id = ?", userID, lessonID).
		Delete(&model.UserQuestionAnswer{}).Error; err != nil {
		return err
	}
	return nil
}
