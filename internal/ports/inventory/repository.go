package inventory

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/inventory"
)

type Repository interface {
	Create(ctx context.Context, i inventory.InventoryItem) (inventory.InventoryItem, error)
	// FindByID and List populate the derived Quantity (sum of the item's
	// InventoryLots - CONTEXT.md "Inventory Lot"); Quantity is never written
	// through this repository.
	FindByID(ctx context.Context, id string) (inventory.InventoryItem, error)
	List(ctx context.Context) ([]inventory.InventoryItem, error)
	// Update persists a fully-loaded item, including cleared optional
	// fields (VendorID/LocationID set back to nil).
	Update(ctx context.Context, i inventory.InventoryItem) (inventory.InventoryItem, error)
	UpdateDefaultVendor(ctx context.Context, id string, vendor string) (inventory.InventoryItem, error)
}

// LotRepository persists InventoryLots - the batches whose quantities sum to
// an item's on-hand total. Receiving stock creates or tops up a lot; a
// Stock Issue (Phase 9) draws one down, possibly below zero (ADR 0008).
type LotRepository interface {
	Create(ctx context.Context, l inventory.InventoryLot) (inventory.InventoryLot, error)
	FindByID(ctx context.Context, id string) (inventory.InventoryLot, error)
	// FindByItemAndLotNo returns shared.ErrNotFound when the item has no lot
	// with that number yet.
	FindByItemAndLotNo(ctx context.Context, itemID, lotNo string) (inventory.InventoryLot, error)
	ListByItem(ctx context.Context, itemID string) ([]inventory.InventoryLot, error)
	UpdateQuantity(ctx context.Context, id string, quantity int) (inventory.InventoryLot, error)
}
