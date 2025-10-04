package main

import (
	"github.com/lac-hong-legacy/ven_api/services"

	"github.com/alphabatem/common/context"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Info("Error loading .env file", err)
	}

	ctx, err := context.NewCtx(
		&services.PostgresService{},
		&services.RedisService{},
		&services.MinIOService{},
		&services.JWTService{},
		&services.RateLimitService{},
		&services.GeolocationService{},
		&services.AuthService{},
		&services.GuestService{},
		&services.ContentService{},
		&services.MediaService{},
		&services.UserService{},
		&services.EmailService{},
		&services.HttpService{},
	)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize context")
		return
	}

	err = ctx.Run()
	if err != nil {
		log.WithError(err).Fatal("Failed to run application")
		return
	}
}
