package location

import (
	"context"
	"errors"
	"fmt"

	domainlocation "github.com/efangly/thanes-lims-backend/internal/domain/location"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portidgen "github.com/efangly/thanes-lims-backend/internal/ports/idgen"
	portlocation "github.com/efangly/thanes-lims-backend/internal/ports/location"
)

type CreateBoxUseCase struct {
	locations portlocation.LocationRepository
	idgen     portidgen.SequenceGenerator
}

func NewCreateBoxUseCase(locations portlocation.LocationRepository, idgen portidgen.SequenceGenerator) *CreateBoxUseCase {
	return &CreateBoxUseCase{locations: locations, idgen: idgen}
}

type CreateBoxInput struct {
	ParentID string
	Name     string
	Rows     int
	Cols     int
}

// Execute creates a Box (level_type 'box') as a child of ParentID. The
// parent must be a Shelf, Slot, or Sub-slot in the Sample tree; a Box
// carries a Grid, never has child Locations, and its name must be unique
// among its siblings (docs/adr/0009).
func (uc *CreateBoxUseCase) Execute(ctx context.Context, in CreateBoxInput) (domainlocation.Location, error) {
	if in.Name == "" {
		return domainlocation.Location{}, fmt.Errorf("%w: box name is required", shared.ErrValidation)
	}
	if err := domainlocation.ValidateBox(in.Rows, in.Cols); err != nil {
		return domainlocation.Location{}, err
	}

	parent, err := uc.locations.GetByID(ctx, in.ParentID)
	if err != nil {
		return domainlocation.Location{}, err
	}
	if !domainlocation.CanParentBox(parent.Kind, parent.LevelType) {
		return domainlocation.Location{}, fmt.Errorf("%w: a box cannot hang off %s", shared.ErrValidation, parent.LevelType)
	}

	if _, err := uc.locations.FindChildByName(ctx, &in.ParentID, in.Name); err == nil {
		return domainlocation.Location{}, fmt.Errorf("%w: sibling %q already exists", shared.ErrConflict, in.Name)
	} else if !errors.Is(err, shared.ErrNotFound) {
		return domainlocation.Location{}, err
	}

	id, err := nextLocationID(ctx, uc.idgen)
	if err != nil {
		return domainlocation.Location{}, err
	}
	barcode, err := nextBarcodeCode(ctx, uc.idgen)
	if err != nil {
		return domainlocation.Location{}, err
	}

	parentID := in.ParentID
	return uc.locations.Create(ctx, domainlocation.Location{
		ID:          id,
		ParentID:    &parentID,
		Name:        in.Name,
		Kind:        parent.Kind,
		LevelType:   domainlocation.LevelBox,
		BarcodeCode: barcode,
		Rows:        in.Rows,
		Cols:        in.Cols,
	})
}
