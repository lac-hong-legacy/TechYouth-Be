package dto

import "time"

// ==================== AUTHENTICATION REQUEST DTOs ====================

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Username string `json:"username" validate:"required,min=3,max=30,alphanum" example:"johndoe"`
	Password string `json:"password" validate:"required,strong_password" example:"SecurePass123!"`
}

func (r RegisterRequest) Validate() error {
	return GetValidator().Struct(r)
}

type LoginRequest struct {
	EmailOrUsername string `json:"email_or_username" validate:"required" example:"user@example.com"`
	Password        string `json:"password" validate:"required" example:"SecurePass123!"`
	DeviceID        string `json:"device_id,omitempty" example:"device_12345"`
	RememberMe      bool   `json:"remember_me,omitempty" example:"true"`
}

func (l LoginRequest) Validate() error {
	return GetValidator().Struct(l)
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

func (r RefreshTokenRequest) Validate() error {
	return GetValidator().Struct(r)
}

type LogoutRequest struct {
	SessionID string `json:"session_id" validate:"required" example:"sess_123456789"`
}

func (l LogoutRequest) Validate() error {
	return GetValidator().Struct(l)
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required" example:"OldPass123!"`
	NewPassword     string `json:"new_password" validate:"required,strong_password" example:"NewPass123!"`
}

func (c ChangePasswordRequest) Validate() error {
	return GetValidator().Struct(c)
}

type ResetPasswordRequest struct {
	Code        string `json:"code" validate:"required,len=6,numeric" example:"123456"`
	NewPassword string `json:"new_password" validate:"required,strong_password" example:"NewPass123!"`
}

func (r ResetPasswordRequest) Validate() error {
	return GetValidator().Struct(r)
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email" example:"user@example.com"`
}

func (f ForgotPasswordRequest) Validate() error {
	return GetValidator().Struct(f)
}

type VerifyEmailRequest struct {
	Email string `json:"email" validate:"required,email" example:"user@example.com"`
	Code  string `json:"code" validate:"required,len=6,numeric" example:"123456"`
}

func (v VerifyEmailRequest) Validate() error {
	return GetValidator().Struct(v)
}

type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email" example:"user@example.com"`
}

func (r ResendVerificationRequest) Validate() error {
	return GetValidator().Struct(r)
}

// ==================== AUTHENTICATION RESPONSE DTOs ====================

type RegisterResponse struct {
	UserID               string `json:"user_id" example:"usr_123456789"`
	RequiresVerification bool   `json:"requires_verification" example:"true"`
	Message              string `json:"message" example:"Registration successful. Please check your email for verification."`
}

type LoginResponse struct {
	AccessToken  string   `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string   `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresIn    int64    `json:"expires_in" example:"900"`
	SessionID    string   `json:"session_id" example:"sess_123456789"`
	User         UserInfo `json:"user"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresIn    int64  `json:"expires_in" example:"900"`
}

type UserInfo struct {
	ID            string     `json:"id" example:"usr_123456789"`
	Username      string     `json:"username" example:"johndoe"`
	Email         string     `json:"email" example:"user@example.com"`
	Role          string     `json:"role" example:"user"`
	EmailVerified bool       `json:"email_verified" example:"true"`
	CreatedAt     time.Time  `json:"created_at" example:"2023-01-01T00:00:00Z"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty" example:"2023-01-15T10:30:00Z"`
}

// ==================== USER PROFILE DTOs ====================

type UserProfileResponse struct {
	ID            string     `json:"id" example:"usr_123456789"`
	Username      string     `json:"username" example:"johndoe"`
	Email         string     `json:"email" example:"user@example.com"`
	Role          string     `json:"role" example:"user"`
	EmailVerified bool       `json:"email_verified" example:"true"`
	CreatedAt     time.Time  `json:"created_at" example:"2023-01-01T00:00:00Z"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty" example:"2023-01-15T10:30:00Z"`
	LastLoginIP   string     `json:"last_login_ip,omitempty" example:"192.168.1.1"`
	IsActive      bool       `json:"is_active" example:"true"`
	Stats         UserStats  `json:"stats"`
}

type UserStats struct {
	TotalLogins        int        `json:"total_logins" example:"42"`
	FailedAttempts     int        `json:"failed_attempts" example:"0"`
	ActiveSessions     int        `json:"active_sessions" example:"2"`
	LastPasswordChange *time.Time `json:"last_password_change,omitempty" example:"2023-01-10T15:30:00Z"`
}

type UpdateProfileRequest struct {
	Username string `json:"username,omitempty" validate:"omitempty,min=3,max=30" example:"newusername"`
	Email    string `json:"email,omitempty" validate:"omitempty,email" example:"newemail@example.com"`
}

func (u UpdateProfileRequest) Validate() error {
	return GetValidator().Struct(u)
}

// ==================== SESSION MANAGEMENT DTOs ====================

type UserSession struct {
	ID               string    `json:"id" example:"sess_123456789"`
	UserID           string    `json:"user_id" example:"usr_123456789"`
	TokenHash        string    `json:"token_hash" example:"hash_abc123"`
	RefreshTokenJTI  string    `json:"refresh_token_jti,omitempty" example:"jti_123456789"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at,omitempty" example:"2023-01-22T10:30:00Z"`
	DeviceID         string    `json:"device_id,omitempty" example:"device_12345"`
	IP               string    `json:"ip" example:"192.168.1.1"`
	UserAgent        string    `json:"user_agent" example:"Mozilla/5.0..."`
	CreatedAt        time.Time `json:"created_at" example:"2023-01-15T10:30:00Z"`
	LastUsed         time.Time `json:"last_used" example:"2023-01-15T11:30:00Z"`
	IsActive         bool      `json:"is_active" example:"true"`
}

type SessionListResponse struct {
	Sessions []UserSessionInfo `json:"sessions"`
	Total    int               `json:"total" example:"3"`
}

type UserSessionInfo struct {
	ID        string    `json:"id" example:"sess_123456789"`
	DeviceID  string    `json:"device_id,omitempty" example:"device_12345"`
	IP        string    `json:"ip" example:"192.168.1.1"`
	UserAgent string    `json:"user_agent" example:"Mozilla/5.0..."`
	CreatedAt time.Time `json:"created_at" example:"2023-01-15T10:30:00Z"`
	LastUsed  time.Time `json:"last_used" example:"2023-01-15T11:30:00Z"`
	IsActive  bool      `json:"is_active" example:"true"`
	IsCurrent bool      `json:"is_current" example:"false"`
}

// ==================== AUDIT LOGGING DTOs ====================

type AuthAuditLog struct {
	ID        string    `json:"id" example:"log_123456789"`
	UserID    string    `json:"user_id,omitempty" example:"usr_123456789"`
	Action    string    `json:"action" example:"login"`
	IP        string    `json:"ip,omitempty" example:"192.168.1.1"`
	UserAgent string    `json:"user_agent,omitempty" example:"Mozilla/5.0..."`
	Timestamp time.Time `json:"timestamp" example:"2023-01-15T10:30:00Z"`
	Success   bool      `json:"success" example:"true"`
	Details   string    `json:"details,omitempty" example:"Login successful"`
}

type AuditLogResponse struct {
	Logs  []AuthAuditLog `json:"logs"`
	Total int            `json:"total" example:"150"`
	Page  int            `json:"page" example:"1"`
	Limit int            `json:"limit" example:"20"`
}

// ==================== PASSWORD RESET DTOs ====================

type PasswordResetCode struct {
	ID        string    `json:"id" example:"rst_123456789"`
	UserID    string    `json:"user_id" example:"usr_123456789"`
	Code      string    `json:"code" example:"123456"`
	ExpiresAt time.Time `json:"expires_at" example:"2023-01-15T11:30:00Z"`
	Used      bool      `json:"used" example:"false"`
	CreatedAt time.Time `json:"created_at" example:"2023-01-15T10:30:00Z"`
}

// ==================== SECURITY SETTINGS DTOs ====================

type SecuritySettings struct {
	TwoFactorEnabled     bool       `json:"two_factor_enabled" example:"false"`
	BackupCodesGenerated bool       `json:"backup_codes_generated" example:"false"`
	LastPasswordChange   *time.Time `json:"last_password_change,omitempty" example:"2023-01-10T15:30:00Z"`
	LoginNotifications   bool       `json:"login_notifications" example:"true"`
	SessionTimeout       int        `json:"session_timeout" example:"1440"`
}

type UpdateSecuritySettingsRequest struct {
	LoginNotifications *bool `json:"login_notifications,omitempty" example:"true"`
	SessionTimeout     *int  `json:"session_timeout,omitempty" validate:"omitempty,min=300,max=43200" example:"720"`
}

func (u UpdateSecuritySettingsRequest) Validate() error {
	return GetValidator().Struct(u)
}

// ==================== TWO-FACTOR AUTHENTICATION DTOs ====================

type EnableTwoFactorResponse struct {
	Secret      string   `json:"secret" example:"JBSWY3DPEHPK3PXP"`
	QRCode      string   `json:"qr_code" example:"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA..."`
	BackupCodes []string `json:"backup_codes" example:"[\"123456\",\"789012\"]"`
}

type VerifyTwoFactorRequest struct {
	Code string `json:"code" validate:"required,len=6,numeric" example:"123456"`
}

func (v VerifyTwoFactorRequest) Validate() error {
	return GetValidator().Struct(v)
}

type TwoFactorLoginRequest struct {
	EmailOrUsername string `json:"email_or_username" validate:"required" example:"user@example.com"`
	Password        string `json:"password" validate:"required" example:"SecurePass123!"`
	Code            string `json:"code" validate:"required,len=6,numeric" example:"123456"`
	DeviceID        string `json:"device_id,omitempty" example:"device_12345"`
}

func (t TwoFactorLoginRequest) Validate() error {
	return GetValidator().Struct(t)
}

// ==================== ADMIN USER MANAGEMENT DTOs ====================

type AdminUserListResponse struct {
	Users []AdminUserInfo `json:"users"`
	Total int             `json:"total" example:"100"`
	Page  int             `json:"page" example:"1"`
	Limit int             `json:"limit" example:"20"`
}

type AdminUserInfo struct {
	ID             string     `json:"id" example:"usr_123456789"`
	Username       string     `json:"username" example:"johndoe"`
	Email          string     `json:"email" example:"user@example.com"`
	Role           string     `json:"role" example:"user"`
	EmailVerified  bool       `json:"email_verified" example:"true"`
	IsActive       bool       `json:"is_active" example:"true"`
	CreatedAt      time.Time  `json:"created_at" example:"2023-01-01T00:00:00Z"`
	LastLoginAt    *time.Time `json:"last_login_at,omitempty" example:"2023-01-15T10:30:00Z"`
	FailedAttempts int        `json:"failed_attempts" example:"0"`
	LockedUntil    *time.Time `json:"locked_until,omitempty" example:"2023-01-15T12:00:00Z"`
}

type AdminUpdateUserRequest struct {
	Role     *string `json:"role,omitempty" validate:"omitempty,oneof=user admin moderator" example:"admin"`
	IsActive *bool   `json:"is_active,omitempty" example:"true"`
}

func (a AdminUpdateUserRequest) Validate() error {
	return GetValidator().Struct(a)
}

// ==================== RATE LIMITING DTOs ====================

type RateLimitInfo struct {
	Allowed      bool       `json:"allowed" example:"true"`
	Remaining    int        `json:"remaining" example:"9"`
	ResetTime    *time.Time `json:"reset_time,omitempty" example:"2023-01-15T11:00:00Z"`
	BlockedUntil *time.Time `json:"blocked_until,omitempty" example:"2023-01-15T12:00:00Z"`
}

// ==================== DEVICE MANAGEMENT DTOs ====================

type DeviceInfo struct {
	ID        string    `json:"id" example:"dev_123456789"`
	Name      string    `json:"name" example:"iPhone 14"`
	Type      string    `json:"type" example:"mobile"`
	OS        string    `json:"os" example:"iOS 16.0"`
	Browser   string    `json:"browser" example:"Safari"`
	IP        string    `json:"ip" example:"192.168.1.1"`
	LastUsed  time.Time `json:"last_used" example:"2023-01-15T10:30:00Z"`
	IsTrusted bool      `json:"is_trusted" example:"true"`
}

type TrustDeviceRequest struct {
	DeviceID string `json:"device_id" validate:"required" example:"device_12345"`
	Trust    bool   `json:"trust" example:"true"`
}

func (t TrustDeviceRequest) Validate() error {
	return GetValidator().Struct(t)
}

type DeviceListResponse struct {
	Devices []DeviceInfo `json:"devices"`
	Total   int          `json:"total" example:"5"`
}

// ==================== ERROR RESPONSE DTOs ====================

type ErrorResponse struct {
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"Invalid request"`
	Error   string `json:"error,omitempty" example:"validation failed"`
}

type ValidationError struct {
	Field   string `json:"field" example:"email"`
	Message string `json:"message" example:"invalid email format"`
}

type ValidationErrorResponse struct {
	Code    int               `json:"code" example:"400"`
	Message string            `json:"message" example:"Validation failed"`
	Errors  []ValidationError `json:"errors"`
}

// ==================== API RESPONSE WRAPPERS ====================

type AuthResponse struct {
	Code    int         `json:"code" example:"200"`
	Message string      `json:"message" example:"Success"`
	Data    interface{} `json:"data,omitempty"`
}

type SuccessResponse struct {
	Code    int         `json:"code" example:"200"`
	Message string      `json:"message" example:"Operation successful"`
	Data    interface{} `json:"data,omitempty"`
}

// ==================== BLACKLISTED TOKEN DTOs ====================

type BlacklistedToken struct {
	JTI       string    `json:"jti" example:"jti_123456789"`
	ExpiresAt time.Time `json:"expires_at" example:"2023-01-15T11:30:00Z"`
	CreatedAt time.Time `json:"created_at" example:"2023-01-15T10:30:00Z"`
}

// ==================== USERNAME AVAILABILITY DTOs ====================

type UsernameCheckResponse struct {
	Username  string `json:"username" example:"johndoe"`
	Available bool   `json:"available" example:"true"`
}

// ==================== HEALTH CHECK DTOs ====================

type HealthCheckResponse struct {
	Status    string                 `json:"status" example:"healthy"`
	Timestamp time.Time              `json:"timestamp" example:"2023-01-15T10:30:00Z"`
	Version   string                 `json:"version" example:"2.0.0"`
	Uptime    string                 `json:"uptime" example:"2h30m15s"`
	Details   map[string]interface{} `json:"details"`
}

// ==================== STATISTICS DTOs ====================

type UserStatisticsResponse struct {
	TotalUsers       int64   `json:"total_users" example:"1000"`
	ActiveUsers      int64   `json:"active_users" example:"850"`
	VerifiedUsers    int64   `json:"verified_users" example:"950"`
	AdminUsers       int64   `json:"admin_users" example:"5"`
	VerificationRate float64 `json:"verification_rate" example:"95.0"`
}

type SystemStatisticsResponse struct {
	Users           UserStatisticsResponse `json:"users"`
	ActiveSessions  int64                  `json:"active_sessions" example:"125"`
	RecentLogins24h int64                  `json:"recent_logins_24h" example:"42"`
	FailedLogins24h int64                  `json:"failed_logins_24h" example:"5"`
}

// ==================== SEARCH AND PAGINATION DTOs ====================

type PaginationRequest struct {
	Page  int `json:"page" form:"page" validate:"omitempty,min=1" example:"1"`
	Limit int `json:"limit" form:"limit" validate:"omitempty,min=1,max=100" example:"20"`
}

func (p PaginationRequest) Validate() error {
	return GetValidator().Struct(p)
}

type PaginationResponse struct {
	Page       int   `json:"page" example:"1"`
	Limit      int   `json:"limit" example:"20"`
	Total      int64 `json:"total" example:"100"`
	TotalPages int   `json:"total_pages" example:"5"`
	HasNext    bool  `json:"has_next" example:"true"`
	HasPrev    bool  `json:"has_prev" example:"false"`
}

type SearchUsersRequest struct {
	PaginationRequest
	Query         string `json:"query" form:"query" validate:"omitempty,min=1,max=100" example:"john"`
	Role          string `json:"role" form:"role" validate:"omitempty,oneof=user admin moderator" example:"user"`
	IsActive      *bool  `json:"is_active" form:"is_active" example:"true"`
	EmailVerified *bool  `json:"email_verified" form:"email_verified" example:"true"`
}

func (s SearchUsersRequest) Validate() error {
	return GetValidator().Struct(s)
}

type SearchUsersResponse struct {
	Users      []AdminUserInfo    `json:"users"`
	Pagination PaginationResponse `json:"pagination"`
}
