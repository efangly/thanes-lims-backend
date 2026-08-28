package inventory

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/inventory"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portidgen "github.com/efangly/thanes-lims-backend/internal/ports/idgen"
	portinventory "github.com/efangly/thanes-lims-backend/internal/ports/inventory"
)

// ReceiveStockUseCase records goods arriving into the store: it creates a
// new InventoryLot for the item, or - when a lot with the same LotNo
// already exists - tops that lot up (CONTEXT.md "Inventory Lot":
// "สร้าง/เพิ่ม InventoryLot").
type ReceiveStockUseCase struct {
	items portinventory.Repository
	lots  portinventory.LotRepository
	idgen portidgen.SequenceGenerator
}

func NewReceiveStockUseCase(items portinventory.Repository, lots portinventory.LotRepository, idgen portidgen.SequenceGenerator) *ReceiveStockUseCase {
	return &ReceiveStockUseCase{items: items, lots: lots, idgen: idgen}
}

type ReceiveStockInput struct {
	ItemID     string
	LotNo      string
	ExpireDate *time.Time
	Quantity   int
}

// ReceiveStockResult carries both the affected lot and the item with its
// recomputed on-hand Quantity.
type ReceiveStockResult struct {
	Item inventory.InventoryItem
	Lot  inventory.InventoryLot
}

func (uc *ReceiveStockUseCase) Execute(ctx context.Context, in ReceiveStockInput) (ReceiveStockResult, error) {
	lotNo := strings.TrimSpace(in.LotNo)
	if lotNo == "" {
		return ReceiveStockResult{}, fmt.Errorf("%w: lot_no is required", shared.ErrValidation)
	}
	if in.Quantity <= 0 {
		return ReceiveStockResult{}, fmt.Errorf("%w: quantity must be greater than zero", shared.ErrValidation)
	}

	if _, err := uc.items.FindByID(ctx, in.ItemID); err != nil {
		return ReceiveStockResult{}, err
	}

	lot, err := uc.receiveLot(ctx, in.ItemID, lotNo, in.ExpireDate, in.Quantity)
	if err != nil {
		return ReceiveStockResult{}, err
	}

	item, err := uc.items.FindByID(ctx, in.ItemID)
	if err != nil {
		return ReceiveStockResult{}, err
	}
	return ReceiveStockResult{Item: item, Lot: lot}, nil
}

// receiveLot tops up an existing lot with the same number or creates a new
// one.
func (uc *ReceiveStockUseCase) receiveLot(ctx context.Context, itemID, lotNo string, expireDate *time.Time, qty int) (inventory.InventoryLot, error) {
	existing, err := uc.lots.FindByItemAndLotNo(ctx, itemID, lotNo)
	switch {
	case err == nil:
		return uc.lots.UpdateQuantity(ctx, existing.ID, existing.Quantity+qty)
	case errors.Is(err, shared.ErrNotFound):
		seq, gerr := uc.idgen.Next(ctx, "inventory_lot", nil)
		if gerr != nil {
			return inventory.InventoryLot{}, gerr
		}
		return uc.lots.Create(ctx, inventory.InventoryLot{
			ID:         fmt.Sprintf("LOT-%05d", seq),
			ItemID:     itemID,
			LotNo:      lotNo,
			ExpireDate: expireDate,
			Quantity:   qty,
		})
	default:
		return inventory.InventoryLot{}, err
	}
}
