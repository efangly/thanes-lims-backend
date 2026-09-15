package environment_test

import (
	"context"
	"errors"
	"testing"

	applicationenvironment "github.com/efangly/thanes-lims-backend/internal/application/environment"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portenvironment "github.com/efangly/thanes-lims-backend/internal/ports/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDiscoverPartnerDevicesByWard_Success(t *testing.T) {
	client := new(mockPartnerAPIClient)
	listing := portenvironment.PartnerDeviceListing{
		Devices: []portenvironment.PartnerDeviceWithReading{
			{Device: portenvironment.PartnerDeviceMetadata{Serial: "SN-00042", Name: "Fridge"}, HasLatest: false},
		},
		Total: 1, Page: 1, Limit: 20,
	}
	client.On("ListDevicesByWard", mock.Anything, "ICU", 1, 20).Return(listing, nil)

	uc := applicationenvironment.NewDiscoverPartnerDevicesByWardUseCase(client)
	got, err := uc.Execute(context.Background(), applicationenvironment.DiscoverPartnerDevicesByWardInput{Ward: "ICU", Page: 1, Limit: 20})

	assert.NoError(t, err)
	assert.Equal(t, listing, got)
}

func TestDiscoverPartnerDevicesByWard_ClampsLimit(t *testing.T) {
	client := new(mockPartnerAPIClient)
	client.On("ListDevicesByWard", mock.Anything, "ICU", 1, 100).Return(portenvironment.PartnerDeviceListing{}, nil)

	uc := applicationenvironment.NewDiscoverPartnerDevicesByWardUseCase(client)
	_, err := uc.Execute(context.Background(), applicationenvironment.DiscoverPartnerDevicesByWardInput{Ward: "ICU", Page: 1, Limit: 500})

	assert.NoError(t, err)
	client.AssertCalled(t, "ListDevicesByWard", mock.Anything, "ICU", 1, 100)
}

func TestDiscoverPartnerDevicesByWard_EmptyWardIsValidationError(t *testing.T) {
	client := new(mockPartnerAPIClient)

	uc := applicationenvironment.NewDiscoverPartnerDevicesByWardUseCase(client)
	_, err := uc.Execute(context.Background(), applicationenvironment.DiscoverPartnerDevicesByWardInput{Ward: "  ", Page: 1, Limit: 20})

	assert.ErrorIs(t, err, shared.ErrValidation)
	client.AssertNotCalled(t, "ListDevicesByWard", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestDiscoverPartnerDevicesByWard_PermanentErrorPassesThrough(t *testing.T) {
	client := new(mockPartnerAPIClient)
	wantErr := errors.New("not found")
	client.On("ListDevicesByWard", mock.Anything, "UNKNOWN", 1, 20).Return(portenvironment.PartnerDeviceListing{}, wantErr)

	uc := applicationenvironment.NewDiscoverPartnerDevicesByWardUseCase(client)
	_, err := uc.Execute(context.Background(), applicationenvironment.DiscoverPartnerDevicesByWardInput{Ward: "UNKNOWN", Page: 1, Limit: 20})

	assert.ErrorIs(t, err, wantErr)
}
