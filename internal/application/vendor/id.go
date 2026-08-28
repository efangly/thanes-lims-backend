package vendor

import (
	"context"
	"fmt"

	portidgen "github.com/efangly/thanes-lims-backend/internal/ports/idgen"
)

// nextVendorID formats a human-readable Vendor ID, e.g. "VEN-00001". The
// "vendor" scope is not Buddhist-year-scoped (master data, like Location).
func nextVendorID(ctx context.Context, idgen portidgen.SequenceGenerator) (string, error) {
	seq, err := idgen.Next(ctx, "vendor", nil)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("VEN-%05d", seq), nil
}
