package notification

import "time"

type Model struct {
	ID              string `gorm:"primaryKey"`
	RecipientUserID *int64
	Tone            string
	Icon            string
	Title           string
	Message         string
	CreatedAt       time.Time
	Read            bool
}

func (Model) TableName() string { return "notifications" }
