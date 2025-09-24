package shared

import (
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

var jsonAPI = sonic.Config{
	UseNumber:            true,
	EscapeHTML:           false,
	SortMapKeys:          false,
	CompactMarshaler:     true,
	NoQuoteTextMarshaler: true,
	NoNullSliceOrMap:     true,
}.Froze()

var (
	successResponse       = mustMarshal(Response{Code: 200, Message: "Success"})
	createdResponse       = mustMarshal(Response{Code: 201, Message: "Created"})
	notFoundResponse      = mustMarshal(Response{Code: 404, Message: "Not Found"})
	unauthorizedResponse  = mustMarshal(Response{Code: 401, Message: "Unauthorized"})
	badRequestResponse    = mustMarshal(Response{Code: 400, Message: "Bad Request"})
	forbiddenResponse     = mustMarshal(Response{Code: 403, Message: "Forbidden"})
	internalErrorResponse = mustMarshal(Response{Code: 500, Message: "Internal Server Error"})
)

func mustMarshal(v interface{}) []byte {
	b, _ := jsonAPI.Marshal(v)
	return b
}

func ResponseJSON(c *fiber.Ctx, httpCode int, message string, data interface{}) error {
	if data == nil {
		switch httpCode {
		case 200:
			if message == "Success" {
				c.Set("Content-Type", "application/json")
				return c.Status(httpCode).Send(successResponse)
			}
		case 201:
			if message == "Created" {
				c.Set("Content-Type", "application/json")
				return c.Status(httpCode).Send(createdResponse)
			}
		case 400:
			if message == "Bad Request" {
				c.Set("Content-Type", "application/json")
				return c.Status(httpCode).Send(badRequestResponse)
			}
		case 404:
			if message == "Not Found" {
				c.Set("Content-Type", "application/json")
				return c.Status(httpCode).Send(notFoundResponse)
			}
		case 401:
			if message == "Unauthorized" {
				c.Set("Content-Type", "application/json")
				return c.Status(httpCode).Send(unauthorizedResponse)
			}
		case 403:
			if message == "Forbidden" {
				c.Set("Content-Type", "application/json")
				return c.Status(httpCode).Send(forbiddenResponse)
			}
		case 500:
			if message == "Internal Server Error" {
				c.Set("Content-Type", "application/json")
				return c.Status(httpCode).Send(internalErrorResponse)
			}
		}
	}

	response := Response{
		Code:    httpCode,
		Message: message,
		Data:    data,
	}

	return c.Status(httpCode).JSON(response)
}

func ResponseOK(c *fiber.Ctx, data interface{}) error {
	return ResponseJSON(c, 200, "Success", data)
}

func ResponseNotFound(c *fiber.Ctx) error {
	return ResponseJSON(c, 404, "Not Found", nil)
}

func ResponseUnauthorized(c *fiber.Ctx) error {
	return ResponseJSON(c, 401, "Unauthorized", nil)
}

func ResponseBadRequest(c *fiber.Ctx, message string) error {
	if message == "" {
		message = "Bad Request"
	}
	return ResponseJSON(c, 400, message, nil)
}

func ResponseForbidden(c *fiber.Ctx) error {
	return ResponseJSON(c, 403, "Forbidden", nil)
}

func ResponseCreated(c *fiber.Ctx, data interface{}) error {
	return ResponseJSON(c, 201, "Created", data)
}

func ResponseInternalError(c *fiber.Ctx, err error) error {
	return ResponseJSON(c, 500, "Internal Server Error", err)
}
