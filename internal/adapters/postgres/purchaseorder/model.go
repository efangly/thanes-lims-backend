package purchaseorder

import (
	"time"

	"gorm.io/gorm"
)

type Model struct {
	ID        string `gorm:"primaryKey"`
	ItemID    string
	Quantity  int
	Vendor    string
	OrderDate time.Time
	Status    string
	// DeletedAt makes Delete a soft delete (Retired, per ADR 0003); no
	// DeletePO use case exists yet, but the column is in place.
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (Model) TableName() string { return "purchase_orders" }
