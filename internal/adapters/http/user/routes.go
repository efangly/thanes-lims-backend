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

	// Self-service (any authenticated User, own record only - see
	// CONTEXT.md "Self-service"). Registered before the /users/:id routes so
	// the literal "me" segment wins over the :id param.
	r.Get("/users/me", authGuard, h.Me)
	r.Patch("/users/me", authGuard, h.UpdateProfile)
	r.Post("/users/me/password", authGuard, h.ChangePassword)

	users := r.Group("/users", authGuard)
	users.Get("/", middleware.RequirePermission(rbac.ModuleUser, rbac.ActionView), h.ListUsers)
	users.Post("/", middleware.RequirePermission(rbac.ModuleUser, rbac.ActionCreate), h.CreateUser)
	users.Patch("/:id", middleware.RequirePermission(rbac.ModuleUser, rbac.ActionEdit), h.UpdateUser)
	users.Delete("/:id", middleware.RequirePermission(rbac.ModuleUser, rbac.ActionDelete), h.RetireUser)
	users.Post("/:id/suspend", middleware.RequirePermission(rbac.ModuleUser, rbac.ActionEdit), h.SuspendUser)
	users.Post("/:id/reactivate", middleware.RequirePermission(rbac.ModuleUser, rbac.ActionEdit), h.ReactivateUser)
	users.Post("/:id/reset-password", middleware.RequirePermission(rbac.ModuleUser, rbac.ActionEdit), h.ResetPassword)
}
