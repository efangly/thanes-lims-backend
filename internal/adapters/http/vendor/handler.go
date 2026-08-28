package vendor

import (
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/response"
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/validate"
	applicationvendor "github.com/efangly/thanes-lims-backend/internal/application/vendor"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	create *applicationvendor.CreateVendorUseCase
	list   *applicationvendor.ListVendorsUseCase
	get    *applicationvendor.GetVendorUseCase
	update *applicationvendor.UpdateVendorUseCase
}

func NewHandler(
	create *applicationvendor.CreateVendorUseCase,
	list *applicationvendor.ListVendorsUseCase,
	get *applicationvendor.GetVendorUseCase,
	update *applicationvendor.UpdateVendorUseCase,
) *Handler {
	return &Handler{create: create, list: list, get: get, update: update}
}

// Create godoc
//
//	@Summary		เพิ่มผู้ขาย (Vendor) ใหม่
//	@Description	master data ผู้ขาย - ชื่อต้อง unique
//	@Tags			vendors
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		CreateVendorRequest	true	"ข้อมูลผู้ขาย"
//	@Success		201		{object}	response.Envelope{data=VendorResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Failure		409		{object}	response.Envelope
//	@Router			/vendors [post]
func (h *Handler) Create(c fiber.Ctx) error {
	var req CreateVendorRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if err := validate.Struct(req); err != nil {
		return err
	}

	v, err := h.create.Execute(c.Context(), applicationvendor.CreateVendorInput{
		Name:         req.Name,
		ContactName:  req.ContactName,
		ContactPhone: req.ContactPhone,
		ContactEmail: req.ContactEmail,
		Address:      req.Address,
	})
	if err != nil {
		return err
	}
	c.Locals(middleware.LocalsAuditChangeSet, middleware.Snapshot(toResponse(v)))
	return response.Created(c, toResponse(v))
}

// List godoc
//
//	@Summary		รายการผู้ขายทั้งหมด
//	@Tags			vendors
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Envelope{data=[]VendorResponse}
//	@Failure		401	{object}	response.Envelope
//	@Router			/vendors [get]
func (h *Handler) List(c fiber.Ctx) error {
	items, err := h.list.Execute(c.Context())
	if err != nil {
		return err
	}
	out := make([]VendorResponse, len(items))
	for i, v := range items {
		out[i] = toResponse(v)
	}
	return response.OK(c, out)
}

// Get godoc
//
//	@Summary		ดึงข้อมูลผู้ขายตาม ID
//	@Tags			vendors
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Vendor ID"
//	@Success		200	{object}	response.Envelope{data=VendorResponse}
//	@Failure		401	{object}	response.Envelope
//	@Failure		404	{object}	response.Envelope
//	@Router			/vendors/{id} [get]
func (h *Handler) Get(c fiber.Ctx) error {
	v, err := h.get.Execute(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, toResponse(v))
}

// Update godoc
//
//	@Summary		แก้ไขข้อมูลผู้ขาย
//	@Tags			vendors
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string				true	"Vendor ID"
//	@Param			request	body		UpdateVendorRequest	true	"ข้อมูลที่ต้องการแก้ไข"
//	@Success		200		{object}	response.Envelope{data=VendorResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope
//	@Failure		409		{object}	response.Envelope
//	@Router			/vendors/{id} [patch]
func (h *Handler) Update(c fiber.Ctx) error {
	id := c.Params("id")

	var req UpdateVendorRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if err := validate.Struct(req); err != nil {
		return err
	}

	before, err := h.get.Execute(c.Context(), id)
	if err != nil {
		return err
	}

	v, err := h.update.Execute(c.Context(), applicationvendor.UpdateVendorInput{
		ID:           id,
		Name:         req.Name,
		ContactName:  req.ContactName,
		ContactPhone: req.ContactPhone,
		ContactEmail: req.ContactEmail,
		Address:      req.Address,
	})
	if err != nil {
		return err
	}
	c.Locals(middleware.LocalsAuditChangeSet, middleware.ChangeSet(toResponse(before), toResponse(v)))
	return response.OK(c, toResponse(v))
}
