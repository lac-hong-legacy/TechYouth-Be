package repositories

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lac-hong-legacy/ven_api/model"
	"gorm.io/gorm"
)

type RateLimitRepository struct {
	BaseRepository
}

func NewRateLimitRepository(db *gorm.DB) *RateLimitRepository {
	return &RateLimitRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (s *RateLimitRepository) GetRateLimit(identifier, endpointType string) (*model.RateLimit, error) {
	var rateLimit model.RateLimit

	err := s.db.Where("identifier = ? AND endpoint_type = ?", identifier, endpointType).First(&rateLimit).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &rateLimit, nil
}

func (s *RateLimitRepository) SaveRateLimit(rateLimit *model.RateLimit) error {
	// Generate ID if not set
	if rateLimit.ID == "" {
		id, _ := uuid.NewV7()
		rateLimit.ID = id.String()
	}

	// Set timestamps if not set
	now := time.Now()
	if rateLimit.CreatedAt.IsZero() {
		rateLimit.CreatedAt = now
	}
	rateLimit.UpdatedAt = now

	// Use GORM's Save method which will INSERT or UPDATE based on primary key
	if err := s.db.Save(rateLimit).Error; err != nil {
		return err
	}
	return nil
}

func (s *RateLimitRepository) UpdateRateLimit(rateLimit *model.RateLimit) error {
	// Update specific fields using GORM's Updates method
	err := s.db.Model(rateLimit).Where("id = ?", rateLimit.ID).Updates(map[string]interface{}{
		"request_count": rateLimit.RequestCount,
		"blocked_until": rateLimit.BlockedUntil,
		"updated_at":    rateLimit.UpdatedAt,
	}).Error

	return err
}

// Cleanup old rate limit records
func (s *RateLimitRepository) CleanupOldRecords() error {
	// Remove records older than 7 days and not currently blocked
	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	now := time.Now()

	err := s.db.Where("created_at < ? AND (blocked_until IS NULL OR blocked_until < ?)", cutoff, now).
		Delete(&model.RateLimit{}).Error

	return err
}
