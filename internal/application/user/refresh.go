package user

import (
	"context"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
)

type RefreshUseCase struct {
	users   portuser.UserRepository
	refresh portuser.RefreshTokenRepository
	tokens  portuser.TokenService
}

func NewRefreshUseCase(users portuser.UserRepository, refresh portuser.RefreshTokenRepository, tokens portuser.TokenService) *RefreshUseCase {
	return &RefreshUseCase{users: users, refresh: refresh, tokens: tokens}
}

// Execute rotates the refresh token: the presented token is revoked and a
// brand new access/refresh pair is issued, so a stolen-then-reused refresh
// token is immediately detectable (it'll already be revoked).
func (uc *RefreshUseCase) Execute(ctx context.Context, refreshTokenRaw string) (TokenPair, error) {
	if _, err := uc.tokens.ParseRefreshToken(refreshTokenRaw); err != nil {
		return TokenPair{}, shared.ErrUnauthorized
	}

	hash := uc.tokens.HashRefreshToken(refreshTokenRaw)
	stored, err := uc.refresh.FindByTokenHash(ctx, hash)
	if err != nil {
		return TokenPair{}, shared.ErrUnauthorized
	}
	if stored.Revoked || stored.ExpiresAt.Before(time.Now()) {
		return TokenPair{}, shared.ErrUnauthorized
	}

	u, err := uc.users.FindByID(ctx, stored.UserID)
	if err != nil {
		return TokenPair{}, shared.ErrUnauthorized
	}

	if err := uc.refresh.Revoke(ctx, stored.ID); err != nil {
		return TokenPair{}, err
	}

	return issueTokenPair(ctx, uc.tokens, uc.refresh, u)
}
