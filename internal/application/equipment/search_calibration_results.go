package equipment

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/equipment"
	portequipment "github.com/efangly/thanes-lims-backend/internal/ports/equipment"
)

// SearchCalibrationResultsUseCase backs the calibration results page
// (requirement 2.2.1) - a flat, searchable list of every logged
// CalibrationEvent across all Equipment.
type SearchCalibrationResultsUseCase struct {
	calibration portequipment.CalibrationRepository
}

func NewSearchCalibrationResultsUseCase(calibration portequipment.CalibrationRepository) *SearchCalibrationResultsUseCase {
	return &SearchCalibrationResultsUseCase{calibration: calibration}
}

func (uc *SearchCalibrationResultsUseCase) Execute(ctx context.Context, f portequipment.CalibrationSearchFilter) ([]equipment.CalibrationEvent, error) {
	return uc.calibration.Search(ctx, f)
}
