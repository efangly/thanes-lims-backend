package document

import (
	"context"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/document"
	portdocument "github.com/efangly/thanes-lims-backend/internal/ports/document"
	portfilestorage "github.com/efangly/thanes-lims-backend/internal/ports/filestorage"
)

type ListDocumentsUseCase struct {
	documents portdocument.Repository
}

func NewListDocumentsUseCase(documents portdocument.Repository) *ListDocumentsUseCase {
	return &ListDocumentsUseCase{documents: documents}
}

func (uc *ListDocumentsUseCase) Execute(ctx context.Context) ([]document.Document, error) {
	return uc.documents.List(ctx)
}

// ExecuteByEquipment lists only the Documents linked to equipmentID (Phase
// 5) - backs GET /documents?equipment_id=.
func (uc *ListDocumentsUseCase) ExecuteByEquipment(ctx context.Context, equipmentID string) ([]document.Document, error) {
	return uc.documents.ListByEquipment(ctx, equipmentID)
}

// ExecuteByCalibrationEvent lists only the Documents linked to one
// CalibrationEvent (Phase 6) - backs GET /documents?calibration_event_id=.
func (uc *ListDocumentsUseCase) ExecuteByCalibrationEvent(ctx context.Context, calibrationEventID int64) ([]document.Document, error) {
	return uc.documents.ListByCalibrationEvent(ctx, calibrationEventID)
}

type GetDocumentUseCase struct {
	documents portdocument.Repository
}

func NewGetDocumentUseCase(documents portdocument.Repository) *GetDocumentUseCase {
	return &GetDocumentUseCase{documents: documents}
}

func (uc *GetDocumentUseCase) Execute(ctx context.Context, id string) (document.Document, error) {
	return uc.documents.FindByID(ctx, id)
}

// GetDownloadURLUseCase issues a short-lived presigned URL for the
// document's current StorageKey rather than proxying bytes through the API.
type GetDownloadURLUseCase struct {
	documents portdocument.Repository
	storage   portfilestorage.FileStorage
}

func NewGetDownloadURLUseCase(documents portdocument.Repository, storage portfilestorage.FileStorage) *GetDownloadURLUseCase {
	return &GetDownloadURLUseCase{documents: documents, storage: storage}
}

func (uc *GetDownloadURLUseCase) Execute(ctx context.Context, id string) (string, error) {
	d, err := uc.documents.FindByID(ctx, id)
	if err != nil {
		return "", err
	}
	return uc.storage.GetPresignedURL(ctx, d.StorageKey, 15*time.Minute)
}

type ListHistoryUseCase struct {
	history portdocument.HistoryRepository
}

func NewListHistoryUseCase(history portdocument.HistoryRepository) *ListHistoryUseCase {
	return &ListHistoryUseCase{history: history}
}

func (uc *ListHistoryUseCase) Execute(ctx context.Context, documentID string) ([]document.DocHistory, error) {
	return uc.history.ListByDocument(ctx, documentID)
}
