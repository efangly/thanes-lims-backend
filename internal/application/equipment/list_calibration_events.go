package equipment

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/equipment"
	portequipment "github.com/efangly/thanes-lims-backend/internal/ports/equipment"
)

type ListCalibrationEventsUseCase struct {
	calibration portequipment.CalibrationRepository
}

func NewListCalibrationEventsUseCase(calibration portequipment.CalibrationRepository) *ListCalibrationEventsUseCase {
	return &ListCalibrationEventsUseCase{calibration: calibration}
}

func (uc *ListCalibrationEventsUseCase) Execute(ctx context.Context, equipmentID string) ([]equipment.CalibrationEvent, error) {
	return uc.calibration.ListByEquipment(ctx, equipmentID)
}
