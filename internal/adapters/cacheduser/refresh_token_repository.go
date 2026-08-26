// Package cacheduser decorates a Postgres-backed
// ports/user.RefreshTokenRepository with a Redis read-through Cache, per
// docs/adr/0005-redis-cache-for-refresh-tokens-and-location-full-path.md.
// Postgres (the wrapped repository) stays the source of truth; the Cache
// only exists to avoid a Postgres round-trip on the common case.
package cacheduser

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/efangly/thanes-lims-backend/internal/domain/user"
	"github.com/efangly/thanes-lims-backend/internal/ports/cache"
	portuser "github.com/efangly/thanes-lims-backend/internal/ports/user"
)

type CachedRefreshTokenRepository struct {
	next  portuser.RefreshTokenRepository
	cache cache.Cache
}

func NewCachedRefreshTokenRepository(next portuser.RefreshTokenRepository, c cache.Cache) *CachedRefreshTokenRepository {
	return &CachedRefreshTokenRepository{next: next, cache: c}
}

func refreshTokenCacheKey(tokenHash string) string {
	return "refresh:" + tokenHash
}

// Create persists rt via the wrapped repository, then best-effort populates
// the cache entry. A cache write failure here is never fatal: the next
// FindByTokenHash for this token will simply miss the cache and fall back
// to Postgres, self-healing once Redis is reachable again.
func (r *CachedRefreshTokenRepository) Create(ctx context.Context, rt user.RefreshToken) (user.RefreshToken, error) {
	created, err := r.next.Create(ctx, rt)
	if err != nil {
		return user.RefreshToken{}, err
	}
	if err := r.populate(ctx, created); err != nil {
		log.Printf("cacheduser: populate cache on create: %v", err)
	}
	return created, nil
}

// FindByTokenHash reads the cached full row when present. A genuine cache
// error (Redis unreachable) is NOT treated as a miss and is not silently
// papered over by falling back to Postgres - per ADR 0005 this path is
// fail-closed, since the caller can't otherwise distinguish "not revoked"
// from "unknown because the cache is down". Only a real cache miss
// (shared.ErrNotFound) falls back to Postgres and repopulates the cache.
func (r *CachedRefreshTokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (user.RefreshToken, error) {
	data, err := r.cache.Get(ctx, refreshTokenCacheKey(tokenHash))
	if err == nil {
		var rt user.RefreshToken
		if unmarshalErr := json.Unmarshal(data, &rt); unmarshalErr != nil {
			return user.RefreshToken{}, fmt.Errorf("cacheduser: decode cached refresh token: %w", unmarshalErr)
		}
		return rt, nil
	}
	if !errors.Is(err, shared.ErrNotFound) {
		return user.RefreshToken{}, fmt.Errorf("cacheduser: cache unreachable: %w", err)
	}

	stored, err := r.next.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return user.RefreshToken{}, err
	}
	if err := r.populate(ctx, stored); err != nil {
		log.Printf("cacheduser: populate cache on miss: %v", err)
	}
	return stored, nil
}

// Revoke updates Postgres first (the source of truth), then deletes the
// cache entry. The cache delete's error is propagated rather than swallowed:
// per ADR 0005 an un-invalidated cache entry for a revoked token is exactly
// the risk this whole cache is designed to avoid.
func (r *CachedRefreshTokenRepository) Revoke(ctx context.Context, id int64, tokenHash string) error {
	if err := r.next.Revoke(ctx, id, tokenHash); err != nil {
		return err
	}
	if err := r.cache.Delete(ctx, refreshTokenCacheKey(tokenHash)); err != nil {
		return fmt.Errorf("cacheduser: invalidate on revoke: %w", err)
	}
	return nil
}

// RevokeAllForUser reads every active token hash for userID before revoking
// in Postgres (see ADR 0005 - no separate user->tokens index is kept in
// Redis, since this is a rare security-path operation, not a hot path), then
// deletes each cache entry. Individual cache-delete failures are logged but
// don't abort the loop or fail the call: Postgres has already revoked every
// row, and each surviving stale cache entry is still bounded by its own TTL.
func (r *CachedRefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID int64) error {
	hashes, err := r.next.FindTokenHashesByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if err := r.next.RevokeAllForUser(ctx, userID); err != nil {
		return err
	}
	for _, hash := range hashes {
		if err := r.cache.Delete(ctx, refreshTokenCacheKey(hash)); err != nil {
			log.Printf("cacheduser: invalidate on revoke-all: %v", err)
		}
	}
	return nil
}

func (r *CachedRefreshTokenRepository) FindTokenHashesByUserID(ctx context.Context, userID int64) ([]string, error) {
	return r.next.FindTokenHashesByUserID(ctx, userID)
}

// populate writes rt into the cache with a TTL bounded by its own expiry -
// an already-expired row (or one expiring immediately) is simply not cached.
func (r *CachedRefreshTokenRepository) populate(ctx context.Context, rt user.RefreshToken) error {
	ttl := time.Until(rt.ExpiresAt)
	if ttl <= 0 {
		return nil
	}
	data, err := json.Marshal(rt)
	if err != nil {
		return fmt.Errorf("cacheduser: encode refresh token: %w", err)
	}
	return r.cache.Set(ctx, refreshTokenCacheKey(rt.TokenHash), data, ttl)
}
