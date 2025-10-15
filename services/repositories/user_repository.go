package repositories

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lac-hong-legacy/ven_api/dto"
	"github.com/lac-hong-legacy/ven_api/model"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserRepository handles user-related database operations
type UserRepository struct {
	BaseRepository
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (ds *UserRepository) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	if err := ds.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (ds *UserRepository) GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	if err := ds.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (ds *UserRepository) GetUserByEmailOrUsername(emailOrUsername string) (*model.User, error) {
	var user model.User
	if err := ds.db.Where("email = ? OR username = ?", emailOrUsername, emailOrUsername).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (ds *UserRepository) GetUser(userID string) (*model.User, error) {
	var user model.User
	if err := ds.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (ds *UserRepository) UpdateUser(user *model.User) error {
	user.UpdatedAt = time.Now()
	if err := ds.db.Save(user).Error; err != nil {
		return err
	}
	return nil
}

func (ds *UserRepository) CreateUser(req dto.RegisterRequest, verificationCode string) (*model.User, error) {
	codeExpiry := time.Now().Add(15 * time.Minute) // Code expires in 15 minutes
	user := &model.User{
		ID:                     uuid.New().String(),
		Username:               req.Username,
		Email:                  req.Email,
		Password:               req.Password,
		Role:                   model.RoleUser,
		IsActive:               true,
		EmailVerified:          false,
		VerificationCode:       verificationCode,
		VerificationCodeExpiry: &codeExpiry,
		FailedAttempts:         0,
		LoginNotifications:     true,
		SessionTimeout:         1440, // 24 hours
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}

	if err := ds.db.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (ds *UserRepository) GetUserByID(userID string) (*model.User, error) {
	var user model.User
	err := ds.db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (ds *UserRepository) GetUserByVerificationCode(email, code string) (*model.User, error) {
	var user model.User
	err := ds.db.Where("email = ? AND verification_code = ?", email, code).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (ds *UserRepository) UpdateUserPassword(userID, hashedPassword string) error {
	now := time.Now()
	return ds.db.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"password":             hashedPassword,
		"last_password_change": &now,
		"updated_at":           now,
	}).Error
}

func (ds *UserRepository) UpdateLastLogin(userID, ip string) error {
	now := time.Now()
	return ds.db.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"last_login_at": &now,
		"last_login_ip": ip,
		"updated_at":    now,
	}).Error
}

func (ds *UserRepository) IncrementFailedAttempts(userID string) error {
	return ds.db.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"failed_attempts": gorm.Expr("failed_attempts + 1"),
		"updated_at":      time.Now(),
	}).Error
}

func (ds *UserRepository) ResetFailedAttempts(userID string) error {
	return ds.db.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"failed_attempts": 0,
		"locked_until":    nil,
		"updated_at":      time.Now(),
	}).Error
}

func (ds *UserRepository) LockAccount(userID string, lockUntil time.Time) error {
	return ds.db.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"locked_until": &lockUntil,
		"updated_at":   time.Now(),
	}).Error
}

func (ds *UserRepository) VerifyUserEmail(userID string) error {
	return ds.db.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"email_verified":           true,
		"verification_code":        nil,
		"verification_code_expiry": nil,
		"updated_at":               time.Now(),
	}).Error
}

func (ds *UserRepository) UpdateVerificationCode(userID, code string) error {
	codeExpiry := time.Now().Add(15 * time.Minute) // Code expires in 15 minutes
	return ds.db.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"verification_code":        code,
		"verification_code_expiry": &codeExpiry,
		"updated_at":               time.Now(),
	}).Error
}

func (ds *UserRepository) IsUsernameAvailable(username string) (bool, error) {
	var count int64
	err := ds.db.Model(&model.User{}).Where("LOWER(username) = LOWER(?) AND deleted_at IS NULL", username).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func (ds *UserRepository) IsEmailAvailable(email string) (bool, error) {
	var count int64
	err := ds.db.Model(&model.User{}).Where("LOWER(email) = LOWER(?) AND deleted_at IS NULL", email).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func (ds *UserRepository) CreateUserSession(session dto.UserSession) (string, error) {
	dbSession := &model.UserSession{
		ID:               uuid.New().String(),
		UserID:           session.UserID,
		TokenHash:        session.TokenHash,
		RefreshTokenJTI:  session.RefreshTokenJTI,
		RefreshExpiresAt: session.RefreshExpiresAt,
		DeviceID:         session.DeviceID,
		IP:               session.IP,
		UserAgent:        session.UserAgent,
		CreatedAt:        session.CreatedAt,
		LastUsed:         session.LastUsed,
		IsActive:         session.IsActive,
		ExpiresAt:        session.CreatedAt.Add(7 * 24 * time.Hour), // 7 days
	}

	if err := ds.db.Create(dbSession).Error; err != nil {
		return "", err
	}
	return dbSession.ID, nil
}

func (ds *UserRepository) GetActiveSession(userID, tokenHash string) (*model.UserSession, error) {
	var session model.UserSession
	err := ds.db.Where("user_id = ? AND token_hash = ? AND is_active = ? AND expires_at > ?",
		userID, tokenHash, true, time.Now()).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (ds *UserRepository) UpdateSessionLastUsed(sessionID string) error {
	return ds.db.Model(&model.UserSession{}).Where("id = ?", sessionID).Update("last_used", time.Now()).Error
}

func (ds *UserRepository) UpdateSessionToken(sessionID, newTokenHash string) error {
	return ds.db.Model(&model.UserSession{}).Where("id = ?", sessionID).Updates(map[string]interface{}{
		"token_hash": newTokenHash,
		"last_used":  time.Now(),
	}).Error
}

func (ds *UserRepository) DeactivateSession(sessionID, userID string) error {
	return ds.db.Model(&model.UserSession{}).Where("id = ? AND user_id = ?", sessionID, userID).Updates(map[string]interface{}{
		"is_active": false,
		"last_used": time.Now(),
	}).Error
}

func (ds *UserRepository) DeactivateAllUserSessions(userID, exceptSessionID string) error {
	query := ds.db.Model(&model.UserSession{}).Where("user_id = ?", userID)
	if exceptSessionID != "" {
		query = query.Where("id != ?", exceptSessionID)
	}

	return query.Updates(map[string]interface{}{
		"is_active": false,
		"last_used": time.Now(),
	}).Error
}

func (ds *UserRepository) GetSessionByID(sessionID string) (*model.UserSession, error) {
	var session model.UserSession
	err := ds.db.Where("id = ?", sessionID).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (ds *UserRepository) GetUserSessions(userID string) ([]model.UserSession, error) {
	var sessions []model.UserSession
	err := ds.db.Where("user_id = ? AND is_active = ?", userID, true).
		Order("last_used DESC").Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func (ds *UserRepository) GetUserActiveSessions(userID string) ([]model.UserSession, error) {
	var sessions []model.UserSession
	err := ds.db.Where("user_id = ? AND is_active = ? AND expires_at > ?", userID, true, time.Now()).
		Order("last_used DESC").Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func (ds *UserRepository) CleanupExpiredSessions() error {
	return ds.db.Model(&model.UserSession{}).
		Where("expires_at < ?", time.Now()).
		Update("is_active", false).Error
}

// ==================== PASSWORD RESET METHODS ====================

func (ds *UserRepository) CreatePasswordResetCode(userID, code string, expiresAt time.Time) error {
	resetToken := &model.PasswordResetCode{
		ID:        uuid.New().String(),
		UserID:    userID,
		Code:      code,
		ExpiresAt: expiresAt,
		Used:      false,
		CreatedAt: time.Now(),
	}

	return ds.db.Create(resetToken).Error
}

func (ds *UserRepository) GetPasswordResetCode(code string) (*model.PasswordResetCode, error) {
	var resetCode model.PasswordResetCode
	err := ds.db.Where("code = ? AND used = ?", code, false).First(&resetCode).Error
	if err != nil {
		return nil, err
	}
	return &resetCode, nil
}

func (ds *UserRepository) InvalidatePasswordResetCode(code string) error {
	return ds.db.Model(&model.PasswordResetCode{}).Where("code = ?", code).Update("used", true).Error
}

func (ds *UserRepository) CleanupExpiredPasswordCodes() error {
	return ds.db.Where("expires_at < ?", time.Now()).Delete(&model.PasswordResetCode{}).Error
}

// ==================== TOKEN BLACKLIST METHODS ====================

func (ds *UserRepository) BlacklistToken(jti string, expiresAt time.Time) error {
	blacklistedToken := &model.BlacklistedToken{
		JTI:       jti,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
	if err := ds.db.Create(blacklistedToken).Error; err != nil {
		log.WithError(err).Errorf("Failed to persist blacklisted token to DB: %s", jti)
		return err
	}

	return nil
}

func (ds *UserRepository) IsTokenBlacklisted(jti string) bool {
	var count int64
	ds.db.Model(&model.BlacklistedToken{}).Where("jti = ? AND expires_at > ?", jti, time.Now()).Count(&count)
	return count > 0
}

func (ds *UserRepository) CleanupExpiredBlacklistedTokens() error {
	return ds.db.Where("expires_at < ?", time.Now()).Delete(&model.BlacklistedToken{}).Error
}

// ==================== AUDIT LOG METHODS ====================

func (ds *UserRepository) CreateAuthAuditLog(log dto.AuthAuditLog) error {
	auditLog := &model.AuthAuditLog{
		ID:        uuid.New().String(),
		Action:    log.Action,
		IP:        log.IP,
		UserAgent: log.UserAgent,
		Timestamp: log.Timestamp,
		Success:   log.Success,
		Details:   log.Details,
	}

	if log.UserID != "" {
		auditLog.UserID = log.UserID
	}

	return ds.db.Create(auditLog).Error
}

func (ds *UserRepository) GetUserAuditLogs(userID string, page, limit int) ([]model.AuthAuditLog, int64, error) {
	var logs []model.AuthAuditLog
	var total int64

	// Get total count
	ds.db.Model(&model.AuthAuditLog{}).Where("user_id = ?", userID).Count(&total)

	// Get paginated results
	offset := (page - 1) * limit
	err := ds.db.Where("user_id = ?", userID).
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error

	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (ds *UserRepository) GetAuditLogs(page, limit int, userID, action string) ([]model.AuthAuditLog, int64, error) {
	var logs []model.AuthAuditLog
	var total int64

	query := ds.db.Model(&model.AuthAuditLog{})

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}

	// Get total count
	query.Count(&total)

	// Get paginated results
	offset := (page - 1) * limit
	err := query.Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error

	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (ds *UserRepository) CleanupOldAuditLogs(olderThan time.Time) error {
	return ds.db.Where("timestamp < ?", olderThan).Delete(&model.AuthAuditLog{}).Error
}

// ==================== TRUSTED DEVICE METHODS ====================

func (ds *UserRepository) CreateTrustedDevice(device *model.TrustedDevice) error {
	device.ID = uuid.New().String()
	device.CreatedAt = time.Now()
	device.LastUsed = time.Now()

	return ds.db.Create(device).Error
}

func (ds *UserRepository) GetTrustedDevice(userID, deviceID string) (*model.TrustedDevice, error) {
	var device model.TrustedDevice
	err := ds.db.Where("user_id = ? AND device_id = ?", userID, deviceID).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (ds *UserRepository) UpdateTrustedDevice(device *model.TrustedDevice) error {
	device.LastUsed = time.Now()
	return ds.db.Save(device).Error
}

func (ds *UserRepository) GetUserTrustedDevices(userID string) ([]model.TrustedDevice, error) {
	var devices []model.TrustedDevice
	err := ds.db.Where("user_id = ?", userID).Order("last_used DESC").Find(&devices).Error
	if err != nil {
		return nil, err
	}
	return devices, nil
}

func (ds *UserRepository) RemoveTrustedDevice(userID, deviceID string) error {
	return ds.db.Where("user_id = ? AND device_id = ?", userID, deviceID).Delete(&model.TrustedDevice{}).Error
}

// ==================== LOGIN ATTEMPT METHODS ====================

func (ds *UserRepository) RecordLoginAttempt(ip, email, userAgent string, success bool) error {
	attempt := &model.LoginAttempt{
		ID:        uuid.New().String(),
		IP:        ip,
		Email:     email,
		Success:   success,
		Timestamp: time.Now(),
		UserAgent: userAgent,
	}

	return ds.db.Create(attempt).Error
}

func (ds *UserRepository) GetRecentLoginAttempts(ip string, since time.Time) ([]model.LoginAttempt, error) {
	var attempts []model.LoginAttempt
	err := ds.db.Where("ip = ? AND timestamp > ?", ip, since).
		Order("timestamp DESC").Find(&attempts).Error
	if err != nil {
		return nil, err
	}
	return attempts, nil
}

func (ds *UserRepository) CleanupOldLoginAttempts(olderThan time.Time) error {
	return ds.db.Where("timestamp < ?", olderThan).Delete(&model.LoginAttempt{}).Error
}

// ==================== ADMIN USER MANAGEMENT ====================

func (ds *UserRepository) AdminGetUsers(page, limit int, search string) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := ds.db.Model(&model.User{}).Where("deleted_at IS NULL")

	if search != "" {
		searchPattern := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(username) LIKE ? OR LOWER(email) LIKE ?", searchPattern, searchPattern)
	}

	// Get total count
	query.Count(&total)

	// Get paginated results
	offset := (page - 1) * limit
	err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error

	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (ds *UserRepository) AdminUpdateUser(userID string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return ds.db.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error
}

func (ds *UserRepository) AdminDeleteUser(userID string) error {
	now := time.Now()
	return ds.db.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"deleted_at": &now,
		"is_active":  false,
		"updated_at": now,
	}).Error
}

// ==================== USER PROFILE & SECURITY METHODS ====================

func (ds *UserRepository) GetUserProfile(userID string) (*model.User, error) {
	var user model.User
	err := ds.db.Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (ds *UserRepository) UpdateUserProfile(userID string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return ds.db.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error
}

func (ds *UserRepository) GetSecuritySettings(userID string) (*dto.SecuritySettings, error) {
	var user model.User
	err := ds.db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}

	settings := &dto.SecuritySettings{
		TwoFactorEnabled:     user.TwoFactorEnabled,
		BackupCodesGenerated: user.BackupCodes != "",
		LastPasswordChange:   user.LastPasswordChange,
		LoginNotifications:   user.LoginNotifications,
		SessionTimeout:       user.SessionTimeout,
	}

	return settings, nil
}

func (ds *UserRepository) UpdateSecuritySettings(userID string, settings dto.UpdateSecuritySettingsRequest) error {
	updates := make(map[string]interface{})
	updates["updated_at"] = time.Now()

	if settings.LoginNotifications != nil {
		updates["login_notifications"] = *settings.LoginNotifications
	}
	if settings.SessionTimeout != nil {
		updates["session_timeout"] = *settings.SessionTimeout
	}

	return ds.db.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error
}

// ==================== CLEANUP AND MAINTENANCE ====================

func (ds *UserRepository) CleanupExpiredData() error {
	now := time.Now()

	ds.CleanupExpiredSessions()

	ds.CleanupExpiredPasswordCodes()

	ds.CleanupExpiredBlacklistedTokens()

	ds.CleanupOldLoginAttempts(now.Add(-30 * 24 * time.Hour))

	ds.CleanupOldAuditLogs(now.Add(-90 * 24 * time.Hour))

	return nil
}

// ==================== STATISTICS METHODS ====================

func (ds *UserRepository) GetUserStats(userID string) (*dto.UserStats, error) {
	var user model.User
	err := ds.db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}

	// Count active sessions
	var sessionCount int64
	ds.db.Model(&model.UserSession{}).Where("user_id = ? AND is_active = ? AND expires_at > ?",
		userID, true, time.Now()).Count(&sessionCount)

	// Count total logins from audit logs
	var loginCount int64
	ds.db.Model(&model.AuthAuditLog{}).Where("user_id = ? AND action = ? AND success = ?",
		userID, model.ActionLogin, true).Count(&loginCount)

	stats := &dto.UserStats{
		TotalLogins:        int(loginCount),
		FailedAttempts:     user.FailedAttempts,
		ActiveSessions:     int(sessionCount),
		LastPasswordChange: user.LastPasswordChange,
	}

	return stats, nil
}

func (ds *UserRepository) SeedInitialData() error {
	err := ds.createDefaultAdmin()
	if err != nil {
		return err
	}

	return nil
}

// Create default admin user
func (ds *UserRepository) createDefaultAdmin() error {
	var count int64
	ds.db.Model(&model.User{}).Where("role = ?", model.RoleAdmin).Count(&count)

	if count == 0 {
		// Hash default password (CHANGE THIS IN PRODUCTION!)
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), 12)
		if err != nil {
			return err
		}

		admin := &model.User{
			ID:                 "admin-" + time.Now().Format("20060102150405"),
			Username:           "admin",
			Email:              "admin@techyouth.com",
			Password:           string(hashedPassword),
			Role:               model.RoleAdmin,
			IsActive:           true,
			EmailVerified:      true,
			LoginNotifications: true,
			SessionTimeout:     1440,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}

		err = ds.db.Create(admin).Error
		if err != nil {
			log.Printf("Failed to create admin user: %v", err)
			return err
		}

		log.Println("Default admin user created - Username: admin, Password: admin123 (CHANGE THIS!)")
	}

	return nil
}
