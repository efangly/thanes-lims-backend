package notification

import (
	"context"
	"fmt"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/notification"
	portidgen "github.com/efangly/thanes-lims-backend/internal/ports/idgen"
	portnotification "github.com/efangly/thanes-lims-backend/internal/ports/notification"
)

type CreateNotificationUseCase struct {
	notifications portnotification.Repository
	idgen         portidgen.SequenceGenerator
}

func NewCreateNotificationUseCase(notifications portnotification.Repository, idgen portidgen.SequenceGenerator) *CreateNotificationUseCase {
	return &CreateNotificationUseCase{notifications: notifications, idgen: idgen}
}

type CreateNotificationInput struct {
	RecipientUserID *int64
	Tone            notification.Tone
	Icon            string
	Title           string
	Message         string
}

func (uc *CreateNotificationUseCase) Execute(ctx context.Context, in CreateNotificationInput) (notification.Notification, error) {
	seq, err := uc.idgen.Next(ctx, "notification", nil)
	if err != nil {
		return notification.Notification{}, err
	}

	return uc.notifications.Create(ctx, notification.Notification{
		ID:              fmt.Sprintf("NTF-%03d", seq),
		RecipientUserID: in.RecipientUserID,
		Tone:            in.Tone,
		Icon:            in.Icon,
		Title:           in.Title,
		Message:         in.Message,
		CreatedAt:       time.Now(),
	})
}

// AsNotifier adapts this use case to ports/notification.Notifier so other
// modules (Environment alerts, Inventory low-stock, TestResult approval,
// ...) can depend on the narrow Notifier interface and inject this without
// pulling in the rest of the notification module's CRUD.
type AsNotifier struct {
	create *CreateNotificationUseCase
}

func NewAsNotifier(create *CreateNotificationUseCase) *AsNotifier {
	return &AsNotifier{create: create}
}

func (n *AsNotifier) Notify(ctx context.Context, notif notification.Notification) error {
	_, err := n.create.Execute(ctx, CreateNotificationInput{
		RecipientUserID: notif.RecipientUserID,
		Tone:            notif.Tone,
		Icon:            notif.Icon,
		Title:           notif.Title,
		Message:         notif.Message,
	})
	return err
}
