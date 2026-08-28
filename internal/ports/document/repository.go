package document

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/document"
	"github.com/efangly/thanes-lims-backend/internal/domain/equipment"
)

type Repository interface {
	Create(ctx context.Context, d document.Document) (document.Document, error)
	FindByID(ctx context.Context, id string) (document.Document, error)
	List(ctx context.Context) ([]document.Document, error)
	// ListByEquipment lists Documents linked to one Equipment (Phase 5).
	ListByEquipment(ctx context.Context, equipmentID string) ([]document.Document, error)
	// ListByCalibrationEvent lists Documents linked to one CalibrationEvent
	// (Phase 6, e.g. calibration certificates).
	ListByCalibrationEvent(ctx context.Context, calibrationEventID int64) ([]document.Document, error)
	Update(ctx context.Context, d document.Document) (document.Document, error)
}

// EquipmentDirectory validates a Document's optional EquipmentID link on
// upload. Returns shared.ErrNotFound when no such Equipment exists.
// *postgres/equipment.Repository satisfies this via FindByID.
type EquipmentDirectory interface {
	FindByID(ctx context.Context, id string) (equipment.Equipment, error)
}

// CalibrationEventDirectory validates a Document's optional
// CalibrationEventID link on upload (Phase 6). Returns shared.ErrNotFound
// when no such event exists. *postgres/equipment.CalibrationRepository
// satisfies this via FindByID.
type CalibrationEventDirectory interface {
	FindByID(ctx context.Context, id int64) (equipment.CalibrationEvent, error)
}

type HistoryRepository interface {
	Append(ctx context.Context, h document.DocHistory) (document.DocHistory, error)
	ListByDocument(ctx context.Context, documentID string) ([]document.DocHistory, error)
}
