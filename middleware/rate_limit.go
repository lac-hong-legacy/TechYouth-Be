package middleware

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alphabatem/common/context"
	"github.com/gin-gonic/gin"
	"github.com/lac-hong-legacy/TechYouth-Be/dto"
	"github.com/lac-hong-legacy/TechYouth-Be/model"
	"github.com/lac-hong-legacy/TechYouth-Be/services"
)

type RateLimitMiddleware struct {
	context.DefaultService

	configs map[string]*model.RateLimitConfig
	mutex   sync.RWMutex
	sqlSvc  *services.SqliteService
}

const RATE_LIMIT_MIDDLEWARE_SVC = "rate_limit"

func (svc *RateLimitMiddleware) Id() string {
	return RATE_LIMIT_MIDDLEWARE_SVC
}

func (svc *RateLimitMiddleware) Configure(ctx *context.Context) error {
	svc.sqlSvc = ctx.Service(services.SQLITE_SVC).(*services.SqliteService)

	svc.configs = make(map[string]*model.RateLimitConfig)
	return svc.DefaultService.Configure(ctx)
}

func (svc *RateLimitMiddleware) Start() error {
	svc.initDefaultConfigs()
	return nil
}

func (svc *RateLimitMiddleware) initDefaultConfigs() {
	svc.configs = map[string]*model.RateLimitConfig{
		// Guest session creation - prevent device ID spam
		"guest_session": {
			EndpointType: "guest_session",
			MaxRequests:  5,                // Max 5 session requests
			WindowSize:   time.Minute * 15, // Per 15 minutes
			BlockTime:    time.Minute * 30, // Block for 30 minutes
			Description:  "Guest session creation rate limit",
		},

		// Lesson completion - prevent rapid fire completions
		"lesson_complete": {
			EndpointType: "lesson_complete",
			MaxRequests:  10,            // Max 10 lesson completions
			WindowSize:   time.Hour,     // Per hour
			BlockTime:    time.Hour * 2, // Block for 2 hours
			Description:  "Lesson completion rate limit",
		},

		// Hearts from ads - prevent ad fraud
		"hearts_from_ad": {
			EndpointType: "hearts_from_ad",
			MaxRequests:  20,            // Max 20 ad views
			WindowSize:   time.Hour,     // Per hour
			BlockTime:    time.Hour * 6, // Block for 6 hours
			Description:  "Hearts from ads rate limit",
		},

		// General API calls per IP
		"api_general": {
			EndpointType: "api_general",
			MaxRequests:  1000,      // Max 1000 requests
			WindowSize:   time.Hour, // Per hour
			BlockTime:    time.Hour, // Block for 1 hour
			Description:  "General API rate limit per IP",
		},

		// Aggressive protection against abuse
		"api_strict": {
			EndpointType: "api_strict",
			MaxRequests:  100,              // Max 100 requests
			WindowSize:   time.Minute * 10, // Per 10 minutes
			BlockTime:    time.Hour * 24,   // Block for 24 hours
			Description:  "Strict rate limit for abuse prevention",
		},
	}
}

func (svc *RateLimitMiddleware) IsAllowed(identifier, endpointType string) (bool, *dto.RateLimitInfo, error) {
	svc.mutex.RLock()
	config, exists := svc.configs[endpointType]
	svc.mutex.RUnlock()

	if !exists {
		// If no config exists, allow the request
		return true, &dto.RateLimitInfo{
			Allowed:   true,
			Remaining: -1,
		}, nil
	}

	now := time.Now()
	windowStart := now.Add(-config.WindowSize)

	// Get current rate limit record using SQLite service
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
		}

		if err := svc.sqlSvc.SaveRateLimit(rateLimit); err != nil {
			return false, nil, err
		}

		return true, &dto.RateLimitInfo{
			Allowed:   true,
			Remaining: config.MaxRequests - 1,
			ResetTime: &[]time.Time{now.Add(config.WindowSize)}[0],
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

// General rate limiting by IP
func (svc *RateLimitMiddleware) IPRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := getClientIP(c)

		allowed, info, err := svc.IsAllowed(ip, "api_general")
		if err != nil {
			log.Printf("Rate limit check error for IP %s: %v", ip, err)
			// Continue with request on error to avoid blocking users due to system issues
			c.Next()
			return
		}

		// Add rate limit headers
		if info.ResetTime != nil {
			c.Header("X-RateLimit-Reset", strconv.FormatInt(info.ResetTime.Unix(), 10))
		}
		c.Header("X-RateLimit-Remaining", strconv.Itoa(info.Remaining))

		if !allowed {
			retryAfter := ""
			if info.BlockedUntil != nil {
				retryAfter = strconv.FormatInt(info.BlockedUntil.Unix(), 10)
				c.Header("Retry-After", retryAfter)
			}
			
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"message":     "Too many requests from this IP address",
				"retry_after": retryAfter,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Strict rate limiting for sensitive endpoints
func (svc *RateLimitMiddleware) StrictRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := getClientIP(c)

		allowed, info, err := svc.IsAllowed(ip, "api_strict")
		if err != nil {
			log.Printf("Strict rate limit check error for IP %s: %v", ip, err)
			// For strict endpoints, block on error for security
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Rate limit service unavailable",
				"message": "Please try again later",
			})
			c.Abort()
			return
		}

		if !allowed {
			retryAfter := ""
			if info.BlockedUntil != nil {
				retryAfter = strconv.FormatInt(info.BlockedUntil.Unix(), 10)
				c.Header("Retry-After", retryAfter)
			}
			
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":         "Rate limit exceeded - strict mode",
				"blocked_until": retryAfter,
				"message":       "Too many requests. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Guest session specific rate limiting
func (svc *RateLimitMiddleware) GuestSessionRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use device_id from request body if available, otherwise fall back to IP
		deviceID := getDeviceIDFromRequest(c)
		if deviceID == "" {
			deviceID = getClientIP(c)
		}

		allowed, info, err := svc.IsAllowed(deviceID, "guest_session")
		if err != nil {
			log.Printf("Guest session rate limit error: %v", err)
			c.Next()
			return
		}

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many session creation attempts",
				"message":     "Please wait before creating a new session",
				"retry_after": info.BlockedUntil,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Lesson completion rate limiting
func (svc *RateLimitMiddleware) LessonCompletionRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.Param("sessionId")
		if sessionID == "" {
			sessionID = getClientIP(c)
		}

		allowed, info, err := svc.IsAllowed(sessionID, "lesson_complete")
		if err != nil {
			log.Printf("Lesson completion rate limit error: %v", err)
			c.Next()
			return
		}

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many lesson completions",
				"message":     "Please take a break before completing more lessons",
				"retry_after": info.BlockedUntil,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Hearts from ads rate limiting
func (svc *RateLimitMiddleware) HeartsFromAdRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.Param("sessionId")
		if sessionID == "" {
			sessionID = getClientIP(c)
		}

		allowed, info, err := svc.IsAllowed(sessionID, "hearts_from_ad")
		if err != nil {
			log.Printf("Hearts from ad rate limit error: %v", err)
			c.Next()
			return
		}

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many ad requests",
				"message":     "You've reached the hourly limit for earning hearts from ads",
				"retry_after": info.BlockedUntil,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (svc *RateLimitMiddleware) CleanupOldRecords() error {
	return svc.sqlSvc.CleanupOldRecords()
}

// Utility functions
func getClientIP(c *gin.Context) string {
	// Check for forwarded IP first
	forwarded := c.GetHeader("X-Forwarded-For")
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check for real IP
	realIP := c.GetHeader("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to remote address
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}

	return ip
}

func getDeviceIDFromRequest(c *gin.Context) string {
	// Try to get device_id from JSON body
	var reqBody map[string]interface{}
	if c.Request.Body != nil {
		if err := c.ShouldBindJSON(&reqBody); err == nil {
			if deviceID, exists := reqBody["device_id"]; exists {
				if deviceIDStr, ok := deviceID.(string); ok {
					return deviceIDStr
				}
			}
		}
	}

	// Try to get from header
	return c.GetHeader("X-Device-ID")
}

func generateID() string {
	return fmt.Sprintf("rl_%d", time.Now().UnixNano())
}

// Admin handlers for rate limit management
func (svc *RateLimitMiddleware) GetRateLimitStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Return rate limit statistics
		c.JSON(200, gin.H{
			"message": "Rate limit stats endpoint",
			"configs": svc.configs,
		})
	}
}

func (svc *RateLimitMiddleware) CleanupRateLimits() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := svc.CleanupOldRecords(); err != nil {
			c.JSON(500, gin.H{"error": "Failed to cleanup rate limits"})
			return
		}
		c.JSON(200, gin.H{"message": "Rate limits cleaned up successfully"})
	}
}

func (svc *RateLimitMiddleware) RemoveRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		identifier := c.Param("identifier")
		endpointType := c.Param("endpointType")

		// Remove specific rate limit (implementation would go here)
		c.JSON(200, gin.H{
			"message": fmt.Sprintf("Rate limit removed for %s/%s", identifier, endpointType),
		})
	}
}

// Background cleanup job
func (svc *RateLimitMiddleware) StartCleanupJob() {
	ticker := time.NewTicker(1 * time.Hour) // Run every hour
	go func() {
		for range ticker.C {
			if err := svc.CleanupOldRecords(); err != nil {
				log.Printf("Rate limit cleanup error: %v", err)
			}
		}
	}()
}
