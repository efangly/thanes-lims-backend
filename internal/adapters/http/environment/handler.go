package environment

import (
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/response"
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/validate"
	applicationenvironment "github.com/efangly/thanes-lims-backend/internal/application/environment"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	record     *applicationenvironment.RecordReadingUseCase
	listGauges *applicationenvironment.ListGaugesUseCase
	getTrend   *applicationenvironment.GetTrendUseCase
	listAlerts *applicationenvironment.ListAlertsUseCase
}

func NewHandler(
	record *applicationenvironment.RecordReadingUseCase,
	listGauges *applicationenvironment.ListGaugesUseCase,
	getTrend *applicationenvironment.GetTrendUseCase,
	listAlerts *applicationenvironment.ListAlertsUseCase,
) *Handler {
	return &Handler{record: record, listGauges: listGauges, getTrend: getTrend, listAlerts: listAlerts}
}

func (h *Handler) ListGauges(c fiber.Ctx) error {
	statuses, err := h.listGauges.Execute(c.Context())
	if err != nil {
		return err
	}
	out := make([]GaugeResponse, len(statuses))
	for i, s := range statuses {
		out[i] = toGaugeResponse(s)
	}
	return response.OK(c, out)
}

func (h *Handler) GetTrend(c fiber.Ctx) error {
	location := c.Params("loc")
	limit := fiber.Query(c, "limit", 50)

	readings, err := h.getTrend.Execute(c.Context(), location, limit)
	if err != nil {
		return err
	}
	out := make([]ReadingResponse, len(readings))
	for i, r := range readings {
		out[i] = toReadingResponse(r)
	}
	return response.OK(c, out)
}

func (h *Handler) ListAlerts(c fiber.Ctx) error {
	alerts, err := h.listAlerts.Execute(c.Context())
	if err != nil {
		return err
	}
	out := make([]AlertResponse, len(alerts))
	for i, a := range alerts {
		out[i] = toAlertResponse(a)
	}
	return response.OK(c, out)
}

func (h *Handler) RecordReading(c fiber.Ctx) error {
	var req RecordReadingRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if err := validate.Struct(req); err != nil {
		return err
	}

	result, err := h.record.Execute(c.Context(), req.Location, req.Value)
	if err != nil {
		return err
	}

	body := fiber.Map{"reading": toReadingResponse(result.Reading)}
	if result.Alert != nil {
		body["alert"] = toAlertResponse(*result.Alert)
	}
	return response.Created(c, body)
}
