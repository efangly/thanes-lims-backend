package equipment

import (
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	"github.com/efangly/thanes-lims-backend/internal/domain/rbac"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, h *Handler, tokens portuser.TokenService) {
	authGuard := middleware.Auth(tokens)

	eq := r.Group("/equipment", authGuard)
	eq.Post("/", middleware.RequirePermission(rbac.ModuleEquipment, rbac.ActionCreate), h.Create)
	eq.Get("/", middleware.RequirePermission(rbac.ModuleEquipment, rbac.ActionView), h.List)
	eq.Get("/:id", middleware.RequirePermission(rbac.ModuleEquipment, rbac.ActionView), h.Get)
	eq.Patch("/:id/calibration", middleware.RequirePermission(rbac.ModuleEquipment, rbac.ActionApprove), h.RecordCalibration)
	eq.Get("/:id/calibration-events", middleware.RequirePermission(rbac.ModuleEquipment, rbac.ActionView), h.ListCalibrationEvents)
}
