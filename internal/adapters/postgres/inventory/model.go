package inventory

import "gorm.io/gorm"

type Model struct {
	ID            string `gorm:"primaryKey"`
	Name          string
	Category      string
	Quantity      int
	Unit          string
	Min           int
	Max           int
	DefaultVendor string
	// DeletedAt makes Delete a soft delete (Retired, per ADR 0003); no
	// DeleteItem use case exists yet, but the column is in place.
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (Model) TableName() string { return "inventory_items" }
