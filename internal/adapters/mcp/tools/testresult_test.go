package tools_test

import (
	"context"
	"testing"

	"github.com/efangly/thanes-lims-backend/internal/adapters/mcp/tools"
	"github.com/efangly/thanes-lims-backend/internal/domain/testresult"
	portstestresult "github.com/efangly/thanes-lims-backend/internal/ports/testresult"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTestResultTools_Search_FiltersByFlagClientSide(t *testing.T) {
	repo := new(mockTestResultRepository)
	tt := &tools.TestResultTools{Results: repo}

	repo.On("List", mock.Anything, portstestresult.ListFilter{}).Return([]testresult.TestResult{
		{ID: "TR-1", Flag: testresult.FlagHi, TestName: "IgG"},
		{ID: "TR-2", Flag: testresult.FlagOk, TestName: "Glucose"},
		{ID: "TR-3", Flag: testresult.FlagLo, TestName: "Lead (Pb)"},
	}, nil)

	flag := "hi"
	_, out, err := tt.Search(context.Background(), nil, tools.SearchTestResultsInput{Flag: &flag})

	assert.NoError(t, err)
	assert.Equal(t, 1, out.Count)
	assert.Equal(t, "TR-1", out.Results[0].ID)
}

func TestTestResultTools_Search_InvalidStatus(t *testing.T) {
	tt := &tools.TestResultTools{}
	status := "bogus"
	result, _, err := tt.Search(context.Background(), nil, tools.SearchTestResultsInput{Status: &status})
	assert.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestTestResultTools_Search_InvalidFlag(t *testing.T) {
	tt := &tools.TestResultTools{}
	flag := "bogus"
	result, _, err := tt.Search(context.Background(), nil, tools.SearchTestResultsInput{Flag: &flag})
	assert.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestTestResultTools_Search_BySampleID(t *testing.T) {
	repo := new(mockTestResultRepository)
	tt := &tools.TestResultTools{Results: repo}

	sampleID := "SMP-1"
	repo.On("List", mock.Anything, portstestresult.ListFilter{SampleID: &sampleID}).Return([]testresult.TestResult{
		{ID: "TR-1", SampleID: "SMP-1", Flag: testresult.FlagOk},
	}, nil)

	_, out, err := tt.Search(context.Background(), nil, tools.SearchTestResultsInput{SampleID: &sampleID})

	assert.NoError(t, err)
	assert.Equal(t, 1, out.Count)
}
