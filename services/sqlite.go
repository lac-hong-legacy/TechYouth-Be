package services

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

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
	ds.database = os.Getenv("DB_DATABASE")

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
	models := []interface{}{}

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
