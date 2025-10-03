package services

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alphabatem/common/context"
	"github.com/google/uuid"
	"github.com/lac-hong-legacy/ven_api/dto"
	"github.com/lac-hong-legacy/ven_api/model"
	"github.com/lac-hong-legacy/ven_api/shared"
	log "github.com/sirupsen/logrus"
)

type MediaService struct {
	context.DefaultService
	sqlSvc   *SqliteService
	minioSvc *MinIOService
	baseURL  string
}

const MEDIA_SVC = "media_svc"

func (svc MediaService) Id() string {
	return MEDIA_SVC
}

func (svc *MediaService) Configure(ctx *context.Context) error {
	svc.baseURL = os.Getenv("BASE_URL")
	if svc.baseURL == "" {
		svc.baseURL = "http://localhost:8000"
	}

	return svc.DefaultService.Configure(ctx)
}

func (svc *MediaService) Start() error {
	svc.sqlSvc = svc.Service(SQLITE_SVC).(*SqliteService)
	svc.minioSvc = svc.Service(MINIO_SVC).(*MinIOService)
	return nil
}

// ==================== MEDIA UPLOAD METHODS ====================

func (svc *MediaService) UploadLessonVideo(lessonID string, file *multipart.FileHeader) (*dto.MediaUploadResponse, error) {
	// Validate file type
	if !svc.isValidVideoFile(file.Filename) {
		return nil, shared.NewBadRequestError(nil, "Invalid video file format. Supported: MP4, MOV, AVI")
	}

	// Check file size (max 100MB for videos)
	if file.Size > 100*1024*1024 {
		return nil, shared.NewBadRequestError(nil, "Video file too large. Maximum size: 100MB")
	}

	return svc.uploadFile(file, "video", lessonID)
}

func (svc *MediaService) UploadLessonSubtitle(lessonID string, file *multipart.FileHeader) (*dto.MediaUploadResponse, error) {
	if !svc.isValidSubtitleFile(file.Filename) {
		return nil, shared.NewBadRequestError(nil, "Invalid subtitle file format. Supported: VTT, SRT")
	}

	return svc.uploadFile(file, "subtitle", lessonID)
}

func (svc *MediaService) UploadBackgroundMusic(lessonID string, file *multipart.FileHeader) (*dto.MediaUploadResponse, error) {
	if !svc.isValidAudioFile(file.Filename) {
		return nil, shared.NewBadRequestError(nil, "Invalid audio file format. Supported: MP3, WAV, AAC")
	}

	if file.Size > 10*1024*1024 {
		return nil, shared.NewBadRequestError(nil, "Audio file too large. Maximum size: 10MB")
	}

	return svc.uploadFile(file, "background_music", lessonID)
}

func (svc *MediaService) UploadVoiceOver(lessonID string, file *multipart.FileHeader) (*dto.MediaUploadResponse, error) {
	if !svc.isValidAudioFile(file.Filename) {
		return nil, shared.NewBadRequestError(nil, "Invalid audio file format. Supported: MP3, WAV, AAC")
	}

	if file.Size > 20*1024*1024 {
		return nil, shared.NewBadRequestError(nil, "Voice-over file too large. Maximum size: 20MB")
	}

	return svc.uploadFile(file, "voice_over", lessonID)
}

func (svc *MediaService) UploadAnimation(lessonID string, file *multipart.FileHeader) (*dto.MediaUploadResponse, error) {
	if !svc.isValidVideoFile(file.Filename) {
		return nil, shared.NewBadRequestError(nil, "Invalid animation file format. Supported: MP4, MOV, WEBM")
	}

	if file.Size > 50*1024*1024 {
		return nil, shared.NewBadRequestError(nil, "Animation file too large. Maximum size: 50MB")
	}

	return svc.uploadFile(file, "animation", lessonID)
}

func (svc *MediaService) UploadIllustration(lessonID string, file *multipart.FileHeader) (*dto.MediaUploadResponse, error) {
	if !svc.isValidImageFile(file.Filename) {
		return nil, shared.NewBadRequestError(nil, "Invalid image file format. Supported: JPG, PNG, WEBP")
	}

	if file.Size > 5*1024*1024 {
		return nil, shared.NewBadRequestError(nil, "Image file too large. Maximum size: 5MB")
	}

	return svc.uploadFile(file, "illustration", lessonID)
}

func (svc *MediaService) UploadThumbnail(lessonID string, file *multipart.FileHeader) (*dto.MediaUploadResponse, error) {
	if !svc.isValidImageFile(file.Filename) {
		return nil, shared.NewBadRequestError(nil, "Invalid image file format. Supported: JPG, PNG, WEBP")
	}

	if file.Size > 2*1024*1024 {
		return nil, shared.NewBadRequestError(nil, "Thumbnail file too large. Maximum size: 2MB")
	}

	return svc.uploadFile(file, "thumbnail", lessonID)
}

func (svc *MediaService) uploadFile(file *multipart.FileHeader, fileType, lessonID string) (*dto.MediaUploadResponse, error) {
	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	fileName := fmt.Sprintf("%s_%s_%d%s", lessonID, fileType, time.Now().Unix(), ext)

	// Determine subdirectory based on file type
	var subDir string
	switch fileType {
	case "video":
		subDir = "videos"
	case "subtitle":
		subDir = "subtitles"
	case "thumbnail":
		subDir = "thumbnails"
	case "audio":
		subDir = "audio"
	case "background_music":
		subDir = "background_music"
	case "voice_over":
		subDir = "voice_over"
	case "animation":
		subDir = "animations"
	case "illustration":
		subDir = "illustrations"
	default:
		subDir = "misc"
	}

	// Create object name for MinIO
	objectName := fmt.Sprintf("%s/%s", subDir, fileName)

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		return nil, shared.NewInternalError(err, "Failed to open uploaded file")
	}
	defer src.Close()

	// Upload to MinIO
	uploadInfo, err := svc.minioSvc.UploadFile(objectName, src, file.Size, file.Header.Get("Content-Type"))
	if err != nil {
		return nil, shared.NewInternalError(err, "Failed to upload file to storage")
	}

	// Generate presigned URL (valid for 24 hours)
	fileURL, err := svc.minioSvc.GetFileURL(objectName, 24*time.Hour)
	if err != nil {
		log.Printf("Failed to generate presigned URL: %v", err)
		fileURL = fmt.Sprintf("%s/%s/%s", svc.baseURL, svc.minioSvc.GetBucketName(), objectName)
	}

	id, _ := uuid.NewV7()

	// Create media asset record
	mediaAsset := &model.MediaAsset{
		ID:           id.String(),
		FileName:     fileName,
		OriginalName: file.Filename,
		FileType:     fileType,
		MimeType:     file.Header.Get("Content-Type"),
		FileSize:     file.Size,
		URL:          fileURL,
		StoragePath:  objectName,
		IsProcessed:  false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save to database
	if err := svc.sqlSvc.CreateMediaAsset(mediaAsset); err != nil {
		// Clean up file if database save fails
		svc.minioSvc.DeleteFile(objectName)
		return nil, err
	}

	// Link to lesson if lessonID provided
	if lessonID != "" {
		lessonMedia := &model.LessonMedia{
			ID:           id.String(),
			LessonID:     lessonID,
			MediaAssetID: mediaAsset.ID,
			MediaType:    fileType,
			IsActive:     true,
			CreatedAt:    time.Now(),
		}

		if err := svc.sqlSvc.CreateLessonMedia(lessonMedia); err != nil {
			log.Printf("Failed to link media to lesson: %v", err)
		}
	}

	log.Printf("Successfully uploaded file %s to MinIO: %s", fileName, uploadInfo.Key)

	return &dto.MediaUploadResponse{
		ID:       mediaAsset.ID,
		URL:      mediaAsset.URL,
		FileName: mediaAsset.FileName,
		FileType: mediaAsset.FileType,
		FileSize: mediaAsset.FileSize,
	}, nil
}

// ==================== MEDIA RETRIEVAL METHODS ====================

func (svc *MediaService) GetLessonMedia(lessonID string) (*dto.LessonMediaResponse, error) {
	mediaAssets, err := svc.sqlSvc.GetLessonMediaAssets(lessonID)
	if err != nil {
		return nil, err
	}

	response := &dto.LessonMediaResponse{
		LessonID: lessonID,
		Media:    make(map[string]*dto.MediaAssetResponse),
	}

	for _, asset := range mediaAssets {
		response.Media[asset.MediaType] = &dto.MediaAssetResponse{
			ID:       asset.MediaAsset.ID,
			URL:      asset.MediaAsset.URL,
			Duration: asset.MediaAsset.Duration,
			FileSize: asset.MediaAsset.FileSize,
		}
	}

	return response, nil
}

// ==================== FILE VALIDATION METHODS ====================

func (svc *MediaService) isValidVideoFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExts := []string{".mp4", ".mov", ".avi", ".mkv", ".webm"}

	for _, validExt := range validExts {
		if ext == validExt {
			return true
		}
	}
	return false
}

func (svc *MediaService) isValidSubtitleFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExts := []string{".vtt", ".srt", ".ass", ".ssa"}

	for _, validExt := range validExts {
		if ext == validExt {
			return true
		}
	}
	return false
}

func (svc *MediaService) isValidAudioFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExts := []string{".mp3", ".wav", ".aac", ".m4a", ".ogg"}

	for _, validExt := range validExts {
		if ext == validExt {
			return true
		}
	}
	return false
}

func (svc *MediaService) isValidImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExts := []string{".jpg", ".jpeg", ".png", ".webp", ".gif"}

	for _, validExt := range validExts {
		if ext == validExt {
			return true
		}
	}
	return false
}

// ==================== MEDIA PROCESSING METHODS ====================

func (svc *MediaService) ProcessVideoMetadata(mediaAssetID string) error {
	// TODO: Implement video metadata extraction
	// - Get video duration, resolution, bitrate
	// - Generate thumbnail
	// - Validate video integrity
	// - Update MediaAsset with metadata

	log.Printf("Processing video metadata for asset %s", mediaAssetID)
	return nil
}

func (svc *MediaService) GenerateVideoThumbnail(mediaAssetID string) error {
	// TODO: Implement thumbnail generation
	// - Extract frame at 10% of video duration
	// - Resize to standard thumbnail size (320x180)
	// - Save as JPEG
	// - Update MediaAsset with thumbnail URL

	log.Printf("Generating thumbnail for video asset %s", mediaAssetID)
	return nil
}

// ==================== CLEANUP METHODS ====================

func (svc *MediaService) DeleteMediaAsset(mediaAssetID string) error {
	asset, err := svc.sqlSvc.GetMediaAsset(mediaAssetID)
	if err != nil {
		return err
	}

	// Delete file from MinIO
	if err := svc.minioSvc.DeleteFile(asset.StoragePath); err != nil {
		log.Printf("Failed to delete file from MinIO %s: %v", asset.StoragePath, err)
	}

	// Delete database records
	return svc.sqlSvc.DeleteMediaAsset(mediaAssetID)
}
