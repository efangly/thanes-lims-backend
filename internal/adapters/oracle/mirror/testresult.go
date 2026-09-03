package mirror

import (
	"context"
	"database/sql"

	dtestresult "github.com/efangly/thanes-lims-backend/internal/domain/testresult"
	porttestresult "github.com/efangly/thanes-lims-backend/internal/ports/testresult"
)

const mergeTestResultSQL = `
MERGE INTO test_results t
USING (SELECT :id AS id FROM dual) s ON (t.id = s.id)
WHEN MATCHED THEN UPDATE SET
    sample_id = :sid, test_name = :tname, analyst = :analyst,
    result = :result, flag = :flag, ref_range = :refrange, status = :status
WHEN NOT MATCHED THEN INSERT
    (id, sample_id, test_name, analyst, result, flag, ref_range, status)
VALUES (:id, :sid, :tname, :analyst, :result, :flag, :refrange, :status)
`

// UpsertTestResult MERGEs one TestResult into the mirror. It maps 1:1 onto the
// mirror table; empty result/flag/ref_range become NULL (flag has a CHECK).
func (m *Mirror) UpsertTestResult(ctx context.Context, t dtestresult.TestResult) error {
	_, err := m.db.ExecContext(ctx, mergeTestResultSQL,
		sql.Named("id", t.ID),
		sql.Named("sid", t.SampleID),
		sql.Named("tname", clip(t.TestName, 200)),
		sql.Named("analyst", clip(t.Analyst, 100)),
		sql.Named("result", clipNull(t.Result, 200)),
		sql.Named("flag", nullText(string(t.Flag))),
		sql.Named("refrange", clipNull(t.RefRange, 100)),
		sql.Named("status", string(t.Status)),
	)
	return err
}

type testResultRepo struct {
	porttestresult.Repository
	mirror *Mirror
}

// WrapTestResult returns next wrapped so Create/Update also mirror to Oracle.
func WrapTestResult(next porttestresult.Repository, m *Mirror) porttestresult.Repository {
	return &testResultRepo{Repository: next, mirror: m}
}

func (r *testResultRepo) Create(ctx context.Context, t dtestresult.TestResult) (dtestresult.TestResult, error) {
	created, err := r.Repository.Create(ctx, t)
	if err != nil {
		return created, err
	}
	run(ctx, "test_result", created.ID, func(ctx context.Context) error {
		return r.mirror.UpsertTestResult(ctx, created)
	})
	return created, nil
}

func (r *testResultRepo) Update(ctx context.Context, t dtestresult.TestResult) (dtestresult.TestResult, error) {
	updated, err := r.Repository.Update(ctx, t)
	if err != nil {
		return updated, err
	}
	run(ctx, "test_result", updated.ID, func(ctx context.Context) error {
		return r.mirror.UpsertTestResult(ctx, updated)
	})
	return updated, nil
}
