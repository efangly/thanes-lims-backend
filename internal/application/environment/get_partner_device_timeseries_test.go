package environment_test

import (
	"context"
	"errors"
	"testing"
	"time"

	applicationenvironment "github.com/efangly/thanes-lims-backend/internal/application/environment"
	portenvironment "github.com/efangly/thanes-lims-backend/internal/ports/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetPartnerDeviceTimeseries_Success(t *testing.T) {
	client := new(mockPartnerAPIClient)
	readings := []portenvironment.PartnerDeviceReading{
		{Serial: "SN-00042", SendTime: time.Now(), TempDisplay: 23.5, HumidityDisplay: 61.5},
		{Serial: "SN-00042", SendTime: time.Now().Add(-5 * time.Minute), TempDisplay: 23.4, HumidityDisplay: 61.0},
	}
	client.On("FetchTimeseries", mock.Anything, "SN-00042").Return(readings, nil)

	uc := applicationenvironment.NewGetPartnerDeviceTimeseriesUseCase(client)
	got, err := uc.Execute(context.Background(), "SN-00042")

	assert.NoError(t, err)
	assert.Equal(t, readings, got)
}

func TestGetPartnerDeviceTimeseries_ErrorPassesThrough(t *testing.T) {
	client := new(mockPartnerAPIClient)
	wantErr := errors.New("not found")
	client.On("FetchTimeseries", mock.Anything, "SN-UNKNOWN").Return(nil, wantErr)

	uc := applicationenvironment.NewGetPartnerDeviceTimeseriesUseCase(client)
	_, err := uc.Execute(context.Background(), "SN-UNKNOWN")

	assert.ErrorIs(t, err, wantErr)
}
