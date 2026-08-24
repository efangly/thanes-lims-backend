package sample

import "time"

type Model struct {
	ID         string `gorm:"primaryKey"`
	Name       string
	Type       string
	Custodian  string
	LocationID *string
	Status     string
	ReceivedAt time.Time
}

func (Model) TableName() string { return "samples" }

type CoCModel struct {
	ID         int64 `gorm:"primaryKey"`
	SampleID   string
	State      string
	Icon       string
	Title      string
	Meta       string
	Who        string
	OccurredAt time.Time
}

func (CoCModel) TableName() string { return "coc_steps" }
