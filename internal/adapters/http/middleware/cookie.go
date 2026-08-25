package middleware

import "github.com/gofiber/fiber/v3"

const (
	// RefreshCookieName is the httpOnly cookie carrying the Refresh Token
	// (see ADR 0004). Scoped to RefreshCookiePath only.
	RefreshCookieName = "refresh_token"
	// RefreshCookiePath restricts the Refresh Cookie to the auth endpoints
	// that actually need it, so it's never sent on unrelated API requests.
	RefreshCookiePath = "/api/v1/auth"
	// CSRFHeaderName must be present on cookie-authenticated calls to
	// /auth/refresh and /auth/logout - a cross-site form submission can
	// trigger the ambient cookie but can't set a custom header (see ADR 0004).
	CSRFHeaderName = "X-SMLIMS-CSRF"
)

// RequireCSRFHeader rejects cookie-authenticated auth requests that don't
// carry CSRFHeaderName. A request with no Refresh Cookie at all is assumed
// to be a non-browser client using the JSON-body fallback, which isn't
// exposed to browser CSRF and so isn't required to send the header.
func RequireCSRFHeader() fiber.Handler {
	return func(c fiber.Ctx) error {
		if c.Cookies(RefreshCookieName) == "" {
			return c.Next()
		}
		if c.Get(CSRFHeaderName) == "" {
			return fiber.NewError(fiber.StatusForbidden, "missing "+CSRFHeaderName+" header")
		}
		return c.Next()
	}
}
