package sample

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/sample"
)

// CoCRepository is also depended on by the testresult module (approving a
// test result appends a CoC step to its linked Sample), so it lives in its
// own file separate from the Sample aggregate's own repository.
type CoCRepository interface {
	AppendStep(ctx context.Context, step sample.CoCStep) (sample.CoCStep, error)
	ListBySample(ctx context.Context, sampleID string) ([]sample.CoCStep, error)
}
