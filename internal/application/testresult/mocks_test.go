package testresult_test

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/notification"
	domainsample "github.com/efangly/thanes-lims-backend/internal/domain/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/testresult"
	portsample "github.com/efangly/thanes-lims-backend/internal/ports/sample"
	porttestresult "github.com/efangly/thanes-lims-backend/internal/ports/testresult"
	"github.com/stretchr/testify/mock"
)

type mockResultRepo struct{ mock.Mock }

func (m *mockResultRepo) Create(ctx context.Context, t testresult.TestResult) (testresult.TestResult, error) {
	args := m.Called(ctx, t)
	return args.Get(0).(testresult.TestResult), args.Error(1)
}
func (m *mockResultRepo) FindByID(ctx context.Context, id string) (testresult.TestResult, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(testresult.TestResult), args.Error(1)
}
func (m *mockResultRepo) List(ctx context.Context, filter porttestresult.ListFilter) ([]testresult.TestResult, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]testresult.TestResult), args.Error(1)
}
func (m *mockResultRepo) Update(ctx context.Context, t testresult.TestResult) (testresult.TestResult, error) {
	args := m.Called(ctx, t)
	return args.Get(0).(testresult.TestResult), args.Error(1)
}

type mockSampleRepo struct{ mock.Mock }

func (m *mockSampleRepo) Create(ctx context.Context, s domainsample.Sample) (domainsample.Sample, error) {
	args := m.Called(ctx, s)
	return args.Get(0).(domainsample.Sample), args.Error(1)
}
func (m *mockSampleRepo) FindByID(ctx context.Context, id string) (domainsample.Sample, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domainsample.Sample), args.Error(1)
}
func (m *mockSampleRepo) List(ctx context.Context, filter portsample.ListFilter) ([]domainsample.Sample, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]domainsample.Sample), args.Error(1)
}
func (m *mockSampleRepo) UpdateStatus(ctx context.Context, s domainsample.Sample) (domainsample.Sample, error) {
	args := m.Called(ctx, s)
	return args.Get(0).(domainsample.Sample), args.Error(1)
}

type mockCoCRepo struct{ mock.Mock }

func (m *mockCoCRepo) AppendStep(ctx context.Context, step domainsample.CoCStep) (domainsample.CoCStep, error) {
	args := m.Called(ctx, step)
	return args.Get(0).(domainsample.CoCStep), args.Error(1)
}
func (m *mockCoCRepo) ListBySample(ctx context.Context, sampleID string) ([]domainsample.CoCStep, error) {
	args := m.Called(ctx, sampleID)
	return args.Get(0).([]domainsample.CoCStep), args.Error(1)
}

type mockIDGen struct{ mock.Mock }

func (m *mockIDGen) Next(ctx context.Context, scope string, year *int) (int64, error) {
	args := m.Called(ctx, scope, year)
	return args.Get(0).(int64), args.Error(1)
}

type mockNotifier struct{ mock.Mock }

func (m *mockNotifier) Notify(ctx context.Context, n notification.Notification) error {
	args := m.Called(ctx, n)
	return args.Error(0)
}
