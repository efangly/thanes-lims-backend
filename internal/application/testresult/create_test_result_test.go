package testresult_test

import (
	"context"
	"testing"

	applicationtestresult "github.com/efangly/thanes-lims-backend/internal/application/testresult"
	domainsample "github.com/efangly/thanes-lims-backend/internal/domain/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/efangly/thanes-lims-backend/internal/domain/testresult"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateTestResultUseCase_SampleNotFound(t *testing.T) {
	results := new(mockResultRepo)
	samples := new(mockSampleRepo)
	idgen := new(mockIDGen)

	samples.On("FindByID", mock.Anything, "SMP-2569-99999").Return(domainsample.Sample{}, shared.ErrNotFound)

	uc := applicationtestresult.NewCreateTestResultUseCase(results, samples, idgen)
	_, err := uc.Execute(context.Background(), applicationtestresult.CreateTestResultInput{
		SampleID: "SMP-2569-99999", TestName: "CBC", Analyst: "somchai",
	})

	assert.ErrorIs(t, err, shared.ErrNotFound)
}

func TestCreateTestResultUseCase_GeneratesID(t *testing.T) {
	results := new(mockResultRepo)
	samples := new(mockSampleRepo)
	idgen := new(mockIDGen)

	samples.On("FindByID", mock.Anything, "SMP-2569-00001").Return(domainsample.Sample{ID: "SMP-2569-00001"}, nil)
	idgen.On("Next", mock.Anything, "testresult", (*int)(nil)).Return(int64(88401), nil)
	results.On("Create", mock.Anything, mock.MatchedBy(func(t testresult.TestResult) bool {
		return t.ID == "TST-88401" && t.Status == testresult.StatusAnalyzing
	})).Return(testresult.TestResult{ID: "TST-88401", Status: testresult.StatusAnalyzing}, nil)

	uc := applicationtestresult.NewCreateTestResultUseCase(results, samples, idgen)
	created, err := uc.Execute(context.Background(), applicationtestresult.CreateTestResultInput{
		SampleID: "SMP-2569-00001", TestName: "CBC", Analyst: "somchai",
	})

	assert.NoError(t, err)
	assert.Equal(t, "TST-88401", created.ID)
}
