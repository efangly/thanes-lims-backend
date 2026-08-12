package user

import (
	"context"
	"errors"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	"golang.org/x/crypto/bcrypt"
)

type CreateUserUseCase struct {
	users portuser.UserRepository
}

func NewCreateUserUseCase(users portuser.UserRepository) *CreateUserUseCase {
	return &CreateUserUseCase{users: users}
}

type CreateUserInput struct {
	Name     string
	Email    string
	Password string
	Role     domainuser.Role
}

func (uc *CreateUserUseCase) Execute(ctx context.Context, in CreateUserInput) (domainuser.User, error) {
	if !in.Role.Valid() {
		return domainuser.User{}, shared.ErrValidation
	}

	if _, err := uc.users.FindByEmail(ctx, in.Email); err == nil {
		return domainuser.User{}, shared.ErrConflict
	} else if !errors.Is(err, shared.ErrNotFound) {
		return domainuser.User{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return domainuser.User{}, err
	}

	now := time.Now()
	return uc.users.Create(ctx, domainuser.User{
		Name:         in.Name,
		Email:        in.Email,
		PasswordHash: string(hash),
		Role:         in.Role,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
}
