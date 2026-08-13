package environment

import (
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/response"
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/validate"
	applicationenvironment "github.com/efangly/thanes-lims-backend/internal/application/environment"
	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/valyala/fasthttp"
)

type Handler struct {
	record     *applicationenvironment.RecordReadingUseCase
	listGauges *applicationenvironment.ListGaugesUseCase
	getTrend   *applicationenvironment.GetTrendUseCase
	listAlerts *applicationenvironment.ListAlertsUseCase
	hub        *Hub
}

func NewHandler(
	record *applicationenvironment.RecordReadingUseCase,
	listGauges *applicationenvironment.ListGaugesUseCase,
	getTrend *applicationenvironment.GetTrendUseCase,
	listAlerts *applicationenvironment.ListAlertsUseCase,
	hub *Hub,
) *Handler {
	return &Handler{record: record, listGauges: listGauges, getTrend: getTrend, listAlerts: listAlerts, hub: hub}
}

// ListGauges godoc
//
//	@Summary		สถานะเกจวัดสภาพแวดล้อมทั้งหมด
//	@Tags			environment
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Envelope{data=[]GaugeResponse}
//	@Failure		401	{object}	response.Envelope
//	@Router			/environment/gauges [get]
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

// GetTrend godoc
//
//	@Summary		ประวัติค่าที่วัดได้ของตำแหน่งหนึ่ง
//	@Tags			environment
//	@Produce		json
//	@Security		BearerAuth
//	@Param			loc		path		string	true	"ตำแหน่งเซนเซอร์"
//	@Param			limit	query		int		false	"จำนวนรายการล่าสุด"	default(50)
//	@Success		200		{object}	response.Envelope{data=[]ReadingResponse}
//	@Failure		401		{object}	response.Envelope
//	@Router			/environment/gauges/{loc}/trend [get]
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

// ListAlerts godoc
//
//	@Summary		รายการ alert สภาพแวดล้อมทั้งหมด
//	@Tags			environment
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Envelope{data=[]AlertResponse}
//	@Failure		401	{object}	response.Envelope
//	@Router			/environment/alerts [get]
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

// RecordReading godoc
//
//	@Summary		บันทึกค่าที่วัดได้จากเซนเซอร์
//	@Description	ประเมิน threshold อัตโนมัติและสร้าง/ยกระดับ alert พร้อมแจ้งเตือนหากเกินขอบเขต
//	@Tags			environment
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		RecordReadingRequest	true	"ค่าที่วัดได้"
//	@Success		201		{object}	response.Envelope
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Router			/environment/readings [post]
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

var alertsUpgrader = websocket.FastHTTPUpgrader{
	CheckOrigin: func(ctx *fasthttp.RequestCtx) bool { return true },
}

// AlertsWS upgrades the connection to a WebSocket and streams every
// newly created/escalated environment alert as JSON until the client
// disconnects. Auth already ran as middleware.AuthQuery before this
// handler, via the `token` query parameter (browsers can't set a custom
// Authorization header on a WS handshake).
func (h *Handler) AlertsWS(c fiber.Ctx) error {
	return alertsUpgrader.Upgrade(c.RequestCtx(), func(conn *websocket.Conn) {
		h.hub.Register(conn)
		defer func() {
			h.hub.Unregister(conn)
			conn.Close()
		}()

		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	})
}
