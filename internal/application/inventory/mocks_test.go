package inventory_test

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/inventory"
	"github.com/efangly/thanes-lims-backend/internal/domain/location"
	"github.com/efangly/thanes-lims-backend/internal/domain/notification"
	"github.com/efangly/thanes-lims-backend/internal/domain/user"
	"github.com/efangly/thanes-lims-backend/internal/domain/vendor"
	"github.com/stretchr/testify/mock"
)

type mockItemRepo struct{ mock.Mock }

func (m *mockItemRepo) Create(ctx context.Context, i inventory.InventoryItem) (inventory.InventoryItem, error) {
	args := m.Called(ctx, i)
	return args.Get(0).(inventory.InventoryItem), args.Error(1)
}
func (m *mockItemRepo) FindByID(ctx context.Context, id string) (inventory.InventoryItem, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(inventory.InventoryItem), args.Error(1)
}
func (m *mockItemRepo) List(ctx context.Context) ([]inventory.InventoryItem, error) {
	args := m.Called(ctx)
	return args.Get(0).([]inventory.InventoryItem), args.Error(1)
}
func (m *mockItemRepo) Update(ctx context.Context, i inventory.InventoryItem) (inventory.InventoryItem, error) {
	args := m.Called(ctx, i)
	return args.Get(0).(inventory.InventoryItem), args.Error(1)
}
func (m *mockItemRepo) UpdateDefaultVendor(ctx context.Context, id string, vendor string) (inventory.InventoryItem, error) {
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

type mockCustodianDir struct{ mock.Mock }

func (m *mockCustodianDir) FindByID(ctx context.Context, id int64) (user.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(user.User), args.Error(1)
}

type mockVendorDir struct{ mock.Mock }

func (m *mockVendorDir) FindByID(ctx context.Context, id string) (vendor.Vendor, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(vendor.Vendor), args.Error(1)
}

type mockLocationDir struct{ mock.Mock }

func (m *mockLocationDir) GetByID(ctx context.Context, id string) (location.Location, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(location.Location), args.Error(1)
}

type stubIDGen struct{}

func (stubIDGen) Next(_ context.Context, _ string, _ *int) (int64, error) { return 1, nil }

type mockNotifier struct{ mock.Mock }

func (m *mockNotifier) Notify(ctx context.Context, n notification.Notification) error {
	args := m.Called(ctx, n)
	return args.Error(0)
}
