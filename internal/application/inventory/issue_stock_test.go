package inventory_test

import (
	"context"
	"testing"

	applicationinventory "github.com/efangly/thanes-lims-backend/internal/application/inventory"
	"github.com/efangly/thanes-lims-backend/internal/domain/inventory"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestIssueStock_DrawsDownChosenLot(t *testing.T) {
	items := new(mockItemRepo)
	lots := new(mockLotRepo)

	items.On("FindByID", mock.Anything, "INV-1").Return(inventory.InventoryItem{ID: "INV-1"}, nil).Once()
	lots.On("FindByID", mock.Anything, "LOT-1").Return(inventory.InventoryLot{ID: "LOT-1", ItemID: "INV-1", Quantity: 10}, nil)
	lots.On("UpdateQuantity", mock.Anything, "LOT-1", 7).Return(inventory.InventoryLot{ID: "LOT-1", ItemID: "INV-1", Quantity: 7}, nil)
	items.On("FindByID", mock.Anything, "INV-1").Return(inventory.InventoryItem{ID: "INV-1", Quantity: 7}, nil).Once()

	uc := applicationinventory.NewIssueStockUseCase(items, lots)
	res, err := uc.Execute(context.Background(), applicationinventory.IssueStockInput{
		ItemID: "INV-1",
		Lines:  []applicationinventory.IssueLine{{LotID: "LOT-1", Quantity: 3}},
	})

	assert.NoError(t, err)
	assert.True(t, res.Applied)
	assert.Empty(t, res.Shortfalls)
	assert.Equal(t, 7, res.Item.Quantity)
}

func TestIssueStock_OverIssueWithoutForce_ReportsBalanceAndAppliesNothing(t *testing.T) {
	items := new(mockItemRepo)
	lots := new(mockLotRepo)

	items.On("FindByID", mock.Anything, "INV-1").Return(inventory.InventoryItem{ID: "INV-1", Quantity: 4}, nil)
	lots.On("FindByID", mock.Anything, "LOT-1").Return(inventory.InventoryLot{ID: "LOT-1", ItemID: "INV-1", LotNo: "A", Quantity: 4}, nil)

	uc := applicationinventory.NewIssueStockUseCase(items, lots)
	res, err := uc.Execute(context.Background(), applicationinventory.IssueStockInput{
		ItemID: "INV-1",
		Lines:  []applicationinventory.IssueLine{{LotID: "LOT-1", Quantity: 10}},
	})

	assert.NoError(t, err)
	assert.False(t, res.Applied)
	assert.Len(t, res.Shortfalls, 1)
	assert.Equal(t, 4, res.Shortfalls[0].Available)
	assert.Equal(t, 10, res.Shortfalls[0].Requested)
	lots.AssertNotCalled(t, "UpdateQuantity", mock.Anything, mock.Anything, mock.Anything)
}

func TestIssueStock_ForceAllowsNegativeQuantity(t *testing.T) {
	items := new(mockItemRepo)
	lots := new(mockLotRepo)

	items.On("FindByID", mock.Anything, "INV-1").Return(inventory.InventoryItem{ID: "INV-1"}, nil).Once()
	lots.On("FindByID", mock.Anything, "LOT-1").Return(inventory.InventoryLot{ID: "LOT-1", ItemID: "INV-1", Quantity: 4}, nil)
	lots.On("UpdateQuantity", mock.Anything, "LOT-1", -6).Return(inventory.InventoryLot{ID: "LOT-1", ItemID: "INV-1", Quantity: -6}, nil)
	items.On("FindByID", mock.Anything, "INV-1").Return(inventory.InventoryItem{ID: "INV-1", Quantity: -6}, nil).Once()

	uc := applicationinventory.NewIssueStockUseCase(items, lots)
	res, err := uc.Execute(context.Background(), applicationinventory.IssueStockInput{
		ItemID: "INV-1",
		Lines:  []applicationinventory.IssueLine{{LotID: "LOT-1", Quantity: 10}},
		Force:  true,
	})

	assert.NoError(t, err)
	assert.True(t, res.Applied)
	assert.Equal(t, -6, res.Lots[0].Quantity)
}

func TestIssueStock_MultiLineAcrossLots(t *testing.T) {
	items := new(mockItemRepo)
	lots := new(mockLotRepo)

	items.On("FindByID", mock.Anything, "INV-1").Return(inventory.InventoryItem{ID: "INV-1"}, nil).Once()
	lots.On("FindByID", mock.Anything, "LOT-1").Return(inventory.InventoryLot{ID: "LOT-1", ItemID: "INV-1", Quantity: 5}, nil)
	lots.On("FindByID", mock.Anything, "LOT-2").Return(inventory.InventoryLot{ID: "LOT-2", ItemID: "INV-1", Quantity: 8}, nil)
	lots.On("UpdateQuantity", mock.Anything, "LOT-1", 0).Return(inventory.InventoryLot{ID: "LOT-1", ItemID: "INV-1", Quantity: 0}, nil)
	lots.On("UpdateQuantity", mock.Anything, "LOT-2", 5).Return(inventory.InventoryLot{ID: "LOT-2", ItemID: "INV-1", Quantity: 5}, nil)
	items.On("FindByID", mock.Anything, "INV-1").Return(inventory.InventoryItem{ID: "INV-1", Quantity: 5}, nil).Once()

	uc := applicationinventory.NewIssueStockUseCase(items, lots)
	res, err := uc.Execute(context.Background(), applicationinventory.IssueStockInput{
		ItemID: "INV-1",
		Lines: []applicationinventory.IssueLine{
			{LotID: "LOT-1", Quantity: 5},
			{LotID: "LOT-2", Quantity: 3},
		},
	})

	assert.NoError(t, err)
	assert.True(t, res.Applied)
	assert.Equal(t, 5, res.Item.Quantity)
}

func TestIssueStock_RejectsLotFromAnotherItem(t *testing.T) {
	items := new(mockItemRepo)
	lots := new(mockLotRepo)

	items.On("FindByID", mock.Anything, "INV-1").Return(inventory.InventoryItem{ID: "INV-1"}, nil)
	lots.On("FindByID", mock.Anything, "LOT-9").Return(inventory.InventoryLot{ID: "LOT-9", ItemID: "INV-2", Quantity: 5}, nil)

	uc := applicationinventory.NewIssueStockUseCase(items, lots)
	_, err := uc.Execute(context.Background(), applicationinventory.IssueStockInput{
		ItemID: "INV-1",
		Lines:  []applicationinventory.IssueLine{{LotID: "LOT-9", Quantity: 1}},
	})

	assert.ErrorIs(t, err, shared.ErrValidation)
}

func TestIssueStock_RejectsDuplicateLotLine(t *testing.T) {
	items := new(mockItemRepo)
	lots := new(mockLotRepo)

	items.On("FindByID", mock.Anything, "INV-1").Return(inventory.InventoryItem{ID: "INV-1"}, nil)
	lots.On("FindByID", mock.Anything, "LOT-1").Return(inventory.InventoryLot{ID: "LOT-1", ItemID: "INV-1", Quantity: 5}, nil)

	uc := applicationinventory.NewIssueStockUseCase(items, lots)
	_, err := uc.Execute(context.Background(), applicationinventory.IssueStockInput{
		ItemID: "INV-1",
		Lines: []applicationinventory.IssueLine{
			{LotID: "LOT-1", Quantity: 1},
			{LotID: "LOT-1", Quantity: 2},
		},
	})

	assert.ErrorIs(t, err, shared.ErrValidation)
}

func TestIssueStock_RejectsEmptyLines(t *testing.T) {
	uc := applicationinventory.NewIssueStockUseCase(new(mockItemRepo), new(mockLotRepo))
	_, err := uc.Execute(context.Background(), applicationinventory.IssueStockInput{ItemID: "INV-1"})
	assert.ErrorIs(t, err, shared.ErrValidation)
}
