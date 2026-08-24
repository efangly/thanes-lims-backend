package location

import (
	"context"

	portlocation "github.com/efangly/thanes-lims-backend/internal/ports/location"
)

type GetFullPathUseCase struct {
	locations portlocation.LocationRepository
}

func NewGetFullPathUseCase(locations portlocation.LocationRepository) *GetFullPathUseCase {
	return &GetFullPathUseCase{locations: locations}
}

// Execute returns the Full Path for id (e.g. "Fridge-A / Shelf-2 / Slot-4"),
// computed from the tree - never stored. See CONTEXT.md#storage-location.
func (uc *GetFullPathUseCase) Execute(ctx context.Context, id string) (string, error) {
	return uc.locations.FullPath(ctx, id)
}
