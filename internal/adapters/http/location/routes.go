package location

import (
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, h *Handler, tokens portuser.TokenService) {
	authGuard := middleware.Auth(tokens)

	locations := r.Group("/locations", authGuard)
	locations.Post("/", h.CreateCabinet)
	locations.Get("/", h.ListChildren)
	locations.Post("/:id/children", h.GenerateChildren)
	locations.Get("/:id/full-path", h.GetFullPath)
	locations.Delete("/:id", h.Delete)
}
