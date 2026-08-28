package equipment

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/location"
	"github.com/efangly/thanes-lims-backend/internal/domain/vendor"
)

// VendorDirectory resolves an Equipment's optional VendorID to a Vendor,
// used to validate the id on write. Returns shared.ErrNotFound when no such
// Vendor exists. *postgres/vendor.Repository satisfies this via FindByID.
type VendorDirectory interface {
	FindByID(ctx context.Context, id string) (vendor.Vendor, error)
}

// LocationDirectory resolves an Equipment's optional LocationID to a
// Location, used to validate the id and its Kind on write. Returns
// shared.ErrNotFound when no such Location exists.
// *postgres/location.Repository satisfies this via GetByID.
type LocationDirectory interface {
	GetByID(ctx context.Context, id string) (location.Location, error)
}
