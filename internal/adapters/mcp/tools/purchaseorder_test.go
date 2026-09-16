package tools_test

import (
	"context"
	"testing"

	"github.com/efangly/thanes-lims-backend/internal/adapters/mcp/tools"
	"github.com/efangly/thanes-lims-backend/internal/domain/purchaseorder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func poFixture() []purchaseorder.PurchaseOrder {
	return []purchaseorder.PurchaseOrder{
		{ID: "PO-1", ItemID: "INV-0002", Status: purchaseorder.StatusReceived},
		{ID: "PO-2", ItemID: "INV-0002", Status: purchaseorder.StatusSentToVendor},
		{ID: "PO-3", ItemID: "INV-0005", Status: purchaseorder.StatusPendingApproval},
	}
}

func TestPurchaseOrderTools_List_FiltersByStatus(t *testing.T) {
	repo := new(mockPurchaseOrderRepository)
	pt := &tools.PurchaseOrderTools{Orders: repo}

	repo.On("List", mock.Anything).Return(poFixture(), nil)

	status := "pending_approval"
	_, out, err := pt.List(context.Background(), nil, tools.ListPurchaseOrdersInput{Status: &status})

	assert.NoError(t, err)
	assert.Equal(t, 1, out.Count)
	assert.Equal(t, "PO-3", out.Orders[0].ID)
}

func TestPurchaseOrderTools_List_InvalidStatus(t *testing.T) {
	pt := &tools.PurchaseOrderTools{}
	status := "bogus"
	result, _, err := pt.List(context.Background(), nil, tools.ListPurchaseOrdersInput{Status: &status})
	assert.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestPurchaseOrderTools_ListByItem(t *testing.T) {
	repo := new(mockPurchaseOrderRepository)
	pt := &tools.PurchaseOrderTools{Orders: repo}

	repo.On("List", mock.Anything).Return(poFixture(), nil)

	_, out, err := pt.ListByItem(context.Background(), nil, tools.ListPurchaseOrdersByItemInput{InventoryItemID: "INV-0002"})

	assert.NoError(t, err)
	assert.Equal(t, 2, out.Count)
}

func TestPurchaseOrderTools_ListByItem_EmptyID(t *testing.T) {
	pt := &tools.PurchaseOrderTools{}
	result, _, err := pt.ListByItem(context.Background(), nil, tools.ListPurchaseOrdersByItemInput{InventoryItemID: " "})
	assert.NoError(t, err)
	assert.True(t, result.IsError)
}
