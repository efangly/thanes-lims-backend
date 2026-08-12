package audit

import (
	"time"

	"gorm.io/datatypes"
)

type Model struct {
	ID         int64 `gorm:"primaryKey"`
	ActorID    *int64
	ActorRole  string
	Method     string
	Path       string
	Resource   string
	ResourceID string
	StatusCode int
	Metadata   datatypes.JSONMap
	CreatedAt  time.Time
}

func (Model) TableName() string { return "audit_logs" }
