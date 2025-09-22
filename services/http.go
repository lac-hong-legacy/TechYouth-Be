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
	"github.com/lac-hong-legacy/TechYouth-Be/model"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/lac-hong-legacy/TechYouth-Be/shared"
)

type HttpService struct {
	context.DefaultService

	jwtSvc     *JWTService
	authSvc    *AuthService
	guestSvc   *GuestService
	contentSvc *ContentService
	userSvc    *UserService
	sqliteSvc  *SqliteService

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
	svc.userSvc = svc.Service(USER_SVC).(*UserService)
	svc.contentSvc = svc.Service(CONTENT_SVC).(*ContentService)
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
	v1.GET("/username/check/:username", svc.CheckUsernameAvailability)

	guest := v1.Group("/guest")
	{
		guest.POST("/session", svc.CreateSession)
		guest.GET("/session/:sessionId/progress", svc.GetProgress)
		guest.GET("/session/:sessionId/lesson/:lessonId/access", svc.CheckLessonAccess)
		guest.POST("/session/:sessionId/lesson/complete", svc.CompleteLesson)
		guest.POST("/session/:sessionId/hearts/add", svc.AddHeartsFromAd)
		guest.POST("/session/:sessionId/hearts/lose", svc.LoseHeart)
	}

	content := v1.Group("/content")
	{
		content.GET("/timeline", svc.GetTimeline)
		content.GET("/characters", svc.GetCharacters)
		content.GET("/characters/:characterId", svc.GetCharacter)
		content.GET("/characters/:characterId/lessons", svc.GetCharacterLessons)
		content.GET("/lessons/:lessonId", svc.GetLesson)
		content.GET("/search", svc.SearchContent)
	}

	user := v1.Group("/user").Use(svc.authSvc.RequiredAuth())
	{
		// Profile management
		user.GET("/profile", svc.GetUserProfile)
		user.PUT("/profile", svc.UpdateUserProfile)
		user.POST("/initialize", svc.InitializeUserProfile)

		// Progress and game state
		user.GET("/progress", svc.GetUserProgress)
		user.GET("/stats", svc.GetUserStats)
		user.GET("/collection", svc.GetUserCollection)

		// Lesson management
		user.GET("/lesson/:lessonId/access", svc.CheckUserLessonAccess)
		user.POST("/lesson/complete", svc.CompleteUserLesson)

		// Hearts management
		user.GET("/hearts", svc.GetHeartStatus)
		user.POST("/hearts/add", svc.AddUserHearts)
		user.POST("/hearts/lose", svc.LoseUserHeart)

		// Social features
		user.POST("/share", svc.ShareAchievement)
	}

	leaderboard := v1.Group("/leaderboard")
	{
		leaderboard.GET("/weekly", svc.GetWeeklyLeaderboard)
		leaderboard.GET("/monthly", svc.GetMonthlyLeaderboard)
		leaderboard.GET("/all-time", svc.GetAllTimeLeaderboard)
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

// @Summary Check Username Availability
// @Description Check if a username is available for registration
// @Tags auth
// @Produce json
// @Param username path string true "Username to check"
// @Success 200 {object} shared.Response{data=map[string]interface{}}
// @Router /api/v1/username/check/{username} [get]
func (svc *HttpService) CheckUsernameAvailability(c *gin.Context) {
	username := c.Param("username")

	available, err := svc.userSvc.CheckUsernameAvailability(username)
	if err != nil {
		shared.ResponseJSON(c, http.StatusBadRequest, "Invalid username", map[string]interface{}{
			"available": false,
			"error":     err.Error(),
		})
		return
	}

	message := "Username is available"
	if !available {
		message = "Username is already taken"
	}

	shared.ResponseJSON(c, http.StatusOK, message, map[string]interface{}{
		"available": available,
		"username":  username,
	})
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

// ==================== CONTENT ENDPOINTS ====================

// @Summary Get Timeline
// @Description Get the historical timeline with eras and dynasties
// @Tags content
// @Accept json
// @Produce json
// @Success 200 {object} shared.Response{data=dto.TimelineCollectionResponse}
// @Router /api/v1/content/timeline [get]
func (svc *HttpService) GetTimeline(c *gin.Context) {
	timeline, err := svc.contentSvc.GetTimeline()
	if err != nil {
		svc.HandleError(c, err)
		return
	}

	shared.ResponseJSON(c, http.StatusOK, "Success", timeline)
}

// @Summary Get Characters
// @Description Get list of historical characters with filtering options
// @Tags content
// @Accept json
// @Produce json
// @Param dynasty query string false "Filter by dynasty"
// @Param rarity query string false "Filter by rarity"
// @Success 200 {object} shared.Response{data=dto.CharacterCollectionResponse}
// @Router /api/v1/content/characters [get]
func (svc *HttpService) GetCharacters(c *gin.Context) {
	dynasty := c.Query("dynasty")
	rarity := c.Query("rarity")

	characters, err := svc.contentSvc.GetCharacters(dynasty, rarity)
	if err != nil {
		svc.HandleError(c, err)
		return
	}

	shared.ResponseJSON(c, http.StatusOK, "Success", characters)
}

// @Summary Get Character
// @Description Get detailed information about a specific character
// @Tags content
// @Accept json
// @Produce json
// @Param characterId path string true "Character ID"
// @Success 200 {object} shared.Response{data=dto.CharacterResponse}
// @Router /api/v1/content/characters/{characterId} [get]
func (svc *HttpService) GetCharacter(c *gin.Context) {
	characterID := c.Param("characterId")

	character, err := svc.contentSvc.GetCharacterDetails(characterID)
	if err != nil {
		svc.HandleError(c, err)
		return
	}

	shared.ResponseJSON(c, http.StatusOK, "Success", character)
}

// @Summary Get Character Lessons
// @Description Get all lessons for a specific character
// @Tags content
// @Accept json
// @Produce json
// @Param characterId path string true "Character ID"
// @Success 200 {object} shared.Response{data=[]dto.LessonResponse}
// @Router /api/v1/content/characters/{characterId}/lessons [get]
func (svc *HttpService) GetCharacterLessons(c *gin.Context) {
	characterID := c.Param("characterId")

	lessons, err := svc.contentSvc.GetCharacterLessons(characterID)
	if err != nil {
		svc.HandleError(c, err)
		return
	}

	shared.ResponseJSON(c, http.StatusOK, "Success", lessons)
}

// @Summary Get Lesson
// @Description Get detailed lesson content including questions
// @Tags content
// @Accept json
// @Produce json
// @Param lessonId path string true "Lesson ID"
// @Success 200 {object} shared.Response{data=dto.LessonResponse}
// @Router /api/v1/content/lessons/{lessonId} [get]
func (svc *HttpService) GetLesson(c *gin.Context) {
	lessonID := c.Param("lessonId")

	lesson, err := svc.contentSvc.GetLessonContent(lessonID)
	if err != nil {
		svc.HandleError(c, err)
		return
	}

	shared.ResponseJSON(c, http.StatusOK, "Success", lesson)
}

// @Summary Search Content
// @Description Search characters and content with various filters
// @Tags content
// @Accept json
// @Produce json
// @Param query query string false "Search query"
// @Param dynasty query string false "Filter by dynasty"
// @Param rarity query string false "Filter by rarity"
// @Param limit query int false "Limit results"
// @Success 200 {object} shared.Response{data=dto.SearchResponse}
// @Router /api/v1/content/search [get]
func (svc *HttpService) SearchContent(c *gin.Context) {
	var req dto.SearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		svc.HandleError(c, shared.NewBadRequestError(err, "Invalid query parameters"))
		return
	}

	if req.Limit == 0 {
		req.Limit = 20
	}

	results, err := svc.contentSvc.SearchContent(req)
	if err != nil {
		svc.HandleError(c, err)
		return
	}

	shared.ResponseJSON(c, http.StatusOK, "Success", results)
}

// ==================== USER PROFILE ENDPOINTS ====================

// @Summary Get User Profile
// @Description Get current user's profile information
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Success 200 {object} shared.Response{data=dto.UserProfileResponse}
// @Router /api/v1/user/profile [get]
func (svc *HttpService) GetUserProfile(c *gin.Context) {
	userID := c.GetString(shared.UserID)

	profile, err := svc.userSvc.GetUserProfile(userID)
	if err != nil {
		svc.HandleError(c, err)
		return
	}

	shared.ResponseJSON(c, http.StatusOK, "Success", profile)
}

// @Summary Update User Profile
// @Description Update user profile information
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Param updateProfileRequest body dto.UpdateProfileRequest true "Update profile request"
// @Success 200 {object} shared.Response{data=dto.UserProfileResponse}
// @Router /api/v1/user/profile [put]
func (svc *HttpService) UpdateUserProfile(c *gin.Context) {
	userID := c.GetString(shared.UserID)

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		svc.HandleError(c, shared.NewBadRequestError(err, "Invalid request"))
		return
	}

	profile, err := svc.userSvc.UpdateUserProfile(userID, req)
	if err != nil {
		svc.HandleError(c, err)
		return
	}

	shared.ResponseJSON(c, http.StatusOK, "Success", profile)
}

// @Summary Initialize User Profile
// @Description Initialize user profile after first login (zodiac setup)
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Param birthYear body map[string]int true "Birth year for zodiac"
// @Success 200 {object} shared.Response{data=dto.UserProgressResponse}
// @Router /api/v1/user/initialize [post]
func (svc *HttpService) InitializeUserProfile(c *gin.Context) {
	userID := c.GetString(shared.UserID)

	var req map[string]int
	if err := c.ShouldBindJSON(&req); err != nil {
		svc.HandleError(c, shared.NewBadRequestError(err, "Invalid request"))
		return
	}

	birthYear, exists := req["birth_year"]
	if !exists || birthYear < 1900 || birthYear > 2020 {
		svc.HandleError(c, shared.NewBadRequestError(nil, "Valid birth year is required"))
		return
	}

	err := svc.userSvc.InitializeUserProfile(userID, birthYear)
	if err != nil {
		svc.HandleError(c, err)
		return
	}

	progress, err := svc.userSvc.GetUserProgress(userID)
	if err != nil {
		svc.HandleError(c, err)
		return
	}

	shared.ResponseJSON(c, http.StatusOK, "Success", progress)
}

// ==================== USER PROGRESS ENDPOINTS ====================

// @Summary Get User Progress
// @Description Get current user's game progress
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Success 200 {object} shared.Response{data=dto.UserProgressResponse}
// @Router /api/v1/user/progress [get]
func (svc *HttpService) GetUserProgress(c *gin.Context) {
	userID := c.GetString(shared.UserID)

	progress, err := svc.userSvc.GetUserProgress(userID)
	if err != nil {
		svc.HandleError(c, err)
		return
	}

	shared.ResponseJSON(c, http.StatusOK, "Success", progress)
}

// @Summary Get User Stats
// @Description Get detailed user statistics and analytics
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Success 200 {object} shared.Response{data=dto.UserStatsResponse}
// @Router /api/v1/user/stats [get]
func (svc *HttpService) GetUserStats(c *gin.Context) {
	userID := c.GetString(shared.UserID)

	stats, err := svc.userSvc.GetUserStats(userID)
	if err != nil {
		svc.HandleError(c, err)
		return
	}

	shared.ResponseJSON(c, http.StatusOK, "Success", stats)
}

// @Summary Get User Collection
// @Description Get user's character collection and achievements
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param UserID query string false "User ID"
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Success 200 {object} shared.Response{data=dto.CollectionResponse}
// @Router /api/v1/user/collection [get]
func (svc *HttpService) GetUserCollection(c *gin.Context) {
	userID := c.Query("userId")
	if userID == "" {
		userID = c.GetString(shared.UserID)
	}

	collection, err := svc.userSvc.GetUserCollection(userID)
	if err != nil {
		svc.HandleError(c, err)
		return
	}

	shared.ResponseJSON(c, http.StatusOK, "Success", collection)
}

// ==================== USER LESSON ENDPOINTS ====================

// @Summary Check User Lesson Access
// @Description Check if user can access a specific lesson
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param lessonId path string true "Lesson ID"
// @Success 200 {object} shared.Response{data=dto.LessonAccessResponse}
// @Router /api/v1/user/lesson/{lessonId}/access [get]
func (svc *HttpService) CheckUserLessonAccess(c *gin.Context) {
	userID := c.GetString(shared.UserID)
	lessonID := c.Param("lessonId")

	access, err := svc.userSvc.CheckLessonAccess(userID, lessonID)
	if err != nil {
		svc.HandleError(c, err)
		return
	}

	shared.ResponseJSON(c, http.StatusOK, "Success", access)
}

// @Summary Complete User Lesson
// @Description Complete a lesson for registered user
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param completeLessonRequest body dto.CompleteLessonRequest true "Complete lesson request"
// @Success 200 {object} shared.Response{data=dto.CompleteLessonResponse}
// @Router /api/v1/user/lesson/complete [post]
func (svc *HttpService) CompleteUserLesson(c *gin.Context) {
	userID := c.GetString(shared.UserID)

	var req dto.CompleteLessonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		svc.HandleError(c, shared.NewBadRequestError(err, "Invalid request"))
		return
	}

	err := svc.userSvc.CompleteLesson(userID, req.LessonID, req.Score, req.TimeSpent)
	if err != nil {
		svc.HandleError(c, err)
		return
	}

	result, err := svc.userSvc.GetUserProgress(userID)
	if err != nil {
		svc.HandleError(c, err)
		return
	}

	shared.ResponseJSON(c, http.StatusOK, "Success", result)
}

// ==================== HEARTS MANAGEMENT ENDPOINTS ====================

// @Summary Get Heart Status
// @Description Get current heart status and reset information
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Success 200 {object} shared.Response{data=dto.HeartStatusResponse}
// @Router /api/v1/user/hearts [get]
func (svc *HttpService) GetHeartStatus(c *gin.Context) {
	userID := c.GetString(shared.UserID)

	status, err := svc.userSvc.GetHeartStatus(userID)
	if err != nil {
		svc.HandleError(c, err)
		return
	}

	shared.ResponseJSON(c, http.StatusOK, "Success", status)
}

// @Summary Add User Hearts
// @Description Add hearts from ads or other sources
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Param addHeartsRequest body dto.AddHeartsRequest true "Add hearts request"
// @Success 200 {object} shared.Response{data=dto.HeartStatusResponse}
// @Router /api/v1/user/hearts/add [post]
func (svc *HttpService) AddUserHearts(c *gin.Context) {
	userID := c.GetString(shared.UserID)

	var req dto.AddHeartsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		svc.HandleError(c, shared.NewBadRequestError(err, "Invalid request"))
		return
	}

	status, err := svc.userSvc.AddHearts(userID, req.Source, req.Amount)
	if err != nil {
		svc.HandleError(c, err)
		return
	}

	shared.ResponseJSON(c, http.StatusOK, "Success", status)
}

// @Summary Lose User Heart
// @Description Deduct a heart when user fails a lesson
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "User Bearer Token" default(Bearer <user_token>)
// @Success 200 {object} shared.Response{data=dto.HeartStatusResponse}
// @Router /api/v1/user/hearts/lose [post]
func (svc *HttpService) LoseUserHeart(c *gin.Context) {
	userID := c.GetString(shared.UserID)

	status, err := svc.userSvc.LoseHeart(userID)
	if err != nil {
		svc.HandleError(c, err)
		return
	}

	shared.ResponseJSON(c, http.StatusOK, "Success", status)
}

// ==================== LEADERBOARD ENDPOINTS ====================

// @Summary Get Weekly Leaderboard
// @Description Get weekly leaderboard rankings
// @Tags leaderboard
// @Accept json
// @Produce json
// @Param limit query int false "Limit results (default 50)"
// @Success 200 {object} shared.Response{data=dto.LeaderboardResponse}
// @Router /api/v1/leaderboard/weekly [get]
func (svc *HttpService) GetWeeklyLeaderboard(c *gin.Context) {
	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	var userID string
	if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		if token, err := svc.jwtSvc.ExtractTokenFromHeader(authHeader); err == nil {
			if uid, err := svc.jwtSvc.VerifyJWTToken(token); err == nil {
				userID = uid
			}
		}
	}

	leaderboard, err := svc.userSvc.GetWeeklyLeaderboard(limit, userID)
	if err != nil {
		svc.HandleError(c, err)
		return
	}

	shared.ResponseJSON(c, http.StatusOK, "Success", leaderboard)
}

// @Summary Get Monthly Leaderboard
// @Description Get monthly leaderboard rankings
// @Tags leaderboard
// @Accept json
// @Produce json
// @Param limit query int false "Limit results (default 50)"
// @Success 200 {object} shared.Response{data=dto.LeaderboardResponse}
// @Router /api/v1/leaderboard/monthly [get]
func (svc *HttpService) GetMonthlyLeaderboard(c *gin.Context) {
	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	var userID string
	if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		if token, err := svc.jwtSvc.ExtractTokenFromHeader(authHeader); err == nil {
			if uid, err := svc.jwtSvc.VerifyJWTToken(token); err == nil {
				userID = uid
			}
		}
	}

	leaderboard, err := svc.userSvc.GetMonthlyLeaderboard(limit, userID)
	if err != nil {
		svc.HandleError(c, err)
		return
	}

	shared.ResponseJSON(c, http.StatusOK, "Success", leaderboard)
}

// @Summary Get All Time Leaderboard
// @Description Get all-time leaderboard rankings
// @Tags leaderboard
// @Accept json
// @Produce json
// @Param limit query int false "Limit results (default 50)"
// @Success 200 {object} shared.Response{data=dto.LeaderboardResponse}
// @Router /api/v1/leaderboard/all-time [get]
func (svc *HttpService) GetAllTimeLeaderboard(c *gin.Context) {
	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	var userID string
	if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		if token, err := svc.jwtSvc.ExtractTokenFromHeader(authHeader); err == nil {
			if uid, err := svc.jwtSvc.VerifyJWTToken(token); err == nil {
				userID = uid
			}
		}
	}

	leaderboard, err := svc.userSvc.GetAllTimeLeaderboard(limit, userID)
	if err != nil {
		svc.HandleError(c, err)
		return
	}

	shared.ResponseJSON(c, http.StatusOK, "Success", leaderboard)
}

// ==================== SOCIAL ENDPOINTS ====================

// @Summary Share Achievement
// @Description Share user achievement or progress on social media
// @Tags user
// @Accept json
// @Produce json
// @Security Bearer
// @Param shareRequest body dto.ShareRequest true "Share request"
// @Success 200 {object} shared.Response{data=dto.ShareResponse}
// @Router /api/v1/user/share [post]
func (svc *HttpService) ShareAchievement(c *gin.Context) {
	userID := c.GetString(shared.UserID)

	var req dto.ShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		svc.HandleError(c, shared.NewBadRequestError(err, "Invalid request"))
		return
	}

	shareData, err := svc.userSvc.CreateShareContent(userID, req)
	if err != nil {
		svc.HandleError(c, err)
		return
	}

	shared.ResponseJSON(c, http.StatusOK, "Success", shareData)
}

// ==================== ADMIN ENDPOINTS (Optional) ====================

// @Summary Create Character (Admin)
// @Description Create a new historical character (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security Bearer
// @Param character body model.Character true "Character data"
// @Success 201 {object} shared.Response{data=dto.CharacterResponse}
// @Router /api/v1/admin/characters [post]
func (svc *HttpService) CreateCharacter(c *gin.Context) {
	// TODO: Add admin authentication check

	var character model.Character
	if err := c.ShouldBindJSON(&character); err != nil {
		svc.HandleError(c, shared.NewBadRequestError(err, "Invalid character data"))
		return
	}

	created, err := svc.contentSvc.CreateCharacter(&character)
	if err != nil {
		svc.HandleError(c, err)
		return
	}

	shared.ResponseJSON(c, http.StatusCreated, "Character created successfully", created)
}

// @Summary Create Lesson (Admin)
// @Description Create a new lesson (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security Bearer
// @Param lesson body model.Lesson true "Lesson data"
// @Success 201 {object} shared.Response{data=dto.LessonResponse}
// @Router /api/v1/admin/lessons [post]
func (svc *HttpService) CreateLesson(c *gin.Context) {
	// TODO: Add admin authentication check

	var lesson model.Lesson
	if err := c.ShouldBindJSON(&lesson); err != nil {
		svc.HandleError(c, shared.NewBadRequestError(err, "Invalid lesson data"))
		return
	}

	created, err := svc.contentSvc.CreateLesson(&lesson)
	if err != nil {
		svc.HandleError(c, err)
		return
	}

	shared.ResponseJSON(c, http.StatusCreated, "Lesson created successfully", created)
}

// ==================== HELPER METHODS ====================

func (svc *HttpService) getUserIDFromOptionalAuth(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	token, err := svc.jwtSvc.ExtractTokenFromHeader(authHeader)
	if err != nil {
		return ""
	}

	userID, err := svc.jwtSvc.VerifyJWTToken(token)
	if err != nil {
		return ""
	}

	return userID
}
