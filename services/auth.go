package services

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lac-hong-legacy/TechYouth-Be/dto"
	"github.com/lac-hong-legacy/TechYouth-Be/model"
	"github.com/lac-hong-legacy/TechYouth-Be/shared"
	"golang.org/x/crypto/bcrypt"

	"github.com/alphabatem/common/context"
	log "github.com/sirupsen/logrus"
)

type AuthService struct {
	context.DefaultService

	sqlSvc       *SqliteService
	jwtSvc       *JWTService
	emailSvc     *EmailService
	rateLimitSvc *RateLimitService

	// Security settings
	maxLoginAttempts   int
	lockoutDuration    time.Duration
	passwordMinLength  int
	requireEmailVerify bool
}

const AUTH_SVC = "auth_svc"

func (svc AuthService) Id() string {
	return AUTH_SVC
}

func (svc *AuthService) Configure(ctx *context.Context) error {
	svc.maxLoginAttempts = 5
	svc.lockoutDuration = 30 * time.Minute
	svc.passwordMinLength = 8
	svc.requireEmailVerify = true

	return svc.DefaultService.Configure(ctx)
}

func (svc *AuthService) Start() error {
	svc.sqlSvc = svc.Service(SQLITE_SVC).(*SqliteService)
	svc.jwtSvc = svc.Service(JWT_SVC).(*JWTService)
	svc.emailSvc = svc.Service(EMAIL_SVC).(*EmailService)
	svc.rateLimitSvc = svc.Service(RATE_LIMIT_SVC).(*RateLimitService)
	return nil
}

// Password hashing
func (svc *AuthService) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

func (svc *AuthService) checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Password validation
func (svc *AuthService) validatePassword(password string) error {
	if len(password) < svc.passwordMinLength {
		return fmt.Errorf("password must be at least %d characters long", svc.passwordMinLength)
	}

	hasUpper := strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	hasLower := strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz")
	hasNumber := strings.ContainsAny(password, "0123456789")
	hasSpecial := strings.ContainsAny(password, "!@#$%^&*()_+-=[]{}|;:,.<>?")

	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return errors.New("password must contain uppercase, lowercase, number and special character")
	}

	return nil
}

// Generate secure random token
func (svc *AuthService) generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// Enhanced Registration
func (svc *AuthService) Register(registerRequest dto.RegisterRequest) (*dto.RegisterResponse, error) {
	// Validate password strength
	if err := svc.validatePassword(registerRequest.Password); err != nil {
		return nil, shared.NewBadRequestError(err, err.Error())
	}

	// Hash password
	hashedPassword, err := svc.hashPassword(registerRequest.Password)
	if err != nil {
		return nil, shared.NewInternalError(err, "Failed to hash password")
	}

	// Generate verification token
	verificationToken, err := svc.generateSecureToken()
	if err != nil {
		return nil, shared.NewInternalError(err, "Failed to generate verification token")
	}

	// Create user with hashed password
	registerRequest.Password = hashedPassword
	user, err := svc.sqlSvc.CreateUser(registerRequest, verificationToken)
	if err != nil {
		return nil, shared.NewInternalError(err, err.Error())
	}

	// Send verification email
	if svc.requireEmailVerify {
		err = svc.emailSvc.SendVerificationEmail(user.Email, user.Username, verificationToken)
		if err != nil {
			// Log error but don't fail registration
			log.WithError(err).Error("Failed to send verification email")
		}
	}

	// Log registration
	svc.logAuthEvent(user.ID, "register", "", "", true)

	return &dto.RegisterResponse{
		UserID:               user.ID,
		RequiresVerification: svc.requireEmailVerify,
	}, nil
}

// Enhanced Login
func (svc *AuthService) Login(loginRequest dto.LoginRequest, clientIP, userAgent string) (*dto.LoginResponse, error) {
	// Check rate limiting
	if blocked := svc.rateLimitSvc.IsBlocked(clientIP, "login"); blocked {
		return nil, shared.NewTooManyRequestsError(errors.New("too many login attempts"), "Too many login attempts. Please try again later.")
	}

	// Get user
	user, err := svc.sqlSvc.GetUserByEmailOrUsername(loginRequest.EmailOrUsername)
	if err != nil {
		svc.logAuthEvent("", "failed_login", clientIP, userAgent, false)
		return nil, shared.NewUnauthorizedError(err, "Invalid credentials")
	}

	// Check if account is locked
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		svc.logAuthEvent(user.ID, "failed_login_locked", clientIP, userAgent, false)
		return nil, shared.NewUnauthorizedError(errors.New("account locked"), "Account is temporarily locked due to too many failed attempts")
	}

	// Check password
	if !svc.checkPasswordHash(loginRequest.Password, user.Password) {
		// Increment failed attempts
		svc.sqlSvc.IncrementFailedAttempts(user.ID)

		// Lock account if too many attempts
		if user.FailedAttempts >= svc.maxLoginAttempts-1 {
			lockUntil := time.Now().Add(svc.lockoutDuration)
			svc.sqlSvc.LockAccount(user.ID, lockUntil)
		}

		svc.logAuthEvent(user.ID, "failed_login", clientIP, userAgent, false)
		return nil, shared.NewUnauthorizedError(errors.New("invalid password"), "Invalid credentials")
	}

	// Check email verification
	if svc.requireEmailVerify && !user.EmailVerified {
		return nil, shared.NewUnauthorizedError(errors.New("email not verified"), "Please verify your email address before logging in")
	}

	// Reset failed attempts on successful login
	svc.sqlSvc.ResetFailedAttempts(user.ID)

	// Generate tokens
	tokenPair, err := svc.jwtSvc.GenerateTokenPair(user.ID)
	if err != nil {
		return nil, shared.NewInternalError(err, "Failed to generate tokens")
	}

	// Create session
	session := dto.UserSession{
		UserID:    user.ID,
		TokenHash: svc.hashToken(tokenPair.RefreshToken),
		DeviceID:  loginRequest.DeviceID,
		IP:        clientIP,
		UserAgent: userAgent,
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
		IsActive:  true,
	}

	sessionID, err := svc.sqlSvc.CreateUserSession(session)
	if err != nil {
		return nil, shared.NewInternalError(err, "Failed to create session")
	}

	// Log successful login
	svc.logAuthEvent(user.ID, "login", clientIP, userAgent, true)

	// Update last login
	svc.sqlSvc.UpdateLastLogin(user.ID, clientIP)

	return &dto.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		SessionID:    sessionID,
		User: dto.UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		},
	}, nil
}

// Refresh Token
func (svc *AuthService) RefreshToken(refreshRequest dto.RefreshTokenRequest, clientIP, userAgent string) (*dto.LoginResponse, error) {
	// Verify refresh token
	userID, err := svc.jwtSvc.VerifyRefreshToken(refreshRequest.RefreshToken)
	if err != nil {
		return nil, shared.NewUnauthorizedError(err, "Invalid refresh token")
	}

	// Check if session exists and is active
	tokenHash := svc.hashToken(refreshRequest.RefreshToken)
	session, err := svc.sqlSvc.GetActiveSession(userID, tokenHash)
	if err != nil {
		return nil, shared.NewUnauthorizedError(err, "Session not found or expired")
	}

	// Update session last used
	svc.sqlSvc.UpdateSessionLastUsed(session.ID)

	// Generate new token pair
	tokenPair, err := svc.jwtSvc.GenerateTokenPair(userID)
	if err != nil {
		return nil, shared.NewInternalError(err, "Failed to generate tokens")
	}

	// Update session with new refresh token hash
	newTokenHash := svc.hashToken(tokenPair.RefreshToken)
	svc.sqlSvc.UpdateSessionToken(session.ID, newTokenHash)

	// Get user info
	user, err := svc.sqlSvc.GetUserByID(userID)
	if err != nil {
		return nil, shared.NewInternalError(err, "Failed to get user info")
	}

	svc.logAuthEvent(userID, "token_refresh", clientIP, userAgent, true)

	return &dto.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		SessionID:    session.ID,
		User: dto.UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		},
	}, nil
}

// Logout
func (svc *AuthService) Logout(userID, sessionID, clientIP, userAgent string) error {
	// Deactivate session
	err := svc.sqlSvc.DeactivateSession(sessionID, userID)
	if err != nil {
		return shared.NewInternalError(err, "Failed to logout")
	}

	svc.logAuthEvent(userID, "logout", clientIP, userAgent, true)
	return nil
}

// Logout from all devices
func (svc *AuthService) LogoutAllDevices(userID, currentSessionID, clientIP, userAgent string) error {
	// Deactivate all sessions except current one
	err := svc.sqlSvc.DeactivateAllUserSessions(userID, currentSessionID)
	if err != nil {
		return shared.NewInternalError(err, "Failed to logout from all devices")
	}

	svc.logAuthEvent(userID, "logout_all", clientIP, userAgent, true)
	return nil
}

// Email Verification
func (svc *AuthService) VerifyEmail(token string) error {
	user, err := svc.sqlSvc.GetUserByVerificationToken(token)
	if err != nil {
		return shared.NewBadRequestError(err, "Invalid verification token")
	}

	// Check if token is expired (24 hours)
	if user.CreatedAt.Add(24 * time.Hour).Before(time.Now()) {
		return shared.NewBadRequestError(errors.New("token expired"), "Verification token has expired")
	}

	// Mark email as verified
	err = svc.sqlSvc.VerifyUserEmail(user.ID)
	if err != nil {
		return shared.NewInternalError(err, "Failed to verify email")
	}

	svc.logAuthEvent(user.ID, "email_verified", "", "", true)
	return nil
}

// Resend Verification Email
func (svc *AuthService) ResendVerificationEmail(email string) error {
	user, err := svc.sqlSvc.GetUserByEmail(email)
	if err != nil {
		return shared.NewBadRequestError(err, "User not found")
	}

	if user.EmailVerified {
		return shared.NewBadRequestError(errors.New("already verified"), "Email is already verified")
	}

	// Generate new verification token
	verificationToken, err := svc.generateSecureToken()
	if err != nil {
		return shared.NewInternalError(err, "Failed to generate verification token")
	}

	// Update verification token
	err = svc.sqlSvc.UpdateVerificationToken(user.ID, verificationToken)
	if err != nil {
		return shared.NewInternalError(err, "Failed to update verification token")
	}

	// Send verification email
	err = svc.emailSvc.SendVerificationEmail(user.Email, user.Username, verificationToken)
	if err != nil {
		return shared.NewInternalError(err, "Failed to send verification email")
	}

	return nil
}

// Forgot Password
func (svc *AuthService) ForgotPassword(email string) error {
	user, err := svc.sqlSvc.GetUserByEmail(email)
	if err != nil {
		// Don't reveal if email exists
		return nil
	}

	// Generate reset token
	resetToken, err := svc.generateSecureToken()
	if err != nil {
		return shared.NewInternalError(err, "Failed to generate reset token")
	}

	// Store reset token (expires in 1 hour)
	expiresAt := time.Now().Add(time.Hour)
	err = svc.sqlSvc.CreatePasswordResetToken(user.ID, resetToken, expiresAt)
	if err != nil {
		return shared.NewInternalError(err, "Failed to create reset token")
	}

	// Send reset email
	err = svc.emailSvc.SendPasswordResetEmail(user.Email, user.Username, resetToken)
	if err != nil {
		return shared.NewInternalError(err, "Failed to send reset email")
	}

	svc.logAuthEvent(user.ID, "password_reset_requested", "", "", true)
	return nil
}

// Reset Password
func (svc *AuthService) ResetPassword(resetRequest dto.ResetPasswordRequest) error {
	// Validate new password
	if err := svc.validatePassword(resetRequest.NewPassword); err != nil {
		return shared.NewBadRequestError(err, err.Error())
	}

	// Verify reset token
	resetToken, err := svc.sqlSvc.GetPasswordResetToken(resetRequest.Token)
	if err != nil {
		return shared.NewBadRequestError(err, "Invalid reset token")
	}

	// Check if token is expired
	if resetToken.ExpiresAt.Before(time.Now()) {
		return shared.NewBadRequestError(errors.New("token expired"), "Reset token has expired")
	}

	// Hash new password
	hashedPassword, err := svc.hashPassword(resetRequest.NewPassword)
	if err != nil {
		return shared.NewInternalError(err, "Failed to hash password")
	}

	// Update password
	err = svc.sqlSvc.UpdateUserPassword(resetToken.UserID, hashedPassword)
	if err != nil {
		return shared.NewInternalError(err, "Failed to update password")
	}

	// Invalidate reset token
	svc.sqlSvc.InvalidatePasswordResetToken(resetRequest.Token)

	// Invalidate all user sessions (force re-login)
	svc.sqlSvc.DeactivateAllUserSessions(resetToken.UserID, "")

	svc.logAuthEvent(resetToken.UserID, "password_reset", "", "", true)
	return nil
}

// Change Password
func (svc *AuthService) ChangePassword(userID string, changeRequest dto.ChangePasswordRequest) error {
	// Get user
	user, err := svc.sqlSvc.GetUserByID(userID)
	if err != nil {
		return shared.NewInternalError(err, "User not found")
	}

	// Verify current password
	if !svc.checkPasswordHash(changeRequest.CurrentPassword, user.Password) {
		return shared.NewUnauthorizedError(errors.New("invalid password"), "Current password is incorrect")
	}

	// Validate new password
	if err := svc.validatePassword(changeRequest.NewPassword); err != nil {
		return shared.NewBadRequestError(err, err.Error())
	}

	// Hash new password
	hashedPassword, err := svc.hashPassword(changeRequest.NewPassword)
	if err != nil {
		return shared.NewInternalError(err, "Failed to hash password")
	}

	// Update password
	err = svc.sqlSvc.UpdateUserPassword(userID, hashedPassword)
	if err != nil {
		return shared.NewInternalError(err, "Failed to update password")
	}

	svc.logAuthEvent(userID, "password_changed", "", "", true)
	return nil
}

// Utility functions
func (svc *AuthService) hashToken(token string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(token), 10)
	return string(hash)
}

func (svc *AuthService) logAuthEvent(userID, action, ip, userAgent string, success bool) {
	auditLog := dto.AuthAuditLog{
		UserID:    userID,
		Action:    action,
		IP:        ip,
		UserAgent: userAgent,
		Timestamp: time.Now(),
		Success:   success,
	}

	// Log to database (implement in SQLite service)
	svc.sqlSvc.CreateAuthAuditLog(auditLog)
}

// Middleware Functions
func (svc *AuthService) RequiredAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		token, err := svc.jwtSvc.ExtractTokenFromHeader(authHeader)
		if err != nil {
			return shared.ResponseJSON(c, http.StatusUnauthorized, "Unauthorized", err.Error())
		}

		userID, err := svc.jwtSvc.VerifyJWTToken(token)
		if err != nil {
			return shared.ResponseJSON(c, http.StatusUnauthorized, "Unauthorized", "Invalid JWT token")
		}

		if userID == "" {
			return shared.ResponseJSON(c, http.StatusUnauthorized, "Unauthorized", "Invalid user ID in token")
		}

		// Check if user exists and is active
		user, err := svc.sqlSvc.GetUserByID(userID)
		if err != nil || !user.IsActive {
			return shared.ResponseJSON(c, http.StatusUnauthorized, "Unauthorized", "User account is inactive")
		}

		c.Locals(shared.UserID, userID)
		c.Locals("user", user)
		return c.Next()
	}
}

func (svc *AuthService) RequireRole(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user")
		if user == nil {
			return shared.ResponseJSON(c, http.StatusUnauthorized, "Unauthorized", "User not found in context")
		}

		userObj := user.(*model.User)
		if userObj.Role != role {
			return shared.ResponseJSON(c, http.StatusForbidden, "Forbidden", "Insufficient permissions")
		}

		return c.Next()
	}
}

func (svc *AuthService) RequireEmailVerified() fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user")
		if user == nil {
			return shared.ResponseJSON(c, http.StatusUnauthorized, "Unauthorized", "User not found in context")
		}

		userObj := user.(*model.User)
		if !userObj.EmailVerified {
			return shared.ResponseJSON(c, http.StatusForbidden, "Forbidden", "Email verification required")
		}

		return c.Next()
	}
}
