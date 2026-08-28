package sample

import (
	"context"
	"errors"

	"github.com/efangly/thanes-lims-backend/internal/domain/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portlocation "github.com/efangly/thanes-lims-backend/internal/ports/location"
	portsample "github.com/efangly/thanes-lims-backend/internal/ports/sample"
)

// StickerData is everything the label renderer needs for one Sample. The
// larger templates use every field; the tiny cap/stem templates use only
// ScanCode + Sample ID.
type StickerData struct {
	Sample           sample.Sample
	CustodianName    string
	LocationFullPath string
	// ScanCode is what the barcode/QR on the sticker encodes: the Sample's
	// Barcode ID when it has one, otherwise its ID (so a sticker is always
	// scannable to something).
	ScanCode string
}

// StickerDataUseCase aggregates the Sample, its Custodian's name and its
// Location Full Path for the label renderer.
type StickerDataUseCase struct {
	samples    portsample.SampleRepository
	custodians portsample.CustodianDirectory
	locations  portlocation.LocationRepository
}

func NewStickerDataUseCase(samples portsample.SampleRepository, custodians portsample.CustodianDirectory, locations portlocation.LocationRepository) *StickerDataUseCase {
	return &StickerDataUseCase{samples: samples, custodians: custodians, locations: locations}
}

func (uc *StickerDataUseCase) Execute(ctx context.Context, sampleID string) (StickerData, error) {
	s, err := uc.samples.FindByID(ctx, sampleID)
	if err != nil {
		return StickerData{}, err
	}

	out := StickerData{Sample: s, CustodianName: "-", LocationFullPath: "-", ScanCode: s.ID}
	if s.BarcodeID != nil && *s.BarcodeID != "" {
		out.ScanCode = *s.BarcodeID
	}

	if custodian, err := uc.custodians.FindByID(ctx, s.CustodianUserID); err == nil {
		out.CustodianName = custodian.Name
	} else if !errors.Is(err, shared.ErrNotFound) {
		return StickerData{}, err
	}

	if s.LocationID != nil {
		if fullPath, err := uc.locations.FullPath(ctx, *s.LocationID); err == nil {
			out.LocationFullPath = fullPath
		} else if !errors.Is(err, shared.ErrNotFound) {
			return StickerData{}, err
		}
	}

	return out, nil
}
