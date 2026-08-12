package sample

import (
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	"github.com/gofiber/fiber/v3"
)

// RegisterRoutes mounts all Sample + Chain-of-Custody endpoints, all behind
// the auth guard - there is no anonymous access to LIMS data.
func RegisterRoutes(r fiber.Router, h *Handler, tokens portuser.TokenService) {
	authGuard := middleware.Auth(tokens)

	samples := r.Group("/samples", authGuard)
	samples.Post("/", h.Create)
	samples.Get("/", h.List)
	samples.Get("/:id", h.Get)
	samples.Patch("/:id/status", h.UpdateStatus)
	samples.Get("/:id/coc", h.ListCoC)
	samples.Post("/:id/coc", h.AppendCoC)
}
