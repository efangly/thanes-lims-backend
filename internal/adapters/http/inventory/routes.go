package inventory

import (
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, h *Handler, tokens portuser.TokenService) {
	authGuard := middleware.Auth(tokens)

	inv := r.Group("/inventory", authGuard)
	inv.Post("/", h.Create)
	inv.Get("/", h.List)
	inv.Get("/:id", h.Get)
	inv.Patch("/:id/quantity", h.UpdateQuantity)
	inv.Post("/:id/reorder", h.Reorder)
}
