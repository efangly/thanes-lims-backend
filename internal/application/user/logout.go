package user

import (
	"context"

	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
)

type LogoutUseCase struct {
	refresh portuser.RefreshTokenRepository
	tokens  portuser.TokenService
}

func NewLogoutUseCase(refresh portuser.RefreshTokenRepository, tokens portuser.TokenService) *LogoutUseCase {
	return &LogoutUseCase{refresh: refresh, tokens: tokens}
}

func (uc *LogoutUseCase) Execute(ctx context.Context, refreshTokenRaw string) error {
	hash := uc.tokens.HashRefreshToken(refreshTokenRaw)
	stored, err := uc.refresh.FindByTokenHash(ctx, hash)
	if err != nil {
		// Already gone/invalid - logout is idempotent, nothing to do.
		return nil
	}
	return uc.refresh.Revoke(ctx, stored.ID)
}
