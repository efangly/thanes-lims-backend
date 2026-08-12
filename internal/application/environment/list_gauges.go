package environment

import (
	"context"
	"errors"

	"github.com/efangly/thanes-lims-backend/internal/domain/environment"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portenvironment "github.com/efangly/thanes-lims-backend/internal/ports/environment"
)

// GaugeStatus is a read-model combining a Gauge's static config with its
// latest reading and derived level - assembled here since neither Gauge nor
// SensorReading alone represents "current state of a monitoring point".
type GaugeStatus struct {
	Gauge   environment.Gauge
	Reading *environment.SensorReading
	Level   environment.Level
}

type ListGaugesUseCase struct {
	gauges   portenvironment.GaugeRepository
	readings portenvironment.ReadingRepository
}

func NewListGaugesUseCase(gauges portenvironment.GaugeRepository, readings portenvironment.ReadingRepository) *ListGaugesUseCase {
	return &ListGaugesUseCase{gauges: gauges, readings: readings}
}

func (uc *ListGaugesUseCase) Execute(ctx context.Context) ([]GaugeStatus, error) {
	gauges, err := uc.gauges.List(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]GaugeStatus, len(gauges))
	for i, g := range gauges {
		status := GaugeStatus{Gauge: g, Level: environment.LevelOK}

		latest, err := uc.readings.LatestByLocation(ctx, g.Location)
		if err != nil && !errors.Is(err, shared.ErrNotFound) {
			return nil, err
		}
		if err == nil {
			status.Reading = &latest
			status.Level = environment.DeriveLevel(latest.Value, g)
		}
		out[i] = status
	}
	return out, nil
}

type GetTrendUseCase struct {
	readings portenvironment.ReadingRepository
}

func NewGetTrendUseCase(readings portenvironment.ReadingRepository) *GetTrendUseCase {
	return &GetTrendUseCase{readings: readings}
}

func (uc *GetTrendUseCase) Execute(ctx context.Context, location string, limit int) ([]environment.SensorReading, error) {
	return uc.readings.ListTrend(ctx, location, limit)
}

type ListAlertsUseCase struct {
	alerts portenvironment.AlertRepository
}

func NewListAlertsUseCase(alerts portenvironment.AlertRepository) *ListAlertsUseCase {
	return &ListAlertsUseCase{alerts: alerts}
}

func (uc *ListAlertsUseCase) Execute(ctx context.Context) ([]environment.EnvAlert, error) {
	return uc.alerts.List(ctx)
}
