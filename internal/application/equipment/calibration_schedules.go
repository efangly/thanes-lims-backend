package equipment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/equipment"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portequipment "github.com/efangly/thanes-lims-backend/internal/ports/equipment"
)

// CalibrationScheduleUseCase covers CRUD for the recurring calibration
// commitments on one Equipment (Phase 6).
type CalibrationScheduleUseCase struct {
	equipment portequipment.Repository
	schedules portequipment.ScheduleRepository
}

func NewCalibrationScheduleUseCase(
	equipment portequipment.Repository,
	schedules portequipment.ScheduleRepository,
) *CalibrationScheduleUseCase {
	return &CalibrationScheduleUseCase{equipment: equipment, schedules: schedules}
}

type CreateCalibrationScheduleInput struct {
	EquipmentID    string
	Label          string
	NextDueDate    time.Time
	IntervalMonths *int
}

func (uc *CalibrationScheduleUseCase) Create(ctx context.Context, in CreateCalibrationScheduleInput) (equipment.CalibrationSchedule, error) {
	if _, err := uc.equipment.FindByID(ctx, in.EquipmentID); err != nil {
		return equipment.CalibrationSchedule{}, err
	}
	label := strings.TrimSpace(in.Label)
	if label == "" {
		return equipment.CalibrationSchedule{}, fmt.Errorf("%w: label is required", shared.ErrValidation)
	}
	if in.NextDueDate.IsZero() {
		return equipment.CalibrationSchedule{}, fmt.Errorf("%w: next_due_date is required", shared.ErrValidation)
	}
	if in.IntervalMonths != nil && *in.IntervalMonths <= 0 {
		return equipment.CalibrationSchedule{}, fmt.Errorf("%w: interval_months must be positive", shared.ErrValidation)
	}
	return uc.schedules.Create(ctx, equipment.CalibrationSchedule{
		EquipmentID:    in.EquipmentID,
		Label:          label,
		NextDueDate:    in.NextDueDate,
		IntervalMonths: in.IntervalMonths,
	})
}

func (uc *CalibrationScheduleUseCase) ListByEquipment(ctx context.Context, equipmentID string) ([]equipment.CalibrationSchedule, error) {
	return uc.schedules.ListByEquipment(ctx, equipmentID)
}

type UpdateCalibrationScheduleInput struct {
	EquipmentID    string
	ID             int64
	Label          *string
	NextDueDate    *time.Time
	IntervalMonths *int
	ClearInterval  bool
}

func (uc *CalibrationScheduleUseCase) Update(ctx context.Context, in UpdateCalibrationScheduleInput) (equipment.CalibrationSchedule, error) {
	s, err := uc.schedules.FindByID(ctx, in.ID)
	if err != nil {
		return equipment.CalibrationSchedule{}, err
	}
	if s.EquipmentID != in.EquipmentID {
		return equipment.CalibrationSchedule{}, shared.ErrNotFound
	}
	if in.Label != nil {
		label := strings.TrimSpace(*in.Label)
		if label == "" {
			return equipment.CalibrationSchedule{}, fmt.Errorf("%w: label is required", shared.ErrValidation)
		}
		s.Label = label
	}
	if in.NextDueDate != nil {
		s.NextDueDate = *in.NextDueDate
	}
	if in.ClearInterval {
		s.IntervalMonths = nil
	} else if in.IntervalMonths != nil {
		if *in.IntervalMonths <= 0 {
			return equipment.CalibrationSchedule{}, fmt.Errorf("%w: interval_months must be positive", shared.ErrValidation)
		}
		s.IntervalMonths = in.IntervalMonths
	}
	return uc.schedules.Update(ctx, s)
}

func (uc *CalibrationScheduleUseCase) Delete(ctx context.Context, equipmentID string, id int64) error {
	s, err := uc.schedules.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if s.EquipmentID != equipmentID {
		return shared.ErrNotFound
	}
	return uc.schedules.Delete(ctx, id)
}
