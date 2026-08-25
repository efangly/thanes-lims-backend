package user_test

import (
	"context"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/rbac"
	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	"github.com/stretchr/testify/mock"
)

type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) Create(ctx context.Context, u domainuser.User) (domainuser.User, error) {
	args := m.Called(ctx, u)
	return args.Get(0).(domainuser.User), args.Error(1)
}
func (m *mockUserRepo) FindByID(ctx context.Context, id int64) (domainuser.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domainuser.User), args.Error(1)
}
func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (domainuser.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(domainuser.User), args.Error(1)
}
func (m *mockUserRepo) List(ctx context.Context) ([]domainuser.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domainuser.User), args.Error(1)
}
func (m *mockUserRepo) Update(ctx context.Context, u domainuser.User) (domainuser.User, error) {
	args := m.Called(ctx, u)
	return args.Get(0).(domainuser.User), args.Error(1)
}
func (m *mockUserRepo) CountByRole(ctx context.Context, role domainuser.Role) (int64, error) {
	args := m.Called(ctx, role)
	return args.Get(0).(int64), args.Error(1)
}

type mockRefreshRepo struct{ mock.Mock }

func (m *mockRefreshRepo) Create(ctx context.Context, rt domainuser.RefreshToken) (domainuser.RefreshToken, error) {
	args := m.Called(ctx, rt)
	return args.Get(0).(domainuser.RefreshToken), args.Error(1)
}
func (m *mockRefreshRepo) FindByTokenHash(ctx context.Context, hash string) (domainuser.RefreshToken, error) {
	args := m.Called(ctx, hash)
	return args.Get(0).(domainuser.RefreshToken), args.Error(1)
}
func (m *mockRefreshRepo) Revoke(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockRefreshRepo) RevokeAllForUser(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

type mockTokenService struct{ mock.Mock }

func (m *mockTokenService) GenerateAccessToken(u domainuser.User, permissions []string) (string, error) {
	args := m.Called(u, permissions)
	return args.String(0), args.Error(1)
}
func (m *mockTokenService) GenerateRefreshToken(u domainuser.User) (string, time.Time, error) {
	args := m.Called(u)
	return args.String(0), args.Get(1).(time.Time), args.Error(2)
}
func (m *mockTokenService) ParseAccessToken(token string) (portuser.Claims, error) {
	args := m.Called(token)
	return args.Get(0).(portuser.Claims), args.Error(1)
}
func (m *mockTokenService) ParseRefreshToken(token string) (portuser.Claims, error) {
	args := m.Called(token)
	return args.Get(0).(portuser.Claims), args.Error(1)
}
func (m *mockTokenService) HashRefreshToken(token string) string {
	args := m.Called(token)
	return args.String(0)
}

type mockRBACRepo struct{ mock.Mock }

func (m *mockRBACRepo) FindPermissionsByRoleName(ctx context.Context, roleName string) ([]rbac.Permission, error) {
	args := m.Called(ctx, roleName)
	return args.Get(0).([]rbac.Permission), args.Error(1)
}
