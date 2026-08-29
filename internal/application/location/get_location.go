package location

import (
	"context"

	domainlocation "github.com/efangly/thanes-lims-backend/internal/domain/location"
	portlocation "github.com/efangly/thanes-lims-backend/internal/ports/location"
)

type GetLocationUseCase struct {
	locations portlocation.LocationRepository
}

func NewGetLocationUseCase(locations portlocation.LocationRepository) *GetLocationUseCase {
	return &GetLocationUseCase{locations: locations}
}

// Execute returns a single Location by id - used by the frontend to resolve a
// deep-linked node (e.g. a Box, whose Grid the tree-drill-down does not carry).
func (uc *GetLocationUseCase) Execute(ctx context.Context, id string) (domainlocation.Location, error) {
	return uc.locations.GetByID(ctx, id)
}
