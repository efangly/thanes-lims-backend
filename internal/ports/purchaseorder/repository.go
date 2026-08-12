package purchaseorder

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/purchaseorder"
)

type Repository interface {
	Create(ctx context.Context, po purchaseorder.PurchaseOrder) (purchaseorder.PurchaseOrder, error)
	FindByID(ctx context.Context, id string) (purchaseorder.PurchaseOrder, error)
	List(ctx context.Context) ([]purchaseorder.PurchaseOrder, error)
	Update(ctx context.Context, po purchaseorder.PurchaseOrder) (purchaseorder.PurchaseOrder, error)
}
