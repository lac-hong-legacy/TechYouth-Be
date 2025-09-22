package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/alphabatem/common/context"
	"github.com/google/uuid"
	"github.com/lac-hong-legacy/TechYouth-Be/model"
	log "github.com/sirupsen/logrus"
)

type UserService struct {
	context.DefaultService
	sqlSvc *SqliteService
}

const USER_SVC = "user_svc"

func (svc UserService) Id() string {
	return USER_SVC
}

func (svc *UserService) Configure(ctx *context.Context) error {
	return svc.DefaultService.Configure(ctx)
}

func (svc *UserService) Start() error {
	svc.sqlSvc = svc.Service(SQLITE_SVC).(*SqliteService)
	return nil
}

// Initialize user profile after registration
func (svc *UserService) InitializeUserProfile(userID string, birthYear int) error {
	// Create user progress
	progress := &model.UserProgress{
		ID:                 uuid.New().String(),
		UserID:             userID,
		Hearts:             5,
		MaxHearts:          5,
		XP:                 0,
		Level:              1,
		CompletedLessons:   json.RawMessage("[]"),
		UnlockedCharacters: json.RawMessage("[]"),
		Streak:             0,
		TotalPlayTime:      0,
		LastHeartReset:     &[]time.Time{time.Now()}[0],
		LastActivityDate:   &[]time.Time{time.Now()}[0],
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if _, err := svc.sqlSvc.CreateUserProgress(progress); err != nil {
		return err
	}

	// Create zodiac-based spirit
	spiritType := svc.getZodiacAnimal(birthYear)
	spirit := &model.Spirit{
		ID:        uuid.New().String(),
		UserID:    userID,
		Type:      spiritType,
		Stage:     1,
		XP:        0,
		XPToNext:  500,
		Name:      "",
		ImageURL:  svc.getSpiritImageURL(spiritType, 1),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err := svc.sqlSvc.CreateSpirit(spirit)
	return err
}

func (svc *UserService) getZodiacAnimal(birthYear int) string {
	zodiacAnimals := []string{
		"rat", "ox", "tiger", "cat", "dragon", "snake",
		"horse", "goat", "monkey", "rooster", "dog", "pig",
	}
	return zodiacAnimals[(birthYear-4)%12]
}

func (svc *UserService) getSpiritImageURL(spiritType string, stage int) string {
	return fmt.Sprintf("/assets/spirits/%s_stage_%d.png", spiritType, stage)
}

func (svc *UserService) ResetDailyHearts() error {
	// Get all users who haven't had hearts reset today
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	users, err := svc.sqlSvc.GetUsersForHeartReset(startOfDay)
	if err != nil {
		return err
	}

	for _, user := range users {
		if err := svc.resetUserHearts(user.UserID); err != nil {
			log.Printf("Failed to reset hearts for user %s: %v", user.UserID, err)
		}
	}

	return nil
}

func (svc *UserService) resetUserHearts(userID string) error {
	progress, err := svc.sqlSvc.GetUserProgress(userID)
	if err != nil {
		return err
	}

	progress.Hearts = progress.MaxHearts
	now := time.Now()
	progress.LastHeartReset = &now
	progress.UpdatedAt = now

	return svc.sqlSvc.UpdateUserProgress(progress)
}

// Complete lesson for registered user
func (svc *UserService) CompleteLesson(userID, lessonID string, score, timeSpent int) error {
	progress, err := svc.sqlSvc.GetUserProgress(userID)
	if err != nil {
		return err
	}

	// Parse completed lessons
	var completedLessons []string
	if err := json.Unmarshal(progress.CompletedLessons, &completedLessons); err != nil {
		return err
	}

	// Check if already completed
	isNewCompletion := true
	for _, completedID := range completedLessons {
		if completedID == lessonID {
			isNewCompletion = false
			break
		}
	}

	if isNewCompletion {
		// Add to completed lessons
		completedLessons = append(completedLessons, lessonID)
		completedLessonsJSON, err := json.Marshal(completedLessons)
		if err != nil {
			return err
		}
		progress.CompletedLessons = completedLessonsJSON

		// Award XP
		xpGained := svc.calculateXP(score)
		progress.XP += xpGained
		oldLevel := progress.Level
		progress.Level = svc.calculateLevel(progress.XP)

		// Update spirit XP
		if err := svc.updateSpiritXP(userID, xpGained); err != nil {
			log.Printf("Failed to update spirit XP: %v", err)
		}

		// Check for level up
		if progress.Level > oldLevel {
			log.Printf("User %s leveled up to %d", userID, progress.Level)
			// TODO: Trigger level up rewards/notifications
		}

		// Check if character should be unlocked
		if err := svc.checkCharacterUnlock(userID, lessonID); err != nil {
			log.Printf("Failed to check character unlock: %v", err)
		}
	}

	// Update play time
	progress.TotalPlayTime += timeSpent / 60
	progress.UpdatedAt = time.Now()

	// Update streak
	if err := svc.updateStreak(userID); err != nil {
		log.Printf("Failed to update streak: %v", err)
	}

	return svc.sqlSvc.UpdateUserProgress(progress)
}

func (svc *UserService) calculateXP(score int) int {
	baseXP := 50
	bonusXP := max(0, (score-60)/10*10) // Bonus for scores above 60%
	return baseXP + bonusXP
}

func (svc *UserService) calculateLevel(totalXP int) int {
	if totalXP < 100 {
		return 1
	}
	return (totalXP / 100) + 1
}

func (svc *UserService) updateSpiritXP(userID string, xpGained int) error {
	spirit, err := svc.sqlSvc.GetUserSpirit(userID)
	if err != nil {
		return err
	}

	spirit.XP += xpGained

	// Check for spirit evolution
	for spirit.XP >= spirit.XPToNext && spirit.Stage < 5 {
		spirit.XP -= spirit.XPToNext
		spirit.Stage++
		spirit.XPToNext = svc.getNextStageXPRequirement(spirit.Stage)
		spirit.ImageURL = svc.getSpiritImageURL(spirit.Type, spirit.Stage)

		log.Printf("Spirit evolved to stage %d for user %s", spirit.Stage, userID)
		// TODO: Trigger evolution animation/notification
	}

	return svc.sqlSvc.UpdateSpirit(spirit)
}

func (svc *UserService) getNextStageXPRequirement(stage int) int {
	requirements := map[int]int{
		2: 1000,
		3: 2000,
		4: 3500,
		5: 5000,
	}
	if req, exists := requirements[stage]; exists {
		return req
	}
	return 5000 // Max stage
}

func (svc *UserService) updateStreak(userID string) error {
	progress, err := svc.sqlSvc.GetUserProgress(userID)
	if err != nil {
		return err
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if progress.LastActivityDate == nil {
		progress.Streak = 1
	} else {
		lastActivityDay := time.Date(
			progress.LastActivityDate.Year(),
			progress.LastActivityDate.Month(),
			progress.LastActivityDate.Day(),
			0, 0, 0, 0, progress.LastActivityDate.Location(),
		)

		daysDiff := int(today.Sub(lastActivityDay).Hours() / 24)

		switch daysDiff {
		case 0:
			// Same day, no change to streak
		case 1:
			// Next day, increment streak
			progress.Streak++
		default:
			// Missed day(s), reset streak
			progress.Streak = 1
		}
	}

	progress.LastActivityDate = &now
	return svc.sqlSvc.UpdateUserProgress(progress)
}

func (svc *UserService) checkCharacterUnlock(userID, lessonID string) error {
	// TODO: Implement character unlock logic based on lesson completion
	// This would check if enough lessons are completed for a character
	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
