// services/content.go
package services

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alphabatem/common/context"
	"github.com/lac-hong-legacy/ven_api/dto"
	"github.com/lac-hong-legacy/ven_api/model"
	log "github.com/sirupsen/logrus"
)

type ContentService struct {
	context.DefaultService
	sqlSvc *PostgresService
}

const CONTENT_SVC = "content_svc"

func (svc ContentService) Id() string {
	return CONTENT_SVC
}

func (svc *ContentService) Configure(ctx *context.Context) error {
	return svc.DefaultService.Configure(ctx)
}

func (svc *ContentService) Start() error {
	svc.sqlSvc = svc.Service(POSTGRES_SVC).(*PostgresService)
	return nil
}

// ==================== TIMELINE METHODS ====================

func (svc *ContentService) GetTimeline() (*dto.TimelineCollectionResponse, error) {
	timelines, err := svc.sqlSvc.GetTimeline()
	if err != nil {
		return nil, err
	}

	// Group timelines by era
	eraMap := make(map[string][]model.Timeline)
	for _, timeline := range timelines {
		eraMap[timeline.Era] = append(eraMap[timeline.Era], timeline)
	}

	var eras []dto.TimelineEraResponse

	// Process each era
	for eraName, eraTimelines := range eraMap {
		var dynasties []dto.TimelineDynastyResponse
		var allEraCharacters []model.Character
		eraUnlocked := true

		// Process each dynasty within the era
		for _, timeline := range eraTimelines {
			// Get characters for this dynasty
			var characterIDs []string
			if timeline.CharacterIds != nil {
				if err := json.Unmarshal(timeline.CharacterIds, &characterIDs); err != nil {
					log.Printf("Failed to unmarshal character IDs for timeline %s: %v", timeline.Dynasty, err)
					characterIDs = []string{}
				}
			}

			// Fetch characters by IDs
			characters := []model.Character{}
			for _, charID := range characterIDs {
				char, err := svc.sqlSvc.GetCharacter(charID)
				if err != nil {
					log.Printf("Failed to get character %s: %v", charID, err)
					continue
				}
				characters = append(characters, *char)
			}

			characterResponses := make([]dto.CharacterResponse, len(characters))
			for j, char := range characters {
				characterResponses[j] = svc.mapCharacterToResponse(&char)
			}

			// Create dynasty response
			dynasty := dto.TimelineDynastyResponse{
				Dynasty:    timeline.Dynasty, // Use actual dynasty name
				StartYear:  timeline.StartYear,
				EndYear:    timeline.EndYear,
				Characters: characterResponses,
				IsUnlocked: timeline.IsUnlocked,
				Progress:   svc.calculateDynastyProgress(characters),
			}

			dynasties = append(dynasties, dynasty)
			allEraCharacters = append(allEraCharacters, characters...)

			if !timeline.IsUnlocked {
				eraUnlocked = false
			}
		}

		// Create era response
		era := dto.TimelineEraResponse{
			Era:        eraName,
			Dynasties:  dynasties,
			IsUnlocked: eraUnlocked,
			Progress:   svc.calculateDynastyProgress(allEraCharacters),
		}

		eras = append(eras, era)
	}

	return &dto.TimelineCollectionResponse{
		Eras: eras,
	}, nil
}

func (svc *ContentService) calculateDynastyProgress(characters []model.Character) float64 {
	if len(characters) == 0 {
		return 0
	}

	unlockedCount := 0
	for _, char := range characters {
		if char.IsUnlocked {
			unlockedCount++
		}
	}

	return float64(unlockedCount) / float64(len(characters)) * 100
}

// ==================== CHARACTER METHODS ====================

func (svc *ContentService) GetCharacters(dynasty, rarity string) (*dto.CharacterCollectionResponse, error) {
	var characters []model.Character
	var err error

	if dynasty != "" {
		characters, err = svc.sqlSvc.GetCharactersByDynasty(dynasty)
	} else if rarity != "" {
		characters, err = svc.sqlSvc.GetCharactersByRarity(rarity)
	} else {
		characters, err = svc.sqlSvc.GetCharactersByDynasty("") // Get all
	}

	if err != nil {
		return nil, err
	}

	characterResponses := make([]dto.CharacterResponse, len(characters))
	unlockedCount := 0

	for i, char := range characters {
		characterResponses[i] = svc.mapCharacterToResponse(&char)
		if char.IsUnlocked {
			unlockedCount++
		}
		lessons, err := svc.sqlSvc.GetLessonsByCharacter(char.ID)
		if err != nil {
			log.Printf("Failed to get lesson count for character %s: %v", char.ID, err)
		} else {
			characterResponses[i].LessonCount = len(lessons)
		}
	}

	return &dto.CharacterCollectionResponse{
		Characters: characterResponses,
		Total:      len(characters),
		Unlocked:   unlockedCount,
	}, nil
}

func (svc *ContentService) GetCharacterDetails(characterID string) (*dto.CharacterResponse, error) {
	character, err := svc.sqlSvc.GetCharacter(characterID)
	if err != nil {
		return nil, err
	}

	response := svc.mapCharacterToResponse(character)

	// Add lesson count
	lessons, err := svc.sqlSvc.GetLessonsByCharacter(characterID)
	if err != nil {
		log.Printf("Failed to get lesson count for character %s: %v", characterID, err)
	} else {
		response.LessonCount = len(lessons)
	}

	return &response, nil
}

func (svc *ContentService) mapCharacterToResponse(char *model.Character) dto.CharacterResponse {
	var achievements []string
	if char.Achievements != nil {
		if err := json.Unmarshal(char.Achievements, &achievements); err != nil {
			log.Printf("Failed to unmarshal achievements for character %s: %v", char.ID, err)
			achievements = []string{}
		}
	}

	return dto.CharacterResponse{
		ID:           char.ID,
		Name:         char.Name,
		Era:          char.Era,
		Dynasty:      char.Dynasty,
		Rarity:       char.Rarity,
		BirthYear:    char.BirthYear,
		DeathYear:    char.DeathYear,
		Description:  char.Description,
		FamousQuote:  char.FamousQuote,
		Achievements: achievements,
		ImageURL:     char.ImageURL,
		IsUnlocked:   char.IsUnlocked,
	}
}

// ==================== LESSON METHODS ====================

func (svc *ContentService) GetCharacterLessons(characterID string) ([]dto.LessonResponse, error) {
	lessons, err := svc.sqlSvc.GetLessonsByCharacter(characterID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.LessonResponse, len(lessons))
	for i, lesson := range lessons {
		responses[i] = svc.mapLessonToResponse(&lesson)
	}

	return responses, nil
}

func (svc *ContentService) GetLessonContent(lessonID string) (*dto.LessonResponse, error) {
	lesson, err := svc.sqlSvc.GetLesson(lessonID)
	if err != nil {
		return nil, err
	}

	response := svc.mapLessonToResponse(lesson)
	return &response, nil
}

func (svc *ContentService) mapLessonToResponse(lesson *model.Lesson) dto.LessonResponse {
	var questions []dto.QuestionResponse
	if lesson.Questions != nil {
		var rawQuestions []model.Question
		if err := json.Unmarshal(lesson.Questions, &rawQuestions); err != nil {
			log.Printf("Failed to unmarshal questions for lesson %s: %v", lesson.ID, err)
			questions = []dto.QuestionResponse{}
		} else {
			questions = make([]dto.QuestionResponse, len(rawQuestions))
			for i, q := range rawQuestions {
				questions[i] = dto.QuestionResponse{
					ID:       q.ID,
					Type:     q.Type,
					Question: q.Question,
					Options:  q.Options,
					Points:   q.Points,
					Metadata: q.Metadata,
					// Note: We don't include the Answer in the response for security
				}
			}
		}
	}

	return dto.LessonResponse{
		ID:          lesson.ID,
		CharacterID: lesson.CharacterID,
		Title:       lesson.Title,
		Order:       lesson.Order,
		Story:       lesson.Story,

		// Media Content
		VideoURL:      lesson.VideoURL,
		SubtitleURL:   lesson.SubtitleURL,
		ThumbnailURL:  lesson.ThumbnailURL,
		VideoDuration: lesson.VideoDuration,
		CanSkipAfter:  lesson.CanSkipAfter,
		HasSubtitles:  lesson.HasSubtitles,

		Questions: questions,
		XPReward:  lesson.XPReward,
		MinScore:  lesson.MinScore,
		Character: svc.mapCharacterToResponse(&lesson.Character),
	}
}

// ==================== SEARCH METHODS ====================

func (svc *ContentService) SearchContent(req dto.SearchRequest) (*dto.SearchResponse, error) {
	characters, err := svc.sqlSvc.SearchCharacters(req.Query, req.Era, req.Dynasty, req.Rarity, req.Limit)
	if err != nil {
		return nil, err
	}

	characterResponses := make([]dto.CharacterResponse, len(characters))
	for i, char := range characters {
		characterResponses[i] = svc.mapCharacterToResponse(&char)
	}

	return &dto.SearchResponse{
		Characters: characterResponses,
		Total:      len(characters),
	}, nil
}

// ==================== ADMIN METHODS ====================

func (svc *ContentService) CreateCharacter(character *model.Character) (*dto.CharacterResponse, error) {
	created, err := svc.sqlSvc.CreateCharacter(character)
	if err != nil {
		return nil, err
	}

	response := svc.mapCharacterToResponse(created)
	return &response, nil
}

func (svc *ContentService) CreateLesson(lesson *model.Lesson) (*dto.LessonResponse, error) {
	created, err := svc.sqlSvc.CreateLesson(lesson)
	if err != nil {
		return nil, err
	}

	response := svc.mapLessonToResponse(created)
	return &response, nil
}

func (svc *ContentService) CreateLessonFromRequest(req dto.CreateLessonRequest) (*dto.LessonResponse, error) {
	// Validate character exists
	_, err := svc.sqlSvc.GetCharacter(req.CharacterID)
	if err != nil {
		return nil, fmt.Errorf("character not found: %v", err)
	}

	// Convert questions to JSON
	var questionsJSON json.RawMessage
	if len(req.Questions) > 0 {
		questions := make([]model.Question, len(req.Questions))
		for i, q := range req.Questions {
			if q.ID == "" {
				q.ID = fmt.Sprintf("q_%d", i+1)
			}
			questions[i] = model.Question{
				ID:       q.ID,
				Type:     q.Type,
				Question: q.Question,
				Options:  q.Options,
				Answer:   q.Answer,
				Points:   q.Points,
				Metadata: q.Metadata,
			}
		}
		questionsJSON, err = json.Marshal(questions)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal questions: %v", err)
		}
	}

	// Set defaults
	if req.XPReward == 0 {
		req.XPReward = 50
	}
	if req.MinScore == 0 {
		req.MinScore = 60
	}
	if req.CanSkipAfter == 0 {
		req.CanSkipAfter = 5
	}

	lesson := &model.Lesson{
		CharacterID:   req.CharacterID,
		Title:         req.Title,
		Order:         req.Order,
		Story:         req.Story,
		VideoURL:      req.VideoURL,
		SubtitleURL:   req.SubtitleURL,
		ThumbnailURL:  req.ThumbnailURL,
		VideoDuration: req.VideoDuration,
		CanSkipAfter:  req.CanSkipAfter,
		HasSubtitles:  req.HasSubtitles,
		Questions:     questionsJSON,
		XPReward:      req.XPReward,
		MinScore:      req.MinScore,
		IsActive:      true,
	}

	return svc.CreateLesson(lesson)
}

// ==================== VALIDATION METHODS ====================

func (svc *ContentService) ValidateLessonAnswers(lessonID string, userAnswers map[string]interface{}) (*dto.ValidateLessonResponse, error) {
	lesson, err := svc.sqlSvc.GetLesson(lessonID)
	if err != nil {
		return nil, err
	}

	var questions []model.Question
	if err := json.Unmarshal(lesson.Questions, &questions); err != nil {
		return nil, fmt.Errorf("failed to parse lesson questions: %v", err)
	}

	totalPoints := 0
	earnedPoints := 0

	for _, question := range questions {
		totalPoints += question.Points

		userAnswer, exists := userAnswers[question.ID]
		if exists && svc.isAnswerCorrect(question, userAnswer) {
			earnedPoints += question.Points
		}
	}

	if totalPoints == 0 {
		return &dto.ValidateLessonResponse{
			Score:       100,
			Passed:      true,
			TotalPoints: 0,
			MinScore:    lesson.MinScore,
		}, nil
	}

	score := (earnedPoints * 100) / totalPoints
	passed := score >= lesson.MinScore

	return &dto.ValidateLessonResponse{
		Score:       score,
		Passed:      passed,
		TotalPoints: totalPoints,
		MinScore:    lesson.MinScore,
	}, nil
}

func (svc *ContentService) isAnswerCorrect(question model.Question, userAnswer interface{}) bool {
	switch question.Type {
	case "multiple_choice":
		// Convert both to strings for comparison
		correctAnswer, ok1 := question.Answer.(string)
		userAnswerStr, ok2 := userAnswer.(string)
		if ok1 && ok2 {
			return strings.EqualFold(strings.TrimSpace(correctAnswer), strings.TrimSpace(userAnswerStr))
		}
		// Fallback to direct comparison
		return question.Answer == userAnswer
	case "fill_blank":
		// Case-insensitive string comparison
		correctAnswer, ok1 := question.Answer.(string)
		userAnswerStr, ok2 := userAnswer.(string)
		if ok1 && ok2 {
			return strings.EqualFold(strings.TrimSpace(correctAnswer), strings.TrimSpace(userAnswerStr))
		}
	case "drag_drop", "connect":
		// For array-based answers, compare as JSON
		correctJSON, _ := json.Marshal(question.Answer)
		userJSON, _ := json.Marshal(userAnswer)
		return string(correctJSON) == string(userJSON)
	}

	return false
}

func (svc *ContentService) GetEras() ([]string, error) {
	return []string{"Bac_Thuoc", "Doc_Lap", "Phong_Kien", "Can_Dai"}, nil
}

func (svc *ContentService) GetDynasties() ([]string, error) {
	return []string{"Văn Lang", "Âu Lạc", "Bắc Thuộc", "Ngô", "Cận Đại", "Đinh - Tiền Lê", "Lý", "Trần", "Hồ", "Nguyễn", "Minh Chiếm Đóng", "Hậu Lê", "Mạc", "Tây Sơn"}, nil
}

// ==================== INDIVIDUAL QUESTION ANSWER METHODS ====================

func (svc *ContentService) SubmitQuestionAnswer(userID, lessonID, questionID string, answer interface{}) (*dto.SubmitQuestionAnswerResponse, error) {
	// Get the lesson to validate the question
	lesson, err := svc.sqlSvc.GetLesson(lessonID)
	if err != nil {
		return nil, err
	}

	var questions []model.Question
	if err := json.Unmarshal(lesson.Questions, &questions); err != nil {
		return nil, fmt.Errorf("failed to parse lesson questions: %v", err)
	}

	// Find the specific question
	var targetQuestion *model.Question
	totalPoints := 0
	for _, q := range questions {
		totalPoints += q.Points
		if q.ID == questionID {
			targetQuestion = &q
		}
	}

	if targetQuestion == nil {
		return nil, fmt.Errorf("question not found: %s", questionID)
	}

	// Check if answer is correct
	isCorrect := svc.isAnswerCorrect(*targetQuestion, answer)
	points := 0
	if isCorrect {
		points = targetQuestion.Points
	}

	// Convert answer to JSON string for storage
	answerJSON, err := json.Marshal(answer)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal answer: %v", err)
	}

	// Save the answer
	userAnswer := &model.UserQuestionAnswer{
		UserID:     userID,
		LessonID:   lessonID,
		QuestionID: questionID,
		Answer:     string(answerJSON),
		IsCorrect:  isCorrect,
		Points:     points,
	}

	if err := svc.sqlSvc.SaveUserQuestionAnswer(userAnswer); err != nil {
		return nil, err
	}

	// Get updated lesson status after this answer
	status, err := svc.CheckLessonStatus(userID, lessonID)
	if err != nil {
		return nil, err
	}

	return &dto.SubmitQuestionAnswerResponse{
		Correct:      isCorrect,
		Points:       points,
		TotalPoints:  totalPoints,
		EarnedPoints: status.EarnedPoints,
		CurrentScore: status.Score,
		Passed:       status.Passed,
		CanStillPass: status.CanStillPass,
		PointsNeeded: status.PointsNeeded,
	}, nil
}

func (svc *ContentService) CheckLessonStatus(userID, lessonID string) (*dto.CheckLessonStatusResponse, error) {
	// Get the lesson
	lesson, err := svc.sqlSvc.GetLesson(lessonID)
	if err != nil {
		return nil, err
	}

	var questions []model.Question
	if err := json.Unmarshal(lesson.Questions, &questions); err != nil {
		return nil, fmt.Errorf("failed to parse lesson questions: %v", err)
	}

	// Get user's answers for this lesson
	userAnswers, err := svc.sqlSvc.GetUserQuestionAnswers(userID, lessonID)
	if err != nil {
		return nil, err
	}

	// Calculate totals
	totalPoints := 0
	earnedPoints := 0
	questionsAnswered := len(userAnswers)

	// Create map of answered questions for quick lookup
	answeredQuestions := make(map[string]bool)
	for _, answer := range userAnswers {
		answeredQuestions[answer.QuestionID] = true
		if answer.IsCorrect {
			earnedPoints += answer.Points
		}
	}

	// Calculate remaining possible points
	remainingPoints := 0
	for _, question := range questions {
		totalPoints += question.Points
		if !answeredQuestions[question.ID] {
			remainingPoints += question.Points
		}
	}

	// Calculate current score
	score := 0
	if totalPoints > 0 {
		score = (earnedPoints * 100) / totalPoints
	}

	// Calculate minimum points needed to pass
	minPointsToPass := (lesson.MinScore*totalPoints + 99) / 100 // Round up

	// Check if user has already passed
	passed := earnedPoints >= minPointsToPass

	// Check if user can still pass (has enough remaining points)
	canStillPass := (earnedPoints + remainingPoints) >= minPointsToPass

	return &dto.CheckLessonStatusResponse{
		Score:             score,
		Passed:            passed,
		TotalPoints:       totalPoints,
		EarnedPoints:      earnedPoints,
		MinScore:          lesson.MinScore,
		QuestionsTotal:    len(questions),
		QuestionsAnswered: questionsAnswered,
		CanStillPass:      canStillPass,
		PointsNeeded:      maxInt(0, minPointsToPass-earnedPoints),
		RemainingPoints:   remainingPoints,
	}, nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
