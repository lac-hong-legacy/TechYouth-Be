package dto

import (
	"regexp"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	validate.RegisterValidation("strong_password", validateStrongPassword)
}

func GetValidator() *validator.Validate {
	return validate
}

func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 {
		return false
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasNumber && hasSpecial
}

func ValidateEmailOrUsername(fl validator.FieldLevel) bool {
	value := fl.Field().String()

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)

	return emailRegex.MatchString(value) || usernameRegex.MatchString(value)
}

func FormatValidationErrors(err error) []ValidationError {
	var errors []ValidationError

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			var message string

			switch fieldError.Tag() {
			case "required":
				message = fieldError.Field() + " is required"
			case "email":
				message = "Invalid email format"
			case "min":
				message = fieldError.Field() + " must be at least " + fieldError.Param() + " characters"
			case "max":
				message = fieldError.Field() + " must be at most " + fieldError.Param() + " characters"
			case "len":
				message = fieldError.Field() + " must be exactly " + fieldError.Param() + " characters"
			case "numeric":
				message = fieldError.Field() + " must contain only numbers"
			case "alphanum":
				message = fieldError.Field() + " must contain only letters and numbers"
			case "strong_password":
				message = "Password must contain at least 8 characters with uppercase, lowercase, number, and special character"
			case "url":
				message = fieldError.Field() + " must be a valid URL"
			case "oneof":
				message = fieldError.Field() + " must be one of: " + fieldError.Param()
			case "dive":
				message = fieldError.Field() + " contains invalid items"
			default:
				message = fieldError.Field() + " is invalid"
			}

			errors = append(errors, ValidationError{
				Field:   fieldError.Field(),
				Message: message,
			})
		}
	}

	return errors
}

type Validator interface {
	Validate() error
}

func CreateValidationErrorResponse(err error) ValidationErrorResponse {
	return ValidationErrorResponse{
		Code:    400,
		Message: "Validation failed",
		Errors:  FormatValidationErrors(err),
	}
}
