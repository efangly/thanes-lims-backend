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

type LoginUseCase struct {
	users   portuser.UserRepository
	refresh portuser.RefreshTokenRepository
	tokens  portuser.TokenService
}

func NewLoginUseCase(users portuser.UserRepository, refresh portuser.RefreshTokenRepository, tokens portuser.TokenService) *LoginUseCase {
	return &LoginUseCase{users: users, refresh: refresh, tokens: tokens}
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// Execute returns shared.ErrUnauthorized for both "user not found" and
// "wrong password" - never leak which one it was.
func (uc *LoginUseCase) Execute(ctx context.Context, email, password string) (TokenPair, error) {
	u, err := uc.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return TokenPair{}, shared.ErrUnauthorized
		}
		return TokenPair{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return TokenPair{}, shared.ErrUnauthorized
	}

	return issueTokenPair(ctx, uc.tokens, uc.refresh, u)
}

// issueTokenPair generates and persists a fresh access/refresh pair for u.
// Shared between login and refresh (rotation).
func issueTokenPair(ctx context.Context, tokens portuser.TokenService, refreshRepo portuser.RefreshTokenRepository, u domainuser.User) (TokenPair, error) {
	access, err := tokens.GenerateAccessToken(u)
	if err != nil {
		return TokenPair{}, err
	}

	refreshRaw, expiresAt, err := tokens.GenerateRefreshToken(u)
	if err != nil {
		return TokenPair{}, err
	}

	_, err = refreshRepo.Create(ctx, domainuser.RefreshToken{
		UserID:    u.ID,
		TokenHash: tokens.HashRefreshToken(refreshRaw),
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	})
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{AccessToken: access, RefreshToken: refreshRaw}, nil
}
