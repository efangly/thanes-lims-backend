package vendor

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/vendor"
)

// Repository is the persistence port for Vendor master data. There is
// deliberately no Delete - removing a Vendor that Equipment/Inventory/PO
// rows point at is out of scope until a phase actually needs it (see
// task.md Phase 1).
type Repository interface {
	Create(ctx context.Context, v vendor.Vendor) (vendor.Vendor, error)
	FindByID(ctx context.Context, id string) (vendor.Vendor, error)
	FindByName(ctx context.Context, name string) (vendor.Vendor, error)
	List(ctx context.Context) ([]vendor.Vendor, error)
	Update(ctx context.Context, v vendor.Vendor) (vendor.Vendor, error)
}
