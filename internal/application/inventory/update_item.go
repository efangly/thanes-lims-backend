package inventory

import (
	"context"
	"strings"

	"github.com/efangly/thanes-lims-backend/internal/domain/inventory"
	portinventory "github.com/efangly/thanes-lims-backend/internal/ports/inventory"
)

type UpdateItemUseCase struct {
	items      portinventory.Repository
	custodians portinventory.CustodianDirectory
	vendors    portinventory.VendorDirectory
	locations  portinventory.LocationDirectory
}

func NewUpdateItemUseCase(
	items portinventory.Repository,
	custodians portinventory.CustodianDirectory,
	vendors portinventory.VendorDirectory,
	locations portinventory.LocationDirectory,
) *UpdateItemUseCase {
	return &UpdateItemUseCase{items: items, custodians: custodians, vendors: vendors, locations: locations}
}

// UpdateItemInput carries a partial update: a nil field is left untouched, a
// non-nil field is written. Quantity moves only through UpdateQuantity;
// DefaultVendor only through UpdateDefaultVendor. VendorID / LocationID:
// nil = unchanged, non-nil "" = clear, non-nil value = set (validated).
type UpdateItemInput struct {
	ID              string
	Name            *string
	Category        *string
	Unit            *string
	Min             *int
	Max             *int
	CustodianUserID *int64
	Manufacturer    *string
	VendorID        *string
	LocationID      *string
}

func (uc *UpdateItemUseCase) Execute(ctx context.Context, in UpdateItemInput) (inventory.InventoryItem, error) {
	item, err := uc.items.FindByID(ctx, in.ID)
	if err != nil {
		return inventory.InventoryItem{}, err
	}

	if in.Name != nil {
		item.Name = strings.TrimSpace(*in.Name)
	}
	if in.Category != nil {
		item.Category = strings.TrimSpace(*in.Category)
	}
	if in.Unit != nil {
		item.Unit = strings.TrimSpace(*in.Unit)
	}
	if in.Min != nil {
		item.Min = *in.Min
	}
	if in.Max != nil {
		item.Max = *in.Max
	}
	if in.Manufacturer != nil {
		item.Manufacturer = strings.TrimSpace(*in.Manufacturer)
	}

	if in.CustodianUserID != nil {
		if err := validateCustodian(ctx, uc.custodians, *in.CustodianUserID); err != nil {
			return inventory.InventoryItem{}, err
		}
		item.CustodianUserID = *in.CustodianUserID
	}
	if in.VendorID != nil {
		ref := optionalRef(*in.VendorID)
		if err := validateVendor(ctx, uc.vendors, ref); err != nil {
			return inventory.InventoryItem{}, err
		}
		item.VendorID = ref
	}
	if in.LocationID != nil {
		ref := optionalRef(*in.LocationID)
		if err := validateLocation(ctx, uc.locations, ref); err != nil {
			return inventory.InventoryItem{}, err
		}
		item.LocationID = ref
	}

	return uc.items.Update(ctx, item)
}
