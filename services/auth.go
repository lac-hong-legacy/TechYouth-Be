package services

import (
	"errors"

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
	user, err := svc.sqlSvc.GetUserByEmail(loginRequest.Email)
	if err != nil {
		return nil, shared.NewInternalError(err, "Internal server error")
	}

	if user.Password != loginRequest.Password {
		return nil, shared.NewUnauthorizedError(errors.New("invalid password"), "Invalid password")
	}

	tokenPair, err := svc.jwtSvc.GenerateTokenPair(user.ID)
	if err != nil {
		return nil, shared.NewInternalError(err, "Internal server error")
	}

	return &dto.LoginResponse{
		AccessToken: tokenPair.AccessToken,
	}, nil
}
