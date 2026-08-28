package inventory

import (
	"context"
	"fmt"
	"strings"

	"github.com/efangly/thanes-lims-backend/internal/domain/inventory"
	portidgen "github.com/efangly/thanes-lims-backend/internal/ports/idgen"
	portinventory "github.com/efangly/thanes-lims-backend/internal/ports/inventory"
)

type CreateItemUseCase struct {
	items      portinventory.Repository
	idgen      portidgen.SequenceGenerator
	custodians portinventory.CustodianDirectory
	vendors    portinventory.VendorDirectory
	locations  portinventory.LocationDirectory
}

func NewCreateItemUseCase(
	items portinventory.Repository,
	idgen portidgen.SequenceGenerator,
	custodians portinventory.CustodianDirectory,
	vendors portinventory.VendorDirectory,
	locations portinventory.LocationDirectory,
) *CreateItemUseCase {
	return &CreateItemUseCase{items: items, idgen: idgen, custodians: custodians, vendors: vendors, locations: locations}
}

type CreateItemInput struct {
	Name          string
	Category      string
	Unit          string
	Min           int
	Max           int
	DefaultVendor string

	CustodianUserID int64
	Manufacturer    string
	VendorID        string
	LocationID      string
}

func (uc *CreateItemUseCase) Execute(ctx context.Context, in CreateItemInput) (inventory.InventoryItem, error) {
	vendorID := optionalRef(in.VendorID)
	locationID := optionalRef(in.LocationID)
	if err := validateCustodian(ctx, uc.custodians, in.CustodianUserID); err != nil {
		return inventory.InventoryItem{}, err
	}
	if err := validateVendor(ctx, uc.vendors, vendorID); err != nil {
		return inventory.InventoryItem{}, err
	}
	if err := validateLocation(ctx, uc.locations, locationID); err != nil {
		return inventory.InventoryItem{}, err
	}

	seq, err := uc.idgen.Next(ctx, "inventory", nil)
	if err != nil {
		return inventory.InventoryItem{}, err
	}

	item := inventory.InventoryItem{
		// Quantity starts at zero: stock only enters through received lots
		// (ReceiveStockUseCase), never a create-time field.
		ID:            fmt.Sprintf("INV-%05d", seq),
		Name:          in.Name,
		Category:      in.Category,
		Unit:          in.Unit,
		Min:           in.Min,
		Max:           in.Max,
		DefaultVendor: in.DefaultVendor,

		CustodianUserID: in.CustodianUserID,
		Manufacturer:    strings.TrimSpace(in.Manufacturer),
		VendorID:        vendorID,
		LocationID:      locationID,
	}

	return uc.items.Create(ctx, item)
}
