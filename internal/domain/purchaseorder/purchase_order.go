package purchaseorder

import "time"

type Status string

const (
	StatusPendingApproval Status = "pending_approval"
	StatusSentToVendor    Status = "sent_to_vendor"
	StatusReceived        Status = "received"
	StatusCancelled       Status = "cancelled"
)

type PurchaseOrder struct {
	ID        string
	ItemID    string
	Quantity  int
	Vendor    string
	OrderDate time.Time
	Status    Status
}
