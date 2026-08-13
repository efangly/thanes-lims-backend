package equipment

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/equipment"
)

type CalibrationRepository interface {
	Append(ctx context.Context, ev equipment.CalibrationEvent) (equipment.CalibrationEvent, error)
	ListByEquipment(ctx context.Context, equipmentID string) ([]equipment.CalibrationEvent, error)
}
