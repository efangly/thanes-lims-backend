package user

import (
	"context"
	"strings"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
)

// UpdateProfileUseCase is the self-service edit: a User changes their own
// Name. Role, email and Status are untouched and remain admin-only (see
// CONTEXT.md "Self-service"). No Session is revoked - a stale name claim in
// an access token clears within one token TTL.
type UpdateProfileUseCase struct {
	users portuser.UserRepository
}

func NewUpdateProfileUseCase(users portuser.UserRepository) *UpdateProfileUseCase {
	return &UpdateProfileUseCase{users: users}
}

func (uc *UpdateProfileUseCase) Execute(ctx context.Context, userID int64, name string) (domainuser.User, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return domainuser.User{}, shared.ErrValidation
	}

	u, err := uc.users.FindByID(ctx, userID)
	if err != nil {
		return domainuser.User{}, err
	}
	u.Name = name
	u.UpdatedAt = time.Now()
	return uc.users.Update(ctx, u)
}
