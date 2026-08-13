package inventory

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/inventory"
)

type Repository interface {
	Create(ctx context.Context, i inventory.InventoryItem) (inventory.InventoryItem, error)
	FindByID(ctx context.Context, id string) (inventory.InventoryItem, error)
	List(ctx context.Context) ([]inventory.InventoryItem, error)
	UpdateQuantity(ctx context.Context, id string, quantity int) (inventory.InventoryItem, error)
	UpdateDefaultVendor(ctx context.Context, id string, vendor string) (inventory.InventoryItem, error)
}
