package audit

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/audit"
	portaudit "github.com/efangly/thanes-lims-backend/internal/ports/audit"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Repository implements ports/audit.AuditLogger directly against Postgres.
type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Log(ctx context.Context, entry audit.AuditLog) error {
	m := Model{
		ActorID:    entry.ActorID,
		ActorRole:  entry.ActorRole,
		Method:     entry.Method,
		Path:       entry.Path,
		Resource:   entry.Resource,
		ResourceID: entry.ResourceID,
		StatusCode: entry.StatusCode,
		Metadata:   datatypes.JSONMap(entry.Metadata),
		CreatedAt:  entry.CreatedAt,
	}
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *Repository) List(ctx context.Context, filter portaudit.ListFilter) ([]audit.AuditLog, error) {
	q := r.db.WithContext(ctx).Order("created_at DESC")
	if filter.From != nil {
		q = q.Where("created_at >= ?", *filter.From)
	}
	if filter.To != nil {
		q = q.Where("created_at <= ?", *filter.To)
	}

	var models []Model
	if err := q.Find(&models).Error; err != nil {
		return nil, err
	}

	out := make([]audit.AuditLog, len(models))
	for i, m := range models {
		out[i] = audit.AuditLog{
			ID:         m.ID,
			ActorID:    m.ActorID,
			ActorRole:  m.ActorRole,
			Method:     m.Method,
			Path:       m.Path,
			Resource:   m.Resource,
			ResourceID: m.ResourceID,
			StatusCode: m.StatusCode,
			Metadata:   m.Metadata,
			CreatedAt:  m.CreatedAt,
		}
	}
	return out, nil
}
