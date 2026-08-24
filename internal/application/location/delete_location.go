package location

import (
	"context"
	"fmt"

	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portlocation "github.com/efangly/thanes-lims-backend/internal/ports/location"
	portsample "github.com/efangly/thanes-lims-backend/internal/ports/sample"
)

type DeleteLocationUseCase struct {
	locations portlocation.LocationRepository
	samples   portsample.SampleRepository
}

func NewDeleteLocationUseCase(locations portlocation.LocationRepository, samples portsample.SampleRepository) *DeleteLocationUseCase {
	return &DeleteLocationUseCase{locations: locations, samples: samples}
}

// Execute deletes a Location. Restricted: it must have no children and no
// Sample (any status) referencing it - see CONTEXT.md#storage-location.
func (uc *DeleteLocationUseCase) Execute(ctx context.Context, id string) error {
	hasChildren, err := uc.locations.HasChildren(ctx, id)
	if err != nil {
		return err
	}
	if hasChildren {
		return fmt.Errorf("%w: location has children", shared.ErrConflict)
	}

	hasSamples, err := uc.samples.ExistsByLocation(ctx, id)
	if err != nil {
		return err
	}
	if hasSamples {
		return fmt.Errorf("%w: location has samples referencing it", shared.ErrConflict)
	}

	return uc.locations.Delete(ctx, id)
}
