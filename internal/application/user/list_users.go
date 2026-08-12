package user

import (
	"context"

	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
)

type ListUsersUseCase struct {
	users portuser.UserRepository
}

func NewListUsersUseCase(users portuser.UserRepository) *ListUsersUseCase {
	return &ListUsersUseCase{users: users}
}

func (uc *ListUsersUseCase) Execute(ctx context.Context) ([]domainuser.User, error) {
	return uc.users.List(ctx)
}
