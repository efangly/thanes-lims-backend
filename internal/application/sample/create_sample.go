package sample

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portidgen "github.com/efangly/thanes-lims-backend/internal/ports/idgen"
	portsample "github.com/efangly/thanes-lims-backend/internal/ports/sample"
)

type CreateSampleUseCase struct {
	samples    portsample.SampleRepository
	coc        portsample.CoCRepository
	custodians portsample.CustodianDirectory
	idgen      portidgen.SequenceGenerator
}

func NewCreateSampleUseCase(samples portsample.SampleRepository, coc portsample.CoCRepository, custodians portsample.CustodianDirectory, idgen portidgen.SequenceGenerator) *CreateSampleUseCase {
	return &CreateSampleUseCase{samples: samples, coc: coc, custodians: custodians, idgen: idgen}
}

type CreateSampleInput struct {
	Name            string
	Type            sample.Type
	CustodianUserID int64
	LocationID      *string
	// BarcodeID is an optional user-supplied scan code. Leave nil to create
	// the Sample without one (it can be generated later via
	// GenerateBarcodeUseCase).
	BarcodeID   *string
	Description string
}

// Execute generates the human-readable SMP-{BuddhistYear}-{seq5} id and
// creates the initial "received" CoC step alongside the sample row.
func (uc *CreateSampleUseCase) Execute(ctx context.Context, in CreateSampleInput) (sample.Sample, error) {
	if !in.Type.Valid() {
		return sample.Sample{}, shared.ErrValidation
	}

	// Custodian must reference a real User (CONTEXT.md "Custodian").
	custodian, err := uc.custodians.FindByID(ctx, in.CustodianUserID)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return sample.Sample{}, fmt.Errorf("%w: custodian user %d not found", shared.ErrValidation, in.CustodianUserID)
		}
		return sample.Sample{}, err
	}

	// Optional user-supplied Barcode ID must be unique across Samples.
	var barcodeID *string
	if in.BarcodeID != nil {
		if code := strings.TrimSpace(*in.BarcodeID); code != "" {
			if _, err := uc.samples.FindByBarcodeID(ctx, code); err == nil {
				return sample.Sample{}, fmt.Errorf("%w: barcode id %q already in use", shared.ErrConflict, code)
			} else if !errors.Is(err, shared.ErrNotFound) {
				return sample.Sample{}, err
			}
			barcodeID = &code
		}
	}

	now := time.Now()
	year := shared.BuddhistYear(now)
	seq, err := uc.idgen.Next(ctx, "sample", &year)
	if err != nil {
		return sample.Sample{}, err
	}

	// LocationID is optional at intake (registration happens before
	// put-away); leaf-only + one-active-sample-per-location validation is
	// enforced by the dedicated assign-location use case, not here.
	s := sample.Sample{
		ID:              fmt.Sprintf("SMP-%d-%05d", year, seq),
		Name:            in.Name,
		Type:            in.Type,
		CustodianUserID: in.CustodianUserID,
		LocationID:      in.LocationID,
		Status:          sample.StatusPending,
		ReceivedAt:      now,
		BarcodeID:       barcodeID,
		Description:     strings.TrimSpace(in.Description),
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
		Who:        custodian.Name,
		OccurredAt: now,
	})
	if err != nil {
		return sample.Sample{}, err
	}

	return created, nil
}
