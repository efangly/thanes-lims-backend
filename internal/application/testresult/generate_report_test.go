package testresult_test

import (
	"context"
	"testing"

	applicationtestresult "github.com/efangly/thanes-lims-backend/internal/application/testresult"
	domainsample "github.com/efangly/thanes-lims-backend/internal/domain/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/testresult"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGenerateReportUseCase_AggregatesResultSampleAndCoC(t *testing.T) {
	results := new(mockResultRepo)
	samples := new(mockSampleRepo)
	coc := new(mockCoCRepo)

	result := testresult.TestResult{ID: "TST-88401", SampleID: "SMP-2569-00001", TestName: "CBC"}
	sample := domainsample.Sample{ID: "SMP-2569-00001", Name: "Blood Sample"}
	steps := []domainsample.CoCStep{{ID: 1, SampleID: "SMP-2569-00001", Title: "รับตัวอย่างเข้าระบบ"}}

	results.On("FindByID", mock.Anything, "TST-88401").Return(result, nil)
	samples.On("FindByID", mock.Anything, "SMP-2569-00001").Return(sample, nil)
	coc.On("ListBySample", mock.Anything, "SMP-2569-00001").Return(steps, nil)

	uc := applicationtestresult.NewGenerateReportUseCase(results, samples, coc)
	data, err := uc.Execute(context.Background(), "TST-88401")

	assert.NoError(t, err)
	assert.Equal(t, result, data.Result)
	assert.Equal(t, sample, data.Sample)
	assert.Equal(t, steps, data.CoCSteps)
}

func TestGenerateReportUseCase_PropagatesNotFound(t *testing.T) {
	results := new(mockResultRepo)
	samples := new(mockSampleRepo)
	coc := new(mockCoCRepo)

	results.On("FindByID", mock.Anything, "TST-missing").Return(testresult.TestResult{}, assert.AnError)

	uc := applicationtestresult.NewGenerateReportUseCase(results, samples, coc)
	_, err := uc.Execute(context.Background(), "TST-missing")

	assert.ErrorIs(t, err, assert.AnError)
	samples.AssertNotCalled(t, "FindByID", mock.Anything, mock.Anything)
}
