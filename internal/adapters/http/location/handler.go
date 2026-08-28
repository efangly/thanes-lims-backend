package location

import (
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/response"
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/validate"
	applicationlocation "github.com/efangly/thanes-lims-backend/internal/application/location"
	domainlocation "github.com/efangly/thanes-lims-backend/internal/domain/location"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	createCabinet    *applicationlocation.CreateCabinetUseCase
	generateChildren *applicationlocation.GenerateChildrenUseCase
	listChildren     *applicationlocation.ListChildrenUseCase
	getFullPath      *applicationlocation.GetFullPathUseCase
	deleteLocation   *applicationlocation.DeleteLocationUseCase
	lookupByBarcode  *applicationlocation.LookupByBarcodeUseCase
}

func NewHandler(
	createCabinet *applicationlocation.CreateCabinetUseCase,
	generateChildren *applicationlocation.GenerateChildrenUseCase,
	listChildren *applicationlocation.ListChildrenUseCase,
	getFullPath *applicationlocation.GetFullPathUseCase,
	deleteLocation *applicationlocation.DeleteLocationUseCase,
	lookupByBarcode *applicationlocation.LookupByBarcodeUseCase,
) *Handler {
	return &Handler{
		createCabinet:    createCabinet,
		generateChildren: generateChildren,
		listChildren:     listChildren,
		getFullPath:      getFullPath,
		deleteLocation:   deleteLocation,
		lookupByBarcode:  lookupByBarcode,
	}
}

// CreateCabinet godoc
//
//	@Summary		สร้างตู้ (Cabinet) ใหม่
//	@Description	สร้าง root Location - ชื่อตู้ต้อง unique ทั้งระบบ
//	@Tags			locations
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		CreateCabinetRequest	true	"ชื่อตู้"
//	@Success		201		{object}	response.Envelope{data=LocationResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Failure		409		{object}	response.Envelope
//	@Router			/locations [post]
func (h *Handler) CreateCabinet(c fiber.Ctx) error {
	var req CreateCabinetRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if err := validate.Struct(req); err != nil {
		return err
	}

	l, err := h.createCabinet.Execute(c.Context(), applicationlocation.CreateCabinetInput{
		Name: req.Name,
		Kind: domainlocation.Kind(req.Kind),
	})
	if err != nil {
		return err
	}
	c.Locals(middleware.LocalsAuditChangeSet, middleware.Snapshot(l))
	return response.Created(c, toLocationResponse(l))
}

// ListChildren godoc
//
//	@Summary		รายการ Location ลูกโดยตรง
//	@Description	ไม่ระบุ query param `parent_id` จะได้รายการตู้ (root) ทั้งหมด
//	@Tags			locations
//	@Produce		json
//	@Security		BearerAuth
//	@Param			parent_id	query		string	false	"Parent Location ID"
//	@Param			kind		query		string	false	"Tree kind for root listing (sample_storage|equipment_storage, default sample_storage)"
//	@Success		200			{object}	response.Envelope{data=[]LocationResponse}
//	@Failure		401			{object}	response.Envelope
//	@Router			/locations [get]
func (h *Handler) ListChildren(c fiber.Ctx) error {
	var parentID *string
	if pid := c.Query("parent_id"); pid != "" {
		parentID = &pid
	}

	children, err := h.listChildren.Execute(c.Context(), parentID, domainlocation.Kind(c.Query("kind")))
	if err != nil {
		return err
	}
	out := make([]LocationResponse, len(children))
	for i, l := range children {
		out[i] = toLocationResponse(l)
	}
	return response.OK(c, out)
}

// GenerateChildren godoc
//
//	@Summary		สร้างลูกอัตโนมัติตามจำนวนที่ระบุ
//	@Description	สร้าง Location ลูกของ :id จำนวน `count` ตัว ชื่อ "{prefix}-1".."{prefix}-{count}" ที่ระดับถัดไปในลำดับชั้น (cabinet > shelf > slot > sub_slot)
//	@Tags			locations
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Parent Location ID"
//	@Param			request	body		GenerateChildrenRequest	true	"prefix + จำนวน"
//	@Success		201		{object}	response.Envelope{data=[]LocationResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope
//	@Failure		409		{object}	response.Envelope
//	@Router			/locations/{id}/children [post]
func (h *Handler) GenerateChildren(c fiber.Ctx) error {
	id := c.Params("id")

	var req GenerateChildrenRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if err := validate.Struct(req); err != nil {
		return err
	}

	children, err := h.generateChildren.Execute(c.Context(), applicationlocation.GenerateChildrenInput{
		ParentID: id,
		Prefix:   req.Prefix,
		Count:    req.Count,
	})
	if err != nil {
		return err
	}
	out := make([]LocationResponse, len(children))
	for i, l := range children {
		out[i] = toLocationResponse(l)
	}
	return response.Created(c, out)
}

// GetFullPath godoc
//
//	@Summary		รหัสเต็มของ Location (Full Path)
//	@Description	คำนวณจาก tree ทุกครั้งที่เรียก เช่น "Fridge-A / Shelf-2 / Slot-4"
//	@Tags			locations
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Location ID"
//	@Success		200	{object}	response.Envelope{data=FullPathResponse}
//	@Failure		401	{object}	response.Envelope
//	@Failure		404	{object}	response.Envelope
//	@Router			/locations/{id}/full-path [get]
func (h *Handler) GetFullPath(c fiber.Ctx) error {
	id := c.Params("id")
	fullPath, err := h.getFullPath.Execute(c.Context(), id)
	if err != nil {
		return err
	}
	return response.OK(c, FullPathResponse{FullPath: fullPath})
}

// LookupByBarcode godoc
//
//	@Summary		ค้นหา Location จาก Barcode
//	@Description	สแกน Location Barcode แล้ว resolve เป็น Location โดยตรง (ใช้ตอนย้าย Sample/วางของ)
//	@Tags			locations
//	@Produce		json
//	@Security		BearerAuth
//	@Param			code	path		string	true	"Location Barcode (เช่น LOC-BC-00001)"
//	@Success		200		{object}	response.Envelope{data=LocationResponse}
//	@Failure		401		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope
//	@Router			/locations/by-barcode/{code} [get]
func (h *Handler) LookupByBarcode(c fiber.Ctx) error {
	l, err := h.lookupByBarcode.Execute(c.Context(), c.Params("code"))
	if err != nil {
		return err
	}
	return response.OK(c, toLocationResponse(l))
}

// Delete godoc
//
//	@Summary		ลบ Location
//	@Description	ลบไม่ได้ถ้ายังมีลูกหรือมี sample อ้างอิงอยู่ (restrict)
//	@Tags			locations
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Location ID"
//	@Success		204	"no content"
//	@Failure		401	{object}	response.Envelope
//	@Failure		404	{object}	response.Envelope
//	@Failure		409	{object}	response.Envelope
//	@Router			/locations/{id} [delete]
func (h *Handler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.deleteLocation.Execute(c.Context(), id); err != nil {
		return err
	}
	c.Locals(middleware.LocalsAuditChangeSet, middleware.DeletedMarker())
	return c.SendStatus(fiber.StatusNoContent)
}
