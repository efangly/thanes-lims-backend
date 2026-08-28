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

func TestReceiveStock_CreatesNewLot(t *testing.T) {
	items := new(mockItemRepo)
	lots := new(mockLotRepo)

	items.On("FindByID", mock.Anything, "INV-00001").Return(inventory.InventoryItem{ID: "INV-00001"}, nil).Once()
	lots.On("FindByItemAndLotNo", mock.Anything, "INV-00001", "LOT-A").Return(inventory.InventoryLot{}, shared.ErrNotFound)
	lots.On("Create", mock.Anything, mock.MatchedBy(func(l inventory.InventoryLot) bool {
		return l.ItemID == "INV-00001" && l.LotNo == "LOT-A" && l.Quantity == 10
	})).Return(inventory.InventoryLot{ID: "LOT-00001", ItemID: "INV-00001", LotNo: "LOT-A", Quantity: 10}, nil)
	items.On("FindByID", mock.Anything, "INV-00001").Return(inventory.InventoryItem{ID: "INV-00001", Quantity: 10}, nil).Once()

	uc := applicationinventory.NewReceiveStockUseCase(items, lots, stubIDGen{})
	res, err := uc.Execute(context.Background(), applicationinventory.ReceiveStockInput{
		ItemID: "INV-00001", LotNo: "LOT-A", Quantity: 10,
	})

	assert.NoError(t, err)
	assert.Equal(t, 10, res.Item.Quantity)
	assert.Equal(t, "LOT-00001", res.Lot.ID)
}

func TestReceiveStock_TopsUpExistingLot(t *testing.T) {
	items := new(mockItemRepo)
	lots := new(mockLotRepo)

	items.On("FindByID", mock.Anything, "INV-00001").Return(inventory.InventoryItem{ID: "INV-00001"}, nil)
	lots.On("FindByItemAndLotNo", mock.Anything, "INV-00001", "LOT-A").
		Return(inventory.InventoryLot{ID: "LOT-00001", Quantity: 5}, nil)
	lots.On("UpdateQuantity", mock.Anything, "LOT-00001", 15).
		Return(inventory.InventoryLot{ID: "LOT-00001", Quantity: 15}, nil)

	uc := applicationinventory.NewReceiveStockUseCase(items, lots, stubIDGen{})
	res, err := uc.Execute(context.Background(), applicationinventory.ReceiveStockInput{
		ItemID: "INV-00001", LotNo: "LOT-A", Quantity: 10,
	})

	assert.NoError(t, err)
	assert.Equal(t, 15, res.Lot.Quantity)
	lots.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestReceiveStock_RejectsNonPositiveQuantity(t *testing.T) {
	uc := applicationinventory.NewReceiveStockUseCase(new(mockItemRepo), new(mockLotRepo), stubIDGen{})
	_, err := uc.Execute(context.Background(), applicationinventory.ReceiveStockInput{
		ItemID: "INV-00001", LotNo: "LOT-A", Quantity: 0,
	})
	assert.ErrorIs(t, err, shared.ErrValidation)
}

func TestReceiveStock_RejectsEmptyLotNo(t *testing.T) {
	uc := applicationinventory.NewReceiveStockUseCase(new(mockItemRepo), new(mockLotRepo), stubIDGen{})
	_, err := uc.Execute(context.Background(), applicationinventory.ReceiveStockInput{
		ItemID: "INV-00001", LotNo: "  ", Quantity: 5,
	})
	assert.ErrorIs(t, err, shared.ErrValidation)
}
