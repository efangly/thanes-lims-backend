package environment_test

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/environment"
	"github.com/efangly/thanes-lims-backend/internal/domain/notification"
	"github.com/stretchr/testify/mock"
)

type mockGaugeRepo struct{ mock.Mock }

func (m *mockGaugeRepo) List(ctx context.Context) ([]environment.Gauge, error) {
	args := m.Called(ctx)
	return args.Get(0).([]environment.Gauge), args.Error(1)
}
func (m *mockGaugeRepo) FindByLocation(ctx context.Context, location string) (environment.Gauge, error) {
	args := m.Called(ctx, location)
	return args.Get(0).(environment.Gauge), args.Error(1)
}

type mockReadingRepo struct{ mock.Mock }

func (m *mockReadingRepo) Record(ctx context.Context, r environment.SensorReading) (environment.SensorReading, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(environment.SensorReading), args.Error(1)
}
func (m *mockReadingRepo) LatestByLocation(ctx context.Context, location string) (environment.SensorReading, error) {
	args := m.Called(ctx, location)
	return args.Get(0).(environment.SensorReading), args.Error(1)
}
func (m *mockReadingRepo) ListTrend(ctx context.Context, location string, limit int) ([]environment.SensorReading, error) {
	args := m.Called(ctx, location, limit)
	return args.Get(0).([]environment.SensorReading), args.Error(1)
}

type mockAlertRepo struct{ mock.Mock }

func (m *mockAlertRepo) Create(ctx context.Context, a environment.EnvAlert) (environment.EnvAlert, error) {
	args := m.Called(ctx, a)
	return args.Get(0).(environment.EnvAlert), args.Error(1)
}
func (m *mockAlertRepo) FindOpenByLocation(ctx context.Context, location string) (*environment.EnvAlert, error) {
	args := m.Called(ctx, location)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*environment.EnvAlert), args.Error(1)
}
func (m *mockAlertRepo) List(ctx context.Context) ([]environment.EnvAlert, error) {
	args := m.Called(ctx)
	return args.Get(0).([]environment.EnvAlert), args.Error(1)
}
func (m *mockAlertRepo) Resolve(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type mockNotifier struct{ mock.Mock }

func (m *mockNotifier) Notify(ctx context.Context, n notification.Notification) error {
	args := m.Called(ctx, n)
	return args.Error(0)
}

type mockBroadcaster struct{ mock.Mock }

func (m *mockBroadcaster) Broadcast(a environment.EnvAlert) {
	m.Called(a)
}
