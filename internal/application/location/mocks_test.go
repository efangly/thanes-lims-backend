package location_test

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/location"
	"github.com/stretchr/testify/mock"
)

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
func (m *mockLocationRepo) UpdateGrid(ctx context.Context, id string, rows, cols int) (location.Location, error) {
	args := m.Called(ctx, id, rows, cols)
	return args.Get(0).(location.Location), args.Error(1)
}
func (m *mockLocationRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockLocationRepo) FullPath(ctx context.Context, id string) (string, error) {
	args := m.Called(ctx, id)
	return args.String(0), args.Error(1)
}

type mockIDGen struct{ mock.Mock }

func (m *mockIDGen) Next(ctx context.Context, scope string, year *int) (int64, error) {
	args := m.Called(ctx, scope, year)
	return args.Get(0).(int64), args.Error(1)
}
