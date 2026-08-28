//go:build integration

package cachedlocation_test

import (
	"context"
	"testing"

	"github.com/efangly/thanes-lims-backend/internal/adapters/cachedlocation"
	"github.com/efangly/thanes-lims-backend/internal/adapters/postgres/location"
	"github.com/efangly/thanes-lims-backend/internal/adapters/postgres/pgtest"
	"github.com/efangly/thanes-lims-backend/internal/adapters/redis/rtest"
	domainlocation "github.com/efangly/thanes-lims-backend/internal/domain/location"
	"github.com/stretchr/testify/require"
)

func TestCachedRepository_FullPath_ServesFromCacheAfterFirstRead(t *testing.T) {
	db := pgtest.SetupPostgres(t)
	c := rtest.SetupRedis(t)
	ctx := context.Background()

	postgresRepo := location.New(db)
	sut := cachedlocation.NewCachedRepository(postgresRepo, c)

	cabinet, err := sut.Create(ctx, domainlocation.Location{
		ID: "LOC-90001", Name: "Fridge-Cache-Test", Kind: domainlocation.KindSampleStorage, LevelType: domainlocation.LevelCabinet,
	})
	require.NoError(t, err)
	cabinetID := cabinet.ID

	shelfID := "LOC-90002"
	_, err = sut.Create(ctx, domainlocation.Location{
		ID: shelfID, ParentID: &cabinetID, Name: "Shelf-Cache-Test", Kind: domainlocation.KindSampleStorage, LevelType: domainlocation.LevelShelf,
	})
	require.NoError(t, err)

	first, err := sut.FullPath(ctx, shelfID)
	require.NoError(t, err)
	require.Equal(t, "Fridge-Cache-Test / Shelf-Cache-Test", first)

	// Rename the underlying row directly in Postgres, bypassing the cache -
	// a cached hit should still return the stale name until the TTL expires
	// (fail-open, TTL-only invalidation, see ADR 0005).
	require.NoError(t, db.Exec("UPDATE locations SET name = ? WHERE id = ?", "Renamed-Without-Invalidation", shelfID).Error)

	cached, err := sut.FullPath(ctx, shelfID)
	require.NoError(t, err)
	require.Equal(t, first, cached, "expected the stale cached Full Path to still be served (TTL-only invalidation)")
}
