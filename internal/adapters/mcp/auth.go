// Package mcp adapts the existing Sample/TestResult/Inventory/PurchaseOrder
// repositories to a read-only Model Context Protocol server, so an external
// NestJS+LangGraph.js chatbot service can query Postgres (the real system of
// record) instead of the old Oracle ADB mirror (docs/adr/0013-mcp-server-transport.md).
package mcp

import (
	"context"
	"crypto/subtle"
	"net/http"
	"strings"

	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
)

type ctxKey string

const claimsCtxKey ctxKey = "mcp_claims"

// ClaimsFromContext returns the caller's JWT claims, previously stashed by
// AuthMiddleware. Tool handlers and the permission wrapper use this instead
// of re-parsing the token.
func ClaimsFromContext(ctx context.Context) (portuser.Claims, bool) {
	c, ok := ctx.Value(claimsCtxKey).(portuser.Claims)
	return c, ok
}

// AuthMiddleware wraps the MCP Streamable HTTP handler with two independent,
// mandatory checks (defense in depth - see the ADR):
//
//  1. A static service-to-service API key (X-Service-Api-Key), compared with
//     constant-time equality against serviceAPIKey, so only the NestJS
//     service itself (not an arbitrary bearer of an end-user JWT) can reach
//     this server at all.
//  2. The forwarded end-user access token (standard "Authorization: Bearer
//     <jwt>" header), parsed with the same HS256 verification
//     internal/adapters/jwt uses. Its claims (user_id/role/permissions) are
//     stashed in the request context for RequirePermission and tool handlers.
//
// Neither check alone is sufficient: the service key proves "this request
// came from the NestJS gateway", the JWT proves "on behalf of this
// authenticated, still-permissioned end user". Both must pass before any MCP
// method (including tools/list) is served.
func AuthMiddleware(next http.Handler, tokens portuser.TokenService, serviceAPIKey string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		presentedKey := r.Header.Get("X-Service-Api-Key")
		if serviceAPIKey == "" || subtle.ConstantTimeCompare([]byte(presentedKey), []byte(serviceAPIKey)) != 1 {
			http.Error(w, "invalid or missing X-Service-Api-Key", http.StatusUnauthorized)
			return
		}

		authHeader := r.Header.Get("Authorization")
		const prefix = "Bearer "
		if !strings.HasPrefix(authHeader, prefix) {
			http.Error(w, "missing bearer token", http.StatusUnauthorized)
			return
		}

		claims, err := tokens.ParseAccessToken(strings.TrimPrefix(authHeader, prefix))
		if err != nil {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), claimsCtxKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// HasPermission reports whether claims carries the compact "module:action"
// permission key (see internal/domain/rbac.Permission.Key) - here always
// "chatbot:view".
func HasPermission(claims portuser.Claims, key string) bool {
	for _, p := range claims.Permissions {
		if p == key {
			return true
		}
	}
	return false
}
