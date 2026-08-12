package testresult

import (
	"context"
	"errors"

	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/efangly/thanes-lims-backend/internal/domain/testresult"
	porttestresult "github.com/efangly/thanes-lims-backend/internal/ports/testresult"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func toDomain(m Model) testresult.TestResult {
	return testresult.TestResult{
		ID:       m.ID,
		SampleID: m.SampleID,
		TestName: m.TestName,
		Analyst:  m.Analyst,
		Result:   m.Result,
		Flag:     testresult.Flag(m.Flag),
		RefRange: m.RefRange,
		Status:   testresult.Status(m.Status),
	}
}

func toModel(t testresult.TestResult) Model {
	flag := t.Flag
	if flag == "" {
		flag = testresult.FlagOk
	}
	return Model{
		ID:       t.ID,
		SampleID: t.SampleID,
		TestName: t.TestName,
		Analyst:  t.Analyst,
		Result:   t.Result,
		Flag:     string(flag),
		RefRange: t.RefRange,
		Status:   string(t.Status),
	}
}

func (r *Repository) Create(ctx context.Context, t testresult.TestResult) (testresult.TestResult, error) {
	m := toModel(t)
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return testresult.TestResult{}, err
	}
	return toDomain(m), nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (testresult.TestResult, error) {
	var m Model
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return testresult.TestResult{}, shared.ErrNotFound
	}
	if err != nil {
		return testresult.TestResult{}, err
	}
	return toDomain(m), nil
}

func (r *Repository) List(ctx context.Context, filter porttestresult.ListFilter) ([]testresult.TestResult, error) {
	q := r.db.WithContext(ctx).Model(&Model{})
	if filter.SampleID != nil {
		q = q.Where("sample_id = ?", *filter.SampleID)
	}
	if filter.Status != nil {
		q = q.Where("status = ?", string(*filter.Status))
	}

	var models []Model
	if err := q.Order("id").Find(&models).Error; err != nil {
		return nil, err
	}
	results := make([]testresult.TestResult, len(models))
	for i, m := range models {
		results[i] = toDomain(m)
	}
	return results, nil
}

func (r *Repository) Update(ctx context.Context, t testresult.TestResult) (testresult.TestResult, error) {
	m := toModel(t)
	if err := r.db.WithContext(ctx).Model(&Model{}).Where("id = ?", t.ID).Updates(&m).Error; err != nil {
		return testresult.TestResult{}, err
	}
	return r.FindByID(ctx, t.ID)
}
