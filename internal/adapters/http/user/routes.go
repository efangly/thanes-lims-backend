package user

import (
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	"github.com/gofiber/fiber/v3"
)

// RegisterRoutes mounts auth (public) and user-management (protected)
// endpoints onto the given router group.
func RegisterRoutes(r fiber.Router, h *Handler, tokens portuser.TokenService) {
	auth := r.Group("/auth")
	auth.Post("/login", h.Login)
	auth.Post("/refresh", h.Refresh)
	auth.Post("/logout", h.Logout)

	authGuard := middleware.Auth(tokens)

	r.Get("/users/me", authGuard, h.Me)

	admin := r.Group("/users", authGuard, middleware.RequireRole(domainuser.RoleAdmin))
	admin.Get("/", h.ListUsers)
	admin.Post("/", h.CreateUser)
	admin.Patch("/:id", h.UpdateUser)
}
