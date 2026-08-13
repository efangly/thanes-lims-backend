package inventory

import (
	"context"
	"fmt"

	"github.com/efangly/thanes-lims-backend/internal/domain/inventory"
	"github.com/efangly/thanes-lims-backend/internal/domain/notification"
	portinventory "github.com/efangly/thanes-lims-backend/internal/ports/inventory"
	portnotification "github.com/efangly/thanes-lims-backend/internal/ports/notification"
)

type UpdateQuantityUseCase struct {
	items    portinventory.Repository
	notifier portnotification.Notifier
}

func NewUpdateQuantityUseCase(items portinventory.Repository, notifier portnotification.Notifier) *UpdateQuantityUseCase {
	return &UpdateQuantityUseCase{items: items, notifier: notifier}
}

func (uc *UpdateQuantityUseCase) Execute(ctx context.Context, id string, quantity int) (inventory.InventoryItem, error) {
	updated, err := uc.items.UpdateQuantity(ctx, id, quantity)
	if err != nil {
		return inventory.InventoryItem{}, err
	}

	if updated.BelowMin() {
		if err := uc.notifier.Notify(ctx, notification.Notification{
			Tone:    notification.ToneAmber,
			Title:   fmt.Sprintf("สินค้าใกล้หมด: %s", updated.Name),
			Message: fmt.Sprintf("คงเหลือ %d %s ต่ำกว่าจุดสั่งซื้อขั้นต่ำ (%d)", updated.Quantity, updated.Unit, updated.Min),
		}); err != nil {
			return inventory.InventoryItem{}, err
		}
	}

	return updated, nil
}
