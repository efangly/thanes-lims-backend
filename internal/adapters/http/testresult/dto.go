package testresult

import "github.com/efangly/thanes-lims-backend/internal/domain/testresult"

type CreateTestResultRequest struct {
	SampleID string `json:"sample_id" validate:"required"`
	TestName string `json:"test_name" validate:"required"`
	Analyst  string `json:"analyst" validate:"required"`
	RefRange string `json:"ref_range"`
}

type SubmitResultRequest struct {
	Result string `json:"result" validate:"required"`
	Flag   string `json:"flag" validate:"required,oneof=hi lo ok"`
}

type TestResultResponse struct {
	ID       string `json:"id"`
	SampleID string `json:"sample_id"`
	TestName string `json:"test_name"`
	Analyst  string `json:"analyst"`
	Result   string `json:"result"`
	Flag     string `json:"flag"`
	RefRange string `json:"ref_range"`
	Status   string `json:"status"`
}

func toResponse(t testresult.TestResult) TestResultResponse {
	return TestResultResponse{
		ID:       t.ID,
		SampleID: t.SampleID,
		TestName: t.TestName,
		Analyst:  t.Analyst,
		Result:   t.Result,
		Flag:     string(t.Flag),
		RefRange: t.RefRange,
		Status:   string(t.Status),
	}
}
