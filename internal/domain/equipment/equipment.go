package equipment

import "time"

type Status string

const (
	StatusReady   Status = "ready"
	StatusDueSoon Status = "due_soon"
	StatusOverdue Status = "overdue"
)

// dueSoonWindow is how far ahead of NextCalibrationDue a machine is flagged
// "due soon" rather than "ready".
const dueSoonWindow = 14 * 24 * time.Hour

type Equipment struct {
	ID                 string
	Name               string
	TypeCode           string // short code used in the ID, e.g. "CENT", "PCR"
	LastCalibratedAt   time.Time
	NextCalibrationDue time.Time
	UsageHours         int
}

// Status is derived from NextCalibrationDue rather than stored, so it can
// never drift out of sync with the actual dates.
func (e Equipment) DerivedStatus(now time.Time) Status {
	if now.After(e.NextCalibrationDue) {
		return StatusOverdue
	}
	if e.NextCalibrationDue.Sub(now) <= dueSoonWindow {
		return StatusDueSoon
	}
	return StatusReady
}

// CalibrationPct is a linear decay from 100% at LastCalibratedAt to 0% at
// NextCalibrationDue.
func (e Equipment) CalibrationPct(now time.Time) int {
	total := e.NextCalibrationDue.Sub(e.LastCalibratedAt)
	if total <= 0 {
		return 0
	}
	remaining := total - now.Sub(e.LastCalibratedAt)
	pct := int(remaining * 100 / total)
	switch {
	case pct > 100:
		return 100
	case pct < 0:
		return 0
	default:
		return pct
	}
}
