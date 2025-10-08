// cmd/seed/main.go
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/lac-hong-legacy/ven_api/seed/seeders"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Parse command line flags
	var (
		seedType = flag.String("type", "all", "Type of seeding: all")
	)
	flag.Parse()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
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

		databaseURL = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
			host, user, password, dbname, port, sslmode, timezone)
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
		return
	}

	log.Printf("Connected to database: %s", databaseURL)

	// Create main seeder
	mainSeeder := seeders.NewMainSeeder(db)

	// Run seeding based on type
	switch *seedType {
	case "all":
		log.Println("Running complete database seeding...")
		if err := mainSeeder.SeedAll(); err != nil {
			log.Fatalf("Failed to seed database: %v", err)
		}
	default:
		log.Fatalf("Unknown seed type: %s. Use 'all'", *seedType)
	}

	log.Println("Seeding operation completed successfully!")
}
