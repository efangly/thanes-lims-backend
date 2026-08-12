package equipment

import (
	"context"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/equipment"
	portequipment "github.com/efangly/thanes-lims-backend/internal/ports/equipment"
)

type RecordCalibrationUseCase struct {
	equipment portequipment.Repository
}

func NewRecordCalibrationUseCase(equipment portequipment.Repository) *RecordCalibrationUseCase {
	return &RecordCalibrationUseCase{equipment: equipment}
}

type RecordCalibrationInput struct {
	ID                 string
	NextCalibrationDue time.Time
}

func (uc *RecordCalibrationUseCase) Execute(ctx context.Context, in RecordCalibrationInput) (equipment.Equipment, error) {
	e, err := uc.equipment.FindByID(ctx, in.ID)
	if err != nil {
		return equipment.Equipment{}, err
	}

	e.LastCalibratedAt = time.Now()
	e.NextCalibrationDue = in.NextCalibrationDue

	return uc.equipment.Update(ctx, e)
}
