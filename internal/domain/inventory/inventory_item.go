package inventory

type Status string

const (
	StatusOK       Status = "ok"
	StatusLow      Status = "low"
	StatusCritical Status = "critical"
)

type InventoryItem struct {
	ID            string
	Name          string
	Category      string
	Quantity      int
	Unit          string
	Min           int
	Max           int
	DefaultVendor string

	// Asset fields (Phase 7). CustodianUserID is the User responsible for
	// the item (FK to users, required - see CONTEXT.md "Custodian").
	// Manufacturer is a plain descriptive string, distinct from VendorID
	// which FKs the Vendor master record (CONTEXT.md#vendors). LocationID
	// FKs a Location of Kind equipment_storage, shared with Equipment
	// (ADR 0007). VendorID and LocationID are optional.
	CustodianUserID int64
	Manufacturer    string
	VendorID        *string
	LocationID      *string
}

// Pct and DerivedStatus are computed on read, never stored, so they can
// never drift out of sync with Quantity/Min/Max.
func (i InventoryItem) Pct() int {
	if i.Max <= 0 {
		return 0
	}
	pct := i.Quantity * 100 / i.Max
	switch {
	case pct > 100:
		return 100
	case pct < 0:
		return 0
	default:
		return pct
	}
}

func (i InventoryItem) DerivedStatus() Status {
	switch {
	case i.Quantity <= i.Min/2:
		return StatusCritical
	case i.Quantity <= i.Min:
		return StatusLow
	default:
		return StatusOK
	}
}

func (i InventoryItem) BelowMin() bool {
	return i.Quantity <= i.Min
}
