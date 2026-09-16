package tools

import (
	"context"
	"errors"
	"fmt"

	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/efangly/thanes-lims-backend/internal/domain/testresult"
	portstestresult "github.com/efangly/thanes-lims-backend/internal/ports/testresult"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestResultTools implements the TestResult-domain MCP tool: searchTestResults.
type TestResultTools struct {
	Results portstestresult.Repository
}

type TestResultOutput struct {
	ID       string `json:"id" jsonschema:"Test result ID, e.g. TR-2569-00005."`
	SampleID string `json:"sampleId" jsonschema:"ID of the Sample this result belongs to."`
	TestName string `json:"testName" jsonschema:"Name of the test performed, e.g. \"IgG\", \"Coliform Count\"."`
	Analyst  string `json:"analyst" jsonschema:"Name of the analyst who ran the test."`
	Result   string `json:"result" jsonschema:"The measured/observed result value, as free text."`
	Flag     string `json:"flag" jsonschema:"Abnormality flag: hi (above reference range), lo (below reference range), ok (within reference range)."`
	RefRange string `json:"refRange" jsonschema:"The reference range the result was checked against."`
	Status   string `json:"status" jsonschema:"Result lifecycle status: analyzing (in progress), pending_verification (awaiting sign-off), approved (finalized)."`
}

func toTestResultOutput(r testresult.TestResult) TestResultOutput {
	return TestResultOutput{
		ID:       r.ID,
		SampleID: r.SampleID,
		TestName: r.TestName,
		Analyst:  r.Analyst,
		Result:   r.Result,
		Flag:     string(r.Flag),
		RefRange: r.RefRange,
		Status:   string(r.Status),
	}
}

type SearchTestResultsInput struct {
	SampleID *string `json:"sampleId,omitempty" jsonschema:"When set, only return results for this Sample ID."`
	Status   *string `json:"status,omitempty" jsonschema:"When set, only return results with this status: analyzing, pending_verification, or approved."`
	// Flag is not on portstestresult.ListFilter (the repository has no
	// flag column filter), so it is applied client-side in the handler
	// after a broader List call, per the same narrowest-change preference
	// used for Sample's olderThanDays.
	Flag *string `json:"flag,omitempty" jsonschema:"When set, only return results with this abnormality flag: hi, lo, or ok. Filtered in this tool, not at the database layer."`
}

type SearchTestResultsOutput struct {
	Results []TestResultOutput `json:"results" jsonschema:"Matching test results."`
	Count   int                `json:"count" jsonschema:"Number of matching test results."`
}

// Search implements the searchTestResults tool.
func (t *TestResultTools) Search(ctx context.Context, _ *mcp.CallToolRequest, in SearchTestResultsInput) (*mcp.CallToolResult, SearchTestResultsOutput, error) {
	filter := portstestresult.ListFilter{SampleID: in.SampleID}
	if in.Status != nil {
		st := testresult.Status(*in.Status)
		switch st {
		case testresult.StatusAnalyzing, testresult.StatusPendingVerification, testresult.StatusApproved:
			filter.Status = &st
		default:
			return errorResult[SearchTestResultsOutput](fmt.Sprintf("invalid status %q: must be one of analyzing, pending_verification, approved", *in.Status))
		}
	}
	if in.Flag != nil {
		switch testresult.Flag(*in.Flag) {
		case testresult.FlagHi, testresult.FlagLo, testresult.FlagOk:
		default:
			return errorResult[SearchTestResultsOutput](fmt.Sprintf("invalid flag %q: must be one of hi, lo, ok", *in.Flag))
		}
	}

	results, err := t.Results.List(ctx, filter)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return nil, SearchTestResultsOutput{Results: []TestResultOutput{}, Count: 0}, nil
		}
		return nil, SearchTestResultsOutput{}, err
	}

	out := make([]TestResultOutput, 0, len(results))
	for _, r := range results {
		if in.Flag != nil && string(r.Flag) != *in.Flag {
			continue
		}
		out = append(out, toTestResultOutput(r))
	}
	return nil, SearchTestResultsOutput{Results: out, Count: len(out)}, nil
}
