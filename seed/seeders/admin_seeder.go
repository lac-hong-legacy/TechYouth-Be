package seeders

import (
	"gorm.io/gorm"
)

// AdminSeeder handles seeding admin users
type AdminSeeder struct {
	db *gorm.DB
}

// NewAdminSeeder creates a new admin seeder
func NewAdminSeeder(db *gorm.DB) *AdminSeeder {
	return &AdminSeeder{db: db}
}

// SeedAdmin creates a default admin user
// func (s *AdminSeeder) SeedAdmin() error {
// 	// Check if admin already exists
// 	var existingAdmin model.User
// 	if err := s.db.Where("role = ?", "admin").First(&existingAdmin).Error; err == nil {
// 		log.Println("Admin user already exists, skipping admin seeding")
// 		return nil
// 	}

// 	// Hash password
// 	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
// 	if err != nil {
// 		return err
// 	}
// 	id, _ := uuid.NewV7()

// 	// Create admin user
// 	admin := model.User{
// 		ID:        id.String(),
// 		Email:     "admin@techyouth.com",
// 		Username:  "admin",
// 		Password:  string(hashedPassword),
// 		Role:      "admin",
// 		LastLogin: time.Now(),
// 		CreatedAt: time.Now(),
// 		UpdatedAt: time.Now(),
// 	}

// 	if err := s.db.Create(&admin).Error; err != nil {
// 		log.Printf("Error creating admin user: %v", err)
// 		return err
// 	}

// 	log.Printf("Created admin user: %s (password: admin123)", admin.Email)
// 	return nil
// }
