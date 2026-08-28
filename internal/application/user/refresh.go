package user

import (
	"context"
	"log"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portrbac "github.com/efangly/thanes-lims-backend/internal/ports/rbac"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
)

// absoluteSessionLifetime caps a Token Family's lifetime regardless of how
// often it's actively rotated, so a Session can't persist forever purely by
// staying in use (see ADR 0004).
const absoluteSessionLifetime = 30 * 24 * time.Hour

type RefreshUseCase struct {
	users   portuser.UserRepository
	refresh portuser.RefreshTokenRepository
	tokens  portuser.TokenService
	rbac    portrbac.Repository
}

func NewRefreshUseCase(users portuser.UserRepository, refresh portuser.RefreshTokenRepository, tokens portuser.TokenService, rbacRepo portrbac.Repository) *RefreshUseCase {
	return &RefreshUseCase{users: users, refresh: refresh, tokens: tokens, rbac: rbacRepo}
}

// Execute rotates the refresh token: the presented token is revoked and a
// brand new access/refresh pair is issued within the same Token Family, so a
// stolen-then-reused refresh token is immediately detectable (it'll already
// be revoked). If a revoked token is presented again (Reuse Detection), every
// Session the user holds is revoked - the token has leaked, so the whole
// account is treated as compromised, not just this Token Family. Permissions
// embedded in the new access token are recomputed fresh from the RBAC
// tables (not copied from the old token's claims), so a Role or Permission
// change since the last login/refresh takes effect here (see ADR 0002).
func (uc *RefreshUseCase) Execute(ctx context.Context, refreshTokenRaw, userAgent, ipAddress string) (TokenPair, error) {
	if _, err := uc.tokens.ParseRefreshToken(refreshTokenRaw); err != nil {
		return TokenPair{}, shared.ErrUnauthorized
	}

	hash := uc.tokens.HashRefreshToken(refreshTokenRaw)
	stored, err := uc.refresh.FindByTokenHash(ctx, hash)
	if err != nil {
		return TokenPair{}, shared.ErrUnauthorized
	}

	if stored.Revoked {
		// Reuse Detection: a revoked token was presented again - it has leaked.
		// If revoke-all itself fails the attacker's Token Family survives, so
		// this must be loud (error log + alerting hook), never swallowed.
		if err := uc.refresh.RevokeAllForUser(ctx, stored.UserID); err != nil {
			log.Printf("ERROR refresh: reuse detected for user %d but RevokeAllForUser failed: %v", stored.UserID, err)
		}
		return TokenPair{}, shared.ErrUnauthorized
	}
	if stored.ExpiresAt.Before(time.Now()) {
		return TokenPair{}, shared.ErrUnauthorized
	}
	if time.Since(stored.FamilyCreatedAt) > absoluteSessionLifetime {
		return TokenPair{}, shared.ErrUnauthorized
	}

	u, err := uc.users.FindByID(ctx, stored.UserID)
	if err != nil {
		return TokenPair{}, shared.ErrUnauthorized
	}

	affected, err := uc.refresh.Revoke(ctx, stored.ID, stored.TokenHash)
	if err != nil {
		return TokenPair{}, err
	}
	if affected == 0 {
		// The row was already revoked between FindByTokenHash and here -
		// another request rotated this same token first (double-spend) or the
		// token was replayed after rotation. Either way it has leaked: kill
		// every Session the user holds, same as explicit Reuse Detection.
		if err := uc.refresh.RevokeAllForUser(ctx, stored.UserID); err != nil {
			log.Printf("ERROR refresh: double-spend detected for user %d but RevokeAllForUser failed: %v", stored.UserID, err)
		}
		return TokenPair{}, shared.ErrUnauthorized
	}

	perms, err := uc.rbac.FindPermissionsByRoleName(ctx, u.Role.DisplayName())
	if err != nil {
		return TokenPair{}, err
	}

	family := refreshFamily{ID: stored.FamilyID, CreatedAt: stored.FamilyCreatedAt, UserAgent: userAgent, IPAddress: ipAddress}
	return issueTokenPair(ctx, uc.tokens, uc.refresh, u, permissionKeys(perms), family)
}
