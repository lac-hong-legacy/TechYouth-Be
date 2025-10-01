package dto

import "github.com/lac-hong-legacy/ven_api/model"

type CreateSessionRequest struct {
	DeviceID string `json:"device_id" validate:"required,min=1,max=100"`
}

func (c CreateSessionRequest) Validate() error {
	return GetValidator().Struct(c)
}

type CreateSessionResponse struct {
	Session  *model.GuestSession  `json:"session"`
	Progress *model.GuestProgress `json:"progress"`
}

type CompleteLessonRequest struct {
	LessonID  string `json:"lesson_id" validate:"required"`
	Score     int    `json:"score" validate:"required,min=0,max=100"`
	TimeSpent int    `json:"time_spent" validate:"required,min=1"`
}

func (c CompleteLessonRequest) Validate() error {
	return GetValidator().Struct(c)
}
