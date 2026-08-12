package purchaseorder_test

import (
	"context"
	"testing"
	"time"

	applicationpurchaseorder "github.com/efangly/thanes-lims-backend/internal/application/purchaseorder"
	"github.com/efangly/thanes-lims-backend/internal/domain/inventory"
	"github.com/efangly/thanes-lims-backend/internal/domain/purchaseorder"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateFromLowStockUseCase_CreatesPOWithReorderQuantity(t *testing.T) {
	pos := new(mockPORepo)
	items := new(mockInventoryRepo)
	idgen := new(mockIDGen)

	item := inventory.InventoryItem{ID: "INV-00001", Quantity: 5, Min: 20, Max: 100}
	items.On("FindByID", mock.Anything, "INV-00001").Return(item, nil)

	year := shared.BuddhistYear(time.Now())
	idgen.On("Next", mock.Anything, "purchase_order", &year).Return(int64(42), nil)
	pos.On("Create", mock.Anything, mock.MatchedBy(func(po purchaseorder.PurchaseOrder) bool {
		return po.ItemID == "INV-00001" && po.Quantity == 95 && po.Status == purchaseorder.StatusPendingApproval
	})).Return(purchaseorder.PurchaseOrder{ItemID: "INV-00001", Quantity: 95}, nil)

	uc := applicationpurchaseorder.NewCreateFromLowStockUseCase(pos, items, idgen)
	po, err := uc.Execute(context.Background(), applicationpurchaseorder.CreateFromLowStockInput{
		ItemID: "INV-00001", Vendor: "Acme Labs",
	})

	assert.NoError(t, err)
	assert.Equal(t, 95, po.Quantity)
}

func TestCreateFromLowStockUseCase_NotBelowMin(t *testing.T) {
	pos := new(mockPORepo)
	items := new(mockInventoryRepo)
	idgen := new(mockIDGen)

	item := inventory.InventoryItem{ID: "INV-00001", Quantity: 80, Min: 20, Max: 100}
	items.On("FindByID", mock.Anything, "INV-00001").Return(item, nil)

	uc := applicationpurchaseorder.NewCreateFromLowStockUseCase(pos, items, idgen)
	_, err := uc.Execute(context.Background(), applicationpurchaseorder.CreateFromLowStockInput{
		ItemID: "INV-00001", Vendor: "Acme Labs",
	})

	assert.ErrorIs(t, err, shared.ErrValidation)
}
