package sample_test

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/location"
	"github.com/efangly/thanes-lims-backend/internal/domain/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/user"
	portsample "github.com/efangly/thanes-lims-backend/internal/ports/sample"
	"github.com/stretchr/testify/mock"
)

type mockCustodianDir struct{ mock.Mock }

func (m *mockCustodianDir) FindByID(ctx context.Context, id int64) (user.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(user.User), args.Error(1)
}

type mockSampleRepo struct{ mock.Mock }

func (m *mockSampleRepo) Create(ctx context.Context, s sample.Sample) (sample.Sample, error) {
	args := m.Called(ctx, s)
	return args.Get(0).(sample.Sample), args.Error(1)
}
func (m *mockSampleRepo) FindByID(ctx context.Context, id string) (sample.Sample, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(sample.Sample), args.Error(1)
}
func (m *mockSampleRepo) FindByBarcodeID(ctx context.Context, barcodeID string) (sample.Sample, error) {
	args := m.Called(ctx, barcodeID)
	return args.Get(0).(sample.Sample), args.Error(1)
}
func (m *mockSampleRepo) UpdateBarcodeID(ctx context.Context, sampleID string, barcodeID *string) (sample.Sample, error) {
	args := m.Called(ctx, sampleID, barcodeID)
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
func (m *mockSampleRepo) UpdateLocation(ctx context.Context, sampleID string, locationID *string) (sample.Sample, error) {
	args := m.Called(ctx, sampleID, locationID)
	return args.Get(0).(sample.Sample), args.Error(1)
}
func (m *mockSampleRepo) ExistsActiveByLocation(ctx context.Context, locationID string) (bool, error) {
	args := m.Called(ctx, locationID)
	return args.Bool(0), args.Error(1)
}
func (m *mockSampleRepo) ExistsByLocation(ctx context.Context, locationID string) (bool, error) {
	args := m.Called(ctx, locationID)
	return args.Bool(0), args.Error(1)
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

type mockLocationRepo struct{ mock.Mock }

func (m *mockLocationRepo) Create(ctx context.Context, l location.Location) (location.Location, error) {
	args := m.Called(ctx, l)
	return args.Get(0).(location.Location), args.Error(1)
}
func (m *mockLocationRepo) CreateMany(ctx context.Context, ls []location.Location) ([]location.Location, error) {
	args := m.Called(ctx, ls)
	return args.Get(0).([]location.Location), args.Error(1)
}
func (m *mockLocationRepo) GetByID(ctx context.Context, id string) (location.Location, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(location.Location), args.Error(1)
}
func (m *mockLocationRepo) ListChildren(ctx context.Context, parentID *string) ([]location.Location, error) {
	args := m.Called(ctx, parentID)
	return args.Get(0).([]location.Location), args.Error(1)
}
func (m *mockLocationRepo) ListRoots(ctx context.Context, kind location.Kind) ([]location.Location, error) {
	args := m.Called(ctx, kind)
	return args.Get(0).([]location.Location), args.Error(1)
}
func (m *mockLocationRepo) FindByBarcode(ctx context.Context, code string) (location.Location, error) {
	args := m.Called(ctx, code)
	return args.Get(0).(location.Location), args.Error(1)
}
func (m *mockLocationRepo) FindChildByName(ctx context.Context, parentID *string, name string) (location.Location, error) {
	args := m.Called(ctx, parentID, name)
	return args.Get(0).(location.Location), args.Error(1)
}
func (m *mockLocationRepo) HasChildren(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}
func (m *mockLocationRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockLocationRepo) FullPath(ctx context.Context, id string) (string, error) {
	args := m.Called(ctx, id)
	return args.String(0), args.Error(1)
}
