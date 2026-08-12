package user

import "time"

type Model struct {
	ID           int64 `gorm:"primaryKey"`
	Name         string
	Email        string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (Model) TableName() string { return "users" }
