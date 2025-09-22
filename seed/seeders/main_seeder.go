package seeders

import (
	"log"

	"gorm.io/gorm"
)

// MainSeeder coordinates all seeding operations
type MainSeeder struct {
	db *gorm.DB
}

// NewMainSeeder creates a new main seeder
func NewMainSeeder(db *gorm.DB) *MainSeeder {
	return &MainSeeder{db: db}
}

// SeedAll runs all seeders in the correct order
func (s *MainSeeder) SeedAll() error {
	log.Println("Starting database seeding...")

	// 1. Seed timelines first (no dependencies)
	timelineSeeder := NewTimelineSeeder(s.db)
	if err := timelineSeeder.SeedTimelines(); err != nil {
		log.Printf("Timeline seeding failed: %v", err)
		return err
	}

	// 2. Seed characters (depends on timelines)
	characterSeeder := NewCharacterSeeder(s.db)
	if err := characterSeeder.SeedCharacters(); err != nil {
		log.Printf("Character seeding failed: %v", err)
		return err
	}

	// 3. Seed lessons (depends on characters)
	lessonSeeder := NewLessonSeeder(s.db)
	if err := lessonSeeder.SeedLessons(); err != nil {
		log.Printf("Lesson seeding failed: %v", err)
		return err
	}

	// 4. Seed achievements (optional)
	if err := s.seedAchievements(); err != nil {
		log.Printf("Achievement seeding failed: %v", err)
		return err
	}

	log.Println("Database seeding completed successfully!")
	return nil
}

// seedAchievements seeds basic achievements
func (s *MainSeeder) seedAchievements() error {
	// You can expand this to include achievement seeding
	log.Println("Achievement seeding skipped (implement as needed)")
	return nil
}

// SeedCharactersOnly seeds only characters
func (s *MainSeeder) SeedCharactersOnly() error {
	characterSeeder := NewCharacterSeeder(s.db)
	return characterSeeder.SeedCharacters()
}

// SeedLessonsOnly seeds only lessons
func (s *MainSeeder) SeedLessonsOnly() error {
	lessonSeeder := NewLessonSeeder(s.db)
	return lessonSeeder.SeedLessons()
}

// SeedTimelinesOnly seeds only timelines
func (s *MainSeeder) SeedTimelinesOnly() error {
	timelineSeeder := NewTimelineSeeder(s.db)
	return timelineSeeder.SeedTimelines()
}
