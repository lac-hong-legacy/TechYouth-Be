package services

import (
	context2 "context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/alphabatem/common/context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	docs "github.com/lac-hong-legacy/TechYouth-Be/docs"
	"github.com/lac-hong-legacy/TechYouth-Be/dto"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/lac-hong-legacy/TechYouth-Be/shared"
)

type HttpService struct {
	context.DefaultService

	jwtSvc    *JWTService
	authSvc   *AuthService
	guestSvc  *GuestService
	sqliteSvc *SqliteService

	port   int
	server *http.Server
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
	svc.sqliteSvc = svc.Service(SQLITE_SVC).(*SqliteService)

	// Set Gin mode
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	if os.Getenv("LOG_LEVEL") == "INFO" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	docs.SwaggerInfo.BasePath = ""
	r.Use(gin.Recovery())

	if os.Getenv("LOG_LEVEL") == "TRACE" {
		r.Use(gin.Logger())
	}
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowCredentials = true
	config.AddAllowHeaders("Authorization")
	r.Use(cors.New(config))

	//Validation endpoints
	r.GET("/ping", svc.ping)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	v1 := r.Group("/api/v1")

	v1.POST("/register", svc.Register)
	v1.POST("/login", svc.Login)

	guest := v1.Group("/guest")
	{
		guest.POST("/session", svc.CreateSession)
		guest.GET("/session/:sessionId/progress", svc.GetProgress)
		guest.GET("/session/:sessionId/lesson/:lessonId/access", svc.CheckLessonAccess)
		guest.POST("/session/:sessionId/lesson/complete", svc.CompleteLesson)
		guest.POST("/session/:sessionId/hearts/add", svc.AddHeartsFromAd)
		guest.POST("/session/:sessionId/hearts/lose", svc.LoseHeart)
	}

	r.NoRoute(func(c *gin.Context) {
		svc.HandleError(c, errors.New("page not found"))
	})

	svc.server = &http.Server{
		Addr:    fmt.Sprintf(":%v", svc.port),
		Handler: r,
	}

	return svc.server.ListenAndServe()
}

func (svc *HttpService) Shutdown() {
	ctx := context2.Background()
	_ = svc.server.Shutdown(ctx)
}

// @Summary Ping
// @Description This endpoint checks the health of the service
// @Tags health
// @Accept  json
// @Produce json
// @Success 200 {object} shared.Response{data=string}
// @Router /ping [get]
func (svc *HttpService) ping(c *gin.Context) {
	c.Header("Cache-Control", "max-age=10")

	shared.ResponseJSON(c, http.StatusOK, "Success", "pong")
}

func (svc *HttpService) HandleError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}

	if appErr, ok := shared.GetAppError(err); ok {
		shared.ResponseJSON(c, appErr.StatusCode, appErr.Message, appErr.Data)
		return true
	}

	shared.ResponseInternalError(c, err)
	return true
}

// @Summary Register
// @Description This endpoint registers a user
// @Tags auth
// @Accept  json
// @Produce json
// @Param registerRequest body dto.RegisterRequest true "Register request"
// @Success 200 {object} shared.Response{data=dto.RegisterResponse}
// @Router /api/v1/register [post]
func (svc *HttpService) Register(c *gin.Context) {
	var registerRequest dto.RegisterRequest
	if err := c.ShouldBindJSON(&registerRequest); err != nil {
		svc.HandleError(c, shared.NewBadRequestError(err, "Invalid request"))
		return
	}

	registerResponse, err := svc.authSvc.Register(registerRequest)
	if err != nil {
		svc.HandleError(c, err)
		return
	}

	shared.ResponseJSON(c, http.StatusOK, "Success", registerResponse)
}

// @Summary Login
// @Description This endpoint logs in a user
// @Tags auth
// @Accept  json
// @Produce json
// @Param loginRequest body dto.LoginRequest true "Login request"
// @Success 200 {object} shared.Response{data=dto.LoginResponse}
// @Router /api/v1/login [post]
func (svc *HttpService) Login(c *gin.Context) {
	var loginRequest dto.LoginRequest
	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		svc.HandleError(c, shared.NewBadRequestError(err, "Invalid request"))
		return
	}

	loginResponse, err := svc.authSvc.Login(loginRequest)
	if err != nil {
		svc.HandleError(c, err)
		return
	}

	shared.ResponseJSON(c, http.StatusOK, "Success", loginResponse)
}

// @Summary Create or Get Guest Session
// @Description This endpoint creates a new guest session or retrieves an existing one based on device ID
// @Tags guest
// @Accept  json
// @Produce json
// @Param createSessionRequest body dto.CreateSessionRequest true "Create session request"
// @Success 200
// @Router /api/v1/guest/session [post]
func (svc *HttpService) CreateSession(c *gin.Context) {
	var req dto.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		svc.HandleError(c, shared.NewBadRequestError(err, "Invalid request"))
		return
	}

	session, err := svc.guestSvc.CreateOrGetSession(req.DeviceID)
	if err != nil {
		svc.HandleError(c, err)
		return
	}

	progress, err := svc.sqliteSvc.GetProgress(session.ID)
	if err != nil {
		svc.HandleError(c, err)
		return
	}

	shared.ResponseJSON(c, http.StatusOK, "Success", dto.CreateSessionResponse{
		Session:  session,
		Progress: progress,
	})
}

// @Summary Get Guest Progress
// @Description This endpoint retrieves the progress of a guest session
// @Tags guest
// @Accept  json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200
// @Router /api/v1/guest/progress/{sessionId} [get]
func (svc *HttpService) GetProgress(c *gin.Context) {
	sessionID := c.Param("sessionId")

	progress, err := svc.sqliteSvc.GetProgress(sessionID)
	if err != nil {
		svc.HandleError(c, err)
		return
	}

	shared.ResponseJSON(c, http.StatusOK, "Success", progress)
}

// @Summary Check Lesson Access
// @Description This endpoint checks if a guest session can access a specific lesson
// @Tags guest
// @Accept  json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param lessonId path string true "Lesson ID"
// @Success 200 {object} shared.Response{data=dto.LessonAccessResponse}
// @Router /api/v1/guest/progress/{sessionId}/lesson/{lessonId}/access [get]
func (svc *HttpService) CheckLessonAccess(c *gin.Context) {
	sessionID := c.Param("sessionId")
	lessonID := c.Param("lessonId")

	canAccess, reason, err := svc.guestSvc.CanAccessLesson(sessionID, lessonID)
	if err != nil {
		svc.HandleError(c, err)
		return
	}

	res := dto.LessonAccessResponse{
		CanAccess: canAccess,
		Reason:    reason,
	}

	shared.ResponseJSON(c, http.StatusOK, "Success", res)
}

// @Summary Complete Lesson
// @Description This endpoint marks a lesson as completed for a guest session
// @Tags guest
// @Accept  json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param completeLessonRequest body dto.CompleteLessonRequest true "Complete lesson request"
// @Success 200
// @Router /api/v1/guest/progress/{sessionId}/complete [post]
func (svc *HttpService) CompleteLesson(c *gin.Context) {
	sessionID := c.Param("sessionId")

	var req dto.CompleteLessonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		svc.HandleError(c, shared.NewBadRequestError(err, "Invalid request"))
		return
	}

	err := svc.guestSvc.CompleteLesson(sessionID, req.LessonID, req.Score, req.TimeSpent)
	if err != nil {
		svc.HandleError(c, err)
		return
	}

	progress, err := svc.sqliteSvc.GetProgress(sessionID)
	if err != nil {
		svc.HandleError(c, err)
		return
	}

	shared.ResponseJSON(c, http.StatusOK, "Success", progress)
}

// @Summary Add Hearts from Ad
// @Description This endpoint adds hearts to a guest session when an ad is watched
// @Tags guest
// @Accept  json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200
// @Router /api/v1/guest/progress/{sessionId}/add-hearts [post]
func (svc *HttpService) AddHeartsFromAd(c *gin.Context) {
	sessionID := c.Param("sessionId")

	err := svc.guestSvc.AddHeartsFromAd(sessionID)
	if err != nil {
		svc.HandleError(c, err)
		return
	}

	progress, err := svc.sqliteSvc.GetProgress(sessionID)
	if err != nil {
		svc.HandleError(c, err)
		return
	}

	shared.ResponseJSON(c, http.StatusOK, "Success", progress)
}

// @Summary Lose Heart
// @Description This endpoint deducts a heart from a guest session when a lesson is failed
// @Tags guest
// @Accept  json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200
// @Router /api/v1/guest/progress/{sessionId}/lose-heart [post]
func (svc *HttpService) LoseHeart(c *gin.Context) {
	sessionID := c.Param("sessionId")

	err := svc.guestSvc.LoseHeart(sessionID)
	if err != nil {
		svc.HandleError(c, err)
		return
	}

	progress, err := svc.sqliteSvc.GetProgress(sessionID)
	if err != nil {
		svc.HandleError(c, err)
		return
	}

	shared.ResponseJSON(c, http.StatusOK, "Success", progress)
}
