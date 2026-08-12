package testresult

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/testresult"
)

type ListFilter struct {
	SampleID *string
	Status   *testresult.Status
}

type Repository interface {
	Create(ctx context.Context, t testresult.TestResult) (testresult.TestResult, error)
	FindByID(ctx context.Context, id string) (testresult.TestResult, error)
	List(ctx context.Context, filter ListFilter) ([]testresult.TestResult, error)
	Update(ctx context.Context, t testresult.TestResult) (testresult.TestResult, error)
}
