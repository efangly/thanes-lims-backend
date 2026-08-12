package sample_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	applicationsample "github.com/efangly/thanes-lims-backend/internal/application/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateSampleUseCase_GeneratesIDAndInitialCoCStep(t *testing.T) {
	samples := new(mockSampleRepo)
	coc := new(mockCoCRepo)
	idgen := new(mockIDGen)

	year := shared.BuddhistYear(time.Now())
	wantID := fmt.Sprintf("SMP-%d-%05d", year, 42)

	idgen.On("Next", mock.Anything, "sample", &year).Return(int64(42), nil)
	samples.On("Create", mock.Anything, mock.MatchedBy(func(s sample.Sample) bool {
		return s.ID == wantID && s.Status == sample.StatusPending
	})).Return(sample.Sample{ID: wantID, Status: sample.StatusPending}, nil)
	coc.On("AppendStep", mock.Anything, mock.MatchedBy(func(step sample.CoCStep) bool {
		return step.SampleID == wantID && step.Icon == sample.IconPlus
	})).Return(sample.CoCStep{}, nil)

	uc := applicationsample.NewCreateSampleUseCase(samples, coc, idgen)
	created, err := uc.Execute(context.Background(), applicationsample.CreateSampleInput{
		Name: "Blood sample", Type: sample.TypeBlood, Custodian: "somchai", Location: "Fridge-A",
	})

	assert.NoError(t, err)
	assert.Equal(t, wantID, created.ID)
}

func TestCreateSampleUseCase_InvalidType(t *testing.T) {
	samples := new(mockSampleRepo)
	coc := new(mockCoCRepo)
	idgen := new(mockIDGen)

	uc := applicationsample.NewCreateSampleUseCase(samples, coc, idgen)
	_, err := uc.Execute(context.Background(), applicationsample.CreateSampleInput{
		Name: "Bad", Type: sample.Type("bogus"), Custodian: "x", Location: "y",
	})

	assert.ErrorIs(t, err, shared.ErrValidation)
}
