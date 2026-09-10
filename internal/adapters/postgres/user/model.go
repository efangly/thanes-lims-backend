package user

import (
	"time"

	"gorm.io/gorm"
)

type Model struct {
	ID           int64 `gorm:"primaryKey"`
	Name         string
	Email        string
	PasswordHash string
	RoleID       int64
	// RoleName is populated only by the joined queries in repository.go
	// (roles.name via role_id) - "->" marks it read-only so plain
	// Create/Update never try to write a non-existent role_name column.
	RoleName string `gorm:"->;column:role_name"`
	// Status is the reversible active/suspended flag (see ADR 0010),
	// distinct from DeletedAt below (Retired, irreversible).
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
	// DeletedAt makes Delete a soft delete (Retired, per ADR 0003). GORM's
	// automatic deleted_at IS NULL scoping keeps Retired Users out of every
	// query that goes through this Model.
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (Model) TableName() string { return "users" }
