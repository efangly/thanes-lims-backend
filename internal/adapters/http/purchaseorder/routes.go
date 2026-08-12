package purchaseorder

import (
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, h *Handler, tokens portuser.TokenService) {
	authGuard := middleware.Auth(tokens)

	po := r.Group("/purchase-orders", authGuard)
	po.Get("/", h.List)
	po.Get("/:id", h.Get)
	po.Patch("/:id/approve", h.Approve)
	po.Patch("/:id/receive", h.MarkReceived)
}
