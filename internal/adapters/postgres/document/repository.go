package document

import (
	"context"
	"errors"

	"github.com/efangly/thanes-lims-backend/internal/domain/document"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func toDomain(m Model) document.Document {
	return document.Document{
		ID:          m.ID,
		Name:        m.Name,
		Type:        document.Type(m.Type),
		Version:     m.Version,
		CreatedBy:   m.CreatedBy,
		IssuedAt:    m.IssuedAt,
		AccessLevel: m.AccessLevel,
		Locked:      m.Locked,
		StorageKey:  m.StorageKey,
	}
}

func toModel(d document.Document) Model {
	return Model{
		ID:          d.ID,
		Name:        d.Name,
		Type:        string(d.Type),
		Version:     d.Version,
		CreatedBy:   d.CreatedBy,
		IssuedAt:    d.IssuedAt,
		AccessLevel: d.AccessLevel,
		Locked:      d.Locked,
		StorageKey:  d.StorageKey,
	}
}

func (r *Repository) Create(ctx context.Context, d document.Document) (document.Document, error) {
	m := toModel(d)
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return document.Document{}, err
	}
	return toDomain(m), nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (document.Document, error) {
	var m Model
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return document.Document{}, shared.ErrNotFound
	}
	if err != nil {
		return document.Document{}, err
	}
	return toDomain(m), nil
}

func (r *Repository) List(ctx context.Context) ([]document.Document, error) {
	var models []Model
	if err := r.db.WithContext(ctx).Order("id").Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]document.Document, len(models))
	for i, m := range models {
		out[i] = toDomain(m)
	}
	return out, nil
}

// Update uses Save (not Updates) so zero-value fields like Locked=false
// persist correctly - GORM's Updates skips zero values on a struct arg,
// which would silently no-op an unlock.
func (r *Repository) Update(ctx context.Context, d document.Document) (document.Document, error) {
	m := toModel(d)
	if err := r.db.WithContext(ctx).Save(&m).Error; err != nil {
		return document.Document{}, err
	}
	return r.FindByID(ctx, d.ID)
}
