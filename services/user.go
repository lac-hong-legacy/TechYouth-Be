// services/user.go
package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/alphabatem/common/context"
	"github.com/google/uuid"
	"github.com/lac-hong-legacy/ven_api/dto"
	"github.com/lac-hong-legacy/ven_api/model"
	"github.com/lac-hong-legacy/ven_api/shared"
	log "github.com/sirupsen/logrus"
)

type UserService struct {
	context.DefaultService

	contentSvc *ContentService
	sqlSvc     *SqliteService
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
	svc.contentSvc = svc.Service(CONTENT_SVC).(*ContentService)

	go svc.startHeartResetScheduler()

	return nil
}

func (svc *UserService) startHeartResetScheduler() {
	for {
		now := time.Now()
		nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		durationUntilMidnight := nextMidnight.Sub(now)

		timer := time.NewTimer(durationUntilMidnight)
		<-timer.C

		svc.ResetDailyHearts()

		ticker := time.NewTicker(1 * time.Minute)
		for range ticker.C {
			svc.ResetDailyHearts()
		}
	}
}

// Initialize user profile after registration
func (svc *UserService) InitializeUserProfile(userID string, birthYear int) error {
	// Create user progress
	progressID, _ := uuid.NewV7()
	progress := &model.UserProgress{
		ID:                 progressID.String(),
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
	spiritID, _ := uuid.NewV7()
	spirit := &model.Spirit{
		ID:        spiritID.String(),
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

	if _, err := svc.sqlSvc.CreateSpirit(spirit); err != nil {
		return err
	}

	return nil
}

func (svc *UserService) getZodiacAnimal(birthYear int) string {
	zodiacAnimals := []string{
		"rat", "ox", "tiger", "rabbit", "dragon", "snake",
		"horse", "goat", "monkey", "rooster", "dog", "pig",
	}
	return zodiacAnimals[(birthYear-4)%12]
}

func (svc *UserService) getSpiritImageURL(spiritType string, stage int) string {
	return fmt.Sprintf("/assets/spirits/%s_stage_%d.png", spiritType, stage)
}

// Daily heart reset (should be called by cron job)
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
	level := 1
	requiredXP := 100 // Base XP for level 2

	for totalXP >= requiredXP {
		totalXP -= requiredXP
		level++
		requiredXP = int(float64(requiredXP) * 1.5) // Each level requires 1.5x more XP
	}

	return level
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

// ==================== PROFILE METHODS ====================

func (svc *UserService) updateUserSpirit(userID string, birthYear int) error {
	spirit, err := svc.sqlSvc.GetUserSpirit(userID)
	if err != nil {
		// If no spirit exists, create one
		if strings.Contains(err.Error(), "not found") {
			return svc.InitializeUserProfile(userID, birthYear)
		}
		return err
	}

	newType := svc.getZodiacAnimal(birthYear)
	if spirit.Type != newType {
		spirit.Type = newType
		spirit.ImageURL = svc.getSpiritImageURL(newType, spirit.Stage)
		return svc.sqlSvc.UpdateSpirit(spirit)
	}

	return nil
}

// ==================== PROGRESS METHODS ====================

func (svc *UserService) GetUserProgress(userID string) (*dto.UserProgressResponse, error) {
	progress, err := svc.sqlSvc.GetUserProgress(userID)
	if err != nil {
		return nil, err
	}

	spirit, err := svc.sqlSvc.GetUserSpirit(userID)
	if err != nil {
		return nil, err
	}

	var completedLessons []string
	if err := json.Unmarshal(progress.CompletedLessons, &completedLessons); err != nil {
		completedLessons = []string{}
	}

	var unlockedCharacters []string
	if err := json.Unmarshal(progress.UnlockedCharacters, &unlockedCharacters); err != nil {
		unlockedCharacters = []string{}
	}

	// Get recent achievements
	achievements, err := svc.sqlSvc.GetUserAchievements(userID)
	if err != nil {
		log.Printf("Failed to get user achievements: %v", err)
		achievements = []model.UserAchievement{}
	}

	recentAchievements := make([]dto.AchievementResponse, 0)
	for _, ua := range achievements {
		// Show achievements from last 7 days
		if time.Since(ua.UnlockedAt) <= 7*24*time.Hour {
			recentAchievements = append(recentAchievements, dto.AchievementResponse{
				ID:          ua.Achievement.ID,
				Name:        ua.Achievement.Name,
				Description: ua.Achievement.Description,
				BadgeURL:    ua.Achievement.BadgeURL,
				Category:    ua.Achievement.Category,
				XPReward:    ua.Achievement.XPReward,
				UnlockedAt:  &ua.UnlockedAt,
			})
		}
	}

	return &dto.UserProgressResponse{
		UserID:             userID,
		Hearts:             progress.Hearts,
		MaxHearts:          progress.MaxHearts,
		XP:                 progress.XP,
		Level:              progress.Level,
		XPToNextLevel:      svc.calculateXPToNextLevel(progress.XP),
		CompletedLessons:   completedLessons,
		UnlockedCharacters: unlockedCharacters,
		Streak:             progress.Streak,
		TotalPlayTime:      progress.TotalPlayTime,
		LastHeartReset:     progress.LastHeartReset,
		LastActivity:       progress.LastActivityDate,
		Spirit: dto.SpiritResponse{
			ID:       spirit.ID,
			Type:     spirit.Type,
			Stage:    spirit.Stage,
			XP:       spirit.XP,
			XPToNext: spirit.XPToNext,
			Name:     spirit.Name,
			ImageURL: spirit.ImageURL,
		},
		Achievements: recentAchievements,
	}, nil
}

func (svc *UserService) calculateXPToNextLevel(currentXP int) int {
	currentLevel := svc.calculateLevel(currentXP)

	// Calculate total XP needed for next level
	totalXPForNextLevel := svc.getTotalXPForLevel(currentLevel + 1)

	return totalXPForNextLevel - currentXP
}

// Helper function to get total XP required to reach a specific level
func (svc *UserService) getTotalXPForLevel(targetLevel int) int {
	if targetLevel <= 1 {
		return 0
	}

	totalXP := 0
	requiredXP := 100 // Base XP for level 2

	for level := 2; level <= targetLevel; level++ {
		totalXP += requiredXP
		requiredXP = int(float64(requiredXP) * 1.5) // Each level requires 1.5x more XP
	}

	return totalXP
}

// ==================== LESSON ACCESS AND COMPLETION ====================

func (svc *UserService) CheckLessonAccess(userID, lessonID string) (*dto.LessonAccessResponse, error) {
	progress, err := svc.sqlSvc.GetUserProgress(userID)
	if err != nil {
		return nil, err
	}

	// Check hearts
	if progress.Hearts <= 0 {
		return &dto.LessonAccessResponse{
			CanAccess:    false,
			Reason:       "Not enough hearts",
			HeartsNeeded: 1,
		}, nil
	}

	// TODO: Add lesson prerequisite checking logic
	// For now, all lessons are accessible if user has hearts

	return &dto.LessonAccessResponse{
		CanAccess:    true,
		Reason:       "Access granted",
		HeartsNeeded: 0,
	}, nil
}

func (svc *UserService) checkCharacterUnlockForLesson(userID, lessonID string) string {
	// TODO: Implement character unlock logic
	// This should check if completing this lesson unlocks a character
	// Based on your PRD's 3-tier system (Common: 1-2 lessons, Rare: 2-3, Legendary: 4-5)
	return ""
}

// ==================== HEARTS MANAGEMENT ====================

func (svc *UserService) GetHeartStatus(userID string) (*dto.HeartStatusResponse, error) {
	progress, err := svc.sqlSvc.GetUserProgress(userID)
	if err != nil {
		return nil, err
	}

	// Calculate next reset time (next day at 00:00)
	now := time.Now()
	nextReset := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())

	// TODO: Track ads watched today
	adsWatchedToday := 0

	return &dto.HeartStatusResponse{
		Hearts:          progress.Hearts,
		MaxHearts:       progress.MaxHearts,
		NextResetTime:   &nextReset,
		CanWatchAd:      adsWatchedToday < 5 && progress.Hearts < progress.MaxHearts,
		AdsWatchedToday: adsWatchedToday,
	}, nil
}

func (svc *UserService) AddHearts(userID, source string, amount int) (*dto.HeartStatusResponse, error) {
	progress, err := svc.sqlSvc.GetUserProgress(userID)
	if err != nil {
		return nil, err
	}

	// Validate source and amount
	switch source {
	case "ad":
		if amount != 3 {
			return nil, fmt.Errorf("invalid amount for ad hearts")
		}
		// TODO: Check ad watch limits
	case "daily_reset":
		progress.Hearts = progress.MaxHearts
		now := time.Now()
		progress.LastHeartReset = &now
	case "purchase":
		// TODO: Validate purchase
	default:
		return nil, fmt.Errorf("invalid heart source")
	}

	if source != "daily_reset" {
		progress.Hearts = min(progress.Hearts+amount, progress.MaxHearts)
	}

	if err := svc.sqlSvc.UpdateUserProgress(progress); err != nil {
		return nil, err
	}

	return svc.GetHeartStatus(userID)
}

func (svc *UserService) LoseHeart(userID string) (*dto.HeartStatusResponse, error) {
	progress, err := svc.sqlSvc.GetUserProgress(userID)
	if err != nil {
		return nil, err
	}

	if progress.Hearts > 0 {
		progress.Hearts--
		if err := svc.sqlSvc.UpdateUserProgress(progress); err != nil {
			return nil, err
		}
	}

	return svc.GetHeartStatus(userID)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ==================== COLLECTION METHODS ====================

func (svc *UserService) GetUserCollection(userID string) (*dto.CollectionResponse, error) {
	progress, err := svc.sqlSvc.GetUserProgress(userID)
	if err != nil {
		return nil, err
	}

	// Get user's unlocked characters
	var unlockedCharacterIDs []string
	if err := json.Unmarshal(progress.UnlockedCharacters, &unlockedCharacterIDs); err != nil {
		unlockedCharacterIDs = []string{}
	}

	// Get all characters to show collection progress
	allCharacters, err := svc.sqlSvc.GetCharactersByDynasty("") // Get all
	if err != nil {
		return nil, err
	}

	// Map characters to responses with unlock status
	characterResponses := make([]dto.CharacterResponse, len(allCharacters))
	unlockedCount := 0
	rarityBreakdown := make(map[string]int)
	dynastyBreakdown := make(map[string]int)

	for i, char := range allCharacters {
		isUnlocked := svc.isCharacterUnlocked(char.ID, unlockedCharacterIDs)

		characterResponses[i] = dto.CharacterResponse{
			ID:          char.ID,
			Name:        char.Name,
			Dynasty:     char.Dynasty,
			Rarity:      char.Rarity,
			BirthYear:   char.BirthYear,
			DeathYear:   char.DeathYear,
			Description: char.Description,
			FamousQuote: char.FamousQuote,
			ImageURL:    char.ImageURL,
			IsUnlocked:  isUnlocked,
		}

		if isUnlocked {
			unlockedCount++
		}

		// Update breakdowns
		rarityBreakdown[char.Rarity]++
		dynastyBreakdown[char.Dynasty]++
	}

	// Get user achievements
	userAchievements, err := svc.sqlSvc.GetUserAchievements(userID)
	if err != nil {
		return nil, err
	}

	achievementResponses := make([]dto.AchievementResponse, len(userAchievements))
	for i, ua := range userAchievements {
		achievementResponses[i] = dto.AchievementResponse{
			ID:          ua.Achievement.ID,
			Name:        ua.Achievement.Name,
			Description: ua.Achievement.Description,
			BadgeURL:    ua.Achievement.BadgeURL,
			Category:    ua.Achievement.Category,
			XPReward:    ua.Achievement.XPReward,
			UnlockedAt:  &ua.UnlockedAt,
		}
	}

	completionRate := float64(unlockedCount) / float64(len(allCharacters)) * 100

	return &dto.CollectionResponse{
		Characters: dto.CharacterCollectionResponse{
			Characters: characterResponses,
			Total:      len(allCharacters),
			Unlocked:   unlockedCount,
		},
		Achievements: achievementResponses,
		Stats: dto.CollectionStatsResponse{
			TotalCharacters:    len(allCharacters),
			UnlockedCharacters: unlockedCount,
			CompletionRate:     completionRate,
			RarityBreakdown:    rarityBreakdown,
			DynastyBreakdown:   dynastyBreakdown,
		},
	}, nil
}

func (svc *UserService) isCharacterUnlocked(characterID string, unlockedIDs []string) bool {
	for _, id := range unlockedIDs {
		if id == characterID {
			return true
		}
	}
	return false
}

// ==================== LEADERBOARD METHODS ====================

func (svc *UserService) GetWeeklyLeaderboard(limit int, currentUserID string) (*dto.LeaderboardResponse, error) {
	users, err := svc.sqlSvc.GetWeeklyLeaderboard(limit)
	if err != nil {
		return nil, err
	}

	return svc.buildLeaderboardResponse("weekly", users, currentUserID)
}

func (svc *UserService) GetMonthlyLeaderboard(limit int, currentUserID string) (*dto.LeaderboardResponse, error) {
	users, err := svc.sqlSvc.GetMonthlyLeaderboard(limit)
	if err != nil {
		return nil, err
	}

	return svc.buildLeaderboardResponse("monthly", users, currentUserID)
}

func (svc *UserService) GetAllTimeLeaderboard(limit int, currentUserID string) (*dto.LeaderboardResponse, error) {
	users, err := svc.sqlSvc.GetAllTimeLeaderboard(limit)
	if err != nil {
		return nil, err
	}

	return svc.buildLeaderboardResponse("all_time", users, currentUserID)
}

func (svc *UserService) buildLeaderboardResponse(period string, users []model.UserProgress, currentUserID string) (*dto.LeaderboardResponse, error) {
	topUsers := make([]dto.LeaderboardUserResponse, len(users))
	var currentUser dto.LeaderboardUserResponse

	for i, user := range users {
		// Get user details
		userDetails, err := svc.sqlSvc.GetUser(user.UserID)
		if err != nil {
			log.Printf("Failed to get user details for %s: %v", user.UserID, err)
			continue
		}

		// Get user's spirit
		spirit, err := svc.sqlSvc.GetUserSpirit(user.UserID)
		if err != nil {
			log.Printf("Failed to get spirit for user %s: %v", user.UserID, err)
			spirit = &model.Spirit{Type: "unknown", Stage: 1}
		}

		leaderboardUser := dto.LeaderboardUserResponse{
			UserID:      user.UserID,
			Username:    userDetails.Username,
			Level:       user.Level,
			XP:          user.XP,
			Rank:        i + 1,
			SpiritType:  spirit.Type,
			SpiritStage: spirit.Stage,
		}

		topUsers[i] = leaderboardUser

		if user.UserID == currentUserID {
			currentUser = leaderboardUser
		}
	}

	// If current user is not in top list, get their rank
	if currentUserID != "" && currentUser.UserID == "" {
		rank, err := svc.sqlSvc.GetUserRank(currentUserID)
		if err == nil {
			userProgress, err := svc.sqlSvc.GetUserProgress(currentUserID)
			if err == nil {
				userDetails, err := svc.sqlSvc.GetUser(currentUserID)
				if err == nil {
					spirit, err := svc.sqlSvc.GetUserSpirit(currentUserID)
					if err != nil {
						spirit = &model.Spirit{Type: "unknown", Stage: 1}
					}

					currentUser = dto.LeaderboardUserResponse{
						UserID:      currentUserID,
						Username:    userDetails.Username,
						Level:       userProgress.Level,
						XP:          userProgress.XP,
						Rank:        rank,
						SpiritType:  spirit.Type,
						SpiritStage: spirit.Stage,
					}
				}
			}
		}
	}

	return &dto.LeaderboardResponse{
		Period:      period,
		CurrentUser: currentUser,
		TopUsers:    topUsers,
	}, nil
}

// ==================== SOCIAL FEATURES ====================

func (svc *UserService) CreateShareContent(userID string, req dto.ShareRequest) (*dto.ShareResponse, error) {
	progress, err := svc.sqlSvc.GetUserProgress(userID)
	if err != nil {
		return nil, err
	}

	var shareText string
	var shareImage string

	switch req.Type {
	case "achievement":
		shareText = fmt.Sprintf("🎉 I just unlocked a new achievement in Ven! Level %d and still learning Vietnamese history!", progress.Level)
		shareImage = "/assets/share/achievement.png"
	case "character_unlock":
		shareText = fmt.Sprintf("📚 Just unlocked a new historical character in Ven! Learning Vietnamese history has never been this fun! Level %d", progress.Level)
		shareImage = "/assets/share/character_unlock.png"
	case "level_up":
		shareText = fmt.Sprintf("⭐ Level UP! I'm now level %d in Ven - the gamified Vietnamese history app! 🇻🇳", progress.Level)
		shareImage = "/assets/share/level_up.png"
	default:
		shareText = fmt.Sprintf("🎮 Learning Vietnamese history with Ven! Currently level %d - join me!", progress.Level)
		shareImage = "/assets/share/general.png"
	}

	// Generate share URL (could include referral tracking)
	shareURL := fmt.Sprintf("https://ven.app/shared/%s/%s", req.Type, userID)

	return &dto.ShareResponse{
		ShareURL:   shareURL,
		ShareImage: shareImage,
		ShareText:  shareText,
		Platforms:  []string{"facebook", "instagram", "tiktok", "twitter"},
	}, nil
}

// ==================== USERNAME VALIDATION ====================

func (svc *UserService) CheckUsernameAvailability(username string) (bool, error) {
	if username == "" {
		return false, fmt.Errorf("username cannot be empty")
	}

	if len(username) < 3 || len(username) > 20 {
		return false, fmt.Errorf("username must be between 3 and 20 characters")
	}

	// Check for valid characters (alphanumeric and underscores only)
	for _, char := range username {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '_') {
			return false, fmt.Errorf("username can only contain letters, numbers, and underscores")
		}
	}

	// Check if username is already taken
	_, err := svc.sqlSvc.GetUserByUsername(username)
	if err == nil {
		return false, nil // Username exists
	}

	return true, nil // Username is available
}

// ==================== USER PROFILE METHODS ====================

func (svc *UserService) GetUserProfile(userID string) (*dto.UserProfileResponse, error) {
	user, err := svc.sqlSvc.GetUserProfile(userID)
	if err != nil {
		return nil, shared.NewInternalError(err, "Failed to get user profile")
	}

	stats, err := svc.sqlSvc.GetUserStats(userID)
	if err != nil {
		// Don't fail if stats can't be retrieved, just log it
		log.WithError(err).WithField("userID", userID).Error("Failed to get user stats")
		stats = &dto.UserStats{} // Empty stats
	}

	profile := &dto.UserProfileResponse{
		ID:            user.ID,
		Username:      user.Username,
		Email:         user.Email,
		Role:          user.Role,
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt,
		LastLoginAt:   user.LastLoginAt,
		LastLoginIP:   user.LastLoginIP,
		IsActive:      user.IsActive,
		Stats:         *stats,
	}

	return profile, nil
}

func (svc *UserService) UpdateUserProfile(userID string, req dto.UpdateProfileRequest) (*dto.UserProfileResponse, error) {
	// Validate updates
	updates := make(map[string]interface{})

	if req.Username != "" {
		// Check if username is available (excluding current user)
		var existingUser model.User
		err := svc.sqlSvc.Db().Where("LOWER(username) = LOWER(?) AND id != ? AND deleted_at IS NULL",
			req.Username, userID).First(&existingUser).Error

		if err == nil {
			return nil, shared.NewBadRequestError(fmt.Errorf("username taken"), "Username is already taken")
		}

		updates["username"] = req.Username
	}

	if req.Email != "" {
		// Check if email is available (excluding current user)
		var existingUser model.User
		err := svc.sqlSvc.Db().Where("LOWER(email) = LOWER(?) AND id != ? AND deleted_at IS NULL",
			req.Email, userID).First(&existingUser).Error

		if err == nil {
			return nil, shared.NewBadRequestError(fmt.Errorf("email taken"), "Email is already taken")
		}

		updates["email"] = req.Email
		updates["email_verified"] = false // Reset verification if email changes
	}

	if len(updates) > 0 {
		err := svc.sqlSvc.UpdateUserProfile(userID, updates)
		if err != nil {
			return nil, shared.NewInternalError(err, "Failed to update profile")
		}
	}

	// Return updated profile
	return svc.GetUserProfile(userID)
}

func (svc *UserService) IsEmailAvailable(email string) (bool, error) {
	available, err := svc.sqlSvc.IsEmailAvailable(email)
	if err != nil {
		return false, shared.NewInternalError(err, "Failed to check email availability")
	}
	return available, nil
}

// ==================== SESSION MANAGEMENT ====================

func (svc *UserService) GetUserSessions(userID, currentSessionID string) (*dto.SessionListResponse, error) {
	sessions, err := svc.sqlSvc.GetUserSessions(userID)
	if err != nil {
		return nil, shared.NewInternalError(err, "Failed to get user sessions")
	}

	sessionInfos := make([]dto.UserSessionInfo, len(sessions))
	for i, session := range sessions {
		sessionInfos[i] = dto.UserSessionInfo{
			ID:        session.ID,
			DeviceID:  session.DeviceID,
			IP:        session.IP,
			UserAgent: session.UserAgent,
			CreatedAt: session.CreatedAt,
			LastUsed:  session.LastUsed,
			IsActive:  session.IsActive,
			IsCurrent: session.ID == currentSessionID,
		}
	}

	return &dto.SessionListResponse{
		Sessions: sessionInfos,
		Total:    len(sessionInfos),
	}, nil
}

func (svc *UserService) RevokeUserSession(userID, sessionID string) error {
	err := svc.sqlSvc.DeactivateSession(sessionID, userID)
	if err != nil {
		return shared.NewInternalError(err, "Failed to revoke session")
	}
	return nil
}

// ==================== SECURITY SETTINGS ====================

func (svc *UserService) GetSecuritySettings(userID string) (*dto.SecuritySettings, error) {
	settings, err := svc.sqlSvc.GetSecuritySettings(userID)
	if err != nil {
		return nil, shared.NewInternalError(err, "Failed to get security settings")
	}
	return settings, nil
}

func (svc *UserService) UpdateSecuritySettings(userID string, req dto.UpdateSecuritySettingsRequest) (*dto.SecuritySettings, error) {
	// Validate settings
	if req.SessionTimeout != nil {
		if *req.SessionTimeout < 15 || *req.SessionTimeout > 10080 { // 15 minutes to 7 days
			return nil, shared.NewBadRequestError(fmt.Errorf("invalid session timeout"),
				"Session timeout must be between 15 minutes and 7 days")
		}
	}

	err := svc.sqlSvc.UpdateSecuritySettings(userID, req)
	if err != nil {
		return nil, shared.NewInternalError(err, "Failed to update security settings")
	}

	// Return updated settings
	return svc.GetSecuritySettings(userID)
}

// ==================== AUDIT LOGS ====================

func (svc *UserService) GetUserAuditLogs(userID string, page, limit int) (*dto.AuditLogResponse, error) {
	logs, total, err := svc.sqlSvc.GetUserAuditLogs(userID, page, limit)
	if err != nil {
		return nil, shared.NewInternalError(err, "Failed to get audit logs")
	}

	auditLogs := make([]dto.AuthAuditLog, len(logs))
	for i, log := range logs {
		auditLogs[i] = dto.AuthAuditLog{
			ID:        log.ID,
			UserID:    log.UserID,
			Action:    log.Action,
			IP:        log.IP,
			UserAgent: log.UserAgent,
			Timestamp: log.Timestamp,
			Success:   log.Success,
			Details:   log.Details,
		}
	}

	return &dto.AuditLogResponse{
		Logs:  auditLogs,
		Total: int(total),
		Page:  page,
		Limit: limit,
	}, nil
}

// ==================== ADMIN USER MANAGEMENT ====================

func (svc *UserService) AdminGetUsers(page, limit int, search string) (*dto.AdminUserListResponse, error) {
	users, total, err := svc.sqlSvc.AdminGetUsers(page, limit, search)
	if err != nil {
		return nil, shared.NewInternalError(err, "Failed to get users")
	}

	userInfos := make([]dto.AdminUserInfo, len(users))
	for i, user := range users {
		userInfos[i] = dto.AdminUserInfo{
			ID:             user.ID,
			Username:       user.Username,
			Email:          user.Email,
			Role:           user.Role,
			EmailVerified:  user.EmailVerified,
			IsActive:       user.IsActive,
			CreatedAt:      user.CreatedAt,
			LastLoginAt:    user.LastLoginAt,
			FailedAttempts: user.FailedAttempts,
			LockedUntil:    user.LockedUntil,
		}
	}

	return &dto.AdminUserListResponse{
		Users: userInfos,
		Total: int(total),
		Page:  page,
		Limit: limit,
	}, nil
}

func (svc *UserService) AdminUpdateUser(userID string, req dto.AdminUpdateUserRequest) (*dto.AdminUserInfo, error) {
	updates := make(map[string]interface{})

	if req.Role != nil {
		// Validate role
		validRoles := []string{model.RoleUser, model.RoleAdmin, model.RoleMod}
		isValidRole := false
		for _, role := range validRoles {
			if *req.Role == role {
				isValidRole = true
				break
			}
		}
		if !isValidRole {
			return nil, shared.NewBadRequestError(fmt.Errorf("invalid role"), "Invalid role specified")
		}
		updates["role"] = *req.Role
	}

	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if len(updates) > 0 {
		err := svc.sqlSvc.AdminUpdateUser(userID, updates)
		if err != nil {
			return nil, shared.NewInternalError(err, "Failed to update user")
		}
	}

	// Get updated user info
	user, err := svc.sqlSvc.GetUserByID(userID)
	if err != nil {
		return nil, shared.NewInternalError(err, "Failed to get updated user")
	}

	return &dto.AdminUserInfo{
		ID:             user.ID,
		Username:       user.Username,
		Email:          user.Email,
		Role:           user.Role,
		EmailVerified:  user.EmailVerified,
		IsActive:       user.IsActive,
		CreatedAt:      user.CreatedAt,
		LastLoginAt:    user.LastLoginAt,
		FailedAttempts: user.FailedAttempts,
		LockedUntil:    user.LockedUntil,
	}, nil
}

func (svc *UserService) AdminDeleteUser(userID string) error {
	err := svc.sqlSvc.AdminDeleteUser(userID)
	if err != nil {
		return shared.NewInternalError(err, "Failed to delete user")
	}
	return nil
}

// ==================== UTILITY METHODS ====================

func (svc *UserService) GetUserInfo(userID string) (*dto.UserInfo, error) {
	user, err := svc.sqlSvc.GetUserByID(userID)
	if err != nil {
		return nil, shared.NewInternalError(err, "Failed to get user info")
	}

	return &dto.UserInfo{
		ID:            user.ID,
		Username:      user.Username,
		Email:         user.Email,
		Role:          user.Role,
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt,
		LastLoginAt:   user.LastLoginAt,
	}, nil
}

func (svc *UserService) ValidateUser(userID string) (*model.User, error) {
	user, err := svc.sqlSvc.GetUserByID(userID)
	if err != nil {
		return nil, shared.NewUnauthorizedError(err, "User not found")
	}

	if !user.IsActive {
		return nil, shared.NewUnauthorizedError(fmt.Errorf("account inactive"), "Account is inactive")
	}

	if user.DeletedAt != nil {
		return nil, shared.NewUnauthorizedError(fmt.Errorf("account deleted"), "Account has been deleted")
	}

	return user, nil
}

// ==================== SEARCH AND FILTERING ====================

func (svc *UserService) SearchUsers(query string, filters map[string]interface{}) ([]dto.UserInfo, error) {
	// This is a basic search implementation
	// You can enhance it based on your specific requirements

	var users []model.User
	dbQuery := svc.sqlSvc.Db().Model(&model.User{}).Where("deleted_at IS NULL")

	if query != "" {
		searchPattern := "%" + strings.ToLower(query) + "%"
		dbQuery = dbQuery.Where("LOWER(username) LIKE ? OR LOWER(email) LIKE ?", searchPattern, searchPattern)
	}

	if role, ok := filters["role"].(string); ok && role != "" {
		dbQuery = dbQuery.Where("role = ?", role)
	}

	if isActive, ok := filters["is_active"].(bool); ok {
		dbQuery = dbQuery.Where("is_active = ?", isActive)
	}

	if emailVerified, ok := filters["email_verified"].(bool); ok {
		dbQuery = dbQuery.Where("email_verified = ?", emailVerified)
	}

	err := dbQuery.Order("created_at DESC").Limit(100).Find(&users).Error
	if err != nil {
		return nil, shared.NewInternalError(err, "Failed to search users")
	}

	userInfos := make([]dto.UserInfo, len(users))
	for i, user := range users {
		userInfos[i] = dto.UserInfo{
			ID:            user.ID,
			Username:      user.Username,
			Email:         user.Email,
			Role:          user.Role,
			EmailVerified: user.EmailVerified,
			CreatedAt:     user.CreatedAt,
			LastLoginAt:   user.LastLoginAt,
		}
	}

	return userInfos, nil
}

// ==================== BULK OPERATIONS ====================

func (svc *UserService) BulkUpdateUsers(userIDs []string, updates map[string]interface{}) error {
	if len(userIDs) == 0 {
		return shared.NewBadRequestError(fmt.Errorf("no users specified"), "No users specified for update")
	}

	err := svc.sqlSvc.Db().Model(&model.User{}).Where("id IN ?", userIDs).Updates(updates).Error
	if err != nil {
		return shared.NewInternalError(err, "Failed to bulk update users")
	}

	return nil
}

func (svc *UserService) BulkDeactivateUsers(userIDs []string) error {
	updates := map[string]interface{}{
		"is_active":  false,
		"updated_at": time.Now(),
	}
	return svc.BulkUpdateUsers(userIDs, updates)
}

func (svc *UserService) BulkActivateUsers(userIDs []string) error {
	updates := map[string]interface{}{
		"is_active":  true,
		"updated_at": time.Now(),
	}
	return svc.BulkUpdateUsers(userIDs, updates)
}

// ==================== USER STATISTICS ====================

func (svc *UserService) GetUserStatistics() (map[string]interface{}, error) {
	var totalUsers, activeUsers, verifiedUsers, adminUsers int64

	// Total users
	svc.sqlSvc.Db().Model(&model.User{}).Where("deleted_at IS NULL").Count(&totalUsers)

	// Active users
	svc.sqlSvc.Db().Model(&model.User{}).Where("deleted_at IS NULL AND is_active = ?", true).Count(&activeUsers)

	// Verified users
	svc.sqlSvc.Db().Model(&model.User{}).Where("deleted_at IS NULL AND email_verified = ?", true).Count(&verifiedUsers)

	// Admin users
	svc.sqlSvc.Db().Model(&model.User{}).Where("deleted_at IS NULL AND role = ?", model.RoleAdmin).Count(&adminUsers)

	stats := map[string]interface{}{
		"total_users":    totalUsers,
		"active_users":   activeUsers,
		"verified_users": verifiedUsers,
		"admin_users":    adminUsers,
		"verification_rate": func() float64 {
			if totalUsers == 0 {
				return 0
			}
			return float64(verifiedUsers) / float64(totalUsers) * 100
		}(),
	}

	return stats, nil
}

// ==================== HELPER METHODS ====================

func (svc *UserService) maskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}

	username := parts[0]
	domain := parts[1]

	if len(username) <= 2 {
		return email
	}

	masked := username[:1] + strings.Repeat("*", len(username)-2) + username[len(username)-1:]
	return masked + "@" + domain
}

func (svc *UserService) maskPhone(phone string) string {
	if len(phone) < 4 {
		return phone
	}
	return phone[:2] + strings.Repeat("*", len(phone)-4) + phone[len(phone)-2:]
}
