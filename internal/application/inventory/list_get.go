package inventory

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/inventory"
	portinventory "github.com/efangly/thanes-lims-backend/internal/ports/inventory"
)

type ListItemsUseCase struct {
	items portinventory.Repository
}

func NewListItemsUseCase(items portinventory.Repository) *ListItemsUseCase {
	return &ListItemsUseCase{items: items}
}

func (uc *ListItemsUseCase) Execute(ctx context.Context) ([]inventory.InventoryItem, error) {
	return uc.items.List(ctx)
}

type GetItemUseCase struct {
	items portinventory.Repository
}

func NewGetItemUseCase(items portinventory.Repository) *GetItemUseCase {
	return &GetItemUseCase{items: items}
}

func (uc *GetItemUseCase) Execute(ctx context.Context, id string) (inventory.InventoryItem, error) {
	return uc.items.FindByID(ctx, id)
}
