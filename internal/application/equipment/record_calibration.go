package equipment

import (
	"context"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/equipment"
	portequipment "github.com/efangly/thanes-lims-backend/internal/ports/equipment"
)

type RecordCalibrationUseCase struct {
	equipment   portequipment.Repository
	calibration portequipment.CalibrationRepository
}

func NewRecordCalibrationUseCase(equipment portequipment.Repository, calibration portequipment.CalibrationRepository) *RecordCalibrationUseCase {
	return &RecordCalibrationUseCase{equipment: equipment, calibration: calibration}
}

type RecordCalibrationInput struct {
	ID                 string
	NextCalibrationDue time.Time
	PerformedBy        string
	Notes              string
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

	if _, err := uc.calibration.Append(ctx, equipment.CalibrationEvent{
		EquipmentID:        e.ID,
		CalibratedAt:       calibratedAt,
		NextCalibrationDue: in.NextCalibrationDue,
		PerformedBy:        in.PerformedBy,
		Notes:              in.Notes,
	}); err != nil {
		return equipment.Equipment{}, err
	}

	return e, nil
}
