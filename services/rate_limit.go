package services

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alphabatem/common/context"
	"github.com/gofiber/fiber/v2"
	"github.com/lac-hong-legacy/TechYouth-Be/dto"
	"github.com/lac-hong-legacy/TechYouth-Be/model"
	"github.com/lac-hong-legacy/TechYouth-Be/shared"
	log "github.com/sirupsen/logrus"
)

type RateLimitService struct {
	context.DefaultService

	configs map[string]*RateLimitConfig
	mutex   sync.RWMutex

	sqlSvc *SqliteService
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	EndpointType string
	MaxRequests  int
	WindowSize   time.Duration
	BlockTime    time.Duration
	Description  string
	IsActive     bool
}

const RATE_LIMIT_SVC = "rate_limit_svc"

func (svc RateLimitService) Id() string {
	return RATE_LIMIT_SVC
}

func (svc *RateLimitService) Configure(ctx *context.Context) error {
	svc.configs = make(map[string]*RateLimitConfig)
	return svc.DefaultService.Configure(ctx)
}

func (svc *RateLimitService) Start() error {
	svc.sqlSvc = svc.Service(SQLITE_SVC).(*SqliteService)
	svc.initDefaultConfigs()

	// Start background cleanup job
	go svc.startCleanupJob()

	return nil
}

// ==================== CONFIGURATION MANAGEMENT ====================

func (svc *RateLimitService) initDefaultConfigs() {
	svc.mutex.Lock()
	defer svc.mutex.Unlock()

	svc.configs = map[string]*RateLimitConfig{
		// Authentication endpoints
		"login": {
			EndpointType: "login",
			MaxRequests:  10,
			WindowSize:   15 * time.Minute,
			BlockTime:    30 * time.Minute,
			Description:  "Login attempts rate limit",
			IsActive:     true,
		},
		"register": {
			EndpointType: "register",
			MaxRequests:  5,
			WindowSize:   15 * time.Minute,
			BlockTime:    60 * time.Minute,
			Description:  "Registration rate limit",
			IsActive:     true,
		},
		"forgot_password": {
			EndpointType: "forgot_password",
			MaxRequests:  3,
			WindowSize:   15 * time.Minute,
			BlockTime:    60 * time.Minute,
			Description:  "Password reset request rate limit",
			IsActive:     true,
		},
		"reset_password": {
			EndpointType: "reset_password",
			MaxRequests:  5,
			WindowSize:   15 * time.Minute,
			BlockTime:    30 * time.Minute,
			Description:  "Password reset rate limit",
			IsActive:     true,
		},
		"refresh": {
			EndpointType: "refresh",
			MaxRequests:  20,
			WindowSize:   15 * time.Minute,
			BlockTime:    5 * time.Minute,
			Description:  "Token refresh rate limit",
			IsActive:     true,
		},
		"resend_verification": {
			EndpointType: "resend_verification",
			MaxRequests:  3,
			WindowSize:   5 * time.Minute,
			BlockTime:    30 * time.Minute,
			Description:  "Resend verification email rate limit",
			IsActive:     true,
		},

		// Game endpoints
		"guest_session": {
			EndpointType: "guest_session",
			MaxRequests:  5,
			WindowSize:   15 * time.Minute,
			BlockTime:    30 * time.Minute,
			Description:  "Guest session creation rate limit",
			IsActive:     true,
		},
		"lesson_complete": {
			EndpointType: "lesson_complete",
			MaxRequests:  20,
			WindowSize:   time.Hour,
			BlockTime:    2 * time.Hour,
			Description:  "Lesson completion rate limit",
			IsActive:     true,
		},
		"hearts_from_ad": {
			EndpointType: "hearts_from_ad",
			MaxRequests:  20,
			WindowSize:   time.Hour,
			BlockTime:    6 * time.Hour,
			Description:  "Hearts from ads rate limit",
			IsActive:     true,
		},

		// API endpoints
		"api_general": {
			EndpointType: "api_general",
			MaxRequests:  1000,
			WindowSize:   time.Hour,
			BlockTime:    time.Hour,
			Description:  "General API rate limit per IP",
			IsActive:     true,
		},
		"api_strict": {
			EndpointType: "api_strict",
			MaxRequests:  100,
			WindowSize:   10 * time.Minute,
			BlockTime:    24 * time.Hour,
			Description:  "Strict rate limit for abuse prevention",
			IsActive:     true,
		},

		// User-specific endpoints
		"change_password": {
			EndpointType: "change_password",
			MaxRequests:  3,
			WindowSize:   time.Hour,
			BlockTime:    2 * time.Hour,
			Description:  "Password change rate limit",
			IsActive:     true,
		},
		"profile_update": {
			EndpointType: "profile_update",
			MaxRequests:  10,
			WindowSize:   time.Hour,
			BlockTime:    30 * time.Minute,
			Description:  "Profile update rate limit",
			IsActive:     true,
		},
		"username_check": {
			EndpointType: "username_check",
			MaxRequests:  50,
			WindowSize:   time.Hour,
			BlockTime:    10 * time.Minute,
			Description:  "Username availability check rate limit",
			IsActive:     true,
		},
	}
}

// ==================== CORE RATE LIMITING LOGIC ====================

func (svc *RateLimitService) IsAllowed(identifier, endpointType string) (bool, *dto.RateLimitInfo, error) {
	svc.mutex.RLock()
	config, exists := svc.configs[endpointType]
	svc.mutex.RUnlock()

	if !exists || !config.IsActive {
		// If no config exists or inactive, allow the request
		return true, &dto.RateLimitInfo{
			Allowed:   true,
			Remaining: -1,
		}, nil
	}

	now := time.Now()
	windowStart := now.Add(-config.WindowSize)

	// Get current rate limit record
	rateLimit, err := svc.sqlSvc.GetRateLimit(identifier, endpointType)
	if err != nil {
		return false, nil, err
	}

	// Check if currently blocked
	if rateLimit != nil && rateLimit.BlockedUntil != nil && now.Before(*rateLimit.BlockedUntil) {
		return false, &dto.RateLimitInfo{
			Allowed:      false,
			Remaining:    0,
			ResetTime:    rateLimit.BlockedUntil,
			BlockedUntil: rateLimit.BlockedUntil,
		}, nil
	}

	// If no existing record or window has passed, create/reset
	if rateLimit == nil || rateLimit.WindowStart.Before(windowStart) {
		rateLimit = &model.RateLimit{
			Identifier:   identifier,
			EndpointType: endpointType,
			RequestCount: 1,
			WindowStart:  now,
			BlockedUntil: nil,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		if err := svc.sqlSvc.SaveRateLimit(rateLimit); err != nil {
			return false, nil, err
		}

		resetTime := now.Add(config.WindowSize)
		return true, &dto.RateLimitInfo{
			Allowed:   true,
			Remaining: config.MaxRequests - 1,
			ResetTime: &resetTime,
		}, nil
	}

	// Check if limit exceeded
	if rateLimit.RequestCount >= config.MaxRequests {
		// Block the identifier
		blockedUntil := now.Add(config.BlockTime)
		rateLimit.BlockedUntil = &blockedUntil
		rateLimit.UpdatedAt = now

		if err := svc.sqlSvc.UpdateRateLimit(rateLimit); err != nil {
			return false, nil, err
		}

		return false, &dto.RateLimitInfo{
			Allowed:      false,
			Remaining:    0,
			ResetTime:    &blockedUntil,
			BlockedUntil: &blockedUntil,
		}, nil
	}

	// Increment request count
	rateLimit.RequestCount++
	rateLimit.UpdatedAt = now

	if err := svc.sqlSvc.UpdateRateLimit(rateLimit); err != nil {
		return false, nil, err
	}

	resetTime := rateLimit.WindowStart.Add(config.WindowSize)
	return true, &dto.RateLimitInfo{
		Allowed:   true,
		Remaining: config.MaxRequests - rateLimit.RequestCount,
		ResetTime: &resetTime,
	}, nil
}

// ==================== MIDDLEWARE FUNCTIONS ====================

// RateLimit creates a rate limiting middleware for specific endpoint types
func (svc *RateLimitService) RateLimit(endpointType string, maxRequests int, window string) fiber.Handler {
	// Parse window duration
	windowDuration, err := time.ParseDuration(window)
	if err != nil {
		log.Printf("Invalid window duration for %s: %s", endpointType, window)
		windowDuration = 15 * time.Minute // default fallback
	}

	// Override config if provided
	if maxRequests > 0 {
		svc.mutex.Lock()
		if config, exists := svc.configs[endpointType]; exists {
			config.MaxRequests = maxRequests
			config.WindowSize = windowDuration
		}
		svc.mutex.Unlock()
	}

	return func(c *fiber.Ctx) error {
		identifier := svc.getIdentifier(c, endpointType)

		allowed, info, err := svc.IsAllowed(identifier, endpointType)
		if err != nil {
			log.Printf("Rate limit check error for %s (%s): %v", endpointType, identifier, err)
			// Continue with request on error to avoid blocking users due to system issues
			return c.Next()
		}

		// Add rate limit headers
		svc.addRateLimitHeaders(c, info)

		if !allowed {
			return svc.handleRateLimitExceeded(c, endpointType, info)
		}

		return c.Next()
	}
}

// IPRateLimit applies general rate limiting by IP address
func (svc *RateLimitService) IPRateLimit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := getClientIP(c)

		allowed, info, err := svc.IsAllowed(ip, "api_general")
		if err != nil {
			log.Printf("IP rate limit check error for %s: %v", ip, err)
			return c.Next()
		}

		svc.addRateLimitHeaders(c, info)

		if !allowed {
			return svc.handleRateLimitExceeded(c, "api_general", info)
		}

		return c.Next()
	}
}

// StrictRateLimit applies strict rate limiting for sensitive endpoints
func (svc *RateLimitService) StrictRateLimit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := getClientIP(c)

		allowed, info, err := svc.IsAllowed(ip, "api_strict")
		if err != nil {
			log.Printf("Strict rate limit check error for %s: %v", ip, err)
			return shared.ResponseJSON(c, http.StatusInternalServerError, "Rate limit service unavailable", nil)
		}

		if !allowed {
			return svc.handleRateLimitExceeded(c, "api_strict", info)
		}

		return c.Next()
	}
}

// UserBasedRateLimit applies rate limiting based on authenticated user
func (svc *RateLimitService) UserBasedRateLimit(endpointType string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals(shared.UserID)
		userIDStr := ""
		if userID != nil {
			userIDStr = userID.(string)
		}
		if userIDStr == "" {
			// Fall back to IP if user not authenticated
			userIDStr = getClientIP(c)
		}

		allowed, info, err := svc.IsAllowed(userIDStr, endpointType)
		if err != nil {
			log.Printf("User rate limit check error for %s (%s): %v", endpointType, userIDStr, err)
			return c.Next()
		}

		svc.addRateLimitHeaders(c, info)

		if !allowed {
			return svc.handleRateLimitExceeded(c, endpointType, info)
		}

		return c.Next()
	}
}

// ==================== HELPER FUNCTIONS ====================

func (svc *RateLimitService) getIdentifier(c *fiber.Ctx, endpointType string) string {
	switch endpointType {
	case "login", "register", "forgot_password", "reset_password", "resend_verification":
		// For auth endpoints, use IP + email if available
		email := svc.getEmailFromRequest(c)
		if email != "" {
			return fmt.Sprintf("%s:%s", getClientIP(c), email)
		}
		return getClientIP(c)

	case "guest_session":
		// For guest sessions, prefer device ID
		deviceID := getDeviceIDFromRequest(c)
		if deviceID != "" {
			return deviceID
		}
		return getClientIP(c)

	case "lesson_complete", "hearts_from_ad":
		// For game actions, use session ID or user ID
		sessionID := c.Params("sessionId")
		if sessionID != "" {
			return sessionID
		}
		userID := c.Locals(shared.UserID)
		if userID != nil {
			if userIDStr, ok := userID.(string); ok && userIDStr != "" {
				return userIDStr
			}
		}
		return getClientIP(c)

	case "change_password", "profile_update":
		// For user actions, use user ID
		userID := c.Locals(shared.UserID)
		if userID != nil {
			if userIDStr, ok := userID.(string); ok && userIDStr != "" {
				return userIDStr
			}
		}
		return getClientIP(c)

	default:
		// Default to IP address
		return getClientIP(c)
	}
}

func (svc *RateLimitService) getEmailFromRequest(c *fiber.Ctx) string {
	var reqBody map[string]interface{}
	if len(c.Body()) > 0 {
		// Try to parse JSON body
		if err := c.BodyParser(&reqBody); err == nil {
			if email, exists := reqBody["email"]; exists {
				if emailStr, ok := email.(string); ok {
					return emailStr
				}
			}
			if emailOrUsername, exists := reqBody["email_or_username"]; exists {
				if emailStr, ok := emailOrUsername.(string); ok {
					return emailStr
				}
			}
		}
	}
	return ""
}

func (svc *RateLimitService) addRateLimitHeaders(c *fiber.Ctx, info *dto.RateLimitInfo) {
	if info == nil {
		return
	}

	if info.Remaining >= 0 {
		c.Set("X-RateLimit-Remaining", strconv.Itoa(info.Remaining))
	}

	if info.ResetTime != nil {
		c.Set("X-RateLimit-Reset", strconv.FormatInt(info.ResetTime.Unix(), 10))
	}

	if info.BlockedUntil != nil {
		retryAfter := int(time.Until(*info.BlockedUntil).Seconds())
		if retryAfter > 0 {
			c.Set("Retry-After", strconv.Itoa(retryAfter))
		}
	}
}

func (svc *RateLimitService) handleRateLimitExceeded(c *fiber.Ctx, endpointType string, info *dto.RateLimitInfo) error {
	message := svc.getRateLimitMessage(endpointType)

	response := map[string]interface{}{
		"error":   "Rate limit exceeded",
		"message": message,
	}

	if info.BlockedUntil != nil {
		response["blocked_until"] = info.BlockedUntil.Unix()
		response["retry_after"] = int(time.Until(*info.BlockedUntil).Seconds())
	}

	return shared.ResponseJSON(c, http.StatusTooManyRequests, message, response)
}

func (svc *RateLimitService) getRateLimitMessage(endpointType string) string {
	messages := map[string]string{
		"login":               "Too many login attempts. Please try again later.",
		"register":            "Too many registration attempts. Please try again later.",
		"forgot_password":     "Too many password reset requests. Please try again later.",
		"reset_password":      "Too many password reset attempts. Please try again later.",
		"refresh":             "Too many token refresh requests. Please try again later.",
		"resend_verification": "Too many verification email requests. Please try again later.",
		"guest_session":       "Too many session creation attempts. Please try again later.",
		"lesson_complete":     "Too many lesson completions. Please take a break.",
		"hearts_from_ad":      "Too many ad requests. You've reached the hourly limit.",
		"change_password":     "Too many password change attempts. Please try again later.",
		"profile_update":      "Too many profile updates. Please try again later.",
		"username_check":      "Too many username checks. Please try again later.",
		"api_general":         "Too many requests. Please slow down.",
		"api_strict":          "Rate limit exceeded. Access temporarily blocked.",
	}

	if message, exists := messages[endpointType]; exists {
		return message
	}

	return "Too many requests. Please try again later."
}

// ==================== UTILITY FUNCTIONS ====================

func getClientIP(c *fiber.Ctx) string {
	// Check for forwarded IP first (for load balancers/proxies)
	forwarded := c.Get("X-Forwarded-For")
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if ip != "" {
				return ip
			}
		}
	}

	// Check for real IP header
	realIP := c.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Check Cloudflare header
	cfIP := c.Get("CF-Connecting-IP")
	if cfIP != "" {
		return cfIP
	}

	// Fall back to remote address
	ip, _, err := net.SplitHostPort(c.Context().RemoteAddr().String())
	if err != nil {
		return c.Context().RemoteAddr().String()
	}

	return ip
}

func getDeviceIDFromRequest(c *fiber.Ctx) string {
	// Try to get device_id from query params first
	if deviceID := c.Query("device_id"); deviceID != "" {
		return deviceID
	}

	// Try to get from headers
	if deviceID := c.Get("X-Device-ID"); deviceID != "" {
		return deviceID
	}

	// Try to get from JSON body
	var reqBody map[string]interface{}
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&reqBody); err == nil {
			if deviceID, exists := reqBody["device_id"]; exists {
				if deviceIDStr, ok := deviceID.(string); ok {
					return deviceIDStr
				}
			}
		}
	}

	return ""
}

// ==================== ADMIN FUNCTIONS ====================

func (svc *RateLimitService) GetRateLimitStats() fiber.Handler {
	return func(c *fiber.Ctx) error {
		svc.mutex.RLock()
		configs := make(map[string]*RateLimitConfig)
		for k, v := range svc.configs {
			configs[k] = v
		}
		svc.mutex.RUnlock()

		// Get current rate limit records count
		var totalRecords int64
		svc.sqlSvc.Db().Model(&model.RateLimit{}).Count(&totalRecords)

		// Get blocked records count
		var blockedRecords int64
		svc.sqlSvc.Db().Model(&model.RateLimit{}).
			Where("blocked_until > ?", time.Now()).
			Count(&blockedRecords)

		stats := map[string]interface{}{
			"configs":         configs,
			"total_records":   totalRecords,
			"blocked_records": blockedRecords,
			"timestamp":       time.Now(),
		}

		return shared.ResponseJSON(c, http.StatusOK, "Rate limit statistics", stats)
	}
}

func (svc *RateLimitService) CleanupRateLimits() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := svc.CleanupOldRecords(); err != nil {
			return shared.ResponseJSON(c, http.StatusInternalServerError, "Failed to cleanup rate limits", err.Error())
		}
		return shared.ResponseJSON(c, http.StatusOK, "Rate limits cleaned up successfully", nil)
	}
}

func (svc *RateLimitService) RemoveRateLimit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		identifier := c.Params("identifier")
		endpointType := c.Params("endpointType")

		if identifier == "" || endpointType == "" {
			return shared.ResponseJSON(c, http.StatusBadRequest, "Missing identifier or endpoint type", nil)
		}

		err := svc.sqlSvc.Db().Where("identifier = ? AND endpoint_type = ?", identifier, endpointType).
			Delete(&model.RateLimit{}).Error

		if err != nil {
			return shared.ResponseJSON(c, http.StatusInternalServerError, "Failed to remove rate limit", err.Error())
		}

		message := fmt.Sprintf("Rate limit removed for %s/%s", identifier, endpointType)
		return shared.ResponseJSON(c, http.StatusOK, message, nil)
	}
}

func (svc *RateLimitService) UpdateConfig() fiber.Handler {
	return func(c *fiber.Ctx) error {
		endpointType := c.Params("endpointType")

		var req struct {
			MaxRequests int    `json:"max_requests"`
			WindowSize  string `json:"window_size"` // e.g., "15m", "1h"
			BlockTime   string `json:"block_time"`  // e.g., "30m", "2h"
			IsActive    *bool  `json:"is_active"`
		}

		if err := c.BodyParser(&req); err != nil {
			return shared.ResponseJSON(c, http.StatusBadRequest, "Invalid request body", err.Error())
		}

		svc.mutex.Lock()
		config, exists := svc.configs[endpointType]
		if !exists {
			svc.mutex.Unlock()
			return shared.ResponseJSON(c, http.StatusNotFound, "Endpoint type not found", nil)
		}

		// Update configuration
		if req.MaxRequests > 0 {
			config.MaxRequests = req.MaxRequests
		}

		if req.WindowSize != "" {
			if duration, err := time.ParseDuration(req.WindowSize); err == nil {
				config.WindowSize = duration
			}
		}

		if req.BlockTime != "" {
			if duration, err := time.ParseDuration(req.BlockTime); err == nil {
				config.BlockTime = duration
			}
		}

		if req.IsActive != nil {
			config.IsActive = *req.IsActive
		}

		svc.mutex.Unlock()

		return shared.ResponseJSON(c, http.StatusOK, "Configuration updated successfully", config)
	}
}

// ==================== BACKGROUND JOBS ====================

func (svc *RateLimitService) CleanupOldRecords() error {
	return svc.sqlSvc.CleanupOldRecords()
}

func (svc *RateLimitService) startCleanupJob() {
	ticker := time.NewTicker(1 * time.Hour) // Run every hour
	defer ticker.Stop()

	for range ticker.C {
		if err := svc.CleanupOldRecords(); err != nil {
			log.Printf("Rate limit cleanup error: %v", err)
		} else {
			log.Printf("Rate limit cleanup completed successfully")
		}
	}
}

// ==================== PUBLIC METHODS ====================

func (svc *RateLimitService) IsBlocked(identifier, endpointType string) bool {
	allowed, _, err := svc.IsAllowed(identifier, endpointType)
	if err != nil {
		log.Printf("Error checking rate limit status: %v", err)
		return false // Don't block on error
	}
	return !allowed
}

func (svc *RateLimitService) GetRemainingRequests(identifier, endpointType string) int {
	_, info, err := svc.IsAllowed(identifier, endpointType)
	if err != nil || info == nil {
		return -1 // Unknown
	}
	return info.Remaining
}

func (svc *RateLimitService) ResetRateLimit(identifier, endpointType string) error {
	return svc.sqlSvc.Db().Where("identifier = ? AND endpoint_type = ?", identifier, endpointType).
		Delete(&model.RateLimit{}).Error
}
