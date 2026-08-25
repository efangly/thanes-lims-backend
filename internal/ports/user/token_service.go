package user

import (
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/user"
)

// Claims are the parsed contents of a validated access token.
type Claims struct {
	UserID int64
	Name   string
	Role   user.Role
	// Permissions is the Actor's resolved Permission set as compact
	// "module:action" strings, embedded at login/refresh (see ADR 0002).
	// Only populated on access tokens.
	Permissions []string
}

// TokenService issues and validates JWT access/refresh tokens. Access and
// refresh tokens use separate signing secrets so leaking one never
// compromises the other.
type TokenService interface {
	// GenerateAccessToken embeds permissions (compact "module:action"
	// strings) alongside the usual claims - see ADR 0002.
	GenerateAccessToken(u user.User, permissions []string) (string, error)
	// GenerateRefreshToken also returns the token's expiry so the caller can
	// persist it alongside the stored hash without duplicating TTL config.
	GenerateRefreshToken(u user.User) (token string, expiresAt time.Time, err error)
	ParseAccessToken(token string) (Claims, error)
	ParseRefreshToken(token string) (Claims, error)
	// HashRefreshToken produces the value that gets persisted via
	// RefreshTokenRepository, checked by equality against a presented raw
	// token's hash on refresh/logout.
	HashRefreshToken(token string) string
}
