package user

import (
	"context"
	"time"

	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
)

// SuspendUserUseCase flips a User to suspended and revokes every Session
// they hold, so the lock-out is immediate (see ADR 0010). ReactivateUser is
// the reverse and shares this type.
type SuspendUserUseCase struct {
	users   portuser.UserRepository
	refresh portuser.RefreshTokenRepository
}

func NewSuspendUserUseCase(users portuser.UserRepository, refresh portuser.RefreshTokenRepository) *SuspendUserUseCase {
	return &SuspendUserUseCase{users: users, refresh: refresh}
}

// Execute suspends targetID. actorID is the admin making the call - they may
// not suspend themselves, nor the last active admin.
func (uc *SuspendUserUseCase) Execute(ctx context.Context, actorID, targetID int64) (domainuser.User, error) {
	if actorID == targetID {
		return domainuser.User{}, errSelfTarget("suspend")
	}

	target, err := uc.users.FindByID(ctx, targetID)
	if err != nil {
		return domainuser.User{}, err
	}
	if target.Status == domainuser.StatusSuspended {
		return target, nil
	}
	if err := guardNotLastActiveAdmin(ctx, uc.users, target, "suspend"); err != nil {
		return domainuser.User{}, err
	}

	target.Status = domainuser.StatusSuspended
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
