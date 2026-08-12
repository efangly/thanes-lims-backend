package audit

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/audit"
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
