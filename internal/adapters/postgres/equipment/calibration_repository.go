package equipment

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/equipment"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portequipment "github.com/efangly/thanes-lims-backend/internal/ports/equipment"
	"gorm.io/gorm"
)

type CalibrationModel struct {
	ID                 int64 `gorm:"primaryKey"`
	EquipmentID        string
	CalibratedAt       time.Time
	NextCalibrationDue time.Time
	PerformedBy        string
	Notes              string
	CalibrationType    string
	CalibrateValue     string
	AcceptanceValue    string
	Result             string
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
		CalibrationType:    m.CalibrationType,
		CalibrateValue:     m.CalibrateValue,
		AcceptanceValue:    m.AcceptanceValue,
		Result:             equipment.CalibrationResult(m.Result),
	}
}

func (r *CalibrationRepository) Append(ctx context.Context, ev equipment.CalibrationEvent) (equipment.CalibrationEvent, error) {
	m := CalibrationModel{
		EquipmentID:        ev.EquipmentID,
		CalibratedAt:       ev.CalibratedAt,
		NextCalibrationDue: ev.NextCalibrationDue,
		PerformedBy:        ev.PerformedBy,
		Notes:              ev.Notes,
		CalibrationType:    ev.CalibrationType,
		CalibrateValue:     ev.CalibrateValue,
		AcceptanceValue:    ev.AcceptanceValue,
		Result:             string(ev.Result),
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

func (r *CalibrationRepository) FindByID(ctx context.Context, id int64) (equipment.CalibrationEvent, error) {
	var m CalibrationModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return equipment.CalibrationEvent{}, shared.ErrNotFound
	}
	if err != nil {
		return equipment.CalibrationEvent{}, err
	}
	return calibrationToDomain(m), nil
}

func (r *CalibrationRepository) Search(ctx context.Context, f portequipment.CalibrationSearchFilter) ([]equipment.CalibrationEvent, error) {
	q := r.db.WithContext(ctx).Model(&CalibrationModel{}).
		Joins("LEFT JOIN equipment ON equipment.id = calibration_events.equipment_id")

	if s := strings.TrimSpace(f.Query); s != "" {
		like := "%" + s + "%"
		q = q.Where(
			r.db.Where("calibration_events.equipment_id ILIKE ?", like).
				Or("equipment.name ILIKE ?", like).
				Or("calibration_events.performed_by ILIKE ?", like).
				Or("calibration_events.calibration_type ILIKE ?", like).
				Or("calibration_events.notes ILIKE ?", like),
		)
	}
	if f.EquipmentID != "" {
		q = q.Where("calibration_events.equipment_id = ?", f.EquipmentID)
	}
	if f.Result != "" {
		q = q.Where("calibration_events.result = ?", string(f.Result))
	}
	if f.From != nil {
		q = q.Where("calibration_events.calibrated_at >= ?", *f.From)
	}
	if f.To != nil {
		q = q.Where("calibration_events.calibrated_at <= ?", *f.To)
	}

	var models []CalibrationModel
	if err := q.Order("calibration_events.calibrated_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]equipment.CalibrationEvent, len(models))
	for i, m := range models {
		out[i] = calibrationToDomain(m)
	}
	return out, nil
}
