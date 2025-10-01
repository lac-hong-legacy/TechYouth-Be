package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lac-hong-legacy/ven_api/dto"
	"github.com/lac-hong-legacy/ven_api/model"
	"github.com/lac-hong-legacy/ven_api/shared"
	"golang.org/x/crypto/bcrypt"

	"github.com/alphabatem/common/context"
	log "github.com/sirupsen/logrus"
)

type VerificationEmail struct {
	Email             string
	Username          string
	VerificationToken string
}

type PasswordResetEmail struct {
	Email      string
	Username   string
	ResetToken string
}

type LoginNotificationEmail struct {
	Email     string
	Username  string
	LoginTime string
	IP        string
	Device    string
	Location  string
}

type AuthService struct {
	context.DefaultService

	sqlSvc       *SqliteService
	jwtSvc       *JWTService
	emailSvc     *EmailService
	rateLimitSvc *RateLimitService

	maxLoginAttempts   int
	lockoutDuration    time.Duration
	passwordMinLength  int
	requireEmailVerify bool

	sendVerificationEmailAsync      chan VerificationEmail
	sendPasswordResetEmailAsync     chan PasswordResetEmail
	sendLoginNotificationEmailAsync chan LoginNotificationEmail
	logAuthEventCh                  chan dto.AuthAuditLog
	dbOperationCh                   chan func()
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

	svc.sendVerificationEmailAsync = make(chan VerificationEmail, 100)
	svc.sendPasswordResetEmailAsync = make(chan PasswordResetEmail, 100)
	svc.sendLoginNotificationEmailAsync = make(chan LoginNotificationEmail, 100)
	svc.logAuthEventCh = make(chan dto.AuthAuditLog, 100)
	svc.dbOperationCh = make(chan func(), 100)

	return svc.DefaultService.Configure(ctx)
}

func (svc *AuthService) Start() error {
	svc.sqlSvc = svc.Service(SQLITE_SVC).(*SqliteService)
	svc.jwtSvc = svc.Service(JWT_SVC).(*JWTService)
	svc.emailSvc = svc.Service(EMAIL_SVC).(*EmailService)
	svc.rateLimitSvc = svc.Service(RATE_LIMIT_SVC).(*RateLimitService)

	go svc.startVerificationEmailJob()
	go svc.startPasswordResetEmailJob()
	go svc.startLoginNotificationEmailJob()
	go svc.startLogAuthEventJob()
	go svc.startDBOperationJob()

	return nil
}

func (svc *AuthService) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

func (svc *AuthService) checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

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

func (svc *AuthService) generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func (svc *AuthService) Register(registerRequest dto.RegisterRequest) (*dto.RegisterResponse, error) {
	_, err := svc.sqlSvc.GetUserByUsername(registerRequest.Username)
	if err == nil {
		return nil, shared.NewBadRequestError(errors.New("username taken"), "Username is already taken")
	}

	if err := svc.validatePassword(registerRequest.Password); err != nil {
		return nil, shared.NewBadRequestError(err, err.Error())
	}

	hashedPassword, err := svc.hashPassword(registerRequest.Password)
	if err != nil {
		return nil, shared.NewInternalError(err, "Failed to hash password")
	}

	verificationToken, err := svc.generateSecureToken()
	if err != nil {
		return nil, shared.NewInternalError(err, "Failed to generate verification token")
	}

	registerRequest.Password = hashedPassword
	user, err := svc.sqlSvc.CreateUser(registerRequest, verificationToken)
	if err != nil {
		return nil, shared.NewInternalError(err, err.Error())
	}

	if svc.requireEmailVerify {
		svc.sendVerificationEmailAsync <- VerificationEmail{
			Email:             registerRequest.Email,
			Username:          registerRequest.Username,
			VerificationToken: verificationToken,
		}
	}

	svc.logAuthEventCh <- dto.AuthAuditLog{
		UserID:    user.ID,
		Action:    "register",
		IP:        "",
		UserAgent: "",
		Timestamp: time.Now(),
		Success:   true,
	}

	return &dto.RegisterResponse{
		UserID:               user.ID,
		RequiresVerification: svc.requireEmailVerify,
	}, nil
}

func (svc *AuthService) Login(loginRequest dto.LoginRequest, clientIP, userAgent string) (*dto.LoginResponse, error) {
	if blocked := svc.rateLimitSvc.IsBlocked(clientIP, "login"); blocked {
		return nil, shared.NewTooManyRequestsError(errors.New("too many login attempts"), "Too many login attempts. Please try again later.")
	}

	user, err := svc.sqlSvc.GetUserByEmailOrUsername(loginRequest.EmailOrUsername)
	if err != nil {
		svc.logAuthEventCh <- dto.AuthAuditLog{
			UserID:    "",
			Action:    "failed_login",
			IP:        clientIP,
			UserAgent: userAgent,
			Timestamp: time.Now(),
			Success:   false,
		}
		return nil, shared.NewUnauthorizedError(err, "Invalid credentials")
	}

	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		svc.logAuthEventCh <- dto.AuthAuditLog{
			UserID:    user.ID,
			Action:    "failed_login_locked",
			IP:        clientIP,
			UserAgent: userAgent,
			Timestamp: time.Now(),
			Success:   false,
		}
		return nil, shared.NewUnauthorizedError(errors.New("account locked"), "Account is temporarily locked due to too many failed attempts")
	}

	if !svc.checkPasswordHash(loginRequest.Password, user.Password) {
		svc.dbOperationCh <- func() {
			svc.sqlSvc.IncrementFailedAttempts(user.ID)
		}

		if user.FailedAttempts >= svc.maxLoginAttempts-1 {
			lockUntil := time.Now().Add(svc.lockoutDuration)
			svc.dbOperationCh <- func() {
				svc.sqlSvc.LockAccount(user.ID, lockUntil)
			}
		}

		svc.logAuthEventCh <- dto.AuthAuditLog{
			UserID:    user.ID,
			Action:    "failed_login",
			IP:        clientIP,
			UserAgent: userAgent,
			Timestamp: time.Now(),
			Success:   false,
		}
		return nil, shared.NewUnauthorizedError(errors.New("invalid password"), "Invalid credentials")
	}

	if svc.requireEmailVerify && !user.EmailVerified {
		return nil, shared.NewUnauthorizedError(errors.New("email not verified"), "Please verify your email address before logging in")
	}

	svc.dbOperationCh <- func() {
		svc.sqlSvc.ResetFailedAttempts(user.ID)
	}

	tokenPair, err := svc.jwtSvc.GenerateTokenPair(user.ID)
	if err != nil {
		return nil, shared.NewInternalError(err, "Failed to generate tokens")
	}

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

	svc.logAuthEventCh <- dto.AuthAuditLog{
		UserID:    user.ID,
		Action:    "login",
		IP:        clientIP,
		UserAgent: userAgent,
		Timestamp: time.Now(),
		Success:   true,
	}

	svc.dbOperationCh <- func() {
		svc.sqlSvc.UpdateLastLogin(user.ID, clientIP)
	}

	// Send login notification email
	svc.sendLoginNotificationEmailAsync <- LoginNotificationEmail{
		Email:     user.Email,
		Username:  user.Username,
		LoginTime: time.Now().Format("2006-01-02 15:04:05"),
		IP:        clientIP,
		Device:    userAgent,
		Location:  "Unknown", // You can integrate with IP geolocation service
	}

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

func (svc *AuthService) RefreshToken(refreshRequest dto.RefreshTokenRequest, clientIP, userAgent string) (*dto.LoginResponse, error) {
	userID, err := svc.jwtSvc.VerifyRefreshToken(refreshRequest.RefreshToken)
	if err != nil {
		return nil, shared.NewUnauthorizedError(err, "Invalid refresh token")
	}

	tokenHash := svc.hashToken(refreshRequest.RefreshToken)
	session, err := svc.sqlSvc.GetActiveSession(userID, tokenHash)
	if err != nil {
		return nil, shared.NewUnauthorizedError(err, "Session not found or expired")
	}

	svc.dbOperationCh <- func() {
		svc.sqlSvc.UpdateSessionLastUsed(session.ID)
	}

	tokenPair, err := svc.jwtSvc.GenerateTokenPair(userID)
	if err != nil {
		return nil, shared.NewInternalError(err, "Failed to generate tokens")
	}

	newTokenHash := svc.hashToken(tokenPair.RefreshToken)
	svc.dbOperationCh <- func() {
		svc.sqlSvc.UpdateSessionToken(session.ID, newTokenHash)
	}

	user, err := svc.sqlSvc.GetUserByID(userID)
	if err != nil {
		return nil, shared.NewInternalError(err, "Failed to get user info")
	}

	svc.logAuthEventCh <- dto.AuthAuditLog{
		UserID:    userID,
		Action:    "token_refresh",
		IP:        clientIP,
		UserAgent: userAgent,
		Timestamp: time.Now(),
		Success:   true,
	}

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

func (svc *AuthService) Logout(userID, sessionID, clientIP, userAgent string) error {
	err := svc.sqlSvc.DeactivateSession(sessionID, userID)
	if err != nil {
		return shared.NewInternalError(err, "Failed to logout")
	}

	svc.logAuthEventCh <- dto.AuthAuditLog{
		UserID:    userID,
		Action:    "logout",
		IP:        clientIP,
		UserAgent: userAgent,
		Timestamp: time.Now(),
		Success:   true,
	}
	return nil
}

func (svc *AuthService) LogoutAllDevices(userID, currentSessionID, clientIP, userAgent string) error {
	err := svc.sqlSvc.DeactivateAllUserSessions(userID, currentSessionID)
	if err != nil {
		return shared.NewInternalError(err, "Failed to logout from all devices")
	}

	svc.logAuthEventCh <- dto.AuthAuditLog{
		UserID:    userID,
		Action:    "logout_all",
		IP:        clientIP,
		UserAgent: userAgent,
		Timestamp: time.Now(),
		Success:   true,
	}
	return nil
}

func (svc *AuthService) VerifyEmail(token string) error {
	user, err := svc.sqlSvc.GetUserByVerificationToken(token)
	if err != nil {
		return shared.NewBadRequestError(err, "Invalid verification token")
	}

	if user.CreatedAt.Add(24 * time.Hour).Before(time.Now()) {
		return shared.NewBadRequestError(errors.New("token expired"), "Verification token has expired")
	}

	err = svc.sqlSvc.VerifyUserEmail(user.ID)
	if err != nil {
		return shared.NewInternalError(err, "Failed to verify email")
	}

	svc.logAuthEventCh <- dto.AuthAuditLog{
		UserID:    user.ID,
		Action:    "email_verified",
		IP:        "",
		UserAgent: "",
		Timestamp: time.Now(),
		Success:   true,
	}
	return nil
}

func (svc *AuthService) ResendVerificationEmail(email string) error {
	user, err := svc.sqlSvc.GetUserByEmail(email)
	if err != nil {
		return shared.NewBadRequestError(err, "User not found")
	}

	if user.EmailVerified {
		return shared.NewBadRequestError(errors.New("already verified"), "Email is already verified")
	}

	verificationToken, err := svc.generateSecureToken()
	if err != nil {
		return shared.NewInternalError(err, "Failed to generate verification token")
	}

	err = svc.sqlSvc.UpdateVerificationToken(user.ID, verificationToken)
	if err != nil {
		return shared.NewInternalError(err, "Failed to update verification token")
	}

	svc.sendVerificationEmailAsync <- VerificationEmail{
		Email:             user.Email,
		Username:          user.Username,
		VerificationToken: verificationToken,
	}

	return nil
}

func (svc *AuthService) ForgotPassword(email string) error {
	user, err := svc.sqlSvc.GetUserByEmail(email)
	if err != nil {
		return nil
	}

	resetToken, err := svc.generateSecureToken()
	if err != nil {
		return shared.NewInternalError(err, "Failed to generate reset token")
	}

	expiresAt := time.Now().Add(time.Hour)
	err = svc.sqlSvc.CreatePasswordResetToken(user.ID, resetToken, expiresAt)
	if err != nil {
		return shared.NewInternalError(err, "Failed to create reset token")
	}

	svc.sendPasswordResetEmailAsync <- PasswordResetEmail{
		Email:      user.Email,
		Username:   user.Username,
		ResetToken: resetToken,
	}

	svc.logAuthEventCh <- dto.AuthAuditLog{
		UserID:    user.ID,
		Action:    "password_reset_requested",
		IP:        "",
		UserAgent: "",
		Timestamp: time.Now(),
		Success:   true,
	}
	return nil
}

func (svc *AuthService) ResetPassword(resetRequest dto.ResetPasswordRequest) error {
	if err := svc.validatePassword(resetRequest.NewPassword); err != nil {
		return shared.NewBadRequestError(err, err.Error())
	}

	resetToken, err := svc.sqlSvc.GetPasswordResetToken(resetRequest.Token)
	if err != nil {
		return shared.NewBadRequestError(err, "Invalid reset token")
	}

	if resetToken.ExpiresAt.Before(time.Now()) {
		return shared.NewBadRequestError(errors.New("token expired"), "Reset token has expired")
	}

	hashedPassword, err := svc.hashPassword(resetRequest.NewPassword)
	if err != nil {
		return shared.NewInternalError(err, "Failed to hash password")
	}

	err = svc.sqlSvc.UpdateUserPassword(resetToken.UserID, hashedPassword)
	if err != nil {
		return shared.NewInternalError(err, "Failed to update password")
	}

	svc.dbOperationCh <- func() {
		svc.sqlSvc.InvalidatePasswordResetToken(resetRequest.Token)
	}

	svc.dbOperationCh <- func() {
		svc.sqlSvc.DeactivateAllUserSessions(resetToken.UserID, "")
	}

	svc.logAuthEventCh <- dto.AuthAuditLog{
		UserID:    resetToken.UserID,
		Action:    "password_reset",
		IP:        "",
		UserAgent: "",
		Timestamp: time.Now(),
		Success:   true,
	}
	return nil
}

func (svc *AuthService) ChangePassword(userID string, changeRequest dto.ChangePasswordRequest) error {
	user, err := svc.sqlSvc.GetUserByID(userID)
	if err != nil {
		return shared.NewInternalError(err, "User not found")
	}

	if !svc.checkPasswordHash(changeRequest.CurrentPassword, user.Password) {
		return shared.NewUnauthorizedError(errors.New("invalid password"), "Current password is incorrect")
	}

	if err := svc.validatePassword(changeRequest.NewPassword); err != nil {
		return shared.NewBadRequestError(err, err.Error())
	}

	hashedPassword, err := svc.hashPassword(changeRequest.NewPassword)
	if err != nil {
		return shared.NewInternalError(err, "Failed to hash password")
	}

	err = svc.sqlSvc.UpdateUserPassword(userID, hashedPassword)
	if err != nil {
		return shared.NewInternalError(err, "Failed to update password")
	}

	svc.logAuthEventCh <- dto.AuthAuditLog{
		UserID:    userID,
		Action:    "password_changed",
		IP:        "",
		UserAgent: "",
		Timestamp: time.Now(),
		Success:   true,
	}
	return nil
}

func (svc *AuthService) startDBOperationJob() {
	for operation := range svc.dbOperationCh {
		operation()
	}
}

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

func (svc *AuthService) startVerificationEmailJob() {
	for email := range svc.sendVerificationEmailAsync {
		err := svc.emailSvc.SendVerificationEmail(email.Email, email.Username, email.VerificationToken)
		if err != nil {
			log.WithError(err).Error("Failed to send verification email")
		}
	}
}

func (svc *AuthService) startPasswordResetEmailJob() {
	for email := range svc.sendPasswordResetEmailAsync {
		err := svc.emailSvc.SendPasswordResetEmail(email.Email, email.Username, email.ResetToken)
		if err != nil {
			log.WithError(err).Error("Failed to send password reset email")
		}
	}
}

func (svc *AuthService) startLoginNotificationEmailJob() {
	for email := range svc.sendLoginNotificationEmailAsync {
		err := svc.emailSvc.SendLoginNotificationEmail(email.Email, email.Username, email.LoginTime, email.IP, email.Device, email.Location)
		if err != nil {
			log.WithError(err).Error("Failed to send login notification email")
		}
	}
}

func (svc *AuthService) startLogAuthEventJob() {
	for auditLog := range svc.logAuthEventCh {
		svc.sqlSvc.CreateAuthAuditLog(auditLog)
	}
}

func (svc *AuthService) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
