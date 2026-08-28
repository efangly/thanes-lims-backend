package sample

import (
	"context"
	"fmt"

	"github.com/efangly/thanes-lims-backend/internal/domain/sample"
	portidgen "github.com/efangly/thanes-lims-backend/internal/ports/idgen"
	portsample "github.com/efangly/thanes-lims-backend/internal/ports/sample"
)

// GenerateBarcodeUseCase assigns an auto-generated scan Barcode ID
// (SMP-BC-{seq5}) to a Sample that doesn't already have one. If the Sample
// already carries a Barcode ID (user-supplied or previously generated) it is
// returned unchanged - generation is idempotent per Sample.
type GenerateBarcodeUseCase struct {
	samples portsample.SampleRepository
	idgen   portidgen.SequenceGenerator
}

func NewGenerateBarcodeUseCase(samples portsample.SampleRepository, idgen portidgen.SequenceGenerator) *GenerateBarcodeUseCase {
	return &GenerateBarcodeUseCase{samples: samples, idgen: idgen}
}

func (uc *GenerateBarcodeUseCase) Execute(ctx context.Context, sampleID string) (sample.Sample, error) {
	s, err := uc.samples.FindByID(ctx, sampleID)
	if err != nil {
		return sample.Sample{}, err
	}
	if s.BarcodeID != nil && *s.BarcodeID != "" {
		return s, nil
	}

	seq, err := uc.idgen.Next(ctx, "sample_barcode", nil)
	if err != nil {
		return sample.Sample{}, err
	}
	code := fmt.Sprintf("SMP-BC-%05d", seq)

	return uc.samples.UpdateBarcodeID(ctx, sampleID, &code)
}
