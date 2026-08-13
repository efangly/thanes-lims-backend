package inventory_test

import (
	"context"
	"testing"

	applicationinventory "github.com/efangly/thanes-lims-backend/internal/application/inventory"
	"github.com/efangly/thanes-lims-backend/internal/domain/inventory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateQuantityUseCase_NotifiesWhenBelowMin(t *testing.T) {
	items := new(mockItemRepo)
	notifier := new(mockNotifier)

	updated := inventory.InventoryItem{ID: "INV-001", Name: "เอทานอล 95%", Quantity: 2, Unit: "ขวด", Min: 5, Max: 20}
	items.On("UpdateQuantity", mock.Anything, "INV-001", 2).Return(updated, nil)
	notifier.On("Notify", mock.Anything, mock.Anything).Return(nil)

	uc := applicationinventory.NewUpdateQuantityUseCase(items, notifier)
	result, err := uc.Execute(context.Background(), "INV-001", 2)

	assert.NoError(t, err)
	assert.Equal(t, updated, result)
	notifier.AssertCalled(t, "Notify", mock.Anything, mock.Anything)
}

func TestUpdateQuantityUseCase_NoNotifyWhenAboveMin(t *testing.T) {
	items := new(mockItemRepo)
	notifier := new(mockNotifier)

	updated := inventory.InventoryItem{ID: "INV-001", Name: "เอทานอล 95%", Quantity: 15, Unit: "ขวด", Min: 5, Max: 20}
	items.On("UpdateQuantity", mock.Anything, "INV-001", 15).Return(updated, nil)

	uc := applicationinventory.NewUpdateQuantityUseCase(items, notifier)
	result, err := uc.Execute(context.Background(), "INV-001", 15)

	assert.NoError(t, err)
	assert.Equal(t, updated, result)
	notifier.AssertNotCalled(t, "Notify", mock.Anything, mock.Anything)
}
