package purchaseorder

import "time"

type Model struct {
	ID        string `gorm:"primaryKey"`
	ItemID    string
	Quantity  int
	Vendor    string
	OrderDate time.Time
	Status    string
}

func (Model) TableName() string { return "purchase_orders" }
