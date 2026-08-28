package equipment

import (
	"context"
	"strings"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/equipment"
	portequipment "github.com/efangly/thanes-lims-backend/internal/ports/equipment"
)

type RecordCalibrationUseCase struct {
	equipment   portequipment.Repository
	calibration portequipment.CalibrationRepository
	schedules   portequipment.ScheduleRepository
}

func NewRecordCalibrationUseCase(
	equipment portequipment.Repository,
	calibration portequipment.CalibrationRepository,
	schedules portequipment.ScheduleRepository,
) *RecordCalibrationUseCase {
	return &RecordCalibrationUseCase{equipment: equipment, calibration: calibration, schedules: schedules}
}

type RecordCalibrationInput struct {
	ID                 string
	NextCalibrationDue time.Time
	PerformedBy        string
	Notes              string

	CalibrationType string
	CalibrateValue  string
	AcceptanceValue string
	Result          equipment.CalibrationResult
}

func (uc *RecordCalibrationUseCase) Execute(ctx context.Context, in RecordCalibrationInput) (equipment.Equipment, error) {
	e, err := uc.equipment.FindByID(ctx, in.ID)
	if err != nil {
		return equipment.Equipment{}, err
	}

	calibratedAt := time.Now()
	e.LastCalibratedAt = calibratedAt
	e.NextCalibrationDue = in.NextCalibrationDue

	e, err = uc.equipment.Update(ctx, e)
	if err != nil {
		return equipment.Equipment{}, err
	}

	calType := strings.TrimSpace(in.CalibrationType)
	if _, err := uc.calibration.Append(ctx, equipment.CalibrationEvent{
		EquipmentID:        e.ID,
		CalibratedAt:       calibratedAt,
		NextCalibrationDue: in.NextCalibrationDue,
		PerformedBy:        in.PerformedBy,
		Notes:              in.Notes,
		CalibrationType:    calType,
		CalibrateValue:     strings.TrimSpace(in.CalibrateValue),
		AcceptanceValue:    strings.TrimSpace(in.AcceptanceValue),
		Result:             in.Result,
	}); err != nil {
		return equipment.Equipment{}, err
	}

	// Auto-advance any schedule whose Label matches this calibration's type
	// and that carries an interval (CONTEXT.md "Calibration Schedule").
	if uc.schedules != nil && calType != "" {
		scheds, err := uc.schedules.ListByEquipment(ctx, e.ID)
		if err != nil {
			return equipment.Equipment{}, err
		}
		for _, s := range scheds {
			if !strings.EqualFold(strings.TrimSpace(s.Label), calType) {
				continue
			}
			next, ok := s.Advance(calibratedAt)
			if !ok {
				continue
			}
			s.NextDueDate = next
			if _, err := uc.schedules.Update(ctx, s); err != nil {
				return equipment.Equipment{}, err
			}
		}
	}

	return e, nil
}
