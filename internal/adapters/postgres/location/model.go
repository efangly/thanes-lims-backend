package location

import "time"

type Model struct {
	ID        string `gorm:"primaryKey"`
	ParentID  *string
	Name      string
	LevelType string
	CreatedAt time.Time
}

func (Model) TableName() string { return "locations" }
