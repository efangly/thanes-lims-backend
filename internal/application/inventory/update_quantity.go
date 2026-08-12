package inventory

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/inventory"
	portinventory "github.com/efangly/thanes-lims-backend/internal/ports/inventory"
)

type UpdateQuantityUseCase struct {
	items portinventory.Repository
}

func NewUpdateQuantityUseCase(items portinventory.Repository) *UpdateQuantityUseCase {
	return &UpdateQuantityUseCase{items: items}
}

func (uc *UpdateQuantityUseCase) Execute(ctx context.Context, id string, quantity int) (inventory.InventoryItem, error) {
	return uc.items.UpdateQuantity(ctx, id, quantity)
}
