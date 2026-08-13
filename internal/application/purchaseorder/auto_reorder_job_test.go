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

func newAutoReorderJob(pos *mockPORepo, items *mockInventoryRepo, idgen *mockIDGen) *applicationpurchaseorder.AutoReorderJob {
	reorder := applicationpurchaseorder.NewCreateFromLowStockUseCase(pos, items, idgen)
	return applicationpurchaseorder.NewAutoReorderJob(items, pos, reorder)
}

func TestAutoReorderJob_ReordersItemsBelowMinWithVendor(t *testing.T) {
	pos := new(mockPORepo)
	items := new(mockInventoryRepo)
	idgen := new(mockIDGen)

	lowStock := inventory.InventoryItem{ID: "INV-00001", Quantity: 5, Min: 20, Max: 100, DefaultVendor: "Acme Labs"}
	wellStocked := inventory.InventoryItem{ID: "INV-00002", Quantity: 80, Min: 20, Max: 100, DefaultVendor: "Acme Labs"}
	items.On("List", mock.Anything).Return([]inventory.InventoryItem{lowStock, wellStocked}, nil)
	items.On("FindByID", mock.Anything, "INV-00001").Return(lowStock, nil)
	pos.On("List", mock.Anything).Return([]purchaseorder.PurchaseOrder{}, nil)

	year := shared.BuddhistYear(time.Now())
	idgen.On("Next", mock.Anything, "purchase_order", &year).Return(int64(1), nil)
	pos.On("Create", mock.Anything, mock.MatchedBy(func(po purchaseorder.PurchaseOrder) bool {
		return po.ItemID == "INV-00001" && po.Vendor == "Acme Labs"
	})).Return(purchaseorder.PurchaseOrder{ItemID: "INV-00001", Vendor: "Acme Labs"}, nil)

	job := newAutoReorderJob(pos, items, idgen)
	result, err := job.Run(context.Background())

	assert.NoError(t, err)
	assert.Len(t, result.Created, 1)
	assert.Equal(t, "INV-00001", result.Created[0].ItemID)
	assert.Empty(t, result.SkippedNoVendor)
	assert.Empty(t, result.SkippedHasOpenPO)
	pos.AssertNotCalled(t, "Create", mock.Anything, mock.MatchedBy(func(po purchaseorder.PurchaseOrder) bool {
		return po.ItemID == "INV-00002"
	}))
}

func TestAutoReorderJob_SkipsItemsWithoutDefaultVendor(t *testing.T) {
	pos := new(mockPORepo)
	items := new(mockInventoryRepo)
	idgen := new(mockIDGen)

	noVendor := inventory.InventoryItem{ID: "INV-00003", Quantity: 2, Min: 20, Max: 100}
	items.On("List", mock.Anything).Return([]inventory.InventoryItem{noVendor}, nil)
	pos.On("List", mock.Anything).Return([]purchaseorder.PurchaseOrder{}, nil)

	job := newAutoReorderJob(pos, items, idgen)
	result, err := job.Run(context.Background())

	assert.NoError(t, err)
	assert.Empty(t, result.Created)
	assert.Equal(t, []string{"INV-00003"}, result.SkippedNoVendor)
	pos.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestAutoReorderJob_SkipsItemsWithOpenPO(t *testing.T) {
	pos := new(mockPORepo)
	items := new(mockInventoryRepo)
	idgen := new(mockIDGen)

	lowStock := inventory.InventoryItem{ID: "INV-00004", Quantity: 5, Min: 20, Max: 100, DefaultVendor: "Acme Labs"}
	items.On("List", mock.Anything).Return([]inventory.InventoryItem{lowStock}, nil)
	pos.On("List", mock.Anything).Return([]purchaseorder.PurchaseOrder{
		{ID: "PO-2569-0001", ItemID: "INV-00004", Status: purchaseorder.StatusPendingApproval},
	}, nil)

	job := newAutoReorderJob(pos, items, idgen)
	result, err := job.Run(context.Background())

	assert.NoError(t, err)
	assert.Empty(t, result.Created)
	assert.Equal(t, []string{"INV-00004"}, result.SkippedHasOpenPO)
	pos.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}
