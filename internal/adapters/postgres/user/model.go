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
	RoleName  string `gorm:"->;column:role_name"`
	CreatedAt time.Time
	UpdatedAt time.Time
	// DeletedAt makes Delete a soft delete (Retired, per ADR 0003); no
	// DeleteUser use case exists yet, but the column and GORM's automatic
	// deleted_at IS NULL scoping are in place for when one is added.
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (Model) TableName() string { return "users" }
