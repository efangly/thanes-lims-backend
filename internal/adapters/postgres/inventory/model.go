package inventory

type Model struct {
	ID            string `gorm:"primaryKey"`
	Name          string
	Category      string
	Quantity      int
	Unit          string
	Min           int
	Max           int
	DefaultVendor string
}

func (Model) TableName() string { return "inventory_items" }
