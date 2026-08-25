package environment

import (
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	"github.com/efangly/thanes-lims-backend/internal/domain/rbac"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, h *Handler, tokens portuser.TokenService) {
	authGuard := middleware.Auth(tokens)
	requireView := middleware.RequirePermission(rbac.ModuleEnvironment, rbac.ActionView)

	// Middleware is applied per-route (not group-level) because a
	// group-level middleware in Fiber v3 is mounted against the prefix
	// itself and would also run on /alerts/ws below, which needs a
	// different auth check (query token, not header) since a browser's WS
	// handshake can't carry an Authorization header.
	env := r.Group("/environment")
	env.Get("/gauges", authGuard, requireView, h.ListGauges)
	env.Get("/gauges/:loc/trend", authGuard, requireView, h.GetTrend)
	env.Get("/alerts", authGuard, requireView, h.ListAlerts)
	env.Post("/readings", authGuard, middleware.RequirePermission(rbac.ModuleEnvironment, rbac.ActionCreate), h.RecordReading)

	env.Get("/alerts/ws", middleware.AuthQuery(tokens), requireView, h.AlertsWS)
}
