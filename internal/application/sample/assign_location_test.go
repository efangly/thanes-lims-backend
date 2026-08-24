package sample_test

import (
	"context"
	"testing"

	applicationsample "github.com/efangly/thanes-lims-backend/internal/application/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/location"
	"github.com/efangly/thanes-lims-backend/internal/domain/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAssignLocationUseCase_Success(t *testing.T) {
	samples := new(mockSampleRepo)
	locations := new(mockLocationRepo)

	samples.On("FindByID", mock.Anything, "SMP-1").Return(sample.Sample{ID: "SMP-1"}, nil)
	locations.On("GetByID", mock.Anything, "LOC-1").Return(location.Location{ID: "LOC-1", LevelType: location.LevelSlot}, nil)
	locations.On("HasChildren", mock.Anything, "LOC-1").Return(false, nil)
	samples.On("ExistsActiveByLocation", mock.Anything, "LOC-1").Return(false, nil)
	samples.On("UpdateLocation", mock.Anything, "SMP-1", mock.MatchedBy(func(id *string) bool {
		return id != nil && *id == "LOC-1"
	})).Return(sample.Sample{ID: "SMP-1", LocationID: strPtr("LOC-1")}, nil)

	uc := applicationsample.NewAssignLocationUseCase(samples, locations)
	got, err := uc.Execute(context.Background(), "SMP-1", "LOC-1")

	assert.NoError(t, err)
	assert.Equal(t, "LOC-1", *got.LocationID)
}

func TestAssignLocationUseCase_RejectsNonLeaf(t *testing.T) {
	samples := new(mockSampleRepo)
	locations := new(mockLocationRepo)

	samples.On("FindByID", mock.Anything, "SMP-1").Return(sample.Sample{ID: "SMP-1"}, nil)
	locations.On("GetByID", mock.Anything, "LOC-1").Return(location.Location{ID: "LOC-1", LevelType: location.LevelShelf}, nil)
	locations.On("HasChildren", mock.Anything, "LOC-1").Return(true, nil)

	uc := applicationsample.NewAssignLocationUseCase(samples, locations)
	_, err := uc.Execute(context.Background(), "SMP-1", "LOC-1")

	assert.ErrorIs(t, err, shared.ErrValidation)
	samples.AssertNotCalled(t, "UpdateLocation", mock.Anything, mock.Anything, mock.Anything)
}

func TestAssignLocationUseCase_RejectsAlreadyOccupiedLeaf(t *testing.T) {
	samples := new(mockSampleRepo)
	locations := new(mockLocationRepo)

	samples.On("FindByID", mock.Anything, "SMP-2").Return(sample.Sample{ID: "SMP-2"}, nil)
	locations.On("GetByID", mock.Anything, "LOC-1").Return(location.Location{ID: "LOC-1", LevelType: location.LevelSlot}, nil)
	locations.On("HasChildren", mock.Anything, "LOC-1").Return(false, nil)
	samples.On("ExistsActiveByLocation", mock.Anything, "LOC-1").Return(true, nil)

	uc := applicationsample.NewAssignLocationUseCase(samples, locations)
	_, err := uc.Execute(context.Background(), "SMP-2", "LOC-1")

	assert.ErrorIs(t, err, shared.ErrConflict)
	samples.AssertNotCalled(t, "UpdateLocation", mock.Anything, mock.Anything, mock.Anything)
}

func strPtr(s string) *string { return &s }
