package user_test

import (
	"context"
	"errors"
	"testing"
	"time"

	applicationuser "github.com/efangly/thanes-lims-backend/internal/application/user"
	"github.com/efangly/thanes-lims-backend/internal/domain/rbac"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRefreshUseCase_ValidRotate(t *testing.T) {
	users := new(mockUserRepo)
	refresh := new(mockRefreshRepo)
	tokens := new(mockTokenService)
	rbacRepo := new(mockRBACRepo)

	u := domainuser.User{ID: 1, Role: domainuser.RoleScientist}
	stored := domainuser.RefreshToken{ID: 10, UserID: 1, FamilyID: "family-1", FamilyCreatedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour), Revoked: false}
	perms := []rbac.Permission{{Module: rbac.ModuleSample, Action: rbac.ActionEdit}}

	tokens.On("ParseRefreshToken", "raw").Return(portuser.Claims{UserID: 1, Role: domainuser.RoleScientist}, nil)
	tokens.On("HashRefreshToken", "raw").Return("hash1")
	refresh.On("FindByTokenHash", mock.Anything, "hash1").Return(stored, nil)
	users.On("FindByID", mock.Anything, int64(1)).Return(u, nil)
	refresh.On("Revoke", mock.Anything, int64(10), "").Return(nil)
	rbacRepo.On("FindPermissionsByRoleName", mock.Anything, "Scientist").Return(perms, nil)
	tokens.On("GenerateAccessToken", u, []string{"sample:edit"}).Return("new-access", nil)
	tokens.On("GenerateRefreshToken", u).Return("new-refresh", time.Now().Add(time.Hour), nil)
	tokens.On("HashRefreshToken", "new-refresh").Return("hash2")
	refresh.On("Create", mock.Anything, mock.AnythingOfType("user.RefreshToken")).Return(domainuser.RefreshToken{}, nil)

	uc := applicationuser.NewRefreshUseCase(users, refresh, tokens, rbacRepo)
	pair, err := uc.Execute(context.Background(), "raw", "test-agent", "127.0.0.1")

	assert.NoError(t, err)
	assert.Equal(t, "new-access", pair.AccessToken)
	assert.Equal(t, "new-refresh", pair.RefreshToken)
}

func TestRefreshUseCase_ExpiredToken(t *testing.T) {
	users := new(mockUserRepo)
	refresh := new(mockRefreshRepo)
	tokens := new(mockTokenService)
	rbacRepo := new(mockRBACRepo)

	stored := domainuser.RefreshToken{ID: 10, UserID: 1, FamilyID: "family-1", FamilyCreatedAt: time.Now(), ExpiresAt: time.Now().Add(-time.Hour), Revoked: false}

	tokens.On("ParseRefreshToken", "raw").Return(portuser.Claims{UserID: 1}, nil)
	tokens.On("HashRefreshToken", "raw").Return("hash1")
	refresh.On("FindByTokenHash", mock.Anything, "hash1").Return(stored, nil)

	uc := applicationuser.NewRefreshUseCase(users, refresh, tokens, rbacRepo)
	_, err := uc.Execute(context.Background(), "raw", "test-agent", "127.0.0.1")

	assert.ErrorIs(t, err, shared.ErrUnauthorized)
}

// TestRefreshUseCase_ReuseDetection covers presenting a refresh token that
// was already rotated away (Revoked = true) - Reuse Detection must treat
// this as a leaked token and revoke every Session the user holds, not just
// this Token Family (see ADR 0004).
func TestRefreshUseCase_ReuseDetection(t *testing.T) {
	users := new(mockUserRepo)
	refresh := new(mockRefreshRepo)
	tokens := new(mockTokenService)
	rbacRepo := new(mockRBACRepo)

	stored := domainuser.RefreshToken{ID: 10, UserID: 1, FamilyID: "family-1", FamilyCreatedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour), Revoked: true}

	tokens.On("ParseRefreshToken", "raw").Return(portuser.Claims{UserID: 1}, nil)
	tokens.On("HashRefreshToken", "raw").Return("hash1")
	refresh.On("FindByTokenHash", mock.Anything, "hash1").Return(stored, nil)
	refresh.On("RevokeAllForUser", mock.Anything, int64(1)).Return(nil)

	uc := applicationuser.NewRefreshUseCase(users, refresh, tokens, rbacRepo)
	_, err := uc.Execute(context.Background(), "raw", "test-agent", "127.0.0.1")

	assert.ErrorIs(t, err, shared.ErrUnauthorized)
	refresh.AssertCalled(t, "RevokeAllForUser", mock.Anything, int64(1))
}

// TestRefreshUseCase_AbsoluteSessionLifetimeExceeded covers a Token Family
// that has been rotated continuously without ever going idle for 7 days -
// once 30 days have passed since the family was first created, Rolling
// Refresh must stop renewing it (see ADR 0004).
func TestRefreshUseCase_AbsoluteSessionLifetimeExceeded(t *testing.T) {
	users := new(mockUserRepo)
	refresh := new(mockRefreshRepo)
	tokens := new(mockTokenService)
	rbacRepo := new(mockRBACRepo)

	stored := domainuser.RefreshToken{
		ID:              10,
		UserID:          1,
		FamilyID:        "family-1",
		FamilyCreatedAt: time.Now().Add(-31 * 24 * time.Hour),
		ExpiresAt:       time.Now().Add(time.Hour),
		Revoked:         false,
	}

	tokens.On("ParseRefreshToken", "raw").Return(portuser.Claims{UserID: 1}, nil)
	tokens.On("HashRefreshToken", "raw").Return("hash1")
	refresh.On("FindByTokenHash", mock.Anything, "hash1").Return(stored, nil)

	uc := applicationuser.NewRefreshUseCase(users, refresh, tokens, rbacRepo)
	_, err := uc.Execute(context.Background(), "raw", "test-agent", "127.0.0.1")

	assert.ErrorIs(t, err, shared.ErrUnauthorized)
}

func TestRefreshUseCase_InvalidSignature(t *testing.T) {
	users := new(mockUserRepo)
	refresh := new(mockRefreshRepo)
	tokens := new(mockTokenService)
	rbacRepo := new(mockRBACRepo)

	tokens.On("ParseRefreshToken", "bad").Return(portuser.Claims{}, errors.New("bad signature"))

	uc := applicationuser.NewRefreshUseCase(users, refresh, tokens, rbacRepo)
	_, err := uc.Execute(context.Background(), "bad", "test-agent", "127.0.0.1")

	assert.ErrorIs(t, err, shared.ErrUnauthorized)
}
