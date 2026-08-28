package equipment

import "time"

// CalibrationResult is the pass/fail verdict of a calibration.
type CalibrationResult string

const (
	CalibrationResultPass CalibrationResult = "pass"
	CalibrationResultFail CalibrationResult = "fail"
)

func (r CalibrationResult) Valid() bool {
	switch r {
	case "", CalibrationResultPass, CalibrationResultFail:
		return true
	default:
		return false
	}
}

// CalibrationEvent is an append-only audit trail entry for a single
// calibration performed on a piece of Equipment.
type CalibrationEvent struct {
	ID                 int64
	EquipmentID        string
	CalibratedAt       time.Time
	NextCalibrationDue time.Time
	PerformedBy        string
	Notes              string

	// Phase 6 measurement fields. CalibrationType is free text and, when it
	// matches a CalibrationSchedule's Label, drives that schedule's
	// auto-advance. CalibrateValue / AcceptanceValue are kept as free text
	// so units and tolerances ("±0.1 g", "< 2 %") can be recorded verbatim.
	CalibrationType string
	CalibrateValue  string
	AcceptanceValue string
	Result          CalibrationResult
}
