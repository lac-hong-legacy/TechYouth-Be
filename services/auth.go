package services

import (
	"errors"

	"github.com/gofiber/fiber/v2"
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

func (svc *AuthService) RequiredAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		token, err := svc.jwtSvc.ExtractTokenFromHeader(authHeader)
		if err != nil {
			return shared.ResponseJSON(c, fiber.StatusUnauthorized, "Unauthorized", err.Error())
		}

		userID, err := svc.jwtSvc.VerifyJWTToken(token)
		if err != nil {
			return shared.ResponseJSON(c, fiber.StatusUnauthorized, "Unauthorized", "Invalid JWT token")
		}

		if userID == "" {
			return shared.ResponseJSON(c, fiber.StatusUnauthorized, "Unauthorized", "Invalid user ID in token")
		}

		c.Locals(shared.UserID, userID)
		return c.Next()
	}
}

// RequiredAdminAuth middleware for admin-only endpoints
func (svc *AuthService) RequiredAdminAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// First check if user is authenticated
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return shared.ResponseJSON(c, fiber.StatusUnauthorized, "Unauthorized", "Authorization header required")
		}

		token, err := svc.jwtSvc.ExtractTokenFromHeader(authHeader)
		if err != nil {
			return shared.ResponseJSON(c, fiber.StatusUnauthorized, "Unauthorized", "Invalid authorization header")
		}

		userID, err := svc.jwtSvc.VerifyJWTToken(token)
		if err != nil {
			return shared.ResponseJSON(c, fiber.StatusUnauthorized, "Unauthorized", "Invalid or expired token")
		}

		if userID == "" {
			return shared.ResponseJSON(c, fiber.StatusUnauthorized, "Unauthorized", "Invalid user ID in token")
		}

		// Check if user is admin
		user, err := svc.sqlSvc.GetUser(userID)
		if err != nil {
			return shared.ResponseJSON(c, fiber.StatusUnauthorized, "Unauthorized", "User not found")
		}

		if user.Role != "admin" {
			return shared.ResponseJSON(c, fiber.StatusForbidden, "Forbidden", "Admin access required")
		}

		c.Locals(shared.UserID, userID)
		return c.Next()
	}
}
