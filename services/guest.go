package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/alphabatem/common/context"
	"github.com/google/uuid"
	"github.com/lac-hong-legacy/TechYouth-Be/model"
	"github.com/lac-hong-legacy/TechYouth-Be/shared"
	log "github.com/sirupsen/logrus"
)

type GuestService struct {
	context.DefaultService

	sqlSvc *SqliteService
}

const GUEST_SVC = "guest_svc"

func (svc GuestService) Id() string {
	return GUEST_SVC
}

func (svc *GuestService) Configure(ctx *context.Context) error {
	return svc.DefaultService.Configure(ctx)
}

func (svc *GuestService) Start() error {
	svc.sqlSvc = svc.Service(SQLITE_SVC).(*SqliteService)
	return nil
}

func (svc *GuestService) CreateOrGetSession(deviceID string) (*model.GuestSession, error) {
	session, err := svc.sqlSvc.GetSessionByDeviceID(deviceID)
	if err == nil && session != nil {
		session.LastActivity = time.Now()
		if err := svc.sqlSvc.UpdateSession(session); err != nil {
			log.Printf("Failed to update session activity: %v", err)
		}
		return session, nil
	}

	session = &model.GuestSession{
		ID:           uuid.New().String(),
		DeviceID:     deviceID,
		SessionStart: time.Now(),
		LastActivity: time.Now(),
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	session, err = svc.sqlSvc.CreateSession(session)
	if err != nil {
		return nil, err
	}

	// Create initial progress
	progress := &model.GuestProgress{
		ID:               uuid.New().String(),
		GuestSessionID:   session.ID,
		Hearts:           5,
		MaxHearts:        5,
		XP:               0,
		Level:            1,
		CompletedLessons: json.RawMessage("[]"),
		Streak:           0,
		TotalPlayTime:    0,
		AdsWatched:       0,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	_, err = svc.sqlSvc.CreateProgress(progress)
	if err != nil {
		log.Printf("Failed to create initial progress: %v", err)
		// Not returning error to avoid blocking session creation
	} else {
		log.Printf("Initial progress created: %+v", progress)
	}

	return session, nil
}

func (svc *GuestService) CanAccessLesson(sessionID, lessonID string) (bool, string, error) {
	progress, err := svc.sqlSvc.GetProgress(sessionID)
	if err != nil {
		return false, "", err
	}

	var completedLessons []string
	if err := json.Unmarshal(progress.CompletedLessons, &completedLessons); err != nil {
		return false, "Failed to parse completed lessons", err
	}

	if len(completedLessons) >= 2 {
		// Check if trying to access already completed lesson (for review)
		for _, completedID := range completedLessons {
			if completedID == lessonID {
				return true, "Review access", nil
			}
		}
		return false, "Guest limit reached. Please register to continue.", nil
	}

	// For simplicity, assume lessons are "lesson_1", "lesson_2", etc.
	// and only "lesson_1" and "lesson_2" are available to guests.
	// TODO: Replace with real lesson availability logic.
	allowedLessons := []string{"lesson_hung_vuong_1", "lesson_hung_vuong_2"}
	for _, allowedID := range allowedLessons {
		if allowedID == lessonID {
			return true, "Access granted", nil
		}
	}

	return false, "Lesson not available for guest users", nil
}

func (svc *GuestService) CompleteLesson(sessionID, lessonID string, score, timeSpent int) error {
	canAccess, reason, err := svc.CanAccessLesson(sessionID, lessonID)
	if err != nil {
		return shared.NewInternalError(err, "Failed to check lesson access")
	}

	if !canAccess {
		return shared.NewForbiddenError(fmt.Errorf("access denied: %s", reason), "Access denied")
	}

	progress, err := svc.sqlSvc.GetProgress(sessionID)
	if err != nil {
		return shared.NewInternalError(err, "Failed to get progress")
	}

	var completedLessons []string
	if err := json.Unmarshal(progress.CompletedLessons, &completedLessons); err != nil {
		return shared.NewInternalError(err, "Failed to parse completed lessons")
	}

	isAlreadyCompleted := false
	for _, completedID := range completedLessons {
		if completedID == lessonID {
			isAlreadyCompleted = true
			break
		}
	}

	if !isAlreadyCompleted {
		completedLessons = append(completedLessons, lessonID)
		completedLessonsJSON, err := json.Marshal(completedLessons)
		if err != nil {
			return shared.NewInternalError(err, "Failed to marshal completed lessons")
		}
		progress.CompletedLessons = completedLessonsJSON

		// Award XP for new completion
		progress.XP += calculateXP(score)
		progress.Level = calculateLevel(progress.XP)
	}

	// Update total play time
	progress.TotalPlayTime += timeSpent / 60 // Convert seconds to minutes

	// Save lesson attempt
	attempt := &model.GuestLessonAttempt{
		ID:             uuid.New().String(),
		GuestSessionID: sessionID,
		LessonID:       lessonID,
		IsCompleted:    true,
		Score:          score,
		TimeSpent:      timeSpent,
		AttemptsCount:  1,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := svc.sqlSvc.CreateLessonAttempt(attempt); err != nil {
		return shared.NewInternalError(err, "Failed to create lesson attempt")
	}

	// Update progress
	return svc.sqlSvc.UpdateProgress(progress)
}

func calculateXP(score int) int {
	baseXP := 50
	bonusXP := (score - 60) / 10 * 10 // Bonus for scores above 60%
	if bonusXP < 0 {
		bonusXP = 0
	}
	return baseXP + bonusXP
}

func calculateLevel(totalXP int) int {
	if totalXP < 100 {
		return 1
	}
	return (totalXP / 100) + 1
}

func (svc *GuestService) AddHeartsFromAd(sessionID string) error {
	progress, err := svc.sqlSvc.GetProgress(sessionID)
	if err != nil {
		return shared.NewInternalError(err, "Failed to get progress")
	}

	progress.Hearts = min(progress.Hearts+3, progress.MaxHearts)
	progress.AdsWatched++

	return svc.sqlSvc.UpdateProgress(progress)
}

func (svc *GuestService) LoseHeart(sessionID string) error {
	progress, err := svc.sqlSvc.GetProgress(sessionID)
	if err != nil {
		return shared.NewInternalError(err, "Failed to get progress")
	}

	if progress.Hearts > 0 {
		progress.Hearts--
	}

	return svc.sqlSvc.UpdateProgress(progress)
}

// func (svc *GuestService) CleanUpInactiveSessions(threshold time.Duration) error {
// 	sessions, err := svc.sqlSvc.GetAllSessions()
// 	if err != nil {
// 		return err
// 	}
