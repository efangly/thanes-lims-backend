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

func activeUser(id int64, role domainuser.Role) domainuser.User {
	return domainuser.User{ID: id, Name: "U", Email: "u@x.io", Role: role, Status: domainuser.StatusActive}
}

func TestSuspendUser_RejectsSelf(t *testing.T) {
	uc := applicationuser.NewSuspendUserUseCase(new(mockUserRepo), new(mockRefreshRepo))
	_, err := uc.Execute(context.Background(), 7, 7)
	assert.ErrorIs(t, err, shared.ErrValidation)
}

func TestSuspendUser_RejectsLastActiveAdmin(t *testing.T) {
	users := new(mockUserRepo)
	users.On("FindByID", mock.Anything, int64(1)).Return(activeUser(1, domainuser.RoleAdmin), nil)
	users.On("CountActiveByRole", mock.Anything, domainuser.RoleAdmin).Return(int64(1), nil)

	uc := applicationuser.NewSuspendUserUseCase(users, new(mockRefreshRepo))
	_, err := uc.Execute(context.Background(), 99, 1)
	assert.ErrorIs(t, err, shared.ErrValidation)
}

func TestSuspendUser_RevokesSessions(t *testing.T) {
	users := new(mockUserRepo)
	refresh := new(mockRefreshRepo)
	target := activeUser(2, domainuser.RoleScientist)
	users.On("FindByID", mock.Anything, int64(2)).Return(target, nil)
	users.On("Update", mock.Anything, mock.MatchedBy(func(u domainuser.User) bool {
		return u.ID == 2 && u.Status == domainuser.StatusSuspended
	})).Return(func() domainuser.User { t := target; t.Status = domainuser.StatusSuspended; return t }(), nil)
	refresh.On("RevokeAllForUser", mock.Anything, int64(2)).Return(nil)

	uc := applicationuser.NewSuspendUserUseCase(users, refresh)
	got, err := uc.Execute(context.Background(), 1, 2)

	assert.NoError(t, err)
	assert.Equal(t, domainuser.StatusSuspended, got.Status)
	refresh.AssertCalled(t, "RevokeAllForUser", mock.Anything, int64(2))
}

func TestRetireUser_BlockedWhileCustodian(t *testing.T) {
	users := new(mockUserRepo)
	custodian := new(mockCustodianChecker)
	users.On("FindByID", mock.Anything, int64(3)).Return(activeUser(3, domainuser.RoleScientist), nil)
	custodian.On("CountCustodianRefs", mock.Anything, int64(3)).Return(int64(2), int64(0), nil)

	uc := applicationuser.NewRetireUserUseCase(users, new(mockRefreshRepo), custodian)
	err := uc.Execute(context.Background(), 1, 3)

	assert.ErrorIs(t, err, shared.ErrConflict)
	users.AssertNotCalled(t, "Retire", mock.Anything, mock.Anything)
}

func TestRetireUser_HappyPathRevokesSessions(t *testing.T) {
	users := new(mockUserRepo)
	refresh := new(mockRefreshRepo)
	custodian := new(mockCustodianChecker)
	users.On("FindByID", mock.Anything, int64(4)).Return(activeUser(4, domainuser.RoleGeneral), nil)
	custodian.On("CountCustodianRefs", mock.Anything, int64(4)).Return(int64(0), int64(0), nil)
	users.On("Retire", mock.Anything, int64(4)).Return(nil)
	refresh.On("RevokeAllForUser", mock.Anything, int64(4)).Return(nil)

	uc := applicationuser.NewRetireUserUseCase(users, refresh, custodian)
	assert.NoError(t, uc.Execute(context.Background(), 1, 4))
	refresh.AssertCalled(t, "RevokeAllForUser", mock.Anything, int64(4))
}

func TestResetPassword_RejectsShort(t *testing.T) {
	uc := applicationuser.NewResetPasswordUseCase(new(mockUserRepo), new(mockRefreshRepo))
	_, err := uc.Execute(context.Background(), 1, "short")
	assert.ErrorIs(t, err, shared.ErrValidation)
}
