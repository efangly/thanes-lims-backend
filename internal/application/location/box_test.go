package location_test

import (
	"context"
	"testing"

	applicationlocation "github.com/efangly/thanes-lims-backend/internal/application/location"
	"github.com/efangly/thanes-lims-backend/internal/domain/location"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateBoxUseCase_Success(t *testing.T) {
	locations := new(mockLocationRepo)
	idgen := new(mockIDGen)

	locations.On("GetByID", mock.Anything, "LOC-SHELF").
		Return(location.Location{ID: "LOC-SHELF", Kind: location.KindSampleStorage, LevelType: location.LevelShelf}, nil)
	locations.On("FindChildByName", mock.Anything, mock.Anything, "Box-1").
		Return(location.Location{}, shared.ErrNotFound)
	idgen.On("Next", mock.Anything, "location", mock.Anything).Return(int64(7), nil)
	idgen.On("Next", mock.Anything, "location_barcode", mock.Anything).Return(int64(9), nil)
	locations.On("Create", mock.Anything, mock.MatchedBy(func(l location.Location) bool {
		return l.LevelType == location.LevelBox && l.Rows == 8 && l.Cols == 12 && l.Kind == location.KindSampleStorage
	})).Return(location.Location{ID: "LOC-00007", LevelType: location.LevelBox, Rows: 8, Cols: 12}, nil)

	uc := applicationlocation.NewCreateBoxUseCase(locations, idgen)
	got, err := uc.Execute(context.Background(), applicationlocation.CreateBoxInput{
		ParentID: "LOC-SHELF", Name: "Box-1", Rows: 8, Cols: 12,
	})

	assert.NoError(t, err)
	assert.Equal(t, location.LevelBox, got.LevelType)
}

func TestCreateBoxUseCase_RejectsBadParent(t *testing.T) {
	locations := new(mockLocationRepo)
	idgen := new(mockIDGen)

	locations.On("GetByID", mock.Anything, "LOC-CAB").
		Return(location.Location{ID: "LOC-CAB", Kind: location.KindSampleStorage, LevelType: location.LevelCabinet}, nil)

	uc := applicationlocation.NewCreateBoxUseCase(locations, idgen)
	_, err := uc.Execute(context.Background(), applicationlocation.CreateBoxInput{
		ParentID: "LOC-CAB", Name: "Box-1", Rows: 8, Cols: 12,
	})

	assert.ErrorIs(t, err, shared.ErrValidation)
	locations.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestCreateBoxUseCase_RejectsBadGrid(t *testing.T) {
	uc := applicationlocation.NewCreateBoxUseCase(new(mockLocationRepo), new(mockIDGen))
	_, err := uc.Execute(context.Background(), applicationlocation.CreateBoxInput{
		ParentID: "LOC-SHELF", Name: "Box-1", Rows: 0, Cols: 12,
	})
	assert.ErrorIs(t, err, shared.ErrValidation)
}

func TestCreateBoxUseCase_RejectsDuplicateName(t *testing.T) {
	locations := new(mockLocationRepo)
	idgen := new(mockIDGen)

	locations.On("GetByID", mock.Anything, "LOC-SHELF").
		Return(location.Location{ID: "LOC-SHELF", Kind: location.KindSampleStorage, LevelType: location.LevelShelf}, nil)
	locations.On("FindChildByName", mock.Anything, mock.Anything, "Box-1").
		Return(location.Location{ID: "LOC-EXIST"}, nil)

	uc := applicationlocation.NewCreateBoxUseCase(locations, idgen)
	_, err := uc.Execute(context.Background(), applicationlocation.CreateBoxInput{
		ParentID: "LOC-SHELF", Name: "Box-1", Rows: 8, Cols: 12,
	})

	assert.ErrorIs(t, err, shared.ErrConflict)
}

func TestEnlargeBoxUseCase_Grows(t *testing.T) {
	locations := new(mockLocationRepo)

	locations.On("GetByID", mock.Anything, "BOX-1").
		Return(location.Location{ID: "BOX-1", LevelType: location.LevelBox, Rows: 8, Cols: 12}, nil)
	locations.On("UpdateGrid", mock.Anything, "BOX-1", 10, 12).
		Return(location.Location{ID: "BOX-1", LevelType: location.LevelBox, Rows: 10, Cols: 12}, nil)

	uc := applicationlocation.NewEnlargeBoxUseCase(locations)
	got, err := uc.Execute(context.Background(), applicationlocation.EnlargeBoxInput{ID: "BOX-1", Rows: 10, Cols: 12})

	assert.NoError(t, err)
	assert.Equal(t, 10, got.Rows)
}

func TestEnlargeBoxUseCase_RejectsShrink(t *testing.T) {
	locations := new(mockLocationRepo)

	locations.On("GetByID", mock.Anything, "BOX-1").
		Return(location.Location{ID: "BOX-1", LevelType: location.LevelBox, Rows: 8, Cols: 12}, nil)

	uc := applicationlocation.NewEnlargeBoxUseCase(locations)
	_, err := uc.Execute(context.Background(), applicationlocation.EnlargeBoxInput{ID: "BOX-1", Rows: 8, Cols: 10})

	assert.ErrorIs(t, err, shared.ErrValidation)
	locations.AssertNotCalled(t, "UpdateGrid", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestEnlargeBoxUseCase_RejectsNonBox(t *testing.T) {
	locations := new(mockLocationRepo)

	locations.On("GetByID", mock.Anything, "LOC-1").
		Return(location.Location{ID: "LOC-1", LevelType: location.LevelSlot}, nil)

	uc := applicationlocation.NewEnlargeBoxUseCase(locations)
	_, err := uc.Execute(context.Background(), applicationlocation.EnlargeBoxInput{ID: "LOC-1", Rows: 8, Cols: 12})

	assert.ErrorIs(t, err, shared.ErrValidation)
}
