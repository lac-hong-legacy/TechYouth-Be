package services

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/alphabatem/common/context"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	docs "github.com/lac-hong-legacy/ven_api/docs"
	"github.com/lac-hong-legacy/ven_api/dto"
	"github.com/lac-hong-legacy/ven_api/model"

	"github.com/lac-hong-legacy/ven_api/shared"
)

type HttpService struct {
	context.DefaultService

	jwtSvc        *JWTService
	authSvc       *AuthService
	guestSvc      *GuestService
	contentSvc    *ContentService
	userSvc       *UserService
	mediaSvc      *MediaService
	sqliteSvc     *PostgresService
	monitoringSvc *MonitoringService

	internalPassword string

	port int
	app  *fiber.App
}

const HTTP_SVC = "http_svc"

func (svc HttpService) Id() string {
	return HTTP_SVC
}

func (svc *HttpService) Configure(ctx *context.Context) error {
	if port := os.Getenv("HTTP_PORT"); port != "" {
		var err error
		if svc.port, err = strconv.Atoi(port); err != nil {
			return err
		}
	} else {
		svc.port = 8000
	}

	svc.internalPassword = os.Getenv("INTERNAL_PASSWORD")

	return svc.DefaultService.Configure(ctx)
}

func (svc *HttpService) Start() error {
	svc.jwtSvc = svc.Service(JWT_SVC).(*JWTService)
	svc.authSvc = svc.Service(AUTH_SVC).(*AuthService)
	svc.guestSvc = svc.Service(GUEST_SVC).(*GuestService)
	svc.userSvc = svc.Service(USER_SVC).(*UserService)
	svc.contentSvc = svc.Service(CONTENT_SVC).(*ContentService)
	svc.mediaSvc = svc.Service(MEDIA_SVC).(*MediaService)
	svc.sqliteSvc = svc.Service(POSTGRES_SVC).(*PostgresService)
	svc.monitoringSvc = svc.Service(MONITORING_SVC).(*MonitoringService)

	// Create Fiber app
	config := fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return svc.HandleError(c, err)
		},
	}

	svc.app = fiber.New(config)
	docs.SwaggerInfo.BasePath = ""

	// Add middleware
	svc.app.Use(recover.New())

	// svc.app.Use(MonitoringMiddleware(svc.monitoringSvc))

	if os.Getenv("LOG_LEVEL") == "TRACE" {
		svc.app.Use(logger.New())
	}

	svc.app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowCredentials: false,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
	}))

	//Validation endpoints
	svc.app.Get("/ping", svc.ping)
	svc.app.Get("/swagger/*", swagger.HandlerDefault)

	v1 := svc.app.Group("/api/v1")

	v1.Post("/register", svc.Register)
	v1.Post("/login", svc.Login)
	v1.Post("/refresh", svc.RefreshToken)
	v1.Post("/logout", svc.authSvc.RequiredAuth(), svc.Logout)
	v1.Post("/logout-all", svc.authSvc.RequiredAuth(), svc.LogoutAll)
	v1.Post("/verify-email", svc.VerifyEmail)
	v1.Post("/resend-verification", svc.ResendVerification)
	v1.Post("/forgot-password", svc.ForgotPassword)
	v1.Post("/reset-password", svc.ResetPassword)
	v1.Post("/change-password", svc.authSvc.RequiredAuth(), svc.ChangePassword)
	v1.Get("/username/check/:username", svc.CheckUsernameAvailability)

	guest := v1.Group("/guest")
	guest.Post("/session", svc.CreateSession)
	guest.Get("/session/:sessionId/progress", svc.GetProgress)
	guest.Get("/session/:sessionId/lesson/:lessonId/access", svc.CheckLessonAccess)
	guest.Post("/session/:sessionId/lesson/complete", svc.CompleteLesson)
	guest.Post("/session/:sessionId/hearts/add", svc.AddHeartsFromAd)
	guest.Post("/session/:sessionId/hearts/lose", svc.LoseHeart)

	content := v1.Group("/content")
	content.Get("/timeline", svc.GetTimeline)
	content.Get("/characters", svc.GetCharacters)
	content.Get("/characters/:characterId", svc.GetCharacter)
	content.Get("/characters/:characterId/lessons", svc.GetCharacterLessons)
	content.Get("/lessons/:lessonId", svc.GetLesson)
	content.Post("/lessons/validate", svc.ValidateLessonAnswers)
	content.Post("/lessons/questions/answer", svc.authSvc.RequiredAuth(), svc.SubmitQuestionAnswer)
	content.Post("/lessons/status", svc.authSvc.RequiredAuth(), svc.CheckLessonStatus)
	content.Get("/search", svc.SearchContent)
	content.Get("/eras", svc.GetEras)
	content.Get("/dynasties", svc.GetDynasties)

	user := v1.Group("/user", svc.authSvc.RequiredAuth())
	// Profile management
	user.Get("/profile", svc.GetUserProfile)
	user.Put("/profile", svc.UpdateUserProfile)
	user.Post("/initialize", svc.InitializeUserProfile)

	// Progress and game state
	user.Get("/progress", svc.GetUserProgress)
	user.Get("/collection", svc.GetUserCollection)

	// Lesson management
	user.Get("/lesson/:lessonId/access", svc.CheckUserLessonAccess)
	user.Post("/lesson/complete", svc.CompleteUserLesson)

	// Hearts management
	user.Get("/hearts", svc.GetHeartStatus)
	user.Post("/hearts/add", svc.AddUserHearts)
	user.Post("/hearts/lose", svc.LoseUserHeart)

	// Session management
	user.Get("/sessions", svc.GetSessions)
	user.Delete("/sessions/:sessionId", svc.RevokeSession)

	// Security management
	user.Get("/security", svc.GetSecuritySettings)
	user.Put("/security", svc.UpdateSecuritySettings)

	// Audit logs
	user.Get("/audit-logs", svc.GetAuditLogs)

	// Trusted devices
	user.Get("/devices", svc.GetUserDevices)
	user.Put("/devices/:deviceId/trust", svc.UpdateDeviceTrust)
	user.Delete("/devices/:deviceId", svc.RemoveUserDevice)

	// Social features
	user.Post("/share", svc.ShareAchievement)

	leaderboard := v1.Group("/leaderboard")
	leaderboard.Get("/weekly", svc.GetWeeklyLeaderboard)
	leaderboard.Get("/monthly", svc.GetMonthlyLeaderboard)
	leaderboard.Get("/all-time", svc.GetAllTimeLeaderboard)

	// Admin endpoints
	admin := v1.Group("/admin", svc.authSvc.RequireRole("admin"))
	// Character & Lesson Management
	admin.Post("/characters", svc.CreateCharacter)
	// admin.Post("/lessons", svc.CreateLesson)
	admin.Post("/lessons/new", svc.CreateLessonFromRequest)

	// Production Workflow (Admin Only) - Professional 3-step process
	admin.Put("/lessons/:lessonId/script", svc.UpdateLessonScript)
	admin.Post("/lessons/:lessonId/audio", svc.UploadLessonAudio)
	admin.Post("/lessons/:lessonId/animation", svc.UploadLessonAnimation)
	admin.Get("/lessons/:lessonId/production-status", svc.GetLessonProductionStatus)

	// Media Management (Admin Only)
	admin.Post("/lessons/:lessonId/subtitle", svc.UploadLessonSubtitle)
	admin.Post("/lessons/:lessonId/thumbnail", svc.UploadThumbnail)
	admin.Get("/lessons/:lessonId/media", svc.GetLessonMedia)
	admin.Delete("/media/assets/:assetId", svc.DeleteMediaAsset)
	admin.Get("/media/statistics", svc.GetMediaStatistics)
	admin.Get("/users", svc.AdminGetUsers)
	admin.Put("/users/:userId", svc.AdminUpdateUser)
	admin.Delete("/users/:userId", svc.AdminDeleteUser)

	// 404 handler
	svc.app.Use(func(c *fiber.Ctx) error {
		return svc.HandleError(c, errors.New("page not found"))
	})

	return svc.app.Listen(fmt.Sprintf(":%v", svc.port))
}

func (svc *HttpService) Shutdown() {
	_ = svc.app.Shutdown()
}

// @Summary Ping
// @Description This endpoint checks the health of the service
// @Tags health
// @Accept  json
// @Produce json
// @Success 200 {object} shared.Response{data=string}
// @Router /ping [get]
func (svc *HttpService) ping(c *fiber.Ctx) error {
	return shared.ResponseJSON(c, fiber.StatusOK, "Success", "pong")
}

func (svc *HttpService) HandleError(c *fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}

	if appErr, ok := shared.GetAppError(err); ok {
		return shared.ResponseJSON(c, appErr.StatusCode, appErr.Message, appErr.Data)
	}

	return shared.ResponseInternalError(c, err)
}

// @Summary Register a new user
// @Description Create a new user account with email verification and password confirmation
// @Tags auth
// @Accept json
// @Produce json
// @Param registerRequest body dto.RegisterRequest true "Registration details with password confirmation"
// @Success 201 {object} shared.Response{data=dto.RegisterResponse}
// @Router /api/v1/register [post]
func (h *HttpService) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return h.HandleError(c, err)
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	resp, err := h.authSvc.Register(req)
	if err != nil {
		return h.HandleError(c, err)
	}

	return shared.ResponseJSON(c, http.StatusCreated, "User registered successfully", resp)
}

// @Summary Login user
// @Description Authenticate user and return access token
// @Tags auth
// @Accept json
// @Produce json
// @Param loginRequest body dto.LoginRequest true "Login credentials"
// @Success 200 {object} shared.Response{data=dto.LoginResponse}
// @Router /api/v1/login [post]
func (h *HttpService) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return h.HandleError(c, err)
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	clientIP := c.IP()
	userAgent := c.Get("User-Agent")

	resp, err := h.authSvc.Login(req, clientIP, userAgent)
	if err != nil {
		return h.HandleError(c, err)
	}

	return shared.ResponseJSON(c, http.StatusOK, "Login successful", resp)
}

// @Summary Refresh access token
// @Description Generate new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param refreshRequest body dto.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} shared.Response{data=dto.LoginResponse}
// @Router /api/v1/refresh [post]
func (h *HttpService) RefreshToken(c *fiber.Ctx) error {
	var req dto.RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return h.HandleError(c, err)
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	clientIP := c.IP()
	userAgent := c.Get("User-Agent")

	resp, err := h.authSvc.RefreshToken(req, clientIP, userAgent)
	if err != nil {
		return h.HandleError(c, err)
	}

	return shared.ResponseJSON(c, http.StatusOK, "Token refreshed successfully", resp)
}

// @Summary Logout user
// @Description Invalidate current session
// @Tags auth
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Success 200 {object} shared.Response{data=nil}
// @Router /api/v1/logout [post]
func (h *HttpService) Logout(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)
	sessionID := c.Locals("session_id").(string)
	clientIP := c.IP()
	userAgent := c.Get("User-Agent")

	authHeader := c.Get("Authorization")
	accessToken, _ := h.jwtSvc.ExtractTokenFromHeader(authHeader)

	err := h.authSvc.Logout(userID, sessionID, accessToken, clientIP, userAgent)
	if err != nil {
		return h.HandleError(c, err)
	}

	return shared.ResponseJSON(c, http.StatusOK, "Logged out successfully", nil)
}

// @Summary Logout from all devices
// @Description Invalidate all sessions for the user
// @Tags auth
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Success 200 {object} shared.Response{data=nil}
// @Router /api/v1/logout-all [post]
func (h *HttpService) LogoutAll(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)
	sessionID := c.Locals("session_id").(string)
	clientIP := c.IP()
	userAgent := c.Get("User-Agent")

	authHeader := c.Get("Authorization")
	accessToken, _ := h.jwtSvc.ExtractTokenFromHeader(authHeader)

	err := h.authSvc.LogoutAllDevices(userID, sessionID, accessToken, clientIP, userAgent)
	if err != nil {
		return h.HandleError(c, err)
	}

	return shared.ResponseJSON(c, http.StatusOK, "Logged out from all devices successfully", nil)
}

// @Summary Verify email
// @Description Verify user email with 6-digit verification code
// @Tags auth
// @Accept json
// @Produce json
// @Param body body dto.VerifyEmailRequest true "Verification code and email"
// @Success 200 {object} shared.Response{data=nil}
// @Router /api/v1/verify-email [post]
func (h *HttpService) VerifyEmail(c *fiber.Ctx) error {
	var req dto.VerifyEmailRequest
	if err := c.BodyParser(&req); err != nil {
		return h.HandleError(c, shared.NewBadRequestError(err, "Invalid request body"))
	}

	if err := req.Validate(); err != nil {
		return h.HandleError(c, shared.NewBadRequestError(err, "Validation failed"))
	}

	err := h.authSvc.VerifyEmail(req.Email, req.Code)
	if err != nil {
		return h.HandleError(c, err)
	}

	return shared.ResponseJSON(c, http.StatusOK, "Email verified successfully", nil)
}

// @Summary Resend verification email
// @Description Send a new verification email to user
// @Tags auth
// @Accept json
// @Produce json
// @Param resendRequest body dto.ResendVerificationRequest true "Email address"
// @Success 200 {object} shared.Response{data=nil}
// @Router /api/v1/resend-verification [post]
func (h *HttpService) ResendVerification(c *fiber.Ctx) error {
	var req dto.ResendVerificationRequest
	if err := c.BodyParser(&req); err != nil {
		return h.HandleError(c, shared.NewBadRequestError(err, "Invalid request"))
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	err := h.authSvc.ResendVerificationEmail(req.Email)
	if err != nil {
		return h.HandleError(c, err)
	}

	return shared.ResponseJSON(c, http.StatusOK, "Verification email sent successfully", nil)
}

// @Summary Request password reset
// @Description Send password reset email to user
// @Tags auth
// @Accept json
// @Produce json
// @Param forgotRequest body dto.ForgotPasswordRequest true "Email address"
// @Success 200 {object} shared.Response{data=nil}
// @Router /api/v1/forgot-password [post]
func (h *HttpService) ForgotPassword(c *fiber.Ctx) error {
	var req dto.ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return h.HandleError(c, shared.NewBadRequestError(err, "Invalid request"))
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	err := h.authSvc.ForgotPassword(req.Email)
	if err != nil {
		return h.HandleError(c, err)
	}

	return shared.ResponseJSON(c, http.StatusOK, "Password reset email sent successfully", nil)
}

// @Summary Reset password
// @Description Reset user password with reset code and password confirmation
// @Tags auth
// @Accept json
// @Produce json
// @Param resetRequest body dto.ResetPasswordRequest true "Reset code, new password, and password confirmation"
// @Success 200 {object} shared.Response{data=nil}
// @Router /api/v1/reset-password [post]
func (h *HttpService) ResetPassword(c *fiber.Ctx) error {
	var req dto.ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return h.HandleError(c, shared.NewBadRequestError(err, "Invalid request"))
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	err := h.authSvc.ResetPassword(req)
	if err != nil {
		return h.HandleError(c, err)
	}

	return shared.ResponseJSON(c, http.StatusOK, "Password reset successfully", nil)
}

// @Summary Change password
// @Description Change user password with confirmation (requires authentication)
// @Tags auth
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Param changeRequest body dto.ChangePasswordRequest true "Current password, new password, and confirmation"
// @Success 200 {object} shared.Response{data=nil}
// @Router /api/v1/change-password [post]
func (h *HttpService) ChangePassword(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)

	var req dto.ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return h.HandleError(c, shared.NewBadRequestError(err, "Invalid request"))
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	err := h.authSvc.ChangePassword(userID, req)
	if err != nil {
		return h.HandleError(c, err)
	}

	return shared.ResponseJSON(c, http.StatusOK, "Password changed successfully", nil)
}

// @Summary Check Username Availability
// @Description Check if a username is available for registration
// @Tags auth
// @Produce json
// @Param username path string true "Username to check"
// @Success 200 {object} shared.Response{data=map[string]interface{}}
// @Router /api/v1/username/check/{username} [get]
func (svc *HttpService) CheckUsernameAvailability(c *fiber.Ctx) error {
	username := c.Params("username")

	available, err := svc.userSvc.CheckUsernameAvailability(username)
	if err != nil {
		return shared.ResponseJSON(c, fiber.StatusBadRequest, "Invalid username", map[string]interface{}{
			"available": false,
			"error":     err.Error(),
		})
	}

	message := "Username is available"
	if !available {
		message = "Username is already taken"
	}

	return shared.ResponseJSON(c, fiber.StatusOK, message, map[string]interface{}{
		"available": available,
		"username":  username,
	})
}

// Admin endpoints

// @Summary Get all users (Admin)
// @Description Get list of all users (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param search query string false "Search term"
// @Success 200 {object} shared.Response{data=dto.AdminUserListResponse}
// @Router /api/v1/admin/users [get]
func (h *HttpService) AdminGetUsers(c *fiber.Ctx) error {
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
		return h.HandleError(c, err)
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Users retrieved successfully", users)
}

// @Summary Update user (Admin)
// @Description Update user information (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security Bearer
// @Param userId path string true "User ID"
// @Param updateRequest body dto.AdminUpdateUserRequest true "User update data"
// @Success 200 {object} shared.Response{data=dto.AdminUserInfo}
// @Router /api/v1/admin/users/{userId} [put]
func (h *HttpService) AdminUpdateUser(c *fiber.Ctx) error {
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
		return h.HandleError(c, err)
	}

	return shared.ResponseJSON(c, http.StatusOK, "User updated successfully", user)
}

// @Summary Delete user (Admin)
// @Description Soft delete user (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security Bearer
// @Param userId path string true "User ID"
// @Success 200 {object} shared.Response{data=nil}
// @Router /api/v1/admin/users/{userId} [delete]
func (h *HttpService) AdminDeleteUser(c *fiber.Ctx) error {
	userID := c.Params("userId")
	if userID == "" {
		return shared.ResponseJSON(c, http.StatusBadRequest, "User ID is required", nil)
	}

	err := h.userSvc.AdminDeleteUser(userID)
	if err != nil {
		return h.HandleError(c, err)
	}

	return shared.ResponseJSON(c, http.StatusOK, "User deleted successfully", nil)
}

// @Summary Get user sessions
// @Description Get list of active user sessions
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} shared.Response{data=dto.SessionListResponse}
// @Router /api/v1/user/sessions [get]
func (h *HttpService) GetSessions(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)
	currentSessionID := c.Locals("session_id").(string)

	sessions, err := h.userSvc.GetUserSessions(userID, currentSessionID)
	if err != nil {
		return h.HandleError(c, err)
	}

	return shared.ResponseJSON(c, http.StatusOK, "Sessions retrieved successfully", sessions)
}

// @Summary Revoke user session
// @Description Revoke a specific user session
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param sessionId path string true "Session ID"
// @Success 200 {object} shared.Response{data=nil}
// @Router /api/v1/user/sessions/{sessionId} [delete]
func (h *HttpService) RevokeSession(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)
	sessionID := c.Params("sessionId")

	if sessionID == "" {
		return shared.ResponseJSON(c, http.StatusBadRequest, "Session ID is required", nil)
	}

	err := h.userSvc.RevokeUserSession(userID, sessionID)
	if err != nil {
		return h.HandleError(c, err)
	}

	return shared.ResponseJSON(c, http.StatusOK, "Session revoked successfully", nil)
}

// @Summary Get security settings
// @Description Get user security settings
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} shared.Response{data=dto.SecuritySettings}
// @Router /api/v1/user/security [get]
func (h *HttpService) GetSecuritySettings(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)

	settings, err := h.userSvc.GetSecuritySettings(userID)
	if err != nil {
		return h.HandleError(c, err)
	}

	return shared.ResponseJSON(c, http.StatusOK, "Security settings retrieved successfully", settings)
}

// @Summary Update security settings
// @Description Update user security settings
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param updateRequest body dto.UpdateSecuritySettingsRequest true "Security settings"
// @Success 200 {object} shared.Response{data=dto.SecuritySettings}
// @Router /api/v1/user/security [put]
func (h *HttpService) UpdateSecuritySettings(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)

	var req dto.UpdateSecuritySettingsRequest
	if err := c.BodyParser(&req); err != nil {
		return h.HandleError(c, shared.NewBadRequestError(err, "Invalid request"))
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	settings, err := h.userSvc.UpdateSecuritySettings(userID, req)
	if err != nil {
		return h.HandleError(c, err)
	}

	return shared.ResponseJSON(c, http.StatusOK, "Security settings updated successfully", settings)
}

// @Summary Get audit logs
// @Description Get user authentication audit logs
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} shared.Response{data=dto.AuditLogResponse}
// @Router /api/v1/user/audit-logs [get]
func (h *HttpService) GetAuditLogs(c *fiber.Ctx) error {
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
		return h.HandleError(c, err)
	}

	return shared.ResponseJSON(c, http.StatusOK, "Audit logs retrieved successfully", logs)
}

// @Summary Get user devices
// @Description Get list of all user's trusted devices
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} shared.Response{data=[]dto.DeviceInfo}
// @Router /api/v1/user/devices [get]
func (h *HttpService) GetUserDevices(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)

	devices, err := h.authSvc.GetUserDevices(userID)
	if err != nil {
		return h.HandleError(c, err)
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
// @Param deviceId path string true "Device ID"
// @Param trustRequest body dto.TrustDeviceRequest true "Trust status"
// @Success 200 {object} shared.Response
// @Router /api/v1/user/devices/{deviceId}/trust [put]
func (h *HttpService) UpdateDeviceTrust(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)
	deviceID := c.Params("deviceId")

	var req dto.TrustDeviceRequest
	if err := c.BodyParser(&req); err != nil {
		return h.HandleError(c, shared.NewBadRequestError(err, "Invalid request"))
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	if err := h.authSvc.UpdateDeviceTrust(userID, deviceID, req.Trust); err != nil {
		return h.HandleError(c, err)
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
// @Param deviceId path string true "Device ID"
// @Success 200 {object} shared.Response
// @Router /api/v1/user/devices/{deviceId} [delete]
func (h *HttpService) RemoveUserDevice(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)
	deviceID := c.Params("deviceId")

	if err := h.authSvc.RemoveDevice(userID, deviceID); err != nil {
		return h.HandleError(c, err)
	}

	return shared.ResponseJSON(c, http.StatusOK, "Device removed successfully", nil)
}

// @Summary Create or Get Guest Session
// @Description This endpoint creates a new guest session or retrieves an existing one based on device ID
// @Tags guest
// @Accept  json
// @Produce json
// @Param createSessionRequest body dto.CreateSessionRequest true "Create session request"
// @Success 200
// @Router /api/v1/guest/session [post]
func (svc *HttpService) CreateSession(c *fiber.Ctx) error {
	var req dto.CreateSessionRequest
	if err := c.BodyParser(&req); err != nil {
		return svc.HandleError(c, shared.NewBadRequestError(err, "Invalid request"))
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	session, err := svc.guestSvc.CreateOrGetSession(req.DeviceID)
	if err != nil {
		return svc.HandleError(c, err)
	}

	progress, err := svc.sqliteSvc.GetProgress(session.ID)
	if err != nil {
		return svc.HandleError(c, err)
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
func (svc *HttpService) GetProgress(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	progress, err := svc.sqliteSvc.GetProgress(sessionID)
	if err != nil {
		return svc.HandleError(c, err)
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
func (svc *HttpService) CheckLessonAccess(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")
	lessonID := c.Params("lessonId")

	canAccess, reason, err := svc.guestSvc.CanAccessLesson(sessionID, lessonID)
	if err != nil {
		return svc.HandleError(c, err)
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
func (svc *HttpService) CompleteLesson(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	var req dto.CompleteLessonRequest
	if err := c.BodyParser(&req); err != nil {
		return svc.HandleError(c, shared.NewBadRequestError(err, "Invalid request"))
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	err := svc.guestSvc.CompleteLesson(sessionID, req.LessonID, req.Score, req.TimeSpent)
	if err != nil {
		return svc.HandleError(c, err)
	}

	progress, err := svc.sqliteSvc.GetProgress(sessionID)
	if err != nil {
		return svc.HandleError(c, err)
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
func (svc *HttpService) AddHeartsFromAd(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	err := svc.guestSvc.AddHeartsFromAd(sessionID)
	if err != nil {
		return svc.HandleError(c, err)
	}

	progress, err := svc.sqliteSvc.GetProgress(sessionID)
	if err != nil {
		return svc.HandleError(c, err)
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
func (svc *HttpService) LoseHeart(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	err := svc.guestSvc.LoseHeart(sessionID)
	if err != nil {
		return svc.HandleError(c, err)
	}

	progress, err := svc.sqliteSvc.GetProgress(sessionID)
	if err != nil {
		return svc.HandleError(c, err)
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", progress)
}

// ==================== CONTENT ENDPOINTS ====================

// @Summary Get Timeline
// @Description Get the historical timeline with eras and dynasties
// @Tags content
// @Accept json
// @Produce json
// @Success 200 {object} shared.Response{data=dto.TimelineCollectionResponse}
// @Router /api/v1/content/timeline [get]
func (svc *HttpService) GetTimeline(c *fiber.Ctx) error {
	timeline, err := svc.contentSvc.GetTimeline()
	if err != nil {
		return svc.HandleError(c, err)
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
func (svc *HttpService) GetCharacters(c *fiber.Ctx) error {
	dynasty := c.Query("dynasty")
	rarity := c.Query("rarity")

	characters, err := svc.contentSvc.GetCharacters(dynasty, rarity)
	if err != nil {
		return svc.HandleError(c, err)
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
func (svc *HttpService) GetCharacter(c *fiber.Ctx) error {
	characterID := c.Params("characterId")

	character, err := svc.contentSvc.GetCharacterDetails(characterID)
	if err != nil {
		return svc.HandleError(c, err)
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
func (svc *HttpService) GetCharacterLessons(c *fiber.Ctx) error {
	characterID := c.Params("characterId")

	lessons, err := svc.contentSvc.GetCharacterLessons(characterID)
	if err != nil {
		return svc.HandleError(c, err)
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
func (svc *HttpService) GetLesson(c *fiber.Ctx) error {
	lessonID := c.Params("lessonId")

	lesson, err := svc.contentSvc.GetLessonContent(lessonID)
	if err != nil {
		return svc.HandleError(c, err)
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
func (svc *HttpService) ValidateLessonAnswers(c *fiber.Ctx) error {
	var req dto.ValidateLessonRequest
	if err := c.BodyParser(&req); err != nil {
		return svc.HandleError(c, shared.NewBadRequestError(err, "Invalid request"))
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	result, err := svc.contentSvc.ValidateLessonAnswers(req.LessonID, req.UserAnswers)
	if err != nil {
		return svc.HandleError(c, err)
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
func (svc *HttpService) SearchContent(c *fiber.Ctx) error {
	var req dto.SearchRequest
	if err := c.QueryParser(&req); err != nil {
		return svc.HandleError(c, shared.NewBadRequestError(err, "Invalid query parameters"))
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	if req.Limit == 0 {
		req.Limit = 20
	}

	results, err := svc.contentSvc.SearchContent(req)
	if err != nil {
		return svc.HandleError(c, err)
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
func (svc *HttpService) SubmitQuestionAnswer(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req dto.SubmitQuestionAnswerRequest
	if err := c.BodyParser(&req); err != nil {
		return svc.HandleError(c, shared.NewBadRequestError(err, "Invalid request"))
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	result, err := svc.contentSvc.SubmitQuestionAnswer(userID, req.LessonID, req.QuestionID, req.Answer)
	if err != nil {
		return svc.HandleError(c, err)
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
func (svc *HttpService) CheckLessonStatus(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req dto.CheckLessonStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return svc.HandleError(c, shared.NewBadRequestError(err, "Invalid request"))
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	result, err := svc.contentSvc.CheckLessonStatus(userID, req.LessonID)
	if err != nil {
		return svc.HandleError(c, err)
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Lesson status retrieved", result)
}

// ==================== USER PROFILE ENDPOINTS ====================

// @Summary Get User Profile
// @Description Get current user's profile information
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Success 200 {object} shared.Response{data=dto.UserProfileResponse}
// @Router /api/v1/user/profile [get]
func (svc *HttpService) GetUserProfile(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)

	profile, err := svc.userSvc.GetUserProfile(userID)
	if err != nil {
		return svc.HandleError(c, err)
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", profile)
}

// @Summary Update User Profile
// @Description Update user profile information
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Param updateProfileRequest body dto.UpdateProfileRequest true "Update profile request"
// @Success 200 {object} shared.Response{data=dto.UserProfileResponse}
// @Router /api/v1/user/profile [put]
func (svc *HttpService) UpdateUserProfile(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)

	var req dto.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return svc.HandleError(c, shared.NewBadRequestError(err, "Invalid request"))
	}

	profile, err := svc.userSvc.UpdateUserProfile(userID, req)
	if err != nil {
		return svc.HandleError(c, err)
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", profile)
}

// @Summary Initialize User Profile
// @Description Initialize user profile after first login (zodiac setup)
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Param birthYear body map[string]int true "Birth year for zodiac"
// @Success 200 {object} shared.Response{data=dto.UserProgressResponse}
// @Router /api/v1/user/initialize [post]
func (svc *HttpService) InitializeUserProfile(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)

	var req map[string]int
	if err := c.BodyParser(&req); err != nil {
		return svc.HandleError(c, shared.NewBadRequestError(err, "Invalid request"))
	}

	birthYear, exists := req["birth_year"]
	if !exists || birthYear < 1900 || birthYear > 2020 {
		return svc.HandleError(c, shared.NewBadRequestError(nil, "Valid birth year is required"))
	}

	err := svc.userSvc.InitializeUserProfile(userID, birthYear)
	if err != nil {
		return svc.HandleError(c, err)
	}

	progress, err := svc.userSvc.GetUserProgress(userID)
	if err != nil {
		return svc.HandleError(c, err)
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", progress)
}

// ==================== USER PROGRESS ENDPOINTS ====================

// @Summary Get User Progress
// @Description Get current user's game progress
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Success 200 {object} shared.Response{data=dto.UserProgressResponse}
// @Router /api/v1/user/progress [get]
func (svc *HttpService) GetUserProgress(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)

	progress, err := svc.userSvc.GetUserProgress(userID)
	if err != nil {
		return svc.HandleError(c, err)
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", progress)
}

// @Summary Get User Collection
// @Description Get user's character collection and achievements
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param UserID query string false "User ID"
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Success 200 {object} shared.Response{data=dto.CollectionResponse}
// @Router /api/v1/user/collection [get]
func (svc *HttpService) GetUserCollection(c *fiber.Ctx) error {
	userID := c.Query("userId")
	if userID == "" {
		userID = c.Locals(shared.UserID).(string)
	}

	collection, err := svc.userSvc.GetUserCollection(userID)
	if err != nil {
		return svc.HandleError(c, err)
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", collection)
}

// ==================== USER LESSON ENDPOINTS ====================

// @Summary Check User Lesson Access
// @Description Check if user can access a specific lesson
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param lessonId path string true "Lesson ID"
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Success 200 {object} shared.Response{data=dto.LessonAccessResponse}
// @Router /api/v1/user/lesson/{lessonId}/access [get]
func (svc *HttpService) CheckUserLessonAccess(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)
	lessonID := c.Params("lessonId")

	access, err := svc.userSvc.CheckLessonAccess(userID, lessonID)
	if err != nil {
		return svc.HandleError(c, err)
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", access)
}

// @Summary Complete User Lesson
// @Description Complete a lesson for registered user
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param completeLessonRequest body dto.CompleteLessonRequest true "Complete lesson request"
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Success 200 {object} shared.Response{data=dto.CompleteLessonResponse}
// @Router /api/v1/user/lesson/complete [post]
func (svc *HttpService) CompleteUserLesson(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)

	var req dto.CompleteLessonRequest
	if err := c.BodyParser(&req); err != nil {
		return svc.HandleError(c, shared.NewBadRequestError(err, "Invalid request"))
	}

	err := svc.userSvc.CompleteLesson(userID, req.LessonID, req.Score, req.TimeSpent)
	if err != nil {
		return svc.HandleError(c, err)
	}

	result, err := svc.userSvc.GetUserProgress(userID)
	if err != nil {
		return svc.HandleError(c, err)
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", result)
}

// ==================== HEARTS MANAGEMENT ENDPOINTS ====================

// @Summary Get Heart Status
// @Description Get current heart status and reset information
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Success 200 {object} shared.Response{data=dto.HeartStatusResponse}
// @Router /api/v1/user/hearts [get]
func (svc *HttpService) GetHeartStatus(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)

	status, err := svc.userSvc.GetHeartStatus(userID)
	if err != nil {
		return svc.HandleError(c, err)
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", status)
}

// @Summary Add User Hearts
// @Description Add hearts from ads or other sources
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Param addHeartsRequest body dto.AddHeartsRequest true "Add hearts request"
// @Success 200 {object} shared.Response{data=dto.HeartStatusResponse}
// @Router /api/v1/user/hearts/add [post]
func (svc *HttpService) AddUserHearts(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)

	var req dto.AddHeartsRequest
	if err := c.BodyParser(&req); err != nil {
		return svc.HandleError(c, shared.NewBadRequestError(err, "Invalid request"))
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	status, err := svc.userSvc.AddHearts(userID, req.Source, req.Amount)
	if err != nil {
		return svc.HandleError(c, err)
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", status)
}

// @Summary Lose User Heart
// @Description Deduct a heart when user fails a lesson
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Success 200 {object} shared.Response{data=dto.HeartStatusResponse}
// @Router /api/v1/user/hearts/lose [post]
func (svc *HttpService) LoseUserHeart(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)

	status, err := svc.userSvc.LoseHeart(userID)
	if err != nil {
		return svc.HandleError(c, err)
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", status)
}

// ==================== LEADERBOARD ENDPOINTS ====================

// @Summary Get Weekly Leaderboard
// @Description Get weekly leaderboard rankings
// @Tags leaderboard
// @Accept json
// @Produce json
// @Param limit query int false "Limit results (default 50)"
// @Success 200 {object} shared.Response{data=dto.LeaderboardResponse}
// @Router /api/v1/leaderboard/weekly [get]
func (svc *HttpService) GetWeeklyLeaderboard(c *fiber.Ctx) error {
	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	var userID string
	if authHeader := c.Get("Authorization"); authHeader != "" {
		if token, err := svc.jwtSvc.ExtractTokenFromHeader(authHeader); err == nil {
			if uid, err := svc.jwtSvc.VerifyJWTToken(token); err == nil {
				userID = uid
			}
		}
	}

	leaderboard, err := svc.userSvc.GetWeeklyLeaderboard(limit, userID)
	if err != nil {
		return svc.HandleError(c, err)
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", leaderboard)
}

// @Summary Get Monthly Leaderboard
// @Description Get monthly leaderboard rankings
// @Tags leaderboard
// @Accept json
// @Produce json
// @Param limit query int false "Limit results (default 50)"
// @Success 200 {object} shared.Response{data=dto.LeaderboardResponse}
// @Router /api/v1/leaderboard/monthly [get]
func (svc *HttpService) GetMonthlyLeaderboard(c *fiber.Ctx) error {
	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	var userID string
	if authHeader := c.Get("Authorization"); authHeader != "" {
		if token, err := svc.jwtSvc.ExtractTokenFromHeader(authHeader); err == nil {
			if uid, err := svc.jwtSvc.VerifyJWTToken(token); err == nil {
				userID = uid
			}
		}
	}

	leaderboard, err := svc.userSvc.GetMonthlyLeaderboard(limit, userID)
	if err != nil {
		return svc.HandleError(c, err)
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", leaderboard)
}

// @Summary Get All Time Leaderboard
// @Description Get all-time leaderboard rankings
// @Tags leaderboard
// @Accept json
// @Produce json
// @Param limit query int false "Limit results (default 50)"
// @Success 200 {object} shared.Response{data=dto.LeaderboardResponse}
// @Router /api/v1/leaderboard/all-time [get]
func (svc *HttpService) GetAllTimeLeaderboard(c *fiber.Ctx) error {
	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	var userID string
	if authHeader := c.Get("Authorization"); authHeader != "" {
		if token, err := svc.jwtSvc.ExtractTokenFromHeader(authHeader); err == nil {
			if uid, err := svc.jwtSvc.VerifyJWTToken(token); err == nil {
				userID = uid
			}
		}
	}

	leaderboard, err := svc.userSvc.GetAllTimeLeaderboard(limit, userID)
	if err != nil {
		return svc.HandleError(c, err)
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", leaderboard)
}

// ==================== SOCIAL ENDPOINTS ====================

// @Summary Share Achievement
// @Description Share user achievement or progress on social media
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Param shareRequest body dto.ShareRequest true "Share request"
// @Success 200 {object} shared.Response{data=dto.ShareResponse}
// @Router /api/v1/user/share [post]
func (svc *HttpService) ShareAchievement(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)

	var req dto.ShareRequest
	if err := c.BodyParser(&req); err != nil {
		return svc.HandleError(c, shared.NewBadRequestError(err, "Invalid request"))
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	shareData, err := svc.userSvc.CreateShareContent(userID, req)
	if err != nil {
		return svc.HandleError(c, err)
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", shareData)
}

// ==================== MEDIA ENDPOINTS ====================

// @Summary Upload Lesson Subtitle (Admin)
// @Description Upload subtitle file for lesson video (Admin only)
// @Tags admin
// @Accept multipart/form-data
// @Produce json
// @Security Bearer
// @Param Authorization header string true "Admin Bearer Token" default(Bearer <admin_token>)
// @Param lessonId path string true "Lesson ID"
// @Param subtitle formData file true "Subtitle file (VTT, SRT)"
// @Success 200 {object} shared.Response{data=dto.MediaUploadResponse}
// @Router /api/v1/admin/lessons/{lessonId}/subtitle [post]
func (svc *HttpService) UploadLessonSubtitle(c *fiber.Ctx) error {
	lessonID := c.Params("lessonId")

	file, err := c.FormFile("subtitle")
	if err != nil {
		return svc.HandleError(c, shared.NewBadRequestError(err, "No subtitle file provided"))
	}

	response, err := svc.mediaSvc.UploadLessonSubtitle(lessonID, file)
	if err != nil {
		return svc.HandleError(c, err)
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Subtitle uploaded successfully", response)
}

// @Summary Upload Lesson Thumbnail (Admin)
// @Description Upload thumbnail image for lesson (Admin only)
// @Tags admin
// @Accept multipart/form-data
// @Produce json
// @Security Bearer
// @Param Authorization header string true "Admin Bearer Token" default(Bearer <admin_token>)
// @Param lessonId path string true "Lesson ID"
// @Param thumbnail formData file true "Thumbnail file (JPG, PNG, WEBP)"
// @Success 200 {object} shared.Response{data=dto.MediaUploadResponse}
// @Router /api/v1/admin/lessons/{lessonId}/thumbnail [post]
func (svc *HttpService) UploadThumbnail(c *fiber.Ctx) error {
	lessonID := c.Params("lessonId")

	file, err := c.FormFile("thumbnail")
	if err != nil {
		return svc.HandleError(c, shared.NewBadRequestError(err, "No thumbnail file provided"))
	}

	response, err := svc.mediaSvc.UploadThumbnail(lessonID, file)
	if err != nil {
		return svc.HandleError(c, err)
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Thumbnail uploaded successfully", response)
}

// @Summary Get Lesson Media (Admin)
// @Description Get all media assets for a lesson (Admin only)
// @Tags admin
// @Produce json
// @Security Bearer
// @Param Authorization header string true "Admin Bearer Token" default(Bearer <admin_token>)
// @Param lessonId path string true "Lesson ID"
// @Success 200 {object} shared.Response{data=dto.LessonMediaResponse}
// @Router /api/v1/admin/lessons/{lessonId}/media [get]
func (svc *HttpService) GetLessonMedia(c *fiber.Ctx) error {
	lessonID := c.Params("lessonId")

	media, err := svc.mediaSvc.GetLessonMedia(lessonID)
	if err != nil {
		return svc.HandleError(c, err)
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", media)
}

// @Summary Delete Media Asset (Admin)
// @Description Delete a media asset and its physical file (Admin only)
// @Tags admin
// @Produce json
// @Security Bearer
// @Param Authorization header string true "Admin Bearer Token" default(Bearer <admin_token>)
// @Param assetId path string true "Media Asset ID"
// @Success 200 {object} shared.Response{data=string}
// @Router /api/v1/admin/media/assets/{assetId} [delete]
func (svc *HttpService) DeleteMediaAsset(c *fiber.Ctx) error {
	assetID := c.Params("assetId")

	err := svc.mediaSvc.DeleteMediaAsset(assetID)
	if err != nil {
		return svc.HandleError(c, err)
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Media asset deleted successfully", "deleted")
}

// @Summary Get Media Statistics (Admin)
// @Description Get statistics about media assets and usage (Admin only)
// @Tags admin
// @Produce json
// @Security Bearer
// @Param Authorization header string true "Admin Bearer Token" default(Bearer <admin_token>)
// @Success 200 {object} shared.Response{data=map[string]interface{}}
// @Router /api/v1/admin/media/statistics [get]
func (svc *HttpService) GetMediaStatistics(c *fiber.Ctx) error {
	stats, err := svc.sqliteSvc.GetMediaStatistics()
	if err != nil {
		return svc.HandleError(c, err)
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", stats)
}

// ==================== PRODUCTION WORKFLOW ENDPOINTS ====================

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
func (svc *HttpService) UpdateLessonScript(c *fiber.Ctx) error {
	lessonID := c.Params("lessonId")

	var req dto.UpdateLessonScriptRequest
	if err := c.BodyParser(&req); err != nil {
		return svc.HandleError(c, shared.NewBadRequestError(err, "Invalid request"))
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	lesson, err := svc.contentSvc.UpdateLessonScript(lessonID, req.Script)
	if err != nil {
		return svc.HandleError(c, err)
	}

	response := svc.contentSvc.MapLessonToResponse(lesson)
	return shared.ResponseJSON(c, fiber.StatusOK, "Script finalized successfully", response)
}

// @Summary Upload Lesson Audio (Admin)
// @Description Upload voice-over audio file - Step 2 of production workflow (Admin only)
// @Tags admin,production
// @Accept multipart/form-data
// @Produce json
// @Security Bearer
// @Param Authorization header string true "Admin Bearer Token" default(Bearer <admin_token>)
// @Param lessonId path string true "Lesson ID"
// @Param audio formData file true "Audio file (MP3, WAV, AAC)"
// @Success 200 {object} shared.Response{data=dto.MediaUploadResponse}
// @Router /api/v1/admin/lessons/{lessonId}/audio [post]
func (svc *HttpService) UploadLessonAudio(c *fiber.Ctx) error {
	lessonID := c.Params("lessonId")

	file, err := c.FormFile("audio")
	if err != nil {
		return svc.HandleError(c, shared.NewBadRequestError(err, "No audio file provided"))
	}

	response, err := svc.mediaSvc.UploadLessonAudio(lessonID, file)
	if err != nil {
		return svc.HandleError(c, err)
	}

	if err := svc.contentSvc.MarkAudioUploaded(lessonID); err != nil {
		return svc.HandleError(c, err)
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Audio uploaded successfully", response)
}

// @Summary Upload Lesson Animation (Admin)
// @Description Upload animation video file - Step 3 of production workflow (Admin only)
// @Tags admin,production
// @Accept multipart/form-data
// @Produce json
// @Security Bearer
// @Param Authorization header string true "Admin Bearer Token" default(Bearer <admin_token>)
// @Param lessonId path string true "Lesson ID"
// @Param animation formData file true "Animation file (MP4, MOV, WEBM)"
// @Success 200 {object} shared.Response{data=dto.MediaUploadResponse}
// @Router /api/v1/admin/lessons/{lessonId}/animation [post]
func (svc *HttpService) UploadLessonAnimation(c *fiber.Ctx) error {
	lessonID := c.Params("lessonId")

	file, err := c.FormFile("animation")
	if err != nil {
		return svc.HandleError(c, shared.NewBadRequestError(err, "No animation file provided"))
	}

	response, err := svc.mediaSvc.UploadLessonAnimation(lessonID, file)
	if err != nil {
		return svc.HandleError(c, err)
	}

	if err := svc.contentSvc.MarkAnimationUploaded(lessonID); err != nil {
		return svc.HandleError(c, err)
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Animation uploaded successfully", response)
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
func (svc *HttpService) GetLessonProductionStatus(c *fiber.Ctx) error {
	lessonID := c.Params("lessonId")

	status, err := svc.contentSvc.GetLessonProductionStatus(lessonID)
	if err != nil {
		return svc.HandleError(c, err)
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", status)
}

// ==================== ADMIN ENDPOINTS (Optional) ====================

// @Summary Create Character (Admin)
// @Description Create a new historical character (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Param X-Internal-Password header string true "Admin internal password"
// @Param character body model.Character true "Character data"
// @Success 201 {object} shared.Response{data=dto.CharacterResponse}
// @Router /api/v1/admin/characters [post]
func (svc *HttpService) CreateCharacter(c *fiber.Ctx) error {
	isAdmin := svc.isAdmin(c)
	if !isAdmin {
		return nil
	}

	var character model.Character
	if err := c.BodyParser(&character); err != nil {
		return svc.HandleError(c, shared.NewBadRequestError(err, "Invalid character data"))
	}

	created, err := svc.contentSvc.CreateCharacter(&character)
	if err != nil {
		return svc.HandleError(c, err)
	}

	return shared.ResponseJSON(c, fiber.StatusCreated, "Character created successfully", created)
}

// // @Summary Create Lesson (Admin)
// // @Description Create a new lesson (admin only)
// // @Tags admin
// // @Accept json
// // @Produce json
// // @Security Bearer
// // @Param X-Internal-Password header string true "Admin internal password"
// // @Param lesson body model.Lesson true "Lesson data"
// // @Success 201 {object} shared.Response{data=dto.LessonResponse}
// // @Router /api/v1/admin/lessons [post]
// func (svc *HttpService) CreateLesson(c *fiber.Ctx) error {
// 	isAdmin := svc.isAdmin(c)
// 	if !isAdmin {
// 		return nil
// 	}

// 	var lesson model.Lesson
// 	if err := c.BodyParser(&lesson); err != nil {
// 		return svc.HandleError(c, shared.NewBadRequestError(err, "Invalid lesson data"))
// 	}

// 	created, err := svc.contentSvc.CreateLesson(&lesson)
// 	if err != nil {
// 		return svc.HandleError(c, err)
// 	}

// 	return shared.ResponseJSON(c, fiber.StatusCreated, "Lesson created successfully", created)
// }

// @Summary Create Lesson from Request (Admin)
// @Description Create a new lesson using structured request (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "Admin Bearer Token" default(Bearer <admin_token>)
// @Param lessonRequest body dto.CreateLessonRequest true "Lesson creation request"
// @Success 201 {object} shared.Response{data=dto.LessonResponse}
// @Router /api/v1/admin/lessons/new [post]
func (svc *HttpService) CreateLessonFromRequest(c *fiber.Ctx) error {
	var req dto.CreateLessonRequest
	if err := c.BodyParser(&req); err != nil {
		return svc.HandleError(c, shared.NewBadRequestError(err, "Invalid lesson request"))
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	created, err := svc.contentSvc.CreateLessonFromRequest(req)
	if err != nil {
		return svc.HandleError(c, err)
	}

	return shared.ResponseJSON(c, fiber.StatusCreated, "Lesson created successfully", created)
}

func (svc *HttpService) isAdmin(c *fiber.Ctx) bool {
	if c.Get("X-Internal-Password") != svc.internalPassword {
		svc.HandleError(c, shared.NewUnauthorizedError(nil, "Admin access required"))
		return false
	}
	return true
}

// @Summary Get Eras
// @Description Get all eras
// @Tags content
// @Accept json
// @Produce json
// @Success 200 {object} shared.Response{data=[]string}
// @Router /api/v1/content/eras [get]
func (svc *HttpService) GetEras(c *fiber.Ctx) error {
	eras, err := svc.contentSvc.GetEras()
	if err != nil {
		return svc.HandleError(c, err)
	}
	return shared.ResponseJSON(c, fiber.StatusOK, "Success", eras)
}

// @Summary Get Dynasties
// @Description Get all dynasties
// @Tags content
// @Accept json
// @Produce json
// @Success 200 {object} shared.Response{data=[]string}
// @Router /api/v1/content/dynasties [get]
func (svc *HttpService) GetDynasties(c *fiber.Ctx) error {
	dynasties, err := svc.contentSvc.GetDynasties()
	if err != nil {
		return svc.HandleError(c, err)
	}
	return shared.ResponseJSON(c, fiber.StatusOK, "Success", dynasties)
}
