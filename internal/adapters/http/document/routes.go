package document

import (
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	"github.com/efangly/thanes-lims-backend/internal/domain/rbac"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, h *Handler, tokens portuser.TokenService) {
	authGuard := middleware.Auth(tokens)

	docs := r.Group("/documents", authGuard)
	docs.Post("/", middleware.RequirePermission(rbac.ModuleDocument, rbac.ActionCreate), h.Upload)
	docs.Get("/", middleware.RequirePermission(rbac.ModuleDocument, rbac.ActionView), h.List)
	docs.Get("/:id", middleware.RequirePermission(rbac.ModuleDocument, rbac.ActionView), h.Get)
	docs.Get("/:id/download", middleware.RequirePermission(rbac.ModuleDocument, rbac.ActionView), h.Download)
	docs.Post("/:id/versions", middleware.RequirePermission(rbac.ModuleDocument, rbac.ActionCreate), h.NewVersion)
	docs.Get("/:id/history", middleware.RequirePermission(rbac.ModuleDocument, rbac.ActionView), h.History)
	docs.Patch("/:id/lock", middleware.RequirePermission(rbac.ModuleDocument, rbac.ActionEdit), h.SetLock)
}
