package inventory

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/location"
	"github.com/efangly/thanes-lims-backend/internal/domain/user"
	"github.com/efangly/thanes-lims-backend/internal/domain/vendor"
)

// CustodianDirectory resolves an InventoryItem's CustodianUserID to a User,
// used to validate the id on write. Returns shared.ErrNotFound when no such
// User exists. *postgres/user.Repository satisfies this via FindByID.
type CustodianDirectory interface {
	FindByID(ctx context.Context, id int64) (user.User, error)
}

// VendorDirectory resolves an InventoryItem's optional VendorID to a Vendor,
// used to validate the id on write. Returns shared.ErrNotFound when no such
// Vendor exists. *postgres/vendor.Repository satisfies this via FindByID.
type VendorDirectory interface {
	FindByID(ctx context.Context, id string) (vendor.Vendor, error)
}

// LocationDirectory resolves an InventoryItem's optional LocationID to a
// Location, used to validate the id and its Kind (equipment_storage, ADR
// 0007) on write. Returns shared.ErrNotFound when no such Location exists.
// *postgres/location.Repository satisfies this via GetByID.
type LocationDirectory interface {
	GetByID(ctx context.Context, id string) (location.Location, error)
}
