package repositories

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/lac-hong-legacy/ven_api/model"
	"gorm.io/gorm"
)

type MediaRepository struct {
	BaseRepository
}

func NewMediaRepository(db *gorm.DB) *MediaRepository {
	return &MediaRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (ds *MediaRepository) CreateMediaAsset(asset *model.MediaAsset) error {
	if asset.ID == "" {
		id, _ := uuid.NewV7()
		asset.ID = id.String()
	}
	asset.CreatedAt = time.Now()
	asset.UpdatedAt = time.Now()

	if err := ds.db.Create(asset).Error; err != nil {
		return err
	}
	return nil
}

func (ds *MediaRepository) GetMediaAsset(id string) (*model.MediaAsset, error) {
	var asset model.MediaAsset
	if err := ds.db.Where("id = ?", id).First(&asset).Error; err != nil {
		return nil, err
	}
	return &asset, nil
}

func (ds *MediaRepository) UpdateMediaAsset(asset *model.MediaAsset) error {
	asset.UpdatedAt = time.Now()
	if err := ds.db.Save(asset).Error; err != nil {
		return err
	}
	return nil
}

func (ds *MediaRepository) DeleteMediaAsset(id string) error {
	// Delete related lesson media records first
	if err := ds.db.Where("media_asset_id = ?", id).Delete(&model.LessonMedia{}).Error; err != nil {
		return err
	}

	// Delete the media asset
	if err := ds.db.Where("id = ?", id).Delete(&model.MediaAsset{}).Error; err != nil {
		return err
	}
	return nil
}

func (ds *MediaRepository) GetMediaAssetsByType(fileType string) ([]model.MediaAsset, error) {
	var assets []model.MediaAsset
	if err := ds.db.Where("file_type = ?", fileType).Find(&assets).Error; err != nil {
		return nil, err
	}
	return assets, nil
}

func (ds *MediaRepository) GetUnprocessedMediaAssets() ([]model.MediaAsset, error) {
	var assets []model.MediaAsset
	if err := ds.db.Where("is_processed = ?", false).Find(&assets).Error; err != nil {
		return nil, err
	}
	return assets, nil
}

// ==================== LESSON MEDIA METHODS ====================

func (ds *MediaRepository) CreateLessonMedia(lessonMedia *model.LessonMedia) error {
	if lessonMedia.ID == "" {
		id, _ := uuid.NewV7()
		lessonMedia.ID = id.String()
	}
	lessonMedia.CreatedAt = time.Now()

	if err := ds.db.Create(lessonMedia).Error; err != nil {
		return err
	}
	return nil
}

func (ds *MediaRepository) GetLessonMediaAssets(lessonID string) ([]model.LessonMedia, error) {
	var lessonMedia []model.LessonMedia
	if err := ds.db.Where("lesson_id = ? AND is_active = ?", lessonID, true).
		Preload("MediaAsset").
		Find(&lessonMedia).Error; err != nil {
		return nil, err
	}
	return lessonMedia, nil
}

func (ds *MediaRepository) GetLessonMediaByType(lessonID, mediaType string) (*model.LessonMedia, error) {
	var lessonMedia model.LessonMedia
	if err := ds.db.Where("lesson_id = ? AND media_type = ? AND is_active = ?", lessonID, mediaType, true).
		Preload("MediaAsset").
		First(&lessonMedia).Error; err != nil {
		return nil, err
	}
	return &lessonMedia, nil
}

func (ds *MediaRepository) UpdateLessonMedia(lessonMedia *model.LessonMedia) error {
	if err := ds.db.Save(lessonMedia).Error; err != nil {
		return err
	}
	return nil
}

func (ds *MediaRepository) DeleteLessonMedia(id string) error {
	if err := ds.db.Where("id = ?", id).Delete(&model.LessonMedia{}).Error; err != nil {
		return err
	}
	return nil
}

func (ds *MediaRepository) DeactivateLessonMediaByType(lessonID, mediaType string) error {
	if err := ds.db.Model(&model.LessonMedia{}).
		Where("lesson_id = ? AND media_type = ?", lessonID, mediaType).
		Update("is_active", false).Error; err != nil {
		return err
	}
	return nil
}

func (ds *MediaRepository) GetMediaStatistics() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Count total media assets by type
	var videoCount, subtitleCount, thumbnailCount int64

	ds.db.Model(&model.MediaAsset{}).Where("file_type = ?", "video").Count(&videoCount)
	ds.db.Model(&model.MediaAsset{}).Where("file_type = ?", "subtitle").Count(&subtitleCount)
	ds.db.Model(&model.MediaAsset{}).Where("file_type = ?", "thumbnail").Count(&thumbnailCount)

	stats["total_videos"] = videoCount
	stats["total_subtitles"] = subtitleCount
	stats["total_thumbnails"] = thumbnailCount

	// Calculate total storage used
	var totalSize int64
	ds.db.Model(&model.MediaAsset{}).Select("COALESCE(SUM(file_size), 0)").Scan(&totalSize)
	stats["total_storage_bytes"] = totalSize
	stats["total_storage_mb"] = totalSize / (1024 * 1024)

	// Count lessons with media
	var lessonsWithVideo, lessonsWithSubtitles int64
	ds.db.Model(&model.LessonMedia{}).
		Where("media_type = ? AND is_active = ?", "video", true).
		Count(&lessonsWithVideo)
	ds.db.Model(&model.LessonMedia{}).
		Where("media_type = ? AND is_active = ?", "subtitle", true).
		Count(&lessonsWithSubtitles)

	stats["lessons_with_video"] = lessonsWithVideo
	stats["lessons_with_subtitles"] = lessonsWithSubtitles

	// Count unprocessed media
	var unprocessedCount int64
	ds.db.Model(&model.MediaAsset{}).Where("is_processed = ?", false).Count(&unprocessedCount)
	stats["unprocessed_media"] = unprocessedCount

	return stats, nil
}

func (ds *MediaRepository) GetLessonsWithoutMedia() ([]model.Lesson, error) {
	var lessons []model.Lesson

	// Find lessons that don't have any active media
	subQuery := ds.db.Model(&model.LessonMedia{}).
		Select("lesson_id").
		Where("is_active = ?", true)

	if err := ds.db.Where("id NOT IN (?)", subQuery).
		Preload("Character").
		Find(&lessons).Error; err != nil {
		return nil, err
	}

	return lessons, nil
}

// ==================== BULK OPERATIONS ====================

func (ds *MediaRepository) BulkCreateMediaAssets(assets []model.MediaAsset) error {
	if len(assets) == 0 {
		return nil
	}

	// Set IDs and timestamps
	now := time.Now()
	for i := range assets {
		if assets[i].ID == "" {
			id, _ := uuid.NewV7()
			assets[i].ID = id.String()
		}
		assets[i].CreatedAt = now
		assets[i].UpdatedAt = now
	}

	if err := ds.db.CreateInBatches(assets, 100).Error; err != nil {
		return err
	}
	return nil
}

func (ds *MediaRepository) BulkCreateLessonMedia(lessonMedia []model.LessonMedia) error {
	if len(lessonMedia) == 0 {
		return nil
	}

	// Set IDs and timestamps
	now := time.Now()
	for i := range lessonMedia {
		if lessonMedia[i].ID == "" {
			id, _ := uuid.NewV7()
			lessonMedia[i].ID = id.String()
		}
		lessonMedia[i].CreatedAt = now
	}

	if err := ds.db.CreateInBatches(lessonMedia, 100).Error; err != nil {
		return err
	}
	return nil
}

func (ds *MediaRepository) GetLessonWithMedia(id string) (*model.Lesson, []model.LessonMedia, error) {
	var lesson model.Lesson
	if err := ds.db.Where("id = ?", id).Preload("Character").First(&lesson).Error; err != nil {
		return nil, nil, err
	}

	mediaAssets, err := ds.GetLessonMediaAssets(id)
	if err != nil {
		return &lesson, nil, err
	}

	return &lesson, mediaAssets, nil
}

func (ds *MediaRepository) GetLessonsWithMediaByCharacter(characterID string) ([]model.Lesson, map[string][]model.LessonMedia, error) {
	var lessons []model.Lesson
	if err := ds.db.Where("character_id = ? AND is_active = ?", characterID, true).
		Order("\"order\" ASC").
		Find(&lessons).Error; err != nil {
		return nil, nil, err
	}

	// Get media for all lessons
	mediaMap := make(map[string][]model.LessonMedia)
	for _, lesson := range lessons {
		media, err := ds.GetLessonMediaAssets(lesson.ID)
		if err != nil {
			log.Printf("Failed to get media for lesson %s: %v", lesson.ID, err)
			continue
		}
		mediaMap[lesson.ID] = media
	}

	return lessons, mediaMap, nil
}
