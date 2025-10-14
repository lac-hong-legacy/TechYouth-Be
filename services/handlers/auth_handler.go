package handlers

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/lac-hong-legacy/ven_api/dto"
	"github.com/lac-hong-legacy/ven_api/shared"
)

type AuthHandler struct {
	authSvc AuthServiceInterface
	jwtSvc  JWTServiceInterface
	userSvc UserServiceInterface
}

func NewAuthHandler(authSvc AuthServiceInterface, jwtSvc JWTServiceInterface, userSvc UserServiceInterface) *AuthHandler {
	return &AuthHandler{
		authSvc: authSvc,
		jwtSvc:  jwtSvc,
		userSvc: userSvc,
	}
}

// @Summary Register a new user
// @Description Create a new user account with email verification and password confirmation
// @Tags auth
// @Accept json
// @Produce json
// @Param registerRequest body dto.RegisterRequest true "Registration details with password confirmation"
// @Success 201 {object} shared.Response{data=dto.RegisterResponse}
// @Router /api/v1/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return err
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	resp, err := h.authSvc.Register(req)
	if err != nil {
		return err
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
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return err
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	clientIP := c.IP()
	userAgent := c.Get("User-Agent")

	resp, err := h.authSvc.Login(req, clientIP, userAgent)
	if err != nil {
		return err
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
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var req dto.RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return err
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	clientIP := c.IP()
	userAgent := c.Get("User-Agent")

	resp, err := h.authSvc.RefreshToken(req, clientIP, userAgent)
	if err != nil {
		return err
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
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)
	sessionID := c.Locals("session_id").(string)
	clientIP := c.IP()
	userAgent := c.Get("User-Agent")

	authHeader := c.Get("Authorization")
	accessToken, _ := h.jwtSvc.ExtractTokenFromHeader(authHeader)

	err := h.authSvc.Logout(userID, sessionID, accessToken, clientIP, userAgent)
	if err != nil {
		return err
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
func (h *AuthHandler) LogoutAll(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)
	sessionID := c.Locals("session_id").(string)
	clientIP := c.IP()
	userAgent := c.Get("User-Agent")

	authHeader := c.Get("Authorization")
	accessToken, _ := h.jwtSvc.ExtractTokenFromHeader(authHeader)

	err := h.authSvc.LogoutAllDevices(userID, sessionID, accessToken, clientIP, userAgent)
	if err != nil {
		return err
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
func (h *AuthHandler) VerifyEmail(c *fiber.Ctx) error {
	var req dto.VerifyEmailRequest
	if err := c.BodyParser(&req); err != nil {
		return shared.NewBadRequestError(err, "Invalid request body")
	}

	if err := req.Validate(); err != nil {
		return shared.NewBadRequestError(err, "Validation failed")
	}

	err := h.authSvc.VerifyEmail(req.Email, req.Code)
	if err != nil {
		return err
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
func (h *AuthHandler) ResendVerification(c *fiber.Ctx) error {
	var req dto.ResendVerificationRequest
	if err := c.BodyParser(&req); err != nil {
		return shared.NewBadRequestError(err, "Invalid request")
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	err := h.authSvc.ResendVerificationEmail(req.Email)
	if err != nil {
		return err
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
func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var req dto.ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return shared.NewBadRequestError(err, "Invalid request")
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	err := h.authSvc.ForgotPassword(req.Email)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, http.StatusOK, "Password reset email sent successfully", nil)
}

// @Summary Reset password
// @Description Reset user password using reset token
// @Tags auth
// @Accept json
// @Produce json
// @Param resetRequest body dto.ResetPasswordRequest true "Reset token and new password"
// @Success 200 {object} shared.Response{data=nil}
// @Router /api/v1/reset-password [post]
func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var req dto.ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return shared.NewBadRequestError(err, "Invalid request")
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	err := h.authSvc.ResetPassword(req)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, http.StatusOK, "Password reset successfully", nil)
}

// @Summary Change password
// @Description Change password for authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Param changeRequest body dto.ChangePasswordRequest true "Current and new password"
// @Success 200 {object} shared.Response{data=nil}
// @Router /api/v1/change-password [post]
func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	userID := c.Locals(shared.UserID).(string)

	var req dto.ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return shared.NewBadRequestError(err, "Invalid request")
	}

	if err := req.Validate(); err != nil {
		validationResp := dto.CreateValidationErrorResponse(err)
		return c.Status(fiber.StatusBadRequest).JSON(validationResp)
	}

	err := h.authSvc.ChangePassword(userID, req)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, http.StatusOK, "Password changed successfully", nil)
}

// @Summary Check username availability
// @Description Check if username is available for registration
// @Tags auth
// @Accept json
// @Produce json
// @Param username path string true "Username to check"
// @Success 200 {object} shared.Response{data=map[string]interface{}}
// @Router /api/v1/username/check/{username} [get]
func (h *AuthHandler) CheckUsernameAvailability(c *fiber.Ctx) error {
	username := c.Params("username")

	available, err := h.userSvc.CheckUsernameAvailability(username)
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
