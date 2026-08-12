package sample

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/sample"
)

type ListFilter struct {
	Status *sample.Status
	Type   *sample.Type
}

type SampleRepository interface {
	Create(ctx context.Context, s sample.Sample) (sample.Sample, error)
	FindByID(ctx context.Context, id string) (sample.Sample, error)
	List(ctx context.Context, filter ListFilter) ([]sample.Sample, error)
	UpdateStatus(ctx context.Context, s sample.Sample) (sample.Sample, error)
}
