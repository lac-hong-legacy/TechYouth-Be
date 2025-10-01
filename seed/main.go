// cmd/seed/main.go
package main

import (
	"flag"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/lac-hong-legacy/ven_api/seed/seeders"
	"gorm.io/driver/sqlite"
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
		seedType = flag.String("type", "all", "Type of seeding: all, characters, lessons, timelines")
		dbPath   = flag.String("db", "", "Database path (overrides DB_NAME env var)")
		help     = flag.Bool("help", false, "Show help message")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// Get database path
	databasePath := *dbPath
	if databasePath == "" {
		databasePath = os.Getenv("DB_NAME")
		if databasePath == "" {
			databasePath = "app.db" // Default database name
		}
	}

	// Connect to database
	db, err := gorm.Open(sqlite.Open(databasePath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Printf("Connected to database: %s", databasePath)

	// Create main seeder
	mainSeeder := seeders.NewMainSeeder(db)

	// Run seeding based on type
	switch *seedType {
	case "all":
		log.Println("Running complete database seeding...")
		if err := mainSeeder.SeedAll(); err != nil {
			log.Fatalf("Failed to seed database: %v", err)
		}
	case "characters":
		log.Println("Seeding characters only...")
		if err := mainSeeder.SeedCharactersOnly(); err != nil {
			log.Fatalf("Failed to seed characters: %v", err)
		}
	case "lessons":
		log.Println("Seeding lessons only...")
		if err := mainSeeder.SeedLessonsOnly(); err != nil {
			log.Fatalf("Failed to seed lessons: %v", err)
		}
	case "timelines":
		log.Println("Seeding timelines only...")
		if err := mainSeeder.SeedTimelinesOnly(); err != nil {
			log.Fatalf("Failed to seed timelines: %v", err)
		}
	default:
		log.Fatalf("Unknown seed type: %s. Use 'all', 'characters', 'lessons', or 'timelines'", *seedType)
	}

	log.Println("Seeding operation completed successfully!")
}

func showHelp() {
	log.Println(`
Database Seeding Tool for Vietnamese History Learning App

Usage: go run cmd/seed/main.go [flags]

Flags:
  -type string
        Type of seeding to perform (default "all")
        Options: all, characters, lessons, timelines
  -db string
        Database path (overrides DB_NAME environment variable)
  -help
        Show this help message

Examples:
  # Seed everything
  go run cmd/seed/main.go

  # Seed only characters
  go run cmd/seed/main.go -type=characters

  # Seed with custom database path
  go run cmd/seed/main.go -db=./custom.db

  # Seed only lessons
  go run cmd/seed/main.go -type=lessons

Environment Variables:
  DB_NAME - Default database path (default: app.db)
`)
}

// Makefile addition for easy seeding
/*
Add these targets to your Makefile:

.PHONY: seed seed-characters seed-lessons seed-timelines

# Seed all data
seed:
	@echo "Seeding database..."
	@go run cmd/seed/main.go -type=all

# Seed only characters
seed-characters:
	@echo "Seeding characters..."
	@go run cmd/seed/main.go -type=characters

# Seed only lessons
seed-lessons:
	@echo "Seeding lessons..."
	@go run cmd/seed/main.go -type=lessons

# Seed only timelines
seed-timelines:
	@echo "Seeding timelines..."
	@go run cmd/seed/main.go -type=timelines

# Clean and reseed database
reseed: clean-db seed
	@echo "Database reseeded successfully!"

# Remove database file
clean-db:
	@echo "Cleaning database..."
	@rm -f app.db *.db
*/
