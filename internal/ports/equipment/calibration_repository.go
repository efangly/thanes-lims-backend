package equipment

import (
	"context"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/equipment"
)

type CalibrationRepository interface {
	Append(ctx context.Context, ev equipment.CalibrationEvent) (equipment.CalibrationEvent, error)
	ListByEquipment(ctx context.Context, equipmentID string) ([]equipment.CalibrationEvent, error)
	FindByID(ctx context.Context, id int64) (equipment.CalibrationEvent, error)
	// Search backs the calibration results page (requirement 2.2.1) - a
	// flat, filterable list across every Equipment.
	Search(ctx context.Context, f CalibrationSearchFilter) ([]equipment.CalibrationEvent, error)
}

// CalibrationSearchFilter is the set of optional narrowing criteria for the
// calibration results page. A zero value matches everything.
type CalibrationSearchFilter struct {
	Query       string // ILIKE against equipment id/name, performed_by, calibration_type, notes
	EquipmentID string
	Result      equipment.CalibrationResult
	From        *time.Time // CalibratedAt >= From
	To          *time.Time // CalibratedAt <= To
}

// ScheduleRepository persists CalibrationSchedule rows (Phase 6).
type ScheduleRepository interface {
	Create(ctx context.Context, s equipment.CalibrationSchedule) (equipment.CalibrationSchedule, error)
	FindByID(ctx context.Context, id int64) (equipment.CalibrationSchedule, error)
	ListByEquipment(ctx context.Context, equipmentID string) ([]equipment.CalibrationSchedule, error)
	Update(ctx context.Context, s equipment.CalibrationSchedule) (equipment.CalibrationSchedule, error)
	Delete(ctx context.Context, id int64) error
}
