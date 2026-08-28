package testresult_test

import (
	"context"
	"testing"

	applicationtestresult "github.com/efangly/thanes-lims-backend/internal/application/testresult"
	domainsample "github.com/efangly/thanes-lims-backend/internal/domain/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/testresult"
	"github.com/efangly/thanes-lims-backend/internal/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGenerateReportUseCase_AggregatesResultSampleCoCAndLocation(t *testing.T) {
	results := new(mockResultRepo)
	samples := new(mockSampleRepo)
	coc := new(mockCoCRepo)
	locations := new(mockLocationRepo)

	locationID := "LOC-00004"
	result := testresult.TestResult{ID: "TST-88401", SampleID: "SMP-2569-00001", TestName: "CBC"}
	sample := domainsample.Sample{ID: "SMP-2569-00001", Name: "Blood Sample", CustodianUserID: 5, LocationID: &locationID}
	steps := []domainsample.CoCStep{{ID: 1, SampleID: "SMP-2569-00001", Title: "รับตัวอย่างเข้าระบบ"}}

	custodians := new(mockCustodianDir)
	custodians.On("FindByID", mock.Anything, int64(5)).Return(user.User{ID: 5, Name: "สมชาย"}, nil)
	results.On("FindByID", mock.Anything, "TST-88401").Return(result, nil)
	samples.On("FindByID", mock.Anything, "SMP-2569-00001").Return(sample, nil)
	coc.On("ListBySample", mock.Anything, "SMP-2569-00001").Return(steps, nil)
	locations.On("FullPath", mock.Anything, locationID).Return("Fridge-A / Shelf-2 / Slot-4", nil)

	uc := applicationtestresult.NewGenerateReportUseCase(results, samples, coc, locations, custodians)
	data, err := uc.Execute(context.Background(), "TST-88401")

	assert.NoError(t, err)
	assert.Equal(t, result, data.Result)
	assert.Equal(t, sample, data.Sample)
	assert.Equal(t, steps, data.CoCSteps)
	assert.Equal(t, "Fridge-A / Shelf-2 / Slot-4", data.LocationFullPath)
	assert.Equal(t, "สมชาย", data.CustodianName)
}

func TestGenerateReportUseCase_NoLocationAssigned(t *testing.T) {
	results := new(mockResultRepo)
	samples := new(mockSampleRepo)
	coc := new(mockCoCRepo)
	locations := new(mockLocationRepo)

	result := testresult.TestResult{ID: "TST-88402", SampleID: "SMP-2569-00002", TestName: "CBC"}
	sample := domainsample.Sample{ID: "SMP-2569-00002", Name: "Blood Sample", CustodianUserID: 5}

	custodians := new(mockCustodianDir)
	custodians.On("FindByID", mock.Anything, int64(5)).Return(user.User{ID: 5, Name: "สมชาย"}, nil)
	results.On("FindByID", mock.Anything, "TST-88402").Return(result, nil)
	samples.On("FindByID", mock.Anything, "SMP-2569-00002").Return(sample, nil)
	coc.On("ListBySample", mock.Anything, "SMP-2569-00002").Return([]domainsample.CoCStep{}, nil)

	uc := applicationtestresult.NewGenerateReportUseCase(results, samples, coc, locations, custodians)
	data, err := uc.Execute(context.Background(), "TST-88402")

	assert.NoError(t, err)
	assert.Equal(t, "-", data.LocationFullPath)
	locations.AssertNotCalled(t, "FullPath", mock.Anything, mock.Anything)
}

func TestGenerateReportUseCase_PropagatesNotFound(t *testing.T) {
	results := new(mockResultRepo)
	samples := new(mockSampleRepo)
	coc := new(mockCoCRepo)
	locations := new(mockLocationRepo)

	results.On("FindByID", mock.Anything, "TST-missing").Return(testresult.TestResult{}, assert.AnError)

	uc := applicationtestresult.NewGenerateReportUseCase(results, samples, coc, locations, new(mockCustodianDir))
	_, err := uc.Execute(context.Background(), "TST-missing")

	assert.ErrorIs(t, err, assert.AnError)
	samples.AssertNotCalled(t, "FindByID", mock.Anything, mock.Anything)
}
