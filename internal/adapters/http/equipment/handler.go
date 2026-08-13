package equipment

import (
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/response"
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/validate"
	applicationequipment "github.com/efangly/thanes-lims-backend/internal/application/equipment"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	create                *applicationequipment.CreateEquipmentUseCase
	list                  *applicationequipment.ListEquipmentUseCase
	get                   *applicationequipment.GetEquipmentUseCase
	recordCalibration     *applicationequipment.RecordCalibrationUseCase
	listCalibrationEvents *applicationequipment.ListCalibrationEventsUseCase
}

func NewHandler(
	create *applicationequipment.CreateEquipmentUseCase,
	list *applicationequipment.ListEquipmentUseCase,
	get *applicationequipment.GetEquipmentUseCase,
	recordCalibration *applicationequipment.RecordCalibrationUseCase,
	listCalibrationEvents *applicationequipment.ListCalibrationEventsUseCase,
) *Handler {
	return &Handler{
		create:                create,
		list:                  list,
		get:                   get,
		recordCalibration:     recordCalibration,
		listCalibrationEvents: listCalibrationEvents,
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
	})
	if err != nil {
		return err
	}
	return response.Created(c, toResponse(e))
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
	})
	if err != nil {
		return err
	}
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
