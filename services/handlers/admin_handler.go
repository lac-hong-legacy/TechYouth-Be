package handlers

import (
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/lac-hong-legacy/ven_api/dto"
	"github.com/lac-hong-legacy/ven_api/model"
	"github.com/lac-hong-legacy/ven_api/shared"
)

type AdminHandler struct {
	userSvc    UserServiceInterface
	contentSvc ContentServiceInterface
}

func NewAdminHandler(userSvc UserServiceInterface, contentSvc ContentServiceInterface) *AdminHandler {
	return &AdminHandler{
		userSvc:    userSvc,
		contentSvc: contentSvc,
	}
}

// @Summary Get all users (Admin)
// @Description Get list of all users (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "Admin Bearer Token" default(Bearer <admin_token>)
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param search query string false "Search term"
// @Success 200 {object} shared.Response{data=dto.AdminUserListResponse}
// @Router /api/v1/admin/users [get]
func (h *AdminHandler) AdminGetUsers(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	users, err := h.userSvc.AdminGetUsers(page, limit, search)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Users retrieved successfully", users)
}

// @Summary Update user (Admin)
// @Description Update user information (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "Admin Bearer Token" default(Bearer <admin_token>)
// @Param userId path string true "User ID"
// @Param updateRequest body dto.AdminUpdateUserRequest true "User update data"
// @Success 200 {object} shared.Response{data=dto.AdminUserInfo}
// @Router /api/v1/admin/users/{userId} [put]
func (h *AdminHandler) AdminUpdateUser(c *fiber.Ctx) error {
	userID := c.Params("userId")
	if userID == "" {
		return shared.ResponseJSON(c, http.StatusBadRequest, "User ID is required", nil)
	}

	var req dto.AdminUpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return shared.ResponseJSON(c, http.StatusBadRequest, "Invalid request", err.Error())
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	user, err := h.userSvc.AdminUpdateUser(userID, req)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, http.StatusOK, "User updated successfully", user)
}

// @Summary Delete user (Admin)
// @Description Soft delete user (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "Admin Bearer Token" default(Bearer <admin_token>)
// @Param userId path string true "User ID"
// @Success 200 {object} shared.Response{data=nil}
// @Router /api/v1/admin/users/{userId} [delete]
func (h *AdminHandler) AdminDeleteUser(c *fiber.Ctx) error {
	userID := c.Params("userId")
	if userID == "" {
		return shared.ResponseJSON(c, http.StatusBadRequest, "User ID is required", nil)
	}

	err := h.userSvc.AdminDeleteUser(userID)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, http.StatusOK, "User deleted successfully", nil)
}

// @Summary Create Character (Admin)
// @Description Create a new historical character (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Param Authorization header string true "Admin Bearer Token" default(Bearer <admin_token>)
// @Param character body model.Character true "Character data"
// @Success 201 {object} shared.Response{data=dto.CharacterResponse}
// @Router /api/v1/admin/characters [post]
func (h *AdminHandler) CreateCharacter(c *fiber.Ctx) error {
	var character model.Character
	if err := c.BodyParser(&character); err != nil {
		return shared.NewBadRequestError(err, "Invalid character data")
	}

	created, err := h.contentSvc.CreateCharacter(&character)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusCreated, "Character created successfully", created)
}

// @Summary Create Lesson from Request (Admin)
// @Description Create a new lesson from request (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Param Authorization header string true "Admin Bearer Token" default(Bearer <admin_token>)
// @Param createLessonRequest body dto.CreateLessonRequest true "Lesson creation request"
// @Success 201 {object} shared.Response{data=dto.LessonResponse}
// @Router /api/v1/admin/lessons/new [post]
func (h *AdminHandler) CreateLessonFromRequest(c *fiber.Ctx) error {
	var req dto.CreateLessonRequest
	if err := c.BodyParser(&req); err != nil {
		return shared.NewBadRequestError(err, "Invalid lesson request")
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	created, err := h.contentSvc.CreateLessonFromRequest(req)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusCreated, "Lesson created successfully", created)
}

// @Summary Update Lesson Script (Admin)
// @Description Finalize the lesson script - Step 1 of production workflow (Admin only)
// @Tags admin,production
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "Admin Bearer Token" default(Bearer <admin_token>)
// @Param lessonId path string true "Lesson ID"
// @Param scriptRequest body dto.UpdateLessonScriptRequest true "Script content"
// @Success 200 {object} shared.Response{data=dto.LessonResponse}
// @Router /api/v1/admin/lessons/{lessonId}/script [put]
func (h *AdminHandler) UpdateLessonScript(c *fiber.Ctx) error {
	lessonID := c.Params("lessonId")

	var req dto.UpdateLessonScriptRequest
	if err := c.BodyParser(&req); err != nil {
		return shared.NewBadRequestError(err, "Invalid request")
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	lesson, err := h.contentSvc.UpdateLessonScript(lessonID, req.Script)
	if err != nil {
		return err
	}

	response := h.contentSvc.MapLessonToResponse(lesson)
	return shared.ResponseJSON(c, fiber.StatusOK, "Script finalized successfully", &response)
}

// @Summary Get Lesson Production Status (Admin)
// @Description Get current status of lesson production workflow (Admin only)
// @Tags admin,production
// @Produce json
// @Security Bearer
// @Param Authorization header string true "Admin Bearer Token" default(Bearer <admin_token>)
// @Param lessonId path string true "Lesson ID"
// @Success 200 {object} shared.Response{data=dto.LessonProductionStatusResponse}
// @Router /api/v1/admin/lessons/{lessonId}/production-status [get]
func (h *AdminHandler) GetLessonProductionStatus(c *fiber.Ctx) error {
	lessonID := c.Params("lessonId")

	status, err := h.contentSvc.GetLessonProductionStatus(lessonID)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", status)
}
