package services

import (
	"net/http"

	"github.com/alphabatem/common/context"
	"github.com/gin-gonic/gin"

	"tech-youth-be/shared"
)

type AuthMiddleware struct {
	context.DefaultService

	sqlSvc *SqliteService
	jwtSvc *JWTService
}

const AUTH_MIDDLEWARE_SVC = "auth"

func (svc AuthMiddleware) Id() string {
	return AUTH_MIDDLEWARE_SVC
}

func (svc *AuthMiddleware) Configure(ctx *context.Context) error {
	svc.sqlSvc = ctx.Service(SQLITE_SVC).(*SqliteService)
	svc.jwtSvc = ctx.Service(JWT_SVC).(*JWTService)
	return svc.DefaultService.Configure(ctx)
}

func (svc *AuthMiddleware) Start() error {
	return nil
}

func (svc *AuthMiddleware) RequiredAuth() gin.HandlerFunc {
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
