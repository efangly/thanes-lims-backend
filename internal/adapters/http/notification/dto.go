package notification

import (
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/notification"
)

type NotificationResponse struct {
	ID        string    `json:"id"`
	Tone      string    `json:"tone"`
	Icon      string    `json:"icon"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
	Read      bool      `json:"read"`
}

func toResponse(n notification.Notification) NotificationResponse {
	return NotificationResponse{
		ID: n.ID, Tone: string(n.Tone), Icon: n.Icon, Title: n.Title,
		Message: n.Message, CreatedAt: n.CreatedAt, Read: n.Read,
	}
}
