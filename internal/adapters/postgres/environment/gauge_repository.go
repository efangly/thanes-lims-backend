package environment

import (
	"context"
	"errors"

	"github.com/efangly/thanes-lims-backend/internal/domain/environment"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"gorm.io/gorm"
)

type GaugeRepository struct {
	db *gorm.DB
}

func NewGaugeRepository(db *gorm.DB) *GaugeRepository {
	return &GaugeRepository{db: db}
}

func gaugeToDomain(m GaugeModel) environment.Gauge {
	return environment.Gauge{Location: m.Location, Unit: m.Unit, RangeMin: m.RangeMin, RangeMax: m.RangeMax}
}

func (r *GaugeRepository) List(ctx context.Context) ([]environment.Gauge, error) {
	var models []GaugeModel
	if err := r.db.WithContext(ctx).Order("location").Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]environment.Gauge, len(models))
	for i, m := range models {
		out[i] = gaugeToDomain(m)
	}
	return out, nil
}

func (r *GaugeRepository) FindByLocation(ctx context.Context, location string) (environment.Gauge, error) {
	var m GaugeModel
	err := r.db.WithContext(ctx).First(&m, "location = ?", location).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return environment.Gauge{}, shared.ErrNotFound
	}
	if err != nil {
		return environment.Gauge{}, err
	}
	return gaugeToDomain(m), nil
}
