package sample

import (
	"time"

	"gorm.io/gorm"
)

type Model struct {
	ID              string `gorm:"primaryKey"`
	Name            string
	Type            string
	CustodianUserID int64
	LocationID      *string
	Status          string
	ReceivedAt      time.Time
	BarcodeID       *string
	Description     string
	// DeletedAt makes Delete a soft delete (Retired, per ADR 0003) - GORM
	// automatically excludes non-NULL rows from Find/First and stamps this
	// instead of removing the row on Delete.
	DeletedAt gorm.DeletedAt `gorm:"index"`
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
