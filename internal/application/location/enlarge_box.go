package location

import (
	"context"
	"fmt"

	domainlocation "github.com/efangly/thanes-lims-backend/internal/domain/location"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portlocation "github.com/efangly/thanes-lims-backend/internal/ports/location"
)

type EnlargeBoxUseCase struct {
	locations portlocation.LocationRepository
}

func NewEnlargeBoxUseCase(locations portlocation.LocationRepository) *EnlargeBoxUseCase {
	return &EnlargeBoxUseCase{locations: locations}
}

type EnlargeBoxInput struct {
	ID   string
	Rows int
	Cols int
}

// Execute grows a Box's Grid to Rows x Cols. Boxes only grow: shrinking would
// need an "are the trailing cells empty?" check for a case that barely
// happens - make a new box and move instead (docs/adr/0009).
func (uc *EnlargeBoxUseCase) Execute(ctx context.Context, in EnlargeBoxInput) (domainlocation.Location, error) {
	if err := domainlocation.ValidateBox(in.Rows, in.Cols); err != nil {
		return domainlocation.Location{}, err
	}

	box, err := uc.locations.GetByID(ctx, in.ID)
	if err != nil {
		return domainlocation.Location{}, err
	}
	if !box.IsBox() {
		return domainlocation.Location{}, fmt.Errorf("%w: location is not a box", shared.ErrValidation)
	}
	if in.Rows < box.Rows || in.Cols < box.Cols {
		return domainlocation.Location{}, fmt.Errorf(
			"%w: a box grid can only grow (from %dx%d)", shared.ErrValidation, box.Rows, box.Cols)
	}

	return uc.locations.UpdateGrid(ctx, in.ID, in.Rows, in.Cols)
}
