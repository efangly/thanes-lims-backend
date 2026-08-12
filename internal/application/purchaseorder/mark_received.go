package purchaseorder

import (
	"context"
	"fmt"

	"github.com/efangly/thanes-lims-backend/internal/domain/purchaseorder"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portinventory "github.com/efangly/thanes-lims-backend/internal/ports/inventory"
	portpurchaseorder "github.com/efangly/thanes-lims-backend/internal/ports/purchaseorder"
)

type MarkReceivedUseCase struct {
	purchaseOrders portpurchaseorder.Repository
	inventory      portinventory.Repository
}

func NewMarkReceivedUseCase(purchaseOrders portpurchaseorder.Repository, inventory portinventory.Repository) *MarkReceivedUseCase {
	return &MarkReceivedUseCase{purchaseOrders: purchaseOrders, inventory: inventory}
}

// Execute is the cross-aggregate orchestration point: marking a PO received
// also bumps the linked InventoryItem's stock back up.
func (uc *MarkReceivedUseCase) Execute(ctx context.Context, id string) (purchaseorder.PurchaseOrder, error) {
	po, err := uc.purchaseOrders.FindByID(ctx, id)
	if err != nil {
		return purchaseorder.PurchaseOrder{}, err
	}
	if po.Status != purchaseorder.StatusSentToVendor {
		return purchaseorder.PurchaseOrder{}, fmt.Errorf("%w: cannot mark PO received from status %s", shared.ErrValidation, po.Status)
	}

	item, err := uc.inventory.FindByID(ctx, po.ItemID)
	if err != nil {
		return purchaseorder.PurchaseOrder{}, err
	}
	if _, err := uc.inventory.UpdateQuantity(ctx, item.ID, item.Quantity+po.Quantity); err != nil {
		return purchaseorder.PurchaseOrder{}, err
	}

	po.Status = purchaseorder.StatusReceived
	return uc.purchaseOrders.Update(ctx, po)
}
