package location

import (
	"context"
	"errors"

	domainlocation "github.com/efangly/thanes-lims-backend/internal/domain/location"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portidgen "github.com/efangly/thanes-lims-backend/internal/ports/idgen"
	portlocation "github.com/efangly/thanes-lims-backend/internal/ports/location"
)

type CreateCabinetUseCase struct {
	locations portlocation.LocationRepository
	idgen     portidgen.SequenceGenerator
}

func NewCreateCabinetUseCase(locations portlocation.LocationRepository, idgen portidgen.SequenceGenerator) *CreateCabinetUseCase {
	return &CreateCabinetUseCase{locations: locations, idgen: idgen}
}

type CreateCabinetInput struct {
	Name string
}

// Execute creates a root Location (Cabinet). Cabinet names must be unique
// across the whole tree, not just among siblings - see
// CONTEXT.md#storage-location.
func (uc *CreateCabinetUseCase) Execute(ctx context.Context, in CreateCabinetInput) (domainlocation.Location, error) {
	if _, err := uc.locations.FindChildByName(ctx, nil, in.Name); err == nil {
		return domainlocation.Location{}, shared.ErrConflict
	} else if !errors.Is(err, shared.ErrNotFound) {
		return domainlocation.Location{}, err
	}

	id, err := nextLocationID(ctx, uc.idgen)
	if err != nil {
		return domainlocation.Location{}, err
	}

	return uc.locations.Create(ctx, domainlocation.Location{
		ID:        id,
		Name:      in.Name,
		LevelType: domainlocation.LevelCabinet,
	})
}
