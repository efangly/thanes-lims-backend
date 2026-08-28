package sample

import (
	"context"
	"errors"

	"github.com/efangly/thanes-lims-backend/internal/domain/sample"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portsample "github.com/efangly/thanes-lims-backend/internal/ports/sample"
	"gorm.io/gorm"
)

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

func (r *Repository) UpdateLocation(ctx context.Context, sampleID string, locationID *string) (sample.Sample, error) {
	if err := r.db.WithContext(ctx).Model(&Model{}).Where("id = ?", sampleID).Update("location_id", locationID).Error; err != nil {
		return sample.Sample{}, err
	}
	return r.FindByID(ctx, sampleID)
}

func (r *Repository) UpdateBarcodeID(ctx context.Context, sampleID string, barcodeID *string) (sample.Sample, error) {
	if err := r.db.WithContext(ctx).Model(&Model{}).Where("id = ?", sampleID).Update("barcode_id", barcodeID).Error; err != nil {
		return sample.Sample{}, err
	}
	return r.FindByID(ctx, sampleID)
}

func (r *Repository) ExistsActiveByLocation(ctx context.Context, locationID string) (bool, error) {
	occupying := make([]string, len(sample.OccupyingStatuses))
	for i, s := range sample.OccupyingStatuses {
		occupying[i] = string(s)
	}

	var count int64
	err := r.db.WithContext(ctx).Model(&Model{}).
		Where("location_id = ? AND status IN ?", locationID, occupying).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) ExistsByLocation(ctx context.Context, locationID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Model{}).Where("location_id = ?", locationID).Count(&count).Error
	return count > 0, err
}
