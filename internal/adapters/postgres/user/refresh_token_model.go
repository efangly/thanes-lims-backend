package user

import "time"

type RefreshTokenModel struct {
	ID        int64 `gorm:"primaryKey"`
	UserID    int64
	TokenHash string
	ExpiresAt time.Time
	Revoked   bool
	CreatedAt time.Time
}

func (RefreshTokenModel) TableName() string { return "refresh_tokens" }
