package sample_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	applicationsample "github.com/efangly/thanes-lims-backend/internal/application/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/efangly/thanes-lims-backend/internal/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateSampleUseCase_GeneratesIDAndInitialCoCStep(t *testing.T) {
	samples := new(mockSampleRepo)
	coc := new(mockCoCRepo)
	idgen := new(mockIDGen)
	custodians := new(mockCustodianDir)

	year := shared.BuddhistYear(time.Now())
	wantID := fmt.Sprintf("SMP-%d-%05d", year, 42)

	custodians.On("FindByID", mock.Anything, int64(7)).Return(user.User{ID: 7, Name: "somchai"}, nil)
	idgen.On("Next", mock.Anything, "sample", &year).Return(int64(42), nil)
	samples.On("Create", mock.Anything, mock.MatchedBy(func(s sample.Sample) bool {
		return s.ID == wantID && s.Status == sample.StatusPending
	})).Return(sample.Sample{ID: wantID, Status: sample.StatusPending}, nil)
	coc.On("AppendStep", mock.Anything, mock.MatchedBy(func(step sample.CoCStep) bool {
		return step.SampleID == wantID && step.Icon == sample.IconPlus
	})).Return(sample.CoCStep{}, nil)

	uc := applicationsample.NewCreateSampleUseCase(samples, coc, custodians, idgen)
	created, err := uc.Execute(context.Background(), applicationsample.CreateSampleInput{
		Name: "Blood sample", Type: sample.TypeBlood, CustodianUserID: 7,
	})

	assert.NoError(t, err)
	assert.Equal(t, wantID, created.ID)
}

func TestCreateSampleUseCase_RejectsDuplicateBarcodeID(t *testing.T) {
	samples := new(mockSampleRepo)
	coc := new(mockCoCRepo)
	idgen := new(mockIDGen)
	custodians := new(mockCustodianDir)

	custodians.On("FindByID", mock.Anything, int64(7)).Return(user.User{ID: 7, Name: "somchai"}, nil)
	samples.On("FindByBarcodeID", mock.Anything, "BC-DUP").Return(sample.Sample{ID: "SMP-OTHER"}, nil)

	uc := applicationsample.NewCreateSampleUseCase(samples, coc, custodians, idgen)
	code := "BC-DUP"
	_, err := uc.Execute(context.Background(), applicationsample.CreateSampleInput{
		Name: "Blood", Type: sample.TypeBlood, CustodianUserID: 7, BarcodeID: &code,
	})

	assert.ErrorIs(t, err, shared.ErrConflict)
	samples.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestCreateSampleUseCase_InvalidType(t *testing.T) {
	samples := new(mockSampleRepo)
	coc := new(mockCoCRepo)
	idgen := new(mockIDGen)

	uc := applicationsample.NewCreateSampleUseCase(samples, coc, new(mockCustodianDir), idgen)
	_, err := uc.Execute(context.Background(), applicationsample.CreateSampleInput{
		Name: "Bad", Type: sample.Type("bogus"), CustodianUserID: 1,
	})

	assert.ErrorIs(t, err, shared.ErrValidation)
}
