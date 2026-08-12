package user

import (
	"context"

	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
)

type GetUserUseCase struct {
	users portuser.UserRepository
}

func NewGetUserUseCase(users portuser.UserRepository) *GetUserUseCase {
	return &GetUserUseCase{users: users}
}

func (uc *GetUserUseCase) Execute(ctx context.Context, id int64) (domainuser.User, error) {
	return uc.users.FindByID(ctx, id)
}
