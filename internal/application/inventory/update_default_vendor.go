package inventory

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/inventory"
	portinventory "github.com/efangly/thanes-lims-backend/internal/ports/inventory"
)

type UpdateDefaultVendorUseCase struct {
	items portinventory.Repository
}

func NewUpdateDefaultVendorUseCase(items portinventory.Repository) *UpdateDefaultVendorUseCase {
	return &UpdateDefaultVendorUseCase{items: items}
}

func (uc *UpdateDefaultVendorUseCase) Execute(ctx context.Context, id string, vendor string) (inventory.InventoryItem, error) {
	return uc.items.UpdateDefaultVendor(ctx, id, vendor)
}
