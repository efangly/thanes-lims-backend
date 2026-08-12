package document

import "time"

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
