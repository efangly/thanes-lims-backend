package equipment

import "time"

// CalibrationEvent is an append-only audit trail entry for a single
// calibration performed on a piece of Equipment.
type CalibrationEvent struct {
	ID                 int64
	EquipmentID        string
	CalibratedAt       time.Time
	NextCalibrationDue time.Time
	PerformedBy        string
	Notes              string
}
