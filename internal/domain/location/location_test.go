package location_test

import (
	"testing"

	"github.com/efangly/thanes-lims-backend/internal/domain/location"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/stretchr/testify/assert"
)

func TestKind_Valid(t *testing.T) {
	assert.True(t, location.KindSampleStorage.Valid())
	assert.True(t, location.KindEquipmentStorage.Valid())
	assert.False(t, location.Kind("warehouse").Valid())
}

func TestKind_RootLevel(t *testing.T) {
	lt, ok := location.KindSampleStorage.RootLevel()
	assert.True(t, ok)
	assert.Equal(t, location.LevelCabinet, lt)

	lt, ok = location.KindEquipmentStorage.RootLevel()
	assert.True(t, ok)
	assert.Equal(t, location.LevelBuilding, lt)

	_, ok = location.Kind("nope").RootLevel()
	assert.False(t, ok)
}

func TestChildLevel(t *testing.T) {
	next, ok := location.ChildLevel(location.KindSampleStorage, location.LevelCabinet)
	assert.True(t, ok)
	assert.Equal(t, location.LevelShelf, next)

	next, ok = location.ChildLevel(location.KindEquipmentStorage, location.LevelZone)
	assert.True(t, ok)
	assert.Equal(t, location.LevelCabinet, next)

	// deepest level of each kind cannot be subdivided
	_, ok = location.ChildLevel(location.KindSampleStorage, location.LevelSubSlot)
	assert.False(t, ok)
	_, ok = location.ChildLevel(location.KindEquipmentStorage, location.LevelShelf)
	assert.False(t, ok)

	// "cabinet" is depth 3 in equipment_storage, not the root
	next, ok = location.ChildLevel(location.KindEquipmentStorage, location.LevelCabinet)
	assert.True(t, ok)
	assert.Equal(t, location.LevelShelf, next)
}

func TestLevelValidForKind(t *testing.T) {
	assert.True(t, location.LevelValidForKind(location.KindSampleStorage, location.LevelSlot))
	assert.False(t, location.LevelValidForKind(location.KindSampleStorage, location.LevelBuilding))
	assert.True(t, location.LevelValidForKind(location.KindEquipmentStorage, location.LevelRoom))
	assert.False(t, location.LevelValidForKind(location.KindEquipmentStorage, location.LevelSlot))
}

func TestValidateChild(t *testing.T) {
	cabinet := location.Location{ID: "LOC-1", Kind: location.KindSampleStorage, LevelType: location.LevelCabinet}

	assert.NoError(t, location.ValidateChild(cabinet, location.Location{
		Kind: location.KindSampleStorage, LevelType: location.LevelShelf,
	}))

	// level skipped
	assert.ErrorIs(t, location.ValidateChild(cabinet, location.Location{
		Kind: location.KindSampleStorage, LevelType: location.LevelSlot,
	}), shared.ErrValidation)

	// kind mismatch
	assert.ErrorIs(t, location.ValidateChild(cabinet, location.Location{
		Kind: location.KindEquipmentStorage, LevelType: location.LevelShelf,
	}), shared.ErrValidation)

	// bogus level
	assert.ErrorIs(t, location.ValidateChild(cabinet, location.Location{
		Kind: location.KindSampleStorage, LevelType: location.LevelType("bogus"),
	}), shared.ErrValidation)

	building := location.Location{ID: "LOC-9", Kind: location.KindEquipmentStorage, LevelType: location.LevelBuilding}
	assert.NoError(t, location.ValidateChild(building, location.Location{
		Kind: location.KindEquipmentStorage, LevelType: location.LevelRoom,
	}))
}

func TestLocation_IsRoot(t *testing.T) {
	assert.True(t, location.Location{ID: "LOC-1"}.IsRoot())

	parentID := "LOC-1"
	assert.False(t, location.Location{ID: "LOC-2", ParentID: &parentID}.IsRoot())
}
