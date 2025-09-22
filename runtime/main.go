package main

import (
	"github.com/lac-hong-legacy/TechYouth-Be/services"

	"github.com/alphabatem/common/context"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading .env file")
	}

	ctx, err := context.NewCtx(
		&services.SqliteService{},
		&services.JWTService{},
		&services.RateLimitService{},
		&services.AuthService{},
		&services.GuestService{},
		&services.UserService{},
		&services.ContentService{},
		&services.HttpService{},
	)
	if err != nil {
		log.Fatal().Err(err)
		return
	}

	err = ctx.Run()
	if err != nil {
		log.Fatal().Err(err)
		return
	}
}
