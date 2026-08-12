package equipment

import (
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, h *Handler, tokens portuser.TokenService) {
	authGuard := middleware.Auth(tokens)

	eq := r.Group("/equipment", authGuard)
	eq.Post("/", h.Create)
	eq.Get("/", h.List)
	eq.Get("/:id", h.Get)
	eq.Patch("/:id/calibration", h.RecordCalibration)
}
