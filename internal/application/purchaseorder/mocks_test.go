package purchaseorder_test

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/inventory"
	"github.com/efangly/thanes-lims-backend/internal/domain/purchaseorder"
	"github.com/stretchr/testify/mock"
)

type mockPORepo struct{ mock.Mock }

func (m *mockPORepo) Create(ctx context.Context, po purchaseorder.PurchaseOrder) (purchaseorder.PurchaseOrder, error) {
	args := m.Called(ctx, po)
	return args.Get(0).(purchaseorder.PurchaseOrder), args.Error(1)
}
func (m *mockPORepo) FindByID(ctx context.Context, id string) (purchaseorder.PurchaseOrder, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(purchaseorder.PurchaseOrder), args.Error(1)
}
func (m *mockPORepo) List(ctx context.Context) ([]purchaseorder.PurchaseOrder, error) {
	args := m.Called(ctx)
	return args.Get(0).([]purchaseorder.PurchaseOrder), args.Error(1)
}
func (m *mockPORepo) Update(ctx context.Context, po purchaseorder.PurchaseOrder) (purchaseorder.PurchaseOrder, error) {
	args := m.Called(ctx, po)
	return args.Get(0).(purchaseorder.PurchaseOrder), args.Error(1)
}

type mockInventoryRepo struct{ mock.Mock }

func (m *mockInventoryRepo) Create(ctx context.Context, i inventory.InventoryItem) (inventory.InventoryItem, error) {
	args := m.Called(ctx, i)
	return args.Get(0).(inventory.InventoryItem), args.Error(1)
}
func (m *mockInventoryRepo) FindByID(ctx context.Context, id string) (inventory.InventoryItem, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(inventory.InventoryItem), args.Error(1)
}
func (m *mockInventoryRepo) List(ctx context.Context) ([]inventory.InventoryItem, error) {
	args := m.Called(ctx)
	return args.Get(0).([]inventory.InventoryItem), args.Error(1)
}
func (m *mockInventoryRepo) Update(ctx context.Context, i inventory.InventoryItem) (inventory.InventoryItem, error) {
	args := m.Called(ctx, i)
	return args.Get(0).(inventory.InventoryItem), args.Error(1)
}
func (m *mockInventoryRepo) UpdateQuantity(ctx context.Context, id string, quantity int) (inventory.InventoryItem, error) {
	args := m.Called(ctx, id, quantity)
	return args.Get(0).(inventory.InventoryItem), args.Error(1)
}
func (m *mockInventoryRepo) UpdateDefaultVendor(ctx context.Context, id string, vendor string) (inventory.InventoryItem, error) {
	args := m.Called(ctx, id, vendor)
	return args.Get(0).(inventory.InventoryItem), args.Error(1)
}

type mockLotRepo struct{ mock.Mock }

func (m *mockLotRepo) Create(ctx context.Context, l inventory.InventoryLot) (inventory.InventoryLot, error) {
	args := m.Called(ctx, l)
	return args.Get(0).(inventory.InventoryLot), args.Error(1)
}
func (m *mockLotRepo) FindByID(ctx context.Context, id string) (inventory.InventoryLot, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(inventory.InventoryLot), args.Error(1)
}
func (m *mockLotRepo) FindByItemAndLotNo(ctx context.Context, itemID, lotNo string) (inventory.InventoryLot, error) {
	args := m.Called(ctx, itemID, lotNo)
	return args.Get(0).(inventory.InventoryLot), args.Error(1)
}
func (m *mockLotRepo) ListByItem(ctx context.Context, itemID string) ([]inventory.InventoryLot, error) {
	args := m.Called(ctx, itemID)
	return args.Get(0).([]inventory.InventoryLot), args.Error(1)
}
func (m *mockLotRepo) UpdateQuantity(ctx context.Context, id string, quantity int) (inventory.InventoryLot, error) {
	args := m.Called(ctx, id, quantity)
	return args.Get(0).(inventory.InventoryLot), args.Error(1)
}

type mockIDGen struct{ mock.Mock }

func (m *mockIDGen) Next(ctx context.Context, scope string, year *int) (int64, error) {
	args := m.Called(ctx, scope, year)
	return args.Get(0).(int64), args.Error(1)
}
