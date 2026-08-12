package equipment

import "time"

type Model struct {
	ID                 string `gorm:"primaryKey"`
	Name               string
	TypeCode           string
	LastCalibratedAt   time.Time
	NextCalibrationDue time.Time
	UsageHours         int
}

func (Model) TableName() string { return "equipment" }
