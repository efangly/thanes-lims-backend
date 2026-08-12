package environment

import (
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, h *Handler, tokens portuser.TokenService) {
	authGuard := middleware.Auth(tokens)

	env := r.Group("/environment", authGuard)
	env.Get("/gauges", h.ListGauges)
	env.Get("/gauges/:loc/trend", h.GetTrend)
	env.Get("/alerts", h.ListAlerts)
	env.Post("/readings", h.RecordReading)
}
