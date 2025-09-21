package services

import (
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

	jwtSvc  *JWTService
	authSvc *AuthService
	port    int
	server  *http.Server
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
	_ = svc.server.Shutdown(nil)
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
