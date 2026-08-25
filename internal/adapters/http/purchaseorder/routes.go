package purchaseorder

import (
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	"github.com/efangly/thanes-lims-backend/internal/domain/rbac"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, h *Handler, tokens portuser.TokenService) {
	authGuard := middleware.Auth(tokens)

	po := r.Group("/purchase-orders", authGuard)
	po.Get("/", middleware.RequirePermission(rbac.ModulePurchaseOrder, rbac.ActionView), h.List)
	po.Get("/:id", middleware.RequirePermission(rbac.ModulePurchaseOrder, rbac.ActionView), h.Get)
	po.Patch("/:id/approve", middleware.RequirePermission(rbac.ModulePurchaseOrder, rbac.ActionApprove), h.Approve)
	po.Patch("/:id/receive", middleware.RequirePermission(rbac.ModulePurchaseOrder, rbac.ActionEdit), h.MarkReceived)
}
