package purchaseorder

import (
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/middleware"
	"github.com/efangly/thanes-lims-backend/internal/adapters/http/response"
	applicationpurchaseorder "github.com/efangly/thanes-lims-backend/internal/application/purchaseorder"
	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	list         *applicationpurchaseorder.ListPOsUseCase
	get          *applicationpurchaseorder.GetPOUseCase
	approve      *applicationpurchaseorder.ApprovePOUseCase
	markReceived *applicationpurchaseorder.MarkReceivedUseCase
}

func NewHandler(
	list *applicationpurchaseorder.ListPOsUseCase,
	get *applicationpurchaseorder.GetPOUseCase,
	approve *applicationpurchaseorder.ApprovePOUseCase,
	markReceived *applicationpurchaseorder.MarkReceivedUseCase,
) *Handler {
	return &Handler{list: list, get: get, approve: approve, markReceived: markReceived}
}

// List godoc
//
//	@Summary		รายการใบสั่งซื้อทั้งหมด
//	@Tags			purchase-orders
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Envelope{data=[]POResponse}
//	@Failure		401	{object}	response.Envelope
//	@Router			/purchase-orders [get]
func (h *Handler) List(c fiber.Ctx) error {
	pos, err := h.list.Execute(c.Context())
	if err != nil {
		return err
	}
	out := make([]POResponse, len(pos))
	for i, po := range pos {
		out[i] = toResponse(po)
	}
	return response.OK(c, out)
}

// Get godoc
//
//	@Summary		ดึงข้อมูลใบสั่งซื้อตาม ID
//	@Tags			purchase-orders
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Purchase Order ID"
//	@Success		200	{object}	response.Envelope{data=POResponse}
//	@Failure		401	{object}	response.Envelope
//	@Failure		404	{object}	response.Envelope
//	@Router			/purchase-orders/{id} [get]
func (h *Handler) Get(c fiber.Ctx) error {
	po, err := h.get.Execute(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, toResponse(po))
}

// Approve godoc
//
//	@Summary		อนุมัติใบสั่งซื้อ
//	@Tags			purchase-orders
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Purchase Order ID"
//	@Success		200	{object}	response.Envelope{data=POResponse}
//	@Failure		401	{object}	response.Envelope
//	@Failure		403	{object}	response.Envelope
//	@Failure		404	{object}	response.Envelope
//	@Failure		409	{object}	response.Envelope
//	@Router			/purchase-orders/{id}/approve [patch]
func (h *Handler) Approve(c fiber.Ctx) error {
	role := fiber.Locals[domainuser.Role](c, middleware.LocalsRole)
	id := c.Params("id")
	before, err := h.get.Execute(c.Context(), id)
	if err != nil {
		return err
	}

	po, err := h.approve.Execute(c.Context(), applicationpurchaseorder.ApprovePOInput{
		ID: id, ActorRole: role,
	})
	if err != nil {
		return err
	}
	c.Locals(middleware.LocalsAuditChangeSet, middleware.ChangeSet(before, po))
	return response.OK(c, toResponse(po))
}

// MarkReceived godoc
//
//	@Summary		บันทึกรับสินค้าตามใบสั่งซื้อ
//	@Description	เพิ่มจำนวนคงคลังตามใบสั่งซื้อโดยอัตโนมัติ
//	@Tags			purchase-orders
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Purchase Order ID"
//	@Success		200	{object}	response.Envelope{data=POResponse}
//	@Failure		401	{object}	response.Envelope
//	@Failure		404	{object}	response.Envelope
//	@Failure		409	{object}	response.Envelope
//	@Router			/purchase-orders/{id}/receive [patch]
func (h *Handler) MarkReceived(c fiber.Ctx) error {
	id := c.Params("id")
	before, err := h.get.Execute(c.Context(), id)
	if err != nil {
		return err
	}

	po, err := h.markReceived.Execute(c.Context(), id)
	if err != nil {
		return err
	}
	c.Locals(middleware.LocalsAuditChangeSet, middleware.ChangeSet(before, po))
	return response.OK(c, toResponse(po))
}
