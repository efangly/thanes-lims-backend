package environment

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/environment"
)

type GaugeRepository interface {
	List(ctx context.Context) ([]environment.Gauge, error)
	FindByLocation(ctx context.Context, location string) (environment.Gauge, error)
}

type ReadingRepository interface {
	Record(ctx context.Context, r environment.SensorReading) (environment.SensorReading, error)
	LatestByLocation(ctx context.Context, location string) (environment.SensorReading, error)
	ListTrend(ctx context.Context, location string, limit int) ([]environment.SensorReading, error)
}

// AlertRepository.FindOpenByLocation returns (nil, nil) when no alert is
// currently open for the location - that's the normal case, not an error.
type AlertRepository interface {
	Create(ctx context.Context, a environment.EnvAlert) (environment.EnvAlert, error)
	FindOpenByLocation(ctx context.Context, location string) (*environment.EnvAlert, error)
	List(ctx context.Context) ([]environment.EnvAlert, error)
	Resolve(ctx context.Context, id int64) error
}
