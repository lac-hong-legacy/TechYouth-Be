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
		log.WithError(err).Fatal("Error loading .env file")
	}

	ctx, err := context.NewCtx(
		&services.SqliteService{},
		&services.JWTService{},
		&services.RateLimitService{},
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
