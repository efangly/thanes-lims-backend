package purchaseorder

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/inventory"
	"github.com/efangly/thanes-lims-backend/internal/domain/purchaseorder"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portidgen "github.com/efangly/thanes-lims-backend/internal/ports/idgen"
	portinventory "github.com/efangly/thanes-lims-backend/internal/ports/inventory"
	portpurchaseorder "github.com/efangly/thanes-lims-backend/internal/ports/purchaseorder"
)

type MarkReceivedUseCase struct {
	purchaseOrders portpurchaseorder.Repository
	inventory      portinventory.Repository
	lots           portinventory.LotRepository
	idgen          portidgen.SequenceGenerator
}

func NewMarkReceivedUseCase(
	purchaseOrders portpurchaseorder.Repository,
	inventory portinventory.Repository,
	lots portinventory.LotRepository,
	idgen portidgen.SequenceGenerator,
) *MarkReceivedUseCase {
	return &MarkReceivedUseCase{purchaseOrders: purchaseOrders, inventory: inventory, lots: lots, idgen: idgen}
}

type MarkReceivedInput struct {
	ID         string
	LotNo      string
	ExpireDate *time.Time
}

// Execute is the cross-aggregate orchestration point: marking a PO received
// also books the delivered goods into the linked InventoryItem as an
// InventoryLot (Phase 8 - stock only ever enters through lots). A lot number
// matching an existing lot for the item tops that lot up.
func (uc *MarkReceivedUseCase) Execute(ctx context.Context, in MarkReceivedInput) (purchaseorder.PurchaseOrder, error) {
	lotNo := strings.TrimSpace(in.LotNo)
	if lotNo == "" {
		return purchaseorder.PurchaseOrder{}, fmt.Errorf("%w: lot_no is required", shared.ErrValidation)
	}

	po, err := uc.purchaseOrders.FindByID(ctx, in.ID)
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

	existing, err := uc.lots.FindByItemAndLotNo(ctx, item.ID, lotNo)
	switch {
	case err == nil:
		if _, uerr := uc.lots.UpdateQuantity(ctx, existing.ID, existing.Quantity+po.Quantity); uerr != nil {
			return purchaseorder.PurchaseOrder{}, uerr
		}
	case errors.Is(err, shared.ErrNotFound):
		seq, gerr := uc.idgen.Next(ctx, "inventory_lot", nil)
		if gerr != nil {
			return purchaseorder.PurchaseOrder{}, gerr
		}
		if _, cerr := uc.lots.Create(ctx, inventory.InventoryLot{
			ID:         fmt.Sprintf("LOT-%05d", seq),
			ItemID:     item.ID,
			LotNo:      lotNo,
			ExpireDate: in.ExpireDate,
			Quantity:   po.Quantity,
		}); cerr != nil {
			return purchaseorder.PurchaseOrder{}, cerr
		}
	default:
		return purchaseorder.PurchaseOrder{}, err
	}

	po.Status = purchaseorder.StatusReceived
	return uc.purchaseOrders.Update(ctx, po)
}
