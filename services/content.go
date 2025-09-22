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

	eras := make([]dto.TimelineEraResponse, len(timelines))

	for i, timeline := range timelines {
		// Get characters for this timeline/era
		var characterIDs []string
		if timeline.CharacterIds != nil {
			if err := json.Unmarshal(timeline.CharacterIds, &characterIDs); err != nil {
				log.Printf("Failed to unmarshal character IDs for timeline %s: %v", timeline.Era, err)
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

		// Create era response (treating each timeline as an era)
		eras[i] = dto.TimelineEraResponse{
			Era: timeline.Era,
			Dynasties: []dto.TimelineDynastyResponse{
				{
					Dynasty:    timeline.Era, // Use era name as dynasty for now
					StartYear:  timeline.StartYear,
					EndYear:    timeline.EndYear,
					Characters: characterResponses,
					IsUnlocked: timeline.IsUnlocked,
					Progress:   svc.calculateDynastyProgress(characters),
				},
			},
			IsUnlocked: timeline.IsUnlocked,
			Progress:   svc.calculateDynastyProgress(characters),
		}
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
