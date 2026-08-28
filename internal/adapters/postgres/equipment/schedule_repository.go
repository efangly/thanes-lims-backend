package equipment

import (
	"context"
	"errors"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/equipment"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"gorm.io/gorm"
)

type ScheduleModel struct {
	ID             int64 `gorm:"primaryKey"`
	EquipmentID    string
	Label          string
	NextDueDate    time.Time
	IntervalMonths *int
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

func (ScheduleModel) TableName() string { return "calibration_schedules" }

type ScheduleRepository struct {
	db *gorm.DB
}

func NewScheduleRepository(db *gorm.DB) *ScheduleRepository {
	return &ScheduleRepository{db: db}
}

func scheduleToDomain(m ScheduleModel) equipment.CalibrationSchedule {
	return equipment.CalibrationSchedule{
		ID:             m.ID,
		EquipmentID:    m.EquipmentID,
		Label:          m.Label,
		NextDueDate:    m.NextDueDate,
		IntervalMonths: m.IntervalMonths,
	}
}

func scheduleToModel(s equipment.CalibrationSchedule) ScheduleModel {
	return ScheduleModel{
		ID:             s.ID,
		EquipmentID:    s.EquipmentID,
		Label:          s.Label,
		NextDueDate:    s.NextDueDate,
		IntervalMonths: s.IntervalMonths,
	}
}

func (r *ScheduleRepository) Create(ctx context.Context, s equipment.CalibrationSchedule) (equipment.CalibrationSchedule, error) {
	m := scheduleToModel(s)
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return equipment.CalibrationSchedule{}, err
	}
	return scheduleToDomain(m), nil
}

func (r *ScheduleRepository) FindByID(ctx context.Context, id int64) (equipment.CalibrationSchedule, error) {
	var m ScheduleModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return equipment.CalibrationSchedule{}, shared.ErrNotFound
	}
	if err != nil {
		return equipment.CalibrationSchedule{}, err
	}
	return scheduleToDomain(m), nil
}

func (r *ScheduleRepository) ListByEquipment(ctx context.Context, equipmentID string) ([]equipment.CalibrationSchedule, error) {
	var models []ScheduleModel
	if err := r.db.WithContext(ctx).Where("equipment_id = ?", equipmentID).Order("next_due_date").Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]equipment.CalibrationSchedule, len(models))
	for i, m := range models {
		out[i] = scheduleToDomain(m)
	}
	return out, nil
}

// Update uses Save so a cleared IntervalMonths (nil) actually persists.
func (r *ScheduleRepository) Update(ctx context.Context, s equipment.CalibrationSchedule) (equipment.CalibrationSchedule, error) {
	m := scheduleToModel(s)
	if err := r.db.WithContext(ctx).Model(&ScheduleModel{}).Where("id = ?", s.ID).
		Select("label", "next_due_date", "interval_months").
		Updates(map[string]any{
			"label":           m.Label,
			"next_due_date":   m.NextDueDate,
			"interval_months": m.IntervalMonths,
		}).Error; err != nil {
		return equipment.CalibrationSchedule{}, err
	}
	return r.FindByID(ctx, s.ID)
}

func (r *ScheduleRepository) Delete(ctx context.Context, id int64) error {
	res := r.db.WithContext(ctx).Delete(&ScheduleModel{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return shared.ErrNotFound
	}
	return nil
}
