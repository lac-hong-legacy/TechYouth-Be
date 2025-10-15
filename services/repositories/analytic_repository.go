package repositories

import "gorm.io/gorm"

type AnalyticRepository struct {
	BaseRepository
}

func NewAnalyticRepository(db *gorm.DB) *AnalyticRepository {
	return &AnalyticRepository{
		BaseRepository: NewBaseRepository(db),
	}
}
