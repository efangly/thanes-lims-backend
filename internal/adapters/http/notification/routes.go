package notification

import (
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, h *Handler, tokens portuser.TokenService) {
	authGuard := middleware.Auth(tokens)

	notif := r.Group("/notifications", authGuard)
	notif.Get("/", h.List)
	notif.Patch("/read-all", h.MarkAllRead)
	notif.Patch("/:id/read", h.MarkRead)
}
