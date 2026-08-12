package notification

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/notification"
	portnotification "github.com/efangly/thanes-lims-backend/internal/ports/notification"
)

type ListNotificationsUseCase struct {
	notifications portnotification.Repository
}

func NewListNotificationsUseCase(notifications portnotification.Repository) *ListNotificationsUseCase {
	return &ListNotificationsUseCase{notifications: notifications}
}

func (uc *ListNotificationsUseCase) Execute(ctx context.Context, userID int64) ([]notification.Notification, error) {
	return uc.notifications.ListForUser(ctx, userID)
}

type MarkReadUseCase struct {
	notifications portnotification.Repository
}

func NewMarkReadUseCase(notifications portnotification.Repository) *MarkReadUseCase {
	return &MarkReadUseCase{notifications: notifications}
}

func (uc *MarkReadUseCase) Execute(ctx context.Context, id string) error {
	return uc.notifications.MarkRead(ctx, id)
}

type MarkAllReadUseCase struct {
	notifications portnotification.Repository
}

func NewMarkAllReadUseCase(notifications portnotification.Repository) *MarkAllReadUseCase {
	return &MarkAllReadUseCase{notifications: notifications}
}

func (uc *MarkAllReadUseCase) Execute(ctx context.Context, userID int64) error {
	return uc.notifications.MarkAllRead(ctx, userID)
}
