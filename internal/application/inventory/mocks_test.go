package inventory_test

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/inventory"
	"github.com/efangly/thanes-lims-backend/internal/domain/notification"
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
func (m *mockItemRepo) UpdateQuantity(ctx context.Context, id string, quantity int) (inventory.InventoryItem, error) {
	args := m.Called(ctx, id, quantity)
	return args.Get(0).(inventory.InventoryItem), args.Error(1)
}

type mockNotifier struct{ mock.Mock }

func (m *mockNotifier) Notify(ctx context.Context, n notification.Notification) error {
	args := m.Called(ctx, n)
	return args.Error(0)
}
