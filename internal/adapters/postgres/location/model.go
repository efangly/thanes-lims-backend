package location

import (
	"time"

	"gorm.io/gorm"
)

type Model struct {
	ID          string `gorm:"primaryKey"`
	ParentID    *string
	Name        string
	Kind        string `gorm:"default:sample_storage"`
	LevelType   string
	BarcodeCode *string
	// Rows/Cols hold a Box's Grid; NULL for every non-Box row (docs/adr/0009).
	Rows      *int16
	Cols      *int16
	CreatedAt time.Time
	// DeletedAt makes Delete a soft delete (Retired, per ADR 0003).
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (Model) TableName() string { return "locations" }
