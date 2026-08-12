package user_test

import (
	"context"
	"testing"

	applicationuser "github.com/efangly/thanes-lims-backend/internal/application/user"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateUserUseCase_DuplicateEmail(t *testing.T) {
	users := new(mockUserRepo)
	users.On("FindByEmail", mock.Anything, "dup@b.com").Return(domainuser.User{ID: 1}, nil)

	uc := applicationuser.NewCreateUserUseCase(users)
	_, err := uc.Execute(context.Background(), applicationuser.CreateUserInput{
		Name: "Dup", Email: "dup@b.com", Password: "password123", Role: domainuser.RoleGeneral,
	})

	assert.ErrorIs(t, err, shared.ErrConflict)
}

func TestCreateUserUseCase_HashesPasswordAndCreates(t *testing.T) {
	users := new(mockUserRepo)
	users.On("FindByEmail", mock.Anything, "new@b.com").Return(domainuser.User{}, shared.ErrNotFound)
	users.On("Create", mock.Anything, mock.MatchedBy(func(u domainuser.User) bool {
		return u.Email == "new@b.com" && u.PasswordHash != "" && u.PasswordHash != "password123"
	})).Return(domainuser.User{ID: 5, Email: "new@b.com", Role: domainuser.RoleQA}, nil)

	uc := applicationuser.NewCreateUserUseCase(users)
	created, err := uc.Execute(context.Background(), applicationuser.CreateUserInput{
		Name: "New", Email: "new@b.com", Password: "password123", Role: domainuser.RoleQA,
	})

	assert.NoError(t, err)
	assert.Equal(t, int64(5), created.ID)
}

func TestCreateUserUseCase_InvalidRole(t *testing.T) {
	users := new(mockUserRepo)

	uc := applicationuser.NewCreateUserUseCase(users)
	_, err := uc.Execute(context.Background(), applicationuser.CreateUserInput{
		Name: "Bad", Email: "bad@b.com", Password: "password123", Role: domainuser.Role("bogus"),
	})

	assert.ErrorIs(t, err, shared.ErrValidation)
}
