package location

import (
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	"github.com/efangly/thanes-lims-backend/internal/domain/rbac"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, h *Handler, tokens portuser.TokenService) {
	authGuard := middleware.Auth(tokens)

	locations := r.Group("/locations", authGuard)
	locations.Post("/", middleware.RequirePermission(rbac.ModuleLocation, rbac.ActionCreate), h.CreateCabinet)
	locations.Get("/", middleware.RequirePermission(rbac.ModuleLocation, rbac.ActionView), h.ListChildren)
	locations.Get("/by-barcode/:code", middleware.RequirePermission(rbac.ModuleLocation, rbac.ActionView), h.LookupByBarcode)
	locations.Post("/:id/children", middleware.RequirePermission(rbac.ModuleLocation, rbac.ActionCreate), h.GenerateChildren)
	locations.Post("/:id/boxes", middleware.RequirePermission(rbac.ModuleLocation, rbac.ActionCreate), h.CreateBox)
	locations.Patch("/:id/grid", middleware.RequirePermission(rbac.ModuleLocation, rbac.ActionEdit), h.EnlargeBox)
	locations.Post("/:id/moves", middleware.RequirePermission(rbac.ModuleSample, rbac.ActionEdit), h.MoveWithinBox)
	locations.Get("/:id/full-path", middleware.RequirePermission(rbac.ModuleLocation, rbac.ActionView), h.GetFullPath)
	locations.Get("/:id", middleware.RequirePermission(rbac.ModuleLocation, rbac.ActionView), h.GetLocation)
	locations.Delete("/:id", middleware.RequirePermission(rbac.ModuleLocation, rbac.ActionDelete), h.Delete)
}
