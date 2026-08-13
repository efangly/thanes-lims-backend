package environment

import (
	"context"
	"fmt"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/environment"
	"github.com/efangly/thanes-lims-backend/internal/domain/notification"
	portenvironment "github.com/efangly/thanes-lims-backend/internal/ports/environment"
	portnotification "github.com/efangly/thanes-lims-backend/internal/ports/notification"
)

type EvaluateThresholdsUseCase struct {
	gauges      portenvironment.GaugeRepository
	alerts      portenvironment.AlertRepository
	notifier    portnotification.Notifier
	broadcaster portenvironment.AlertBroadcaster
}

func NewEvaluateThresholdsUseCase(gauges portenvironment.GaugeRepository, alerts portenvironment.AlertRepository, notifier portnotification.Notifier, broadcaster portenvironment.AlertBroadcaster) *EvaluateThresholdsUseCase {
	return &EvaluateThresholdsUseCase{gauges: gauges, alerts: alerts, notifier: notifier, broadcaster: broadcaster}
}

// Execute derives the reading's Level and manages alert lifecycle: at most
// one open alert per location at a time (no duplicate-alert spam for a
// sustained excursion), auto-resolved once a reading returns to OK. Returns
// the alert if one is currently open (newly created or pre-existing), or
// nil if the location is OK. Kept as its own use case, separate from
// RecordReading, so the WebSocket broadcaster and Notifier hook in here
// without touching ingestion logic.
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

	title := fmt.Sprintf("ค่าผิดปกติที่ %s", location)
	message := fmt.Sprintf("ค่าที่อ่านได้ %.2f อยู่นอกช่วงที่กำหนด (%.2f-%.2f)", value, gauge.RangeMin, gauge.RangeMax)

	created, err := uc.alerts.Create(ctx, environment.EnvAlert{
		Location:    location,
		Level:       level,
		Title:       title,
		Message:     message,
		TriggeredAt: time.Now(),
	})
	if err != nil {
		return nil, err
	}

	tone := notification.ToneAmber
	if level == environment.LevelCrit {
		tone = notification.ToneRed
	}
	if err := uc.notifier.Notify(ctx, notification.Notification{
		Tone:    tone,
		Title:   title,
		Message: message,
	}); err != nil {
		return nil, err
	}
	uc.broadcaster.Broadcast(created)

	return &created, nil
}
