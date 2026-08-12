package testresult

import (
	"context"
	"fmt"
	"time"

	domainsample "github.com/efangly/thanes-lims-backend/internal/domain/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/efangly/thanes-lims-backend/internal/domain/testresult"
	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	portsample "github.com/efangly/thanes-lims-backend/internal/ports/sample"
	porttestresult "github.com/efangly/thanes-lims-backend/internal/ports/testresult"
)

type ApproveResultUseCase struct {
	results porttestresult.Repository
	coc     portsample.CoCRepository
}

func NewApproveResultUseCase(results porttestresult.Repository, coc portsample.CoCRepository) *ApproveResultUseCase {
	return &ApproveResultUseCase{results: results, coc: coc}
}

type ApproveResultInput struct {
	ID        string
	ActorRole domainuser.Role
	ActorName string
}

// Execute is the cross-aggregate orchestration point: approving a
// TestResult also appends a CoC step to its linked Sample. That side effect
// lives here in the application layer (not inside either aggregate) since
// it spans two bounded contexts.
func (uc *ApproveResultUseCase) Execute(ctx context.Context, in ApproveResultInput) (testresult.TestResult, error) {
	if !in.ActorRole.Can(domainuser.PermApprove) {
		return testresult.TestResult{}, shared.ErrForbidden
	}

	t, err := uc.results.FindByID(ctx, in.ID)
	if err != nil {
		return testresult.TestResult{}, err
	}

	if t.Status != testresult.StatusPendingVerification {
		return testresult.TestResult{}, fmt.Errorf("%w: cannot approve result from status %s", shared.ErrValidation, t.Status)
	}

	t.Status = testresult.StatusApproved
	updated, err := uc.results.Update(ctx, t)
	if err != nil {
		return testresult.TestResult{}, err
	}

	_, err = uc.coc.AppendStep(ctx, domainsample.CoCStep{
		SampleID:   t.SampleID,
		State:      domainsample.CoCStateDone,
		Icon:       domainsample.IconCheck,
		Title:      "อนุมัติผลการทดสอบ",
		Who:        in.ActorName,
		OccurredAt: time.Now(),
	})
	if err != nil {
		return testresult.TestResult{}, err
	}

	return updated, nil
}
