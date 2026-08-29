package sample

import (
	"context"
	"fmt"

	"github.com/efangly/thanes-lims-backend/internal/domain/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portlocation "github.com/efangly/thanes-lims-backend/internal/ports/location"
	portsample "github.com/efangly/thanes-lims-backend/internal/ports/sample"
)

type MoveWithinBoxUseCase struct {
	samples   portsample.SampleRepository
	locations portlocation.LocationRepository
}

func NewMoveWithinBoxUseCase(samples portsample.SampleRepository, locations portlocation.LocationRepository) *MoveWithinBoxUseCase {
	return &MoveWithinBoxUseCase{samples: samples, locations: locations}
}

type MoveWithinBoxInput struct {
	BoxID string
	Moves []portsample.PositionAssignment
}

// Execute rearranges Cells inside one Box as a single atomic batch - a drag
// of a multi-selection or a two-Cell swap either all lands or none does. A
// resulting position clash fails the whole batch with shared.ErrConflict.
//
// Every Sample in the batch must already sit in this Box: moving a Sample in
// from elsewhere is ordinary Put-away, not a grid drag (docs/adr/0009).
func (uc *MoveWithinBoxUseCase) Execute(ctx context.Context, in MoveWithinBoxInput) ([]sample.Sample, error) {
	if len(in.Moves) == 0 {
		return nil, fmt.Errorf("%w: no moves given", shared.ErrValidation)
	}

	box, err := uc.locations.GetByID(ctx, in.BoxID)
	if err != nil {
		return nil, err
	}
	if !box.IsBox() {
		return nil, fmt.Errorf("%w: location is not a box", shared.ErrValidation)
	}

	occupants, err := uc.samples.ListActiveByLocation(ctx, in.BoxID)
	if err != nil {
		return nil, err
	}

	// Current Cell layout, and which Samples are actually in this Box.
	posBySample := make(map[string]string, len(occupants))
	occupiedBy := make(map[string]string, len(occupants))
	for _, s := range occupants {
		pos := ""
		if s.Position != nil {
			pos = *s.Position
		}
		posBySample[s.ID] = pos
		if pos != "" {
			occupiedBy[pos] = s.ID
		}
	}

	moved := make(map[string]bool, len(in.Moves))
	targets := make(map[string]string, len(in.Moves))
	for _, mv := range in.Moves {
		if _, ok := posBySample[mv.SampleID]; !ok {
			return nil, fmt.Errorf("%w: sample %s is not in this box - use put-away instead", shared.ErrValidation, mv.SampleID)
		}
		if !box.PositionInGrid(mv.Position) {
			return nil, fmt.Errorf("%w: cell %q is outside the box grid", shared.ErrValidation, mv.Position)
		}
		if moved[mv.SampleID] {
			return nil, fmt.Errorf("%w: sample %s listed twice", shared.ErrValidation, mv.SampleID)
		}
		if _, dup := targets[mv.Position]; dup {
			return nil, fmt.Errorf("%w: two samples target cell %q", shared.ErrConflict, mv.Position)
		}
		moved[mv.SampleID] = true
		targets[mv.Position] = mv.SampleID
	}

	// Freed Cells (Samples being moved out) open up; every target Cell must
	// then be either empty or freed by this same batch.
	for pos, sampleID := range targets {
		holder, taken := occupiedBy[pos]
		if taken && holder != sampleID && !moved[holder] {
			return nil, fmt.Errorf("%w: cell %q is already occupied", shared.ErrConflict, pos)
		}
	}

	return uc.samples.MoveWithinBox(ctx, in.BoxID, in.Moves)
}
