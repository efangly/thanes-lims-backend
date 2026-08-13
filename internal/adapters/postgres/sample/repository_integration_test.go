//go:build integration

package sample_test

import (
	"context"
	"testing"

	"github.com/efangly/thanes-lims-backend/internal/adapters/postgres/pgtest"
	"github.com/efangly/thanes-lims-backend/internal/adapters/postgres/sample"
	domainsample "github.com/efangly/thanes-lims-backend/internal/domain/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portsample "github.com/efangly/thanes-lims-backend/internal/ports/sample"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSampleRepository_CRUD(t *testing.T) {
	db := pgtest.SetupPostgres(t)
	repo := sample.New(db)
	ctx := context.Background()

	created, err := repo.Create(ctx, domainsample.Sample{
		ID:        "SMP-0001",
		Name:      "Blood panel A",
		Type:      domainsample.TypeBlood,
		Custodian: "Somchai",
		Location:  "Lab A",
		Status:    domainsample.StatusPending,
	})
	require.NoError(t, err)
	assert.Equal(t, "SMP-0001", created.ID)

	byID, err := repo.FindByID(ctx, "SMP-0001")
	require.NoError(t, err)
	assert.Equal(t, created.Name, byID.Name)
	assert.Equal(t, domainsample.StatusPending, byID.Status)

	all, err := repo.List(ctx, portsample.ListFilter{})
	require.NoError(t, err)
	assert.Len(t, all, 1)

	pendingStatus := domainsample.StatusPending
	filtered, err := repo.List(ctx, portsample.ListFilter{Status: &pendingStatus})
	require.NoError(t, err)
	assert.Len(t, filtered, 1)

	testingStatus := domainsample.StatusTesting
	noneFiltered, err := repo.List(ctx, portsample.ListFilter{Status: &testingStatus})
	require.NoError(t, err)
	assert.Len(t, noneFiltered, 0)

	updated := byID
	updated.Status = domainsample.StatusTesting
	result, err := repo.UpdateStatus(ctx, updated)
	require.NoError(t, err)
	assert.Equal(t, domainsample.StatusTesting, result.Status)
}

func TestSampleRepository_NotFound(t *testing.T) {
	db := pgtest.SetupPostgres(t)
	repo := sample.New(db)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, "does-not-exist")
	assert.ErrorIs(t, err, shared.ErrNotFound)
}
