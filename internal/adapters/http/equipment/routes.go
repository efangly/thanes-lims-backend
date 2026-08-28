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
	eq.Patch("/:id", middleware.RequirePermission(rbac.ModuleEquipment, rbac.ActionEdit), h.Update)
	eq.Patch("/:id/calibration", middleware.RequirePermission(rbac.ModuleEquipment, rbac.ActionApprove), h.RecordCalibration)
	eq.Get("/:id/calibration-events", middleware.RequirePermission(rbac.ModuleEquipment, rbac.ActionView), h.ListCalibrationEvents)

	eq.Get("/:id/calibration-schedules", middleware.RequirePermission(rbac.ModuleEquipment, rbac.ActionView), h.ListCalibrationSchedules)
	eq.Post("/:id/calibration-schedules", middleware.RequirePermission(rbac.ModuleEquipment, rbac.ActionEdit), h.CreateCalibrationSchedule)
	eq.Patch("/:id/calibration-schedules/:scheduleId", middleware.RequirePermission(rbac.ModuleEquipment, rbac.ActionEdit), h.UpdateCalibrationSchedule)
	eq.Delete("/:id/calibration-schedules/:scheduleId", middleware.RequirePermission(rbac.ModuleEquipment, rbac.ActionEdit), h.DeleteCalibrationSchedule)

	r.Get("/calibration-results", authGuard, middleware.RequirePermission(rbac.ModuleEquipment, rbac.ActionView), h.SearchCalibrationResults)
}
