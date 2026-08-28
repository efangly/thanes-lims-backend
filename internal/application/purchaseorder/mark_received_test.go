package purchaseorder_test

import (
	"context"
	"testing"

	applicationpurchaseorder "github.com/efangly/thanes-lims-backend/internal/application/purchaseorder"
	"github.com/efangly/thanes-lims-backend/internal/domain/inventory"
	"github.com/efangly/thanes-lims-backend/internal/domain/purchaseorder"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMarkReceivedUseCase_BooksDeliveredGoodsAsLot(t *testing.T) {
	pos := new(mockPORepo)
	items := new(mockInventoryRepo)
	lots := new(mockLotRepo)
	idgen := new(mockIDGen)

	po := purchaseorder.PurchaseOrder{ID: "PO-2569-0001", ItemID: "INV-00001", Quantity: 50, Status: purchaseorder.StatusSentToVendor}
	pos.On("FindByID", mock.Anything, "PO-2569-0001").Return(po, nil)
	items.On("FindByID", mock.Anything, "INV-00001").Return(inventory.InventoryItem{ID: "INV-00001"}, nil)
	lots.On("FindByItemAndLotNo", mock.Anything, "INV-00001", "LOT-9").Return(inventory.InventoryLot{}, shared.ErrNotFound)
	idgen.On("Next", mock.Anything, "inventory_lot", (*int)(nil)).Return(int64(7), nil)
	lots.On("Create", mock.Anything, mock.MatchedBy(func(l inventory.InventoryLot) bool {
		return l.ItemID == "INV-00001" && l.LotNo == "LOT-9" && l.Quantity == 50
	})).Return(inventory.InventoryLot{ID: "LOT-00007"}, nil)
	pos.On("Update", mock.Anything, mock.MatchedBy(func(p purchaseorder.PurchaseOrder) bool {
		return p.Status == purchaseorder.StatusReceived
	})).Return(purchaseorder.PurchaseOrder{ID: "PO-2569-0001", Status: purchaseorder.StatusReceived}, nil)

	uc := applicationpurchaseorder.NewMarkReceivedUseCase(pos, items, lots, idgen)
	updated, err := uc.Execute(context.Background(), applicationpurchaseorder.MarkReceivedInput{ID: "PO-2569-0001", LotNo: "LOT-9"})

	assert.NoError(t, err)
	assert.Equal(t, purchaseorder.StatusReceived, updated.Status)
	lots.AssertCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestMarkReceivedUseCase_WrongStatus(t *testing.T) {
	pos := new(mockPORepo)

	po := purchaseorder.PurchaseOrder{ID: "PO-2569-0001", Status: purchaseorder.StatusPendingApproval}
	pos.On("FindByID", mock.Anything, "PO-2569-0001").Return(po, nil)

	uc := applicationpurchaseorder.NewMarkReceivedUseCase(pos, new(mockInventoryRepo), new(mockLotRepo), new(mockIDGen))
	_, err := uc.Execute(context.Background(), applicationpurchaseorder.MarkReceivedInput{ID: "PO-2569-0001", LotNo: "LOT-9"})

	assert.ErrorIs(t, err, shared.ErrValidation)
}

func TestMarkReceivedUseCase_RequiresLotNo(t *testing.T) {
	uc := applicationpurchaseorder.NewMarkReceivedUseCase(new(mockPORepo), new(mockInventoryRepo), new(mockLotRepo), new(mockIDGen))
	_, err := uc.Execute(context.Background(), applicationpurchaseorder.MarkReceivedInput{ID: "PO-2569-0001"})

	assert.ErrorIs(t, err, shared.ErrValidation)
}
