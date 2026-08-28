package inventory

import (
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	"github.com/efangly/thanes-lims-backend/internal/domain/rbac"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, h *Handler, tokens portuser.TokenService) {
	authGuard := middleware.Auth(tokens)

	inv := r.Group("/inventory", authGuard)
	inv.Post("/", middleware.RequirePermission(rbac.ModuleInventory, rbac.ActionCreate), h.Create)
	inv.Get("/", middleware.RequirePermission(rbac.ModuleInventory, rbac.ActionView), h.List)
	inv.Get("/:id", middleware.RequirePermission(rbac.ModuleInventory, rbac.ActionView), h.Get)
	inv.Patch("/:id", middleware.RequirePermission(rbac.ModuleInventory, rbac.ActionEdit), h.Update)
	inv.Get("/:id/lots", middleware.RequirePermission(rbac.ModuleInventory, rbac.ActionView), h.ListLots)
	inv.Post("/:id/receive", middleware.RequirePermission(rbac.ModuleInventory, rbac.ActionEdit), h.ReceiveStock)
	inv.Post("/:id/issue", middleware.RequirePermission(rbac.ModuleInventory, rbac.ActionEdit), h.IssueStock)
	inv.Patch("/:id/default-vendor", middleware.RequirePermission(rbac.ModuleInventory, rbac.ActionEdit), h.UpdateDefaultVendor)
	inv.Post("/:id/reorder", middleware.RequirePermission(rbac.ModuleInventory, rbac.ActionCreate), h.Reorder)
}
