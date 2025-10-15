package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lac-hong-legacy/ven_api/shared"
)

type MediaHandler struct {
	mediaSvc    MediaServiceInterface
	contentSvc  ContentServiceInterface
}

func NewMediaHandler(mediaSvc MediaServiceInterface, contentSvc ContentServiceInterface) *MediaHandler {
	return &MediaHandler{
		mediaSvc:    mediaSvc,
		contentSvc:  contentSvc,
	}
}

// @Summary Upload Lesson Subtitle (Admin)
// @Description Upload subtitle file for lesson video (Admin only)
// @Tags admin
// @Accept multipart/form-data
// @Produce json
// @Security Bearer
// @Param Authorization header string true "Admin Bearer Token" default(Bearer <admin_token>)
// @Param lessonId path string true "Lesson ID"
// @Param subtitle formData file true "Subtitle file (VTT, SRT)"
// @Success 200 {object} shared.Response{data=dto.MediaUploadResponse}
// @Router /api/v1/admin/lessons/{lessonId}/subtitle [post]
func (h *MediaHandler) UploadLessonSubtitle(c *fiber.Ctx) error {
	lessonID := c.Params("lessonId")

	file, err := c.FormFile("subtitle")
	if err != nil {
		return shared.NewBadRequestError(err, "No subtitle file provided")
	}

	response, err := h.mediaSvc.UploadLessonSubtitle(lessonID, file)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Subtitle uploaded successfully", response)
}

// @Summary Upload Lesson Thumbnail (Admin)
// @Description Upload thumbnail image for lesson (Admin only)
// @Tags admin
// @Accept multipart/form-data
// @Produce json
// @Security Bearer
// @Param Authorization header string true "Admin Bearer Token" default(Bearer <admin_token>)
// @Param lessonId path string true "Lesson ID"
// @Param thumbnail formData file true "Thumbnail file (JPG, PNG, WEBP)"
// @Success 200 {object} shared.Response{data=dto.MediaUploadResponse}
// @Router /api/v1/admin/lessons/{lessonId}/thumbnail [post]
func (h *MediaHandler) UploadThumbnail(c *fiber.Ctx) error {
	lessonID := c.Params("lessonId")

	file, err := c.FormFile("thumbnail")
	if err != nil {
		return shared.NewBadRequestError(err, "No thumbnail file provided")
	}

	response, err := h.mediaSvc.UploadThumbnail(lessonID, file)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Thumbnail uploaded successfully", response)
}

// @Summary Get Lesson Media (Admin)
// @Description Get all media assets for a lesson (Admin only)
// @Tags admin
// @Produce json
// @Security Bearer
// @Param Authorization header string true "Admin Bearer Token" default(Bearer <admin_token>)
// @Param lessonId path string true "Lesson ID"
// @Success 200 {object} shared.Response{data=dto.LessonMediaResponse}
// @Router /api/v1/admin/lessons/{lessonId}/media [get]
func (h *MediaHandler) GetLessonMedia(c *fiber.Ctx) error {
	lessonID := c.Params("lessonId")

	media, err := h.mediaSvc.GetLessonMedia(lessonID)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", media)
}

// @Summary Delete Media Asset (Admin)
// @Description Delete a media asset and its physical file (Admin only)
// @Tags admin
// @Produce json
// @Security Bearer
// @Param Authorization header string true "Admin Bearer Token" default(Bearer <admin_token>)
// @Param assetId path string true "Media Asset ID"
// @Success 200 {object} shared.Response{data=string}
// @Router /api/v1/admin/media/assets/{assetId} [delete]
func (h *MediaHandler) DeleteMediaAsset(c *fiber.Ctx) error {
	assetID := c.Params("assetId")

	err := h.mediaSvc.DeleteMediaAsset(assetID)
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Media asset deleted successfully", "deleted")
}

// @Summary Get Media Statistics (Admin)
// @Description Get statistics about media assets and usage (Admin only)
// @Tags admin
// @Produce json
// @Security Bearer
// @Param Authorization header string true "Admin Bearer Token" default(Bearer <admin_token>)
// @Success 200 {object} shared.Response{data=map[string]interface{}}
// @Router /api/v1/admin/media/statistics [get]
func (h *MediaHandler) GetMediaStatistics(c *fiber.Ctx) error {
	stats, err := h.mediaSvc.GetMediaStatistics()
	if err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Success", stats)
}

// @Summary Upload Lesson Audio (Admin)
// @Description Upload voice-over audio file - Step 2 of production workflow (Admin only)
// @Tags admin,production
// @Accept multipart/form-data
// @Produce json
// @Security Bearer
// @Param Authorization header string true "Admin Bearer Token" default(Bearer <admin_token>)
// @Param lessonId path string true "Lesson ID"
// @Param audio formData file true "Audio file (MP3, WAV, AAC)"
// @Success 200 {object} shared.Response{data=dto.MediaUploadResponse}
// @Router /api/v1/admin/lessons/{lessonId}/audio [post]
func (h *MediaHandler) UploadLessonAudio(c *fiber.Ctx) error {
	lessonID := c.Params("lessonId")

	file, err := c.FormFile("audio")
	if err != nil {
		return shared.NewBadRequestError(err, "No audio file provided")
	}

	response, err := h.mediaSvc.UploadLessonAudio(lessonID, file)
	if err != nil {
		return err
	}

	if err := h.contentSvc.MarkAudioUploaded(lessonID); err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Audio uploaded successfully", response)
}

// @Summary Upload Lesson Animation (Admin)
// @Description Upload animation video file - Step 3 of production workflow (Admin only)
// @Tags admin,production
// @Accept multipart/form-data
// @Produce json
// @Security Bearer
// @Param Authorization header string true "Admin Bearer Token" default(Bearer <admin_token>)
// @Param lessonId path string true "Lesson ID"
// @Param animation formData file true "Animation file (MP4, MOV, WEBM)"
// @Success 200 {object} shared.Response{data=dto.MediaUploadResponse}
// @Router /api/v1/admin/lessons/{lessonId}/animation [post]
func (h *MediaHandler) UploadLessonAnimation(c *fiber.Ctx) error {
	lessonID := c.Params("lessonId")

	file, err := c.FormFile("animation")
	if err != nil {
		return shared.NewBadRequestError(err, "No animation file provided")
	}

	response, err := h.mediaSvc.UploadLessonAnimation(lessonID, file)
	if err != nil {
		return err
	}

	if err := h.contentSvc.MarkAnimationUploaded(lessonID); err != nil {
		return err
	}

	return shared.ResponseJSON(c, fiber.StatusOK, "Animation uploaded successfully", response)
}
