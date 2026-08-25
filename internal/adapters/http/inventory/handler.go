package inventory

import (
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/response"
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/validate"
	applicationinventory "github.com/efangly/thanes-lims-backend/internal/application/inventory"
	applicationpurchaseorder "github.com/efangly/thanes-lims-backend/internal/application/purchaseorder"
	"github.com/gofiber/fiber/v3"
)

type ReorderRequest struct {
	Vendor string `json:"vendor" validate:"required"`
}

type Handler struct {
	create              *applicationinventory.CreateItemUseCase
	list                *applicationinventory.ListItemsUseCase
	get                 *applicationinventory.GetItemUseCase
	updateQuantity      *applicationinventory.UpdateQuantityUseCase
	updateDefaultVendor *applicationinventory.UpdateDefaultVendorUseCase
	reorder             *applicationpurchaseorder.CreateFromLowStockUseCase
}

func NewHandler(
	create *applicationinventory.CreateItemUseCase,
	list *applicationinventory.ListItemsUseCase,
	get *applicationinventory.GetItemUseCase,
	updateQuantity *applicationinventory.UpdateQuantityUseCase,
	updateDefaultVendor *applicationinventory.UpdateDefaultVendorUseCase,
	reorder *applicationpurchaseorder.CreateFromLowStockUseCase,
) *Handler {
	return &Handler{create: create, list: list, get: get, updateQuantity: updateQuantity, updateDefaultVendor: updateDefaultVendor, reorder: reorder}
}

// Create godoc
//
//	@Summary		เพิ่มรายการวัสดุคงคลังใหม่
//	@Tags			inventory
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		CreateItemRequest	true	"ข้อมูลรายการ"
//	@Success		201		{object}	response.Envelope{data=ItemResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Router			/inventory [post]
func (h *Handler) Create(c fiber.Ctx) error {
	var req CreateItemRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if err := validate.Struct(req); err != nil {
		return err
	}

	item, err := h.create.Execute(c.Context(), applicationinventory.CreateItemInput{
		Name: req.Name, Category: req.Category, Quantity: req.Quantity, Unit: req.Unit, Min: req.Min, Max: req.Max,
		DefaultVendor: req.DefaultVendor,
	})
	if err != nil {
		return err
	}
	c.Locals(middleware.LocalsAuditChangeSet, middleware.Snapshot(item))
	return response.Created(c, toResponse(item))
}

// List godoc
//
//	@Summary		รายการวัสดุคงคลังทั้งหมด
//	@Tags			inventory
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Envelope{data=[]ItemResponse}
//	@Failure		401	{object}	response.Envelope
//	@Router			/inventory [get]
func (h *Handler) List(c fiber.Ctx) error {
	items, err := h.list.Execute(c.Context())
	if err != nil {
		return err
	}
	out := make([]ItemResponse, len(items))
	for i, item := range items {
		out[i] = toResponse(item)
	}
	return response.OK(c, out)
}

// Get godoc
//
//	@Summary		ดึงข้อมูลรายการวัสดุคงคลังตาม ID
//	@Tags			inventory
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Item ID"
//	@Success		200	{object}	response.Envelope{data=ItemResponse}
//	@Failure		401	{object}	response.Envelope
//	@Failure		404	{object}	response.Envelope
//	@Router			/inventory/{id} [get]
func (h *Handler) Get(c fiber.Ctx) error {
	item, err := h.get.Execute(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, toResponse(item))
}

// UpdateQuantity godoc
//
//	@Summary		ปรับจำนวนคงคลัง
//	@Description	แจ้งเตือนอัตโนมัติหากจำนวนต่ำกว่า min
//	@Tags			inventory
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Item ID"
//	@Param			request	body		UpdateQuantityRequest	true	"จำนวนใหม่"
//	@Success		200		{object}	response.Envelope{data=ItemResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope
//	@Router			/inventory/{id}/quantity [patch]
func (h *Handler) UpdateQuantity(c fiber.Ctx) error {
	var req UpdateQuantityRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if err := validate.Struct(req); err != nil {
		return err
	}

	id := c.Params("id")
	before, err := h.get.Execute(c.Context(), id)
	if err != nil {
		return err
	}

	item, err := h.updateQuantity.Execute(c.Context(), id, req.Quantity)
	if err != nil {
		return err
	}
	c.Locals(middleware.LocalsAuditChangeSet, middleware.ChangeSet(before, item))
	return response.OK(c, toResponse(item))
}

// UpdateDefaultVendor godoc
//
//	@Summary		ตั้งค่าผู้ขายเริ่มต้นของรายการ
//	@Description	ใช้เป็นผู้ขายเมื่อระบบสั่งซื้อเพิ่มอัตโนมัติ (auto-reorder)
//	@Tags			inventory
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string						true	"Item ID"
//	@Param			request	body		UpdateDefaultVendorRequest	true	"ผู้ขายเริ่มต้น"
//	@Success		200		{object}	response.Envelope{data=ItemResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope
//	@Router			/inventory/{id}/default-vendor [patch]
func (h *Handler) UpdateDefaultVendor(c fiber.Ctx) error {
	var req UpdateDefaultVendorRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if err := validate.Struct(req); err != nil {
		return err
	}

	id := c.Params("id")
	before, err := h.get.Execute(c.Context(), id)
	if err != nil {
		return err
	}

	item, err := h.updateDefaultVendor.Execute(c.Context(), id, req.Vendor)
	if err != nil {
		return err
	}
	c.Locals(middleware.LocalsAuditChangeSet, middleware.ChangeSet(before, item))
	return response.OK(c, toResponse(item))
}

// Reorder godoc
//
//	@Summary		สั่งซื้อเพิ่ม (สร้างใบสั่งซื้อ)
//	@Description	สร้าง purchase order จากรายการที่ต่ำกว่า min ด้วยตนเอง
//	@Tags			inventory
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string			true	"Item ID"
//	@Param			request	body		ReorderRequest	true	"ผู้ขาย"
//	@Success		201		{object}	response.Envelope
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope
//	@Router			/inventory/{id}/reorder [post]
func (h *Handler) Reorder(c fiber.Ctx) error {
	var req ReorderRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if err := validate.Struct(req); err != nil {
		return err
	}

	po, err := h.reorder.Execute(c.Context(), applicationpurchaseorder.CreateFromLowStockInput{
		ItemID: c.Params("id"), Vendor: req.Vendor,
	})
	if err != nil {
		return err
	}
	return response.Created(c, fiber.Map{
		"id":         po.ID,
		"item_id":    po.ItemID,
		"quantity":   po.Quantity,
		"vendor":     po.Vendor,
		"order_date": po.OrderDate,
		"status":     string(po.Status),
	})
}
