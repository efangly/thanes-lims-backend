package purchaseorder

import (
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/purchaseorder"
)

// MarkReceivedRequest carries the lot the delivered goods are booked into
// (Phase 8 - stock enters inventory only through lots). ExpireDate is
// optional (RFC3339).
type MarkReceivedRequest struct {
	LotNo      string     `json:"lot_no" validate:"required"`
	ExpireDate *time.Time `json:"expire_date"`
}

type POResponse struct {
	ID        string    `json:"id"`
	ItemID    string    `json:"item_id"`
	Quantity  int       `json:"quantity"`
	Vendor    string    `json:"vendor"`
	OrderDate time.Time `json:"order_date"`
	Status    string    `json:"status"`
}

func toResponse(po purchaseorder.PurchaseOrder) POResponse {
	return POResponse{
		ID:        po.ID,
		ItemID:    po.ItemID,
		Quantity:  po.Quantity,
		Vendor:    po.Vendor,
		OrderDate: po.OrderDate,
		Status:    string(po.Status),
	}
}
