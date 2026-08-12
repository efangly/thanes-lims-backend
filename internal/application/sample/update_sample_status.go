package sample

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	portsample "github.com/efangly/thanes-lims-backend/internal/ports/sample"
)

type UpdateSampleStatusUseCase struct {
	samples portsample.SampleRepository
	coc     portsample.CoCRepository
}

func NewUpdateSampleStatusUseCase(samples portsample.SampleRepository, coc portsample.CoCRepository) *UpdateSampleStatusUseCase {
	return &UpdateSampleStatusUseCase{samples: samples, coc: coc}
}

type UpdateSampleStatusInput struct {
	SampleID  string
	NewStatus sample.Status
	ActorRole domainuser.Role
	ActorName string
}

// Execute enforces RBAC (edit permission) and the Sample state machine, then
// appends the state-machine-generated CoC step alongside the status update.
func (uc *UpdateSampleStatusUseCase) Execute(ctx context.Context, in UpdateSampleStatusInput) (sample.Sample, error) {
	if !in.ActorRole.Can(domainuser.PermEdit) {
		return sample.Sample{}, shared.ErrForbidden
	}

	s, err := uc.samples.FindByID(ctx, in.SampleID)
	if err != nil {
		return sample.Sample{}, err
	}

	step, err := s.Transition(in.NewStatus, in.ActorName)
	if err != nil {
		return sample.Sample{}, err
	}

	updated, err := uc.samples.UpdateStatus(ctx, s)
	if err != nil {
		return sample.Sample{}, err
	}

	if _, err := uc.coc.AppendStep(ctx, step); err != nil {
		return sample.Sample{}, err
	}

	return updated, nil
}
