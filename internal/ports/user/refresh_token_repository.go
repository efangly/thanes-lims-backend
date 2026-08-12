package user

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/user"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, rt user.RefreshToken) (user.RefreshToken, error)
	FindByTokenHash(ctx context.Context, tokenHash string) (user.RefreshToken, error)
	Revoke(ctx context.Context, id int64) error
	RevokeAllForUser(ctx context.Context, userID int64) error
}
