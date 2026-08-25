package user

import (
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	"github.com/efangly/thanes-lims-backend/internal/domain/rbac"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	"github.com/gofiber/fiber/v3"
)

// RegisterRoutes mounts auth (public) and user-management (protected)
// endpoints onto the given router group.
func RegisterRoutes(r fiber.Router, h *Handler, tokens portuser.TokenService) {
	authGuard := middleware.Auth(tokens)
	csrf := middleware.RequireCSRFHeader()

	auth := r.Group("/auth")
	auth.Post("/login", h.Login)
	auth.Post("/refresh", csrf, h.Refresh)
	auth.Post("/logout", csrf, h.Logout)
	auth.Post("/logout-all", authGuard, h.LogoutAll)

	r.Get("/users/me", authGuard, h.Me)

	users := r.Group("/users", authGuard)
	users.Get("/", middleware.RequirePermission(rbac.ModuleUser, rbac.ActionView), h.ListUsers)
	users.Post("/", middleware.RequirePermission(rbac.ModuleUser, rbac.ActionCreate), h.CreateUser)
	users.Patch("/:id", middleware.RequirePermission(rbac.ModuleUser, rbac.ActionEdit), h.UpdateUser)
}
