package testresult

import (
	"context"
	"fmt"

	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/efangly/thanes-lims-backend/internal/domain/testresult"
	porttestresult "github.com/efangly/thanes-lims-backend/internal/ports/testresult"
)

type SubmitResultUseCase struct {
	results porttestresult.Repository
}

func NewSubmitResultUseCase(results porttestresult.Repository) *SubmitResultUseCase {
	return &SubmitResultUseCase{results: results}
}

type SubmitResultInput struct {
	ID     string
	Result string
	Flag   testresult.Flag
}

func (uc *SubmitResultUseCase) Execute(ctx context.Context, in SubmitResultInput) (testresult.TestResult, error) {
	t, err := uc.results.FindByID(ctx, in.ID)
	if err != nil {
		return testresult.TestResult{}, err
	}

	if t.Status != testresult.StatusAnalyzing {
		return testresult.TestResult{}, fmt.Errorf("%w: cannot submit result from status %s", shared.ErrValidation, t.Status)
	}

	t.Result = in.Result
	t.Flag = in.Flag
	t.Status = testresult.StatusPendingVerification

	return uc.results.Update(ctx, t)
}
