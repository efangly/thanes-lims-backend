package location_test

import (
	"testing"

	"github.com/efangly/thanes-lims-backend/internal/domain/location"
	"github.com/stretchr/testify/assert"
)

func TestCanParentBox(t *testing.T) {
	assert.True(t, location.CanParentBox(location.KindSampleStorage, location.LevelShelf))
	assert.True(t, location.CanParentBox(location.KindSampleStorage, location.LevelSlot))
	assert.True(t, location.CanParentBox(location.KindSampleStorage, location.LevelSubSlot))
	assert.False(t, location.CanParentBox(location.KindSampleStorage, location.LevelCabinet))
	assert.False(t, location.CanParentBox(location.KindSampleStorage, location.LevelBox))
	assert.False(t, location.CanParentBox(location.KindEquipmentStorage, location.LevelShelf))
}

func TestValidateBox(t *testing.T) {
	assert.NoError(t, location.ValidateBox(8, 12))
	assert.NoError(t, location.ValidateBox(1, 1))
	assert.NoError(t, location.ValidateBox(26, 99))
	assert.Error(t, location.ValidateBox(0, 12))
	assert.Error(t, location.ValidateBox(27, 12))
	assert.Error(t, location.ValidateBox(8, 0))
	assert.Error(t, location.ValidateBox(8, 100))
}

func TestParsePosition(t *testing.T) {
	cases := []struct {
		in       string
		row, col int
		ok       bool
	}{
		{"A1", 1, 1, true},
		{"H12", 8, 12, true},
		{"Z99", 26, 99, true},
		{"", 0, 0, false},
		{"1A", 0, 0, false},
		{"a1", 0, 0, false},
		{"A0", 0, 0, false},
		{"A01", 0, 0, false},
		{"AA1", 0, 0, false},
		{"A100", 0, 0, false},
	}
	for _, c := range cases {
		row, col, err := location.ParsePosition(c.in)
		if c.ok {
			assert.NoErrorf(t, err, "input %q", c.in)
			assert.Equal(t, c.row, row, "row for %q", c.in)
			assert.Equal(t, c.col, col, "col for %q", c.in)
		} else {
			assert.Errorf(t, err, "input %q should be rejected", c.in)
		}
	}
}

func TestPositionInGrid(t *testing.T) {
	box := location.Location{LevelType: location.LevelBox, Rows: 8, Cols: 12}
	assert.True(t, box.PositionInGrid("A1"))
	assert.True(t, box.PositionInGrid("H12"))
	assert.False(t, box.PositionInGrid("H13"))
	assert.False(t, box.PositionInGrid("I1"))
	assert.False(t, box.PositionInGrid("bad"))

	notBox := location.Location{LevelType: location.LevelSlot}
	assert.False(t, notBox.PositionInGrid("A1"))
}
