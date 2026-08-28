package equipment

import (
	"context"
	"errors"
	"fmt"
	"strings"

	domainlocation "github.com/efangly/thanes-lims-backend/internal/domain/location"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portequipment "github.com/efangly/thanes-lims-backend/internal/ports/equipment"
)

// optionalRef trims s and returns nil for an empty string, otherwise a
// pointer to the trimmed value - the shape both VendorID and LocationID
// take on Equipment.
func optionalRef(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}

// validateVendor returns nil when id is nil/empty or names an existing
// Vendor; shared.ErrValidation otherwise.
func validateVendor(ctx context.Context, vendors portequipment.VendorDirectory, id *string) error {
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
func validateLocation(ctx context.Context, locations portequipment.LocationDirectory, id *string) error {
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
