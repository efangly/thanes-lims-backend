//go:build integration

package environment_test

import (
	"context"
	"testing"

	postgresenvironment "github.com/efangly/thanes-lims-backend/internal/adapters/postgres/environment"
	"github.com/efangly/thanes-lims-backend/internal/adapters/postgres/pgtest"
	"github.com/efangly/thanes-lims-backend/internal/domain/environment"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Gauge is config-only, seeded out of band - there's no CreateGauge use
// case, so each test below inserts one directly. PartnerDevice.Location has
// a real FK to gauges(location), so a Partner Device row needs one to
// point at.

func TestPartnerDeviceRepository_CRUD(t *testing.T) {
	db := pgtest.SetupPostgres(t)
	require.NoError(t, db.Exec(`INSERT INTO gauges (location, unit, range_min, range_max) VALUES (?, ?, ?, ?)`,
		"ward-3", "C", 2.0, 8.0).Error)

	repo := postgresenvironment.NewPartnerDeviceRepository(db)
	ctx := context.Background()

	created, err := repo.Create(ctx, environment.PartnerDevice{Serial: "SN-00042", Location: "ward-3", Active: true})
	require.NoError(t, err)
	assert.Equal(t, "SN-00042", created.Serial)
	assert.True(t, created.Active)

	found, err := repo.FindBySerial(ctx, "SN-00042")
	require.NoError(t, err)
	assert.Equal(t, "ward-3", found.Location)

	_, err = repo.FindBySerial(ctx, "missing")
	assert.ErrorIs(t, err, shared.ErrNotFound)

	updated, err := repo.Update(ctx, environment.PartnerDevice{Serial: "SN-00042", Location: "ward-3", Active: false})
	require.NoError(t, err)
	assert.False(t, updated.Active)

	afterUpdate, err := repo.FindBySerial(ctx, "SN-00042")
	require.NoError(t, err)
	assert.False(t, afterUpdate.Active)

	list, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, list, 1)
}

func TestPartnerDeviceRepository_SerialIsPrimaryKey(t *testing.T) {
	db := pgtest.SetupPostgres(t)
	require.NoError(t, db.Exec(`INSERT INTO gauges (location, unit, range_min, range_max) VALUES (?, ?, ?, ?)`,
		"ward-3", "C", 2.0, 8.0).Error)

	repo := postgresenvironment.NewPartnerDeviceRepository(db)
	ctx := context.Background()

	_, err := repo.Create(ctx, environment.PartnerDevice{Serial: "SN-00042", Location: "ward-3", Active: true})
	require.NoError(t, err)

	_, err = repo.Create(ctx, environment.PartnerDevice{Serial: "SN-00042", Location: "ward-3", Active: true})
	assert.Error(t, err)
}

func TestPartnerDeviceRepository_LocationRequiresExistingGauge(t *testing.T) {
	db := pgtest.SetupPostgres(t)
	repo := postgresenvironment.NewPartnerDeviceRepository(db)
	ctx := context.Background()

	_, err := repo.Create(ctx, environment.PartnerDevice{Serial: "SN-00042", Location: "no-such-location", Active: true})
	assert.Error(t, err)
}
