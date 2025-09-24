package services

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lac-hong-legacy/TechYouth-Be/dto"
	"github.com/lac-hong-legacy/TechYouth-Be/shared"

	"github.com/alphabatem/common/context"
)

type AuthService struct {
	context.DefaultService

	sqlSvc *SqliteService
	jwtSvc *JWTService
}

const AUTH_SVC = "auth_svc"

func (svc AuthService) Id() string {
	return AUTH_SVC
}

func (svc *AuthService) Configure(ctx *context.Context) error {
	return svc.DefaultService.Configure(ctx)
}

func (svc *AuthService) Start() error {
	svc.sqlSvc = svc.Service(SQLITE_SVC).(*SqliteService)
	svc.jwtSvc = svc.Service(JWT_SVC).(*JWTService)
	return nil
}

func (svc *AuthService) Register(registerRequest dto.RegisterRequest) (*dto.RegisterResponse, error) {
	user, err := svc.sqlSvc.CreateUser(registerRequest)
	if err != nil {
		return nil, shared.NewInternalError(err, err.Error())
	}

	return &dto.RegisterResponse{UserID: user.ID}, nil
}

func (svc *AuthService) Login(loginRequest dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := svc.sqlSvc.GetUserByEmailOrUsername(loginRequest.EmailOrUsername)
	if err != nil {
		return nil, shared.NewUnauthorizedError(err, "Invalid credentials")
	}

	if user.Password != loginRequest.Password {
		return nil, shared.NewUnauthorizedError(errors.New("invalid password"), "Invalid credentials")
	}

	tokenPair, err := svc.jwtSvc.GenerateTokenPair(user.ID)
	if err != nil {
		return nil, shared.NewInternalError(err, "Internal server error")
	}

	return &dto.LoginResponse{
		AccessToken: tokenPair.AccessToken,
	}, nil
}

func (svc *AuthService) RequiredAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		token, err := svc.jwtSvc.ExtractTokenFromHeader(authHeader)
		if err != nil {
			shared.ResponseJSON(c, http.StatusUnauthorized, "Unauthorized", err.Error())
			c.Abort()
			return
		}

		userID, err := svc.jwtSvc.VerifyJWTToken(token)
		if err != nil {
			shared.ResponseJSON(c, http.StatusUnauthorized, "Unauthorized", "Invalid JWT token")
			c.Abort()
			return
		}

		if userID == "" {
			shared.ResponseJSON(c, http.StatusUnauthorized, "Unauthorized", "Invalid user ID in token")
			c.Abort()
			return
		}

		c.Set(shared.UserID, userID)
		c.Next()
	}
}

// RequiredAdminAuth middleware for admin-only endpoints
func (svc *AuthService) RequiredAdminAuth() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// First check if user is authenticated
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			shared.ResponseJSON(c, http.StatusUnauthorized, "Unauthorized", "Authorization header required")
			c.Abort()
			return
		}

		token, err := svc.jwtSvc.ExtractTokenFromHeader(authHeader)
		if err != nil {
			shared.ResponseJSON(c, http.StatusUnauthorized, "Unauthorized", "Invalid authorization header")
			c.Abort()
			return
		}

		userID, err := svc.jwtSvc.VerifyJWTToken(token)
		if err != nil {
			shared.ResponseJSON(c, http.StatusUnauthorized, "Unauthorized", "Invalid or expired token")
			c.Abort()
			return
		}

		if userID == "" {
			shared.ResponseJSON(c, http.StatusUnauthorized, "Unauthorized", "Invalid user ID in token")
			c.Abort()
			return
		}

		// Check if user is admin
		user, err := svc.sqlSvc.GetUser(userID)
		if err != nil {
			shared.ResponseJSON(c, http.StatusUnauthorized, "Unauthorized", "User not found")
			c.Abort()
			return
		}

		if user.Role != "admin" {
			shared.ResponseJSON(c, http.StatusForbidden, "Forbidden", "Admin access required")
			c.Abort()
			return
		}

		c.Set(shared.UserID, userID)
		c.Next()
	})
}
