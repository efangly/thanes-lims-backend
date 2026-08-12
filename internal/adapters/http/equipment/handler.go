package equipment

import (
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/response"
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/validate"
	applicationequipment "github.com/efangly/thanes-lims-backend/internal/application/equipment"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	create            *applicationequipment.CreateEquipmentUseCase
	list              *applicationequipment.ListEquipmentUseCase
	get               *applicationequipment.GetEquipmentUseCase
	recordCalibration *applicationequipment.RecordCalibrationUseCase
}

func NewHandler(
	create *applicationequipment.CreateEquipmentUseCase,
	list *applicationequipment.ListEquipmentUseCase,
	get *applicationequipment.GetEquipmentUseCase,
	recordCalibration *applicationequipment.RecordCalibrationUseCase,
) *Handler {
	return &Handler{create: create, list: list, get: get, recordCalibration: recordCalibration}
}

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

func (h *Handler) Get(c fiber.Ctx) error {
	e, err := h.get.Execute(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, toResponse(e))
}

func (h *Handler) RecordCalibration(c fiber.Ctx) error {
	var req RecordCalibrationRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if err := validate.Struct(req); err != nil {
		return err
	}

	e, err := h.recordCalibration.Execute(c.Context(), applicationequipment.RecordCalibrationInput{
		ID:                 c.Params("id"),
		NextCalibrationDue: req.NextCalibrationDue,
	})
	if err != nil {
		return err
	}
	return response.OK(c, toResponse(e))
}
