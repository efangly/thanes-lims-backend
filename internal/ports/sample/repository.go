package sample

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/sample"
)

type ListFilter struct {
	Status *sample.Status
	Type   *sample.Type
	// BarcodeID, when set, matches a Sample's scan Barcode ID exactly
	// (the scan-to-filter path in the registry).
	BarcodeID *string
	// CustodianUserID, when set, matches the Sample's Custodian.
	CustodianUserID *int64
	// LocationText, when set, case-insensitively matches a substring of the
	// Sample's assigned Location name (join through LocationID).
	LocationText *string
	// LocationID, when set, matches the Sample's assigned Location exactly -
	// the "what is in this Box" query for the Cell grid (docs/adr/0009).
	LocationID *string
}

// PositionAssignment is one leg of a Move-within-box batch: put SampleID at
// Cell Position. The whole batch is applied in one transaction.
type PositionAssignment struct {
	SampleID string
	Position string
}

type SampleRepository interface {
	Create(ctx context.Context, s sample.Sample) (sample.Sample, error)
	FindByID(ctx context.Context, id string) (sample.Sample, error)
	// FindByBarcodeID resolves a scanned Barcode ID to its Sample; returns
	// shared.ErrNotFound if no Sample carries that code.
	FindByBarcodeID(ctx context.Context, barcodeID string) (sample.Sample, error)
	List(ctx context.Context, filter ListFilter) ([]sample.Sample, error)
	UpdateStatus(ctx context.Context, s sample.Sample) (sample.Sample, error)
	// UpdateLocation sets a Sample's put-away spot. position is non-nil only
	// when locationID is a Box (the Cell); it is cleared to NULL otherwise.
	UpdateLocation(ctx context.Context, sampleID string, locationID, position *string) (sample.Sample, error)
	// ListActiveByLocation returns the Samples still occupying locationID
	// (status pending/testing/completed) - used to render a Box's Cell grid
	// and to validate a Move-within-box batch.
	ListActiveByLocation(ctx context.Context, locationID string) ([]sample.Sample, error)
	// MoveWithinBox applies a batch of Cell reassignments within boxID in one
	// transaction: all land or none do. A resulting (location_id, position)
	// clash returns shared.ErrConflict.
	MoveWithinBox(ctx context.Context, boxID string, moves []PositionAssignment) ([]sample.Sample, error)
	// UpdateBarcodeID sets (or clears) a Sample's scan Barcode ID.
	UpdateBarcodeID(ctx context.Context, sampleID string, barcodeID *string) (sample.Sample, error)
	// ExistsActiveByLocation reports whether any Sample with status
	// pending/testing/completed currently occupies locationID - the leaf
	// Location capacity check for AssignLocationUseCase.
	ExistsActiveByLocation(ctx context.Context, locationID string) (bool, error)
	// ExistsActiveByLocationPosition is the Box-Cell capacity check: whether
	// an active Sample already occupies (locationID, position).
	ExistsActiveByLocationPosition(ctx context.Context, locationID, position string) (bool, error)
	// ExistsByLocation reports whether any Sample, regardless of status,
	// references locationID - the restrict-delete check for
	// DeleteLocationUseCase.
	ExistsByLocation(ctx context.Context, locationID string) (bool, error)
}
