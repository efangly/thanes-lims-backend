package purchaseorder

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/purchaseorder"
	portpurchaseorder "github.com/efangly/thanes-lims-backend/internal/ports/purchaseorder"
)

type ListPOsUseCase struct {
	purchaseOrders portpurchaseorder.Repository
}

func NewListPOsUseCase(purchaseOrders portpurchaseorder.Repository) *ListPOsUseCase {
	return &ListPOsUseCase{purchaseOrders: purchaseOrders}
}

func (uc *ListPOsUseCase) Execute(ctx context.Context) ([]purchaseorder.PurchaseOrder, error) {
	return uc.purchaseOrders.List(ctx)
}

type GetPOUseCase struct {
	purchaseOrders portpurchaseorder.Repository
}

func NewGetPOUseCase(purchaseOrders portpurchaseorder.Repository) *GetPOUseCase {
	return &GetPOUseCase{purchaseOrders: purchaseOrders}
}

func (uc *GetPOUseCase) Execute(ctx context.Context, id string) (purchaseorder.PurchaseOrder, error) {
	return uc.purchaseOrders.FindByID(ctx, id)
}
