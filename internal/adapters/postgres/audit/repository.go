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
	q := applyFilter(r.db.WithContext(ctx), filter).Order("created_at DESC")
	if filter.Limit > 0 {
		q = q.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		q = q.Offset(filter.Offset)
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

func (r *Repository) Count(ctx context.Context, filter portaudit.ListFilter) (int64, error) {
	var count int64
	err := applyFilter(r.db.WithContext(ctx), filter).Model(&Model{}).Count(&count).Error
	return count, err
}

func applyFilter(q *gorm.DB, filter portaudit.ListFilter) *gorm.DB {
	if filter.From != nil {
		q = q.Where("created_at >= ?", *filter.From)
	}
	if filter.To != nil {
		q = q.Where("created_at <= ?", *filter.To)
	}
	if filter.ActorID != nil {
		q = q.Where("actor_id = ?", *filter.ActorID)
	}
	if filter.Resource != "" {
		q = q.Where("resource = ?", filter.Resource)
	}
	if filter.Method != "" {
		q = q.Where("method = ?", filter.Method)
	}
	return q
}
