package user

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/user"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, rt user.RefreshToken) (user.RefreshToken, error)
	FindByTokenHash(ctx context.Context, tokenHash string) (user.RefreshToken, error)
	// Revoke marks the row identified by id revoked and returns the number of
	// rows actually transitioned from not-revoked to revoked (0 or 1). A
	// return of 0 means the row was already revoked - for the refresh-rotation
	// path that is a Reuse Detection signal and must be treated as a leaked
	// token, not a no-op. tokenHash is the same row's hash, passed alongside
	// id so a caching decorator can invalidate its cache entry (keyed by hash)
	// without an extra lookup - callers always already have it from the
	// FindByTokenHash that preceded Revoke.
	Revoke(ctx context.Context, id int64, tokenHash string) (int64, error)
	RevokeAllForUser(ctx context.Context, userID int64) error
	// FindTokenHashesByUserID returns the hashes of every not-yet-revoked
	// Refresh Token row for userID, so a caching decorator's
	// RevokeAllForUser can invalidate each one's cache entry individually.
	FindTokenHashesByUserID(ctx context.Context, userID int64) ([]string, error)
}
