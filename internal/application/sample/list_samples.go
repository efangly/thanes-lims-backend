package sample

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/sample"
	portsample "github.com/efangly/thanes-lims-backend/internal/ports/sample"
)

type ListSamplesUseCase struct {
	samples portsample.SampleRepository
}

func NewListSamplesUseCase(samples portsample.SampleRepository) *ListSamplesUseCase {
	return &ListSamplesUseCase{samples: samples}
}

func (uc *ListSamplesUseCase) Execute(ctx context.Context, filter portsample.ListFilter) ([]sample.Sample, error) {
	return uc.samples.List(ctx, filter)
}
