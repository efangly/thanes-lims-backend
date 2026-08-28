package document_test

import (
	"context"
	"io"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/document"
	"github.com/stretchr/testify/mock"
)

type mockDocRepo struct{ mock.Mock }

func (m *mockDocRepo) Create(ctx context.Context, d document.Document) (document.Document, error) {
	args := m.Called(ctx, d)
	return args.Get(0).(document.Document), args.Error(1)
}
func (m *mockDocRepo) FindByID(ctx context.Context, id string) (document.Document, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(document.Document), args.Error(1)
}
func (m *mockDocRepo) List(ctx context.Context) ([]document.Document, error) {
	args := m.Called(ctx)
	return args.Get(0).([]document.Document), args.Error(1)
}
func (m *mockDocRepo) ListByEquipment(ctx context.Context, equipmentID string) ([]document.Document, error) {
	args := m.Called(ctx, equipmentID)
	return args.Get(0).([]document.Document), args.Error(1)
}
func (m *mockDocRepo) ListByCalibrationEvent(ctx context.Context, calibrationEventID int64) ([]document.Document, error) {
	args := m.Called(ctx, calibrationEventID)
	return args.Get(0).([]document.Document), args.Error(1)
}
func (m *mockDocRepo) Update(ctx context.Context, d document.Document) (document.Document, error) {
	args := m.Called(ctx, d)
	return args.Get(0).(document.Document), args.Error(1)
}

type mockHistoryRepo struct{ mock.Mock }

func (m *mockHistoryRepo) Append(ctx context.Context, h document.DocHistory) (document.DocHistory, error) {
	args := m.Called(ctx, h)
	return args.Get(0).(document.DocHistory), args.Error(1)
}
func (m *mockHistoryRepo) ListByDocument(ctx context.Context, documentID string) ([]document.DocHistory, error) {
	args := m.Called(ctx, documentID)
	return args.Get(0).([]document.DocHistory), args.Error(1)
}

type mockFileStorage struct{ mock.Mock }

func (m *mockFileStorage) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	args := m.Called(ctx, key, reader, size, contentType)
	return args.Error(0)
}
func (m *mockFileStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}
func (m *mockFileStorage) GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	args := m.Called(ctx, key, expiry)
	return args.String(0), args.Error(1)
}
func (m *mockFileStorage) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

type mockIDGen struct{ mock.Mock }

func (m *mockIDGen) Next(ctx context.Context, scope string, year *int) (int64, error) {
	args := m.Called(ctx, scope, year)
	return args.Get(0).(int64), args.Error(1)
}
