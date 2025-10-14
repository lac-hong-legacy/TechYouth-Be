package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lac-hong-legacy/ven_api/dto"
	"github.com/lac-hong-legacy/ven_api/shared"
)

type ContentHandler struct {
	contentSvc ContentServiceInterface
}

func NewContentHandler(contentSvc ContentServiceInterface) *ContentHandler {
	return &ContentHandler{
		contentSvc: contentSvc,
	}
}

// @Summary Get Timeline
// @Description Get the historical timeline with eras and dynasties
// @Tags content
// @Accept json
// @Produce json
// @Success 200 {object} shared.Response{data=dto.TimelineCollectionResponse}
// @Router /api/v1/content/timeline [get]
func (h *ContentHandler) GetTimeline(c *fiber.Ctx) error {
	timeline, err := h.contentSvc.GetTimeline()
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", timeline)
}

// @Summary Get Characters
// @Description Get list of historical characters with filtering options
// @Tags content
// @Accept json
// @Produce json
// @Param dynasty query string false "Filter by dynasty"
// @Param rarity query string false "Filter by rarity"
// @Success 200 {object} shared.Response{data=dto.CharacterCollectionResponse}
// @Router /api/v1/content/characters [get]
func (h *ContentHandler) GetCharacters(c *fiber.Ctx) error {
	dynasty := c.Query("dynasty")
	rarity := c.Query("rarity")

	characters, err := h.contentSvc.GetCharacters(dynasty, rarity)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", characters)
}

// @Summary Get Character
// @Description Get detailed information about a specific character
// @Tags content
// @Accept json
// @Produce json
// @Param characterId path string true "Character ID"
// @Success 200 {object} shared.Response{data=dto.CharacterResponse}
// @Router /api/v1/content/characters/{characterId} [get]
func (h *ContentHandler) GetCharacter(c *fiber.Ctx) error {
	characterID := c.Params("characterId")

	character, err := h.contentSvc.GetCharacterDetails(characterID)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", character)
}

// @Summary Get Character Lessons
// @Description Get all lessons for a specific character
// @Tags content
// @Accept json
// @Produce json
// @Param characterId path string true "Character ID"
// @Success 200 {object} shared.Response{data=[]dto.LessonResponse}
// @Router /api/v1/content/characters/{characterId}/lessons [get]
func (h *ContentHandler) GetCharacterLessons(c *fiber.Ctx) error {
	characterID := c.Params("characterId")

	lessons, err := h.contentSvc.GetCharacterLessons(characterID)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", lessons)
}

// @Summary Get Lesson
// @Description Get detailed lesson content including questions
// @Tags content
// @Accept json
// @Produce json
// @Param lessonId path string true "Lesson ID"
// @Success 200 {object} shared.Response{data=dto.LessonResponse}
// @Router /api/v1/content/lessons/{lessonId} [get]
func (h *ContentHandler) GetLesson(c *fiber.Ctx) error {
	lessonID := c.Params("lessonId")

	lesson, err := h.contentSvc.GetLessonContent(lessonID)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", lesson)
}

// @Summary Validate Lesson Answers
// @Description Validate user answers for a lesson and return score
// @Tags content
// @Accept json
// @Produce json
// @Param validateRequest body dto.ValidateLessonRequest true "Validation request"
// @Success 200 {object} shared.Response{data=dto.ValidateLessonResponse}
// @Router /api/v1/content/lessons/validate [post]
func (h *ContentHandler) ValidateLessonAnswers(c *fiber.Ctx) error {
	var req dto.ValidateLessonRequest
	if err := c.BodyParser(&req); err != nil {
		return shared.NewBadRequestError(err, "Invalid request")
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	result, err := h.contentSvc.ValidateLessonAnswers(req.LessonID, req.UserAnswers)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Validation completed", result)
}

// @Summary Search Content
// @Description Search characters and content with various filters
// @Tags content
// @Accept json
// @Produce json
// @Param query query string false "Search query"
// @Param dynasty query string false "Filter by dynasty"
// @Param rarity query string false "Filter by rarity"
// @Param limit query int false "Limit results"
// @Success 200 {object} shared.Response{data=dto.SearchResponse}
// @Router /api/v1/content/search [get]
func (h *ContentHandler) SearchContent(c *fiber.Ctx) error {
	var req dto.SearchRequest
	if err := c.QueryParser(&req); err != nil {
		return shared.NewBadRequestError(err, "Invalid query parameters")
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	if req.Limit == 0 {
		req.Limit = 20
	}

	results, err := h.contentSvc.SearchContent(req)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", results)
}

// @Summary Submit Question Answer
// @Description Submit answer for individual question in a lesson
// @Tags content
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Param submitRequest body dto.SubmitQuestionAnswerRequest true "Question answer request"
// @Success 200 {object} shared.Response{data=dto.SubmitQuestionAnswerResponse}
// @Router /api/v1/content/lessons/questions/answer [post]
func (h *ContentHandler) SubmitQuestionAnswer(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req dto.SubmitQuestionAnswerRequest
	if err := c.BodyParser(&req); err != nil {
		return shared.NewBadRequestError(err, "Invalid request")
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	result, err := h.contentSvc.SubmitQuestionAnswer(userID, req.LessonID, req.QuestionID, req.Answer)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Answer submitted", result)
}

// @Summary Check Lesson Status
// @Description Check current lesson completion status and score
// @Tags content
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Param statusRequest body dto.CheckLessonStatusRequest true "Lesson status request"
// @Success 200 {object} shared.Response{data=dto.CheckLessonStatusResponse}
// @Router /api/v1/content/lessons/status [post]
func (h *ContentHandler) CheckLessonStatus(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req dto.CheckLessonStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return shared.NewBadRequestError(err, "Invalid request")
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	result, err := h.contentSvc.CheckLessonStatus(userID, req.LessonID)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Lesson status retrieved", result)
}

// @Summary Get Eras
// @Description Get list of eras
// @Tags content
// @Accept json
// @Produce json
// @Success 200 {object} shared.Response{data=[]string}
// @Router /api/v1/content/eras [get]
func (h *ContentHandler) GetEras(c *fiber.Ctx) error {
	eras, err := h.contentSvc.GetEras()
	if err != nil {
		return err
	}
	return shared.ResponseJSON(c, fiber.StatusOK, "Success", eras)
}

// @Summary Get Dynasties
// @Description Get list of dynasties
// @Tags content
// @Accept json
// @Produce json
// @Success 200 {object} shared.Response{data=[]string}
// @Router /api/v1/content/dynasties [get]
func (h *ContentHandler) GetDynasties(c *fiber.Ctx) error {
	dynasties, err := h.contentSvc.GetDynasties()
	if err != nil {
		return err
	}
	return shared.ResponseJSON(c, fiber.StatusOK, "Success", dynasties)
}
