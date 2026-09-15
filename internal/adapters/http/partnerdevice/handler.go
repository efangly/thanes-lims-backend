package partnerdevice

import (
	"bufio"
	"fmt"

	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/response"
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/validate"
	applicationenvironment "github.com/efangly/thanes-lims-backend/internal/application/environment"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	create      *applicationenvironment.CreatePartnerDeviceUseCase
	update      *applicationenvironment.UpdatePartnerDeviceUseCase
	list        *applicationenvironment.ListPartnerDevicesUseCase
	get         *applicationenvironment.GetPartnerDeviceUseCase
	getSnapshot *applicationenvironment.GetPartnerDeviceSnapshotUseCase
	discover    *applicationenvironment.DiscoverPartnerDevicesByWardUseCase
	timeseries  *applicationenvironment.GetPartnerDeviceTimeseriesUseCase
	hub         *SSEHub
}

// discover and timeseries are nil when PARTNER_API_ENABLED=false -
// RegisterRoutes only mounts /discover and /:serial/timeseries when they're
// non-nil, so Handler.Discover/GetTimeseries are never called with a nil
// use case.
func NewHandler(
	create *applicationenvironment.CreatePartnerDeviceUseCase,
	update *applicationenvironment.UpdatePartnerDeviceUseCase,
	list *applicationenvironment.ListPartnerDevicesUseCase,
	get *applicationenvironment.GetPartnerDeviceUseCase,
	getSnapshot *applicationenvironment.GetPartnerDeviceSnapshotUseCase,
	discover *applicationenvironment.DiscoverPartnerDevicesByWardUseCase,
	timeseries *applicationenvironment.GetPartnerDeviceTimeseriesUseCase,
	hub *SSEHub,
) *Handler {
	return &Handler{create: create, update: update, list: list, get: get, getSnapshot: getSnapshot, discover: discover, timeseries: timeseries, hub: hub}
}

// Create godoc
//
//	@Summary		เพิ่มการผูก Partner Device (Serial) กับ Location
//	@Description	Location ต้องเป็น Gauge ที่มีอยู่แล้ว - ไม่สร้างใหม่ให้อัตโนมัติ
//	@Tags			partner-devices
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		CreatePartnerDeviceRequest	true	"ข้อมูล Partner Device"
//	@Success		201		{object}	response.Envelope{data=PartnerDeviceResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope
//	@Failure		409		{object}	response.Envelope
//	@Router			/partner-devices [post]
func (h *Handler) Create(c fiber.Ctx) error {
	var req CreatePartnerDeviceRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if err := validate.Struct(req); err != nil {
		return err
	}

	d, err := h.create.Execute(c.Context(), applicationenvironment.CreatePartnerDeviceInput{
		Serial:   req.Serial,
		Location: req.Location,
		Active:   req.Active,
	})
	if err != nil {
		return err
	}
	c.Locals(middleware.LocalsAuditChangeSet, middleware.Snapshot(toResponse(d)))
	return response.Created(c, toResponse(d))
}

// List godoc
//
//	@Summary		รายการ Partner Device ทั้งหมด
//	@Tags			partner-devices
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Envelope{data=[]PartnerDeviceResponse}
//	@Failure		401	{object}	response.Envelope
//	@Router			/partner-devices [get]
func (h *Handler) List(c fiber.Ctx) error {
	items, err := h.list.Execute(c.Context())
	if err != nil {
		return err
	}
	out := make([]PartnerDeviceResponse, len(items))
	for i, d := range items {
		out[i] = toResponse(d)
	}
	return response.OK(c, out)
}

// Get godoc
//
//	@Summary		ดึงข้อมูลการผูก Partner Device ตาม Serial
//	@Tags			partner-devices
//	@Produce		json
//	@Security		BearerAuth
//	@Param			serial	path		string	true	"Serial ของ Partner Device"
//	@Success		200		{object}	response.Envelope{data=PartnerDeviceResponse}
//	@Failure		401		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope
//	@Router			/partner-devices/{serial} [get]
func (h *Handler) Get(c fiber.Ctx) error {
	d, err := h.get.Execute(c.Context(), c.Params("serial"))
	if err != nil {
		return err
	}
	return response.OK(c, toResponse(d))
}

// Update godoc
//
//	@Summary		แก้ไขการผูก Partner Device (เปลี่ยน Location หรือปิด/เปิด Active)
//	@Tags			partner-devices
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			serial	path		string						true	"Serial ของ Partner Device"
//	@Param			request	body		UpdatePartnerDeviceRequest	true	"ข้อมูลที่ต้องการแก้ไข"
//	@Success		200		{object}	response.Envelope{data=PartnerDeviceResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope
//	@Router			/partner-devices/{serial} [patch]
func (h *Handler) Update(c fiber.Ctx) error {
	serial := c.Params("serial")

	var req UpdatePartnerDeviceRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if err := validate.Struct(req); err != nil {
		return err
	}

	before, err := h.get.Execute(c.Context(), serial)
	if err != nil {
		return err
	}

	d, err := h.update.Execute(c.Context(), applicationenvironment.UpdatePartnerDeviceInput{
		Serial:   serial,
		Location: req.Location,
		Active:   req.Active,
	})
	if err != nil {
		return err
	}
	c.Locals(middleware.LocalsAuditChangeSet, middleware.ChangeSet(toResponse(before), toResponse(d)))
	return response.OK(c, toResponse(d))
}

// GetSnapshot godoc
//
//	@Summary		ข้อมูลล่าสุดของ Partner Device (metadata + reading + alert)
//	@Description	อ่านจาก cache ที่ background poller เก็บไว้เท่านั้น (ไม่เรียก Partner API ตรงๆ) - stale:true เมื่อข้อมูลเก่ากว่า cache TTL
//	@Tags			partner-devices
//	@Produce		json
//	@Security		BearerAuth
//	@Param			serial	path		string	true	"Serial ของ Partner Device"
//	@Success		200		{object}	response.Envelope{data=SnapshotResponse}
//	@Failure		401		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope
//	@Router			/partner-devices/{serial}/snapshot [get]
func (h *Handler) GetSnapshot(c fiber.Ctx) error {
	snap, err := h.getSnapshot.Execute(c.Context(), c.Params("serial"))
	if err != nil {
		return err
	}
	return response.OK(c, toSnapshotResponse(snap))
}

// GetTimeseries godoc
//
//	@Summary		ข้อมูล time-series ย้อนหลัง 1 ชั่วโมงของ Partner Device (สำหรับทำกราฟ)
//	@Description	เรียกสดจาก SMtrack ทุกครั้ง (ไม่ผ่าน cache ของ poller เหมือน .../snapshot) - ได้ข้อมูลย้อนหลังสูงสุด 1 ชั่วโมงตามข้อจำกัดของ Partner API เอง เรียงจากใหม่ไปเก่า ว่างได้ถ้าอุปกรณ์ไม่มีค่าส่งเข้ามาในชั่วโมงที่ผ่านมา
//	@Tags			partner-devices
//	@Produce		json
//	@Security		BearerAuth
//	@Param			serial	path		string	true	"Serial ของ Partner Device"
//	@Success		200		{object}	response.Envelope{data=TimeseriesResponse}
//	@Failure		401		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope
//	@Router			/partner-devices/{serial}/timeseries [get]
func (h *Handler) GetTimeseries(c fiber.Ctx) error {
	serial := c.Params("serial")
	readings, err := h.timeseries.Execute(c.Context(), serial)
	if err != nil {
		return err
	}
	return response.OK(c, toTimeseriesResponse(serial, readings))
}

// Discover godoc
//
//	@Summary		ค้นหาอุปกรณ์ SMtrack ตาม Ward (ช่วยหา Serial ก่อนสร้าง Partner Device)
//	@Description	อ่านตรงจาก SMtrack (ไม่ persist) - ใช้ช่วยแอดมินหา Serial ของอุปกรณ์ใน Ward ก่อนสร้างการผูก Serial<->Location
//	@Tags			partner-devices
//	@Produce		json
//	@Security		BearerAuth
//	@Param			ward	query		string	true	"Ward ฝั่ง SMtrack"
//	@Param			page	query		int		false	"หน้า (default 1)"
//	@Param			limit	query		int		false	"จำนวนต่อหน้า (default 20, max 100)"
//	@Success		200		{object}	response.Envelope{data=DiscoverDevicesResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope
//	@Router			/partner-devices/discover [get]
func (h *Handler) Discover(c fiber.Ctx) error {
	listing, err := h.discover.Execute(c.Context(), applicationenvironment.DiscoverPartnerDevicesByWardInput{
		Ward:  c.Query("ward"),
		Page:  fiber.Query(c, "page", 1),
		Limit: fiber.Query(c, "limit", 20),
	})
	if err != nil {
		return err
	}
	return response.OK(c, toDiscoverResponse(listing))
}

// Stream godoc
//
//	@Summary		SSE stream ของ snapshot Partner Device ทุกตัวที่ถูก poll
//	@Description	text/event-stream - ส่ง event ใหม่ทุกครั้งที่ background poller ได้ข้อมูลใหม่ (ทุก PARTNER_API_POLL_INTERVAL วินาทีต่อ device)
//	@Tags			partner-devices
//	@Produce		text/event-stream
//	@Security		BearerAuth
//	@Success		200	{string}	string	"text/event-stream"
//	@Failure		401	{object}	response.Envelope
//	@Router			/partner-devices/stream [get]
func (h *Handler) Stream(c fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")

	ch := h.hub.Subscribe()
	return c.SendStreamWriter(func(w *bufio.Writer) {
		defer h.hub.Unsubscribe(ch)
		for payload := range ch {
			if _, err := fmt.Fprintf(w, "data: %s\n\n", payload); err != nil {
				return
			}
			if err := w.Flush(); err != nil {
				return
			}
		}
	})
}
