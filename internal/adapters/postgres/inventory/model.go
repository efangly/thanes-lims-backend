package inventory

import (
	"time"

	"gorm.io/gorm"
)

type Model struct {
	ID            string `gorm:"primaryKey"`
	Name          string
	Category      string
	Unit          string
	Min           int
	Max           int
	DefaultVendor string

	// Asset fields (Phase 7). custodian_user_id FKs users (NOT NULL);
	// manufacturer is a plain string; vendor_id FKs vendors; location_id
	// FKs a Location of Kind equipment_storage (checked in the use case,
	// not SQL - ADR 0007).
	CustodianUserID int64
	Manufacturer    string
	VendorID        *string
	LocationID      *string
	// DeletedAt makes Delete a soft delete (Retired, per ADR 0003); no
	// DeleteItem use case exists yet, but the column is in place.
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (Model) TableName() string { return "inventory_items" }

// LotModel is one InventoryLot row. The (item_id, lot_no) pair is unique -
// receiving more of an existing lot tops up its quantity rather than
// inserting a duplicate. Quantity carries no CHECK constraint: it may go
// negative on a forced over-issue (ADR 0008).
type LotModel struct {
	ID         string `gorm:"primaryKey"`
	ItemID     string
	LotNo      string
	ExpireDate *time.Time
	Quantity   int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (LotModel) TableName() string { return "inventory_lots" }
