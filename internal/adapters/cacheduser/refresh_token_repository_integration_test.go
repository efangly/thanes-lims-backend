//go:build integration

package cacheduser_test

import (
	"context"
	"testing"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/adapters/cacheduser"
	"github.com/efangly/thanes-lims-backend/internal/adapters/postgres/pgtest"
	postgresuser "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/user"
	"github.com/efangly/thanes-lims-backend/internal/adapters/redis/rtest"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	domainuser "github.com/efangly/thanes-lims-backend/internal/domain/user"
	"github.com/stretchr/testify/require"
)

func TestCachedRefreshTokenRepository_ReadThroughAndInvalidation(t *testing.T) {
	db := pgtest.SetupPostgres(t)
	c := rtest.SetupRedis(t)
	ctx := context.Background()

	users := postgresuser.New(db)
	u, err := users.Create(ctx, domainuser.User{
		Name: "Cache Test", Email: "cache-test@example.com",
		PasswordHash: "hashed", Role: domainuser.RoleScientist,
	})
	require.NoError(t, err)

	postgresRepo := postgresuser.NewRefreshTokenRepository(db)
	sut := cacheduser.NewCachedRefreshTokenRepository(postgresRepo, c)

	created, err := sut.Create(ctx, domainuser.RefreshToken{
		UserID: u.ID, FamilyID: "family-1", FamilyCreatedAt: time.Now(),
		TokenHash: "integration-hash-1", ExpiresAt: time.Now().Add(time.Hour), CreatedAt: time.Now(),
	})
	require.NoError(t, err)

	// First read populates the cache (miss -> fallback -> populate);
	// deleting the underlying Postgres row afterwards proves the second
	// read comes from the cache, not from Postgres.
	first, err := sut.FindByTokenHash(ctx, "integration-hash-1")
	require.NoError(t, err)
	require.Equal(t, created.ID, first.ID)

	require.NoError(t, db.Exec("DELETE FROM refresh_tokens WHERE id = ?", created.ID).Error)

	cached, err := sut.FindByTokenHash(ctx, "integration-hash-1")
	require.NoError(t, err)
	require.Equal(t, created.ID, cached.ID, "expected the cached row to still be served after the Postgres row was deleted")

	// Revoke invalidates the cache entry, so the read falls through to
	// Postgres again - which now genuinely has nothing to find.
	require.NoError(t, sut.Revoke(ctx, created.ID, created.TokenHash))
	_, err = sut.FindByTokenHash(ctx, "integration-hash-1")
	require.ErrorIs(t, err, shared.ErrNotFound)
}

func TestCachedRefreshTokenRepository_RevokeAllForUser_InvalidatesEveryToken(t *testing.T) {
	db := pgtest.SetupPostgres(t)
	c := rtest.SetupRedis(t)
	ctx := context.Background()

	users := postgresuser.New(db)
	u, err := users.Create(ctx, domainuser.User{
		Name: "Cache Test 2", Email: "cache-test-2@example.com",
		PasswordHash: "hashed", Role: domainuser.RoleScientist,
	})
	require.NoError(t, err)

	postgresRepo := postgresuser.NewRefreshTokenRepository(db)
	sut := cacheduser.NewCachedRefreshTokenRepository(postgresRepo, c)

	for _, hash := range []string{"multi-hash-1", "multi-hash-2"} {
		_, err := sut.Create(ctx, domainuser.RefreshToken{
			UserID: u.ID, FamilyID: hash, FamilyCreatedAt: time.Now(),
			TokenHash: hash, ExpiresAt: time.Now().Add(time.Hour), CreatedAt: time.Now(),
		})
		require.NoError(t, err)
		// Warm the cache for each token before revoking all.
		_, err = sut.FindByTokenHash(ctx, hash)
		require.NoError(t, err)
	}

	require.NoError(t, sut.RevokeAllForUser(ctx, u.ID))

	for _, hash := range []string{"multi-hash-1", "multi-hash-2"} {
		got, err := sut.FindByTokenHash(ctx, hash)
		require.NoError(t, err)
		require.True(t, got.Revoked, "expected %s to be reported revoked after cache invalidation forced a Postgres re-read", hash)
	}
}
