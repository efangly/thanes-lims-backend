package user

import (
	"context"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portrbac "github.com/efangly/thanes-lims-backend/internal/ports/rbac"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// ChangePasswordUseCase is the self-service password change: the caller
// proves the current password, sets a new one, and every OTHER Session is
// revoked. The caller's own Session is kept alive by issuing a fresh token
// pair (a new Token Family) in the same call - so a password change does not
// log you out of the device you changed it on.
type ChangePasswordUseCase struct {
	users   portuser.UserRepository
	refresh portuser.RefreshTokenRepository
	tokens  portuser.TokenService
	rbac    portrbac.Repository
}

func NewChangePasswordUseCase(users portuser.UserRepository, refresh portuser.RefreshTokenRepository, tokens portuser.TokenService, rbacRepo portrbac.Repository) *ChangePasswordUseCase {
	return &ChangePasswordUseCase{users: users, refresh: refresh, tokens: tokens, rbac: rbacRepo}
}

func (uc *ChangePasswordUseCase) Execute(ctx context.Context, userID int64, currentPassword, newPassword, userAgent, ipAddress string) (TokenPair, error) {
	if len(newPassword) < minPasswordLen {
		return TokenPair{}, shared.ErrValidation
	}

	u, err := uc.users.FindByID(ctx, userID)
	if err != nil {
		return TokenPair{}, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(currentPassword)); err != nil {
		return TokenPair{}, shared.ErrUnauthorized
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return TokenPair{}, err
	}
	u.PasswordHash = string(hash)
	u.UpdatedAt = time.Now()
	if _, err := uc.users.Update(ctx, u); err != nil {
		return TokenPair{}, err
	}

	// Revoke every existing Session, then mint a fresh one for this caller.
	if err := uc.refresh.RevokeAllForUser(ctx, userID); err != nil {
		return TokenPair{}, err
	}

	perms, err := uc.rbac.FindPermissionsByRoleName(ctx, u.Role.DisplayName())
	if err != nil {
		return TokenPair{}, err
	}
	family := refreshFamily{ID: uuid.NewString(), CreatedAt: time.Now(), UserAgent: userAgent, IPAddress: ipAddress}
	return issueTokenPair(ctx, uc.tokens, uc.refresh, u, permissionKeys(perms), family)
}
