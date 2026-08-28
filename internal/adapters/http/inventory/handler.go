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
	update              *applicationinventory.UpdateItemUseCase
	receiveStock        *applicationinventory.ReceiveStockUseCase
	issueStock          *applicationinventory.IssueStockUseCase
	listLots            *applicationinventory.ListLotsUseCase
	updateDefaultVendor *applicationinventory.UpdateDefaultVendorUseCase
	reorder             *applicationpurchaseorder.CreateFromLowStockUseCase
}

func NewHandler(
	create *applicationinventory.CreateItemUseCase,
	list *applicationinventory.ListItemsUseCase,
	get *applicationinventory.GetItemUseCase,
	update *applicationinventory.UpdateItemUseCase,
	receiveStock *applicationinventory.ReceiveStockUseCase,
	issueStock *applicationinventory.IssueStockUseCase,
	listLots *applicationinventory.ListLotsUseCase,
	updateDefaultVendor *applicationinventory.UpdateDefaultVendorUseCase,
	reorder *applicationpurchaseorder.CreateFromLowStockUseCase,
) *Handler {
	return &Handler{create: create, list: list, get: get, update: update, receiveStock: receiveStock, issueStock: issueStock, listLots: listLots, updateDefaultVendor: updateDefaultVendor, reorder: reorder}
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
		Name: req.Name, Category: req.Category, Unit: req.Unit, Min: req.Min, Max: req.Max,
		DefaultVendor:   req.DefaultVendor,
		CustodianUserID: req.CustodianUserID,
		Manufacturer:    req.Manufacturer,
		VendorID:        req.VendorID,
		LocationID:      req.LocationID,
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

// Update godoc
//
//	@Summary		แก้ไขข้อมูลรายการวัสดุคงคลัง
//	@Description	แก้ไขบางส่วน (partial) — จำนวนคงคลังมาจากการรับล็อต (/receive), ผู้ขายเริ่มต้นแก้ผ่าน /default-vendor
//	@Tags			inventory
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string				true	"Item ID"
//	@Param			request	body		UpdateItemRequest	true	"ข้อมูลที่แก้ไข"
//	@Success		200		{object}	response.Envelope{data=ItemResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope
//	@Router			/inventory/{id} [patch]
func (h *Handler) Update(c fiber.Ctx) error {
	var req UpdateItemRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}

	id := c.Params("id")
	before, err := h.get.Execute(c.Context(), id)
	if err != nil {
		return err
	}

	item, err := h.update.Execute(c.Context(), applicationinventory.UpdateItemInput{
		ID:              id,
		Name:            req.Name,
		Category:        req.Category,
		Unit:            req.Unit,
		Min:             req.Min,
		Max:             req.Max,
		CustodianUserID: req.CustodianUserID,
		Manufacturer:    req.Manufacturer,
		VendorID:        req.VendorID,
		LocationID:      req.LocationID,
	})
	if err != nil {
		return err
	}
	c.Locals(middleware.LocalsAuditChangeSet, middleware.ChangeSet(before, item))
	return response.OK(c, toResponse(item))
}

// ReceiveStock godoc
//
//	@Summary		รับของเข้าคลัง (สร้าง/เพิ่มล็อต)
//	@Description	รับสินค้าเข้าคลังตามล็อต — ถ้าเลขล็อตซ้ำกับล็อตเดิมของรายการจะบวกจำนวนเข้าล็อตนั้น ไม่งั้นสร้างล็อตใหม่ จำนวนคงคลังของรายการคือผลรวมของทุกล็อต
//	@Tags			inventory
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Item ID"
//	@Param			request	body		ReceiveStockRequest		true	"ข้อมูลล็อตที่รับเข้า"
//	@Success		200		{object}	response.Envelope{data=ReceiveStockResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope
//	@Router			/inventory/{id}/receive [post]
func (h *Handler) ReceiveStock(c fiber.Ctx) error {
	var req ReceiveStockRequest
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

	res, err := h.receiveStock.Execute(c.Context(), applicationinventory.ReceiveStockInput{
		ItemID:     id,
		LotNo:      req.LotNo,
		ExpireDate: req.ExpireDate,
		Quantity:   req.Quantity,
	})
	if err != nil {
		return err
	}
	c.Locals(middleware.LocalsAuditChangeSet, middleware.ChangeSet(before, res.Item))
	return response.OK(c, ReceiveStockResponse{
		Item: toResponse(res.Item),
		Lot:  toLotResponse(res.Lot),
	})
}

// IssueStock godoc
//
//	@Summary		เบิกของออกจากคลัง
//	@Description	เบิกออกตามล็อตที่ผู้ใช้เลือกเอง (ไม่ใช่ FEFO อัตโนมัติ) — ถ้าจำนวนที่เบิกเกินยอดคงเหลือของล็อต ระบบจะไม่หักและตอบยอดคงเหลือจริงกลับมา (applied=false) ให้ระบุล็อตเพิ่มหรือส่งใหม่พร้อม force=true เพื่อยอมให้ยอดล็อตติดลบ (ADR 0008)
//	@Tags			inventory
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string				true	"Item ID"
//	@Param			request	body		IssueStockRequest	true	"รายการล็อตที่เบิกออก"
//	@Success		200		{object}	response.Envelope{data=IssueStockResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope
//	@Router			/inventory/{id}/issue [post]
func (h *Handler) IssueStock(c fiber.Ctx) error {
	var req IssueStockRequest
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

	lines := make([]applicationinventory.IssueLine, len(req.Lines))
	for i, l := range req.Lines {
		lines[i] = applicationinventory.IssueLine{LotID: l.LotID, Quantity: l.Quantity}
	}

	res, err := h.issueStock.Execute(c.Context(), applicationinventory.IssueStockInput{
		ItemID: id,
		Lines:  lines,
		Force:  req.Force,
	})
	if err != nil {
		return err
	}

	out := IssueStockResponse{
		Applied: res.Applied,
		Item:    toResponse(res.Item),
		Lots:    make([]LotResponse, len(res.Lots)),
	}
	for i, l := range res.Lots {
		out.Lots[i] = toLotResponse(l)
	}
	for _, s := range res.Shortfalls {
		out.Shortfalls = append(out.Shortfalls, ShortfallResponse{
			LotID: s.LotID, LotNo: s.LotNo, Requested: s.Requested, Available: s.Available,
		})
	}
	if res.Applied {
		c.Locals(middleware.LocalsAuditChangeSet, middleware.ChangeSet(before, res.Item))
	}
	return response.OK(c, out)
}

// ListLots godoc
//
//	@Summary		รายการล็อตของวัสดุคงคลัง
//	@Description	ทุกล็อตของรายการ — ใช้เลือกล็อตตอนเบิกออก (Stock Issue)
//	@Tags			inventory
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Item ID"
//	@Success		200	{object}	response.Envelope{data=[]LotResponse}
//	@Failure		401	{object}	response.Envelope
//	@Failure		404	{object}	response.Envelope
//	@Router			/inventory/{id}/lots [get]
func (h *Handler) ListLots(c fiber.Ctx) error {
	lots, err := h.listLots.Execute(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	out := make([]LotResponse, len(lots))
	for i, l := range lots {
		out[i] = toLotResponse(l)
	}
	return response.OK(c, out)
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
