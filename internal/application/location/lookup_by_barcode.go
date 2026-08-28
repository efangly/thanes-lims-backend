package location

import (
	"context"

	domainlocation "github.com/efangly/thanes-lims-backend/internal/domain/location"
	portlocation "github.com/efangly/thanes-lims-backend/internal/ports/location"
)

type LookupByBarcodeUseCase struct {
	locations portlocation.LocationRepository
}

func NewLookupByBarcodeUseCase(locations portlocation.LocationRepository) *LookupByBarcodeUseCase {
	return &LookupByBarcodeUseCase{locations: locations}
}

// Execute resolves a scanned Location Barcode to its Location - e.g. to pick
// a destination when moving a Sample without navigating the tree by hand.
func (uc *LookupByBarcodeUseCase) Execute(ctx context.Context, code string) (domainlocation.Location, error) {
	return uc.locations.FindByBarcode(ctx, code)
}
