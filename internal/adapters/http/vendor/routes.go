package vendor

import (
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	"github.com/efangly/thanes-lims-backend/internal/domain/rbac"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, h *Handler, tokens portuser.TokenService) {
	authGuard := middleware.Auth(tokens)

	v := r.Group("/vendors", authGuard)
	v.Post("/", middleware.RequirePermission(rbac.ModuleVendor, rbac.ActionCreate), h.Create)
	v.Get("/", middleware.RequirePermission(rbac.ModuleVendor, rbac.ActionView), h.List)
	v.Get("/:id", middleware.RequirePermission(rbac.ModuleVendor, rbac.ActionView), h.Get)
	v.Patch("/:id", middleware.RequirePermission(rbac.ModuleVendor, rbac.ActionEdit), h.Update)
}
