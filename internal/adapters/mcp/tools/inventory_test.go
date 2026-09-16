package tools_test

import (
	"context"
	"testing"

	"github.com/efangly/thanes-lims-backend/internal/adapters/mcp/tools"
	"github.com/efangly/thanes-lims-backend/internal/domain/inventory"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestInventoryTools_ListLowStock_FiltersBelowMin(t *testing.T) {
	items := new(mockInventoryRepository)
	it := &tools.InventoryTools{Items: items}

	items.On("List", mock.Anything).Return([]inventory.InventoryItem{
		{ID: "INV-0001", Name: "Glucose Reagent", Quantity: 5, Min: 20},
		{ID: "INV-0005", Name: "Well stocked", Quantity: 80, Min: 20},
		{ID: "INV-0002", Name: "ถุงมือไนไตรไซส์ M", Quantity: 3, Min: 15},
	}, nil)

	_, out, err := it.ListLowStock(context.Background(), nil, nil)

	assert.NoError(t, err)
	assert.Equal(t, 2, out.Count)
	ids := []string{out.Items[0].ID, out.Items[1].ID}
	assert.Contains(t, ids, "INV-0001")
	assert.Contains(t, ids, "INV-0002")
}

func TestInventoryTools_GetByID_IncludesLots(t *testing.T) {
	items := new(mockInventoryRepository)
	lots := new(mockLotRepository)
	it := &tools.InventoryTools{Items: items, Lots: lots}

	items.On("FindByID", mock.Anything, "INV-0002").Return(inventory.InventoryItem{ID: "INV-0002", Name: "ถุงมือไนไตรไซส์ M", Quantity: 3, Min: 15}, nil)
	lots.On("ListByItem", mock.Anything, "INV-0002").Return([]inventory.InventoryLot{
		{ID: "LOT-1", LotNo: "A100", Quantity: 3},
	}, nil)

	_, out, err := it.GetByID(context.Background(), nil, tools.GetInventoryItemByIDInput{ID: "INV-0002"})

	assert.NoError(t, err)
	assert.True(t, out.BelowMin)
	assert.Len(t, out.Lots, 1)
	assert.Equal(t, "A100", out.Lots[0].LotNo)
}

func TestInventoryTools_GetByID_NotFound(t *testing.T) {
	items := new(mockInventoryRepository)
	it := &tools.InventoryTools{Items: items}

	items.On("FindByID", mock.Anything, "MISSING").Return(inventory.InventoryItem{}, shared.ErrNotFound)

	result, _, err := it.GetByID(context.Background(), nil, tools.GetInventoryItemByIDInput{ID: "MISSING"})

	assert.NoError(t, err)
	assert.True(t, result.IsError)
}
