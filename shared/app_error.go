package shared

import (
	"errors"
	"net/http"
)

type AppError struct {
	Err        error
	StatusCode int
	Message    string
	Code       string // Optional error code for API clients
	Data       interface{} // Optional data to include in response
}

func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return "unknown error"
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

func GetAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

func NewNotFoundError(err error, message string) *AppError {
	if message == "" {
		message = "Not Found"
	}
	return &AppError{
		Err:        err,
		StatusCode: http.StatusNotFound,
		Message:    message,
		Code:       "NOT_FOUND",
	}
}

func NewBadRequestError(err error, message string) *AppError {
	if message == "" {
		message = "Bad Request"
	}
	return &AppError{
		Err:        err,
		StatusCode: http.StatusBadRequest,
		Message:    message,
		Code:       "BAD_REQUEST",
	}
}

func NewUnauthorizedError(err error, message string) *AppError {
	if message == "" {
		message = "Unauthorized"
	}
	return &AppError{
		Err:        err,
		StatusCode: http.StatusUnauthorized,
		Message:    message,
		Code:       "UNAUTHORIZED",
	}
}

func NewForbiddenError(err error, message string) *AppError {
	if message == "" {
		message = "Forbidden"
	}
	return &AppError{
		Err:        err,
		StatusCode: http.StatusForbidden,
		Message:    message,
		Code:       "FORBIDDEN",
	}
}

func NewInternalError(err error, message string) *AppError {
	if message == "" {
		message = "Internal Server Error"
	}
	return &AppError{
		Err:        err,
		StatusCode: http.StatusInternalServerError,
		Message:    message,
		Code:       "INTERNAL_ERROR",
	}
}

func (e *AppError) WithData(data interface{}) *AppError {
	e.Data = data
	return e
}