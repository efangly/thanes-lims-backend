package user

import (
	"context"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
)

type UpdateUserUseCase struct {
	users portuser.UserRepository
}

func NewUpdateUserUseCase(users portuser.UserRepository) *UpdateUserUseCase {
	return &UpdateUserUseCase{users: users}
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

	existing.Name = in.Name
	existing.Role = in.Role
	existing.UpdatedAt = time.Now()

	return uc.users.Update(ctx, existing)
}
