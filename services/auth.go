package services

import (
	"net/http"
	"sync"
	"time"

	"tech-youth-be/dto"

	"github.com/sirupsen/logrus"

	"github.com/alphabatem/common/context"
	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	context.DefaultService

	sqlSvc      *SqliteService
	jwtSvc      *JWTService
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

		payload, loginType, err := svc.jwtSvc.VerifyUnifiedJWT(token)
		if err != nil {
			shared.ResponseJSON(c, http.StatusUnauthorized, "Unauthorized", "Invalid JWT token")
			c.Abort()
			return
		}

		userID := payload.AddressId
		// Validate user ID is not empty
		if userID == "" {
			shared.ResponseJSON(c, http.StatusUnauthorized, "Unauthorized", "Invalid user ID in token")
			c.Abort()
			return
		}

		_, err = svc.sqlSvc.findUserByWallet(userID)
		if err != nil {
			user := dto.UserCreationDTO{
				Id: userID,
			}

			uId, err := svc.sqlSvc.CreateOrGetUserByWallet(user)
			if err != nil {
				shared.ResponseJSON(c, http.StatusInternalServerError, "Internal Server Error", err.Error())
				c.Abort()
				return
			}

			userID = *uId

			go func() {
				if err := svc.imageSvc.randomImage(userID); err != nil {
					logrus.WithError(err).WithField("user_id", userID).Error("Failed to get random image")
				}
			}()
		}

		c.Set(shared.UserId, userID)
		c.Next()
	}
}
