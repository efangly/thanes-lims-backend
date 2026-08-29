package sample_test

import (
	"context"
	"testing"

	applicationsample "github.com/efangly/thanes-lims-backend/internal/application/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/location"
	"github.com/efangly/thanes-lims-backend/internal/domain/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portsample "github.com/efangly/thanes-lims-backend/internal/ports/sample"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func occupant(id, pos string) sample.Sample {
	p := pos
	return sample.Sample{ID: id, Status: sample.StatusPending, LocationID: strPtr("BOX-1"), Position: &p}
}

func TestMoveWithinBox_SwapSucceeds(t *testing.T) {
	samples := new(mockSampleRepo)
	locations := new(mockLocationRepo)

	locations.On("GetByID", mock.Anything, "BOX-1").Return(boxLoc(), nil)
	samples.On("ListActiveByLocation", mock.Anything, "BOX-1").
		Return([]sample.Sample{occupant("S1", "A1"), occupant("S2", "A2")}, nil)
	moves := []portsample.PositionAssignment{{SampleID: "S1", Position: "A2"}, {SampleID: "S2", Position: "A1"}}
	samples.On("MoveWithinBox", mock.Anything, "BOX-1", moves).
		Return([]sample.Sample{occupant("S1", "A2"), occupant("S2", "A1")}, nil)

	uc := applicationsample.NewMoveWithinBoxUseCase(samples, locations)
	_, err := uc.Execute(context.Background(), applicationsample.MoveWithinBoxInput{BoxID: "BOX-1", Moves: moves})

	assert.NoError(t, err)
}

func TestMoveWithinBox_RejectsForeignSample(t *testing.T) {
	samples := new(mockSampleRepo)
	locations := new(mockLocationRepo)

	locations.On("GetByID", mock.Anything, "BOX-1").Return(boxLoc(), nil)
	samples.On("ListActiveByLocation", mock.Anything, "BOX-1").Return([]sample.Sample{occupant("S1", "A1")}, nil)

	uc := applicationsample.NewMoveWithinBoxUseCase(samples, locations)
	_, err := uc.Execute(context.Background(), applicationsample.MoveWithinBoxInput{
		BoxID: "BOX-1", Moves: []portsample.PositionAssignment{{SampleID: "OTHER", Position: "B2"}},
	})

	assert.ErrorIs(t, err, shared.ErrValidation)
	samples.AssertNotCalled(t, "MoveWithinBox", mock.Anything, mock.Anything, mock.Anything)
}

func TestMoveWithinBox_RejectsClashWithStayingSample(t *testing.T) {
	samples := new(mockSampleRepo)
	locations := new(mockLocationRepo)

	locations.On("GetByID", mock.Anything, "BOX-1").Return(boxLoc(), nil)
	samples.On("ListActiveByLocation", mock.Anything, "BOX-1").
		Return([]sample.Sample{occupant("S1", "A1"), occupant("S2", "A2")}, nil)

	uc := applicationsample.NewMoveWithinBoxUseCase(samples, locations)
	_, err := uc.Execute(context.Background(), applicationsample.MoveWithinBoxInput{
		BoxID: "BOX-1", Moves: []portsample.PositionAssignment{{SampleID: "S1", Position: "A2"}},
	})

	assert.ErrorIs(t, err, shared.ErrConflict)
}

func TestMoveWithinBox_RejectsOutOfGrid(t *testing.T) {
	samples := new(mockSampleRepo)
	locations := new(mockLocationRepo)

	locations.On("GetByID", mock.Anything, "BOX-1").Return(boxLoc(), nil)
	samples.On("ListActiveByLocation", mock.Anything, "BOX-1").Return([]sample.Sample{occupant("S1", "A1")}, nil)

	uc := applicationsample.NewMoveWithinBoxUseCase(samples, locations)
	_, err := uc.Execute(context.Background(), applicationsample.MoveWithinBoxInput{
		BoxID: "BOX-1", Moves: []portsample.PositionAssignment{{SampleID: "S1", Position: "Z99"}},
	})

	assert.ErrorIs(t, err, shared.ErrValidation)
}

func TestMoveWithinBox_RejectsNonBox(t *testing.T) {
	samples := new(mockSampleRepo)
	locations := new(mockLocationRepo)

	locations.On("GetByID", mock.Anything, "LOC-1").
		Return(location.Location{ID: "LOC-1", LevelType: location.LevelSlot}, nil)

	uc := applicationsample.NewMoveWithinBoxUseCase(samples, locations)
	_, err := uc.Execute(context.Background(), applicationsample.MoveWithinBoxInput{
		BoxID: "LOC-1", Moves: []portsample.PositionAssignment{{SampleID: "S1", Position: "A1"}},
	})

	assert.ErrorIs(t, err, shared.ErrValidation)
}
