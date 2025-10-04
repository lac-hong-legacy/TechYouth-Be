package model

import "time"

const (
	RoleAdmin            = "admin"
	RoleUser             = "user"
	RoleMod              = "mod"
	ActionLogin          = "login"
	ActionLogout         = "logout"
	ActionRegister       = "register"
	ActionForgotPassword = "forgot_password"
	ActionResetPassword  = "reset_password"
	ActionVerifyEmail    = "verify_email"
	ActionUpdateProfile  = "update_profile"
	ActionUpdatePassword = "update_password"
)

type User struct {
	// Basic Information
	ID        string `json:"id" gorm:"primaryKey;type:text;not null"`
	Username  string `json:"username" gorm:"uniqueIndex:idx_username;not null;size:50"`
	Email     string `json:"email" gorm:"uniqueIndex:idx_email;not null;size:255"`
	BirthYear int    `json:"birth_year" gorm:"default:0;not null"`
	Password  string `json:"-" gorm:"not null;size:255"` // Never expose in JSON

	// Role and Status
	Role     string `json:"role" gorm:"default:user;not null;size:20;index"`
	IsActive bool   `json:"is_active" gorm:"default:true;not null;index"`

	// Email Verification
	EmailVerified          bool       `json:"email_verified" gorm:"default:false;not null;index"`
	VerificationCode       string     `json:"-" gorm:"size:6;index"`
	VerificationCodeExpiry *time.Time `json:"-" gorm:"index"`

	// Security Fields
	FailedAttempts     int        `json:"failed_attempts" gorm:"default:0;not null"`
	LockedUntil        *time.Time `json:"locked_until,omitempty" gorm:"index"`
	LastLoginAt        *time.Time `json:"last_login_at,omitempty"`
	LastLoginIP        string     `json:"last_login_ip,omitempty" gorm:"size:45"`
	LastPasswordChange *time.Time `json:"last_password_change,omitempty"`

	// Two-Factor Authentication
	TwoFactorEnabled bool   `json:"two_factor_enabled" gorm:"default:false;not null"`
	TwoFactorSecret  string `json:"-" gorm:"size:255"`
	BackupCodes      string `json:"-" gorm:"type:text"` // JSON array stored as string

	// User Preferences
	LoginNotifications bool `json:"login_notifications" gorm:"default:true;not null"`
	SessionTimeout     int  `json:"session_timeout" gorm:"default:1440;not null"` // minutes, default 24h

	// Timestamps
	CreatedAt time.Time  `json:"created_at" gorm:"not null;index"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"not null"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

// UserSession represents an active user session
type UserSession struct {
	ID               string    `json:"id" gorm:"primaryKey;type:text;not null"`
	UserID           string    `json:"user_id" gorm:"not null;index;size:50"`
	TokenHash        string    `json:"token_hash" gorm:"not null;index;size:255"`
	RefreshTokenJTI  string    `json:"refresh_token_jti" gorm:"index;size:255"` // Nullable for existing sessions
	RefreshExpiresAt time.Time `json:"refresh_expires_at" gorm:"not null"`
	DeviceID         string    `json:"device_id,omitempty" gorm:"index;size:100"`
	IP               string    `json:"ip" gorm:"not null;size:45"`
	UserAgent        string    `json:"user_agent" gorm:"type:text"`
	CreatedAt        time.Time `json:"created_at" gorm:"not null"`
	LastUsed         time.Time `json:"last_used" gorm:"not null;index"`
	IsActive         bool      `json:"is_active" gorm:"default:true;not null;index"`
	ExpiresAt        time.Time `json:"expires_at" gorm:"not null;index"`

	// Relationships
	User User `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// AuthAuditLog represents authentication audit logs
type AuthAuditLog struct {
	ID        string    `json:"id" gorm:"primaryKey;type:text;not null"`
	UserID    string    `json:"user_id,omitempty" gorm:"index;size:50"`
	Action    string    `json:"action" gorm:"not null;index;size:50"`
	IP        string    `json:"ip,omitempty" gorm:"index;size:45"`
	UserAgent string    `json:"user_agent,omitempty" gorm:"type:text"`
	Timestamp time.Time `json:"timestamp" gorm:"not null;index"`
	Success   bool      `json:"success" gorm:"not null;index"`
	Details   string    `json:"details,omitempty" gorm:"type:text"`

	// Relationships
	User *User `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL"`
}

// PasswordResetCode represents password reset codes
type PasswordResetCode struct {
	ID        string    `json:"id" gorm:"primaryKey;type:text;not null"`
	UserID    string    `json:"user_id" gorm:"not null;index;size:50"`
	Code      string    `json:"code" gorm:"not null;uniqueIndex;size:255"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null;index"`
	Used      bool      `json:"used" gorm:"default:false;not null;index"`
	CreatedAt time.Time `json:"created_at" gorm:"not null"`

	// Relationships
	User User `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// BlacklistedToken represents blacklisted JWT tokens
type BlacklistedToken struct {
	JTI       string    `json:"jti" gorm:"primaryKey;size:255"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null;index"`
	CreatedAt time.Time `json:"created_at" gorm:"not null"`
}

// TrustedDevice represents trusted devices for users
type TrustedDevice struct {
	ID        string    `json:"id" gorm:"primaryKey;type:text;not null"`
	UserID    string    `json:"user_id" gorm:"not null;index;size:50"`
	DeviceID  string    `json:"device_id" gorm:"not null;size:100"`
	Name      string    `json:"name" gorm:"size:255"`
	Type      string    `json:"type" gorm:"size:20"` // mobile, desktop, tablet
	OS        string    `json:"os" gorm:"size:50"`
	Browser   string    `json:"browser" gorm:"size:50"`
	IP        string    `json:"ip" gorm:"size:45"`
	IsTrusted bool      `json:"is_trusted" gorm:"default:false;not null;index"`
	LastUsed  time.Time `json:"last_used" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"not null"`

	// Relationships
	User User `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// LoginAttempt represents login attempts for rate limiting and security monitoring
type LoginAttempt struct {
	ID        string    `json:"id" gorm:"primaryKey;type:text;not null"`
	IP        string    `json:"ip" gorm:"not null;index;size:45"`
	Email     string    `json:"email,omitempty" gorm:"index;size:255"`
	Success   bool      `json:"success" gorm:"not null;index"`
	Timestamp time.Time `json:"timestamp" gorm:"not null;index"`
	UserAgent string    `json:"user_agent,omitempty" gorm:"type:text"`
}
