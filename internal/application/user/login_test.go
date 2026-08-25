package user_test

import (
	"context"
	"testing"
	"time"

	applicationuser "github.com/efangly/thanes-lims-backend/internal/application/user"
	"github.com/efangly/thanes-lims-backend/internal/domain/rbac"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func hashPassword(t *testing.T, pw string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	assert.NoError(t, err)
	return string(hash)
}

func TestLoginUseCase_Success(t *testing.T) {
	users := new(mockUserRepo)
	refresh := new(mockRefreshRepo)
	tokens := new(mockTokenService)
	rbacRepo := new(mockRBACRepo)

	u := domainuser.User{ID: 1, Email: "a@b.com", PasswordHash: hashPassword(t, "correct-password"), Role: domainuser.RoleAdmin}
	perms := []rbac.Permission{{Module: rbac.ModuleUser, Action: rbac.ActionView}}

	users.On("FindByEmail", mock.Anything, "a@b.com").Return(u, nil)
	rbacRepo.On("FindPermissionsByRoleName", mock.Anything, "Admin").Return(perms, nil)
	tokens.On("GenerateAccessToken", u, []string{"user:view"}).Return("access-token", nil)
	tokens.On("GenerateRefreshToken", u).Return("refresh-token", time.Now().Add(time.Hour), nil)
	tokens.On("HashRefreshToken", "refresh-token").Return("hashed")
	refresh.On("Create", mock.Anything, mock.AnythingOfType("user.RefreshToken")).Return(domainuser.RefreshToken{}, nil)

	uc := applicationuser.NewLoginUseCase(users, refresh, tokens, rbacRepo)
	pair, err := uc.Execute(context.Background(), "a@b.com", "correct-password", "test-agent", "127.0.0.1")

	assert.NoError(t, err)
	assert.Equal(t, "access-token", pair.AccessToken)
	assert.Equal(t, "refresh-token", pair.RefreshToken)
}

func TestLoginUseCase_WrongPassword(t *testing.T) {
	users := new(mockUserRepo)
	refresh := new(mockRefreshRepo)
	tokens := new(mockTokenService)
	rbacRepo := new(mockRBACRepo)

	u := domainuser.User{ID: 1, Email: "a@b.com", PasswordHash: hashPassword(t, "correct-password"), Role: domainuser.RoleAdmin}
	users.On("FindByEmail", mock.Anything, "a@b.com").Return(u, nil)

	uc := applicationuser.NewLoginUseCase(users, refresh, tokens, rbacRepo)
	_, err := uc.Execute(context.Background(), "a@b.com", "wrong-password", "test-agent", "127.0.0.1")

	assert.ErrorIs(t, err, shared.ErrUnauthorized)
}

func TestLoginUseCase_UserNotFound(t *testing.T) {
	users := new(mockUserRepo)
	refresh := new(mockRefreshRepo)
	tokens := new(mockTokenService)
	rbacRepo := new(mockRBACRepo)

	users.On("FindByEmail", mock.Anything, "missing@b.com").Return(domainuser.User{}, shared.ErrNotFound)

	uc := applicationuser.NewLoginUseCase(users, refresh, tokens, rbacRepo)
	_, err := uc.Execute(context.Background(), "missing@b.com", "whatever", "test-agent", "127.0.0.1")

	assert.ErrorIs(t, err, shared.ErrUnauthorized)
}
