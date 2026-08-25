package notification

import (
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	"github.com/efangly/thanes-lims-backend/internal/domain/rbac"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, h *Handler, tokens portuser.TokenService) {
	authGuard := middleware.Auth(tokens)

	notif := r.Group("/notifications", authGuard)
	notif.Get("/", middleware.RequirePermission(rbac.ModuleNotification, rbac.ActionView), h.List)
	notif.Patch("/read-all", middleware.RequirePermission(rbac.ModuleNotification, rbac.ActionEdit), h.MarkAllRead)
	notif.Patch("/:id/read", middleware.RequirePermission(rbac.ModuleNotification, rbac.ActionEdit), h.MarkRead)
}
