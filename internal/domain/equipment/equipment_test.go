package equipment_test

import (
	"testing"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/equipment"
	"github.com/stretchr/testify/assert"
)

func TestEquipment_DerivedStatus(t *testing.T) {
	now := time.Now()

	cases := []struct {
		name string
		due  time.Time
		want equipment.Status
	}{
		{"far in future", now.Add(60 * 24 * time.Hour), equipment.StatusReady},
		{"exactly at due-soon boundary", now.Add(14 * 24 * time.Hour), equipment.StatusDueSoon},
		{"just inside due-soon window", now.Add(13 * 24 * time.Hour), equipment.StatusDueSoon},
		{"exactly now (past due)", now.Add(-time.Second), equipment.StatusOverdue},
		{"past due", now.Add(-24 * time.Hour), equipment.StatusOverdue},
	}

	for _, tc := range cases {
		e := equipment.Equipment{NextCalibrationDue: tc.due}
		assert.Equal(t, tc.want, e.DerivedStatus(now), tc.name)
	}
}

func TestEquipment_CalibrationPct(t *testing.T) {
	now := time.Now()
	e := equipment.Equipment{
		LastCalibratedAt:   now.Add(-30 * 24 * time.Hour),
		NextCalibrationDue: now.Add(30 * 24 * time.Hour),
	}
	assert.Equal(t, 50, e.CalibrationPct(now))

	overdue := equipment.Equipment{
		LastCalibratedAt:   now.Add(-60 * 24 * time.Hour),
		NextCalibrationDue: now.Add(-1 * time.Hour),
	}
	assert.Equal(t, 0, overdue.CalibrationPct(now))

	freshlyCalibrated := equipment.Equipment{
		LastCalibratedAt:   now,
		NextCalibrationDue: now.Add(30 * 24 * time.Hour),
	}
	assert.Equal(t, 100, freshlyCalibrated.CalibrationPct(now))
}
