package location

import (
	"context"
	"errors"
	"strings"

	"github.com/efangly/thanes-lims-backend/internal/domain/location"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func toDomain(m Model) location.Location {
	return location.Location{
		ID:        m.ID,
		ParentID:  m.ParentID,
		Name:      m.Name,
		LevelType: location.LevelType(m.LevelType),
	}
}

func toModel(l location.Location) Model {
	return Model{
		ID:        l.ID,
		ParentID:  l.ParentID,
		Name:      l.Name,
		LevelType: string(l.LevelType),
	}
}

func (r *Repository) Create(ctx context.Context, l location.Location) (location.Location, error) {
	m := toModel(l)
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return location.Location{}, err
	}
	return toDomain(m), nil
}

// CreateMany inserts the batch produced by "generate children" in one
// round-trip; all-or-nothing.
func (r *Repository) CreateMany(ctx context.Context, ls []location.Location) ([]location.Location, error) {
	models := make([]Model, len(ls))
	for i, l := range ls {
		models[i] = toModel(l)
	}
	if err := r.db.WithContext(ctx).Create(&models).Error; err != nil {
		return nil, err
	}

	out := make([]location.Location, len(models))
	for i, m := range models {
		out[i] = toDomain(m)
	}
	return out, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (location.Location, error) {
	var m Model
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return location.Location{}, shared.ErrNotFound
	}
	if err != nil {
		return location.Location{}, err
	}
	return toDomain(m), nil
}

func (r *Repository) ListChildren(ctx context.Context, parentID *string) ([]location.Location, error) {
	q := r.db.WithContext(ctx).Model(&Model{})
	if parentID == nil {
		q = q.Where("parent_id IS NULL")
	} else {
		q = q.Where("parent_id = ?", *parentID)
	}

	var models []Model
	if err := q.Order("name").Find(&models).Error; err != nil {
		return nil, err
	}

	out := make([]location.Location, len(models))
	for i, m := range models {
		out[i] = toDomain(m)
	}
	return out, nil
}

func (r *Repository) FindChildByName(ctx context.Context, parentID *string, name string) (location.Location, error) {
	q := r.db.WithContext(ctx).Model(&Model{})
	if parentID == nil {
		q = q.Where("parent_id IS NULL")
	} else {
		q = q.Where("parent_id = ?", *parentID)
	}

	var m Model
	err := q.Where("name = ?", name).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return location.Location{}, shared.ErrNotFound
	}
	if err != nil {
		return location.Location{}, err
	}
	return toDomain(m), nil
}

func (r *Repository) HasChildren(ctx context.Context, id string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Model{}).Where("parent_id = ?", id).Count(&count).Error
	return count > 0, err
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Where("id = ?", id).Delete(&Model{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return shared.ErrNotFound
	}
	return nil
}

// FullPath walks the tree up from id to its root Cabinet via a recursive
// CTE and joins the ancestor names root-first (e.g. "Fridge-A / Shelf-2").
// Computed fresh on every call rather than stored - see
// docs/adr/0001-self-referencing-tree-for-storage-location.md.
func (r *Repository) FullPath(ctx context.Context, id string) (string, error) {
	// Raw SQL bypasses GORM's automatic deleted_at IS NULL scoping, so both
	// legs of the recursive CTE filter Retired rows explicitly (see ADR
	// 0003) - a Full Path must never resolve through a Retired Location.
	var names []string
	err := r.db.WithContext(ctx).Raw(`
		WITH RECURSIVE ancestors AS (
			SELECT id, parent_id, name, 0 AS depth FROM locations WHERE id = ? AND deleted_at IS NULL
			UNION ALL
			SELECT l.id, l.parent_id, l.name, a.depth + 1
			FROM locations l
			JOIN ancestors a ON l.id = a.parent_id
			WHERE l.deleted_at IS NULL
		)
		SELECT name FROM ancestors ORDER BY depth DESC
	`, id).Scan(&names).Error
	if err != nil {
		return "", err
	}
	if len(names) == 0 {
		return "", shared.ErrNotFound
	}
	return strings.Join(names, " / "), nil
}
