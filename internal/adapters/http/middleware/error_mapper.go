package middleware

import (
	"errors"

	"github.com/efangly/thanes-lims-backend/internal/adapters/http/response"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/gofiber/fiber/v3"
)

// ErrorMapper is wired as fiber.Config.ErrorHandler so every handler can
// just `return err` and get a consistent {success:false, error:{code,message}}
// envelope, keyed off the sentinel domain errors in domain/shared.
func ErrorMapper(c fiber.Ctx, err error) error {
	code, msg := StatusAndCodeForError(err)
	return response.Error(c, code, msg, errorMessage(err, msg))
}

// StatusAndCodeForError maps an error to its (HTTP status, envelope error
// code). Shared between ErrorMapper (writes the response) and the Audit
// middleware (needs the real status for logging) - Fiber's ErrorHandler
// only runs at the very top of the request dispatch, after every
// middleware's own c.Next() has already returned, so Audit can't just read
// c.Response().StatusCode() after calling Next() and expect it to reflect
// an error a downstream handler returned.
func StatusAndCodeForError(err error) (int, string) {
	if err == nil {
		return fiber.StatusOK, ""
	}

	var fe *fiber.Error
	if errors.As(err, &fe) {
		return fe.Code, "http_error"
	}

	switch {
	case errors.Is(err, shared.ErrNotFound):
		return fiber.StatusNotFound, "not_found"
	case errors.Is(err, shared.ErrConflict):
		return fiber.StatusConflict, "conflict"
	case errors.Is(err, shared.ErrValidation):
		return fiber.StatusBadRequest, "validation_failed"
	case errors.Is(err, shared.ErrUnauthorized):
		return fiber.StatusUnauthorized, "unauthorized"
	case errors.Is(err, shared.ErrForbidden):
		return fiber.StatusForbidden, "forbidden"
	default:
		return fiber.StatusInternalServerError, "internal_error"
	}
}

func errorMessage(err error, code string) string {
	if code == "internal_error" {
		return "internal server error"
	}
	var fe *fiber.Error
	if errors.As(err, &fe) {
		return fe.Message
	}
	return err.Error()
}
