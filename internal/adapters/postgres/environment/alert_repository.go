package environment

import (
	"context"
	"errors"

	"github.com/efangly/thanes-lims-backend/internal/domain/environment"
	"gorm.io/gorm"
)

type AlertRepository struct {
	db *gorm.DB
}

func NewAlertRepository(db *gorm.DB) *AlertRepository {
	return &AlertRepository{db: db}
}

func alertToDomain(m AlertModel) environment.EnvAlert {
	return environment.EnvAlert{
		ID:          m.ID,
		Location:    m.Location,
		Level:       environment.Level(m.Level),
		Title:       m.Title,
		Message:     m.Message,
		TriggeredAt: m.TriggeredAt,
		ResolvedAt:  m.ResolvedAt,
	}
}

func (r *AlertRepository) Create(ctx context.Context, a environment.EnvAlert) (environment.EnvAlert, error) {
	m := AlertModel{
		Location:    a.Location,
		Level:       string(a.Level),
		Title:       a.Title,
		Message:     a.Message,
		TriggeredAt: a.TriggeredAt,
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return environment.EnvAlert{}, err
	}
	return alertToDomain(m), nil
}

func (r *AlertRepository) FindOpenByLocation(ctx context.Context, location string) (*environment.EnvAlert, error) {
	var m AlertModel
	err := r.db.WithContext(ctx).Where("location = ? AND resolved_at IS NULL", location).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	a := alertToDomain(m)
	return &a, nil
}

func (r *AlertRepository) List(ctx context.Context) ([]environment.EnvAlert, error) {
	var models []AlertModel
	if err := r.db.WithContext(ctx).Order("triggered_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]environment.EnvAlert, len(models))
	for i, m := range models {
		out[i] = alertToDomain(m)
	}
	return out, nil
}

func (r *AlertRepository) Resolve(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&AlertModel{}).Where("id = ?", id).Update("resolved_at", gorm.Expr("now()")).Error
}
