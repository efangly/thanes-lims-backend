package sample

import (
	"context"
	"fmt"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portidgen "github.com/efangly/thanes-lims-backend/internal/ports/idgen"
	portsample "github.com/efangly/thanes-lims-backend/internal/ports/sample"
)

type CreateSampleUseCase struct {
	samples portsample.SampleRepository
	coc     portsample.CoCRepository
	idgen   portidgen.SequenceGenerator
}

func NewCreateSampleUseCase(samples portsample.SampleRepository, coc portsample.CoCRepository, idgen portidgen.SequenceGenerator) *CreateSampleUseCase {
	return &CreateSampleUseCase{samples: samples, coc: coc, idgen: idgen}
}

type CreateSampleInput struct {
	Name      string
	Type      sample.Type
	Custodian string
	Location  string
}

// Execute generates the human-readable SMP-{BuddhistYear}-{seq5} id and
// creates the initial "received" CoC step alongside the sample row.
func (uc *CreateSampleUseCase) Execute(ctx context.Context, in CreateSampleInput) (sample.Sample, error) {
	if !in.Type.Valid() {
		return sample.Sample{}, shared.ErrValidation
	}

	now := time.Now()
	year := shared.BuddhistYear(now)
	seq, err := uc.idgen.Next(ctx, "sample", &year)
	if err != nil {
		return sample.Sample{}, err
	}

	s := sample.Sample{
		ID:         fmt.Sprintf("SMP-%d-%05d", year, seq),
		Name:       in.Name,
		Type:       in.Type,
		Custodian:  in.Custodian,
		Location:   in.Location,
		Status:     sample.StatusPending,
		ReceivedAt: now,
	}

	created, err := uc.samples.Create(ctx, s)
	if err != nil {
		return sample.Sample{}, err
	}

	_, err = uc.coc.AppendStep(ctx, sample.CoCStep{
		SampleID:   created.ID,
		State:      sample.CoCStateDone,
		Icon:       sample.IconPlus,
		Title:      "รับตัวอย่างเข้าระบบ",
		Who:        in.Custodian,
		OccurredAt: now,
	})
	if err != nil {
		return sample.Sample{}, err
	}

	return created, nil
}
