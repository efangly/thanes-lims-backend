package environment

import (
	"context"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/environment"
	portenvironment "github.com/efangly/thanes-lims-backend/internal/ports/environment"
)

type RecordReadingUseCase struct {
	readings  portenvironment.ReadingRepository
	threshold *EvaluateThresholdsUseCase
}

func NewRecordReadingUseCase(readings portenvironment.ReadingRepository, threshold *EvaluateThresholdsUseCase) *RecordReadingUseCase {
	return &RecordReadingUseCase{readings: readings, threshold: threshold}
}

type RecordReadingResult struct {
	Reading environment.SensorReading
	Alert   *environment.EnvAlert
}

func (uc *RecordReadingUseCase) Execute(ctx context.Context, location string, value float64) (RecordReadingResult, error) {
	reading, err := uc.readings.Record(ctx, environment.SensorReading{
		Location:   location,
		Value:      value,
		RecordedAt: time.Now(),
	})
	if err != nil {
		return RecordReadingResult{}, err
	}

	alert, err := uc.threshold.Execute(ctx, location, value)
	if err != nil {
		return RecordReadingResult{}, err
	}

	return RecordReadingResult{Reading: reading, Alert: alert}, nil
}
