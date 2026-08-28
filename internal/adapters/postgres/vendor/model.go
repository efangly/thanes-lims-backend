package vendor

import (
	"time"

	"gorm.io/gorm"
)

type Model struct {
	ID           string `gorm:"primaryKey"`
	Name         string
	ContactName  string
	ContactPhone string
	ContactEmail string
	Address      string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	// DeletedAt makes Delete a soft delete (Retired, per ADR 0003); no
	// DeleteVendor use case exists yet, but the column and the partial
	// unique index on name are already in place.
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (Model) TableName() string { return "vendors" }
