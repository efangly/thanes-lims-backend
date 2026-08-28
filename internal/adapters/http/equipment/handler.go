package equipment

import (
	"strconv"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/response"
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/validate"
	applicationequipment "github.com/efangly/thanes-lims-backend/internal/application/equipment"
	domainequipment "github.com/efangly/thanes-lims-backend/internal/domain/equipment"
	portequipment "github.com/efangly/thanes-lims-backend/internal/ports/equipment"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	create                *applicationequipment.CreateEquipmentUseCase
	update                *applicationequipment.UpdateEquipmentUseCase
	list                  *applicationequipment.ListEquipmentUseCase
	get                   *applicationequipment.GetEquipmentUseCase
	recordCalibration     *applicationequipment.RecordCalibrationUseCase
	listCalibrationEvents *applicationequipment.ListCalibrationEventsUseCase
	schedules             *applicationequipment.CalibrationScheduleUseCase
	searchResults         *applicationequipment.SearchCalibrationResultsUseCase
}

func NewHandler(
	create *applicationequipment.CreateEquipmentUseCase,
	update *applicationequipment.UpdateEquipmentUseCase,
	list *applicationequipment.ListEquipmentUseCase,
	get *applicationequipment.GetEquipmentUseCase,
	recordCalibration *applicationequipment.RecordCalibrationUseCase,
	listCalibrationEvents *applicationequipment.ListCalibrationEventsUseCase,
	schedules *applicationequipment.CalibrationScheduleUseCase,
	searchResults *applicationequipment.SearchCalibrationResultsUseCase,
) *Handler {
	return &Handler{
		create:                create,
		update:                update,
		list:                  list,
		get:                   get,
		recordCalibration:     recordCalibration,
		listCalibrationEvents: listCalibrationEvents,
		schedules:             schedules,
		searchResults:         searchResults,
	}
}

// Create godoc
//
//	@Summary		เพิ่มอุปกรณ์ใหม่
//	@Tags			equipment
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		CreateEquipmentRequest	true	"ข้อมูลอุปกรณ์"
//	@Success		201		{object}	response.Envelope{data=EquipmentResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Router			/equipment [post]
func (h *Handler) Create(c fiber.Ctx) error {
	var req CreateEquipmentRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if err := validate.Struct(req); err != nil {
		return err
	}

	e, err := h.create.Execute(c.Context(), applicationequipment.CreateEquipmentInput{
		Name:               req.Name,
		TypeCode:           req.TypeCode,
		NextCalibrationDue: req.NextCalibrationDue,
		SerialNumber:       req.SerialNumber,
		Category:           req.Category,
		Manufacturer:       req.Manufacturer,
		Model:              req.Model,
		InstallationDate:   req.InstallationDate,
		VendorID:           req.VendorID,
		LocationID:         req.LocationID,
	})
	if err != nil {
		return err
	}
	c.Locals(middleware.LocalsAuditChangeSet, middleware.Snapshot(e))
	return response.Created(c, toResponse(e))
}

// Update godoc
//
//	@Summary		แก้ไขข้อมูลอุปกรณ์
//	@Description	แก้ไขแบบบางส่วน — ส่งเฉพาะ field ที่ต้องการเปลี่ยน (วันสอบเทียบแก้ผ่าน /calibration เท่านั้น)
//	@Tags			equipment
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Equipment ID"
//	@Param			request	body		UpdateEquipmentRequest	true	"ข้อมูลที่แก้ไข"
//	@Success		200		{object}	response.Envelope{data=EquipmentResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope
//	@Router			/equipment/{id} [patch]
func (h *Handler) Update(c fiber.Ctx) error {
	var req UpdateEquipmentRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}

	id := c.Params("id")
	before, err := h.get.Execute(c.Context(), id)
	if err != nil {
		return err
	}

	e, err := h.update.Execute(c.Context(), applicationequipment.UpdateEquipmentInput{
		ID:               id,
		Name:             req.Name,
		TypeCode:         req.TypeCode,
		SerialNumber:     req.SerialNumber,
		Category:         req.Category,
		Manufacturer:     req.Manufacturer,
		Model:            req.Model,
		InstallationDate: req.InstallationDate,
		ClearInstallDate: req.ClearInstallationDate,
		VendorID:         req.VendorID,
		LocationID:       req.LocationID,
	})
	if err != nil {
		return err
	}
	c.Locals(middleware.LocalsAuditChangeSet, middleware.ChangeSet(before, e))
	return response.OK(c, toResponse(e))
}

// List godoc
//
//	@Summary		รายการอุปกรณ์ทั้งหมด
//	@Tags			equipment
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Envelope{data=[]EquipmentResponse}
//	@Failure		401	{object}	response.Envelope
//	@Router			/equipment [get]
func (h *Handler) List(c fiber.Ctx) error {
	items, err := h.list.Execute(c.Context())
	if err != nil {
		return err
	}
	out := make([]EquipmentResponse, len(items))
	for i, e := range items {
		out[i] = toResponse(e)
	}
	return response.OK(c, out)
}

// Get godoc
//
//	@Summary		ดึงข้อมูลอุปกรณ์ตาม ID
//	@Tags			equipment
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Equipment ID"
//	@Success		200	{object}	response.Envelope{data=EquipmentResponse}
//	@Failure		401	{object}	response.Envelope
//	@Failure		404	{object}	response.Envelope
//	@Router			/equipment/{id} [get]
func (h *Handler) Get(c fiber.Ctx) error {
	e, err := h.get.Execute(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, toResponse(e))
}

// RecordCalibration godoc
//
//	@Summary		บันทึกการสอบเทียบอุปกรณ์
//	@Tags			equipment
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string						true	"Equipment ID"
//	@Param			request	body		RecordCalibrationRequest	true	"วันสอบเทียบครั้งถัดไป"
//	@Success		200		{object}	response.Envelope{data=EquipmentResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope
//	@Router			/equipment/{id}/calibration [patch]
func (h *Handler) RecordCalibration(c fiber.Ctx) error {
	var req RecordCalibrationRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if err := validate.Struct(req); err != nil {
		return err
	}

	name := fiber.Locals[string](c, middleware.LocalsName)

	e, err := h.recordCalibration.Execute(c.Context(), applicationequipment.RecordCalibrationInput{
		ID:                 c.Params("id"),
		NextCalibrationDue: req.NextCalibrationDue,
		PerformedBy:        name,
		Notes:              req.Notes,
		CalibrationType:    req.CalibrationType,
		CalibrateValue:     req.CalibrateValue,
		AcceptanceValue:    req.AcceptanceValue,
		Result:             domainequipment.CalibrationResult(req.Result),
	})
	if err != nil {
		return err
	}
	// RecordCalibrationUseCase.Execute returns the updated Equipment, not
	// the CalibrationEvent it appends internally - but the event's
	// CalibratedAt/NextCalibrationDue end up mirrored onto Equipment's own
	// fields, so it can be reconstructed here for the append-only snapshot
	// (Calibration Event is audited as its own resource - see ADR 0003).
	c.Locals(middleware.LocalsAuditResource, "calibration_event")
	c.Locals(middleware.LocalsAuditChangeSet, middleware.Snapshot(domainequipment.CalibrationEvent{
		EquipmentID:        e.ID,
		CalibratedAt:       e.LastCalibratedAt,
		NextCalibrationDue: e.NextCalibrationDue,
		PerformedBy:        name,
		Notes:              req.Notes,
	}))
	return response.OK(c, toResponse(e))
}

// ListCalibrationEvents godoc
//
//	@Summary		ประวัติการสอบเทียบของอุปกรณ์
//	@Tags			equipment
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Equipment ID"
//	@Success		200	{object}	response.Envelope{data=[]CalibrationEventResponse}
//	@Failure		401	{object}	response.Envelope
//	@Router			/equipment/{id}/calibration-events [get]
func (h *Handler) ListCalibrationEvents(c fiber.Ctx) error {
	events, err := h.listCalibrationEvents.Execute(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	out := make([]CalibrationEventResponse, len(events))
	for i, ev := range events {
		out[i] = toCalibrationEventResponse(ev)
	}
	return response.OK(c, out)
}

// ListCalibrationSchedules godoc
//
//	@Summary		รายการตารางสอบเทียบของอุปกรณ์
//	@Tags			equipment
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Equipment ID"
//	@Success		200	{object}	response.Envelope{data=[]CalibrationScheduleResponse}
//	@Failure		401	{object}	response.Envelope
//	@Router			/equipment/{id}/calibration-schedules [get]
func (h *Handler) ListCalibrationSchedules(c fiber.Ctx) error {
	scheds, err := h.schedules.ListByEquipment(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	out := make([]CalibrationScheduleResponse, len(scheds))
	for i, s := range scheds {
		out[i] = toCalibrationScheduleResponse(s)
	}
	return response.OK(c, out)
}

// CreateCalibrationSchedule godoc
//
//	@Summary		เพิ่มตารางสอบเทียบให้อุปกรณ์
//	@Tags			equipment
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string								true	"Equipment ID"
//	@Param			request	body		CreateCalibrationScheduleRequest	true	"ข้อมูลตารางสอบเทียบ"
//	@Success		201		{object}	response.Envelope{data=CalibrationScheduleResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope
//	@Router			/equipment/{id}/calibration-schedules [post]
func (h *Handler) CreateCalibrationSchedule(c fiber.Ctx) error {
	var req CreateCalibrationScheduleRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if err := validate.Struct(req); err != nil {
		return err
	}
	s, err := h.schedules.Create(c.Context(), applicationequipment.CreateCalibrationScheduleInput{
		EquipmentID:    c.Params("id"),
		Label:          req.Label,
		NextDueDate:    req.NextDueDate,
		IntervalMonths: req.IntervalMonths,
	})
	if err != nil {
		return err
	}
	c.Locals(middleware.LocalsAuditChangeSet, middleware.Snapshot(s))
	return response.Created(c, toCalibrationScheduleResponse(s))
}

// UpdateCalibrationSchedule godoc
//
//	@Summary		แก้ไขตารางสอบเทียบ
//	@Tags			equipment
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id			path		string								true	"Equipment ID"
//	@Param			scheduleId	path		integer								true	"Calibration Schedule ID"
//	@Param			request		body		UpdateCalibrationScheduleRequest	true	"ข้อมูลที่แก้ไข"
//	@Success		200			{object}	response.Envelope{data=CalibrationScheduleResponse}
//	@Failure		400			{object}	response.Envelope
//	@Failure		404			{object}	response.Envelope
//	@Router			/equipment/{id}/calibration-schedules/{scheduleId} [patch]
func (h *Handler) UpdateCalibrationSchedule(c fiber.Ctx) error {
	scheduleID, err := strconv.ParseInt(c.Params("scheduleId"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid schedule id")
	}
	var req UpdateCalibrationScheduleRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if err := validate.Struct(req); err != nil {
		return err
	}
	s, err := h.schedules.Update(c.Context(), applicationequipment.UpdateCalibrationScheduleInput{
		EquipmentID:    c.Params("id"),
		ID:             scheduleID,
		Label:          req.Label,
		NextDueDate:    req.NextDueDate,
		IntervalMonths: req.IntervalMonths,
		ClearInterval:  req.ClearInterval,
	})
	if err != nil {
		return err
	}
	return response.OK(c, toCalibrationScheduleResponse(s))
}

// DeleteCalibrationSchedule godoc
//
//	@Summary		ลบตารางสอบเทียบ
//	@Tags			equipment
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id			path	string	true	"Equipment ID"
//	@Param			scheduleId	path	integer	true	"Calibration Schedule ID"
//	@Success		204
//	@Failure		404	{object}	response.Envelope
//	@Router			/equipment/{id}/calibration-schedules/{scheduleId} [delete]
func (h *Handler) DeleteCalibrationSchedule(c fiber.Ctx) error {
	scheduleID, err := strconv.ParseInt(c.Params("scheduleId"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid schedule id")
	}
	if err := h.schedules.Delete(c.Context(), c.Params("id"), scheduleID); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// SearchCalibrationResults godoc
//
//	@Summary		ค้นหาผลการสอบเทียบทั้งหมด
//	@Description	รายการผลสอบเทียบข้ามทุกอุปกรณ์ พร้อม search bar (requirement 2.2.1)
//	@Tags			equipment
//	@Produce		json
//	@Security		BearerAuth
//	@Param			q				query		string	false	"ค้นหา (รหัส/ชื่ออุปกรณ์, ผู้สอบเทียบ, ประเภท, หมายเหตุ)"
//	@Param			equipment_id	query		string	false	"กรองตามอุปกรณ์"
//	@Param			result			query		string	false	"pass หรือ fail"
//	@Param			from			query		string	false	"วันสอบเทียบตั้งแต่ (RFC3339)"
//	@Param			to				query		string	false	"วันสอบเทียบถึง (RFC3339)"
//	@Success		200				{object}	response.Envelope{data=[]CalibrationEventResponse}
//	@Failure		401				{object}	response.Envelope
//	@Router			/calibration-results [get]
func (h *Handler) SearchCalibrationResults(c fiber.Ctx) error {
	f := portequipment.CalibrationSearchFilter{
		Query:       c.Query("q"),
		EquipmentID: c.Query("equipment_id"),
		Result:      domainequipment.CalibrationResult(c.Query("result")),
	}
	if v := c.Query("from"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid from date")
		}
		f.From = &t
	}
	if v := c.Query("to"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid to date")
		}
		f.To = &t
	}
	events, err := h.searchResults.Execute(c.Context(), f)
	if err != nil {
		return err
	}
	out := make([]CalibrationEventResponse, len(events))
	for i, ev := range events {
		out[i] = toCalibrationEventResponse(ev)
	}
	return response.OK(c, out)
}
