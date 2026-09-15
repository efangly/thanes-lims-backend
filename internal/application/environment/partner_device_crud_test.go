package environment_test

import (
	"context"
	"testing"

	applicationenvironment "github.com/efangly/thanes-lims-backend/internal/application/environment"
	"github.com/efangly/thanes-lims-backend/internal/domain/environment"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreatePartnerDevice_RejectsUnknownLocation(t *testing.T) {
	devices := new(mockPartnerDeviceRepo)
	gauges := new(mockGaugeRepo)
	gauges.On("FindByLocation", mock.Anything, "ward-3").Return(environment.Gauge{}, shared.ErrNotFound)

	uc := applicationenvironment.NewCreatePartnerDeviceUseCase(devices, gauges)
	_, err := uc.Execute(context.Background(), applicationenvironment.CreatePartnerDeviceInput{
		Serial: "SN-00042", Location: "ward-3", Active: true,
	})

	assert.ErrorIs(t, err, shared.ErrNotFound)
	devices.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestCreatePartnerDevice_RejectsDuplicateSerial(t *testing.T) {
	devices := new(mockPartnerDeviceRepo)
	gauges := new(mockGaugeRepo)
	gauges.On("FindByLocation", mock.Anything, "ward-3").Return(environment.Gauge{Location: "ward-3"}, nil)
	devices.On("FindBySerial", mock.Anything, "SN-00042").Return(environment.PartnerDevice{Serial: "SN-00042"}, nil)

	uc := applicationenvironment.NewCreatePartnerDeviceUseCase(devices, gauges)
	_, err := uc.Execute(context.Background(), applicationenvironment.CreatePartnerDeviceInput{
		Serial: "SN-00042", Location: "ward-3", Active: true,
	})

	assert.ErrorIs(t, err, shared.ErrConflict)
}

func TestCreatePartnerDevice_Succeeds(t *testing.T) {
	devices := new(mockPartnerDeviceRepo)
	gauges := new(mockGaugeRepo)
	gauges.On("FindByLocation", mock.Anything, "ward-3").Return(environment.Gauge{Location: "ward-3"}, nil)
	devices.On("FindBySerial", mock.Anything, "SN-00042").Return(environment.PartnerDevice{}, shared.ErrNotFound)
	devices.On("Create", mock.Anything, environment.PartnerDevice{Serial: "SN-00042", Location: "ward-3", Active: true}).
		Return(environment.PartnerDevice{Serial: "SN-00042", Location: "ward-3", Active: true}, nil)

	uc := applicationenvironment.NewCreatePartnerDeviceUseCase(devices, gauges)
	d, err := uc.Execute(context.Background(), applicationenvironment.CreatePartnerDeviceInput{
		Serial: "SN-00042", Location: "ward-3", Active: true,
	})

	assert.NoError(t, err)
	assert.Equal(t, "SN-00042", d.Serial)
}

func TestUpdatePartnerDevice_RejectsUnknownLocation(t *testing.T) {
	devices := new(mockPartnerDeviceRepo)
	gauges := new(mockGaugeRepo)
	devices.On("FindBySerial", mock.Anything, "SN-00042").Return(environment.PartnerDevice{Serial: "SN-00042", Location: "ward-3", Active: true}, nil)
	gauges.On("FindByLocation", mock.Anything, "ward-9").Return(environment.Gauge{}, shared.ErrNotFound)

	uc := applicationenvironment.NewUpdatePartnerDeviceUseCase(devices, gauges)
	_, err := uc.Execute(context.Background(), applicationenvironment.UpdatePartnerDeviceInput{
		Serial: "SN-00042", Location: "ward-9", Active: false,
	})

	assert.ErrorIs(t, err, shared.ErrNotFound)
	devices.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}
