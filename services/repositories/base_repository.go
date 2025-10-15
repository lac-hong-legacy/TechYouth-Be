package repositories

import (
	"gorm.io/gorm"
)

// BaseRepository provides common database functionality
type BaseRepository struct {
	db *gorm.DB
}

func NewBaseRepository(db *gorm.DB) BaseRepository {
	return BaseRepository{db: db}
}

// DB returns the underlying database connection
func (r *BaseRepository) DB() *gorm.DB {
	return r.db
}
