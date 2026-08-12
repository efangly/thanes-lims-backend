package notification

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/notification"
)

type Repository interface {
	Create(ctx context.Context, n notification.Notification) (notification.Notification, error)
	FindByID(ctx context.Context, id string) (notification.Notification, error)
	// ListForUser returns notifications addressed to userID plus every
	// broadcast notification, unread-first.
	ListForUser(ctx context.Context, userID int64) ([]notification.Notification, error)
	MarkRead(ctx context.Context, id string) error
	MarkAllRead(ctx context.Context, userID int64) error
}

// Notifier is what other modules depend on to emit a notification, kept
// separate from Repository so a module like Environment or Inventory only
// needs this narrow interface, not full notification CRUD.
type Notifier interface {
	Notify(ctx context.Context, n notification.Notification) error
}
