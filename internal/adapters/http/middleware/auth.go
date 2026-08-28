package middleware

import (
	"strings"

	"github.com/efangly/thanes-lims-backend/internal/domain/rbac"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	"github.com/gofiber/fiber/v3"
)

type localsKey string

const (
	LocalsUserID localsKey = "auth_user_id"
	LocalsName   localsKey = "auth_name"
	LocalsRole   localsKey = "auth_role"
	// LocalsPermissions holds the caller's JWT-embedded Permission set
	// (compact "module:action" strings - see ADR 0002) for RequirePermission
	// to check without a DB call.
	LocalsPermissions localsKey = "auth_permissions"
)

// Auth validates the Bearer access token on every route it's mounted on and
// stores the resulting claims in Locals for downstream handlers/middleware
// (RequirePermission, the audit middleware) to read.
func Auth(tokens portuser.TokenService) fiber.Handler {
	return func(c fiber.Ctx) error {
		header := c.Get(fiber.HeaderAuthorization)
		const prefix = "Bearer "
		if !strings.HasPrefix(header, prefix) {
			return fiber.NewError(fiber.StatusUnauthorized, "missing bearer token")
		}
		return authenticate(c, tokens, strings.TrimPrefix(header, prefix))
	}
}

// AuthQuery validates the access token passed as a `token` query parameter,
// for routes a browser can't attach an Authorization header to - namely the
// WebSocket upgrade handshake. A token in the query string leaks into proxy
// access logs and browser history, so this is deliberately restricted to
// genuine WebSocket upgrade requests; every other route must use header
// Bearer auth via Auth. (A short-lived single-use ticket token issued from an
// authenticated endpoint and exchanged here would remove the leak entirely -
// tracked as a follow-up.)
func AuthQuery(tokens portuser.TokenService) fiber.Handler {
	return func(c fiber.Ctx) error {
		if !strings.Contains(strings.ToLower(c.Get(fiber.HeaderConnection)), "upgrade") ||
			!strings.EqualFold(c.Get(fiber.HeaderUpgrade), "websocket") {
			return fiber.NewError(fiber.StatusUpgradeRequired, "query-string auth is only valid for websocket upgrades")
		}
		token := fiber.Query[string](c, "token")
		if token == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing token query parameter")
		}
		return authenticate(c, tokens, token)
	}
}

func authenticate(c fiber.Ctx, tokens portuser.TokenService, rawToken string) error {
	claims, err := tokens.ParseAccessToken(rawToken)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired token")
	}

	c.Locals(LocalsUserID, claims.UserID)
	c.Locals(LocalsName, claims.Name)
	c.Locals(LocalsRole, claims.Role)
	c.Locals(LocalsPermissions, claims.Permissions)
	return c.Next()
}

// RequirePermission gates a route to callers whose JWT-embedded Permission
// set (resolved from the normalized RBAC tables at login/refresh - see ADR
// 0002) includes module:action. Checked against Fiber locals only, no DB
// call per request.
func RequirePermission(module rbac.Module, action rbac.Action) fiber.Handler {
	want := string(module) + ":" + string(action)
	return func(c fiber.Ctx) error {
		perms, _ := c.Locals(LocalsPermissions).([]string)
		for _, p := range perms {
			if p == want {
				return c.Next()
			}
		}
		return fiber.NewError(fiber.StatusForbidden, "insufficient permission")
	}
}
