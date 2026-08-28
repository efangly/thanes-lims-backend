package testresult

import (
	"context"
	"errors"

	domainsample "github.com/efangly/thanes-lims-backend/internal/domain/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/efangly/thanes-lims-backend/internal/domain/testresult"
	portlocation "github.com/efangly/thanes-lims-backend/internal/ports/location"
	portsample "github.com/efangly/thanes-lims-backend/internal/ports/sample"
	porttestresult "github.com/efangly/thanes-lims-backend/internal/ports/testresult"
)

type ReportData struct {
	Result testresult.TestResult
	Sample domainsample.Sample
	// CustodianName is the display name of the User the Sample's
	// CustodianUserID points at, resolved for the report; "-" if the User
	// can no longer be found.
	CustodianName string
	CoCSteps      []domainsample.CoCStep
	// LocationFullPath is the sample's storage Location rendered as its
	// Full Path (e.g. "Fridge-A / Shelf-2 / Slot-4"), or "-" if the sample
	// has no LocationID assigned. See CONTEXT.md#storage-location.
	LocationFullPath string
}

// GenerateReportUseCase aggregates everything the PDF report needs: the
// result itself, its linked sample, the sample's chain-of-custody trail,
// and the sample's storage Location resolved to a human-readable Full Path.
type GenerateReportUseCase struct {
	results    porttestresult.Repository
	samples    portsample.SampleRepository
	coc        portsample.CoCRepository
	locations  portlocation.LocationRepository
	custodians portsample.CustodianDirectory
}

func NewGenerateReportUseCase(results porttestresult.Repository, samples portsample.SampleRepository, coc portsample.CoCRepository, locations portlocation.LocationRepository, custodians portsample.CustodianDirectory) *GenerateReportUseCase {
	return &GenerateReportUseCase{results: results, samples: samples, coc: coc, locations: locations, custodians: custodians}
}

func (uc *GenerateReportUseCase) Execute(ctx context.Context, id string) (ReportData, error) {
	result, err := uc.results.FindByID(ctx, id)
	if err != nil {
		return ReportData{}, err
	}

	s, err := uc.samples.FindByID(ctx, result.SampleID)
	if err != nil {
		return ReportData{}, err
	}

	steps, err := uc.coc.ListBySample(ctx, result.SampleID)
	if err != nil {
		return ReportData{}, err
	}

	custodianName := "-"
	if custodian, err := uc.custodians.FindByID(ctx, s.CustodianUserID); err == nil {
		custodianName = custodian.Name
	} else if !errors.Is(err, shared.ErrNotFound) {
		return ReportData{}, err
	}

	locationFullPath := "-"
	if s.LocationID != nil {
		fullPath, err := uc.locations.FullPath(ctx, *s.LocationID)
		if err != nil && !errors.Is(err, shared.ErrNotFound) {
			return ReportData{}, err
		}
		if err == nil {
			locationFullPath = fullPath
		}
	}

	return ReportData{Result: result, Sample: s, CustodianName: custodianName, CoCSteps: steps, LocationFullPath: locationFullPath}, nil
}
