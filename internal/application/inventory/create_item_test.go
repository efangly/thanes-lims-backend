package inventory_test

import (
	"context"
	"testing"

	applicationinventory "github.com/efangly/thanes-lims-backend/internal/application/inventory"
	"github.com/efangly/thanes-lims-backend/internal/domain/inventory"
	domainlocation "github.com/efangly/thanes-lims-backend/internal/domain/location"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/efangly/thanes-lims-backend/internal/domain/user"
	"github.com/efangly/thanes-lims-backend/internal/domain/vendor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateItem_PersistsAssetFields(t *testing.T) {
	items := new(mockItemRepo)
	custodians := new(mockCustodianDir)
	vendors := new(mockVendorDir)
	locations := new(mockLocationDir)

	custodians.On("FindByID", mock.Anything, int64(7)).Return(user.User{ID: 7}, nil)
	vendors.On("FindByID", mock.Anything, "VEN-00001").Return(vendor.Vendor{ID: "VEN-00001"}, nil)
	locations.On("GetByID", mock.Anything, "LOC-1").Return(domainlocation.Location{ID: "LOC-1", Kind: domainlocation.KindEquipmentStorage}, nil)
	var saved inventory.InventoryItem
	items.On("Create", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) { saved = args.Get(1).(inventory.InventoryItem) }).
		Return(inventory.InventoryItem{}, nil)

	uc := applicationinventory.NewCreateItemUseCase(items, stubIDGen{}, custodians, vendors, locations)
	_, err := uc.Execute(context.Background(), applicationinventory.CreateItemInput{
		Name: "เอทานอล", Category: "รีเอเจนต์", Unit: "ลิตร",
		CustodianUserID: 7, Manufacturer: "  Acme ", VendorID: "VEN-00001", LocationID: "LOC-1",
	})
	require.NoError(t, err)
	assert.Equal(t, "INV-00001", saved.ID)
	assert.Equal(t, int64(7), saved.CustodianUserID)
	assert.Equal(t, "Acme", saved.Manufacturer)
	require.NotNil(t, saved.VendorID)
	assert.Equal(t, "VEN-00001", *saved.VendorID)
	require.NotNil(t, saved.LocationID)
	assert.Equal(t, "LOC-1", *saved.LocationID)
}

func TestCreateItem_RejectsMissingCustodian(t *testing.T) {
	items := new(mockItemRepo)
	custodians := new(mockCustodianDir)

	uc := applicationinventory.NewCreateItemUseCase(items, stubIDGen{}, custodians, nil, nil)
	_, err := uc.Execute(context.Background(), applicationinventory.CreateItemInput{
		Name: "x", Category: "y", Unit: "z",
	})
	assert.ErrorIs(t, err, shared.ErrValidation)
	items.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestCreateItem_RejectsUnknownCustodian(t *testing.T) {
	items := new(mockItemRepo)
	custodians := new(mockCustodianDir)
	custodians.On("FindByID", mock.Anything, int64(99)).Return(user.User{}, shared.ErrNotFound)

	uc := applicationinventory.NewCreateItemUseCase(items, stubIDGen{}, custodians, nil, nil)
	_, err := uc.Execute(context.Background(), applicationinventory.CreateItemInput{
		Name: "x", Category: "y", Unit: "z", CustodianUserID: 99,
	})
	assert.ErrorIs(t, err, shared.ErrValidation)
}

func TestCreateItem_RejectsNonEquipmentStorageLocation(t *testing.T) {
	items := new(mockItemRepo)
	custodians := new(mockCustodianDir)
	locations := new(mockLocationDir)
	custodians.On("FindByID", mock.Anything, int64(7)).Return(user.User{ID: 7}, nil)
	locations.On("GetByID", mock.Anything, "LOC-S").Return(domainlocation.Location{ID: "LOC-S", Kind: domainlocation.KindSampleStorage}, nil)

	uc := applicationinventory.NewCreateItemUseCase(items, stubIDGen{}, custodians, nil, locations)
	_, err := uc.Execute(context.Background(), applicationinventory.CreateItemInput{
		Name: "x", Category: "y", Unit: "z", CustodianUserID: 7, LocationID: "LOC-S",
	})
	assert.ErrorIs(t, err, shared.ErrValidation)
}
