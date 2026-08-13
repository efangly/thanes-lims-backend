package testresult_test

import (
	"context"
	"testing"

	applicationtestresult "github.com/efangly/thanes-lims-backend/internal/application/testresult"
	"github.com/efangly/thanes-lims-backend/internal/domain/notification"
	domainsample "github.com/efangly/thanes-lims-backend/internal/domain/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/efangly/thanes-lims-backend/internal/domain/testresult"
	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestApproveResultUseCase_RBACDenial(t *testing.T) {
	results := new(mockResultRepo)
	coc := new(mockCoCRepo)
	notifier := new(mockNotifier)

	uc := applicationtestresult.NewApproveResultUseCase(results, coc, notifier)
	_, err := uc.Execute(context.Background(), applicationtestresult.ApproveResultInput{
		ID: "TST-88401", ActorRole: domainuser.RoleScientist, ActorName: "somchai",
	})

	assert.ErrorIs(t, err, shared.ErrForbidden)
	notifier.AssertNotCalled(t, "Notify", mock.Anything, mock.Anything)
}

func TestApproveResultUseCase_AppendsCoCStepOnLinkedSample(t *testing.T) {
	results := new(mockResultRepo)
	coc := new(mockCoCRepo)
	notifier := new(mockNotifier)

	existing := testresult.TestResult{ID: "TST-88401", SampleID: "SMP-2569-00001", Status: testresult.StatusPendingVerification}
	results.On("FindByID", mock.Anything, "TST-88401").Return(existing, nil)
	results.On("Update", mock.Anything, mock.MatchedBy(func(t testresult.TestResult) bool {
		return t.Status == testresult.StatusApproved
	})).Return(testresult.TestResult{ID: "TST-88401", Status: testresult.StatusApproved}, nil)
	coc.On("AppendStep", mock.Anything, mock.MatchedBy(func(step domainsample.CoCStep) bool {
		return step.SampleID == "SMP-2569-00001" && step.Icon == domainsample.IconCheck
	})).Return(domainsample.CoCStep{}, nil)
	notifier.On("Notify", mock.Anything, mock.MatchedBy(func(n notification.Notification) bool {
		return n.Title != ""
	})).Return(nil)

	uc := applicationtestresult.NewApproveResultUseCase(results, coc, notifier)
	updated, err := uc.Execute(context.Background(), applicationtestresult.ApproveResultInput{
		ID: "TST-88401", ActorRole: domainuser.RoleQA, ActorName: "priya",
	})

	assert.NoError(t, err)
	assert.Equal(t, testresult.StatusApproved, updated.Status)
	coc.AssertCalled(t, "AppendStep", mock.Anything, mock.Anything)
	notifier.AssertCalled(t, "Notify", mock.Anything, mock.Anything)
}

func TestApproveResultUseCase_WrongStatus(t *testing.T) {
	results := new(mockResultRepo)
	coc := new(mockCoCRepo)
	notifier := new(mockNotifier)

	existing := testresult.TestResult{ID: "TST-88401", SampleID: "SMP-2569-00001", Status: testresult.StatusAnalyzing}
	results.On("FindByID", mock.Anything, "TST-88401").Return(existing, nil)

	uc := applicationtestresult.NewApproveResultUseCase(results, coc, notifier)
	_, err := uc.Execute(context.Background(), applicationtestresult.ApproveResultInput{
		ID: "TST-88401", ActorRole: domainuser.RoleAdmin, ActorName: "somchai",
	})

	assert.ErrorIs(t, err, shared.ErrValidation)
	notifier.AssertNotCalled(t, "Notify", mock.Anything, mock.Anything)
}
