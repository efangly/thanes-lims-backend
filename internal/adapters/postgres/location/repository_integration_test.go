//go:build integration

package location_test

import (
	"context"
	"testing"

	"github.com/efangly/thanes-lims-backend/internal/adapters/postgres/location"
	"github.com/efangly/thanes-lims-backend/internal/adapters/postgres/pgtest"
	domainlocation "github.com/efangly/thanes-lims-backend/internal/domain/location"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocationRepository_TreeCRUD(t *testing.T) {
	db := pgtest.SetupPostgres(t)
	repo := location.New(db)
	ctx := context.Background()

	cabinet, err := repo.Create(ctx, domainlocation.Location{
		ID: "LOC-00001", Name: "Fridge-A", Kind: domainlocation.KindSampleStorage,
		LevelType: domainlocation.LevelCabinet, BarcodeCode: "LOC-BC-00001",
	})
	require.NoError(t, err)
	assert.True(t, cabinet.IsRoot())

	byBarcode, err := repo.FindByBarcode(ctx, "LOC-BC-00001")
	require.NoError(t, err)
	assert.Equal(t, "LOC-00001", byBarcode.ID)

	cabinetID := cabinet.ID
	shelves, err := repo.CreateMany(ctx, []domainlocation.Location{
		{ID: "LOC-00002", ParentID: &cabinetID, Name: "Shelf-1", Kind: domainlocation.KindSampleStorage, LevelType: domainlocation.LevelShelf, BarcodeCode: "LOC-BC-00002"},
		{ID: "LOC-00003", ParentID: &cabinetID, Name: "Shelf-2", Kind: domainlocation.KindSampleStorage, LevelType: domainlocation.LevelShelf, BarcodeCode: "LOC-BC-00003"},
	})
	require.NoError(t, err)
	require.Len(t, shelves, 2)

	// Root listing is scoped to one Kind.
	sampleRoots, err := repo.ListRoots(ctx, domainlocation.KindSampleStorage)
	require.NoError(t, err)
	assert.Len(t, sampleRoots, 1)
	eqRoots, err := repo.ListRoots(ctx, domainlocation.KindEquipmentStorage)
	require.NoError(t, err)
	assert.Empty(t, eqRoots)

	byID, err := repo.GetByID(ctx, cabinetID)
	require.NoError(t, err)
	assert.Equal(t, "Fridge-A", byID.Name)

	roots, err := repo.ListChildren(ctx, nil)
	require.NoError(t, err)
	assert.Len(t, roots, 1)

	children, err := repo.ListChildren(ctx, &cabinetID)
	require.NoError(t, err)
	assert.Len(t, children, 2)

	found, err := repo.FindChildByName(ctx, &cabinetID, "Shelf-1")
	require.NoError(t, err)
	assert.Equal(t, "LOC-00002", found.ID)

	_, err = repo.FindChildByName(ctx, &cabinetID, "Shelf-99")
	assert.ErrorIs(t, err, shared.ErrNotFound)

	hasChildren, err := repo.HasChildren(ctx, cabinetID)
	require.NoError(t, err)
	assert.True(t, hasChildren)

	shelf1ID := shelves[0].ID
	hasChildren, err = repo.HasChildren(ctx, shelf1ID)
	require.NoError(t, err)
	assert.False(t, hasChildren)

	fullPath, err := repo.FullPath(ctx, shelf1ID)
	require.NoError(t, err)
	assert.Equal(t, "Fridge-A / Shelf-1", fullPath)

	// A parent with children can't be deleted - the FK on locations.parent_id
	// has no ON DELETE CASCADE, so Postgres restricts it. Application-layer
	// pre-checks (DeleteLocationUseCase) are what turn this into a clean
	// shared.ErrConflict for API callers; the repository itself just
	// surfaces whatever the DB returns.
	err = repo.Delete(ctx, cabinetID)
	assert.Error(t, err)

	// A leaf can be deleted freely.
	err = repo.Delete(ctx, shelf1ID)
	assert.NoError(t, err)
	_, err = repo.GetByID(ctx, shelf1ID)
	assert.ErrorIs(t, err, shared.ErrNotFound)
}

func TestLocationRepository_NotFound(t *testing.T) {
	db := pgtest.SetupPostgres(t)
	repo := location.New(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "does-not-exist")
	assert.ErrorIs(t, err, shared.ErrNotFound)

	_, err = repo.FullPath(ctx, "does-not-exist")
	assert.ErrorIs(t, err, shared.ErrNotFound)

	err = repo.Delete(ctx, "does-not-exist")
	assert.ErrorIs(t, err, shared.ErrNotFound)
}

func TestLocationRepository_RootNameUniqueness(t *testing.T) {
	db := pgtest.SetupPostgres(t)
	repo := location.New(db)
	ctx := context.Background()

	_, err := repo.Create(ctx, domainlocation.Location{
		ID: "LOC-00010", Name: "Fridge-A", Kind: domainlocation.KindSampleStorage, LevelType: domainlocation.LevelCabinet,
	})
	require.NoError(t, err)

	_, err = repo.Create(ctx, domainlocation.Location{
		ID: "LOC-00011", Name: "Fridge-A", Kind: domainlocation.KindSampleStorage, LevelType: domainlocation.LevelCabinet,
	})
	assert.Error(t, err)
}
