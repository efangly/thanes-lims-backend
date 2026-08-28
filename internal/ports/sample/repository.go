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
}

type SampleRepository interface {
	Create(ctx context.Context, s sample.Sample) (sample.Sample, error)
	FindByID(ctx context.Context, id string) (sample.Sample, error)
	// FindByBarcodeID resolves a scanned Barcode ID to its Sample; returns
	// shared.ErrNotFound if no Sample carries that code.
	FindByBarcodeID(ctx context.Context, barcodeID string) (sample.Sample, error)
	List(ctx context.Context, filter ListFilter) ([]sample.Sample, error)
	UpdateStatus(ctx context.Context, s sample.Sample) (sample.Sample, error)
	UpdateLocation(ctx context.Context, sampleID string, locationID *string) (sample.Sample, error)
	// UpdateBarcodeID sets (or clears) a Sample's scan Barcode ID.
	UpdateBarcodeID(ctx context.Context, sampleID string, barcodeID *string) (sample.Sample, error)
	// ExistsActiveByLocation reports whether any Sample with status
	// pending/testing/completed currently occupies locationID - the leaf
	// Location capacity check for AssignLocationUseCase.
	ExistsActiveByLocation(ctx context.Context, locationID string) (bool, error)
	// ExistsByLocation reports whether any Sample, regardless of status,
	// references locationID - the restrict-delete check for
	// DeleteLocationUseCase.
	ExistsByLocation(ctx context.Context, locationID string) (bool, error)
}
