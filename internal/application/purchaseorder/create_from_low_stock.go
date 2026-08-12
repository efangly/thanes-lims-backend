package purchaseorder

import (
	"context"
	"fmt"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/purchaseorder"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portidgen "github.com/efangly/thanes-lims-backend/internal/ports/idgen"
	portinventory "github.com/efangly/thanes-lims-backend/internal/ports/inventory"
	portpurchaseorder "github.com/efangly/thanes-lims-backend/internal/ports/purchaseorder"
)

type CreateFromLowStockUseCase struct {
	purchaseOrders portpurchaseorder.Repository
	inventory      portinventory.Repository
	idgen          portidgen.SequenceGenerator
}

func NewCreateFromLowStockUseCase(purchaseOrders portpurchaseorder.Repository, inventory portinventory.Repository, idgen portidgen.SequenceGenerator) *CreateFromLowStockUseCase {
	return &CreateFromLowStockUseCase{purchaseOrders: purchaseOrders, inventory: inventory, idgen: idgen}
}

type CreateFromLowStockInput struct {
	ItemID string
	Vendor string
}

// Execute is a manual reorder action (POST /inventory/:id/reorder) rather
// than an automated cron trigger for MVP - reorder quantity is topped up to
// the item's configured Max.
func (uc *CreateFromLowStockUseCase) Execute(ctx context.Context, in CreateFromLowStockInput) (purchaseorder.PurchaseOrder, error) {
	item, err := uc.inventory.FindByID(ctx, in.ItemID)
	if err != nil {
		return purchaseorder.PurchaseOrder{}, err
	}
	if !item.BelowMin() {
		return purchaseorder.PurchaseOrder{}, fmt.Errorf("%w: item is not below its minimum threshold", shared.ErrValidation)
	}

	now := time.Now()
	year := shared.BuddhistYear(now)
	seq, err := uc.idgen.Next(ctx, "purchase_order", &year)
	if err != nil {
		return purchaseorder.PurchaseOrder{}, err
	}

	po := purchaseorder.PurchaseOrder{
		ID:        fmt.Sprintf("PO-%d-%04d", year, seq),
		ItemID:    item.ID,
		Quantity:  item.Max - item.Quantity,
		Vendor:    in.Vendor,
		OrderDate: now,
		Status:    purchaseorder.StatusPendingApproval,
	}

	return uc.purchaseOrders.Create(ctx, po)
}
