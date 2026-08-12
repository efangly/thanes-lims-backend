package document

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/document"
	"gorm.io/gorm"
)

type HistoryRepository struct {
	db *gorm.DB
}

func NewHistoryRepository(db *gorm.DB) *HistoryRepository {
	return &HistoryRepository{db: db}
}

func historyToDomain(m HistoryModel) document.DocHistory {
	return document.DocHistory{
		ID:         m.ID,
		DocumentID: m.DocumentID,
		Version:    m.Version,
		Change:     m.Change,
		Date:       m.Date,
		Who:        m.Who,
	}
}

func (r *HistoryRepository) Append(ctx context.Context, h document.DocHistory) (document.DocHistory, error) {
	m := HistoryModel{
		DocumentID: h.DocumentID,
		Version:    h.Version,
		Change:     h.Change,
		Date:       h.Date,
		Who:        h.Who,
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return document.DocHistory{}, err
	}
	return historyToDomain(m), nil
}

func (r *HistoryRepository) ListByDocument(ctx context.Context, documentID string) ([]document.DocHistory, error) {
	var models []HistoryModel
	if err := r.db.WithContext(ctx).Where("document_id = ?", documentID).Order("date").Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]document.DocHistory, len(models))
	for i, m := range models {
		out[i] = historyToDomain(m)
	}
	return out, nil
}
