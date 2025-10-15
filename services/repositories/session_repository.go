package repositories

import (
	"github.com/google/uuid"
	"github.com/lac-hong-legacy/ven_api/model"
	"gorm.io/gorm"
)

// SessionRepository handles user session and device-related database operations
type SessionRepository struct {
	BaseRepository
}

func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (ds *SessionRepository) GetSessionByDeviceID(deviceID string) (*model.GuestSession, error) {
	var session model.GuestSession
	if err := ds.db.Where("device_id = ?", deviceID).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (ds *SessionRepository) CreateSession(session *model.GuestSession) (*model.GuestSession, error) {
	id, _ := uuid.NewV7()
	session.ID = id.String()
	if err := ds.db.Create(session).Error; err != nil {
		return nil, err
	}
	return session, nil
}

func (ds *SessionRepository) UpdateSession(session *model.GuestSession) error {
	if err := ds.db.Save(session).Error; err != nil {
		return err
	}
	return nil
}
