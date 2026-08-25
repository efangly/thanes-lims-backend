package user

import "time"

// RefreshToken is stored hashed so a leaked database dump can't be replayed
// as a valid refresh token; only the hash is ever persisted.
//
// FamilyID identifies the Token Family (one continuous Session): every
// Rolling Refresh rotation creates a new row carrying the same FamilyID and
// FamilyCreatedAt as the row it replaces, so the chain can be revoked as a
// unit and checked against the Absolute Session Lifetime (see ADR 0004).
type RefreshToken struct {
	ID              int64
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
