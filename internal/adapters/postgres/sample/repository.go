package sample

import (
	"context"
	"errors"

	"github.com/efangly/thanes-lims-backend/internal/domain/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portsample "github.com/efangly/thanes-lims-backend/internal/ports/sample"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func toDomain(m Model) sample.Sample {
	return sample.Sample{
		ID:         m.ID,
		Name:       m.Name,
		Type:       sample.Type(m.Type),
		Custodian:  m.Custodian,
		Location:   m.Location,
		Status:     sample.Status(m.Status),
		ReceivedAt: m.ReceivedAt,
	}
}

func toModel(s sample.Sample) Model {
	return Model{
		ID:         s.ID,
		Name:       s.Name,
		Type:       string(s.Type),
		Custodian:  s.Custodian,
		Location:   s.Location,
		Status:     string(s.Status),
		ReceivedAt: s.ReceivedAt,
	}
}

func (r *Repository) Create(ctx context.Context, s sample.Sample) (sample.Sample, error) {
	m := toModel(s)
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return sample.Sample{}, err
	}
	return toDomain(m), nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (sample.Sample, error) {
	var m Model
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return sample.Sample{}, shared.ErrNotFound
	}
	if err != nil {
		return sample.Sample{}, err
	}
	return toDomain(m), nil
}

func (r *Repository) List(ctx context.Context, filter portsample.ListFilter) ([]sample.Sample, error) {
	q := r.db.WithContext(ctx).Model(&Model{})
	if filter.Status != nil {
		q = q.Where("status = ?", string(*filter.Status))
	}
	if filter.Type != nil {
		q = q.Where("type = ?", string(*filter.Type))
	}

	var models []Model
	if err := q.Order("received_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}

	samples := make([]sample.Sample, len(models))
	for i, m := range models {
		samples[i] = toDomain(m)
	}
	return samples, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, s sample.Sample) (sample.Sample, error) {
	if err := r.db.WithContext(ctx).Model(&Model{}).Where("id = ?", s.ID).Update("status", string(s.Status)).Error; err != nil {
		return sample.Sample{}, err
	}
	return r.FindByID(ctx, s.ID)
}
