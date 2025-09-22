package dto

import "github.com/lac-hong-legacy/TechYouth-Be/model"

type CreateSessionRequest struct {
	DeviceID string `json:"device_id" binding:"required"`
}

type CreateSessionResponse struct {
	Session  *model.GuestSession  `json:"session"`
	Progress *model.GuestProgress `json:"progress"`
}

type CompleteLessonRequest struct {
	LessonID  string `json:"lesson_id" binding:"required"`
	Score     int    `json:"score" binding:"required"`
	TimeSpent int    `json:"time_spent" binding:"required"`
}
