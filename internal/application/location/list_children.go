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

// Execute lists the direct children of parentID. A nil parentID lists the
// root Locations of kind (defaulting to KindSampleStorage) - the two trees
// have separate roots, so a root listing is always scoped to one Kind.
func (uc *ListChildrenUseCase) Execute(ctx context.Context, parentID *string, kind domainlocation.Kind) ([]domainlocation.Location, error) {
	if parentID == nil {
		if kind == "" {
			kind = domainlocation.KindSampleStorage
		}
		return uc.locations.ListRoots(ctx, kind)
	}
	return uc.locations.ListChildren(ctx, parentID)
}
