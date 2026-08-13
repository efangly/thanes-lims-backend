package audit

import (
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	"github.com/gofiber/fiber/v3"
)

// RegisterRoutes gates /audit/export to admin and QA - it's a compliance
// export of every write action in the system, not something every role
// should be able to pull.
func RegisterRoutes(r fiber.Router, h *Handler, tokens portuser.TokenService) {
	authGuard := middleware.Auth(tokens)
	requireRole := middleware.RequireRole(domainuser.RoleAdmin, domainuser.RoleQA)

	r.Get("/audit/export", authGuard, requireRole, h.Export)
}
