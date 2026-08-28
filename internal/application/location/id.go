package location

import (
	"context"
	"fmt"

	portidgen "github.com/efangly/thanes-lims-backend/internal/ports/idgen"
)

// nextLocationID formats a human-readable Location ID, e.g. "LOC-00001".
// The "location" scope is not Buddhist-year-scoped, unlike Sample/PurchaseOrder.
func nextLocationID(ctx context.Context, idgen portidgen.SequenceGenerator) (string, error) {
	seq, err := idgen.Next(ctx, "location", nil)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("LOC-%05d", seq), nil
}

// nextBarcodeCode formats a Location Barcode, e.g. "LOC-BC-00001",
// auto-generated for every Location node on creation (see
// CONTEXT.md "Location Barcode"). Its own id_sequences scope, independent
// of the "location" ID counter.
func nextBarcodeCode(ctx context.Context, idgen portidgen.SequenceGenerator) (string, error) {
	seq, err := idgen.Next(ctx, "location_barcode", nil)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("LOC-BC-%05d", seq), nil
}
