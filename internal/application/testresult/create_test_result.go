package testresult

import (
	"context"
	"fmt"

	"github.com/efangly/thanes-lims-backend/internal/domain/testresult"
	portidgen "github.com/efangly/thanes-lims-backend/internal/ports/idgen"
	portsample "github.com/efangly/thanes-lims-backend/internal/ports/sample"
	porttestresult "github.com/efangly/thanes-lims-backend/internal/ports/testresult"
)

type CreateTestResultUseCase struct {
	results porttestresult.Repository
	samples portsample.SampleRepository
	idgen   portidgen.SequenceGenerator
}

func NewCreateTestResultUseCase(results porttestresult.Repository, samples portsample.SampleRepository, idgen portidgen.SequenceGenerator) *CreateTestResultUseCase {
	return &CreateTestResultUseCase{results: results, samples: samples, idgen: idgen}
}

type CreateTestResultInput struct {
	SampleID string
	TestName string
	Analyst  string
	RefRange string
}

// Execute validates the referenced Sample exists (FK integrity across
// aggregates enforced at the application layer, not a DB constraint, since
// samples and test_results are separate bounded contexts) before generating
// the TST-{seq5} id and creating the row with status "analyzing".
func (uc *CreateTestResultUseCase) Execute(ctx context.Context, in CreateTestResultInput) (testresult.TestResult, error) {
	if _, err := uc.samples.FindByID(ctx, in.SampleID); err != nil {
		return testresult.TestResult{}, err
	}

	seq, err := uc.idgen.Next(ctx, "testresult", nil)
	if err != nil {
		return testresult.TestResult{}, err
	}

	t := testresult.TestResult{
		ID:       fmt.Sprintf("TST-%05d", seq),
		SampleID: in.SampleID,
		TestName: in.TestName,
		Analyst:  in.Analyst,
		RefRange: in.RefRange,
		Status:   testresult.StatusAnalyzing,
	}

	return uc.results.Create(ctx, t)
}
