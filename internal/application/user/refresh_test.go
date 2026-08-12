package user_test

import (
	"context"
	"errors"
	"testing"
	"time"

	applicationuser "github.com/efangly/thanes-lims-backend/internal/application/user"
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

	u := domainuser.User{ID: 1, Role: domainuser.RoleScientist}
	stored := domainuser.RefreshToken{ID: 10, UserID: 1, ExpiresAt: time.Now().Add(time.Hour), Revoked: false}

	tokens.On("ParseRefreshToken", "raw").Return(portuser.Claims{UserID: 1, Role: domainuser.RoleScientist}, nil)
	tokens.On("HashRefreshToken", "raw").Return("hash1")
	refresh.On("FindByTokenHash", mock.Anything, "hash1").Return(stored, nil)
	users.On("FindByID", mock.Anything, int64(1)).Return(u, nil)
	refresh.On("Revoke", mock.Anything, int64(10)).Return(nil)
	tokens.On("GenerateAccessToken", u).Return("new-access", nil)
	tokens.On("GenerateRefreshToken", u).Return("new-refresh", time.Now().Add(time.Hour), nil)
	tokens.On("HashRefreshToken", "new-refresh").Return("hash2")
	refresh.On("Create", mock.Anything, mock.AnythingOfType("user.RefreshToken")).Return(domainuser.RefreshToken{}, nil)

	uc := applicationuser.NewRefreshUseCase(users, refresh, tokens)
	pair, err := uc.Execute(context.Background(), "raw")

	assert.NoError(t, err)
	assert.Equal(t, "new-access", pair.AccessToken)
	assert.Equal(t, "new-refresh", pair.RefreshToken)
}

func TestRefreshUseCase_ExpiredToken(t *testing.T) {
	users := new(mockUserRepo)
	refresh := new(mockRefreshRepo)
	tokens := new(mockTokenService)

	stored := domainuser.RefreshToken{ID: 10, UserID: 1, ExpiresAt: time.Now().Add(-time.Hour), Revoked: false}

	tokens.On("ParseRefreshToken", "raw").Return(portuser.Claims{UserID: 1}, nil)
	tokens.On("HashRefreshToken", "raw").Return("hash1")
	refresh.On("FindByTokenHash", mock.Anything, "hash1").Return(stored, nil)

	uc := applicationuser.NewRefreshUseCase(users, refresh, tokens)
	_, err := uc.Execute(context.Background(), "raw")

	assert.ErrorIs(t, err, shared.ErrUnauthorized)
}

func TestRefreshUseCase_RevokedToken(t *testing.T) {
	users := new(mockUserRepo)
	refresh := new(mockRefreshRepo)
	tokens := new(mockTokenService)

	stored := domainuser.RefreshToken{ID: 10, UserID: 1, ExpiresAt: time.Now().Add(time.Hour), Revoked: true}

	tokens.On("ParseRefreshToken", "raw").Return(portuser.Claims{UserID: 1}, nil)
	tokens.On("HashRefreshToken", "raw").Return("hash1")
	refresh.On("FindByTokenHash", mock.Anything, "hash1").Return(stored, nil)

	uc := applicationuser.NewRefreshUseCase(users, refresh, tokens)
	_, err := uc.Execute(context.Background(), "raw")

	assert.ErrorIs(t, err, shared.ErrUnauthorized)
}

func TestRefreshUseCase_InvalidSignature(t *testing.T) {
	users := new(mockUserRepo)
	refresh := new(mockRefreshRepo)
	tokens := new(mockTokenService)

	tokens.On("ParseRefreshToken", "bad").Return(portuser.Claims{}, errors.New("bad signature"))

	uc := applicationuser.NewRefreshUseCase(users, refresh, tokens)
	_, err := uc.Execute(context.Background(), "bad")

	assert.ErrorIs(t, err, shared.ErrUnauthorized)
}
