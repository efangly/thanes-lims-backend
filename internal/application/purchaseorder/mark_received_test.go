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

func TestMarkReceivedUseCase_BumpsInventoryStock(t *testing.T) {
	pos := new(mockPORepo)
	items := new(mockInventoryRepo)

	po := purchaseorder.PurchaseOrder{ID: "PO-2569-0001", ItemID: "INV-00001", Quantity: 50, Status: purchaseorder.StatusSentToVendor}
	pos.On("FindByID", mock.Anything, "PO-2569-0001").Return(po, nil)

	item := inventory.InventoryItem{ID: "INV-00001", Quantity: 10, Max: 100}
	items.On("FindByID", mock.Anything, "INV-00001").Return(item, nil)
	items.On("UpdateQuantity", mock.Anything, "INV-00001", 60).Return(inventory.InventoryItem{ID: "INV-00001", Quantity: 60}, nil)

	pos.On("Update", mock.Anything, mock.MatchedBy(func(p purchaseorder.PurchaseOrder) bool {
		return p.Status == purchaseorder.StatusReceived
	})).Return(purchaseorder.PurchaseOrder{ID: "PO-2569-0001", Status: purchaseorder.StatusReceived}, nil)

	uc := applicationpurchaseorder.NewMarkReceivedUseCase(pos, items)
	updated, err := uc.Execute(context.Background(), "PO-2569-0001")

	assert.NoError(t, err)
	assert.Equal(t, purchaseorder.StatusReceived, updated.Status)
	items.AssertCalled(t, "UpdateQuantity", mock.Anything, "INV-00001", 60)
}

func TestMarkReceivedUseCase_WrongStatus(t *testing.T) {
	pos := new(mockPORepo)
	items := new(mockInventoryRepo)

	po := purchaseorder.PurchaseOrder{ID: "PO-2569-0001", Status: purchaseorder.StatusPendingApproval}
	pos.On("FindByID", mock.Anything, "PO-2569-0001").Return(po, nil)

	uc := applicationpurchaseorder.NewMarkReceivedUseCase(pos, items)
	_, err := uc.Execute(context.Background(), "PO-2569-0001")

	assert.ErrorIs(t, err, shared.ErrValidation)
}
