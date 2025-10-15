package services

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/cloakd/common/context"
	serviceContext "github.com/cloakd/common/services"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	docs "github.com/lac-hong-legacy/ven_api/docs"
	"github.com/lac-hong-legacy/ven_api/services/handlers"
	"github.com/lac-hong-legacy/ven_api/shared"
)

type HttpService struct {
	serviceContext.DefaultService

	jwtSvc      *JWTService
	authSvc     *AuthService
	guestSvc    *GuestService
	contentSvc  *ContentService
	userSvc     *UserService
	mediaSvc    *MediaService
	postgresSvc *PostgresService

	authHandler        *handlers.AuthHandler
	userHandler        *handlers.UserHandler
	guestHandler       *handlers.GuestHandler
	contentHandler     *handlers.ContentHandler
	leaderboardHandler *handlers.LeaderboardHandler
	adminHandler       *handlers.AdminHandler
	mediaHandler       *handlers.MediaHandler

	port int
	app  *fiber.App
}

const HTTP_SVC = "http_svc"

func (svc HttpService) Id() string {
	return HTTP_SVC
}

func (svc *HttpService) Configure(ctx *context.Context) error {
	if port := os.Getenv("HTTP_PORT"); port != "" {
		var err error
		if svc.port, err = strconv.Atoi(port); err != nil {
			return err
		}
	} else {
		svc.port = 8000
	}

	return svc.DefaultService.Configure(ctx)
}

func (svc *HttpService) Start() error {
	svc.jwtSvc = svc.Service(JWT_SVC).(*JWTService)
	svc.authSvc = svc.Service(AUTH_SVC).(*AuthService)
	svc.guestSvc = svc.Service(GUEST_SVC).(*GuestService)
	svc.userSvc = svc.Service(USER_SVC).(*UserService)
	svc.contentSvc = svc.Service(CONTENT_SVC).(*ContentService)
	svc.mediaSvc = svc.Service(MEDIA_SVC).(*MediaService)
	svc.postgresSvc = svc.Service(POSTGRES_SVC).(*PostgresService)

	svc.authHandler = handlers.NewAuthHandler(svc.authSvc, svc.jwtSvc, svc.userSvc)
	svc.userHandler = handlers.NewUserHandler(svc.userSvc, svc.authSvc)
	svc.guestHandler = handlers.NewGuestHandler(svc.guestSvc, svc.postgresSvc)
	svc.contentHandler = handlers.NewContentHandler(svc.contentSvc)
	svc.leaderboardHandler = handlers.NewLeaderboardHandler(svc.userSvc, svc.jwtSvc)
	svc.adminHandler = handlers.NewAdminHandler(svc.userSvc, svc.contentSvc)
	svc.mediaHandler = handlers.NewMediaHandler(svc.mediaSvc, svc.contentSvc, svc.postgresSvc)

	config := fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return svc.HandleError(c, err)
		},
	}

	svc.app = fiber.New(config)
	docs.SwaggerInfo.BasePath = ""

	svc.app.Use(recover.New())

	if os.Getenv("LOG_LEVEL") == "TRACE" {
		svc.app.Use(logger.New())
	}

	svc.app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowCredentials: false,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
	}))

	svc.setupRoutes()

	svc.app.Use(func(c *fiber.Ctx) error {
		return svc.HandleError(c, errors.New("page not found"))
	})

	return svc.app.Listen(fmt.Sprintf(":%v", svc.port))
}

func (svc *HttpService) setupRoutes() {
	svc.app.Get("/ping", svc.ping)
	svc.app.Get("/swagger/*", swagger.HandlerDefault)

	v1 := svc.app.Group("/api/v1")

	svc.setupAuthRoutes(v1)
	svc.setupGuestRoutes(v1)
	svc.setupContentRoutes(v1)
	svc.setupUserRoutes(v1)
	svc.setupLeaderboardRoutes(v1)
	svc.setupAdminRoutes(v1)
}

func (svc *HttpService) setupAuthRoutes(v1 fiber.Router) {
	v1.Post("/register", svc.authHandler.Register)
	v1.Post("/login", svc.authHandler.Login)
	v1.Post("/refresh", svc.authHandler.RefreshToken)
	v1.Post("/logout", svc.authSvc.RequiredAuth(), svc.authHandler.Logout)
	v1.Post("/logout-all", svc.authSvc.RequiredAuth(), svc.authHandler.LogoutAll)
	v1.Post("/verify-email", svc.authHandler.VerifyEmail)
	v1.Post("/resend-verification", svc.authHandler.ResendVerification)
	v1.Post("/forgot-password", svc.authHandler.ForgotPassword)
	v1.Post("/reset-password", svc.authHandler.ResetPassword)
	v1.Post("/change-password", svc.authSvc.RequiredAuth(), svc.authHandler.ChangePassword)
	v1.Get("/username/check/:username", svc.authHandler.CheckUsernameAvailability)
}

func (svc *HttpService) setupGuestRoutes(v1 fiber.Router) {
	guest := v1.Group("/guest")
	guest.Post("/session", svc.guestHandler.CreateSession)
	guest.Get("/session/:sessionId/progress", svc.guestHandler.GetProgress)
	guest.Get("/session/:sessionId/lesson/:lessonId/access", svc.guestHandler.CheckLessonAccess)
	guest.Post("/session/:sessionId/lesson/complete", svc.guestHandler.CompleteLesson)
	guest.Post("/session/:sessionId/hearts/add", svc.guestHandler.AddHeartsFromAd)
	guest.Post("/session/:sessionId/hearts/lose", svc.guestHandler.LoseHeart)
}

func (svc *HttpService) setupContentRoutes(v1 fiber.Router) {
	content := v1.Group("/content")
	content.Get("/timeline", svc.contentHandler.GetTimeline)
	content.Get("/characters", svc.contentHandler.GetCharacters)
	content.Get("/characters/:characterId", svc.contentHandler.GetCharacter)
	content.Get("/characters/:characterId/lessons", svc.contentHandler.GetCharacterLessons)
	content.Get("/lessons/:lessonId", svc.contentHandler.GetLesson)
	content.Post("/lessons/validate", svc.contentHandler.ValidateLessonAnswers)
	content.Post("/lessons/questions/answer", svc.authSvc.RequiredAuth(), svc.contentHandler.SubmitQuestionAnswer)
	content.Post("/lessons/status", svc.authSvc.RequiredAuth(), svc.contentHandler.CheckLessonStatus)
	content.Get("/search", svc.contentHandler.SearchContent)
	content.Get("/eras", svc.contentHandler.GetEras)
	content.Get("/dynasties", svc.contentHandler.GetDynasties)
}

func (svc *HttpService) setupUserRoutes(v1 fiber.Router) {
	user := v1.Group("/user", svc.authSvc.RequiredAuth())
	user.Get("/profile", svc.userHandler.GetUserProfile)
	user.Put("/profile", svc.userHandler.UpdateUserProfile)
	user.Post("/initialize", svc.userHandler.InitializeUserProfile)

	user.Get("/progress", svc.userHandler.GetUserProgress)
	user.Get("/collection", svc.userHandler.GetUserCollection)

	user.Get("/lesson/:lessonId/access", svc.userHandler.CheckUserLessonAccess)
	user.Post("/lesson/complete", svc.userHandler.CompleteUserLesson)

	user.Get("/hearts", svc.userHandler.GetHeartStatus)
	user.Post("/hearts/add", svc.userHandler.AddUserHearts)
	user.Post("/hearts/lose", svc.userHandler.LoseUserHeart)

	user.Get("/sessions", svc.userHandler.GetSessions)
	user.Delete("/sessions/:sessionId", svc.userHandler.RevokeSession)

	user.Get("/security", svc.userHandler.GetSecuritySettings)
	user.Put("/security", svc.userHandler.UpdateSecuritySettings)

	user.Get("/audit-logs", svc.userHandler.GetAuditLogs)

	user.Get("/devices", svc.userHandler.GetUserDevices)
	user.Put("/devices/:deviceId/trust", svc.userHandler.UpdateDeviceTrust)
	user.Delete("/devices/:deviceId", svc.userHandler.RemoveUserDevice)

	user.Post("/share", svc.userHandler.ShareAchievement)
}

func (svc *HttpService) setupLeaderboardRoutes(v1 fiber.Router) {
	leaderboard := v1.Group("/leaderboard")
	leaderboard.Get("/weekly", svc.leaderboardHandler.GetWeeklyLeaderboard)
	leaderboard.Get("/monthly", svc.leaderboardHandler.GetMonthlyLeaderboard)
	leaderboard.Get("/all-time", svc.leaderboardHandler.GetAllTimeLeaderboard)
}

func (svc *HttpService) setupAdminRoutes(v1 fiber.Router) {
	admin := v1.Group("/admin", svc.authSvc.RequireRole("admin"))
	admin.Post("/characters", svc.adminHandler.CreateCharacter)
	admin.Post("/lessons/new", svc.adminHandler.CreateLessonFromRequest)

	admin.Put("/lessons/:lessonId/script", svc.adminHandler.UpdateLessonScript)
	admin.Post("/lessons/:lessonId/audio", svc.mediaHandler.UploadLessonAudio)
	admin.Post("/lessons/:lessonId/animation", svc.mediaHandler.UploadLessonAnimation)
	admin.Get("/lessons/:lessonId/production-status", svc.adminHandler.GetLessonProductionStatus)

	admin.Post("/lessons/:lessonId/subtitle", svc.mediaHandler.UploadLessonSubtitle)
	admin.Post("/lessons/:lessonId/thumbnail", svc.mediaHandler.UploadThumbnail)
	admin.Get("/lessons/:lessonId/media", svc.mediaHandler.GetLessonMedia)
	admin.Delete("/media/assets/:assetId", svc.mediaHandler.DeleteMediaAsset)
	admin.Get("/media/statistics", svc.mediaHandler.GetMediaStatistics)
	admin.Get("/users", svc.adminHandler.AdminGetUsers)
	admin.Put("/users/:userId", svc.adminHandler.AdminUpdateUser)
	admin.Delete("/users/:userId", svc.adminHandler.AdminDeleteUser)
}

func (svc *HttpService) Shutdown() {
	_ = svc.app.Shutdown()
}

func (svc *HttpService) ping(c *fiber.Ctx) error {
	return shared.ResponseJSON(c, fiber.StatusOK, "Success", "pong")
}

func (svc *HttpService) HandleError(c *fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}

	if appErr, ok := shared.GetAppError(err); ok {
		return shared.ResponseJSON(c, appErr.StatusCode, appErr.Message, appErr.Data)
	}

	return shared.ResponseInternalError(c, err)
}
