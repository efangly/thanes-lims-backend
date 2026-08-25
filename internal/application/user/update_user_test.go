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

func TestUpdateUserUseCase_LastAdminGuard(t *testing.T) {
	users := new(mockUserRepo)
	refresh := new(mockRefreshRepo)

	existing := domainuser.User{ID: 1, Name: "Admin One", Role: domainuser.RoleAdmin}
	users.On("FindByID", mock.Anything, int64(1)).Return(existing, nil)
	users.On("CountByRole", mock.Anything, domainuser.RoleAdmin).Return(int64(1), nil)

	uc := applicationuser.NewUpdateUserUseCase(users, refresh)
	_, err := uc.Execute(context.Background(), applicationuser.UpdateUserInput{
		ID: 1, Name: "Admin One", Role: domainuser.RoleGeneral,
	})

	assert.ErrorIs(t, err, shared.ErrValidation)
	refresh.AssertNotCalled(t, "RevokeAllForUser", mock.Anything, mock.Anything)
}

func TestUpdateUserUseCase_RoleChangeRevokesRefreshTokens(t *testing.T) {
	users := new(mockUserRepo)
	refresh := new(mockRefreshRepo)

	existing := domainuser.User{ID: 2, Name: "Somchai", Role: domainuser.RoleGeneral}
	updated := domainuser.User{ID: 2, Name: "Somchai", Role: domainuser.RoleScientist}
	users.On("FindByID", mock.Anything, int64(2)).Return(existing, nil)
	users.On("Update", mock.Anything, mock.MatchedBy(func(u domainuser.User) bool {
		return u.ID == 2 && u.Role == domainuser.RoleScientist
	})).Return(updated, nil)
	refresh.On("RevokeAllForUser", mock.Anything, int64(2)).Return(nil)

	uc := applicationuser.NewUpdateUserUseCase(users, refresh)
	result, err := uc.Execute(context.Background(), applicationuser.UpdateUserInput{
		ID: 2, Name: "Somchai", Role: domainuser.RoleScientist,
	})

	assert.NoError(t, err)
	assert.Equal(t, domainuser.RoleScientist, result.Role)
	refresh.AssertCalled(t, "RevokeAllForUser", mock.Anything, int64(2))
}

func TestUpdateUserUseCase_NoRoleChangeSkipsRevocation(t *testing.T) {
	users := new(mockUserRepo)
	refresh := new(mockRefreshRepo)

	existing := domainuser.User{ID: 3, Name: "Somchai", Role: domainuser.RoleGeneral}
	updated := domainuser.User{ID: 3, Name: "Somchai Updated", Role: domainuser.RoleGeneral}
	users.On("FindByID", mock.Anything, int64(3)).Return(existing, nil)
	users.On("Update", mock.Anything, mock.AnythingOfType("user.User")).Return(updated, nil)

	uc := applicationuser.NewUpdateUserUseCase(users, refresh)
	_, err := uc.Execute(context.Background(), applicationuser.UpdateUserInput{
		ID: 3, Name: "Somchai Updated", Role: domainuser.RoleGeneral,
	})

	assert.NoError(t, err)
	refresh.AssertNotCalled(t, "RevokeAllForUser", mock.Anything, mock.Anything)
}

func TestUpdateUserUseCase_InvalidRole(t *testing.T) {
	users := new(mockUserRepo)
	refresh := new(mockRefreshRepo)

	uc := applicationuser.NewUpdateUserUseCase(users, refresh)
	_, err := uc.Execute(context.Background(), applicationuser.UpdateUserInput{
		ID: 1, Name: "Bad", Role: domainuser.Role("bogus"),
	})

	assert.ErrorIs(t, err, shared.ErrValidation)
}
