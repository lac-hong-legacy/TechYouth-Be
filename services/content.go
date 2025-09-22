// services/content.go
package services

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alphabatem/common/context"
	"github.com/lac-hong-legacy/TechYouth-Be/dto"
	"github.com/lac-hong-legacy/TechYouth-Be/model"
	log "github.com/sirupsen/logrus"
)

type ContentService struct {
	context.DefaultService
	sqlSvc *SqliteService
}

const CONTENT_SVC = "content_svc"

func (svc ContentService) Id() string {
	return CONTENT_SVC
}

func (svc *ContentService) Configure(ctx *context.Context) error {
	return svc.DefaultService.Configure(ctx)
}

func (svc *ContentService) Start() error {
	svc.sqlSvc = svc.Service(SQLITE_SVC).(*SqliteService)
	return nil
}

// ==================== TIMELINE METHODS ====================

func (svc *ContentService) GetTimeline() (*dto.TimelineCollectionResponse, error) {
	timelines, err := svc.sqlSvc.GetTimeline()
	if err != nil {
		return nil, err
	}

	// Group timelines by era
	eraMap := make(map[string]*dto.TimelineEraResponse)

	for _, timeline := range timelines {
		era, exists := eraMap[timeline.Era]
		if !exists {
			era = &dto.TimelineEraResponse{
				Era:        timeline.Era,
				Dynasties:  []dto.TimelineDynastyResponse{},
				IsUnlocked: timeline.IsUnlocked,
			}
			eraMap[timeline.Era] = era
		}

		// Get characters for this dynasty
		characters, err := svc.sqlSvc.GetCharactersByDynasty(timeline.Dynasty)
		if err != nil {
			log.Printf("Failed to get characters for dynasty %s: %v", timeline.Dynasty, err)
			characters = []model.Character{}
		}

		characterResponses := make([]dto.CharacterResponse, len(characters))
		for i, char := range characters {
			characterResponses[i] = svc.mapCharacterToResponse(&char)
		}

		dynasty := dto.TimelineDynastyResponse{
			Dynasty:    timeline.Dynasty,
			StartYear:  timeline.StartYear,
			EndYear:    timeline.EndYear,
			Characters: characterResponses,
			IsUnlocked: timeline.IsUnlocked,
			Progress:   svc.calculateDynastyProgress(characters),
		}

		era.Dynasties = append(era.Dynasties, dynasty)
	}

	// Convert map to slice
	eras := make([]dto.TimelineEraResponse, 0, len(eraMap))
	for _, era := range eraMap {
		eras = append(eras, *era)
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
		ID:           lesson.ID,
		CharacterID:  lesson.CharacterID,
		Title:        lesson.Title,
		Order:        lesson.Order,
		Story:        lesson.Story,
		VoiceOverURL: lesson.VoiceOverURL,
		Questions:    questions,
		XPReward:     lesson.XPReward,
		MinScore:     lesson.MinScore,
		Character:    svc.mapCharacterToResponse(&lesson.Character),
	}
}

// ==================== SEARCH METHODS ====================

func (svc *ContentService) SearchContent(req dto.SearchRequest) (*dto.SearchResponse, error) {
	characters, err := svc.sqlSvc.SearchCharacters(req.Query, req.Dynasty, req.Rarity, req.Limit)
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

// ==================== VALIDATION METHODS ====================

func (svc *ContentService) ValidateLessonAnswers(lessonID string, userAnswers map[string]interface{}) (int, error) {
	lesson, err := svc.sqlSvc.GetLesson(lessonID)
	if err != nil {
		return 0, err
	}

	var questions []model.Question
	if err := json.Unmarshal(lesson.Questions, &questions); err != nil {
		return 0, fmt.Errorf("failed to parse lesson questions: %v", err)
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
		return 100, nil // Default to 100% if no questions
	}

	score := (earnedPoints * 100) / totalPoints
	return score, nil
}

func (svc *ContentService) isAnswerCorrect(question model.Question, userAnswer interface{}) bool {
	switch question.Type {
	case "multiple_choice":
		return question.Answer == userAnswer
	case "fill_blank":
		// Case-insensitive string comparison
		correctAnswer, ok1 := question.Answer.(string)
		userAnswerStr, ok2 := userAnswer.(string)
		if ok1 && ok2 {
			return strings.ToLower(strings.TrimSpace(correctAnswer)) == strings.ToLower(strings.TrimSpace(userAnswerStr))
		}
	case "drag_drop", "connect":
		// For array-based answers, compare as JSON
		correctJSON, _ := json.Marshal(question.Answer)
		userJSON, _ := json.Marshal(userAnswer)
		return string(correctJSON) == string(userJSON)
	}

	return false
}
