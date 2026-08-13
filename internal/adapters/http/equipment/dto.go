package equipment

import (
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/equipment"
)

type CreateEquipmentRequest struct {
	Name               string    `json:"name" validate:"required"`
	TypeCode           string    `json:"type_code" validate:"required"`
	NextCalibrationDue time.Time `json:"next_calibration_due" validate:"required"`
}

type RecordCalibrationRequest struct {
	NextCalibrationDue time.Time `json:"next_calibration_due" validate:"required"`
	Notes              string    `json:"notes"`
}

type EquipmentResponse struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	TypeCode           string    `json:"type_code"`
	LastCalibratedAt   time.Time `json:"last_calibrated_at"`
	NextCalibrationDue time.Time `json:"next_calibration_due"`
	UsageHours         int       `json:"usage_hours"`
	CalibrationPct     int       `json:"calibration_pct"`
	Status             string    `json:"status"`
}

func toResponse(e equipment.Equipment) EquipmentResponse {
	now := time.Now()
	return EquipmentResponse{
		ID:                 e.ID,
		Name:               e.Name,
		TypeCode:           e.TypeCode,
		LastCalibratedAt:   e.LastCalibratedAt,
		NextCalibrationDue: e.NextCalibrationDue,
		UsageHours:         e.UsageHours,
		CalibrationPct:     e.CalibrationPct(now),
		Status:             string(e.DerivedStatus(now)),
	}
}

type CalibrationEventResponse struct {
	ID                 int64     `json:"id"`
	EquipmentID        string    `json:"equipment_id"`
	CalibratedAt       time.Time `json:"calibrated_at"`
	NextCalibrationDue time.Time `json:"next_calibration_due"`
	PerformedBy        string    `json:"performed_by"`
	Notes              string    `json:"notes"`
}

func toCalibrationEventResponse(ev equipment.CalibrationEvent) CalibrationEventResponse {
	return CalibrationEventResponse{
		ID:                 ev.ID,
		EquipmentID:        ev.EquipmentID,
		CalibratedAt:       ev.CalibratedAt,
		NextCalibrationDue: ev.NextCalibrationDue,
		PerformedBy:        ev.PerformedBy,
		Notes:              ev.Notes,
	}
}
