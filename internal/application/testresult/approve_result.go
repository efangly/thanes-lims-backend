package testresult

import (
	"context"
	"fmt"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/notification"
	domainsample "github.com/efangly/thanes-lims-backend/internal/domain/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/efangly/thanes-lims-backend/internal/domain/testresult"
	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	portnotification "github.com/efangly/thanes-lims-backend/internal/ports/notification"
	portsample "github.com/efangly/thanes-lims-backend/internal/ports/sample"
	porttestresult "github.com/efangly/thanes-lims-backend/internal/ports/testresult"
)

type ApproveResultUseCase struct {
	results  porttestresult.Repository
	coc      portsample.CoCRepository
	notifier portnotification.Notifier
}

func NewApproveResultUseCase(results porttestresult.Repository, coc portsample.CoCRepository, notifier portnotification.Notifier) *ApproveResultUseCase {
	return &ApproveResultUseCase{results: results, coc: coc, notifier: notifier}
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
	// Approve permission: same Admin/QA set the removed Can(PermApprove)
	// matrix granted (see ADR 0002 - wiring this to the normalized RBAC
	// model is a later phase, out of scope here).
	if in.ActorRole != domainuser.RoleAdmin && in.ActorRole != domainuser.RoleQA {
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

	if err := uc.notifier.Notify(ctx, notification.Notification{
		Tone:    notification.ToneGreen,
		Title:   "ผลการทดสอบได้รับการอนุมัติ",
		Message: fmt.Sprintf("ผลการทดสอบ %s ได้รับการอนุมัติโดย %s", t.ID, in.ActorName),
	}); err != nil {
		return testresult.TestResult{}, err
	}

	return updated, nil
}
