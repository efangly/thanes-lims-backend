package environment_test

import (
	"context"
	"testing"

	applicationenvironment "github.com/efangly/thanes-lims-backend/internal/application/environment"
	"github.com/efangly/thanes-lims-backend/internal/domain/environment"
	"github.com/efangly/thanes-lims-backend/internal/domain/notification"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEvaluateThresholds_OKClearsExistingAlert(t *testing.T) {
	gauges := new(mockGaugeRepo)
	alerts := new(mockAlertRepo)
	notifier := new(mockNotifier)
	broadcaster := new(mockBroadcaster)

	gauges.On("FindByLocation", mock.Anything, "Fridge-A").Return(environment.Gauge{Location: "Fridge-A", RangeMin: 2, RangeMax: 8}, nil)
	existing := &environment.EnvAlert{ID: 1, Location: "Fridge-A"}
	alerts.On("FindOpenByLocation", mock.Anything, "Fridge-A").Return(existing, nil)
	alerts.On("Resolve", mock.Anything, int64(1)).Return(nil)

	uc := applicationenvironment.NewEvaluateThresholdsUseCase(gauges, alerts, notifier, broadcaster)
	alert, err := uc.Execute(context.Background(), "Fridge-A", 5)

	assert.NoError(t, err)
	assert.Nil(t, alert)
	alerts.AssertCalled(t, "Resolve", mock.Anything, int64(1))
	notifier.AssertNotCalled(t, "Notify", mock.Anything, mock.Anything)
	broadcaster.AssertNotCalled(t, "Broadcast", mock.Anything)
}

func TestEvaluateThresholds_CreatesNewAlertWhenNoneOpen(t *testing.T) {
	gauges := new(mockGaugeRepo)
	alerts := new(mockAlertRepo)
	notifier := new(mockNotifier)
	broadcaster := new(mockBroadcaster)

	gauges.On("FindByLocation", mock.Anything, "Fridge-A").Return(environment.Gauge{Location: "Fridge-A", RangeMin: 2, RangeMax: 8}, nil)
	alerts.On("FindOpenByLocation", mock.Anything, "Fridge-A").Return(nil, nil)
	created := environment.EnvAlert{ID: 2, Location: "Fridge-A", Level: environment.LevelCrit}
	alerts.On("Create", mock.Anything, mock.MatchedBy(func(a environment.EnvAlert) bool {
		return a.Location == "Fridge-A" && a.Level == environment.LevelCrit
	})).Return(created, nil)
	notifier.On("Notify", mock.Anything, mock.MatchedBy(func(n notification.Notification) bool {
		return n.Tone == notification.ToneRed
	})).Return(nil)
	broadcaster.On("Broadcast", created)

	uc := applicationenvironment.NewEvaluateThresholdsUseCase(gauges, alerts, notifier, broadcaster)
	alert, err := uc.Execute(context.Background(), "Fridge-A", 20)

	assert.NoError(t, err)
	assert.NotNil(t, alert)
	assert.Equal(t, int64(2), alert.ID)
	notifier.AssertCalled(t, "Notify", mock.Anything, mock.Anything)
	broadcaster.AssertCalled(t, "Broadcast", created)
}

func TestEvaluateThresholds_DedupsWhileAlertOpen(t *testing.T) {
	gauges := new(mockGaugeRepo)
	alerts := new(mockAlertRepo)
	notifier := new(mockNotifier)
	broadcaster := new(mockBroadcaster)

	gauges.On("FindByLocation", mock.Anything, "Fridge-A").Return(environment.Gauge{Location: "Fridge-A", RangeMin: 2, RangeMax: 8}, nil)
	existing := &environment.EnvAlert{ID: 3, Location: "Fridge-A", Level: environment.LevelCrit}
	alerts.On("FindOpenByLocation", mock.Anything, "Fridge-A").Return(existing, nil)

	uc := applicationenvironment.NewEvaluateThresholdsUseCase(gauges, alerts, notifier, broadcaster)
	alert, err := uc.Execute(context.Background(), "Fridge-A", 25)

	assert.NoError(t, err)
	assert.Equal(t, int64(3), alert.ID)
	alerts.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	notifier.AssertNotCalled(t, "Notify", mock.Anything, mock.Anything)
	broadcaster.AssertNotCalled(t, "Broadcast", mock.Anything)
}
