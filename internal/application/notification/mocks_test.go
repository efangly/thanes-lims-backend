package notification_test

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/notification"
	"github.com/stretchr/testify/mock"
)

type mockNotificationRepo struct{ mock.Mock }

func (m *mockNotificationRepo) Create(ctx context.Context, n notification.Notification) (notification.Notification, error) {
	args := m.Called(ctx, n)
	return args.Get(0).(notification.Notification), args.Error(1)
}
func (m *mockNotificationRepo) FindByID(ctx context.Context, id string) (notification.Notification, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(notification.Notification), args.Error(1)
}
func (m *mockNotificationRepo) ListForUser(ctx context.Context, userID int64) ([]notification.Notification, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]notification.Notification), args.Error(1)
}
func (m *mockNotificationRepo) MarkRead(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockNotificationRepo) MarkAllRead(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

type mockIDGen struct{ mock.Mock }

func (m *mockIDGen) Next(ctx context.Context, scope string, year *int) (int64, error) {
	args := m.Called(ctx, scope, year)
	return args.Get(0).(int64), args.Error(1)
}
