package inventory_test

import (
	"context"
	"testing"

	applicationinventory "github.com/efangly/thanes-lims-backend/internal/application/inventory"
	"github.com/efangly/thanes-lims-backend/internal/domain/inventory"
	domainlocation "github.com/efangly/thanes-lims-backend/internal/domain/location"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/efangly/thanes-lims-backend/internal/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func strPtr(s string) *string { return &s }
func i64Ptr(i int64) *int64   { return &i }

func TestUpdateItem_PartialUpdateAndClearVendor(t *testing.T) {
	items := new(mockItemRepo)
	custodians := new(mockCustodianDir)

	existing := inventory.InventoryItem{
		ID: "INV-00001", Name: "old", Category: "c", Unit: "u", Min: 1, Max: 10,
		CustodianUserID: 3, Manufacturer: "OldCo", VendorID: strPtr("VEN-1"),
	}
	items.On("FindByID", mock.Anything, "INV-00001").Return(existing, nil)
	custodians.On("FindByID", mock.Anything, int64(9)).Return(user.User{ID: 9}, nil)
	var saved inventory.InventoryItem
	items.On("Update", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) { saved = args.Get(1).(inventory.InventoryItem) }).
		Return(inventory.InventoryItem{}, nil)

	uc := applicationinventory.NewUpdateItemUseCase(items, custodians, nil, nil)
	_, err := uc.Execute(context.Background(), applicationinventory.UpdateItemInput{
		ID:              "INV-00001",
		Name:            strPtr("new"),
		CustodianUserID: i64Ptr(9),
		Manufacturer:    strPtr(" NewCo "),
		VendorID:        strPtr(""),
	})
	require.NoError(t, err)
	assert.Equal(t, "new", saved.Name)
	assert.Equal(t, "c", saved.Category)
	assert.Equal(t, int64(9), saved.CustodianUserID)
	assert.Equal(t, "NewCo", saved.Manufacturer)
	assert.Nil(t, saved.VendorID)
}

func TestUpdateItem_RejectsUnknownCustodian(t *testing.T) {
	items := new(mockItemRepo)
	custodians := new(mockCustodianDir)
	items.On("FindByID", mock.Anything, "INV-00001").Return(inventory.InventoryItem{ID: "INV-00001"}, nil)
	custodians.On("FindByID", mock.Anything, int64(99)).Return(user.User{}, shared.ErrNotFound)

	uc := applicationinventory.NewUpdateItemUseCase(items, custodians, nil, nil)
	_, err := uc.Execute(context.Background(), applicationinventory.UpdateItemInput{
		ID: "INV-00001", CustodianUserID: i64Ptr(99),
	})
	assert.ErrorIs(t, err, shared.ErrValidation)
	items.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestUpdateItem_RejectsNonEquipmentStorageLocation(t *testing.T) {
	items := new(mockItemRepo)
	locations := new(mockLocationDir)
	items.On("FindByID", mock.Anything, "INV-00001").Return(inventory.InventoryItem{ID: "INV-00001"}, nil)
	locations.On("GetByID", mock.Anything, "LOC-S").Return(domainlocation.Location{ID: "LOC-S", Kind: domainlocation.KindSampleStorage}, nil)

	uc := applicationinventory.NewUpdateItemUseCase(items, nil, nil, locations)
	_, err := uc.Execute(context.Background(), applicationinventory.UpdateItemInput{
		ID: "INV-00001", LocationID: strPtr("LOC-S"),
	})
	assert.ErrorIs(t, err, shared.ErrValidation)
}
