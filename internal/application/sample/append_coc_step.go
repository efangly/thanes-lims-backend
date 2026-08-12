package sample

import (
	"context"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/sample"
	portsample "github.com/efangly/thanes-lims-backend/internal/ports/sample"
)

// AppendCoCStepUseCase records a manual custody-transfer entry, distinct
// from the automatic steps a status Transition() generates.
type AppendCoCStepUseCase struct {
	coc portsample.CoCRepository
}

func NewAppendCoCStepUseCase(coc portsample.CoCRepository) *AppendCoCStepUseCase {
	return &AppendCoCStepUseCase{coc: coc}
}

type AppendCoCStepInput struct {
	SampleID string
	Title    string
	Meta     string
	Who      string
}

func (uc *AppendCoCStepUseCase) Execute(ctx context.Context, in AppendCoCStepInput) (sample.CoCStep, error) {
	return uc.coc.AppendStep(ctx, sample.CoCStep{
		SampleID:   in.SampleID,
		State:      sample.CoCStateDone,
		Icon:       sample.IconArrow,
		Title:      in.Title,
		Meta:       in.Meta,
		Who:        in.Who,
		OccurredAt: time.Now(),
	})
}
