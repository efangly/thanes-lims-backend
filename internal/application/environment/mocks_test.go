package environment_test

import (
	"context"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/environment"
	"github.com/efangly/thanes-lims-backend/internal/domain/notification"
	portenvironment "github.com/efangly/thanes-lims-backend/internal/ports/environment"
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

type mockPartnerDeviceRepo struct{ mock.Mock }

func (m *mockPartnerDeviceRepo) Create(ctx context.Context, d environment.PartnerDevice) (environment.PartnerDevice, error) {
	args := m.Called(ctx, d)
	return args.Get(0).(environment.PartnerDevice), args.Error(1)
}
func (m *mockPartnerDeviceRepo) Update(ctx context.Context, d environment.PartnerDevice) (environment.PartnerDevice, error) {
	args := m.Called(ctx, d)
	return args.Get(0).(environment.PartnerDevice), args.Error(1)
}
func (m *mockPartnerDeviceRepo) FindBySerial(ctx context.Context, serial string) (environment.PartnerDevice, error) {
	args := m.Called(ctx, serial)
	return args.Get(0).(environment.PartnerDevice), args.Error(1)
}
func (m *mockPartnerDeviceRepo) List(ctx context.Context) ([]environment.PartnerDevice, error) {
	args := m.Called(ctx)
	return args.Get(0).([]environment.PartnerDevice), args.Error(1)
}

type mockPartnerAPIClient struct{ mock.Mock }

func (m *mockPartnerAPIClient) FetchSnapshot(ctx context.Context, serial string) (portenvironment.PartnerDeviceMetadata, portenvironment.PartnerDeviceReading, bool, error) {
	args := m.Called(ctx, serial)
	return args.Get(0).(portenvironment.PartnerDeviceMetadata), args.Get(1).(portenvironment.PartnerDeviceReading), args.Bool(2), args.Error(3)
}
func (m *mockPartnerAPIClient) ListDevicesByWard(ctx context.Context, ward string, page, limit int) (portenvironment.PartnerDeviceListing, error) {
	args := m.Called(ctx, ward, page, limit)
	return args.Get(0).(portenvironment.PartnerDeviceListing), args.Error(1)
}

type mockPartnerDeviceBroadcaster struct{ mock.Mock }

func (m *mockPartnerDeviceBroadcaster) Broadcast(s environment.PartnerDeviceSnapshot) {
	m.Called(s)
}

// mockRetryableError implements portenvironment.RetryableError for tests
// that need to control whether PollPartnerDeviceUseCase treats a failure as
// retryable (stale-cache fallback) or permanent (hard error).
type mockRetryableError struct {
	msg       string
	retryable bool
}

func (e *mockRetryableError) Error() string   { return e.msg }
func (e *mockRetryableError) Retryable() bool { return e.retryable }

type mockCache struct{ mock.Mock }

func (m *mockCache) Get(ctx context.Context, key string) ([]byte, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}
func (m *mockCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}
func (m *mockCache) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}
