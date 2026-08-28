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

type GenerateChildrenUseCase struct {
	locations portlocation.LocationRepository
	idgen     portidgen.SequenceGenerator
}

func NewGenerateChildrenUseCase(locations portlocation.LocationRepository, idgen portidgen.SequenceGenerator) *GenerateChildrenUseCase {
	return &GenerateChildrenUseCase{locations: locations, idgen: idgen}
}

type GenerateChildrenInput struct {
	ParentID string
	Prefix   string
	Count    int
}

// Execute creates Count children under ParentID, named "{Prefix}-1"
// .. "{Prefix}-{Count}", one LevelType below the parent's - the "generate
// children" operation from CONTEXT.md#storage-location. Fails if the parent
// is already at the deepest level (sub_slot) or if any generated name
// collides with an existing sibling.
func (uc *GenerateChildrenUseCase) Execute(ctx context.Context, in GenerateChildrenInput) ([]domainlocation.Location, error) {
	if in.Count <= 0 {
		return nil, fmt.Errorf("%w: count must be positive", shared.ErrValidation)
	}

	parent, err := uc.locations.GetByID(ctx, in.ParentID)
	if err != nil {
		return nil, err
	}

	childLevel, ok := domainlocation.ChildLevel(parent.Kind, parent.LevelType)
	if !ok {
		return nil, fmt.Errorf("%w: %s cannot have children", shared.ErrValidation, parent.LevelType)
	}

	children := make([]domainlocation.Location, in.Count)
	for i := 0; i < in.Count; i++ {
		name := fmt.Sprintf("%s-%d", in.Prefix, i+1)
		if _, err := uc.locations.FindChildByName(ctx, &in.ParentID, name); err == nil {
			return nil, fmt.Errorf("%w: sibling %q already exists", shared.ErrConflict, name)
		} else if !errors.Is(err, shared.ErrNotFound) {
			return nil, err
		}

		id, err := nextLocationID(ctx, uc.idgen)
		if err != nil {
			return nil, err
		}
		barcode, err := nextBarcodeCode(ctx, uc.idgen)
		if err != nil {
			return nil, err
		}

		parentID := in.ParentID
		children[i] = domainlocation.Location{
			ID:          id,
			ParentID:    &parentID,
			Name:        name,
			Kind:        parent.Kind,
			LevelType:   childLevel,
			BarcodeCode: barcode,
		}
	}

	return uc.locations.CreateMany(ctx, children)
}
