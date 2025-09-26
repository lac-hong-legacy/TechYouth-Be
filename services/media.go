package services

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alphabatem/common/context"
	"github.com/google/uuid"
	"github.com/lac-hong-legacy/TechYouth-Be/dto"
	"github.com/lac-hong-legacy/TechYouth-Be/model"
	"github.com/lac-hong-legacy/TechYouth-Be/shared"
	log "github.com/sirupsen/logrus"
)

type MediaService struct {
	context.DefaultService
	sqlSvc     *SqliteService
	uploadPath string
	baseURL    string
}

const MEDIA_SVC = "media_svc"

func (svc MediaService) Id() string {
	return MEDIA_SVC
}

func (svc *MediaService) Configure(ctx *context.Context) error {
	svc.uploadPath = os.Getenv("UPLOAD_PATH")
	if svc.uploadPath == "" {
		svc.uploadPath = "./uploads"
	}

	svc.baseURL = os.Getenv("BASE_URL")
	if svc.baseURL == "" {
		svc.baseURL = "http://localhost:8000"
	}

	// Create upload directories
	dirs := []string{"videos", "subtitles", "thumbnails"}
	for _, dir := range dirs {
		fullPath := filepath.Join(svc.uploadPath, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return fmt.Errorf("failed to create upload directory %s: %v", fullPath, err)
		}
	}

	return svc.DefaultService.Configure(ctx)
}

func (svc *MediaService) Start() error {
	svc.sqlSvc = svc.Service(SQLITE_SVC).(*SqliteService)
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

func (svc *MediaService) uploadFile(file *multipart.FileHeader, fileType, lessonID string) (*dto.MediaUploadResponse, error) {
	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	fileName := fmt.Sprintf("%s_%s_%d%s", lessonID, fileType, time.Now().Unix(), ext)

	// Determine subdirectory
	var subDir string
	switch fileType {
	case "video":
		subDir = "videos"
	case "subtitle":
		subDir = "subtitles"
	case "thumbnail":
		subDir = "thumbnails"
	default:
		subDir = "misc"
	}

	// Full file path
	filePath := filepath.Join(svc.uploadPath, subDir, fileName)

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		return nil, shared.NewInternalError(err, "Failed to open uploaded file")
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, shared.NewInternalError(err, "Failed to create destination file")
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, src); err != nil {
		return nil, shared.NewInternalError(err, "Failed to save file")
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
		URL:          fmt.Sprintf("%s/uploads/%s/%s", svc.baseURL, subDir, fileName),
		StoragePath:  filePath,
		IsProcessed:  false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save to database
	if err := svc.sqlSvc.CreateMediaAsset(mediaAsset); err != nil {
		// Clean up file if database save fails
		os.Remove(filePath)
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

	// Delete physical file
	if err := os.Remove(asset.StoragePath); err != nil {
		log.Printf("Failed to delete physical file %s: %v", asset.StoragePath, err)
	}

	// Delete database records
	return svc.sqlSvc.DeleteMediaAsset(mediaAssetID)
}
