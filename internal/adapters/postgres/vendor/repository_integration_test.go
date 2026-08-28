//go:build integration

package vendor_test

import (
	"context"
	"testing"

	"github.com/efangly/thanes-lims-backend/internal/adapters/postgres/pgtest"
	"github.com/efangly/thanes-lims-backend/internal/adapters/postgres/vendor"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	domainvendor "github.com/efangly/thanes-lims-backend/internal/domain/vendor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVendorRepository_CRUD(t *testing.T) {
	db := pgtest.SetupPostgres(t)
	repo := vendor.New(db)
	ctx := context.Background()

	created, err := repo.Create(ctx, domainvendor.Vendor{
		ID: "VEN-00001", Name: "Acme Lab Supplies",
		ContactName: "Jane", ContactPhone: "02-111-2222", ContactEmail: "jane@acme.com",
	})
	require.NoError(t, err)
	assert.Equal(t, "VEN-00001", created.ID)

	byID, err := repo.FindByID(ctx, "VEN-00001")
	require.NoError(t, err)
	assert.Equal(t, "Acme Lab Supplies", byID.Name)

	byName, err := repo.FindByName(ctx, "Acme Lab Supplies")
	require.NoError(t, err)
	assert.Equal(t, "VEN-00001", byName.ID)

	_, err = repo.FindByName(ctx, "Nope")
	assert.ErrorIs(t, err, shared.ErrNotFound)

	updated, err := repo.Update(ctx, domainvendor.Vendor{
		ID: "VEN-00001", Name: "Acme Lab Supplies", ContactName: "John", Address: "",
	})
	require.NoError(t, err)
	assert.Equal(t, "John", updated.ContactName)
	assert.Empty(t, updated.ContactPhone)

	list, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, list, 1)

	_, err = repo.FindByID(ctx, "missing")
	assert.ErrorIs(t, err, shared.ErrNotFound)
}

func TestVendorRepository_NameUniqueness(t *testing.T) {
	db := pgtest.SetupPostgres(t)
	repo := vendor.New(db)
	ctx := context.Background()

	_, err := repo.Create(ctx, domainvendor.Vendor{ID: "VEN-00001", Name: "Acme"})
	require.NoError(t, err)

	_, err = repo.Create(ctx, domainvendor.Vendor{ID: "VEN-00002", Name: "Acme"})
	assert.Error(t, err)
}
