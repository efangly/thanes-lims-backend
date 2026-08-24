package location_test

import (
	"testing"

	"github.com/efangly/thanes-lims-backend/internal/domain/location"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/stretchr/testify/assert"
)

func TestLevelType_Valid(t *testing.T) {
	assert.True(t, location.LevelCabinet.Valid())
	assert.True(t, location.LevelShelf.Valid())
	assert.True(t, location.LevelSlot.Valid())
	assert.True(t, location.LevelSubSlot.Valid())
	assert.False(t, location.LevelType("shelf-unit").Valid())
}

func TestLevelType_Next(t *testing.T) {
	next, ok := location.LevelCabinet.Next()
	assert.True(t, ok)
	assert.Equal(t, location.LevelShelf, next)

	next, ok = location.LevelShelf.Next()
	assert.True(t, ok)
	assert.Equal(t, location.LevelSlot, next)

	next, ok = location.LevelSlot.Next()
	assert.True(t, ok)
	assert.Equal(t, location.LevelSubSlot, next)

	// sub_slot is the deepest level - it cannot be subdivided further.
	_, ok = location.LevelSubSlot.Next()
	assert.False(t, ok)
}

func TestLevelType_CanBeChildOf(t *testing.T) {
	assert.True(t, location.LevelShelf.CanBeChildOf(location.LevelCabinet))
	assert.True(t, location.LevelSlot.CanBeChildOf(location.LevelShelf))
	assert.True(t, location.LevelSubSlot.CanBeChildOf(location.LevelSlot))

	// levels cannot be skipped
	assert.False(t, location.LevelSlot.CanBeChildOf(location.LevelCabinet))
	assert.False(t, location.LevelSubSlot.CanBeChildOf(location.LevelCabinet))

	// a Cabinet is always root - nothing can be its parent, and it can't be
	// anyone's child
	assert.False(t, location.LevelCabinet.CanBeChildOf(location.LevelCabinet))
}

func TestValidateChild(t *testing.T) {
	cabinet := location.Location{ID: "LOC-1", LevelType: location.LevelCabinet}

	err := location.ValidateChild(cabinet, location.Location{LevelType: location.LevelShelf})
	assert.NoError(t, err)

	err = location.ValidateChild(cabinet, location.Location{LevelType: location.LevelSlot})
	assert.ErrorIs(t, err, shared.ErrValidation)

	err = location.ValidateChild(cabinet, location.Location{LevelType: location.LevelType("bogus")})
	assert.ErrorIs(t, err, shared.ErrValidation)
}

func TestLocation_IsRoot(t *testing.T) {
	assert.True(t, location.Location{ID: "LOC-1"}.IsRoot())

	parentID := "LOC-1"
	assert.False(t, location.Location{ID: "LOC-2", ParentID: &parentID}.IsRoot())
}
