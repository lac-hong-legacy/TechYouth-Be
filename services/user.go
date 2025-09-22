// services/user.go
package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/alphabatem/common/context"
	"github.com/google/uuid"
	"github.com/lac-hong-legacy/TechYouth-Be/dto"
	"github.com/lac-hong-legacy/TechYouth-Be/model"
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

	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for range ticker.C {
			svc.ResetDailyHearts()
		}
	}()

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

// ==================== PROFILE METHODS ====================

func (svc *UserService) GetUserProfile(userID string) (*dto.UserProfileResponse, error) {
	user, err := svc.sqlSvc.GetUser(userID)
	if err != nil {
		return nil, err
	}

	return &dto.UserProfileResponse{
		UserID:      user.ID,
		Email:       user.Email,
		Username:    user.Username,
		JoinedAt:    user.CreatedAt,
		LastLoginAt: user.LastLogin,
	}, nil
}

func (svc *UserService) UpdateUserProfile(userID string, req dto.UpdateProfileRequest) (*dto.UserProfileResponse, error) {
	user, err := svc.sqlSvc.GetUser(userID)
	if err != nil {
		return nil, err
	}

	// Update user fields if provided
	if req.Username != "" && req.Username != user.Username {
		// Check if username is already taken
		if _, err := svc.sqlSvc.GetUserByUsername(req.Username); err == nil {
			return nil, fmt.Errorf("username already exists")
		}
		user.Username = req.Username
	}

	if req.BirthYear > 0 {
		// Update spirit if birth year changed
		if err := svc.updateUserSpirit(userID, req.BirthYear); err != nil {
			log.Printf("Failed to update user spirit: %v", err)
		}
	}

	if err := svc.sqlSvc.UpdateUser(user); err != nil {
		return nil, err
	}

	return svc.GetUserProfile(userID)
}

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
	nextLevelXP := (currentLevel) * 100 // Next level requires currentLevel * 100 XP
	return nextLevelXP - currentXP
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

// ==================== STATISTICS AND ANALYTICS ====================

func (svc *UserService) GetUserStats(userID string) (*dto.UserStatsResponse, error) {
	stats, err := svc.sqlSvc.GetUserStats(userID)
	if err != nil {
		return nil, err
	}

	// Convert to DTO
	response := &dto.UserStatsResponse{
		UserID:             userID,
		Level:              stats["level"].(int),
		XP:                 stats["xp"].(int),
		Hearts:             stats["hearts"].(int),
		Streak:             stats["streak"].(int),
		TotalPlayTime:      stats["total_play_time"].(int),
		CompletedLessons:   stats["completed_lessons"].(int),
		UnlockedCharacters: stats["unlocked_characters"].(int),
		Achievements:       stats["achievements"].(int),
		Rank:               stats["rank"].(int),
		SpiritStage:        stats["spirit_stage"].(int),
		SpiritType:         stats["spirit_type"].(string),
	}

	// Add weekly and monthly stats
	response.WeeklyStats = svc.getWeeklyStats(userID)
	response.MonthlyStats = svc.getMonthlyStats(userID)

	return response, nil
}

func (svc *UserService) getWeeklyStats(userID string) map[string]interface{} {
	// TODO: Implement weekly statistics calculation
	return map[string]interface{}{
		"lessons_completed": 0,
		"xp_gained":         0,
		"time_played":       0,
	}
}

func (svc *UserService) getMonthlyStats(userID string) map[string]interface{} {
	// TODO: Implement monthly statistics calculation
	return map[string]interface{}{
		"lessons_completed": 0,
		"xp_gained":         0,
		"time_played":       0,
		"streak_record":     0,
	}
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

func (svc *UserService) initializeUserProfile(userID string, birthYear int) error {
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

	if _, err := svc.sqlSvc.CreateSpirit(spirit); err != nil {
		return err
	}

	return nil
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
