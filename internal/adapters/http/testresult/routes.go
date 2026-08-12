package testresult

import (
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, h *Handler, tokens portuser.TokenService) {
	authGuard := middleware.Auth(tokens)

	tests := r.Group("/tests", authGuard)
	tests.Post("/", h.Create)
	tests.Get("/", h.List)
	tests.Get("/:id", h.Get)
	tests.Patch("/:id/result", h.SubmitResult)
	tests.Patch("/:id/approve", h.Approve)

	// Nested convenience route, registered independently of the sample
	// module's own route group.
	r.Get("/samples/:id/tests", authGuard, h.ListBySample)
}
