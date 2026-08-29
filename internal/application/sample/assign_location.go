package sample

import (
	"context"
	"fmt"

	"github.com/efangly/thanes-lims-backend/internal/domain/location"
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

// ExecuteByBarcode resolves a scanned Location Barcode to its Location id
// and then assigns it exactly like Execute - lets a caller move a Sample by
// scanning the destination shelf (or box) instead of passing its LocationID.
func (uc *AssignLocationUseCase) ExecuteByBarcode(ctx context.Context, sampleID, barcodeCode string, position *string) (sample.Sample, error) {
	loc, err := uc.locations.FindByBarcode(ctx, barcodeCode)
	if err != nil {
		return sample.Sample{}, err
	}
	return uc.Execute(ctx, sampleID, loc.ID, position)
}

// Execute assigns sampleID to locationID - its put-away spot.
//
//   - A plain Leaf (no children, not a box): position must be nil, and no
//     other active Sample may occupy the Leaf.
//   - A Box: position is required, must fall inside the Box's Grid, and no
//     other active Sample may occupy that Cell.
//
// Cross-box moves are ordinary Put-away and go through this path too - only
// rearranging Cells inside one Box is a MoveWithinBox (docs/adr/0009).
func (uc *AssignLocationUseCase) Execute(ctx context.Context, sampleID, locationID string, position *string) (sample.Sample, error) {
	if _, err := uc.samples.FindByID(ctx, sampleID); err != nil {
		return sample.Sample{}, err
	}

	loc, err := uc.locations.GetByID(ctx, locationID)
	if err != nil {
		return sample.Sample{}, err
	}
	if loc.Kind != "" && loc.Kind != location.KindSampleStorage {
		return sample.Sample{}, fmt.Errorf("%w: location is not sample storage", shared.ErrValidation)
	}

	if loc.IsBox() {
		return uc.assignToBox(ctx, sampleID, loc, position)
	}

	if position != nil {
		return sample.Sample{}, fmt.Errorf("%w: position is only valid for a box location", shared.ErrValidation)
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

	return uc.samples.UpdateLocation(ctx, sampleID, &locationID, nil)
}

func (uc *AssignLocationUseCase) assignToBox(ctx context.Context, sampleID string, box location.Location, position *string) (sample.Sample, error) {
	if position == nil || *position == "" {
		return sample.Sample{}, fmt.Errorf("%w: a box location requires a cell position", shared.ErrValidation)
	}
	if !box.PositionInGrid(*position) {
		return sample.Sample{}, fmt.Errorf("%w: cell %q is outside the box grid", shared.ErrValidation, *position)
	}

	occupied, err := uc.samples.ExistsActiveByLocationPosition(ctx, box.ID, *position)
	if err != nil {
		return sample.Sample{}, err
	}
	if occupied {
		return sample.Sample{}, fmt.Errorf("%w: cell already occupied", shared.ErrConflict)
	}

	return uc.samples.UpdateLocation(ctx, sampleID, &box.ID, position)
}
