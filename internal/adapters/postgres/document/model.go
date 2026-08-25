package document

import (
	"time"

	"gorm.io/gorm"
)

type Model struct {
	ID          string `gorm:"primaryKey"`
	Name        string
	Type        string
	Version     string
	CreatedBy   string
	IssuedAt    time.Time
	AccessLevel string
	Locked      bool
	StorageKey  string
	// DeletedAt makes Delete a soft delete (Retired, per ADR 0003); no
	// DeleteDocument use case exists yet, but the column is in place.
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (Model) TableName() string { return "documents" }

type HistoryModel struct {
	ID         int64 `gorm:"primaryKey"`
	DocumentID string
	Version    string
	Change     string
	Date       time.Time
	Who        string
}

func (HistoryModel) TableName() string { return "doc_history" }
