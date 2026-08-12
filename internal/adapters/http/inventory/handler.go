package inventory

import (
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
	create         *applicationinventory.CreateItemUseCase
	list           *applicationinventory.ListItemsUseCase
	get            *applicationinventory.GetItemUseCase
	updateQuantity *applicationinventory.UpdateQuantityUseCase
	reorder        *applicationpurchaseorder.CreateFromLowStockUseCase
}

func NewHandler(
	create *applicationinventory.CreateItemUseCase,
	list *applicationinventory.ListItemsUseCase,
	get *applicationinventory.GetItemUseCase,
	updateQuantity *applicationinventory.UpdateQuantityUseCase,
	reorder *applicationpurchaseorder.CreateFromLowStockUseCase,
) *Handler {
	return &Handler{create: create, list: list, get: get, updateQuantity: updateQuantity, reorder: reorder}
}

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
	})
	if err != nil {
		return err
	}
	return response.Created(c, toResponse(item))
}

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

func (h *Handler) Get(c fiber.Ctx) error {
	item, err := h.get.Execute(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, toResponse(item))
}

func (h *Handler) UpdateQuantity(c fiber.Ctx) error {
	var req UpdateQuantityRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if err := validate.Struct(req); err != nil {
		return err
	}

	item, err := h.updateQuantity.Execute(c.Context(), c.Params("id"), req.Quantity)
	if err != nil {
		return err
	}
	return response.OK(c, toResponse(item))
}

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
