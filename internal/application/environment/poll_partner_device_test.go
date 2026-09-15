package environment_test

import (
	"bytes"
	"context"
	"encoding/gob"
	"testing"
	"time"

	applicationenvironment "github.com/efangly/thanes-lims-backend/internal/application/environment"
	"github.com/efangly/thanes-lims-backend/internal/domain/environment"
	portenvironment "github.com/efangly/thanes-lims-backend/internal/ports/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newEvaluateThresholds(gauges *mockGaugeRepo, alerts *mockAlertRepo) *applicationenvironment.EvaluateThresholdsUseCase {
	notifier := new(mockNotifier)
	notifier.On("Notify", mock.Anything, mock.Anything).Return(nil)
	broadcaster := new(mockBroadcaster)
	broadcaster.On("Broadcast", mock.Anything).Return()
	return applicationenvironment.NewEvaluateThresholdsUseCase(gauges, alerts, notifier, broadcaster)
}

func TestPollPartnerDevice_SuccessCachesAndBroadcasts(t *testing.T) {
	device := environment.PartnerDevice{Serial: "SN-00042", Location: "ward-3", Active: true}

	client := new(mockPartnerAPIClient)
	client.On("FetchMetadata", mock.Anything, "SN-00042").
		Return(portenvironment.PartnerDeviceMetadata{Serial: "SN-00042", Name: "Fridge", Status: true, Online: true}, nil)
	client.On("FetchLatestReading", mock.Anything, "SN-00042").
		Return(portenvironment.PartnerDeviceReading{Serial: "SN-00042", TempDisplay: 4.8, SendTime: time.Now()}, true, nil)

	gauges := new(mockGaugeRepo)
	gauges.On("FindByLocation", mock.Anything, "ward-3").Return(environment.Gauge{Location: "ward-3", RangeMin: 2, RangeMax: 8}, nil)
	alerts := new(mockAlertRepo)
	alerts.On("FindOpenByLocation", mock.Anything, "ward-3").Return(nil, nil)
	evaluate := newEvaluateThresholds(gauges, alerts)

	cache := new(mockCache)
	cache.On("Set", mock.Anything, "env:partnerdevice:snapshot:SN-00042", mock.Anything, mock.Anything).Return(nil)

	broadcaster := new(mockPartnerDeviceBroadcaster)
	broadcaster.On("Broadcast", mock.MatchedBy(func(s environment.PartnerDeviceSnapshot) bool {
		return s.Serial == "SN-00042" && !s.Stale && s.Level == environment.LevelOK
	})).Return()

	uc := applicationenvironment.NewPollPartnerDeviceUseCase(client, cache, evaluate, broadcaster, 45*time.Second, 5*time.Minute)
	snap, err := uc.Execute(context.Background(), device)

	assert.NoError(t, err)
	assert.False(t, snap.Stale)
	assert.Equal(t, environment.LevelOK, snap.Level)
	cache.AssertCalled(t, "Set", mock.Anything, "env:partnerdevice:snapshot:SN-00042", mock.Anything, mock.Anything)
	broadcaster.AssertExpectations(t)
}

func TestPollPartnerDevice_RetryableFailureFallsBackToStaleCache(t *testing.T) {
	device := environment.PartnerDevice{Serial: "SN-00042", Location: "ward-3", Active: true}

	client := new(mockPartnerAPIClient)
	client.On("FetchMetadata", mock.Anything, "SN-00042").
		Return(portenvironment.PartnerDeviceMetadata{}, &mockRetryableError{msg: "rate limited", retryable: true})

	cached := environment.PartnerDeviceSnapshot{Serial: "SN-00042", Location: "ward-3", Level: environment.LevelOK, FetchedAt: time.Now()}
	var buf bytes.Buffer
	assert.NoError(t, gob.NewEncoder(&buf).Encode(cached))

	cache := new(mockCache)
	cache.On("Get", mock.Anything, "env:partnerdevice:snapshot:SN-00042").Return(buf.Bytes(), nil)

	evaluate := newEvaluateThresholds(new(mockGaugeRepo), new(mockAlertRepo))
	broadcaster := new(mockPartnerDeviceBroadcaster)
	broadcaster.On("Broadcast", mock.MatchedBy(func(s environment.PartnerDeviceSnapshot) bool {
		return s.Serial == "SN-00042" && s.Stale
	})).Return()

	uc := applicationenvironment.NewPollPartnerDeviceUseCase(client, cache, evaluate, broadcaster, 45*time.Second, 5*time.Minute)
	snap, err := uc.Execute(context.Background(), device)

	assert.NoError(t, err)
	assert.True(t, snap.Stale)
	broadcaster.AssertExpectations(t)
}

func TestPollPartnerDevice_PermanentFailureReturnsError(t *testing.T) {
	device := environment.PartnerDevice{Serial: "SN-00042", Location: "ward-3", Active: true}

	client := new(mockPartnerAPIClient)
	client.On("FetchMetadata", mock.Anything, "SN-00042").
		Return(portenvironment.PartnerDeviceMetadata{}, &mockRetryableError{msg: "not found", retryable: false})

	cache := new(mockCache)
	evaluate := newEvaluateThresholds(new(mockGaugeRepo), new(mockAlertRepo))
	broadcaster := new(mockPartnerDeviceBroadcaster)

	uc := applicationenvironment.NewPollPartnerDeviceUseCase(client, cache, evaluate, broadcaster, 45*time.Second, 5*time.Minute)
	_, err := uc.Execute(context.Background(), device)

	assert.Error(t, err)
	cache.AssertNotCalled(t, "Get", mock.Anything, mock.Anything)
	broadcaster.AssertNotCalled(t, "Broadcast", mock.Anything)
}
