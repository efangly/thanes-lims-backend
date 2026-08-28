package sample

import (
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	"github.com/efangly/thanes-lims-backend/internal/domain/rbac"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	"github.com/gofiber/fiber/v3"
)

// RegisterRoutes mounts all Sample + Chain-of-Custody endpoints, all behind
// the auth guard - there is no anonymous access to LIMS data.
func RegisterRoutes(r fiber.Router, h *Handler, tokens portuser.TokenService) {
	authGuard := middleware.Auth(tokens)

	samples := r.Group("/samples", authGuard)
	samples.Post("/", middleware.RequirePermission(rbac.ModuleSample, rbac.ActionCreate), h.Create)
	samples.Get("/", middleware.RequirePermission(rbac.ModuleSample, rbac.ActionView), h.List)
	samples.Get("/:id", middleware.RequirePermission(rbac.ModuleSample, rbac.ActionView), h.Get)
	samples.Get("/:id/sticker", middleware.RequirePermission(rbac.ModuleSample, rbac.ActionView), h.Sticker)
	samples.Patch("/:id/status", middleware.RequirePermission(rbac.ModuleSample, rbac.ActionEdit), h.UpdateStatus)
	samples.Patch("/:id/location", middleware.RequirePermission(rbac.ModuleSample, rbac.ActionEdit), h.AssignLocation)
	samples.Post("/:id/barcode", middleware.RequirePermission(rbac.ModuleSample, rbac.ActionEdit), h.GenerateBarcode)
	samples.Get("/:id/coc", middleware.RequirePermission(rbac.ModuleSample, rbac.ActionView), h.ListCoC)
	samples.Post("/:id/coc", middleware.RequirePermission(rbac.ModuleSample, rbac.ActionCreate), h.AppendCoC)
}
