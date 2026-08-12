package sample

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/sample"
	"gorm.io/gorm"
)

type CoCRepository struct {
	db *gorm.DB
}

func NewCoCRepository(db *gorm.DB) *CoCRepository {
	return &CoCRepository{db: db}
}

func cocToDomain(m CoCModel) sample.CoCStep {
	return sample.CoCStep{
		ID:         m.ID,
		SampleID:   m.SampleID,
		State:      sample.CoCState(m.State),
		Icon:       sample.CoCIcon(m.Icon),
		Title:      m.Title,
		Meta:       m.Meta,
		Who:        m.Who,
		OccurredAt: m.OccurredAt,
	}
}

func (r *CoCRepository) AppendStep(ctx context.Context, step sample.CoCStep) (sample.CoCStep, error) {
	m := CoCModel{
		SampleID:   step.SampleID,
		State:      string(step.State),
		Icon:       string(step.Icon),
		Title:      step.Title,
		Meta:       step.Meta,
		Who:        step.Who,
		OccurredAt: step.OccurredAt,
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return sample.CoCStep{}, err
	}
	return cocToDomain(m), nil
}

func (r *CoCRepository) ListBySample(ctx context.Context, sampleID string) ([]sample.CoCStep, error) {
	var models []CoCModel
	if err := r.db.WithContext(ctx).Where("sample_id = ?", sampleID).Order("occurred_at").Find(&models).Error; err != nil {
		return nil, err
	}
	steps := make([]sample.CoCStep, len(models))
	for i, m := range models {
		steps[i] = cocToDomain(m)
	}
	return steps, nil
}
