package equipment

import (
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/equipment"
)

type CreateEquipmentRequest struct {
	Name               string     `json:"name" validate:"required"`
	TypeCode           string     `json:"type_code" validate:"required"`
	NextCalibrationDue time.Time  `json:"next_calibration_due" validate:"required"`
	SerialNumber       string     `json:"serial_number"`
	Category           string     `json:"category"`
	Manufacturer       string     `json:"manufacturer"`
	Model              string     `json:"model"`
	InstallationDate   *time.Time `json:"installation_date"`
	VendorID           string     `json:"vendor_id"`
	LocationID         string     `json:"location_id"`
}

type UpdateEquipmentRequest struct {
	Name                  *string    `json:"name"`
	TypeCode              *string    `json:"type_code"`
	SerialNumber          *string    `json:"serial_number"`
	Category              *string    `json:"category"`
	Manufacturer          *string    `json:"manufacturer"`
	Model                 *string    `json:"model"`
	InstallationDate      *time.Time `json:"installation_date"`
	ClearInstallationDate bool       `json:"clear_installation_date"`
	VendorID              *string    `json:"vendor_id"`
	LocationID            *string    `json:"location_id"`
}

type RecordCalibrationRequest struct {
	NextCalibrationDue time.Time `json:"next_calibration_due" validate:"required"`
	Notes              string    `json:"notes"`
	CalibrationType    string    `json:"calibration_type"`
	CalibrateValue     string    `json:"calibrate_value"`
	AcceptanceValue    string    `json:"acceptance_value"`
	Result             string    `json:"result" validate:"omitempty,oneof=pass fail"`
}

type CreateCalibrationScheduleRequest struct {
	Label          string    `json:"label" validate:"required"`
	NextDueDate    time.Time `json:"next_due_date" validate:"required"`
	IntervalMonths *int      `json:"interval_months" validate:"omitempty,gt=0"`
}

type UpdateCalibrationScheduleRequest struct {
	Label          *string    `json:"label"`
	NextDueDate    *time.Time `json:"next_due_date"`
	IntervalMonths *int       `json:"interval_months" validate:"omitempty,gt=0"`
	ClearInterval  bool       `json:"clear_interval"`
}

type CalibrationScheduleResponse struct {
	ID             int64     `json:"id"`
	EquipmentID    string    `json:"equipment_id"`
	Label          string    `json:"label"`
	NextDueDate    time.Time `json:"next_due_date"`
	IntervalMonths *int      `json:"interval_months"`
}

func toCalibrationScheduleResponse(s equipment.CalibrationSchedule) CalibrationScheduleResponse {
	return CalibrationScheduleResponse{
		ID:             s.ID,
		EquipmentID:    s.EquipmentID,
		Label:          s.Label,
		NextDueDate:    s.NextDueDate,
		IntervalMonths: s.IntervalMonths,
	}
}

type EquipmentResponse struct {
	ID                 string     `json:"id"`
	Name               string     `json:"name"`
	TypeCode           string     `json:"type_code"`
	LastCalibratedAt   time.Time  `json:"last_calibrated_at"`
	NextCalibrationDue time.Time  `json:"next_calibration_due"`
	UsageHours         int        `json:"usage_hours"`
	CalibrationPct     int        `json:"calibration_pct"`
	Status             string     `json:"status"`
	SerialNumber       string     `json:"serial_number"`
	Category           string     `json:"category"`
	Manufacturer       string     `json:"manufacturer"`
	Model              string     `json:"model"`
	InstallationDate   *time.Time `json:"installation_date"`
	VendorID           *string    `json:"vendor_id"`
	LocationID         *string    `json:"location_id"`
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
		SerialNumber:       e.SerialNumber,
		Category:           e.Category,
		Manufacturer:       e.Manufacturer,
		Model:              e.Model,
		InstallationDate:   e.InstallationDate,
		VendorID:           e.VendorID,
		LocationID:         e.LocationID,
	}
}

type CalibrationEventResponse struct {
	ID                 int64     `json:"id"`
	EquipmentID        string    `json:"equipment_id"`
	CalibratedAt       time.Time `json:"calibrated_at"`
	NextCalibrationDue time.Time `json:"next_calibration_due"`
	PerformedBy        string    `json:"performed_by"`
	Notes              string    `json:"notes"`
	CalibrationType    string    `json:"calibration_type"`
	CalibrateValue     string    `json:"calibrate_value"`
	AcceptanceValue    string    `json:"acceptance_value"`
	Result             string    `json:"result"`
}

func toCalibrationEventResponse(ev equipment.CalibrationEvent) CalibrationEventResponse {
	return CalibrationEventResponse{
		ID:                 ev.ID,
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
}
