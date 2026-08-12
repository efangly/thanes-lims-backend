package testresult

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/testresult"
	porttestresult "github.com/efangly/thanes-lims-backend/internal/ports/testresult"
)

type ListTestResultsUseCase struct {
	results porttestresult.Repository
}

func NewListTestResultsUseCase(results porttestresult.Repository) *ListTestResultsUseCase {
	return &ListTestResultsUseCase{results: results}
}

func (uc *ListTestResultsUseCase) Execute(ctx context.Context, filter porttestresult.ListFilter) ([]testresult.TestResult, error) {
	return uc.results.List(ctx, filter)
}

type GetTestResultUseCase struct {
	results porttestresult.Repository
}

func NewGetTestResultUseCase(results porttestresult.Repository) *GetTestResultUseCase {
	return &GetTestResultUseCase{results: results}
}

func (uc *GetTestResultUseCase) Execute(ctx context.Context, id string) (testresult.TestResult, error) {
	return uc.results.FindByID(ctx, id)
}
