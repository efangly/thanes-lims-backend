package audit

import (
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	"github.com/efangly/thanes-lims-backend/internal/domain/rbac"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	"github.com/gofiber/fiber/v3"
)

// RegisterRoutes gates /audit/export to audit:export - it's a compliance
// export of every write action in the system, not something every role
// should be able to pull (Admin and QA hold it; Lab Manager only holds
// audit:view). /audit/logs, the JSON browse endpoint, only needs audit:view.
func RegisterRoutes(r fiber.Router, h *Handler, tokens portuser.TokenService) {
	authGuard := middleware.Auth(tokens)
	requireExport := middleware.RequirePermission(rbac.ModuleAudit, rbac.ActionExport)
	requireView := middleware.RequirePermission(rbac.ModuleAudit, rbac.ActionView)

	r.Get("/audit/export", authGuard, requireExport, h.Export)
	r.Get("/audit/logs", authGuard, requireView, h.ListLogs)
}
