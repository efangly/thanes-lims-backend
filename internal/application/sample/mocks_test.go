package sample_test

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/sample"
	portsample "github.com/efangly/thanes-lims-backend/internal/ports/sample"
	"github.com/stretchr/testify/mock"
)

type mockSampleRepo struct{ mock.Mock }

func (m *mockSampleRepo) Create(ctx context.Context, s sample.Sample) (sample.Sample, error) {
	args := m.Called(ctx, s)
	return args.Get(0).(sample.Sample), args.Error(1)
}
func (m *mockSampleRepo) FindByID(ctx context.Context, id string) (sample.Sample, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(sample.Sample), args.Error(1)
}
func (m *mockSampleRepo) List(ctx context.Context, filter portsample.ListFilter) ([]sample.Sample, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]sample.Sample), args.Error(1)
}
func (m *mockSampleRepo) UpdateStatus(ctx context.Context, s sample.Sample) (sample.Sample, error) {
	args := m.Called(ctx, s)
	return args.Get(0).(sample.Sample), args.Error(1)
}

type mockCoCRepo struct{ mock.Mock }

func (m *mockCoCRepo) AppendStep(ctx context.Context, step sample.CoCStep) (sample.CoCStep, error) {
	args := m.Called(ctx, step)
	return args.Get(0).(sample.CoCStep), args.Error(1)
}
func (m *mockCoCRepo) ListBySample(ctx context.Context, sampleID string) ([]sample.CoCStep, error) {
	args := m.Called(ctx, sampleID)
	return args.Get(0).([]sample.CoCStep), args.Error(1)
}

type mockIDGen struct{ mock.Mock }

func (m *mockIDGen) Next(ctx context.Context, scope string, year *int) (int64, error) {
	args := m.Called(ctx, scope, year)
	return args.Get(0).(int64), args.Error(1)
}
