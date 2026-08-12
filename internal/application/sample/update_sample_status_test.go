package sample_test

import (
	"context"
	"testing"

	applicationsample "github.com/efangly/thanes-lims-backend/internal/application/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateSampleStatusUseCase_RBACDenial(t *testing.T) {
	samples := new(mockSampleRepo)
	coc := new(mockCoCRepo)

	uc := applicationsample.NewUpdateSampleStatusUseCase(samples, coc)
	_, err := uc.Execute(context.Background(), applicationsample.UpdateSampleStatusInput{
		SampleID: "SMP-2569-00001", NewStatus: sample.StatusTesting, ActorRole: domainuser.RoleGeneral, ActorName: "somchai",
	})

	assert.ErrorIs(t, err, shared.ErrForbidden)
}

func TestUpdateSampleStatusUseCase_ValidTransition(t *testing.T) {
	samples := new(mockSampleRepo)
	coc := new(mockCoCRepo)

	existing := sample.Sample{ID: "SMP-2569-00001", Status: sample.StatusPending}
	samples.On("FindByID", mock.Anything, "SMP-2569-00001").Return(existing, nil)
	samples.On("UpdateStatus", mock.Anything, mock.MatchedBy(func(s sample.Sample) bool {
		return s.Status == sample.StatusTesting
	})).Return(sample.Sample{ID: "SMP-2569-00001", Status: sample.StatusTesting}, nil)
	coc.On("AppendStep", mock.Anything, mock.AnythingOfType("sample.CoCStep")).Return(sample.CoCStep{}, nil)

	uc := applicationsample.NewUpdateSampleStatusUseCase(samples, coc)
	updated, err := uc.Execute(context.Background(), applicationsample.UpdateSampleStatusInput{
		SampleID: "SMP-2569-00001", NewStatus: sample.StatusTesting, ActorRole: domainuser.RoleScientist, ActorName: "somchai",
	})

	assert.NoError(t, err)
	assert.Equal(t, sample.StatusTesting, updated.Status)
}

func TestUpdateSampleStatusUseCase_InvalidTransition(t *testing.T) {
	samples := new(mockSampleRepo)
	coc := new(mockCoCRepo)

	existing := sample.Sample{ID: "SMP-2569-00001", Status: sample.StatusCompleted}
	samples.On("FindByID", mock.Anything, "SMP-2569-00001").Return(existing, nil)

	uc := applicationsample.NewUpdateSampleStatusUseCase(samples, coc)
	_, err := uc.Execute(context.Background(), applicationsample.UpdateSampleStatusInput{
		SampleID: "SMP-2569-00001", NewStatus: sample.StatusPending, ActorRole: domainuser.RoleAdmin, ActorName: "somchai",
	})

	assert.ErrorIs(t, err, shared.ErrValidation)
}
