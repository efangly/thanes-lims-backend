package inventory

import (
	"context"
	"errors"
	"fmt"
	"strings"

	domainlocation "github.com/efangly/thanes-lims-backend/internal/domain/location"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portinventory "github.com/efangly/thanes-lims-backend/internal/ports/inventory"
)

// optionalRef trims s and returns nil for an empty string, otherwise a
// pointer to the trimmed value - the shape both VendorID and LocationID
// take on InventoryItem.
func optionalRef(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}

// validateCustodian returns nil when id names an existing User;
// shared.ErrValidation otherwise. CustodianUserID is required, so id 0 is
// itself invalid.
func validateCustodian(ctx context.Context, custodians portinventory.CustodianDirectory, id int64) error {
	if custodians == nil {
		return nil
	}
	if id == 0 {
		return fmt.Errorf("%w: custodian_user_id is required", shared.ErrValidation)
	}
	if _, err := custodians.FindByID(ctx, id); err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return fmt.Errorf("%w: custodian user %d not found", shared.ErrValidation, id)
		}
		return err
	}
	return nil
}

// validateVendor returns nil when id is nil/empty or names an existing
// Vendor; shared.ErrValidation otherwise.
func validateVendor(ctx context.Context, vendors portinventory.VendorDirectory, id *string) error {
	if id == nil || vendors == nil {
		return nil
	}
	if _, err := vendors.FindByID(ctx, *id); err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return fmt.Errorf("%w: vendor %q not found", shared.ErrValidation, *id)
		}
		return err
	}
	return nil
}

// validateLocation returns nil when id is nil/empty or names an existing
// Location of Kind equipment_storage (ADR 0007); shared.ErrValidation
// otherwise.
func validateLocation(ctx context.Context, locations portinventory.LocationDirectory, id *string) error {
	if id == nil || locations == nil {
		return nil
	}
	loc, err := locations.GetByID(ctx, *id)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return fmt.Errorf("%w: location %q not found", shared.ErrValidation, *id)
		}
		return err
	}
	if loc.Kind != "" && loc.Kind != domainlocation.KindEquipmentStorage {
		return fmt.Errorf("%w: location %q is not equipment storage", shared.ErrValidation, *id)
	}
	return nil
}
