package equipment

import (
	"time"

	"gorm.io/gorm"
)

type Model struct {
	ID                 string `gorm:"primaryKey"`
	Name               string
	TypeCode           string
	LastCalibratedAt   time.Time
	NextCalibrationDue time.Time
	UsageHours         int

	SerialNumber     string
	Category         string
	Manufacturer     string
	Model            string
	InstallationDate *time.Time
	VendorID         *string
	LocationID       *string
	// DeletedAt makes Delete a soft delete (Retired, per ADR 0003); no
	// DeleteEquipment use case exists yet, but the column is in place.
	// CalibrationModel (calibration_events) is append-only and deliberately
	// excluded - its repository has no Delete.
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (Model) TableName() string { return "equipment" }
