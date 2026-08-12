package notification

import (
	"context"
	"errors"

	"github.com/efangly/thanes-lims-backend/internal/domain/notification"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func toDomain(m Model) notification.Notification {
	return notification.Notification{
		ID:              m.ID,
		RecipientUserID: m.RecipientUserID,
		Tone:            notification.Tone(m.Tone),
		Icon:            m.Icon,
		Title:           m.Title,
		Message:         m.Message,
		CreatedAt:       m.CreatedAt,
		Read:            m.Read,
	}
}

func (r *Repository) Create(ctx context.Context, n notification.Notification) (notification.Notification, error) {
	m := Model{
		ID:              n.ID,
		RecipientUserID: n.RecipientUserID,
		Tone:            string(n.Tone),
		Icon:            n.Icon,
		Title:           n.Title,
		Message:         n.Message,
		CreatedAt:       n.CreatedAt,
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return notification.Notification{}, err
	}
	return toDomain(m), nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (notification.Notification, error) {
	var m Model
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return notification.Notification{}, shared.ErrNotFound
	}
	if err != nil {
		return notification.Notification{}, err
	}
	return toDomain(m), nil
}

func (r *Repository) ListForUser(ctx context.Context, userID int64) ([]notification.Notification, error) {
	var models []Model
	err := r.db.WithContext(ctx).
		Where("recipient_user_id = ? OR recipient_user_id IS NULL", userID).
		Order("read ASC, created_at DESC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	out := make([]notification.Notification, len(models))
	for i, m := range models {
		out[i] = toDomain(m)
	}
	return out, nil
}

func (r *Repository) MarkRead(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&Model{}).Where("id = ?", id).Update("read", true).Error
}

func (r *Repository) MarkAllRead(ctx context.Context, userID int64) error {
	return r.db.WithContext(ctx).Model(&Model{}).
		Where("recipient_user_id = ? OR recipient_user_id IS NULL", userID).
		Update("read", true).Error
}
