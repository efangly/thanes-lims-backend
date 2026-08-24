package sample

import (
	"context"
	"fmt"

	"github.com/efangly/thanes-lims-backend/internal/domain/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portlocation "github.com/efangly/thanes-lims-backend/internal/ports/location"
	portsample "github.com/efangly/thanes-lims-backend/internal/ports/sample"
)

type AssignLocationUseCase struct {
	samples   portsample.SampleRepository
	locations portlocation.LocationRepository
}

func NewAssignLocationUseCase(samples portsample.SampleRepository, locations portlocation.LocationRepository) *AssignLocationUseCase {
	return &AssignLocationUseCase{samples: samples, locations: locations}
}

// Execute assigns sampleID to locationID - its put-away spot. locationID
// must be a Leaf Location (no children) with no other active Sample already
// occupying it. See CONTEXT.md#storage-location.
func (uc *AssignLocationUseCase) Execute(ctx context.Context, sampleID, locationID string) (sample.Sample, error) {
	if _, err := uc.samples.FindByID(ctx, sampleID); err != nil {
		return sample.Sample{}, err
	}

	if _, err := uc.locations.GetByID(ctx, locationID); err != nil {
		return sample.Sample{}, err
	}

	hasChildren, err := uc.locations.HasChildren(ctx, locationID)
	if err != nil {
		return sample.Sample{}, err
	}
	if hasChildren {
		return sample.Sample{}, fmt.Errorf("%w: location is not a leaf", shared.ErrValidation)
	}

	occupied, err := uc.samples.ExistsActiveByLocation(ctx, locationID)
	if err != nil {
		return sample.Sample{}, err
	}
	if occupied {
		return sample.Sample{}, fmt.Errorf("%w: location already occupied", shared.ErrConflict)
	}

	return uc.samples.UpdateLocation(ctx, sampleID, &locationID)
}
