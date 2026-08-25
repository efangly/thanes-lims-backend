package user

import (
	"context"
	"errors"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/rbac"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	portrbac "github.com/efangly/thanes-lims-backend/internal/ports/rbac"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type LoginUseCase struct {
	users   portuser.UserRepository
	refresh portuser.RefreshTokenRepository
	tokens  portuser.TokenService
	rbac    portrbac.Repository
}

func NewLoginUseCase(users portuser.UserRepository, refresh portuser.RefreshTokenRepository, tokens portuser.TokenService, rbacRepo portrbac.Repository) *LoginUseCase {
	return &LoginUseCase{users: users, refresh: refresh, tokens: tokens, rbac: rbacRepo}
}

type TokenPair struct {
	AccessToken      string
	RefreshToken     string
	RefreshExpiresAt time.Time
}

// refreshFamily identifies the Token Family a new refresh token row belongs
// to: a fresh family on Login, the same family carried forward on Rolling
// Refresh rotation (see ADR 0004).
type refreshFamily struct {
	ID        string
	CreatedAt time.Time
	UserAgent string
	IPAddress string
}

// Execute returns shared.ErrUnauthorized for both "user not found" and
// "wrong password" - never leak which one it was.
func (uc *LoginUseCase) Execute(ctx context.Context, email, password, userAgent, ipAddress string) (TokenPair, error) {
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

	perms, err := uc.rbac.FindPermissionsByRoleName(ctx, u.Role.DisplayName())
	if err != nil {
		return TokenPair{}, err
	}

	family := refreshFamily{ID: uuid.NewString(), CreatedAt: time.Now(), UserAgent: userAgent, IPAddress: ipAddress}
	return issueTokenPair(ctx, uc.tokens, uc.refresh, u, permissionKeys(perms), family)
}

// issueTokenPair generates and persists a fresh access/refresh pair for u,
// embedding permissions (compact "module:action" strings) into the access
// token. Shared between login and refresh (rotation).
func issueTokenPair(ctx context.Context, tokens portuser.TokenService, refreshRepo portuser.RefreshTokenRepository, u domainuser.User, permissions []string, family refreshFamily) (TokenPair, error) {
	access, err := tokens.GenerateAccessToken(u, permissions)
	if err != nil {
		return TokenPair{}, err
	}

	refreshRaw, expiresAt, err := tokens.GenerateRefreshToken(u)
	if err != nil {
		return TokenPair{}, err
	}

	_, err = refreshRepo.Create(ctx, domainuser.RefreshToken{
		UserID:          u.ID,
		FamilyID:        family.ID,
		FamilyCreatedAt: family.CreatedAt,
		TokenHash:       tokens.HashRefreshToken(refreshRaw),
		ExpiresAt:       expiresAt,
		CreatedAt:       time.Now(),
		UserAgent:       family.UserAgent,
		IPAddress:       family.IPAddress,
	})
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{AccessToken: access, RefreshToken: refreshRaw, RefreshExpiresAt: expiresAt}, nil
}

// permissionKeys converts resolved Permissions into the compact
// "module:action" strings embedded in the JWT (see rbac.Permission.Key and
// ADR 0002).
func permissionKeys(perms []rbac.Permission) []string {
	out := make([]string, len(perms))
	for i, p := range perms {
		out[i] = p.Key()
	}
	return out
}
