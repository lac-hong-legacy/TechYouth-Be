package services

import (
	"fmt"
	"os"
	"time"

	"github.com/cloakd/common/context"
	serviceContext "github.com/cloakd/common/services"
	"github.com/lac-hong-legacy/ven_api/model"
	"github.com/lac-hong-legacy/ven_api/services/repositories"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type PostgresService struct {
	serviceContext.DefaultService
	db       *gorm.DB
	database string

	userRepo      *repositories.UserRepository
	sessionRepo   *repositories.SessionRepository
	rateLimitRepo *repositories.RateLimitRepository
	mediaRepo     *repositories.MediaRepository
	contentRepo   *repositories.ContentRepository
	analyticRepo  *repositories.AnalyticRepository
}

const POSTGRES_SVC = "postgres_svc"

func (ds PostgresService) Id() string {
	return POSTGRES_SVC
}

func (ds PostgresService) Db() *gorm.DB {
	return ds.db
}

func (ds *PostgresService) Configure(ctx *context.Context) error {
	ds.database = os.Getenv("DATABASE_URL")
	if ds.database == "" {
		// Fallback to individual environment variables
		host := os.Getenv("DB_HOST")
		if host == "" {
			host = "localhost"
		}
		port := os.Getenv("DB_PORT")
		if port == "" {
			port = "5432"
		}
		user := os.Getenv("DB_USER")
		if user == "" {
			user = "ven_user"
		}
		password := os.Getenv("DB_PASSWORD")
		if password == "" {
			password = "ven_password"
		}
		dbname := os.Getenv("DB_NAME")
		if dbname == "" {
			dbname = "ven_api"
		}
		sslmode := os.Getenv("DB_SSLMODE")
		if sslmode == "" {
			sslmode = "disable"
		}
		timezone := os.Getenv("DB_TIMEZONE")
		if timezone == "" {
			timezone = "UTC"
		}

		ds.database = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
			host, user, password, dbname, port, sslmode, timezone)
	}

	return ds.DefaultService.Configure(ctx)
}

func (ds *PostgresService) Start() (err error) {
	ds.userRepo = repositories.NewUserRepository(ds.db)
	ds.sessionRepo = repositories.NewSessionRepository(ds.db)
	ds.rateLimitRepo = repositories.NewRateLimitRepository(ds.db)
	ds.mediaRepo = repositories.NewMediaRepository(ds.db)
	ds.contentRepo = repositories.NewContentRepository(ds.db)
	ds.analyticRepo = repositories.NewAnalyticRepository(ds.db)

	maxRetries := 10
	retryDelay := time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Printf("Attempting to connect to database (attempt %d/%d)...", attempt, maxRetries)

		ds.db, err = gorm.Open(postgres.Open(ds.database), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Error),
		})

		if err == nil {
			sqlDB, dbErr := ds.db.DB()
			if dbErr == nil {
				pingErr := sqlDB.Ping()
				if pingErr == nil {
					log.Println("Successfully connected to database")
					break
				}
				err = pingErr
			} else {
				err = dbErr
			}
		}

		if attempt == maxRetries {
			log.Printf("Failed to connect to database after %d attempts: %v", maxRetries, err)
			return err
		}

		log.Printf("Database connection failed: %v. Retrying in %v...", err, retryDelay)
		time.Sleep(retryDelay)

		retryDelay *= 2
		if retryDelay > 10*time.Second {
			retryDelay = 10 * time.Second
		}
	}

	models := []interface{}{
		// Existing models
		&model.User{},
		&model.GuestSession{},
		&model.GuestProgress{},
		&model.GuestLessonAttempt{},
		&model.RateLimit{},
		&model.RateLimitConfig{},

		// Content models
		&model.Character{},
		&model.Lesson{},
		&model.Timeline{},
		&model.MediaAsset{},
		&model.LessonMedia{},

		// User progress models
		&model.UserProgress{},
		&model.Spirit{},
		&model.Achievement{},
		&model.UserAchievement{},
		&model.UserLessonAttempt{},
		&model.UserQuestionAnswer{},

		// New authentication models
		&model.UserSession{},
		&model.AuthAuditLog{},
		&model.PasswordResetCode{},
		&model.BlacklistedToken{},
		&model.TrustedDevice{},
		&model.LoginAttempt{},
	}

	if err := ds.fixJSONBColumns(); err != nil {
		log.Printf("Failed to fix JSONB columns: %v", err)
		return err
	}

	err = ds.db.AutoMigrate(models...)
	if err != nil {
		log.Printf("Failed to migrate database: %v", err)
		return err
	}

	err = ds.userRepo.SeedInitialData()
	if err != nil {
		log.Printf("Failed to seed initial data: %v", err)
		return err
	}

	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for range ticker.C {
			err := ds.userRepo.CleanupExpiredData()
			if err != nil {
				log.Printf("Failed to cleanup expired data: %v", err)
			}
		}
	}()

	log.Println("Database connected and migrated successfully")
	return nil
}

func (ds *PostgresService) fixJSONBColumns() error {
	tables := []struct {
		table  string
		column string
	}{
		{"user_progresses", "completed_lessons"},
		{"user_progresses", "unlocked_characters"},
		{"guest_progresses", "completed_lessons"},
	}

	for _, t := range tables {
		var dataType string
		err := ds.db.Raw(`
			SELECT data_type 
			FROM information_schema.columns 
			WHERE table_name = ? AND column_name = ?
		`, t.table, t.column).Scan(&dataType).Error

		if err != nil || dataType == "" {
			// Column doesn't exist yet, skip
			continue
		}

		if dataType == "bytea" || dataType == "text" || dataType == "character varying" {
			log.Printf("Migrating %s.%s from %s to jsonb...", t.table, t.column, dataType)

			// First, ensure all values are valid JSON or convert them
			err = ds.db.Exec(fmt.Sprintf(`
				UPDATE %s 
				SET %s = '[]'::bytea 
				WHERE %s IS NULL OR %s = ''::bytea
			`, t.table, t.column, t.column, t.column)).Error

			if err != nil {
				log.Printf("Warning: Failed to update NULL/empty values in %s.%s: %v", t.table, t.column, err)
			}

			// Drop the column and recreate it as jsonb
			err = ds.db.Exec(fmt.Sprintf(`
				ALTER TABLE %s 
				DROP COLUMN IF EXISTS %s CASCADE
			`, t.table, t.column)).Error

			if err != nil {
				log.Printf("Failed to drop column %s.%s: %v", t.table, t.column, err)
				return err
			}

			// Add the column back as jsonb with default value
			err = ds.db.Exec(fmt.Sprintf(`
				ALTER TABLE %s 
				ADD COLUMN %s jsonb DEFAULT '[]'::jsonb
			`, t.table, t.column)).Error

			if err != nil {
				log.Printf("Failed to add jsonb column %s.%s: %v", t.table, t.column, err)
				return err
			}

			log.Printf("Successfully migrated %s.%s to jsonb", t.table, t.column)
		}
	}

	return nil
}

func (ds *PostgresService) Shutdown() {
	sqlDB, err := ds.db.DB()
	if err == nil {
		sqlDB.Close()
	}
}
