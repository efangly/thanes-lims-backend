package user

import (
	"context"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	"golang.org/x/crypto/bcrypt"
)

const minPasswordLen = 8

// ResetPasswordUseCase lets an admin set a new password for another User
// (there is no email infrastructure, so no reset link). Every Session the
// target holds is revoked - they must log in again with the new password.
type ResetPasswordUseCase struct {
	users   portuser.UserRepository
	refresh portuser.RefreshTokenRepository
}

func NewResetPasswordUseCase(users portuser.UserRepository, refresh portuser.RefreshTokenRepository) *ResetPasswordUseCase {
	return &ResetPasswordUseCase{users: users, refresh: refresh}
}

func (uc *ResetPasswordUseCase) Execute(ctx context.Context, targetID int64, newPassword string) (domainuser.User, error) {
	if len(newPassword) < minPasswordLen {
		return domainuser.User{}, shared.ErrValidation
	}

	target, err := uc.users.FindByID(ctx, targetID)
	if err != nil {
		return domainuser.User{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return domainuser.User{}, err
	}

	target.PasswordHash = string(hash)
	target.UpdatedAt = time.Now()
	updated, err := uc.users.Update(ctx, target)
	if err != nil {
		return domainuser.User{}, err
	}

	if err := uc.refresh.RevokeAllForUser(ctx, targetID); err != nil {
		return domainuser.User{}, err
	}
	return updated, nil
}
