package sample

import (
	"context"
	"errors"
	"fmt"

	"github.com/efangly/thanes-lims-backend/internal/domain/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portsample "github.com/efangly/thanes-lims-backend/internal/ports/sample"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// occupyingStatuses is sample.OccupyingStatuses as plain strings for a
// GORM `IN` clause.
func occupyingStatuses() []string {
	out := make([]string, len(sample.OccupyingStatuses))
	for i, s := range sample.OccupyingStatuses {
		out[i] = string(s)
	}
	return out
}

// toConflict maps a Postgres unique-violation (23505) - in practice only
// uq_samples_box_cell_active - onto shared.ErrConflict so the HTTP layer
// answers 409 instead of 500. Any other error passes through untouched.
func toConflict(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return fmt.Errorf("%w: cell already occupied", shared.ErrConflict)
	}
	return err
}

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func toDomain(m Model) sample.Sample {
	return sample.Sample{
		ID:              m.ID,
		Name:            m.Name,
		Type:            sample.Type(m.Type),
		CustodianUserID: m.CustodianUserID,
		LocationID:      m.LocationID,
		Status:          sample.Status(m.Status),
		ReceivedAt:      m.ReceivedAt,
		BarcodeID:       m.BarcodeID,
		Description:     m.Description,
		Position:        m.Position,
	}
}

func toModel(s sample.Sample) Model {
	return Model{
		ID:              s.ID,
		Name:            s.Name,
		Type:            string(s.Type),
		CustodianUserID: s.CustodianUserID,
		LocationID:      s.LocationID,
		Status:          string(s.Status),
		ReceivedAt:      s.ReceivedAt,
		BarcodeID:       s.BarcodeID,
		Description:     s.Description,
		Position:        s.Position,
	}
}

func (r *Repository) Create(ctx context.Context, s sample.Sample) (sample.Sample, error) {
	m := toModel(s)
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return sample.Sample{}, err
	}
	return toDomain(m), nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (sample.Sample, error) {
	var m Model
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return sample.Sample{}, shared.ErrNotFound
	}
	if err != nil {
		return sample.Sample{}, err
	}
	return toDomain(m), nil
}

func (r *Repository) FindByBarcodeID(ctx context.Context, barcodeID string) (sample.Sample, error) {
	var m Model
	err := r.db.WithContext(ctx).First(&m, "barcode_id = ?", barcodeID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return sample.Sample{}, shared.ErrNotFound
	}
	if err != nil {
		return sample.Sample{}, err
	}
	return toDomain(m), nil
}

func (r *Repository) List(ctx context.Context, filter portsample.ListFilter) ([]sample.Sample, error) {
	q := r.db.WithContext(ctx).Model(&Model{})
	if filter.Status != nil {
		q = q.Where("status = ?", string(*filter.Status))
	}
	if filter.Type != nil {
		q = q.Where("type = ?", string(*filter.Type))
	}
	if filter.BarcodeID != nil {
		q = q.Where("barcode_id = ?", *filter.BarcodeID)
	}
	if filter.CustodianUserID != nil {
		q = q.Where("custodian_user_id = ?", *filter.CustodianUserID)
	}
	if filter.LocationID != nil {
		q = q.Where("samples.location_id = ?", *filter.LocationID)
	}
	if filter.LocationText != nil {
		// Match against the assigned Location's own name; the join is to
		// locations, so Samples with no LocationID are excluded when this
		// filter is active.
		q = q.Select("samples.*").
			Joins("JOIN locations ON locations.id = samples.location_id").
			Where("locations.name ILIKE ?", "%"+*filter.LocationText+"%")
	}

	var models []Model
	if err := q.Order("received_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}

	samples := make([]sample.Sample, len(models))
	for i, m := range models {
		samples[i] = toDomain(m)
	}
	return samples, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, s sample.Sample) (sample.Sample, error) {
	if err := r.db.WithContext(ctx).Model(&Model{}).Where("id = ?", s.ID).Update("status", string(s.Status)).Error; err != nil {
		return sample.Sample{}, err
	}
	return r.FindByID(ctx, s.ID)
}

func (r *Repository) UpdateLocation(ctx context.Context, sampleID string, locationID, position *string) (sample.Sample, error) {
	err := r.db.WithContext(ctx).Model(&Model{}).Where("id = ?", sampleID).
		Updates(map[string]any{"location_id": locationID, "position": position}).Error
	if err != nil {
		return sample.Sample{}, toConflict(err)
	}
	return r.FindByID(ctx, sampleID)
}

func (r *Repository) ListActiveByLocation(ctx context.Context, locationID string) ([]sample.Sample, error) {
	var models []Model
	err := r.db.WithContext(ctx).
		Where("location_id = ? AND status IN ?", locationID, occupyingStatuses()).
		Order("position").Find(&models).Error
	if err != nil {
		return nil, err
	}
	out := make([]sample.Sample, len(models))
	for i, m := range models {
		out[i] = toDomain(m)
	}
	return out, nil
}

// MoveWithinBox reassigns Cells inside boxID atomically. Each Sample's
// location stays boxID; only position changes. A (location_id, position)
// clash - caught by uq_samples_box_cell_active - rolls back the whole batch
// and surfaces as shared.ErrConflict.
func (r *Repository) MoveWithinBox(ctx context.Context, boxID string, moves []portsample.PositionAssignment) ([]sample.Sample, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Park each moved Sample at a distinct temporary Cell first, so an
		// in-batch swap does not trip uq_samples_box_cell_active mid-loop.
		// The parking values stay non-NULL (a NULL position would instead
		// collide under the leaf-occupancy index uq_samples_active_location).
		for i, mv := range moves {
			parked := fmt.Sprintf("#%d", i)
			res := tx.Model(&Model{}).Where("id = ? AND location_id = ?", mv.SampleID, boxID).
				Update("position", parked)
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return shared.ErrNotFound
			}
		}
		for _, mv := range moves {
			pos := mv.Position
			if err := tx.Model(&Model{}).Where("id = ? AND location_id = ?", mv.SampleID, boxID).
				Update("position", &pos).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, toConflict(err)
	}
	return r.ListActiveByLocation(ctx, boxID)
}

func (r *Repository) UpdateBarcodeID(ctx context.Context, sampleID string, barcodeID *string) (sample.Sample, error) {
	if err := r.db.WithContext(ctx).Model(&Model{}).Where("id = ?", sampleID).Update("barcode_id", barcodeID).Error; err != nil {
		return sample.Sample{}, err
	}
	return r.FindByID(ctx, sampleID)
}

func (r *Repository) ExistsActiveByLocation(ctx context.Context, locationID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Model{}).
		Where("location_id = ? AND status IN ?", locationID, occupyingStatuses()).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) ExistsActiveByLocationPosition(ctx context.Context, locationID, position string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Model{}).
		Where("location_id = ? AND position = ? AND status IN ?", locationID, position, occupyingStatuses()).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) ExistsByLocation(ctx context.Context, locationID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Model{}).Where("location_id = ?", locationID).Count(&count).Error
	return count > 0, err
}
