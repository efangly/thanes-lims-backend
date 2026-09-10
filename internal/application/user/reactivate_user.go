package user

import (
	"context"
	"time"

	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
)

// ReactivateUserUseCase flips a suspended User back to active. It does not
// restore any Session - the User simply logs in again.
type ReactivateUserUseCase struct {
	users portuser.UserRepository
}

func NewReactivateUserUseCase(users portuser.UserRepository) *ReactivateUserUseCase {
	return &ReactivateUserUseCase{users: users}
}

func (uc *ReactivateUserUseCase) Execute(ctx context.Context, targetID int64) (domainuser.User, error) {
	target, err := uc.users.FindByID(ctx, targetID)
	if err != nil {
		return domainuser.User{}, err
	}
	if target.Status == domainuser.StatusActive {
		return target, nil
	}
	target.Status = domainuser.StatusActive
	target.UpdatedAt = time.Now()
	return uc.users.Update(ctx, target)
}
