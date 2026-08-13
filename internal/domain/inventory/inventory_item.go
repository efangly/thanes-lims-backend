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
