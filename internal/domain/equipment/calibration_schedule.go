package equipment

import "time"

// CalibrationSchedule is one recurring calibration commitment on a piece of
// Equipment - e.g. "สอบเทียบภายใน" due every 12 months. An Equipment may
// carry several at once, each tracked independently (CONTEXT.md
// "Calibration Schedule").
//
// When IntervalMonths is set, logging a CalibrationEvent whose
// CalibrationType matches Label auto-advances NextDueDate by that interval.
// When it is nil, the user sets the next NextDueDate by hand each time.
type CalibrationSchedule struct {
	ID             int64
	EquipmentID    string
	Label          string
	NextDueDate    time.Time
	IntervalMonths *int
}

// Advance returns NextDueDate moved forward by IntervalMonths, counted from
// the later of the current NextDueDate and calibratedAt so a late
// calibration does not leave the schedule permanently in the past. Returns
// (zero, false) when IntervalMonths is not set.
func (s CalibrationSchedule) Advance(calibratedAt time.Time) (time.Time, bool) {
	if s.IntervalMonths == nil || *s.IntervalMonths <= 0 {
		return time.Time{}, false
	}
	base := s.NextDueDate
	if calibratedAt.After(base) {
		base = calibratedAt
	}
	return base.AddDate(0, *s.IntervalMonths, 0), true
}
