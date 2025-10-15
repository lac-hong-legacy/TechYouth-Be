package services

import (
	"crypto/rand"
	"crypto/sha256"
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

	"github.com/cloakd/common/context"
	serviceContext "github.com/cloakd/common/services"
	log "github.com/sirupsen/logrus"
)

type VerificationEmail struct {
	Email            string
	Username         string
	VerificationCode string
}

type PasswordResetEmail struct {
	Email     string
	Username  string
	ResetCode string
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
	serviceContext.DefaultService

	sqlSvc         *PostgresService
	jwtSvc         *JWTService
	emailSvc       *EmailService
	rateLimitSvc   *RateLimitService
	geolocationSvc *GeolocationService

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
	svc.sqlSvc = svc.Service(POSTGRES_SVC).(*PostgresService)
	svc.jwtSvc = svc.Service(JWT_SVC).(*JWTService)
	svc.emailSvc = svc.Service(EMAIL_SVC).(*EmailService)
	svc.rateLimitSvc = svc.Service(RATE_LIMIT_SVC).(*RateLimitService)
	svc.geolocationSvc = svc.Service(GEOLOCATION_SVC).(*GeolocationService)

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

func (svc *AuthService) generateVerificationCode() (string, error) {
	// Generate a random 6-digit code (100000 to 999999)
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	// Convert to number between 100000 and 999999
	code := 100000 + (int(bytes[0])<<16|int(bytes[1])<<8|int(bytes[2]))%900000
	return fmt.Sprintf("%06d", code), nil
}

func (svc *AuthService) Register(registerRequest dto.RegisterRequest) (*dto.RegisterResponse, error) {
	_, err := svc.sqlSvc.userRepo.GetUserByUsername(registerRequest.Username)
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

	verificationCode, err := svc.generateVerificationCode()
	if err != nil {
		return nil, shared.NewInternalError(err, "Failed to generate verification code")
	}

	registerRequest.Password = hashedPassword
	user, err := svc.sqlSvc.userRepo.CreateUser(registerRequest, verificationCode)
	if err != nil {
		return nil, shared.NewInternalError(err, err.Error())
	}

	if svc.requireEmailVerify {
		svc.sendVerificationEmailAsync <- VerificationEmail{
			Email:            registerRequest.Email,
			Username:         registerRequest.Username,
			VerificationCode: verificationCode,
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
		Message:              "Registration successful. Please check your email for verification.",
	}, nil
}

func (svc *AuthService) Login(loginRequest dto.LoginRequest, clientIP, userAgent string) (*dto.LoginResponse, error) {
	// if blocked := svc.rateLimitSvc.IsBlocked(clientIP, "login"); blocked {
	// 	return nil, shared.NewTooManyRequestsError(errors.New("too many login attempts"), "Too many login attempts. Please try again later.")
	// }

	user, err := svc.sqlSvc.userRepo.GetUserByEmailOrUsername(loginRequest.EmailOrUsername)
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
			svc.sqlSvc.userRepo.IncrementFailedAttempts(user.ID)
		}

		if user.FailedAttempts >= svc.maxLoginAttempts-1 {
			lockUntil := time.Now().Add(svc.lockoutDuration)
			svc.dbOperationCh <- func() {
				svc.sqlSvc.userRepo.LockAccount(user.ID, lockUntil)
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
		svc.sqlSvc.userRepo.ResetFailedAttempts(user.ID)
	}

	// Generate tokens
	tokenPair, err := svc.jwtSvc.GenerateTokenPair(user.ID)
	if err != nil {
		return nil, shared.NewInternalError(err, "Failed to generate tokens")
	}

	refreshClaims, err := svc.jwtSvc.GetTokenClaims(tokenPair.RefreshToken)
	if err != nil {
		return nil, shared.NewInternalError(err, "Failed to extract refresh token claims")
	}

	// Create session with refresh token hash and JTI
	session := dto.UserSession{
		UserID:           user.ID,
		TokenHash:        svc.hashToken(tokenPair.RefreshToken),
		RefreshTokenJTI:  refreshClaims.ID,
		RefreshExpiresAt: refreshClaims.ExpiresAt.Time,
		DeviceID:         loginRequest.DeviceID,
		IP:               clientIP,
		UserAgent:        userAgent,
		CreatedAt:        time.Now(),
		LastUsed:         time.Now(),
		IsActive:         true,
	}

	sessionID, err := svc.sqlSvc.userRepo.CreateUserSession(session)
	if err != nil {
		return nil, shared.NewInternalError(err, "Failed to create session")
	}

	accessToken, err := svc.jwtSvc.GenerateAccessTokenWithSession(user.ID, sessionID)
	if err != nil {
		return nil, shared.NewInternalError(err, "Failed to generate access token with session")
	}

	tokenPair.AccessToken = accessToken

	svc.logAuthEventCh <- dto.AuthAuditLog{
		UserID:    user.ID,
		Action:    "login",
		IP:        clientIP,
		UserAgent: userAgent,
		Timestamp: time.Now(),
		Success:   true,
	}

	svc.dbOperationCh <- func() {
		svc.sqlSvc.userRepo.UpdateLastLogin(user.ID, clientIP)
	}

	location, geoErr := svc.geolocationSvc.GetLocationByIP(clientIP)
	if geoErr != nil {
		location = "Unknown"
	}

	// Send login notification email
	svc.sendLoginNotificationEmailAsync <- LoginNotificationEmail{
		Email:     user.Email,
		Username:  user.Username,
		LoginTime: time.Now().Local().Format("2006-01-02 15:04:05"),
		IP:        clientIP,
		Device:    userAgent,
		Location:  location,
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
	session, err := svc.sqlSvc.userRepo.GetActiveSession(userID, tokenHash)
	if err != nil {
		return nil, shared.NewUnauthorizedError(err, "Session not found or expired")
	}

	svc.dbOperationCh <- func() {
		svc.sqlSvc.userRepo.UpdateSessionLastUsed(session.ID)
	}

	// Generate tokens with session_id
	tokenPair, err := svc.jwtSvc.GenerateTokenPairWithSession(userID, session.ID)
	if err != nil {
		return nil, shared.NewInternalError(err, "Failed to generate tokens")
	}

	newTokenHash := svc.hashToken(tokenPair.RefreshToken)
	svc.dbOperationCh <- func() {
		svc.sqlSvc.userRepo.UpdateSessionToken(session.ID, newTokenHash)
	}

	user, err := svc.sqlSvc.userRepo.GetUserByID(userID)
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

func (svc *AuthService) BlacklistToken(accessToken, refreshToken string) error {
	if err := svc.jwtSvc.BlacklistToken(accessToken); err != nil {
		return shared.NewInternalError(err, "Failed to blacklist access token")
	}

	if err := svc.jwtSvc.BlacklistToken(refreshToken); err != nil {
		return shared.NewInternalError(err, "Failed to blacklist refresh token")
	}
	return nil
}

func (svc *AuthService) Logout(userID, sessionID, accessToken, clientIP, userAgent string) error {
	if accessToken != "" {
		if err := svc.jwtSvc.BlacklistToken(accessToken); err != nil {
			log.WithError(err).Error("Failed to blacklist access token")
		}
	}

	session, err := svc.sqlSvc.userRepo.GetSessionByID(sessionID)
	if err == nil && session != nil && session.RefreshTokenJTI != "" {
		if err := svc.sqlSvc.userRepo.BlacklistToken(session.RefreshTokenJTI, session.RefreshExpiresAt); err != nil {
			log.WithError(err).Error("Failed to blacklist refresh token")
		}
	}

	err = svc.sqlSvc.userRepo.DeactivateSession(sessionID, userID)
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
		Details:   "Access & Refresh tokens blacklisted, session deactivated",
	}
	return nil
}

func (svc *AuthService) LogoutAllDevices(userID, currentSessionID, accessToken, clientIP, userAgent string) error {
	if accessToken != "" {
		if err := svc.jwtSvc.BlacklistToken(accessToken); err != nil {
			log.WithError(err).Error("Failed to blacklist current access token")
		}
	}

	sessions, err := svc.sqlSvc.userRepo.GetUserActiveSessions(userID)
	if err == nil {
		for _, session := range sessions {
			if session.RefreshTokenJTI != "" {
				if err := svc.sqlSvc.userRepo.BlacklistToken(session.RefreshTokenJTI, session.RefreshExpiresAt); err != nil {
					log.WithError(err).Errorf("Failed to blacklist refresh token for session %s", session.ID)
				}
			}
		}
	}

	err = svc.sqlSvc.userRepo.DeactivateAllUserSessions(userID, currentSessionID)
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
		Details:   "All access & refresh tokens blacklisted, all sessions deactivated",
	}
	return nil
}

func (svc *AuthService) VerifyEmail(email, code string) error {
	user, err := svc.sqlSvc.userRepo.GetUserByVerificationCode(email, code)
	if err != nil {
		return shared.NewBadRequestError(err, "Invalid verification code or email")
	}

	if user.EmailVerified {
		return shared.NewBadRequestError(errors.New("already verified"), "Email is already verified")
	}

	// Check if code has expired
	if user.VerificationCodeExpiry == nil || user.VerificationCodeExpiry.Before(time.Now()) {
		return shared.NewBadRequestError(errors.New("code expired"), "Verification code has expired. Please request a new one")
	}

	err = svc.sqlSvc.userRepo.VerifyUserEmail(user.ID)
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
	user, err := svc.sqlSvc.userRepo.GetUserByEmail(email)
	if err != nil {
		return shared.NewBadRequestError(err, "User not found")
	}

	if user.EmailVerified {
		return shared.NewBadRequestError(errors.New("already verified"), "Email is already verified")
	}

	verificationCode, err := svc.generateVerificationCode()
	if err != nil {
		return shared.NewInternalError(err, "Failed to generate verification code")
	}

	err = svc.sqlSvc.userRepo.UpdateVerificationCode(user.ID, verificationCode)
	if err != nil {
		return shared.NewInternalError(err, "Failed to update verification code")
	}

	svc.sendVerificationEmailAsync <- VerificationEmail{
		Email:            user.Email,
		Username:         user.Username,
		VerificationCode: verificationCode,
	}

	return nil
}

func (svc *AuthService) ForgotPassword(email string) error {
	user, err := svc.sqlSvc.userRepo.GetUserByEmail(email)
	if err != nil {
		return nil
	}

	resetCode, err := svc.generateVerificationCode()
	if err != nil {
		return shared.NewInternalError(err, "Failed to generate reset code")
	}

	expiresAt := time.Now().Add(time.Hour)
	err = svc.sqlSvc.userRepo.CreatePasswordResetCode(user.ID, resetCode, expiresAt)
	if err != nil {
		return shared.NewInternalError(err, "Failed to create reset code")
	}

	svc.sendPasswordResetEmailAsync <- PasswordResetEmail{
		Email:     user.Email,
		Username:  user.Username,
		ResetCode: resetCode,
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

	resetCode, err := svc.sqlSvc.userRepo.GetPasswordResetCode(resetRequest.Code)
	if err != nil {
		return shared.NewBadRequestError(err, "Invalid reset code")
	}

	if resetCode.ExpiresAt.Before(time.Now()) {
		return shared.NewBadRequestError(errors.New("code expired"), "Reset code has expired")
	}

	hashedPassword, err := svc.hashPassword(resetRequest.NewPassword)
	if err != nil {
		return shared.NewInternalError(err, "Failed to hash password")
	}

	err = svc.sqlSvc.userRepo.UpdateUserPassword(resetCode.UserID, hashedPassword)
	if err != nil {
		return shared.NewInternalError(err, "Failed to update password")
	}

	svc.dbOperationCh <- func() {
		svc.sqlSvc.userRepo.InvalidatePasswordResetCode(resetRequest.Code)
	}

	svc.dbOperationCh <- func() {
		svc.sqlSvc.userRepo.DeactivateAllUserSessions(resetCode.UserID, "")
	}

	svc.logAuthEventCh <- dto.AuthAuditLog{
		UserID:    resetCode.UserID,
		Action:    "password_reset",
		IP:        "",
		UserAgent: "",
		Timestamp: time.Now(),
		Success:   true,
	}
	return nil
}

func (svc *AuthService) ChangePassword(userID string, changeRequest dto.ChangePasswordRequest) error {
	user, err := svc.sqlSvc.userRepo.GetUserByID(userID)
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

	err = svc.sqlSvc.userRepo.UpdateUserPassword(userID, hashedPassword)
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

		claims, err := svc.jwtSvc.VerifyAndGetClaims(token)
		if err != nil {
			return shared.ResponseJSON(c, http.StatusUnauthorized, "Unauthorized", "Invalid JWT token")
		}

		if claims.UserID == "" {
			return shared.ResponseJSON(c, http.StatusUnauthorized, "Unauthorized", "Invalid user ID in token")
		}

		// Check if user exists and is active
		user, err := svc.sqlSvc.userRepo.GetUserByID(claims.UserID)
		if err != nil || !user.IsActive {
			return shared.ResponseJSON(c, http.StatusUnauthorized, "Unauthorized", "User account is inactive")
		}

		c.Locals(shared.UserID, claims.UserID)
		c.Locals("user", user)
		c.Locals("session_id", claims.SessionID)
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
		err := svc.emailSvc.SendVerificationEmail(email.Email, email.Username, email.VerificationCode)
		if err != nil {
			log.WithError(err).Error("Failed to send verification email")
		}
	}
}

func (svc *AuthService) startPasswordResetEmailJob() {
	for email := range svc.sendPasswordResetEmailAsync {
		err := svc.emailSvc.SendPasswordResetEmail(email.Email, email.Username, email.ResetCode)
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
		svc.sqlSvc.userRepo.CreateAuthAuditLog(auditLog)
	}
}

func (svc *AuthService) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (svc *AuthService) GetDetailedLocationInfo(ip string) (*GeolocationResponse, error) {
	return svc.geolocationSvc.GetDetailedLocationByIP(ip)
}

func (svc *AuthService) GetUserDevices(userID string) ([]dto.DeviceInfo, error) {
	devices, err := svc.sqlSvc.userRepo.GetUserTrustedDevices(userID)
	if err != nil {
		return nil, shared.NewInternalError(err, "Failed to get user devices")
	}

	deviceInfos := make([]dto.DeviceInfo, 0, len(devices))
	for _, device := range devices {
		deviceInfos = append(deviceInfos, dto.DeviceInfo{
			ID:        device.ID,
			Name:      device.Name,
			Type:      device.Type,
			OS:        device.OS,
			Browser:   device.Browser,
			IP:        device.IP,
			LastUsed:  device.LastUsed,
			IsTrusted: device.IsTrusted,
		})
	}

	return deviceInfos, nil
}

func (svc *AuthService) UpdateDeviceTrust(userID, deviceID string, trust bool) error {
	device, err := svc.sqlSvc.userRepo.GetTrustedDevice(userID, deviceID)
	if err != nil {
		return shared.NewNotFoundError(err, "Device not found")
	}

	device.IsTrusted = trust
	if err := svc.sqlSvc.userRepo.UpdateTrustedDevice(device); err != nil {
		return shared.NewInternalError(err, "Failed to update device trust")
	}

	action := "device_untrusted"
	if trust {
		action = "device_trusted"
	}

	svc.logAuthEventCh <- dto.AuthAuditLog{
		UserID:    userID,
		Action:    action,
		Timestamp: time.Now(),
		Success:   true,
		Details:   fmt.Sprintf("Device %s", deviceID),
	}

	return nil
}

func (svc *AuthService) RemoveDevice(userID, deviceID string) error {
	if err := svc.sqlSvc.userRepo.RemoveTrustedDevice(userID, deviceID); err != nil {
		return shared.NewInternalError(err, "Failed to remove device")
	}

	svc.logAuthEventCh <- dto.AuthAuditLog{
		UserID:    userID,
		Action:    "device_removed",
		Timestamp: time.Now(),
		Success:   true,
		Details:   fmt.Sprintf("Device %s", deviceID),
	}

	return nil
}

func (svc *AuthService) RegisterOrUpdateDevice(userID, deviceID, name, deviceType, os, browser, ip string) error {
	device, err := svc.sqlSvc.userRepo.GetTrustedDevice(userID, deviceID)
	if err == nil {
		device.LastUsed = time.Now()
		device.IP = ip
		return svc.sqlSvc.userRepo.UpdateTrustedDevice(device)
	}

	newDevice := &model.TrustedDevice{
		UserID:    userID,
		DeviceID:  deviceID,
		Name:      name,
		Type:      deviceType,
		OS:        os,
		Browser:   browser,
		IP:        ip,
		IsTrusted: false,
	}

	return svc.sqlSvc.userRepo.CreateTrustedDevice(newDevice)
}
