package environment

import (
	"context"
	"fmt"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/environment"
	portenvironment "github.com/efangly/thanes-lims-backend/internal/ports/environment"
)

type EvaluateThresholdsUseCase struct {
	gauges portenvironment.GaugeRepository
	alerts portenvironment.AlertRepository
}

func NewEvaluateThresholdsUseCase(gauges portenvironment.GaugeRepository, alerts portenvironment.AlertRepository) *EvaluateThresholdsUseCase {
	return &EvaluateThresholdsUseCase{gauges: gauges, alerts: alerts}
}

// Execute derives the reading's Level and manages alert lifecycle: at most
// one open alert per location at a time (no duplicate-alert spam for a
// sustained excursion), auto-resolved once a reading returns to OK. Returns
// the alert if one is currently open (newly created or pre-existing), or
// nil if the location is OK. Kept as its own use case, separate from
// RecordReading, so a future WebSocket broadcaster can hook in here without
// touching ingestion logic.
func (uc *EvaluateThresholdsUseCase) Execute(ctx context.Context, location string, value float64) (*environment.EnvAlert, error) {
	gauge, err := uc.gauges.FindByLocation(ctx, location)
	if err != nil {
		return nil, err
	}
	level := environment.DeriveLevel(value, gauge)

	existing, err := uc.alerts.FindOpenByLocation(ctx, location)
	if err != nil {
		return nil, err
	}

	if level == environment.LevelOK {
		if existing != nil {
			if err := uc.alerts.Resolve(ctx, existing.ID); err != nil {
				return nil, err
			}
		}
		return nil, nil
	}

	if existing != nil {
		return existing, nil
	}

	created, err := uc.alerts.Create(ctx, environment.EnvAlert{
		Location:    location,
		Level:       level,
		Title:       fmt.Sprintf("ค่าผิดปกติที่ %s", location),
		Message:     fmt.Sprintf("ค่าที่อ่านได้ %.2f อยู่นอกช่วงที่กำหนด (%.2f-%.2f)", value, gauge.RangeMin, gauge.RangeMax),
		TriggeredAt: time.Now(),
	})
	if err != nil {
		return nil, err
	}
	return &created, nil
}
