package vendor

import (
	"strings"

	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
)

// Vendor is master data for a supplier, referenced by FK (VendorID) from
// Equipment, Inventory Item, and Purchase Order rather than duplicated as
// free text - see CONTEXT.md#vendors. Manufacturer is a separate plain
// descriptive field on those entities and is deliberately NOT modelled
// here (a Manufacturer carries no contact details and is not master data).
type Vendor struct {
	ID           string
	Name         string
	ContactName  string
	ContactPhone string
	ContactEmail string
	Address      string
}

// Validate enforces the invariants that hold regardless of transport:
// Name is required, and ContactEmail - when given - looks like an address.
func (v Vendor) Validate() error {
	if strings.TrimSpace(v.Name) == "" {
		return shared.ErrValidation
	}
	if v.ContactEmail != "" && !strings.Contains(v.ContactEmail, "@") {
		return shared.ErrValidation
	}
	return nil
}
