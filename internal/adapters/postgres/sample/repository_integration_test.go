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
	"gorm.io/gorm"
)

func seedCustodian(t *testing.T, db *gorm.DB) int64 {
	t.Helper()
	require.NoError(t, db.Exec(
		`INSERT INTO users (name, email, password_hash, role_id)
		 VALUES ('Somchai', 'somchai@test.local', 'x', (SELECT id FROM roles WHERE name = 'Admin'))`,
	).Error)
	var id int64
	require.NoError(t, db.Raw(`SELECT id FROM users WHERE email = 'somchai@test.local'`).Scan(&id).Error)
	return id
}

func TestSampleRepository_CRUD(t *testing.T) {
	db := pgtest.SetupPostgres(t)
	repo := sample.New(db)
	ctx := context.Background()

	custodianID := seedCustodian(t, db)

	created, err := repo.Create(ctx, domainsample.Sample{
		ID:              "SMP-0001",
		Name:            "Blood panel A",
		Type:            domainsample.TypeBlood,
		CustodianUserID: custodianID,
		Status:          domainsample.StatusPending,
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

func TestSampleRepository_BarcodeAndFilters(t *testing.T) {
	db := pgtest.SetupPostgres(t)
	repo := sample.New(db)
	ctx := context.Background()

	custodianID := seedCustodian(t, db)
	require.NoError(t, db.Exec(
		`INSERT INTO locations (id, name, kind, level_type, barcode_code)
		 VALUES ('LOC-INT-1', 'Fridge-Integration', 'sample_storage', 'cabinet', 'LOC-BC-90001')`,
	).Error)
	locID := "LOC-INT-1"

	_, err := repo.Create(ctx, domainsample.Sample{
		ID: "SMP-B1", Name: "Has barcode", Type: domainsample.TypeBlood,
		CustodianUserID: custodianID, LocationID: &locID, Status: domainsample.StatusPending,
		BarcodeID: strPtrI("BC-INT-001"), Description: "notes",
	})
	require.NoError(t, err)
	_, err = repo.Create(ctx, domainsample.Sample{
		ID: "SMP-B2", Name: "No barcode", Type: domainsample.TypeUrine,
		CustodianUserID: custodianID, Status: domainsample.StatusPending,
	})
	require.NoError(t, err)

	byBarcode, err := repo.FindByBarcodeID(ctx, "BC-INT-001")
	require.NoError(t, err)
	assert.Equal(t, "SMP-B1", byBarcode.ID)
	assert.Equal(t, "notes", byBarcode.Description)

	_, err = repo.FindByBarcodeID(ctx, "nope")
	assert.ErrorIs(t, err, shared.ErrNotFound)

	code := "BC-INT-001"
	filtered, err := repo.List(ctx, portsample.ListFilter{BarcodeID: &code})
	require.NoError(t, err)
	assert.Len(t, filtered, 1)

	locText := "integration"
	byLoc, err := repo.List(ctx, portsample.ListFilter{LocationText: &locText})
	require.NoError(t, err)
	assert.Len(t, byLoc, 1)
	assert.Equal(t, "SMP-B1", byLoc[0].ID)

	updated, err := repo.UpdateBarcodeID(ctx, "SMP-B2", strPtrI("BC-INT-002"))
	require.NoError(t, err)
	assert.Equal(t, "BC-INT-002", *updated.BarcodeID)
}

func strPtrI(s string) *string { return &s }

func TestSampleRepository_BoxCellsAndMoves(t *testing.T) {
	db := pgtest.SetupPostgres(t)
	repo := sample.New(db)
	ctx := context.Background()

	custodianID := seedCustodian(t, db)
	require.NoError(t, db.Exec(
		`INSERT INTO locations (id, name, kind, level_type, barcode_code, rows, cols)
		 VALUES ('BOX-INT-1', 'Cryobox-1', 'sample_storage', 'box', 'LOC-BC-95001', 8, 12)`,
	).Error)
	boxID := "BOX-INT-1"

	mk := func(id string) {
		_, err := repo.Create(ctx, domainsample.Sample{
			ID: id, Name: id, Type: domainsample.TypeBlood,
			CustodianUserID: custodianID, Status: domainsample.StatusPending,
		})
		require.NoError(t, err)
	}
	mk("SMP-C1")
	mk("SMP-C2")

	_, err := repo.UpdateLocation(ctx, "SMP-C1", &boxID, strPtrI("A1"))
	require.NoError(t, err)
	_, err = repo.UpdateLocation(ctx, "SMP-C2", &boxID, strPtrI("A2"))
	require.NoError(t, err)

	occupied, err := repo.ExistsActiveByLocationPosition(ctx, boxID, "A1")
	require.NoError(t, err)
	assert.True(t, occupied)

	// A second active sample cannot take A1 - the partial unique index bites.
	mk("SMP-C3")
	_, err = repo.UpdateLocation(ctx, "SMP-C3", &boxID, strPtrI("A1"))
	assert.ErrorIs(t, err, shared.ErrConflict)

	inBox, err := repo.List(ctx, portsample.ListFilter{LocationID: &boxID})
	require.NoError(t, err)
	assert.Len(t, inBox, 2)

	// Atomic swap A1 <-> A2.
	moves := []portsample.PositionAssignment{
		{SampleID: "SMP-C1", Position: "A2"},
		{SampleID: "SMP-C2", Position: "A1"},
	}
	after, err := repo.MoveWithinBox(ctx, boxID, moves)
	require.NoError(t, err)
	assert.Len(t, after, 2)

	c1, err := repo.FindByID(ctx, "SMP-C1")
	require.NoError(t, err)
	assert.Equal(t, "A2", *c1.Position)
}

func TestSampleRepository_NotFound(t *testing.T) {
	db := pgtest.SetupPostgres(t)
	repo := sample.New(db)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, "does-not-exist")
	assert.ErrorIs(t, err, shared.ErrNotFound)
}
