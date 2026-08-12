package user

import "time"

// RefreshToken is stored hashed so a leaked database dump can't be replayed
// as a valid refresh token; only the hash is ever persisted.
type RefreshToken struct {
	ID        int64
	UserID    int64
	TokenHash string
	ExpiresAt time.Time
	Revoked   bool
	CreatedAt time.Time
}
