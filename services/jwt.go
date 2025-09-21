package services

import (
	"errors"
	"fmt"
	"github.com/lac-hong-legacy/TechYouth-Be/dto"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/alphabatem/common/context"
)

type JWTService struct {
	context.DefaultService

	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	jwtSecretKey         string
	sqlSvc               *SqliteService
}

type CustomClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

const JWT_SVC = "jwt_svc"

func (svc JWTService) Id() string {
	return JWT_SVC
}
func (svc *JWTService) Configure(ctx *context.Context) error {
	svc.sqlSvc = ctx.Service(SQLITE_SVC).(*SqliteService)

	svc.AccessTokenDuration = time.Duration(24 * time.Hour) // 24 hours
	svc.RefreshTokenDuration = time.Duration(24 * time.Hour)
	svc.jwtSecretKey = os.Getenv("JWT_OAUTH_SECRET")
	return svc.DefaultService.Configure(ctx)
}

func (svc *JWTService) Start() error {
	return nil
}

func (svc *JWTService) VerifyJWTToken(jwtToken string) (string, error) {
	token, err := jwt.ParseWithClaims(jwtToken, &CustomClaims{}, svc.getJWTKey)
	if err == nil && token.Valid {
		claims, ok := token.Claims.(*CustomClaims)
		if ok && claims != nil {
			// Validate expiration
			expTime, err := claims.GetExpirationTime()
			if err != nil {
				return "", fmt.Errorf("failed to get expiration time: %v", err)
			}
			now := jwt.NewNumericDate(time.Now())
			if expTime.Unix() < now.Unix() {
				return "", errors.New("token has expired")
			}

			return claims.UserID, nil
		}
	}

	return "", errors.New("unsupported JWT format")
}

func (svc *JWTService) getJWTKey(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}

	return []byte(svc.jwtSecretKey), nil
}

func (svc *JWTService) GenerateTokenPair(userId string) (*dto.TokenPair, error) {
	accessToken, err := svc.ToJWT(userId)
	if err != nil {
		return nil, err
	}

	return &dto.TokenPair{
		AccessToken: accessToken,
		ExpiresIn:   int64(svc.AccessTokenDuration.Seconds()),
	}, nil
}

func (svc *JWTService) ToJWT(userID string) (string, error) {
	expirationTime := svc.AccessTokenDuration
	expTime := time.Now().Add(expirationTime)

	claims := &CustomClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "TechYouth",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(svc.jwtSecretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %v", err)
	}

	return tokenString, nil
}

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
