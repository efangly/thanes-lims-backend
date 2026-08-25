package user

import (
	"context"
	"fmt"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
)

type UpdateUserUseCase struct {
	users   portuser.UserRepository
	refresh portuser.RefreshTokenRepository
}

func NewUpdateUserUseCase(users portuser.UserRepository, refresh portuser.RefreshTokenRepository) *UpdateUserUseCase {
	return &UpdateUserUseCase{users: users, refresh: refresh}
}

type UpdateUserInput struct {
	ID   int64
	Name string
	Role domainuser.Role
}

func (uc *UpdateUserUseCase) Execute(ctx context.Context, in UpdateUserInput) (domainuser.User, error) {
	if !in.Role.Valid() {
		return domainuser.User{}, shared.ErrValidation
	}

	existing, err := uc.users.FindByID(ctx, in.ID)
	if err != nil {
		return domainuser.User{}, err
	}

	roleChanged := existing.Role != in.Role

	// Last-admin guard: block a role change that would leave zero Users
	// with the Admin role.
	if roleChanged && existing.Role == domainuser.RoleAdmin {
		adminCount, err := uc.users.CountByRole(ctx, domainuser.RoleAdmin)
		if err != nil {
			return domainuser.User{}, err
		}
		if adminCount <= 1 {
			return domainuser.User{}, fmt.Errorf("%w: cannot change role - this is the last remaining admin", shared.ErrValidation)
		}
	}

	existing.Name = in.Name
	existing.Role = in.Role
	existing.UpdatedAt = time.Now()

	updated, err := uc.users.Update(ctx, existing)
	if err != nil {
		return domainuser.User{}, err
	}

	// A role change invalidates any long-lived refresh tokens issued under
	// the old role, so the User must log in again to pick up the new
	// Permission set. Access tokens still expire naturally (see ADR 0002).
	if roleChanged {
		if err := uc.refresh.RevokeAllForUser(ctx, in.ID); err != nil {
			return domainuser.User{}, err
		}
	}

	return updated, nil
}
