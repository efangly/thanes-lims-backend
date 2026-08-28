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

type CreateCabinetUseCase struct {
	locations portlocation.LocationRepository
	idgen     portidgen.SequenceGenerator
}

func NewCreateCabinetUseCase(locations portlocation.LocationRepository, idgen portidgen.SequenceGenerator) *CreateCabinetUseCase {
	return &CreateCabinetUseCase{locations: locations, idgen: idgen}
}

type CreateCabinetInput struct {
	Name string
	// Kind selects which tree the root belongs to. Empty defaults to
	// KindSampleStorage for backward compatibility with existing callers.
	Kind domainlocation.Kind
}

// Execute creates a root Location - a Cabinet for KindSampleStorage, a
// Building for KindEquipmentStorage. Root names must be unique across the
// whole tree, not just among siblings - see CONTEXT.md#storage-location.
func (uc *CreateCabinetUseCase) Execute(ctx context.Context, in CreateCabinetInput) (domainlocation.Location, error) {
	kind := in.Kind
	if kind == "" {
		kind = domainlocation.KindSampleStorage
	}
	if !kind.Valid() {
		return domainlocation.Location{}, fmt.Errorf("%w: invalid location kind %q", shared.ErrValidation, in.Kind)
	}
	rootLevel, _ := kind.RootLevel()

	if _, err := uc.locations.FindChildByName(ctx, nil, in.Name); err == nil {
		return domainlocation.Location{}, shared.ErrConflict
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

	return uc.locations.Create(ctx, domainlocation.Location{
		ID:          id,
		Name:        in.Name,
		Kind:        kind,
		LevelType:   rootLevel,
		BarcodeCode: barcode,
	})
}
