package environment

import (
	"context"
	"errors"

	"github.com/efangly/thanes-lims-backend/internal/domain/environment"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"gorm.io/gorm"
)

type ReadingRepository struct {
	db *gorm.DB
}

func NewReadingRepository(db *gorm.DB) *ReadingRepository {
	return &ReadingRepository{db: db}
}

func readingToDomain(m ReadingModel) environment.SensorReading {
	return environment.SensorReading{ID: m.ID, Location: m.Location, Value: m.Value, RecordedAt: m.RecordedAt}
}

func (r *ReadingRepository) Record(ctx context.Context, reading environment.SensorReading) (environment.SensorReading, error) {
	m := ReadingModel{Location: reading.Location, Value: reading.Value, RecordedAt: reading.RecordedAt}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return environment.SensorReading{}, err
	}
	return readingToDomain(m), nil
}

func (r *ReadingRepository) LatestByLocation(ctx context.Context, location string) (environment.SensorReading, error) {
	var m ReadingModel
	err := r.db.WithContext(ctx).Where("location = ?", location).Order("recorded_at DESC").First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return environment.SensorReading{}, shared.ErrNotFound
	}
	if err != nil {
		return environment.SensorReading{}, err
	}
	return readingToDomain(m), nil
}

func (r *ReadingRepository) ListTrend(ctx context.Context, location string, limit int) ([]environment.SensorReading, error) {
	var models []ReadingModel
	if err := r.db.WithContext(ctx).Where("location = ?", location).Order("recorded_at DESC").Limit(limit).Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]environment.SensorReading, len(models))
	for i, m := range models {
		out[i] = readingToDomain(m)
	}
	return out, nil
}
