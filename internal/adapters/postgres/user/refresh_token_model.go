package user

import "time"

type RefreshTokenModel struct {
	ID              int64 `gorm:"primaryKey"`
	UserID          int64
	FamilyID        string
	FamilyCreatedAt time.Time
	TokenHash       string
	ExpiresAt       time.Time
	Revoked         bool
	CreatedAt       time.Time
	UserAgent       string
	IPAddress       string
}

func (RefreshTokenModel) TableName() string { return "refresh_tokens" }
