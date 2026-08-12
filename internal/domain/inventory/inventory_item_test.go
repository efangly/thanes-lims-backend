package inventory_test

import (
	"testing"

	"github.com/efangly/thanes-lims-backend/internal/domain/inventory"
	"github.com/stretchr/testify/assert"
)

func TestInventoryItem_DerivedStatus(t *testing.T) {
	cases := []struct {
		name string
		item inventory.InventoryItem
		want inventory.Status
	}{
		{"well stocked", inventory.InventoryItem{Quantity: 80, Min: 20, Max: 100}, inventory.StatusOK},
		{"at min", inventory.InventoryItem{Quantity: 20, Min: 20, Max: 100}, inventory.StatusLow},
		{"at half min", inventory.InventoryItem{Quantity: 10, Min: 20, Max: 100}, inventory.StatusCritical},
		{"empty", inventory.InventoryItem{Quantity: 0, Min: 20, Max: 100}, inventory.StatusCritical},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.want, tc.item.DerivedStatus(), tc.name)
	}
}

func TestInventoryItem_Pct(t *testing.T) {
	item := inventory.InventoryItem{Quantity: 50, Max: 100}
	assert.Equal(t, 50, item.Pct())

	overMax := inventory.InventoryItem{Quantity: 150, Max: 100}
	assert.Equal(t, 100, overMax.Pct())

	zeroMax := inventory.InventoryItem{Quantity: 10, Max: 0}
	assert.Equal(t, 0, zeroMax.Pct())
}

func TestInventoryItem_BelowMin(t *testing.T) {
	assert.True(t, inventory.InventoryItem{Quantity: 5, Min: 10}.BelowMin())
	assert.False(t, inventory.InventoryItem{Quantity: 15, Min: 10}.BelowMin())
}
