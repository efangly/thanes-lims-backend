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

func (h *Handler) Get(c fiber.Ctx) error {
	po, err := h.get.Execute(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, toResponse(po))
}

func (h *Handler) Approve(c fiber.Ctx) error {
	role := fiber.Locals[domainuser.Role](c, middleware.LocalsRole)
	po, err := h.approve.Execute(c.Context(), applicationpurchaseorder.ApprovePOInput{
		ID: c.Params("id"), ActorRole: role,
	})
	if err != nil {
		return err
	}
	return response.OK(c, toResponse(po))
}

func (h *Handler) MarkReceived(c fiber.Ctx) error {
	po, err := h.markReceived.Execute(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return response.OK(c, toResponse(po))
}
