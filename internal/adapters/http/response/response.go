package response

import "github.com/gofiber/fiber/v3"

// Envelope is the standard shape every endpoint responds with:
// {success, data, error: {code, message}, meta?}
type Envelope struct {
	Success bool       `json:"success"`
	Data    any        `json:"data,omitempty"`
	Error   *ErrorBody `json:"error,omitempty"`
	Meta    any        `json:"meta,omitempty"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// OK writes a 200 success envelope. Pass meta to include pagination or
// similar side-channel info.
func OK(c fiber.Ctx, data any, meta ...any) error {
	env := Envelope{Success: true, Data: data}
	if len(meta) > 0 {
		env.Meta = meta[0]
	}
	return c.JSON(env)
}

// Created writes a 201 success envelope.
func Created(c fiber.Ctx, data any) error {
	return c.Status(fiber.StatusCreated).JSON(Envelope{Success: true, Data: data})
}

// Error writes a failure envelope with the given HTTP status and error code.
func Error(c fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(Envelope{
		Success: false,
		Error:   &ErrorBody{Code: code, Message: message},
	})
}
