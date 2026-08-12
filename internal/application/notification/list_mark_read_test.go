package notification_test

import (
	"context"
	"testing"

	applicationnotification "github.com/efangly/thanes-lims-backend/internal/application/notification"
	"github.com/efangly/thanes-lims-backend/internal/domain/notification"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMarkReadUseCase_Idempotent(t *testing.T) {
	repo := new(mockNotificationRepo)
	repo.On("MarkRead", mock.Anything, "NTF-001").Return(nil).Twice()

	uc := applicationnotification.NewMarkReadUseCase(repo)

	assert.NoError(t, uc.Execute(context.Background(), "NTF-001"))
	assert.NoError(t, uc.Execute(context.Background(), "NTF-001"))
	repo.AssertNumberOfCalls(t, "MarkRead", 2)
}

func TestListNotificationsUseCase_UnreadFirst(t *testing.T) {
	repo := new(mockNotificationRepo)
	want := []notification.Notification{
		{ID: "NTF-002", Read: false},
		{ID: "NTF-001", Read: true},
	}
	repo.On("ListForUser", mock.Anything, int64(1)).Return(want, nil)

	uc := applicationnotification.NewListNotificationsUseCase(repo)
	got, err := uc.Execute(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, want, got)
}
