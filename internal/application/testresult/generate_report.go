package testresult

import (
	"context"

	domainsample "github.com/efangly/thanes-lims-backend/internal/domain/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/testresult"
	portsample "github.com/efangly/thanes-lims-backend/internal/ports/sample"
	porttestresult "github.com/efangly/thanes-lims-backend/internal/ports/testresult"
)

type ReportData struct {
	Result   testresult.TestResult
	Sample   domainsample.Sample
	CoCSteps []domainsample.CoCStep
}

// GenerateReportUseCase aggregates everything the PDF report needs: the
// result itself, its linked sample, and the sample's chain-of-custody trail.
type GenerateReportUseCase struct {
	results porttestresult.Repository
	samples portsample.SampleRepository
	coc     portsample.CoCRepository
}

func NewGenerateReportUseCase(results porttestresult.Repository, samples portsample.SampleRepository, coc portsample.CoCRepository) *GenerateReportUseCase {
	return &GenerateReportUseCase{results: results, samples: samples, coc: coc}
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

	return ReportData{Result: result, Sample: s, CoCSteps: steps}, nil
}
