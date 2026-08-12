package sample

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/sample"
	portsample "github.com/efangly/thanes-lims-backend/internal/ports/sample"
)

type GetSampleUseCase struct {
	samples portsample.SampleRepository
}

func NewGetSampleUseCase(samples portsample.SampleRepository) *GetSampleUseCase {
	return &GetSampleUseCase{samples: samples}
}

func (uc *GetSampleUseCase) Execute(ctx context.Context, id string) (sample.Sample, error) {
	return uc.samples.FindByID(ctx, id)
}

type ListCoCStepsUseCase struct {
	coc portsample.CoCRepository
}

func NewListCoCStepsUseCase(coc portsample.CoCRepository) *ListCoCStepsUseCase {
	return &ListCoCStepsUseCase{coc: coc}
}

func (uc *ListCoCStepsUseCase) Execute(ctx context.Context, sampleID string) ([]sample.CoCStep, error) {
	return uc.coc.ListBySample(ctx, sampleID)
}
