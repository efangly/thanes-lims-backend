package testresult

import (
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	"github.com/efangly/thanes-lims-backend/internal/domain/rbac"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, h *Handler, tokens portuser.TokenService) {
	authGuard := middleware.Auth(tokens)

	tests := r.Group("/tests", authGuard)
	tests.Post("/", middleware.RequirePermission(rbac.ModuleTestResult, rbac.ActionCreate), h.Create)
	tests.Get("/", middleware.RequirePermission(rbac.ModuleTestResult, rbac.ActionView), h.List)
	tests.Get("/:id", middleware.RequirePermission(rbac.ModuleTestResult, rbac.ActionView), h.Get)
	tests.Patch("/:id/result", middleware.RequirePermission(rbac.ModuleTestResult, rbac.ActionEdit), h.SubmitResult)
	tests.Patch("/:id/approve", middleware.RequirePermission(rbac.ModuleTestResult, rbac.ActionApprove), h.Approve)
	tests.Get("/:id/report", middleware.RequirePermission(rbac.ModuleTestResult, rbac.ActionView), h.GetReport)

	// Nested convenience route, registered independently of the sample
	// module's own route group.
	r.Get("/samples/:id/tests", authGuard, middleware.RequirePermission(rbac.ModuleTestResult, rbac.ActionView), h.ListBySample)
}
