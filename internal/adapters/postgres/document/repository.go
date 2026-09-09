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
		ID:                 m.ID,
		Name:               m.Name,
		Type:               document.Type(m.Type),
		Version:            m.Version,
		CreatedBy:          m.CreatedBy,
		IssuedAt:           m.IssuedAt,
		AccessLevel:        m.AccessLevel,
		Locked:             m.Locked,
		StorageKey:         m.StorageKey,
		EquipmentID:        m.EquipmentID,
		CalibrationEventID: m.CalibrationEventID,
	}
}

func toModel(d document.Document) Model {
	return Model{
		ID:                 d.ID,
		Name:               d.Name,
		Type:               string(d.Type),
		Version:            d.Version,
		CreatedBy:          d.CreatedBy,
		IssuedAt:           d.IssuedAt,
		AccessLevel:        d.AccessLevel,
		Locked:             d.Locked,
		StorageKey:         d.StorageKey,
		EquipmentID:        d.EquipmentID,
		CalibrationEventID: d.CalibrationEventID,
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

// FindByIDIncludingDeleted bypasses GORM's soft-delete scope so the
// restore flow can find a Retired Document.
func (r *Repository) FindByIDIncludingDeleted(ctx context.Context, id string) (document.Document, error) {
	var m Model
	err := r.db.WithContext(ctx).Unscoped().First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return document.Document{}, shared.ErrNotFound
	}
	if err != nil {
		return document.Document{}, err
	}
	return toDomain(m), nil
}

// Delete soft-deletes the Document - GORM stamps deleted_at because Model
// carries gorm.DeletedAt.
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

// Restore clears deleted_at. It distinguishes "no such id" (ErrNotFound)
// from "id exists but is still active" (ErrConflict).
func (r *Repository) Restore(ctx context.Context, id string) error {
	var m Model
	err := r.db.WithContext(ctx).Unscoped().First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return shared.ErrNotFound
	}
	if err != nil {
		return err
	}
	if !m.DeletedAt.Valid {
		return shared.ErrConflict
	}
	return r.db.WithContext(ctx).Unscoped().Model(&Model{}).Where("id = ?", id).Update("deleted_at", nil).Error
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

func (r *Repository) ListByEquipment(ctx context.Context, equipmentID string) ([]document.Document, error) {
	var models []Model
	if err := r.db.WithContext(ctx).Where("equipment_id = ?", equipmentID).Order("id").Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]document.Document, len(models))
	for i, m := range models {
		out[i] = toDomain(m)
	}
	return out, nil
}

func (r *Repository) ListByCalibrationEvent(ctx context.Context, calibrationEventID int64) ([]document.Document, error) {
	var models []Model
	if err := r.db.WithContext(ctx).Where("calibration_event_id = ?", calibrationEventID).Order("id").Find(&models).Error; err != nil {
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
