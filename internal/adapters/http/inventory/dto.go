package inventory

import (
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/inventory"
)

type CreateItemRequest struct {
	Name          string `json:"name" validate:"required"`
	Category      string `json:"category" validate:"required"`
	Unit          string `json:"unit" validate:"required"`
	Min           int    `json:"min"`
	Max           int    `json:"max"`
	DefaultVendor string `json:"default_vendor"`

	CustodianUserID int64  `json:"custodian_user_id" validate:"required"`
	Manufacturer    string `json:"manufacturer"`
	VendorID        string `json:"vendor_id"`
	LocationID      string `json:"location_id"`
}

// UpdateItemRequest is a partial update: a nil field is left untouched.
// Quantity is derived from received lots (POST /inventory/{id}/receive),
// DefaultVendor moves only through PATCH /inventory/{id}/default-vendor. For
// VendorID / LocationID a non-nil empty string clears the link.
type UpdateItemRequest struct {
	Name            *string `json:"name"`
	Category        *string `json:"category"`
	Unit            *string `json:"unit"`
	Min             *int    `json:"min"`
	Max             *int    `json:"max"`
	CustodianUserID *int64  `json:"custodian_user_id"`
	Manufacturer    *string `json:"manufacturer"`
	VendorID        *string `json:"vendor_id"`
	LocationID      *string `json:"location_id"`
}

// ReceiveStockRequest records goods arriving into the store against a lot.
// A lot number matching an existing lot for the item tops that lot up;
// otherwise a new lot is created. ExpireDate is optional (RFC3339).
type ReceiveStockRequest struct {
	LotNo      string     `json:"lot_no" validate:"required"`
	ExpireDate *time.Time `json:"expire_date"`
	Quantity   int        `json:"quantity" validate:"gt=0"`
}

type LotResponse struct {
	ID         string     `json:"id"`
	ItemID     string     `json:"item_id"`
	LotNo      string     `json:"lot_no"`
	ExpireDate *time.Time `json:"expire_date"`
	Quantity   int        `json:"quantity"`
}

func toLotResponse(l inventory.InventoryLot) LotResponse {
	return LotResponse{
		ID:         l.ID,
		ItemID:     l.ItemID,
		LotNo:      l.LotNo,
		ExpireDate: l.ExpireDate,
		Quantity:   l.Quantity,
	}
}

type ReceiveStockResponse struct {
	Item ItemResponse `json:"item"`
	Lot  LotResponse  `json:"lot"`
}

// IssueStockRequest withdraws stock (เบิกออก) against one or more lots the
// user picks explicitly. Set force=true to withdraw past a lot's recorded
// balance (its quantity may then go negative - ADR 0008).
type IssueStockRequest struct {
	Lines []IssueLineRequest `json:"lines" validate:"required,min=1,dive"`
	Force bool               `json:"force"`
}

type IssueLineRequest struct {
	LotID    string `json:"lot_id" validate:"required"`
	Quantity int    `json:"quantity" validate:"gt=0"`
}

// ShortfallResponse reports a line that asked for more than the lot holds.
type ShortfallResponse struct {
	LotID     string `json:"lot_id"`
	LotNo     string `json:"lot_no"`
	Requested int    `json:"requested"`
	Available int    `json:"available"`
}

// IssueStockResponse reports the outcome. When applied is false one or more
// lines over-issued and force was not set - no stock moved; shortfalls
// carries the real remaining balances so the caller can add another lot
// line or re-submit with force=true.
type IssueStockResponse struct {
	Applied    bool                `json:"applied"`
	Shortfalls []ShortfallResponse `json:"shortfalls,omitempty"`
	Item       ItemResponse        `json:"item"`
	Lots       []LotResponse       `json:"lots"`
}

type UpdateDefaultVendorRequest struct {
	Vendor string `json:"vendor" validate:"required"`
}

type ItemResponse struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Category        string  `json:"category"`
	Quantity        int     `json:"quantity"`
	Unit            string  `json:"unit"`
	Min             int     `json:"min"`
	Max             int     `json:"max"`
	Pct             int     `json:"pct"`
	Status          string  `json:"status"`
	DefaultVendor   string  `json:"default_vendor"`
	CustodianUserID int64   `json:"custodian_user_id"`
	Manufacturer    string  `json:"manufacturer"`
	VendorID        *string `json:"vendor_id"`
	LocationID      *string `json:"location_id"`
}

func toResponse(i inventory.InventoryItem) ItemResponse {
	return ItemResponse{
		ID:              i.ID,
		Name:            i.Name,
		Category:        i.Category,
		Quantity:        i.Quantity,
		Unit:            i.Unit,
		Min:             i.Min,
		Max:             i.Max,
		Pct:             i.Pct(),
		Status:          string(i.DerivedStatus()),
		DefaultVendor:   i.DefaultVendor,
		CustodianUserID: i.CustodianUserID,
		Manufacturer:    i.Manufacturer,
		VendorID:        i.VendorID,
		LocationID:      i.LocationID,
	}
}
