package handlers

import (
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/lac-hong-legacy/ven_api/dto"
	"github.com/lac-hong-legacy/ven_api/shared"
)

type UserHandler struct {
	userSvc UserServiceInterface
	authSvc AuthServiceInterface
}

func NewUserHandler(userSvc UserServiceInterface, authSvc AuthServiceInterface) *UserHandler {
	return &UserHandler{
		userSvc: userSvc,
		authSvc: authSvc,
	}
}

// @Summary Get user profile
// @Description Get user profile
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Success 200 {object} shared.Response{data=dto.UserProfileResponse}
// @Router /api/v1/user/profile [get]
func (h *UserHandler) GetUserProfile(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)

	profile, err := h.userSvc.GetUserProfile(userID)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", profile)
}

// @Summary Update user profile
// @Description Update user profile
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Param updateRequest body dto.UpdateProfileRequest true "User profile"
// @Success 200 {object} shared.Response{data=dto.UserProfileResponse}
// @Router /api/v1/user/profile [put]
func (h *UserHandler) UpdateUserProfile(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)

	var req dto.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return shared.NewBadRequestError(err, "Invalid request")
	}

	profile, err := h.userSvc.UpdateUserProfile(userID, req)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", profile)
}

// @Summary Initialize user profile
// @Description Initialize user profile
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Param initializeRequest body map[string]int true "User profile"
// @Success 200 {object} shared.Response{data=dto.UserProgressResponse}
// @Router /api/v1/user/initialize [post]
func (h *UserHandler) InitializeUserProfile(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)

	var req map[string]int
	if err := c.BodyParser(&req); err != nil {
		return shared.NewBadRequestError(err, "Invalid request")
	}

	birthYear, exists := req["birth_year"]
	if !exists || birthYear < 1900 || birthYear > 2020 {
		return shared.NewBadRequestError(nil, "Valid birth year is required")
	}

	err := h.userSvc.InitializeUserProfile(userID, birthYear)
	if err != nil {
		return err
	}

	progress, err := h.userSvc.GetUserProgress(userID)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", progress)
}

// @Summary Get user progress
// @Description Get user progress
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Success 200 {object} shared.Response{data=dto.UserProgressResponse}
// @Router /api/v1/user/progress [get]
func (h *UserHandler) GetUserProgress(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)

	progress, err := h.userSvc.GetUserProgress(userID)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", progress)
}

// @Summary Get user collection
// @Description Get user collection
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Success 200 {object} shared.Response{data=dto.CollectionResponse}
// @Router /api/v1/user/collection [get]
func (h *UserHandler) GetUserCollection(c *fiber.Ctx) error {
	userID := c.Query("userId")
	if userID == "" {
		userID = c.Locals(shared.UserID).(string)
	}

	collection, err := h.userSvc.GetUserCollection(userID)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", collection)
}

// @Summary Check user lesson access
// @Description Check user lesson access
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Param lessonId path string true "Lesson ID"
// @Success 200 {object} shared.Response{data=dto.LessonAccessResponse}
// @Router /api/v1/user/lesson/{lessonId}/access [get]
func (h *UserHandler) CheckUserLessonAccess(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)
	lessonID := c.Params("lessonId")

	access, err := h.userSvc.CheckLessonAccess(userID, lessonID)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", access)
}

// @Summary Complete user lesson
// @Description Complete user lesson
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Param completeRequest body dto.CompleteLessonRequest true "Complete lesson request"
// @Success 200 {object} shared.Response{data=dto.UserProgressResponse}
// @Router /api/v1/user/lesson/complete [post]
func (h *UserHandler) CompleteUserLesson(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)

	var req dto.CompleteLessonRequest
	if err := c.BodyParser(&req); err != nil {
		return shared.NewBadRequestError(err, "Invalid request")
	}

	err := h.userSvc.CompleteLesson(userID, req.LessonID, req.Score, req.TimeSpent)
	if err != nil {
		return err
	}

	result, err := h.userSvc.GetUserProgress(userID)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", result)
}

// @Summary Get user heart status
// @Description Get user heart status
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Success 200 {object} shared.Response{data=dto.HeartStatusResponse}
// @Router /api/v1/user/hearts [get]
func (h *UserHandler) GetHeartStatus(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)

	status, err := h.userSvc.GetHeartStatus(userID)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", status)
}

// @Summary Add user hearts
// @Description Add user hearts
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Param addRequest body dto.AddHeartsRequest true "Add hearts request"
// @Success 200 {object} shared.Response{data=dto.HeartStatusResponse}
// @Router /api/v1/user/hearts/add [post]
func (h *UserHandler) AddUserHearts(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)

	var req dto.AddHeartsRequest
	if err := c.BodyParser(&req); err != nil {
		return shared.NewBadRequestError(err, "Invalid request")
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	status, err := h.userSvc.AddHearts(userID, req.Source, req.Amount)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", status)
}

// @Summary Lose user heart
// @Description Lose user heart
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Success 200 {object} shared.Response{data=dto.HeartStatusResponse}
// @Router /api/v1/user/hearts/lose [post]
func (h *UserHandler) LoseUserHeart(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)

	status, err := h.userSvc.LoseHeart(userID)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", status)
}

// @Summary Get user sessions
// @Description Get list of active user sessions
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Success 200 {object} shared.Response{data=dto.SessionListResponse}
// @Router /api/v1/user/sessions [get]
func (h *UserHandler) GetSessions(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)
	currentSessionID := c.Locals("session_id").(string)

	sessions, err := h.userSvc.GetUserSessions(userID, currentSessionID)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, http.StatusOK, "Sessions retrieved successfully", sessions)
}

// @Summary Revoke user session
// @Description Revoke a specific user session
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Param sessionId path string true "Session ID"
// @Success 200 {object} shared.Response{data=nil}
// @Router /api/v1/user/sessions/{sessionId} [delete]
func (h *UserHandler) RevokeSession(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)
	sessionID := c.Params("sessionId")

	if sessionID == "" {
		return shared.ResponseJSON(c, http.StatusBadRequest, "Session ID is required", nil)
	}

	err := h.userSvc.RevokeUserSession(userID, sessionID)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, http.StatusOK, "Session revoked successfully", nil)
}

// @Summary Get security settings
// @Description Get user security settings
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Success 200 {object} shared.Response{data=dto.SecuritySettings}
// @Router /api/v1/user/security [get]
func (h *UserHandler) GetSecuritySettings(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)

	settings, err := h.userSvc.GetSecuritySettings(userID)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, http.StatusOK, "Security settings retrieved successfully", settings)
}

// @Summary Update security settings
// @Description Update user security settings
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Param updateRequest body dto.UpdateSecuritySettingsRequest true "Security settings"
// @Success 200 {object} shared.Response{data=dto.SecuritySettings}
// @Router /api/v1/user/security [put]
func (h *UserHandler) UpdateSecuritySettings(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)

	var req dto.UpdateSecuritySettingsRequest
	if err := c.BodyParser(&req); err != nil {
		return shared.NewBadRequestError(err, "Invalid request")
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	settings, err := h.userSvc.UpdateSecuritySettings(userID, req)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, http.StatusOK, "Security settings updated successfully", settings)
}

// @Summary Get audit logs
// @Description Get user authentication audit logs
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} shared.Response{data=dto.AuditLogResponse}
// @Router /api/v1/user/audit-logs [get]
func (h *UserHandler) GetAuditLogs(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	logs, err := h.userSvc.GetUserAuditLogs(userID, page, limit)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, http.StatusOK, "Audit logs retrieved successfully", logs)
}

// @Summary Get user devices
// @Description Get list of all user's trusted devices
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Success 200 {object} shared.Response{data=[]dto.DeviceInfo}
// @Router /api/v1/user/devices [get]
func (h *UserHandler) GetUserDevices(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)

	devices, err := h.authSvc.GetUserDevices(userID)
	if err != nil {
		return err
	}

	response := dto.DeviceListResponse{
		Devices: devices,
		Total:   len(devices),
	}

	return shared.ResponseJSON(c, http.StatusOK, "Devices retrieved successfully", response)
}

// @Summary Update device trust status
// @Description Trust or untrust a specific device
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Param deviceId path string true "Device ID"
// @Param trustRequest body dto.TrustDeviceRequest true "Trust status"
// @Success 200 {object} shared.Response
// @Router /api/v1/user/devices/{deviceId}/trust [put]
func (h *UserHandler) UpdateDeviceTrust(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)
	deviceID := c.Params("deviceId")

	var req dto.TrustDeviceRequest
	if err := c.BodyParser(&req); err != nil {
		return shared.NewBadRequestError(err, "Invalid request")
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	if err := h.authSvc.UpdateDeviceTrust(userID, deviceID, req.Trust); err != nil {
		return err
	}

	message := "Device untrusted successfully"
	if req.Trust {
		message = "Device trusted successfully"
	}

	return shared.ResponseJSON(c, http.StatusOK, message, nil)
}

// @Summary Remove user device
// @Description Remove a device from user's device list
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Param deviceId path string true "Device ID"
// @Success 200 {object} shared.Response
// @Router /api/v1/user/devices/{deviceId} [delete]
func (h *UserHandler) RemoveUserDevice(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)
	deviceID := c.Params("deviceId")

	if err := h.authSvc.RemoveDevice(userID, deviceID); err != nil {
		return err
	}

	return shared.ResponseJSON(c, http.StatusOK, "Device removed successfully", nil)
}

// @Summary Share achievement
// @Description Share achievement
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Param shareRequest body dto.ShareRequest true "Share request"
// @Success 200 {object} shared.Response{data=dto.ShareResponse}
// @Router /api/v1/user/share [post]
func (h *UserHandler) ShareAchievement(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)

	var req dto.ShareRequest
	if err := c.BodyParser(&req); err != nil {
		return shared.NewBadRequestError(err, "Invalid request")
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	shareData, err := h.userSvc.CreateShareContent(userID, req)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", shareData)
}
