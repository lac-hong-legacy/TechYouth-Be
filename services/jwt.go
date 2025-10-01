package services

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lac-hong-legacy/ven_api/dto"

	"github.com/alphabatem/common/context"
)

type JWTService struct {
	context.DefaultService

	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	jwtSecretKey         string
	refreshSecretKey     string
	sqlSvc               *SqliteService
}

type CustomClaims struct {
	UserID    string `json:"user_id"`
	TokenType string `json:"token_type"` // "access" or "refresh"
	SessionID string `json:"session_id,omitempty"`
	jwt.RegisteredClaims
}

const JWT_SVC = "jwt_svc"

func (svc JWTService) Id() string {
	return JWT_SVC
}

func (svc *JWTService) Configure(ctx *context.Context) error {
	svc.sqlSvc = ctx.Service(SQLITE_SVC).(*SqliteService)

	// Access tokens: 15 minutes (short-lived for security)
	svc.AccessTokenDuration = time.Duration(15 * time.Minute)

	// Refresh tokens: 7 days (longer-lived)
	svc.RefreshTokenDuration = time.Duration(7 * 24 * time.Hour)

	svc.jwtSecretKey = os.Getenv("JWT_ACCESS_SECRET")
	if svc.jwtSecretKey == "" {
		svc.jwtSecretKey = os.Getenv("JWT_OAUTH_SECRET") // fallback to existing
	}

	svc.refreshSecretKey = os.Getenv("JWT_REFRESH_SECRET")
	if svc.refreshSecretKey == "" {
		svc.refreshSecretKey = svc.jwtSecretKey + "_refresh" // fallback
	}

	return svc.DefaultService.Configure(ctx)
}

func (svc *JWTService) Start() error {
	return nil
}

// Generate both access and refresh tokens
func (svc *JWTService) GenerateTokenPair(userID string) (*dto.TokenPair, error) {
	// Generate access token
	accessToken, err := svc.generateAccessToken(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %v", err)
	}

	// Generate refresh token
	refreshToken, err := svc.generateRefreshToken(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %v", err)
	}

	return &dto.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(svc.AccessTokenDuration.Seconds()),
	}, nil
}

// Generate access token (short-lived)
func (svc *JWTService) generateAccessToken(userID string) (string, error) {
	now := time.Now()
	expirationTime := now.Add(svc.AccessTokenDuration)

	claims := &CustomClaims{
		UserID:    userID,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "TechYouth",
			Subject:   userID,
			ID:        svc.generateJTI(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(svc.jwtSecretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %v", err)
	}

	return tokenString, nil
}

// Generate refresh token (long-lived)
func (svc *JWTService) generateRefreshToken(userID string) (string, error) {
	now := time.Now()
	expirationTime := now.Add(svc.RefreshTokenDuration)

	claims := &CustomClaims{
		UserID:    userID,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "TechYouth",
			Subject:   userID,
			ID:        svc.generateJTI(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(svc.refreshSecretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %v", err)
	}

	return tokenString, nil
}

// Verify access token
func (svc *JWTService) VerifyJWTToken(jwtToken string) (string, error) {
	token, err := jwt.ParseWithClaims(jwtToken, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return svc.getAccessTokenKey(token)
	})

	if err != nil {
		return "", fmt.Errorf("failed to parse token: %v", err)
	}

	if !token.Valid {
		return "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	// Verify token type
	if claims.TokenType != "access" {
		return "", errors.New("invalid token type")
	}

	// Check if token is blacklisted
	if svc.isTokenBlacklisted(claims.ID) {
		return "", errors.New("token has been revoked")
	}

	// Validate expiration
	if claims.ExpiresAt.Time.Before(time.Now()) {
		return "", errors.New("token has expired")
	}

	return claims.UserID, nil
}

func (svc *JWTService) VerifyRefreshToken(refreshToken string) (string, error) {
	token, err := jwt.ParseWithClaims(refreshToken, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return svc.getRefreshTokenKey(token)
	})

	if err != nil {
		return "", fmt.Errorf("failed to parse refresh token: %v", err)
	}

	if !token.Valid {
		return "", errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return "", errors.New("invalid refresh token claims")
	}

	if claims.TokenType != "refresh" {
		return "", errors.New("invalid token type")
	}

	if svc.isTokenBlacklisted(claims.ID) {
		return "", errors.New("refresh token has been revoked")
	}

	if claims.ExpiresAt.Time.Before(time.Now()) {
		return "", errors.New("refresh token has expired")
	}

	return claims.UserID, nil
}

// Get access token signing key
func (svc *JWTService) getAccessTokenKey(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	return []byte(svc.jwtSecretKey), nil
}

// Get refresh token signing key
func (svc *JWTService) getRefreshTokenKey(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	return []byte(svc.refreshSecretKey), nil
}

// Extract token from Authorization header
func (svc *JWTService) ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is missing")
	}

	// Check if the header starts with "Bearer "
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return "", errors.New("invalid authorization header format")
	}

	// Extract the token
	return authHeader[7:], nil
}

// Generate unique token ID (JTI)
func (svc *JWTService) generateJTI() string {
	return fmt.Sprintf("jti_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}

// Check if token is blacklisted
func (svc *JWTService) isTokenBlacklisted(jti string) bool {
	return svc.sqlSvc.IsTokenBlacklisted(jti)
}

// Blacklist token (for logout)
func (svc *JWTService) BlacklistToken(jwtToken string) error {
	token, err := jwt.ParseWithClaims(jwtToken, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return svc.getAccessTokenKey(token)
	})

	if err != nil {
		return fmt.Errorf("failed to parse token: %v", err)
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return errors.New("invalid token claims")
	}

	// Add to blacklist with expiration time
	return svc.sqlSvc.BlacklistToken(claims.ID, claims.ExpiresAt.Time)
}

// Get token claims without verification (for logout)
func (svc *JWTService) GetTokenClaims(jwtToken string) (*CustomClaims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(jwtToken, &CustomClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// Clean up expired blacklisted tokens
func (svc *JWTService) CleanupExpiredTokens() error {
	return svc.sqlSvc.CleanupExpiredBlacklistedTokens()
}
