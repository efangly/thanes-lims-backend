package middleware

import (
	"strings"

	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	"github.com/gofiber/fiber/v3"
)

type localsKey string

const (
	LocalsUserID localsKey = "auth_user_id"
	LocalsName   localsKey = "auth_name"
	LocalsRole   localsKey = "auth_role"
)

// Auth validates the Bearer access token on every route it's mounted on and
// stores the resulting claims in Locals for downstream handlers/middleware
// (RequireRole, the audit middleware) to read.
func Auth(tokens portuser.TokenService) fiber.Handler {
	return func(c fiber.Ctx) error {
		header := c.Get(fiber.HeaderAuthorization)
		const prefix = "Bearer "
		if !strings.HasPrefix(header, prefix) {
			return fiber.NewError(fiber.StatusUnauthorized, "missing bearer token")
		}

		claims, err := tokens.ParseAccessToken(strings.TrimPrefix(header, prefix))
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired token")
		}

		c.Locals(LocalsUserID, claims.UserID)
		c.Locals(LocalsName, claims.Name)
		c.Locals(LocalsRole, claims.Role)
		return c.Next()
	}
}

// RequireRole gates a route to a set of roles, checked against the RBAC
// permission each role needs (global-per-role, not module-scoped).
func RequireRole(roles ...domainuser.Role) fiber.Handler {
	allowed := make(map[domainuser.Role]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(c fiber.Ctx) error {
		role, _ := c.Locals(LocalsRole).(domainuser.Role)
		if !allowed[role] {
			return fiber.NewError(fiber.StatusForbidden, "insufficient role")
		}
		return c.Next()
	}
}

// RequirePermission gates a route to roles holding a given permission.
func RequirePermission(perm domainuser.Permission) fiber.Handler {
	return func(c fiber.Ctx) error {
		role, _ := c.Locals(LocalsRole).(domainuser.Role)
		if !role.Can(perm) {
			return fiber.NewError(fiber.StatusForbidden, "insufficient permission")
		}
		return c.Next()
	}
}
