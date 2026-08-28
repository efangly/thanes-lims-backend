package equipment_test

import (
	"context"
	"errors"
	"testing"
	"time"

	appequipment "github.com/efangly/thanes-lims-backend/internal/application/equipment"
	domainlocation "github.com/efangly/thanes-lims-backend/internal/domain/location"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateEquipment_PersistsAssetFields(t *testing.T) {
	repo := newStubRepo()
	vendors := stubVendors{known: map[string]bool{"VEN-00001": true}}
	locs := stubLocations{byID: map[string]domainlocation.Location{
		"LOC-1": {ID: "LOC-1", Kind: domainlocation.KindEquipmentStorage},
	}}
	uc := appequipment.NewCreateEquipmentUseCase(repo, &stubIDGen{}, vendors, locs)

	install := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)
	e, err := uc.Execute(context.Background(), appequipment.CreateEquipmentInput{
		Name:               "เครื่อง PCR",
		TypeCode:           "PCR",
		NextCalibrationDue: time.Now().AddDate(0, 1, 0),
		SerialNumber:       "  SN-42 ",
		Category:           "Molecular",
		Manufacturer:       "Acme",
		Model:              "X1",
		InstallationDate:   &install,
		VendorID:           "VEN-00001",
		LocationID:         "LOC-1",
	})
	require.NoError(t, err)
	assert.Equal(t, "EQ-PCR-001", e.ID)
	assert.Equal(t, "SN-42", e.SerialNumber)
	assert.Equal(t, "Molecular", e.Category)
	require.NotNil(t, e.VendorID)
	assert.Equal(t, "VEN-00001", *e.VendorID)
	require.NotNil(t, e.LocationID)
	assert.Equal(t, "LOC-1", *e.LocationID)
}

func TestCreateEquipment_RejectsUnknownVendor(t *testing.T) {
	uc := appequipment.NewCreateEquipmentUseCase(newStubRepo(), &stubIDGen{}, stubVendors{}, stubLocations{})
	_, err := uc.Execute(context.Background(), appequipment.CreateEquipmentInput{
		Name: "x", TypeCode: "PCR", NextCalibrationDue: time.Now(), VendorID: "VEN-999",
	})
	assert.True(t, errors.Is(err, shared.ErrValidation))
}

func TestCreateEquipment_RejectsNonEquipmentStorageLocation(t *testing.T) {
	locs := stubLocations{byID: map[string]domainlocation.Location{
		"LOC-S": {ID: "LOC-S", Kind: domainlocation.KindSampleStorage},
	}}
	uc := appequipment.NewCreateEquipmentUseCase(newStubRepo(), &stubIDGen{}, stubVendors{}, locs)
	_, err := uc.Execute(context.Background(), appequipment.CreateEquipmentInput{
		Name: "x", TypeCode: "PCR", NextCalibrationDue: time.Now(), LocationID: "LOC-S",
	})
	assert.True(t, errors.Is(err, shared.ErrValidation))
}

func TestCreateEquipment_NilDirectoriesSkipValidation(t *testing.T) {
	uc := appequipment.NewCreateEquipmentUseCase(newStubRepo(), &stubIDGen{}, nil, nil)
	e, err := uc.Execute(context.Background(), appequipment.CreateEquipmentInput{
		Name: "x", TypeCode: "PCR", NextCalibrationDue: time.Now(), VendorID: "whatever",
	})
	require.NoError(t, err)
	assert.Equal(t, "whatever", *e.VendorID)
}
