package equipment

import (
	"context"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/equipment"
	"gorm.io/gorm"
)

type CalibrationModel struct {
	ID                 int64 `gorm:"primaryKey"`
	EquipmentID        string
	CalibratedAt       time.Time
	NextCalibrationDue time.Time
	PerformedBy        string
	Notes              string
}

func (CalibrationModel) TableName() string { return "calibration_events" }

type CalibrationRepository struct {
	db *gorm.DB
}

func NewCalibrationRepository(db *gorm.DB) *CalibrationRepository {
	return &CalibrationRepository{db: db}
}

func calibrationToDomain(m CalibrationModel) equipment.CalibrationEvent {
	return equipment.CalibrationEvent{
		ID:                 m.ID,
		EquipmentID:        m.EquipmentID,
		CalibratedAt:       m.CalibratedAt,
		NextCalibrationDue: m.NextCalibrationDue,
		PerformedBy:        m.PerformedBy,
		Notes:              m.Notes,
	}
}

func (r *CalibrationRepository) Append(ctx context.Context, ev equipment.CalibrationEvent) (equipment.CalibrationEvent, error) {
	m := CalibrationModel{
		EquipmentID:        ev.EquipmentID,
		CalibratedAt:       ev.CalibratedAt,
		NextCalibrationDue: ev.NextCalibrationDue,
		PerformedBy:        ev.PerformedBy,
		Notes:              ev.Notes,
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return equipment.CalibrationEvent{}, err
	}
	return calibrationToDomain(m), nil
}

func (r *CalibrationRepository) ListByEquipment(ctx context.Context, equipmentID string) ([]equipment.CalibrationEvent, error) {
	var models []CalibrationModel
	if err := r.db.WithContext(ctx).Where("equipment_id = ?", equipmentID).Order("calibrated_at").Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]equipment.CalibrationEvent, len(models))
	for i, m := range models {
		out[i] = calibrationToDomain(m)
	}
	return out, nil
}
