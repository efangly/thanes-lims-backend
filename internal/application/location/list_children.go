package location

import (
	"context"

	domainlocation "github.com/efangly/thanes-lims-backend/internal/domain/location"
	portlocation "github.com/efangly/thanes-lims-backend/internal/ports/location"
)

type ListChildrenUseCase struct {
	locations portlocation.LocationRepository
}

func NewListChildrenUseCase(locations portlocation.LocationRepository) *ListChildrenUseCase {
	return &ListChildrenUseCase{locations: locations}
}

// Execute lists the direct children of parentID. A nil parentID lists root
// Locations (Cabinets).
func (uc *ListChildrenUseCase) Execute(ctx context.Context, parentID *string) ([]domainlocation.Location, error) {
	return uc.locations.ListChildren(ctx, parentID)
}
