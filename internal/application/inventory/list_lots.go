package inventory

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/inventory"
	portinventory "github.com/efangly/thanes-lims-backend/internal/ports/inventory"
)

// ListLotsUseCase returns every InventoryLot for one item - the list a Stock
// Issue (Phase 9) picks from and the receiving page shows after a top-up.
type ListLotsUseCase struct {
	items portinventory.Repository
	lots  portinventory.LotRepository
}

func NewListLotsUseCase(items portinventory.Repository, lots portinventory.LotRepository) *ListLotsUseCase {
	return &ListLotsUseCase{items: items, lots: lots}
}

func (uc *ListLotsUseCase) Execute(ctx context.Context, itemID string) ([]inventory.InventoryLot, error) {
	if _, err := uc.items.FindByID(ctx, itemID); err != nil {
		return nil, err
	}
	return uc.lots.ListByItem(ctx, itemID)
}
