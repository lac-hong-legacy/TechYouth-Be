package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lac-hong-legacy/ven_api/dto"
	"github.com/lac-hong-legacy/ven_api/shared"
)

type GuestHandler struct {
	guestSvc   GuestServiceInterface
	contentSvc ContentServiceInterface
}

func NewGuestHandler(guestSvc GuestServiceInterface, contentSvc ContentServiceInterface) *GuestHandler {
	return &GuestHandler{
		guestSvc:   guestSvc,
		contentSvc: contentSvc,
	}
}

// @Summary Create or Get Guest Session
// @Description This endpoint creates a new guest session or retrieves an existing one based on device ID
// @Tags guest
// @Accept  json
// @Produce json
// @Param createSessionRequest body dto.CreateSessionRequest true "Create session request"
// @Success 200
// @Router /api/v1/guest/session [post]
func (h *GuestHandler) CreateSession(c *fiber.Ctx) error {
	var req dto.CreateSessionRequest
	if err := c.BodyParser(&req); err != nil {
		return shared.NewBadRequestError(err, "Invalid request")
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	session, err := h.guestSvc.CreateOrGetSession(req.DeviceID)
	if err != nil {
		return err
	}

	progress, err := h.contentSvc.GetProgress(session.ID)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", dto.CreateSessionResponse{
		Session:  session,
		Progress: progress,
	})
}

// @Summary Get Guest Progress
// @Description This endpoint retrieves the progress of a guest session
// @Tags guest
// @Accept  json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200
// @Router /api/v1/guest/session/{sessionId}/progress [get]
func (h *GuestHandler) GetProgress(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	progress, err := h.contentSvc.GetProgress(sessionID)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", progress)
}

// @Summary Check Lesson Access
// @Description This endpoint checks if a guest session can access a specific lesson
// @Tags guest
// @Accept  json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param lessonId path string true "Lesson ID"
// @Success 200 {object} shared.Response{data=dto.LessonAccessResponse}
// @Router /api/v1/guest/session/{sessionId}/lesson/{lessonId}/access [get]
func (h *GuestHandler) CheckLessonAccess(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")
	lessonID := c.Params("lessonId")

	canAccess, reason, err := h.guestSvc.CanAccessLesson(sessionID, lessonID)
	if err != nil {
		return err
	}

	res := dto.LessonAccessResponse{
		CanAccess: canAccess,
		Reason:    reason,
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", res)
}

// @Summary Complete Lesson
// @Description This endpoint marks a lesson as completed for a guest session
// @Tags guest
// @Accept  json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param completeLessonRequest body dto.CompleteLessonRequest true "Complete lesson request"
// @Success 200
// @Router /api/v1/guest/session/{sessionId}/lesson/complete [post]
func (h *GuestHandler) CompleteLesson(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	var req dto.CompleteLessonRequest
	if err := c.BodyParser(&req); err != nil {
		return shared.NewBadRequestError(err, "Invalid request")
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	err := h.guestSvc.CompleteLesson(sessionID, req.LessonID, req.Score, req.TimeSpent)
	if err != nil {
		return err
	}

	progress, err := h.contentSvc.GetProgress(sessionID)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", progress)
}

// @Summary Add Hearts from Ad
// @Description This endpoint adds hearts to a guest session when an ad is watched
// @Tags guest
// @Accept  json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200
// @Router /api/v1/guest/session/{sessionId}/hearts/add [post]
func (h *GuestHandler) AddHeartsFromAd(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	err := h.guestSvc.AddHeartsFromAd(sessionID)
	if err != nil {
		return err
	}

	progress, err := h.contentSvc.GetProgress(sessionID)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", progress)
}

// @Summary Lose Heart
// @Description This endpoint deducts a heart from a guest session when a lesson is failed
// @Tags guest
// @Accept  json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200
// @Router /api/v1/guest/session/{sessionId}/hearts/lose [post]
func (h *GuestHandler) LoseHeart(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	err := h.guestSvc.LoseHeart(sessionID)
	if err != nil {
		return err
	}

	progress, err := h.contentSvc.GetProgress(sessionID)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", progress)
}
