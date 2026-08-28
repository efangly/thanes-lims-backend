package equipment_test

import (
	"context"
	"errors"
	"testing"
	"time"

	appequipment "github.com/efangly/thanes-lims-backend/internal/application/equipment"
	domainequipment "github.com/efangly/thanes-lims-backend/internal/domain/equipment"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptr(s string) *string { return &s }

func TestUpdateEquipment_PartialLeavesUnsetFields(t *testing.T) {
	repo := newStubRepo()
	vid := "VEN-1"
	repo.store["EQ-PCR-001"] = domainequipment.Equipment{
		ID: "EQ-PCR-001", Name: "old", Category: "keep", VendorID: &vid,
	}
	uc := appequipment.NewUpdateEquipmentUseCase(repo, nil, nil)

	e, err := uc.Execute(context.Background(), appequipment.UpdateEquipmentInput{
		ID:   "EQ-PCR-001",
		Name: ptr("new"),
	})
	require.NoError(t, err)
	assert.Equal(t, "new", e.Name)
	assert.Equal(t, "keep", e.Category)
	require.NotNil(t, e.VendorID)
	assert.Equal(t, "VEN-1", *e.VendorID)
}

func TestUpdateEquipment_ClearsVendorAndInstallDate(t *testing.T) {
	repo := newStubRepo()
	vid := "VEN-1"
	now := time.Now()
	repo.store["EQ-1"] = domainequipment.Equipment{ID: "EQ-1", VendorID: &vid, InstallationDate: &now}
	uc := appequipment.NewUpdateEquipmentUseCase(repo, nil, nil)

	e, err := uc.Execute(context.Background(), appequipment.UpdateEquipmentInput{
		ID:               "EQ-1",
		VendorID:         ptr(""),
		ClearInstallDate: true,
	})
	require.NoError(t, err)
	assert.Nil(t, e.VendorID)
	assert.Nil(t, e.InstallationDate)
}

func TestUpdateEquipment_NotFound(t *testing.T) {
	uc := appequipment.NewUpdateEquipmentUseCase(newStubRepo(), nil, nil)
	_, err := uc.Execute(context.Background(), appequipment.UpdateEquipmentInput{ID: "nope"})
	assert.True(t, errors.Is(err, shared.ErrNotFound))
}
