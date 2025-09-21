package services

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/lac-hong-legacy/TechYouth-Be/dto"
	"github.com/lac-hong-legacy/TechYouth-Be/model"

	"github.com/alphabatem/common/context"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type SqliteService struct {
	context.DefaultService
	db *gorm.DB

	database string
}

const SQLITE_SVC = "sqlite_svc"

// Id returns Service ID
func (ds SqliteService) Id() string {
	return SQLITE_SVC
}

// Db Access to raw SqliteService db
func (ds SqliteService) Db() *gorm.DB {
	return ds.db
}

// Configure the service
func (ds *SqliteService) Configure(ctx *context.Context) error {
	ds.database = os.Getenv("DB_NAME")

	return ds.DefaultService.Configure(ctx)
}

// Start the service and open connection to the database
// Migrate any tables that have changed since last runtime
func (ds *SqliteService) Start() (err error) {
	ds.db, err = gorm.Open(sqlite.Open(ds.database), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		return err
	}
	models := []interface{}{
		&model.User{},
		&model.GuestSession{},
		&model.GuestProgress{},
		&model.GuestLessonAttempt{},
	}

	err = ds.db.AutoMigrate(models...)
	if err != nil {
		log.Printf("Failed to migrate database: %v", err)
		return err
	}

	log.Println("Database connected and migrated successfully")
	return nil
}

func (ds *SqliteService) Shutdown() {
}

func (ds *SqliteService) HandleError(err error) error {
	if err == nil {
		return nil
	}

	var statusCode int
	var errorType string

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		statusCode = http.StatusNotFound // 404
		errorType = "NOT_FOUND"
	case errors.Is(err, gorm.ErrDuplicatedKey):
		statusCode = http.StatusConflict // 409
		errorType = "CONFLICT"
	case errors.Is(err, gorm.ErrForeignKeyViolated):
		statusCode = http.StatusBadRequest // 400
		errorType = "FOREIGN_KEY_VIOLATION"
	case errors.Is(err, gorm.ErrInvalidTransaction):
		statusCode = http.StatusInternalServerError // 500
		errorType = "TRANSACTION_ERROR"
	default:
		// Check for SQLite-specific errors
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			statusCode = http.StatusConflict // 409
			errorType = "UNIQUE_CONSTRAINT"
		} else if strings.Contains(err.Error(), "no such table") {
			statusCode = http.StatusInternalServerError // 500
			errorType = "SCHEMA_ERROR"
		} else {
			statusCode = http.StatusInternalServerError // 500
			errorType = "INTERNAL_ERROR"
		}
	}

	logEntry := log.WithFields(log.Fields{
		"status_code": statusCode,
		"error_type":  errorType,
		"error":       err.Error(),
	})

	if statusCode >= 500 {
		logEntry.Error("Database error occurred")
	} else {
		logEntry.Warn("Database operation failed")
	}

	return fmt.Errorf("%s: %w", errorType, err)
}

func (ds *SqliteService) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	if err := ds.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (ds *SqliteService) CreateUser(user dto.RegisterRequest) (*model.User, error) {
	if _, err := ds.GetUserByEmail(user.Email); err == nil {
		return nil, errors.New("user already exists")
	}

	userModel := model.User{
		ID:       uuid.New().String(),
		Email:    user.Email,
		Password: user.Password,
	}

	if err := ds.db.Create(&userModel).Error; err != nil {
		return nil, err
	}
	return &userModel, nil
}

func (ds *SqliteService) GetSessionByDeviceID(deviceID string) (*model.GuestSession, error) {
	var session model.GuestSession
	if err := ds.db.Where("device_id = ?", deviceID).First(&session).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return &session, nil
}

func (ds *SqliteService) CreateSession(session *model.GuestSession) (*model.GuestSession, error) {
	session.ID = uuid.New().String()
	if err := ds.db.Create(session).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return session, nil
}

func (ds *SqliteService) UpdateSession(session *model.GuestSession) error {
	if err := ds.db.Save(session).Error; err != nil {
		return ds.HandleError(err)
	}
	return nil
}

func (ds *SqliteService) GetProgress(sessionID string) (*model.GuestProgress, error) {
	var progress model.GuestProgress
	if err := ds.db.Where("session_id = ?", sessionID).First(&progress).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return &progress, nil
}

func (ds *SqliteService) CreateProgress(progress *model.GuestProgress) (*model.GuestProgress, error) {
	progress.ID = uuid.New().String()
	if err := ds.db.Create(progress).Error; err != nil {
		return nil, ds.HandleError(err)
	}
	return progress, nil
}

func (ds *SqliteService) UpdateProgress(progress *model.GuestProgress) error {
	if err := ds.db.Save(progress).Error; err != nil {
		return ds.HandleError(err)
	}
	return nil
}

func (ds *SqliteService) CreateLessonAttempt(attempt *model.GuestLessonAttempt) error {
	attempt.ID = uuid.New().String()
	if err := ds.db.Create(attempt).Error; err != nil {
		return ds.HandleError(err)
	}
	return nil
}
