package user

import (
	"context"

	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
)

// LogoutAllUseCase revokes every Session a user holds - "logout all
// devices" - as distinct from LogoutUseCase, which revokes only the
// Session presenting the given refresh token (see ADR 0004).
type LogoutAllUseCase struct {
	refresh portuser.RefreshTokenRepository
}

func NewLogoutAllUseCase(refresh portuser.RefreshTokenRepository) *LogoutAllUseCase {
	return &LogoutAllUseCase{refresh: refresh}
}

func (uc *LogoutAllUseCase) Execute(ctx context.Context, userID int64) error {
	return uc.refresh.RevokeAllForUser(ctx, userID)
}
